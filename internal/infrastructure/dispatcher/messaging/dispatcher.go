package messaging

import (
	"bytes"
	"container/list"
	"context"
	"go.uber.org/zap"
	"sync"
	"time"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/infrastructure/dispatcher/messaging/ports"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/dto"
)

type Dispatcher struct {
	// Зависимости
	gateway ports.MessageGateway
	limiter ports.GlobalRateLimiter
	logger  logger.Logger

	// Внутреннее состояние
	mu              sync.Mutex
	activeCampaigns *list.List            // list.Element.Value is campaignID (string)
	queues          map[string]*list.List // campaignID -> list of *dto.Message
	resultsChans    map[string]chan<- *dto.MessageSendResult

	// Управление
	jobsChan chan dispatcherJobRequest
	stopOnce sync.Once
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewDispatcher(gateway ports.MessageGateway, limiter ports.GlobalRateLimiter, logger logger.Logger) *Dispatcher {
	return &Dispatcher{
		gateway:         gateway,
		limiter:         limiter,
		logger:          logger,
		activeCampaigns: list.New(),
		queues:          make(map[string]*list.List),
		resultsChans:    make(map[string]chan<- *dto.MessageSendResult),
		jobsChan:        make(chan dispatcherJobRequest),
		stopChan:        make(chan struct{}),
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	d.wg.Add(1)
	go d.run(ctx)
}

func (d *Dispatcher) Stop(ctx context.Context) error {
	d.stopOnce.Do(func() {
		close(d.stopChan)
	})

	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *Dispatcher) Submit(ctx context.Context, newJob *dto.DispatcherJob) (<-chan *dto.MessageSendResult, error) {
	d.logger.Info("Submitting new job", zap.String("campaignID", newJob.CampaignID), zap.Int("message_count", len(newJob.Messages)))

	resultsCh := make(chan *dto.MessageSendResult, len(newJob.Messages))
	errChan := make(chan error, 1)

	req := dispatcherJobRequest{
		job:         newJob,
		resultsChan: resultsCh,
		errChan:     errChan,
	}

	select {
	case d.jobsChan <- req:
		// Ждем ответа от run(), была ли принята работа
		select {
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
			return resultsCh, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	case <-d.stopChan:
		return nil, ErrDispatcherClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (d *Dispatcher) run(ctx context.Context) {
	defer d.wg.Done()
	ticker := time.NewTicker(100 * time.Millisecond) // Тикер для проверки очередей
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			d.handleShutdown()
			return
		case req := <-d.jobsChan:
			d.addJob(req)
			req.errChan <- nil // Сигнализируем, что работа принята
		case <-ticker.C:
			d.processNextMessage(ctx)
		}
	}
}

func (d *Dispatcher) addJob(req dispatcherJobRequest) {
	d.mu.Lock()
	defer d.mu.Unlock()

	id := req.job.CampaignID
	if _, exists := d.queues[id]; !exists {
		// Новая кампания
		d.queues[id] = list.New()
		d.activeCampaigns.PushBack(id)
		d.resultsChans[id] = req.resultsChan
	}
	// Добавляем сообщения в очередь
	for i := range req.job.Messages {
		d.queues[id].PushBack(&req.job.Messages[i])
	}
}

func (d *Dispatcher) processNextMessage(ctx context.Context) {
	d.mu.Lock()
	if d.activeCampaigns.Len() == 0 {
		d.mu.Unlock()
		return
	}

	element := d.activeCampaigns.Front()
	campaignID := element.Value.(string)
	queue := d.queues[campaignID]

	if queue.Len() == 0 {
		// Очередь пуста, кампания завершена
		d.activeCampaigns.Remove(element)
		if resultsChan, ok := d.resultsChans[campaignID]; ok {
			close(resultsChan)
			delete(d.resultsChans, campaignID)
		}
		delete(d.queues, campaignID)
		d.mu.Unlock()
		return
	}

	msgElement := queue.Front()
	message := queue.Remove(msgElement).(*dto.Message)
	d.mu.Unlock()

	// --- Длительные операции ---
	if err := d.limiter.Wait(ctx); err != nil {
		d.logger.Error("Rate limiter wait error", zap.Error(err), zap.String("campaignID", campaignID))
		return
	}

	// Отправляем сообщение
	result := d.send(ctx, *message)

	// --- Конец длительных операций ---

	d.mu.Lock()
	defer d.mu.Unlock()

	if resultsChan, ok := d.resultsChans[campaignID]; ok {
		resultsChan <- result
	}

	// Перемещаем кампанию в конец, чтобы обеспечить round-robin
	d.activeCampaigns.MoveToBack(element)
}

func (d *Dispatcher) handleShutdown() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.logger.Info("Shutting down dispatcher, closing active channels.")
	// Закрываем все оставшиеся каналы результатов
	for id, ch := range d.resultsChans {
		close(ch)
		delete(d.resultsChans, id)
	}
	// Очищаем очереди
	d.queues = make(map[string]*list.List)
	d.activeCampaigns = list.New()
}

func (d *Dispatcher) send(ctx context.Context, msg dto.Message) *dto.MessageSendResult {
	var result *dto.MessageSendResult
	var err error

	if msg.Media != nil {
		mediaReader := bytes.NewReader(msg.Media.Data)
		result, err = d.gateway.SendMediaMessage(ctx, msg.PhoneNumber, campaign.MessageType(msg.Media.MessageType), msg.Text, msg.Media.Filename, mediaReader, msg.Media.MimeType, false)
	} else {
		result, err = d.gateway.SendTextMessage(ctx, msg.PhoneNumber, msg.Text, false)
	}

	if err != nil {
		return &dto.MessageSendResult{
			PhoneNumber: msg.PhoneNumber,
			Success:     false,
			Error:       err.Error(),
			Timestamp:   time.Now(),
		}
	}
	if result == nil { // На случай, если gateway вернет nil, nil
		return &dto.MessageSendResult{
			PhoneNumber: msg.PhoneNumber,
			Success:     false,
			Error:       "gateway returned nil result and nil error",
			Timestamp:   time.Now(),
		}
	}
	return result
}

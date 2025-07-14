package messaging

import (
	"bytes"
	"container/list"
	"context"
	"sync"
	"time"
	"whatsapp-service/internal/entities/campaign"
	"whatsapp-service/internal/interfaces"
	"whatsapp-service/internal/usecases/dto"

	"go.uber.org/zap"
)

type Dispatcher struct {
	// Зависимости
	gateway interfaces.MessageGateway
	limiter GlobalRateLimiter
	logger  interfaces.Logger

	// Внутреннее состояние
	mu              sync.Mutex
	activeCampaigns *list.List
	queues          map[string]*list.List
	resultsChans    map[string]chan<- *dto.MessageSendResult

	// Управление
	jobsChan chan dispatcherJobRequest
	stopOnce sync.Once
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewDispatcher(gateway interfaces.MessageGateway, limiter GlobalRateLimiter, logger interfaces.Logger) *Dispatcher {
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
	d.logger.Info("Dispatcher starting")
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
		select {
		case <-done:
			return nil
		case <-time.After(5 * time.Second):
			return ctx.Err()
		}
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
	d.logger.Info("Dispatcher run loop started")
	ticker := time.NewTicker(100 * time.Millisecond) // Тикер для проверки очередей
	defer ticker.Stop()

	for {
		select {
		case <-d.stopChan:
			d.logger.Info("Dispatcher received stop signal")
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

		// Устанавливаем лимит для кампании
		d.limiter.SetRateForCampaign(id, req.job.MessagesPerHour)
		d.logger.Info("Set rate limit for campaign", zap.String("campaignID", id), zap.Int("messagesPerHour", req.job.MessagesPerHour))
	}
	// Добавляем сообщения в очередь
	for i := range req.job.Messages {
		d.queues[id].PushBack(&req.job.Messages[i])
	}

	d.logger.Info("Job added to queue", zap.String("campaignID", id), zap.Int("messagesInQueue", d.queues[id].Len()), zap.Int("activeCampaigns", d.activeCampaigns.Len()))
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
		d.logger.Info("Campaign completed, queue empty", zap.String("campaignID", campaignID))
		d.mu.Unlock()
		return
	}

	msgElement := queue.Front()
	message := queue.Remove(msgElement).(*dto.Message)
	d.mu.Unlock()

	// --- Длительные операции ---
	if err := d.limiter.WaitForCampaign(ctx, campaignID); err != nil {
		d.logger.Error("Campaign rate limiter wait error", zap.Error(err), zap.String("campaignID", campaignID))
		return
	}

	// Отправляем сообщение
	result := d.send(ctx, *message)

	// --- Конец длительных операций ---

	d.mu.Lock()
	defer d.mu.Unlock()

	if resultsChan, ok := d.resultsChans[campaignID]; ok {
		resultsChan <- result
	} else {
		d.logger.Warn("No results channel found for campaign", zap.String("campaignID", campaignID))
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

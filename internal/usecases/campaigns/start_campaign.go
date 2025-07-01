package campaigns

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate/types"

	"whatsapp-service/internal/entities"
	"whatsapp-service/internal/usecases/interfaces"
)

// StartCampaignRequest запрос на запуск кампании
type StartCampaignRequest struct {
	CampaignID string
	Async      bool
}

// StartCampaignResponse ответ на запрос запуска кампании
type StartCampaignResponse struct {
	CampaignID     string
	Status         entities.CampaignStatus
	TotalNumbers   int
	EstimatedTime  int
	AsyncStarted   bool
	InitialResults []types.MessageResult
}

// StartCampaignUseCase use case для запуска кампании
type StartCampaignUseCase struct {
	campaignRepo       interfaces.CampaignRepository
	campaignStatusRepo interfaces.CampaignStatusRepository
	messageGateway     interfaces.MessageGateway
	rateLimiter        interfaces.RateLimiter
	activeWorkers      map[string]context.CancelFunc
	workersMutex       sync.RWMutex
}

// NewStartCampaignUseCase создает новый экземпляр use case
func NewStartCampaignUseCase(
	campaignRepo interfaces.CampaignRepository,
	campaignStatusRepo interfaces.CampaignStatusRepository,
	messageGateway interfaces.MessageGateway,
	rateLimiter interfaces.RateLimiter,
) *StartCampaignUseCase {
	return &StartCampaignUseCase{
		campaignRepo:       campaignRepo,
		campaignStatusRepo: campaignStatusRepo,
		messageGateway:     messageGateway,
		rateLimiter:        rateLimiter,
		activeWorkers:      make(map[string]context.CancelFunc),
	}
}

// Execute выполняет запуск кампании
func (uc *StartCampaignUseCase) Execute(ctx context.Context, req StartCampaignRequest) (*StartCampaignResponse, error) {
	// Получаем кампанию
	campaign, err := uc.campaignRepo.GetByID(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	// Проверяем можно ли запустить
	if !campaign.CanBeStarted() {
		return nil, fmt.Errorf("campaign cannot be started")
	}

	// Запускаем кампанию
	err = campaign.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start campaign: %w", err)
	}

	// Получаем статусы для отправки
	statuses, err := uc.campaignStatusRepo.ListByCampaignID(ctx, req.CampaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign statuses: %w", err)
	}

	// Настраиваем rate limiter
	uc.rateLimiter.SetRate(req.CampaignID, campaign.MessagesPerHour())

	response := &StartCampaignResponse{
		CampaignID:     req.CampaignID,
		Status:         campaign.Status(),
		TotalNumbers:   len(statuses),
		EstimatedTime:  (len(statuses) * 60) / campaign.MessagesPerHour(),
		AsyncStarted:   req.Async,
		InitialResults: []types.MessageResult{},
	}

	if req.Async {
		go uc.startAsyncSending(ctx, campaign, statuses)
	} else {
		// Отправляем первые 3 сообщения синхронно
		if len(statuses) > 0 {
			maxBatch := 3
			if len(statuses) < maxBatch {
				maxBatch = len(statuses)
			}
			for i := 0; i < maxBatch; i++ {
				result := uc.sendMessage(ctx, campaign, statuses[i])
				response.InitialResults = append(response.InitialResults, result)
			}

			// Остальные асинхронно
			if len(statuses) > maxBatch {
				go uc.startAsyncSending(ctx, campaign, statuses[maxBatch:])
			}
		}
	}

	// Обновляем статус кампании
	uc.campaignRepo.UpdateStatus(ctx, req.CampaignID, entities.CampaignStatusStarted)

	return response, nil
}

// startAsyncSending запускает асинхронную отправку
func (uc *StartCampaignUseCase) startAsyncSending(ctx context.Context, campaign *entities.Campaign, statuses []*entities.CampaignPhoneStatus) {
	campaignID := campaign.ID()

	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	uc.workersMutex.Lock()
	uc.activeWorkers[campaignID] = cancel
	uc.workersMutex.Unlock()

	defer func() {
		uc.workersMutex.Lock()
		delete(uc.activeWorkers, campaignID)
		uc.workersMutex.Unlock()
	}()

	log.Printf("Starting async sending for campaign %s", campaignID)

	totalSent := 0
	totalFailed := 0

	for _, status := range statuses {
		select {
		case <-workerCtx.Done():
			log.Printf("Campaign %s was cancelled", campaignID)
			return
		default:
		}

		if !uc.rateLimiter.CanSend(campaignID) {
			waitTime := time.Duration(uc.rateLimiter.GetWaitTime(campaignID)) * time.Second
			select {
			case <-time.After(waitTime):
			case <-workerCtx.Done():
				return
			}
		}

		result := uc.sendMessage(workerCtx, campaign, status)

		if result.Success {
			status.MarkAsSent()
			totalSent++
		} else {
			status.MarkAsFailed(result.Error)
			totalFailed++
		}

		uc.campaignStatusRepo.Update(ctx, status)
		uc.rateLimiter.MessageSent(campaignID)
	}

	if totalFailed > totalSent/2 {
		campaign.Fail()
		uc.campaignRepo.UpdateStatus(ctx, campaign.ID(), entities.CampaignStatusFailed)
	} else {
		campaign.Finish()
		uc.campaignRepo.UpdateStatus(ctx, campaign.ID(), entities.CampaignStatusFinished)
	}

	log.Printf("Campaign %s completed: %d sent, %d failed", campaignID, totalSent, totalFailed)
}

// sendMessage отправляет одно сообщение
func (uc *StartCampaignUseCase) sendMessage(ctx context.Context, campaign *entities.Campaign, status *entities.CampaignPhoneStatus) types.MessageResult {
	phoneNumber := status.PhoneNumber()

	if campaign.Media() != nil {
		// Отправляем медиа-сообщение
		media := campaign.Media()
		mediaReader := &simpleReader{data: media.Data()}

		result, err := uc.messageGateway.SendMediaMessage(
			ctx,
			phoneNumber,
			media.MessageType(),
			campaign.Message(),
			media.Filename(),
			mediaReader,
			media.MimeType(),
			true)
		if err != nil {
			return types.MessageResult{
				PhoneNumber: phoneNumber,
				Success:     false,
				Error:       err.Error(),
				Timestamp:   time.Now().Format(time.RFC3339),
			}
		}
		return result
	}

	// Отправляем текстовое сообщение
	result, err := uc.messageGateway.SendTextMessage(ctx, phoneNumber, campaign.Message(), true)
	if err != nil {
		return types.MessageResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Error:       err.Error(),
			Timestamp:   time.Now().Format(time.RFC3339),
		}
	}
	return result
}

// StopCampaign останавливает активную кампанию
func (uc *StartCampaignUseCase) StopCampaign(campaignID string) error {
	uc.workersMutex.Lock()
	defer uc.workersMutex.Unlock()

	if cancel, exists := uc.activeWorkers[campaignID]; exists {
		cancel()
		return nil
	}

	return fmt.Errorf("no active worker found for campaign %s", campaignID)
}

// simpleReader простая реализация io.Reader
type simpleReader struct {
	data []byte
	pos  int
}

func (r *simpleReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

package ports

import (
	"context"
	"whatsapp-service/internal/usecases/dto"
)

// DispatcherJob представляет собой задание для диспетчера —
// набор сообщений для одной конкретной кампании.
type DispatcherJob struct {
	CampaignID string
	Messages   []dto.Message
}

// Dispatcher отвечает за оркестрацию отправки сообщений из нескольких кампаний,
// используя логику round-robin и соблюдая единый глобальный лимит скорости.
type Dispatcher interface {
	// Submit отправляет новое задание в очередь диспетчера.
	// Возвращает канал, из которого можно читать результаты отправки каждого сообщения.
	// Этот канал будет закрыт диспетчером, когда все сообщения из задания будут обработаны.
	Submit(ctx context.Context, job *dto.DispatcherJob) (<-chan *dto.MessageSendResult, error)

	// Start запускает фоновый процесс диспетчера. Должен быть вызван один раз при старте приложения.
	Start(ctx context.Context)

	// Stop плавно останавливает диспетчер, дождавшись завершения текущих задач.
	Stop(ctx context.Context) error
}

package messaging

import "whatsapp-service/internal/usecases/dto"

type dispatcherJobRequest struct {
	job         *dto.DispatcherJob
	resultsChan chan<- *dto.MessageSendResult
	errChan     chan<- error
}

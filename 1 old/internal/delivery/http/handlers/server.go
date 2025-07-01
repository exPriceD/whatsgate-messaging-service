package handlers

import (
	"whatsapp-service/internal/bulk/interfaces"
	"whatsapp-service/internal/logger"
	usecase "whatsapp-service/internal/whatsgate/usecase"
)

// Server представляет HTTP-обработчики
type Server struct {
	log              logger.Logger
	whatsgateService *usecase.SettingsUsecase
	bulkRepo         interfaces.BulkCampaignStorage
	statusRepo       interfaces.BulkCampaignStatusStorage
}

// NewServer создает новый сервер обработчиков
func NewServer(log logger.Logger, whatsgateService *usecase.SettingsUsecase, bulkRepo interfaces.BulkCampaignStorage, statusRepo interfaces.BulkCampaignStatusStorage) *Server {
	return &Server{
		log:              log,
		whatsgateService: whatsgateService,
		bulkRepo:         bulkRepo,
		statusRepo:       statusRepo,
	}
}

package handlers

import (
	"whatsapp-service/internal/logger"
	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"
)

// Server представляет HTTP-обработчики
type Server struct {
	log              logger.Logger
	whatsgateService *whatsgateDomain.SettingsService
}

// NewServer создает новый сервер обработчиков
func NewServer(log logger.Logger, whatsgateService *whatsgateDomain.SettingsService) *Server {
	return &Server{
		log:              log,
		whatsgateService: whatsgateService,
	}
}

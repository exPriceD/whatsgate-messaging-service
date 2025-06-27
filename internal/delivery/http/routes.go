package http

import (
	_ "whatsapp-service/internal/docs" // swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"whatsapp-service/internal/delivery/http/handlers/health"
	"whatsapp-service/internal/delivery/http/handlers/messages"
	"whatsapp-service/internal/delivery/http/handlers/settings"
)

const apiVersion = "1.0.0"

// setupRoutes настраивает маршруты сервера.
func (s *Server) setupRoutes() {
	// Swagger документация
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Корневые роуты для удобства (дублируют /api/v1)
	s.engine.GET("/health", health.ServiceHealthHandler())
	s.engine.GET("/status", health.StatusHandler(apiVersion))
	s.engine.GET("/settings", settings.GetSettingsHandler(s.whatsgateService))
	s.engine.PUT("/settings", settings.UpdateSettingsHandler(s.whatsgateService))
	s.engine.DELETE("/settings/reset", settings.ResetSettingsHandler(s.whatsgateService))
	s.engine.POST("/messages/send", messages.SendMessageHandler(s.whatsgateService))
	s.engine.POST("/messages/send-media", messages.SendMediaMessageHandler(s.whatsgateService))
	s.engine.POST("/messages/bulk-send", messages.BulkSendHandler(s.whatsgateService, s.bulkStorage, s.statusStorage))

	// API v1 (версионированные роуты)
	v1 := s.engine.Group("/api/v1")
	{
		v1.GET("/health", health.ServiceHealthHandler())
		v1.GET("/status", health.StatusHandler(apiVersion))

		// Настройки
		v1.GET("/settings", settings.GetSettingsHandler(s.whatsgateService))
		v1.PUT("/settings", settings.UpdateSettingsHandler(s.whatsgateService))
		v1.DELETE("/settings/reset", settings.ResetSettingsHandler(s.whatsgateService))

		// Отправка сообщений
		v1.POST("/messages/send", messages.SendMessageHandler(s.whatsgateService))
		v1.POST("/messages/send-media", messages.SendMediaMessageHandler(s.whatsgateService))
		v1.POST("/messages/bulk-send", messages.BulkSendHandler(s.whatsgateService, s.bulkStorage, s.statusStorage))
	}
}

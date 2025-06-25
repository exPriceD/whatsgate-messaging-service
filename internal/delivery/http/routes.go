package http

import (
	_ "whatsapp-service/internal/docs" // swagger docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// setupRoutes настраивает маршруты сервера.
func (s *Server) setupRoutes() {
	// Swagger документация
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Корневые роуты для удобства (дублируют /api/v1)
	s.engine.GET("/health", s.handlers.HealthHandler)
	s.engine.GET("/status", s.handlers.StatusHandler)
	s.engine.GET("/settings", s.handlers.GetSettingsHandler)
	s.engine.PUT("/settings", s.handlers.UpdateSettingsHandler)
	s.engine.DELETE("/settings/reset", s.handlers.ResetSettingsHandler)
	s.engine.POST("/messages/send", s.handlers.SendMessageHandler)
	s.engine.POST("/messages/send-media", s.handlers.SendMediaMessageHandler)
	s.engine.POST("/messages/bulk-send", s.handlers.BulkSendHandler)

	// API v1 (версионированные роуты)
	v1 := s.engine.Group("/api/v1")
	{
		v1.GET("/health", s.handlers.HealthHandler)
		v1.GET("/status", s.handlers.StatusHandler)

		// Настройки WhatGate
		v1.GET("/settings", s.handlers.GetSettingsHandler)
		v1.PUT("/settings", s.handlers.UpdateSettingsHandler)
		v1.DELETE("/settings/reset", s.handlers.ResetSettingsHandler)

		// Отправка сообщений
		v1.POST("/messages/send", s.handlers.SendMessageHandler)
		v1.POST("/messages/send-media", s.handlers.SendMediaMessageHandler)
		v1.POST("/messages/bulk-send", s.handlers.BulkSendHandler)
	}
}

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

	// API v1
	v1 := s.engine.Group("/api/" + APIVersion)
	{
		v1.GET(HealthPath, s.healthHandler)
		v1.GET(StatusPath, s.statusHandler)
		// Здесь будут добавлены другие эндпоинты
	}
}

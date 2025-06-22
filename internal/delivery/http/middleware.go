package http

import (
	"whatsapp-service/internal/middleware"
)

// setupMiddleware настраивает middleware для сервера.
func (s *Server) setupMiddleware() {
	s.engine.Use(middleware.Recovery(s.logger))
	s.engine.Use(middleware.CORSMiddleware(s.config.CORS))
	s.engine.Use(middleware.Logging(s.logger))
}

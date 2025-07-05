package http

import (
	"whatsapp-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

// LoggerToContextMiddleware кладёт логгер в gin.Context
func (s *Server) loggerToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("logger", s.logger)
		c.Next()
	}
}

// setupMiddleware настраивает middleware для сервера.
func (s *Server) setupMiddleware() {
	s.engine.Use(middleware.RequestIDMiddleware())
	s.engine.Use(s.loggerToContextMiddleware())
	s.engine.Use(middleware.Recovery(s.logger))
	s.engine.Use(middleware.ErrorHandler(s.logger))
	s.engine.Use(middleware.CORSMiddleware(s.config.CORS))
	s.engine.Use(middleware.Logging(s.logger))
}

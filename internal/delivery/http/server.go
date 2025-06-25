package http

import (
	"context"
	"fmt"
	"net/http"

	"whatsapp-service/internal/config"
	"whatsapp-service/internal/delivery/http/handlers"
	"whatsapp-service/internal/logger"
	whatsgate_domain "whatsapp-service/internal/whatsgate/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server представляет HTTP-сервер приложения.
type Server struct {
	engine           *gin.Engine
	server           *http.Server
	logger           logger.Logger
	config           config.HTTPConfig
	handlers         *handlers.Server
	whatsgateService *whatsgate_domain.SettingsService
}

// NewServer создает новый HTTP-сервер.
func NewServer(cfg config.HTTPConfig, log logger.Logger, whatsgateService *whatsgate_domain.SettingsService) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	handlersServer := handlers.NewServer(log, whatsgateService)

	server := &Server{
		engine:           engine,
		logger:           log,
		config:           cfg,
		handlers:         handlersServer,
		whatsgateService: whatsgateService,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:      engine,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}

	server.setupMiddleware()
	server.setupRoutes()
	return server
}

// Start запускает HTTP-сервер.
func (s *Server) Start() error {
	s.logger.Info("starting HTTP server",
		zap.String("address", s.server.Addr),
		zap.String("host", s.config.Host),
		zap.Int("port", s.config.Port),
	)

	return s.server.ListenAndServe()
}

// Shutdown корректно завершает работу HTTP-сервера.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// GetEngine возвращает gin.Engine для добавления дополнительных роутов.
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

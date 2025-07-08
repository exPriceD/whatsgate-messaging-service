package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"whatsapp-service/internal/delivery/http/handlers"
	"whatsapp-service/internal/shared/logger"
)

// HTTPServer представляет HTTP сервер
type HTTPServer struct {
	server *http.Server
	router *Router
	logger logger.Logger
}

// NewHTTPServer создает новый HTTP сервер
func NewHTTPServer(
	port int,
	campaignHandler *handlers.CampaignsHandler,
	messagingHandler *handlers.MessagingHandler,
	settingsHandler *handlers.SettingsHandler,
	healthHandler *handlers.HealthHandler,
	logger logger.Logger,
) *HTTPServer {
	router := NewRouter(campaignHandler, messagingHandler, settingsHandler, healthHandler)

	return &HTTPServer{
		router: router,
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      router.SetupRoutes(),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start запускает HTTP сервер
func (s *HTTPServer) Start() error {
	s.logger.Info("Starting HTTP server", "addr", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop останавливает HTTP сервер
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server...")
	return s.server.Shutdown(ctx)
}

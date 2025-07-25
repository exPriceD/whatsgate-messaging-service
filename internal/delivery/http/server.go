package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"whatsapp-service/internal/interfaces"

	"whatsapp-service/internal/delivery/http/handlers"
)

// HTTPServer представляет HTTP сервер
type HTTPServer struct {
	server *http.Server
	router *Router
	logger interfaces.Logger
}

// NewHTTPServer создает новый HTTP сервер
func NewHTTPServer(
	port int,
	campaignHandler *handlers.CampaignsHandler,
	messagingHandler *handlers.MessagingHandler,
	whatsgateSettingsHandler *handlers.WhatsgateSettingsHandler,
	retailCRMSettingsHandler *handlers.RetailCRMSettingsHandler,
	retailCRMHandler *handlers.RetailCRMHandler,
	healthHandler *handlers.HealthHandler,
	logger interfaces.Logger,
) *HTTPServer {
	router := NewRouter(campaignHandler, messagingHandler, whatsgateSettingsHandler, retailCRMSettingsHandler, healthHandler, retailCRMHandler, logger)

	return &HTTPServer{
		router: router,
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      router.SetupRoutes(),
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
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

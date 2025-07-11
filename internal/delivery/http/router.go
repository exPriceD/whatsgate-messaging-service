package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "whatsapp-service/docs"
	"whatsapp-service/internal/delivery/http/handlers"
	customMiddleware "whatsapp-service/internal/delivery/http/middleware"
)

// Router обрабатывает HTTP маршрутизацию и настройку middleware
type Router struct {
	campaigns         *handlers.CampaignsHandler
	messaging         *handlers.MessagingHandler
	whatsgateSettings *handlers.WhatsgateSettingsHandler
	retailcrmSettings *handlers.RetailCRMSettingsHandler
	health            *handlers.HealthHandler
	retailcrm         *handlers.RetailCRMHandler
}

// NewRouter создает новый роутер со всеми обработчиками
func NewRouter(
	campaignHandler *handlers.CampaignsHandler,
	messagingHandler *handlers.MessagingHandler,
	whatsgateSettingsHandler *handlers.WhatsgateSettingsHandler,
	retailcrmSettingsHandler *handlers.RetailCRMSettingsHandler,
	healthHandler *handlers.HealthHandler,
	retailcrmHandler *handlers.RetailCRMHandler,
) *Router {
	return &Router{
		campaigns:         campaignHandler,
		messaging:         messagingHandler,
		whatsgateSettings: whatsgateSettingsHandler,
		retailcrmSettings: retailcrmSettingsHandler,
		health:            healthHandler,
		retailcrm:         retailcrmHandler,
	}
}

// SetupRoutes настраивает все маршруты с middleware
func (rt *Router) SetupRoutes() http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.CORS)
	r.Use(customMiddleware.Logging)

	// Health check endpoints - no versioning
	r.Get("/health", rt.health.Check)
	r.Get("/health/ready", rt.health.Ready)
	r.Get("/health/live", rt.health.Live)

	// Swagger UI
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Campaigns
		r.Route("/campaigns", func(r chi.Router) {
			// Список кампаний
			r.Get("/", rt.campaigns.List)

			// Создание новой кампании
			r.Post("/", rt.campaigns.Create)

			r.Route("/{id}", func(r chi.Router) {
				// Получение кампании по ID
				r.Get("/", rt.campaigns.GetByID)

				// Операции с кампанией
				r.Post("/start", rt.campaigns.Start)
				r.Post("/cancel", rt.campaigns.Cancel)
			})
		})

		// Messaging
		r.Post("/test-message", rt.messaging.SendTestMessage)

		// Whatsgate Settings
		r.Route("/whatsgate-settings", func(r chi.Router) {
			r.Get("/", rt.whatsgateSettings.Get)
			r.Put("/", rt.whatsgateSettings.Update)
			r.Delete("/reset", rt.whatsgateSettings.Reset)
		})

		// RetailCRM Settings
		r.Route("/retailcrm-settings", func(r chi.Router) {
			r.Get("/", rt.retailcrmSettings.Get)
			r.Put("/", rt.retailcrmSettings.Update)
			r.Delete("/reset", rt.retailcrmSettings.Reset)
		})

		// RetailCRM
		r.Route("/retailcrm", func(r chi.Router) {
			r.Get("/categories", rt.retailcrm.GetAvailableCategories)
			r.Post("/filter-customers", rt.retailcrm.FilterCustomersByCategory)
			r.Get("/test-connection", rt.retailcrm.TestConnection)
		})
	})

	// 404 handler
	r.NotFound(rt.health.NotFound)

	return r
}

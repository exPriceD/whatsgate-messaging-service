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
	campaigns *handlers.CampaignsHandler
	messaging *handlers.MessagingHandler
	settings  *handlers.SettingsHandler
	health    *handlers.HealthHandler
}

// NewRouter создает новый роутер со всеми обработчиками
func NewRouter(
	campaignHandler *handlers.CampaignsHandler,
	messagingHandler *handlers.MessagingHandler,
	settingsHandler *handlers.SettingsHandler,
	healthHandler *handlers.HealthHandler,
) *Router {
	return &Router{
		campaigns: campaignHandler,
		messaging: messagingHandler,
		settings:  settingsHandler,
		health:    healthHandler,
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

		// Settings
		r.Route("/settings", func(r chi.Router) {
			r.Get("/", rt.settings.Get)
			r.Put("/", rt.settings.Update)
			r.Delete("/reset", rt.settings.Reset)
		})
	})

	// 404 handler
	r.NotFound(rt.health.NotFound)

	return r
}

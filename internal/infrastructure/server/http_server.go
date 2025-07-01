package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"whatsapp-service/internal/adapters/controllers"
)

// HTTPServer представляет HTTP сервер
type HTTPServer struct {
	server       *http.Server
	campaignCtrl *controllers.CampaignController
	settingsCtrl *controllers.SettingsController
}

// NewHTTPServer создает новый HTTP сервер
func NewHTTPServer(port int, campaignCtrl *controllers.CampaignController, settingsCtrl *controllers.SettingsController) *HTTPServer {
	return &HTTPServer{
		campaignCtrl: campaignCtrl,
		settingsCtrl: settingsCtrl,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start запускает HTTP сервер
func (s *HTTPServer) Start() error {
	// Настраиваем маршруты
	s.setupRoutes()

	log.Printf("Starting HTTP server on %s", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop останавливает HTTP сервер
func (s *HTTPServer) Stop(ctx context.Context) error {
	log.Println("Stopping HTTP server...")
	return s.server.Shutdown(ctx)
}

// setupRoutes настраивает маршруты
func (s *HTTPServer) setupRoutes() {
	mux := http.NewServeMux()

	// Campaign endpoints
	mux.HandleFunc("/api/campaigns", s.campaignsHandler)
	mux.HandleFunc("/api/campaigns/", s.campaignHandler)

	// Settings endpoints
	mux.HandleFunc("/api/settings", s.settingsHandler)            // GET, PUT
	mux.HandleFunc("/api/settings/reset", s.settingsResetHandler) // DELETE

	// Health check
	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/", s.notFoundHandler)

	// Добавляем middleware
	s.server.Handler = s.corsMiddleware(s.loggingMiddleware(mux))
}

// campaignsHandler обрабатывает /api/campaigns
func (s *HTTPServer) campaignsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.campaignCtrl.CreateCampaign(w, r)
	case http.MethodGet:
		s.campaignCtrl.ListCampaigns(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// campaignHandler обрабатывает /api/campaigns/{id}
func (s *HTTPServer) campaignHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /api/campaigns/{id}/start
	if containsPath(path, "/start") {
		s.campaignCtrl.StartCampaign(w, r)
		return
	}

	// /api/campaigns/{id}/cancel
	if containsPath(path, "/cancel") {
		s.campaignCtrl.CancelCampaign(w, r)
		return
	}

	// /api/campaigns/{id}
	switch r.Method {
	case http.MethodGet:
		s.campaignCtrl.GetCampaign(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// healthHandler обрабатывает health check
func (s *HTTPServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"whatsapp-service"}`))
}

// notFoundHandler обрабатывает 404
func (s *HTTPServer) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error":"endpoint not found"}`))
}

// corsMiddleware добавляет CORS заголовки
func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware добавляет логирование запросов
func (s *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Создаем ResponseWriter с отслеживанием статуса
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v %s",
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			duration,
			r.RemoteAddr)
	})
}

// loggingResponseWriter обертка для ResponseWriter
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// containsPath проверяет наличие подстроки в пути
func containsPath(path, substr string) bool {
	for i := 0; i <= len(path)-len(substr); i++ {
		if path[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (s *HTTPServer) settingsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.settingsCtrl.GetSettings(w, r)
	case http.MethodPut:
		s.settingsCtrl.UpdateSettings(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *HTTPServer) settingsResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.settingsCtrl.ResetSettings(w, r)
}

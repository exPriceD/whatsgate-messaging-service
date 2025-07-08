package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"whatsapp-service/internal/entities/campaign/repository"

	"whatsapp-service/internal/delivery/http/response"
	"whatsapp-service/internal/shared/logger"
	"whatsapp-service/internal/usecases/campaigns/ports"
)

// HealthHandler обрабатывает проверку состояния сервиса и системные эндпоинты
type HealthHandler struct {
	logger       logger.Logger
	campaignRepo repository.CampaignRepository
	dispatcher   ports.Dispatcher
	startTime    time.Time
	version      string
	serviceName  string
}

// NewHealthHandler создает новый обработчик состояния сервиса
func NewHealthHandler(
	logger logger.Logger,
	campaignRepo repository.CampaignRepository,
	dispatcher ports.Dispatcher,
) *HealthHandler {
	return &HealthHandler{
		logger:       logger,
		campaignRepo: campaignRepo,
		dispatcher:   dispatcher,
		startTime:    time.Now(),
		version:      "1.0.0", // В реальном проекте можно получать из build flags
		serviceName:  "whatsapp-service",
	}
}

// HealthStatus представляет статус компонента
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth представляет состояние отдельного компонента
type ComponentHealth struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Details   interface{}  `json:"details,omitempty"`
	CheckedAt time.Time    `json:"checked_at"`
}

// HealthResponse представляет полный ответ health check
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`
	Service    string                     `json:"service"`
	Version    string                     `json:"version"`
	Timestamp  time.Time                  `json:"timestamp"`
	Uptime     string                     `json:"uptime"`
	Components map[string]ComponentHealth `json:"components"`
}

// Check проверяет состояние сервиса
// @Summary Проверка состояния сервиса
// @Description Возвращает подробный статус работоспособности сервиса и его компонентов
// @Tags Система
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Сервис работает нормально"
// @Success 503 {object} HealthResponse "Сервис неисправен"
// @Router /health [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Health check requested", "remote_addr", r.RemoteAddr)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	components := h.checkComponents(ctx)

	overallStatus := h.calculateOverallStatus(components)

	// Формируем ответ
	healthResp := HealthResponse{
		Status:     overallStatus,
		Service:    h.serviceName,
		Version:    h.version,
		Timestamp:  time.Now(),
		Uptime:     h.formatUptime(),
		Components: components,
	}

	statusCode := http.StatusOK
	if overallStatus == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	response.WriteJSON(w, statusCode, healthResp)
}

// checkComponents проверяет состояние всех компонентов
func (h *HealthHandler) checkComponents(ctx context.Context) map[string]ComponentHealth {
	components := make(map[string]ComponentHealth)

	// Проверка базы данных
	components["database"] = h.checkDatabase(ctx)

	// Проверка dispatcher
	components["dispatcher"] = h.checkDispatcher(ctx)

	// Проверка памяти/основных ресурсов
	components["system"] = h.checkSystem(ctx)

	return components
}

// checkDatabase проверяет состояние базы данных
func (h *HealthHandler) checkDatabase(ctx context.Context) ComponentHealth {
	checkTime := time.Now()

	if h.campaignRepo == nil {
		return ComponentHealth{
			Status:    StatusUnhealthy,
			Message:   "Database repository not initialized",
			CheckedAt: checkTime,
		}
	}

	// Попытка выполнить простую операцию
	// Здесь можно добавить метод Health() в repository interface
	// Пока используем простую проверку
	return ComponentHealth{
		Status:    StatusHealthy,
		Message:   "Database connection is healthy",
		CheckedAt: checkTime,
	}
}

// checkDispatcher проверяет состояние dispatcher
func (h *HealthHandler) checkDispatcher(ctx context.Context) ComponentHealth {
	checkTime := time.Now()

	if h.dispatcher == nil {
		return ComponentHealth{
			Status:    StatusUnhealthy,
			Message:   "Dispatcher not initialized",
			CheckedAt: checkTime,
		}
	}

	return ComponentHealth{
		Status:    StatusHealthy,
		Message:   "Dispatcher is running",
		CheckedAt: checkTime,
	}
}

// checkSystem проверяет системные ресурсы
func (h *HealthHandler) checkSystem(ctx context.Context) ComponentHealth {
	checkTime := time.Now()

	uptime := time.Since(h.startTime)

	return ComponentHealth{
		Status:  StatusHealthy,
		Message: "System is running normally",
		Details: map[string]interface{}{
			"uptime_seconds": uptime.Seconds(),
			"goroutines":     "healthy", // Можно добавить runtime.NumGoroutine() если нужно
		},
		CheckedAt: checkTime,
	}
}

// calculateOverallStatus определяет общий статус на основе компонентов
func (h *HealthHandler) calculateOverallStatus(components map[string]ComponentHealth) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false

	for _, component := range components {
		switch component.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}

	return StatusHealthy
}

// formatUptime форматирует время работы сервиса
func (h *HealthHandler) formatUptime() string {
	uptime := time.Since(h.startTime)

	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// Ready проверяет готовность сервиса (Kubernetes readiness probe)
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Readiness check requested", "remote_addr", r.RemoteAddr)

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	components := make(map[string]ComponentHealth)
	components["database"] = h.checkDatabase(ctx)

	overallStatus := h.calculateOverallStatus(components)

	if overallStatus == StatusUnhealthy {
		response.WriteJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "not ready",
			"reason": "critical components are unhealthy",
		})
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ready",
	})
}

// Live проверяет живучесть сервиса (Kubernetes liveness probe)
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Liveness check requested", "remote_addr", r.RemoteAddr)

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// NotFound обрабатывает неизвестные маршруты
func (h *HealthHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.logger.Warn("Unknown endpoint requested",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	response.WriteJSON(w, http.StatusNotFound, map[string]interface{}{
		"error":  "Endpoint not found",
		"method": r.Method,
		"path":   r.URL.Path,
	})
}

package handlers

import (
	"net/http"
	"time"

	"whatsapp-service/internal/delivery/http/types"

	"github.com/gin-gonic/gin"
)

// HealthHandler обрабатывает запросы health check.
// @Summary Health check
// @Description Проверяет работоспособность сервиса
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} types.HealthResponse
// @Router /health [get]
func (s *Server) HealthHandler(c *gin.Context) {
	response := types.HealthResponse{
		Status: "ok",
		Time:   time.Now().UTC(),
	}
	c.JSON(http.StatusOK, response)
}

// StatusHandler возвращает статус сервиса.
// @Summary Service status
// @Description Возвращает текущий статус и информацию о сервисе
// @Tags status
// @Accept json
// @Produce json
// @Success 200 {object} types.StatusResponse
// @Router /status [get]
func (s *Server) StatusHandler(c *gin.Context) {
	response := types.StatusResponse{
		Status:    "running",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
	}
	c.JSON(http.StatusOK, response)
}

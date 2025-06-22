package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// healthHandler обрабатывает запросы health check.
// @Summary Health check
// @Description Проверяет работоспособность сервиса
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (s *Server) healthHandler(c *gin.Context) {
	response := HealthResponse{
		Status: StatusOK,
		Time:   time.Now().UTC(),
	}
	c.JSON(http.StatusOK, response)
}

// statusHandler возвращает статус сервиса.
// @Summary Service status
// @Description Возвращает текущий статус и информацию о сервисе
// @Tags status
// @Accept json
// @Produce json
// @Success 200 {object} StatusResponse
// @Router /status [get]
func (s *Server) statusHandler(c *gin.Context) {
	response := StatusResponse{
		Status:    StatusRunning,
		Timestamp: time.Now().UTC(),
		Version:   AppVersion,
	}
	c.JSON(http.StatusOK, response)
}

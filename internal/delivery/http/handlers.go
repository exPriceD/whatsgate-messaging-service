package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// healthHandler обрабатывает запросы health check.
func (s *Server) healthHandler(c *gin.Context) {
	response := HealthResponse{
		Status: StatusOK,
		Time:   time.Now().UTC(),
	}
	c.JSON(http.StatusOK, response)
}

// statusHandler возвращает статус сервиса.
func (s *Server) statusHandler(c *gin.Context) {
	response := StatusResponse{
		Status:    StatusRunning,
		Timestamp: time.Now().UTC(),
		Version:   AppVersion,
	}
	c.JSON(http.StatusOK, response)
}

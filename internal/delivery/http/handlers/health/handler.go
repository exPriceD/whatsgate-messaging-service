package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceHealthHandler возвращает gin.HandlerFunc
func ServiceHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := HealthResponse{
			Status: "ok",
			Time:   time.Now().UTC(),
		}
		c.JSON(http.StatusOK, response)
	}
}

// StatusHandler возвращает gin.HandlerFunc
func StatusHandler(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := StatusResponse{
			Status:    "running",
			Timestamp: time.Now().UTC(),
			Version:   version,
		}
		c.JSON(http.StatusOK, response)
	}
}

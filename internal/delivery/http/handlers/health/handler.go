package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceHealthHandler godoc
// @Summary Проверка работоспособности сервиса
// @Description Проверяет работоспособность сервиса
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "OK"
// @Failure 400 {object} types.AppErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.AppErrorResponse "Внутренняя ошибка сервера"
// @Router /health [get]
func ServiceHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		response := HealthResponse{
			Status: "ok",
			Time:   time.Now().UTC(),
		}
		c.JSON(http.StatusOK, response)
	}
}

// StatusHandler godoc
// @Summary Статус сервиса
// @Description Возвращает текущий статус и информацию о сервисе
// @Tags status
// @Accept json
// @Produce json
// @Success 200 {object} StatusResponse "OK"
// @Router /status [get]
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

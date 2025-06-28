package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceHealthHandler возвращает статус здоровья сервиса
// @Summary Проверка здоровья сервиса
// @Description Возвращает статус здоровья сервиса
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "Сервис работает"
// @Failure 400 {object} types.ClientErrorResponse "Ошибка валидации"
// @Failure 500 {object} types.ClientErrorResponse "Внутренняя ошибка сервера"
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

// StatusHandler возвращает информацию о статусе сервиса
// @Summary Статус сервиса
// @Description Возвращает информацию о статусе сервиса
// @Tags status
// @Accept json
// @Produce json
// @Success 200 {object} StatusResponse "Статус сервиса"
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

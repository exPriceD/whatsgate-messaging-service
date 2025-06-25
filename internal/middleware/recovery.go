package middleware

import (
	"net/http"
	"whatsapp-service/internal/delivery/http/types"
	"whatsapp-service/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery перехватывает панику, логирует её и возвращает структурированную ошибку
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.Any("error", r),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
				)

				c.JSON(http.StatusInternalServerError, types.ErrorResponse{
					Error:   "Internal server error",
					Message: "An unexpected error occurred",
					Code:    http.StatusInternalServerError,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

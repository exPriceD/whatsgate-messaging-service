package middleware

import (
	"time"

	"whatsapp-service/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logging логирует все HTTP-запросы через logger.Logger.
func Logging(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetString("request_id")

		log.Info("request started",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("request_id", requestID),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		c.Next()
		latency := time.Since(start)

		log.Info("request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("request_id", requestID),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}

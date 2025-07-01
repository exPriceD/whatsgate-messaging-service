package middleware

import (
	"net/http"
	"strings"

	"whatsapp-service/internal/config"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware возвращает middleware для CORS на основе настроек.
func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}
		origin := c.GetHeader("Origin")
		allowed := false
		for _, o := range cfg.AllowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ","))
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ","))
			c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ","))
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			c.Header("Access-Control-Max-Age", string(rune(cfg.MaxAge)))
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

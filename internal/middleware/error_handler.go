package middleware

import (
	"net/http"
	"whatsapp-service/internal/delivery/http/types"
	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandler middleware для централизованной обработки ошибок
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Логируем ошибку
			log.Error("request error",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Error(err),
			)

			var statusCode int
			var errorResponse types.ErrorResponse

			switch e := err.(type) {
			case *appErr.AppError:
				switch e.Code {
				case "VALIDATION_ERROR":
					statusCode = http.StatusBadRequest
					errorResponse = types.ErrorResponse{
						Error:   "Validation failed",
						Message: e.Message,
						Code:    statusCode,
					}
				case "NOT_CONFIGURED":
					statusCode = http.StatusBadRequest
					errorResponse = types.ErrorResponse{
						Error:   "WhatGate not configured",
						Message: e.Message,
						Code:    statusCode,
					}
				case "UNAUTHORIZED":
					statusCode = http.StatusUnauthorized
					errorResponse = types.ErrorResponse{
						Error:   "Unauthorized",
						Message: e.Message,
						Code:    statusCode,
					}
				default:
					statusCode = http.StatusInternalServerError
					errorResponse = types.ErrorResponse{
						Error:   "Internal server error",
						Message: e.Message,
						Code:    statusCode,
					}
				}
			default:
				statusCode = http.StatusInternalServerError
				errorResponse = types.ErrorResponse{
					Error:   "Internal server error",
					Message: err.Error(),
					Code:    statusCode,
				}
			}

			c.JSON(statusCode, errorResponse)
			c.Abort()
		}
	}
}

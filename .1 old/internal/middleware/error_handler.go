package middleware

import (
	"errors"
	"net/http"
	"time"
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

			// Логируем полную ошибку на сервере
			log.Error("request error",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Error(err),
			)

			var statusCode int
			var clientErrorResponse types.ClientErrorResponse

			var e *appErr.AppError
			switch {
			case errors.As(err, &e):
				switch e.Code {
				case "VALIDATION_ERROR":
					statusCode = http.StatusBadRequest
				case "NOT_CONFIGURED":
					statusCode = http.StatusBadRequest
				case "UNAUTHORIZED":
					statusCode = http.StatusUnauthorized
				case "NOT_FOUND":
					statusCode = http.StatusNotFound
				case "SEND_ERROR", "API_ERROR", "SERVER_ERROR":
					statusCode = http.StatusBadGateway
				case "DB_STORAGE_ERROR", "DB_INIT_ERROR", "DB_LOAD_ERROR", "DB_QUERY_ERROR", "DB_SAVE_ERROR", "DB_DELETE_ERROR", "DB_HISTORY_ERROR", "DB_SCAN_ERROR", "DB_ROWS_ERROR", "DB_POOL_CREATE_ERROR", "DB_POOL_PARSE_ERROR", "DB_CONFIG_INVALID", "DB_BULK_INIT_ERROR", "DB_BULK_CREATE_ERROR", "DB_BULK_UPDATE_STATUS_ERROR", "DB_BULK_UPDATE_PROCESSED_ERROR", "DB_BULK_GET_ERROR", "DB_BULK_LIST_ERROR", "DB_BULK_SCAN_ERROR", "DB_BULK_STATUS_INIT_ERROR", "DB_BULK_STATUS_CREATE_ERROR", "DB_BULK_STATUS_UPDATE_ERROR", "DB_BULK_STATUS_LIST_ERROR", "DB_BULK_STATUS_SCAN_ERROR":
					statusCode = http.StatusServiceUnavailable
				case "STORAGE_ERROR", "BULK_STORAGE_CREATE_ERROR", "BULK_STORAGE_UPDATE_STATUS_ERROR", "BULK_STORAGE_UPDATE_PROCESSED_ERROR", "BULK_STORAGE_GET_ERROR", "BULK_STORAGE_LIST_ERROR", "BULK_STATUS_STORAGE_CREATE_ERROR", "BULK_STATUS_STORAGE_UPDATE_ERROR", "BULK_STATUS_STORAGE_LIST_ERROR":
					statusCode = http.StatusServiceUnavailable
				case "CONFIG_FILE_OPEN_ERROR", "CONFIG_DECODE_ERROR", "CONFIG_VALIDATE_ERROR", "CONFIG_LOAD_ERROR":
					statusCode = http.StatusInternalServerError
				case "RESET_ERROR", "LIST_ERROR":
					statusCode = http.StatusInternalServerError
				default:
					statusCode = http.StatusInternalServerError
				}

				// Создаем упрощенную ошибку для клиента
				clientErrorResponse = types.ClientErrorResponse{
					Message:     e.Message,
					Description: e.Description,
					Code:        e.Code,
					HTTPStatus:  statusCode,
					Timestamp:   e.Timestamp,
				}

				// Дополнительно логируем детальную ошибку на сервере
				log.Error("detailed error",
					zap.String("type", string(e.Type)),
					zap.String("code", e.Code),
					zap.String("severity", string(e.Severity)),
					zap.Any("context", e.Context),
					zap.Any("stack", e.Stack),
					zap.String("version", e.Version),
				)
			default:
				statusCode = http.StatusInternalServerError
				clientErrorResponse = types.ClientErrorResponse{
					Message:     "An unexpected error occurred",
					Description: "Please try again later or contact support if the problem persists",
					Code:        "INTERNAL_ERROR",
					HTTPStatus:  statusCode,
					Timestamp:   time.Now(),
				}
			}

			c.JSON(statusCode, clientErrorResponse)
			c.Abort()
		}
	}
}

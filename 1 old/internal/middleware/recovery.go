package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"
	"whatsapp-service/internal/delivery/http/types"
	appErrors "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery перехватывает панику, логирует её и возвращает структурированную ошибку
func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				startTime := time.Now()
				stack := debug.Stack()

				errorContext := &appErrors.ErrorContext{
					RequestID: c.GetString("request_id"),
					Method:    c.Request.Method,
					Path:      c.Request.URL.Path,
					IP:        c.ClientIP(),
					UserAgent: c.Request.UserAgent(),
					Component: "panic_recovery",
				}

				var panicErr error
				if err, ok := r.(error); ok {
					panicErr = err
				} else {
					panicErr = errors.New("panic occurred")
				}

				appErr := appErrors.New(
					appErrors.ErrorTypeInternal,
					"PANIC_RECOVERED",
					"An unexpected panic occurred",
					panicErr,
				).WithContext(errorContext).WithMetadata("panic_value", r)

				log.Error("panic recovered",
					append(appErr.ToZapFields(),
						zap.Duration("latency", time.Since(startTime)),
						zap.String("stack_trace", string(stack)),
						zap.Any("panic_value", r),
					)...,
				)

				errorResponse := types.AppErrorResponse{
					Type:        string(appErr.Type),
					Code:        appErr.Code,
					Message:     appErr.Message,
					Description: "An unexpected panic occurred and was recovered",
					Severity:    string(appErr.Severity),
					Context: &types.ErrorContext{
						RequestID: errorContext.RequestID,
						Method:    errorContext.Method,
						Path:      errorContext.Path,
						IP:        errorContext.IP,
						UserAgent: errorContext.UserAgent,
						Component: errorContext.Component,
						Metadata:  appErr.Context.Metadata,
					},
					Stack: []types.StackFrame{
						{
							Function: "panic_recovery",
							File:     "recovery.go",
							Line:     55,
						},
					},
					Timestamp:  time.Now(),
					HTTPStatus: http.StatusInternalServerError,
					Version:    "1.0.0",
				}

				c.Header("X-Error-Type", string(appErr.Type))
				c.Header("X-Error-Code", appErr.Code)
				c.Header("X-Request-ID", errorContext.RequestID)

				c.JSON(http.StatusInternalServerError, errorResponse)
				c.Abort()
			}
		}()
		c.Next()
	}
}

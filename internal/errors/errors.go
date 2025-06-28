package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrorType определяет тип ошибки
type ErrorType string

const (
	// Client errors (4xx)
	ErrorTypeValidation      ErrorType = "VALIDATION_ERROR"
	ErrorTypeUnauthorized    ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden       ErrorType = "FORBIDDEN"
	ErrorTypeNotFound        ErrorType = "NOT_FOUND"
	ErrorTypeConflict        ErrorType = "CONFLICT"
	ErrorTypeTooManyRequests ErrorType = "TOO_MANY_REQUESTS"

	// Server errors (5xx)
	ErrorTypeInternal           ErrorType = "INTERNAL_ERROR"
	ErrorTypeServiceUnavailable ErrorType = "SERVICE_UNAVAILABLE"
	ErrorTypeGatewayTimeout     ErrorType = "GATEWAY_TIMEOUT"

	// External service errors
	ErrorTypeExternalService ErrorType = "EXTERNAL_SERVICE_ERROR"
	ErrorTypeNetwork         ErrorType = "NETWORK_ERROR"
	ErrorTypeTimeout         ErrorType = "TIMEOUT_ERROR"

	// Business logic errors
	ErrorTypeBusinessLogic ErrorType = "BUSINESS_LOGIC_ERROR"
	ErrorTypeConfiguration ErrorType = "CONFIGURATION_ERROR"
	ErrorTypeDatabase      ErrorType = "DATABASE_ERROR"
	ErrorTypeStorage       ErrorType = "STORAGE_ERROR"
)

// ErrorSeverity определяет серьезность ошибки
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "LOW"
	ErrorSeverityMedium   ErrorSeverity = "MEDIUM"
	ErrorSeverityHigh     ErrorSeverity = "HIGH"
	ErrorSeverityCritical ErrorSeverity = "CRITICAL"
)

// ErrorContext содержит контекстную информацию об ошибке
type ErrorContext struct {
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	Component  string                 `json:"component,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	IP         string                 `json:"ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AppError представляет структурированную ошибку приложения
type AppError struct {
	// Основная информация
	Type        ErrorType `json:"type"`
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Description string    `json:"description,omitempty"`

	// Метаданные
	Severity ErrorSeverity `json:"severity"`
	Context  *ErrorContext `json:"context,omitempty"`

	// Техническая информация
	Cause     error     `json:"-"`
	Stack     []Frame   `json:"stack,omitempty"`
	Timestamp time.Time `json:"timestamp"`

	// HTTP информация
	HTTPStatus int `json:"http_status,omitempty"`

	// Версионирование
	Version string `json:"version"`
}

// Frame представляет кадр стека вызовов
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// Error возвращает строковое представление ошибки
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap возвращает причину ошибки
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is проверяет, является ли ошибка определенного типа
func (e *AppError) Is(target error) bool {
	if target == nil {
		return false
	}

	if appErr, ok := target.(*AppError); ok {
		return e.Type == appErr.Type && e.Code == appErr.Code
	}

	return false
}

// WithContext добавляет контекст к ошибке
func (e *AppError) WithContext(ctx *ErrorContext) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}

	if ctx.RequestID != "" {
		e.Context.RequestID = ctx.RequestID
	}
	if ctx.UserID != "" {
		e.Context.UserID = ctx.UserID
	}
	if ctx.SessionID != "" {
		e.Context.SessionID = ctx.SessionID
	}
	if ctx.ResourceID != "" {
		e.Context.ResourceID = ctx.ResourceID
	}
	if ctx.Operation != "" {
		e.Context.Operation = ctx.Operation
	}
	if ctx.Component != "" {
		e.Context.Component = ctx.Component
	}
	if ctx.Method != "" {
		e.Context.Method = ctx.Method
	}
	if ctx.Path != "" {
		e.Context.Path = ctx.Path
	}
	if ctx.IP != "" {
		e.Context.IP = ctx.IP
	}
	if ctx.UserAgent != "" {
		e.Context.UserAgent = ctx.UserAgent
	}
	if ctx.Metadata != nil {
		if e.Context.Metadata == nil {
			e.Context.Metadata = make(map[string]interface{})
		}
		for k, v := range ctx.Metadata {
			e.Context.Metadata[k] = v
		}
	}

	return e
}

// WithMetadata добавляет метаданные к ошибке
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = &ErrorContext{}
	}
	if e.Context.Metadata == nil {
		e.Context.Metadata = make(map[string]interface{})
	}
	e.Context.Metadata[key] = value
	return e
}

// WithSeverity устанавливает серьезность ошибки
func (e *AppError) WithSeverity(severity ErrorSeverity) *AppError {
	e.Severity = severity
	return e
}

// ToZapFields конвертирует ошибку в поля для zap логгера
func (e *AppError) ToZapFields() []zap.Field {
	fields := []zap.Field{
		zap.String("error_type", string(e.Type)),
		zap.String("error_code", e.Code),
		zap.String("error_message", e.Message),
		zap.String("error_severity", string(e.Severity)),
		zap.Time("error_timestamp", e.Timestamp),
		zap.String("error_version", e.Version),
	}

	if e.Description != "" {
		fields = append(fields, zap.String("error_description", e.Description))
	}

	if e.HTTPStatus != 0 {
		fields = append(fields, zap.Int("http_status", e.HTTPStatus))
	}

	if e.Context != nil {
		if e.Context.RequestID != "" {
			fields = append(fields, zap.String("request_id", e.Context.RequestID))
		}
		if e.Context.UserID != "" {
			fields = append(fields, zap.String("user_id", e.Context.UserID))
		}
		if e.Context.Operation != "" {
			fields = append(fields, zap.String("operation", e.Context.Operation))
		}
		if e.Context.Component != "" {
			fields = append(fields, zap.String("component", e.Context.Component))
		}
		if e.Context.Method != "" {
			fields = append(fields, zap.String("method", e.Context.Method))
		}
		if e.Context.Path != "" {
			fields = append(fields, zap.String("path", e.Context.Path))
		}
		if e.Context.IP != "" {
			fields = append(fields, zap.String("ip", e.Context.IP))
		}
		if e.Context.Metadata != nil {
			fields = append(fields, zap.Any("metadata", e.Context.Metadata))
		}
	}

	if e.Cause != nil {
		fields = append(fields, zap.Error(e.Cause))
	}

	if len(e.Stack) > 0 {
		fields = append(fields, zap.Any("stack", e.Stack))
	}

	return fields
}

// ToJSON конвертирует ошибку в JSON
func (e *AppError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// New создает новую ошибку
func New(errorType ErrorType, code, message string, cause error) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Severity:   getDefaultSeverity(errorType),
		Cause:      cause,
		Stack:      captureStack(),
		Timestamp:  time.Now(),
		HTTPStatus: getDefaultHTTPStatus(errorType),
		Version:    "1.0.0",
	}
}

// NewValidationError создает ошибку валидации
func NewValidationError(message string, details ...string) *AppError {
	desc := ""
	if len(details) > 0 {
		desc = details[0]
	}

	return &AppError{
		Type:        ErrorTypeValidation,
		Code:        "VALIDATION_ERROR",
		Message:     message,
		Description: desc,
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// NewNotFoundError создает ошибку "не найдено"
func NewNotFoundError(resource, id string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s with id '%s' not found", resource, id),
		Severity:   ErrorSeverityLow,
		Stack:      captureStack(),
		Timestamp:  time.Now(),
		HTTPStatus: http.StatusNotFound,
		Version:    "1.0.0",
	}
}

// NewUnauthorizedError создает ошибку авторизации
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:       ErrorTypeUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    message,
		Severity:   ErrorSeverityHigh,
		Stack:      captureStack(),
		Timestamp:  time.Now(),
		HTTPStatus: http.StatusUnauthorized,
		Version:    "1.0.0",
	}
}

// NewInternalError создает внутреннюю ошибку сервера
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Code:       "INTERNAL_ERROR",
		Message:    message,
		Severity:   ErrorSeverityCritical,
		Cause:      cause,
		Stack:      captureStack(),
		Timestamp:  time.Now(),
		HTTPStatus: http.StatusInternalServerError,
		Version:    "1.0.0",
	}
}

// NewExternalServiceError создает ошибку внешнего сервиса
func NewExternalServiceError(service, operation string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeExternalService,
		Code:       "EXTERNAL_SERVICE_ERROR",
		Message:    fmt.Sprintf("External service '%s' failed during '%s'", service, operation),
		Severity:   ErrorSeverityHigh,
		Cause:      cause,
		Stack:      captureStack(),
		Timestamp:  time.Now(),
		HTTPStatus: http.StatusBadGateway,
		Version:    "1.0.0",
	}
}

// NewDatabaseError создает ошибку базы данных
func NewDatabaseError(operation string, cause error) *AppError {
	return &AppError{
		Type:       ErrorTypeDatabase,
		Code:       "DATABASE_ERROR",
		Message:    fmt.Sprintf("Database operation '%s' failed", operation),
		Severity:   ErrorSeverityHigh,
		Cause:      cause,
		Stack:      captureStack(),
		Timestamp:  time.Now(),
		HTTPStatus: http.StatusServiceUnavailable,
		Version:    "1.0.0",
	}
}

// NewTimeoutError создает ошибку таймаута
func NewTimeoutError(operation string, timeout time.Duration) *AppError {
	return &AppError{
		Type:       ErrorTypeTimeout,
		Code:       "TIMEOUT_ERROR",
		Message:    fmt.Sprintf("Operation '%s' timed out after %v", operation, timeout),
		Severity:   ErrorSeverityMedium,
		Stack:      captureStack(),
		Timestamp:  time.Now(),
		HTTPStatus: http.StatusGatewayTimeout,
		Version:    "1.0.0",
	}
}

// Wrap оборачивает существующую ошибку в AppError
func Wrap(err error, errorType ErrorType, code, message string) *AppError {
	if err == nil {
		return nil
	}

	// Если ошибка уже AppError, добавляем контекст
	if appErr, ok := err.(*AppError); ok {
		appErr.Type = errorType
		appErr.Code = code
		appErr.Message = message
		appErr.Severity = getDefaultSeverity(errorType)
		appErr.HTTPStatus = getDefaultHTTPStatus(errorType)
		return appErr
	}

	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		Severity:   getDefaultSeverity(errorType),
		Cause:      err,
		Stack:      captureStack(),
		Timestamp:  time.Now(),
		HTTPStatus: getDefaultHTTPStatus(errorType),
		Version:    "1.0.0",
	}
}

// FromContext создает контекст ошибки из gin.Context
func FromContext(c context.Context) *ErrorContext {
	ctx := &ErrorContext{}

	// Извлекаем request_id из контекста
	if requestID, ok := c.Value("request_id").(string); ok {
		ctx.RequestID = requestID
	}

	// Извлекаем user_id из контекста
	if userID, ok := c.Value("user_id").(string); ok {
		ctx.UserID = userID
	}

	// Извлекаем session_id из контекста
	if sessionID, ok := c.Value("session_id").(string); ok {
		ctx.SessionID = sessionID
	}

	return ctx
}

// Вспомогательные функции

func getDefaultSeverity(errorType ErrorType) ErrorSeverity {
	switch errorType {
	case ErrorTypeValidation, ErrorTypeNotFound:
		return ErrorSeverityLow
	case ErrorTypeUnauthorized, ErrorTypeForbidden, ErrorTypeConflict, ErrorTypeTooManyRequests:
		return ErrorSeverityMedium
	case ErrorTypeExternalService, ErrorTypeNetwork, ErrorTypeTimeout, ErrorTypeDatabase, ErrorTypeStorage:
		return ErrorSeverityHigh
	case ErrorTypeInternal, ErrorTypeServiceUnavailable, ErrorTypeGatewayTimeout:
		return ErrorSeverityCritical
	default:
		return ErrorSeverityMedium
	}
}

func getDefaultHTTPStatus(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	case ErrorTypeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorTypeGatewayTimeout:
		return http.StatusGatewayTimeout
	case ErrorTypeExternalService:
		return http.StatusBadGateway
	case ErrorTypeNetwork, ErrorTypeTimeout:
		return http.StatusGatewayTimeout
	case ErrorTypeDatabase, ErrorTypeStorage:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func captureStack() []Frame {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := make([]Frame, 0, n)

	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i])
		if fn == nil {
			continue
		}
		file, line := fn.FileLine(pcs[i])
		frames = append(frames, Frame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// Глобальные экземпляры для удобства использования
var (
	globalValidator *ErrorValidator
	globalUtils     *ErrorUtils
	initOnce        sync.Once
)

// InitErrorSystem инициализирует глобальную систему ошибок
func InitErrorSystem() {
	initOnce.Do(func() {
		globalValidator = NewErrorValidator()
		globalUtils = NewErrorUtils()
	})
}

// GetErrorValidator возвращает глобальный экземпляр валидатора
func GetErrorValidator() *ErrorValidator {
	if globalValidator == nil {
		InitErrorSystem()
	}
	return globalValidator
}

// GetErrorUtils возвращает глобальный экземпляр утилит
func GetErrorUtils() *ErrorUtils {
	if globalUtils == nil {
		InitErrorSystem()
	}
	return globalUtils
}

// NewWithValidation создает новую ошибку с валидацией
func NewWithValidation(errorType ErrorType, code, message string, cause error) (*AppError, error) {
	err := New(errorType, code, message, cause)

	// Валидируем ошибку
	validator := GetErrorValidator()
	if validationErr := validator.ValidateAppError(err); validationErr != nil {
		return nil, fmt.Errorf("error validation failed: %w", validationErr)
	}

	return err, nil
}

// NewValidationErrorWithValidation создает ошибку валидации с проверкой
func NewValidationErrorWithValidation(message string, details ...string) (*AppError, error) {
	err := NewValidationError(message, details...)

	// Валидируем ошибку
	validator := GetErrorValidator()
	if validationErr := validator.ValidateAppError(err); validationErr != nil {
		return nil, fmt.Errorf("error validation failed: %w", validationErr)
	}

	return err, nil
}

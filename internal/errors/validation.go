package errors

import (
	"fmt"
	"strings"
	"time"
)

// ErrorValidator предоставляет валидацию для ошибок
type ErrorValidator struct{}

// NewErrorValidator создает новый экземпляр ErrorValidator
func NewErrorValidator() *ErrorValidator {
	return &ErrorValidator{}
}

// ValidateAppError проверяет корректность AppError
func (ev *ErrorValidator) ValidateAppError(err *AppError) error {
	if err == nil {
		return fmt.Errorf("AppError cannot be nil")
	}

	var validationErrors []string

	// Проверяем обязательные поля
	if err.Type == "" {
		validationErrors = append(validationErrors, "Type is required")
	}

	if err.Code == "" {
		validationErrors = append(validationErrors, "Code is required")
	}

	if err.Message == "" {
		validationErrors = append(validationErrors, "Message is required")
	}

	// Проверяем корректность кода ошибки
	if err.Code != "" && !ev.isValidErrorCode(err.Code) {
		validationErrors = append(validationErrors, "Code must be in UPPER_SNAKE_CASE format")
	}

	// Проверяем корректность типа ошибки
	if err.Type != "" && !ev.isValidErrorType(err.Type) {
		validationErrors = append(validationErrors, "Invalid error type")
	}

	// Проверяем корректность серьезности
	if err.Severity != "" && !ev.isValidErrorSeverity(err.Severity) {
		validationErrors = append(validationErrors, "Invalid error severity")
	}

	// Проверяем HTTP статус
	if err.HTTPStatus != 0 && !ev.isValidHTTPStatus(err.HTTPStatus) {
		validationErrors = append(validationErrors, "Invalid HTTP status code")
	}

	// Проверяем временную метку
	if err.Timestamp.IsZero() {
		validationErrors = append(validationErrors, "Timestamp is required")
	}

	if !err.Timestamp.IsZero() && err.Timestamp.After(time.Now().Add(time.Minute)) {
		validationErrors = append(validationErrors, "Timestamp cannot be in the future")
	}

	// Проверяем версию
	if err.Version == "" {
		validationErrors = append(validationErrors, "Version is required")
	}

	// Проверяем контекст
	if err.Context != nil {
		if contextErr := ev.validateErrorContext(err.Context); contextErr != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Context validation failed: %v", contextErr))
		}
	}

	// Проверяем метаданные
	if err.Context != nil && err.Context.Metadata != nil {
		if metadataErr := ev.validateMetadata(err.Context.Metadata); metadataErr != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Metadata validation failed: %v", metadataErr))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("AppError validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// validateErrorContext проверяет корректность ErrorContext
func (ev *ErrorValidator) validateErrorContext(ctx *ErrorContext) error {
	if ctx == nil {
		return nil
	}

	var validationErrors []string

	// Проверяем RequestID
	if ctx.RequestID != "" && !ev.isValidRequestID(ctx.RequestID) {
		validationErrors = append(validationErrors, "Invalid RequestID format")
	}

	// Проверяем Operation
	if ctx.Operation != "" && !ev.isValidOperation(ctx.Operation) {
		validationErrors = append(validationErrors, "Invalid Operation format")
	}

	// Проверяем Component
	if ctx.Component != "" && !ev.isValidComponent(ctx.Component) {
		validationErrors = append(validationErrors, "Invalid Component format")
	}

	// Проверяем Method
	if ctx.Method != "" && !ev.isValidHTTPMethod(ctx.Method) {
		validationErrors = append(validationErrors, "Invalid HTTP method")
	}

	// Проверяем Path
	if ctx.Path != "" && !ev.isValidPath(ctx.Path) {
		validationErrors = append(validationErrors, "Invalid Path format")
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("ErrorContext validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// validateMetadata проверяет корректность метаданных
func (ev *ErrorValidator) validateMetadata(metadata map[string]interface{}) error {
	if metadata == nil {
		return nil
	}

	var validationErrors []string

	for key, value := range metadata {
		// Проверяем ключ
		if key == "" {
			validationErrors = append(validationErrors, "Metadata key cannot be empty")
			continue
		}

		if !ev.isValidMetadataKey(key) {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid metadata key format: %s", key))
		}

		// Проверяем значение
		if value == nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Metadata value cannot be nil for key: %s", key))
			continue
		}

		// Проверяем тип значения
		if !ev.isValidMetadataValue(value) {
			validationErrors = append(validationErrors, fmt.Sprintf("Invalid metadata value type for key: %s", key))
		}

		// Рекурсивно проверяем вложенные метаданные
		if nestedMap, ok := value.(map[string]interface{}); ok {
			if nestedErr := ev.validateMetadata(nestedMap); nestedErr != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("Nested metadata validation failed for key %s: %v", key, nestedErr))
			}
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("Metadata validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}

// isValidErrorCode проверяет формат кода ошибки
func (ev *ErrorValidator) isValidErrorCode(code string) bool {
	if code == "" {
		return false
	}

	// Код должен быть в формате UPPER_SNAKE_CASE
	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// isValidErrorType проверяет корректность типа ошибки
func (ev *ErrorValidator) isValidErrorType(errorType ErrorType) bool {
	validTypes := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeUnauthorized,
		ErrorTypeForbidden,
		ErrorTypeNotFound,
		ErrorTypeConflict,
		ErrorTypeTooManyRequests,
		ErrorTypeInternal,
		ErrorTypeServiceUnavailable,
		ErrorTypeGatewayTimeout,
		ErrorTypeExternalService,
		ErrorTypeNetwork,
		ErrorTypeTimeout,
		ErrorTypeDatabase,
		ErrorTypeStorage,
		ErrorTypeConfiguration,
		ErrorTypeBusinessLogic,
	}

	for _, validType := range validTypes {
		if errorType == validType {
			return true
		}
	}

	return false
}

// isValidErrorSeverity проверяет корректность серьезности ошибки
func (ev *ErrorValidator) isValidErrorSeverity(severity ErrorSeverity) bool {
	validSeverities := []ErrorSeverity{
		ErrorSeverityLow,
		ErrorSeverityMedium,
		ErrorSeverityHigh,
		ErrorSeverityCritical,
	}

	for _, validSeverity := range validSeverities {
		if severity == validSeverity {
			return true
		}
	}

	return false
}

// isValidHTTPStatus проверяет корректность HTTP статуса
func (ev *ErrorValidator) isValidHTTPStatus(status int) bool {
	return status >= 100 && status <= 599
}

// isValidRequestID проверяет формат RequestID
func (ev *ErrorValidator) isValidRequestID(requestID string) bool {
	if requestID == "" {
		return false
	}

	// RequestID должен содержать только буквы, цифры и дефисы
	for _, char := range requestID {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	return true
}

// isValidOperation проверяет формат операции
func (ev *ErrorValidator) isValidOperation(operation string) bool {
	if operation == "" {
		return false
	}

	// Операция должна быть в snake_case
	for _, char := range operation {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// isValidComponent проверяет формат компонента
func (ev *ErrorValidator) isValidComponent(component string) bool {
	if component == "" {
		return false
	}

	// Компонент должен быть в snake_case
	for _, char := range component {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// isValidHTTPMethod проверяет корректность HTTP метода
func (ev *ErrorValidator) isValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}

	return false
}

// isValidPath проверяет формат пути
func (ev *ErrorValidator) isValidPath(path string) bool {
	if path == "" {
		return false
	}

	// Путь должен начинаться с /
	if !strings.HasPrefix(path, "/") {
		return false
	}

	// Путь должен содержать только допустимые символы
	for _, char := range path {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') ||
			char == '/' || char == '-' || char == '_' || char == '.' || char == '?') {
			return false
		}
	}

	return true
}

// isValidMetadataKey проверяет формат ключа метаданных
func (ev *ErrorValidator) isValidMetadataKey(key string) bool {
	if key == "" {
		return false
	}

	// Ключ должен быть в snake_case
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}

	return true
}

// isValidMetadataValue проверяет тип значения метаданных
func (ev *ErrorValidator) isValidMetadataValue(value interface{}) bool {
	switch value.(type) {
	case string, int, int64, float64, bool, map[string]interface{}:
		return true
	default:
		return false
	}
}

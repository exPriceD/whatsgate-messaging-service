package errors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrorValidator_ValidateAppError(t *testing.T) {
	validator := NewErrorValidator()

	// Тестируем корректную ошибку
	validErr := New(ErrorTypeValidation, "TEST_ERROR", "Test error", nil)
	err := validator.ValidateAppError(validErr)
	assert.NoError(t, err)

	// Тестируем nil ошибку
	err = validator.ValidateAppError(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Тестируем ошибку без типа
	invalidErr := &AppError{
		Code:    "TEST_ERROR",
		Message: "Test error",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Type is required")

	// Тестируем ошибку без кода
	invalidErr = &AppError{
		Type:    ErrorTypeValidation,
		Message: "Test error",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Code is required")

	// Тестируем ошибку без сообщения
	invalidErr = &AppError{
		Type: ErrorTypeValidation,
		Code: "TEST_ERROR",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Message is required")

	// Тестируем некорректный код ошибки
	invalidErr = &AppError{
		Type:    ErrorTypeValidation,
		Code:    "test-error",
		Message: "Test error",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UPPER_SNAKE_CASE")

	// Тестируем некорректный тип ошибки
	invalidErr = &AppError{
		Type:    "INVALID_TYPE",
		Code:    "TEST_ERROR",
		Message: "Test error",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid error type")

	// Тестируем некорректную серьезность
	invalidErr = &AppError{
		Type:      ErrorTypeValidation,
		Code:      "TEST_ERROR",
		Message:   "Test error",
		Severity:  "INVALID_SEVERITY",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid error severity")

	// Тестируем некорректный HTTP статус
	invalidErr = &AppError{
		Type:       ErrorTypeValidation,
		Code:       "TEST_ERROR",
		Message:    "Test error",
		HTTPStatus: 999,
		Timestamp:  time.Now(),
		Version:    "1.0.0",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid HTTP status")

	// Тестируем ошибку без временной метки
	invalidErr = &AppError{
		Type:    ErrorTypeValidation,
		Code:    "TEST_ERROR",
		Message: "Test error",
		Version: "1.0.0",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Timestamp is required")

	// Тестируем ошибку с будущей временной меткой
	invalidErr = &AppError{
		Type:      ErrorTypeValidation,
		Code:      "TEST_ERROR",
		Message:   "Test error",
		Timestamp: time.Now().Add(time.Hour),
		Version:   "1.0.0",
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be in the future")

	// Тестируем ошибку без версии
	invalidErr = &AppError{
		Type:      ErrorTypeValidation,
		Code:      "TEST_ERROR",
		Message:   "Test error",
		Timestamp: time.Now(),
	}
	err = validator.ValidateAppError(invalidErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Version is required")
}

func TestErrorValidator_ValidateErrorContext(t *testing.T) {
	validator := NewErrorValidator()

	// Тестируем корректный контекст
	validContext := &ErrorContext{
		RequestID: "req-123",
		Operation: "test_operation",
		Component: "test_component",
		Method:    "GET",
		Path:      "/api/test",
	}
	err := validator.validateErrorContext(validContext)
	assert.NoError(t, err)

	// Тестируем некорректный RequestID
	invalidContext := &ErrorContext{
		RequestID: "req@123",
	}
	err = validator.validateErrorContext(invalidContext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid RequestID")

	// Тестируем некорректную операцию
	invalidContext = &ErrorContext{
		Operation: "test-operation",
	}
	err = validator.validateErrorContext(invalidContext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Operation")

	// Тестируем некорректный компонент
	invalidContext = &ErrorContext{
		Component: "test-component",
	}
	err = validator.validateErrorContext(invalidContext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Component")

	// Тестируем некорректный HTTP метод
	invalidContext = &ErrorContext{
		Method: "INVALID",
	}
	err = validator.validateErrorContext(invalidContext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid HTTP method")

	// Тестируем некорректный путь
	invalidContext = &ErrorContext{
		Path: "invalid path",
	}
	err = validator.validateErrorContext(invalidContext)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Path")
}

func TestErrorValidator_ValidateMetadata(t *testing.T) {
	validator := NewErrorValidator()

	// Тестируем корректные метаданные
	validMetadata := map[string]interface{}{
		"user_id":   123,
		"operation": "test",
		"nested": map[string]interface{}{
			"key": "value",
		},
	}
	err := validator.validateMetadata(validMetadata)
	assert.NoError(t, err)

	// Тестируем пустой ключ
	invalidMetadata := map[string]interface{}{
		"": "value",
	}
	err = validator.validateMetadata(invalidMetadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Тестируем некорректный формат ключа
	invalidMetadata = map[string]interface{}{
		"invalid-key": "value",
	}
	err = validator.validateMetadata(invalidMetadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid metadata key format")

	// Тестируем nil значение
	invalidMetadata = map[string]interface{}{
		"key": nil,
	}
	err = validator.validateMetadata(invalidMetadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")

	// Тестируем некорректный тип значения
	invalidMetadata = map[string]interface{}{
		"key": make(chan int),
	}
	err = validator.validateMetadata(invalidMetadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid metadata value type")
}

func TestErrorValidator_HelperFunctions(t *testing.T) {
	validator := NewErrorValidator()

	// Тестируем isValidErrorCode
	assert.True(t, validator.isValidErrorCode("TEST_ERROR"))
	assert.True(t, validator.isValidErrorCode("ERROR_123"))
	assert.False(t, validator.isValidErrorCode("test-error"))
	assert.False(t, validator.isValidErrorCode(""))

	// Тестируем isValidErrorType
	assert.True(t, validator.isValidErrorType(ErrorTypeValidation))
	assert.True(t, validator.isValidErrorType(ErrorTypeInternal))
	assert.False(t, validator.isValidErrorType("INVALID_TYPE"))

	// Тестируем isValidErrorSeverity
	assert.True(t, validator.isValidErrorSeverity(ErrorSeverityLow))
	assert.True(t, validator.isValidErrorSeverity(ErrorSeverityCritical))
	assert.False(t, validator.isValidErrorSeverity("INVALID_SEVERITY"))

	// Тестируем isValidHTTPStatus
	assert.True(t, validator.isValidHTTPStatus(200))
	assert.True(t, validator.isValidHTTPStatus(404))
	assert.True(t, validator.isValidHTTPStatus(500))
	assert.False(t, validator.isValidHTTPStatus(999))
	assert.False(t, validator.isValidHTTPStatus(0))

	// Тестируем isValidRequestID
	assert.True(t, validator.isValidRequestID("req-123"))
	assert.True(t, validator.isValidRequestID("abc123"))
	assert.False(t, validator.isValidRequestID("req@123"))
	assert.False(t, validator.isValidRequestID(""))

	// Тестируем isValidOperation
	assert.True(t, validator.isValidOperation("test_operation"))
	assert.True(t, validator.isValidOperation("op123"))
	assert.False(t, validator.isValidOperation("test-operation"))
	assert.False(t, validator.isValidOperation(""))

	// Тестируем isValidComponent
	assert.True(t, validator.isValidComponent("test_component"))
	assert.True(t, validator.isValidComponent("comp123"))
	assert.False(t, validator.isValidComponent("test-component"))
	assert.False(t, validator.isValidComponent(""))

	// Тестируем isValidHTTPMethod
	assert.True(t, validator.isValidHTTPMethod("GET"))
	assert.True(t, validator.isValidHTTPMethod("POST"))
	assert.False(t, validator.isValidHTTPMethod("INVALID"))

	// Тестируем isValidPath
	assert.True(t, validator.isValidPath("/api/test"))
	assert.True(t, validator.isValidPath("/"))
	assert.False(t, validator.isValidPath("invalid path"))
	assert.False(t, validator.isValidPath(""))

	// Тестируем isValidMetadataKey
	assert.True(t, validator.isValidMetadataKey("test_key"))
	assert.True(t, validator.isValidMetadataKey("key123"))
	assert.False(t, validator.isValidMetadataKey("test-key"))
	assert.False(t, validator.isValidMetadataKey(""))

	// Тестируем isValidMetadataValue
	assert.True(t, validator.isValidMetadataValue("string"))
	assert.True(t, validator.isValidMetadataValue(123))
	assert.True(t, validator.isValidMetadataValue(true))
	assert.True(t, validator.isValidMetadataValue(map[string]interface{}{"key": "value"}))
	assert.False(t, validator.isValidMetadataValue(make(chan int)))
}

package errors

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("Invalid input")

	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.Equal(t, "Invalid input", err.Message)
	assert.Equal(t, ErrorSeverityMedium, err.Severity)
	assert.Equal(t, http.StatusBadRequest, err.HTTPStatus)
	assert.Equal(t, "1.0.0", err.Version)
	assert.NotNil(t, err.Stack)
	assert.NotZero(t, err.Timestamp)
}

func TestNewValidationErrorWithDescription(t *testing.T) {
	err := NewValidationError("Invalid phone", "Phone must be in international format")

	assert.Equal(t, "Invalid phone", err.Message)
	assert.Equal(t, "Phone must be in international format", err.Description)
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("user", "123")

	assert.Equal(t, ErrorTypeNotFound, err.Type)
	assert.Equal(t, "NOT_FOUND", err.Code)
	assert.Equal(t, "user with id '123' not found", err.Message)
	assert.Equal(t, ErrorSeverityLow, err.Severity)
	assert.Equal(t, http.StatusNotFound, err.HTTPStatus)
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("Invalid credentials")

	assert.Equal(t, ErrorTypeUnauthorized, err.Type)
	assert.Equal(t, "UNAUTHORIZED", err.Code)
	assert.Equal(t, "Invalid credentials", err.Message)
	assert.Equal(t, ErrorSeverityHigh, err.Severity)
	assert.Equal(t, http.StatusUnauthorized, err.HTTPStatus)
}

func TestNewInternalError(t *testing.T) {
	cause := errors.New("database connection failed")
	err := NewInternalError("Database operation failed", cause)

	assert.Equal(t, ErrorTypeInternal, err.Type)
	assert.Equal(t, "INTERNAL_ERROR", err.Code)
	assert.Equal(t, "Database operation failed", err.Message)
	assert.Equal(t, ErrorSeverityCritical, err.Severity)
	assert.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
	assert.Equal(t, cause, err.Cause)
}

func TestNewExternalServiceError(t *testing.T) {
	cause := errors.New("connection timeout")
	err := NewExternalServiceError("WhatsApp", "send_message", cause)

	assert.Equal(t, ErrorTypeExternalService, err.Type)
	assert.Equal(t, "EXTERNAL_SERVICE_ERROR", err.Code)
	assert.Equal(t, "External service 'WhatsApp' failed during 'send_message'", err.Message)
	assert.Equal(t, ErrorSeverityHigh, err.Severity)
	assert.Equal(t, http.StatusBadGateway, err.HTTPStatus)
	assert.Equal(t, cause, err.Cause)
}

func TestNewDatabaseError(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewDatabaseError("SELECT users", cause)

	assert.Equal(t, ErrorTypeDatabase, err.Type)
	assert.Equal(t, "DATABASE_ERROR", err.Code)
	assert.Equal(t, "Database operation 'SELECT users' failed", err.Message)
	assert.Equal(t, ErrorSeverityHigh, err.Severity)
	assert.Equal(t, http.StatusServiceUnavailable, err.HTTPStatus)
	assert.Equal(t, cause, err.Cause)
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("database query", 30*time.Second)

	assert.Equal(t, ErrorTypeTimeout, err.Type)
	assert.Equal(t, "TIMEOUT_ERROR", err.Code)
	assert.Equal(t, "Operation 'database query' timed out after 30s", err.Message)
	assert.Equal(t, ErrorSeverityMedium, err.Severity)
	assert.Equal(t, http.StatusGatewayTimeout, err.HTTPStatus)
}

func TestAppError_Error(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewInternalError("Test error", cause)

	errorStr := err.Error()
	assert.Contains(t, errorStr, "INTERNAL_ERROR")
	assert.Contains(t, errorStr, "Test error")
	assert.Contains(t, errorStr, "underlying error")
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := NewInternalError("Test error", cause)

	unwrapped := err.Unwrap()
	assert.Equal(t, cause, unwrapped)
}

func TestAppError_Is(t *testing.T) {
	err1 := NewValidationError("Error 1")
	err2 := NewValidationError("Error 2")
	err3 := NewInternalError("Error 3", nil)

	assert.True(t, err1.Is(err2))
	assert.False(t, err1.Is(err3))
	assert.False(t, err1.Is(nil))
}

func TestAppError_WithContext(t *testing.T) {
	err := NewValidationError("Test error")
	ctx := &ErrorContext{
		RequestID: "req-123",
		UserID:    "user-456",
		Operation: "test_operation",
		Component: "test_component",
		Method:    "GET",
		Path:      "/test",
		IP:        "192.168.1.1",
		UserAgent: "test-agent",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}

	result := err.WithContext(ctx)

	assert.Equal(t, "req-123", result.Context.RequestID)
	assert.Equal(t, "user-456", result.Context.UserID)
	assert.Equal(t, "test_operation", result.Context.Operation)
	assert.Equal(t, "test_component", result.Context.Component)
	assert.Equal(t, "GET", result.Context.Method)
	assert.Equal(t, "/test", result.Context.Path)
	assert.Equal(t, "192.168.1.1", result.Context.IP)
	assert.Equal(t, "test-agent", result.Context.UserAgent)
	assert.Equal(t, "value1", result.Context.Metadata["key1"])
	assert.Equal(t, "value2", result.Context.Metadata["key2"])
}

func TestAppError_WithMetadata(t *testing.T) {
	err := NewValidationError("Test error")

	result := err.WithMetadata("key1", "value1").
		WithMetadata("key2", "value2").
		WithMetadata("number", 42)

	assert.Equal(t, "value1", result.Context.Metadata["key1"])
	assert.Equal(t, "value2", result.Context.Metadata["key2"])
	assert.Equal(t, 42, result.Context.Metadata["number"])
}

func TestAppError_ToZapFields(t *testing.T) {
	err := NewValidationError("Test error").
		WithMetadata("test_key", "test_value")

	fields := err.ToZapFields()

	// Проверяем наличие основных полей
	fieldNames := make(map[string]bool)
	for _, field := range fields {
		fieldNames[field.Key] = true
	}

	assert.True(t, fieldNames["error_type"])
	assert.True(t, fieldNames["error_code"])
	assert.True(t, fieldNames["error_message"])
	assert.True(t, fieldNames["error_severity"])
	assert.True(t, fieldNames["error_timestamp"])
	assert.True(t, fieldNames["error_version"])
	assert.True(t, fieldNames["http_status"])
	assert.True(t, fieldNames["metadata"])
}

func TestAppError_ToJSON(t *testing.T) {
	appErr := NewValidationError("Test error").
		WithMetadata("test_key", "test_value")

	jsonData, jsonErr := appErr.ToJSON()
	require.NoError(t, jsonErr)

	assert.Contains(t, string(jsonData), "VALIDATION_ERROR")
	assert.Contains(t, string(jsonData), "Test error")
	assert.Contains(t, string(jsonData), "test_key")
	assert.Contains(t, string(jsonData), "test_value")
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr, ErrorTypeExternalService, "WRAPPED_ERROR", "Wrapped message")

	assert.Equal(t, ErrorTypeExternalService, wrapped.Type)
	assert.Equal(t, "WRAPPED_ERROR", wrapped.Code)
	assert.Equal(t, "Wrapped message", wrapped.Message)
	assert.Equal(t, originalErr, wrapped.Cause)
}

func TestWrap_WithAppError(t *testing.T) {
	originalErr := NewValidationError("Original error")
	wrapped := Wrap(originalErr, ErrorTypeInternal, "WRAPPED_ERROR", "Wrapped message")

	assert.Equal(t, ErrorTypeInternal, wrapped.Type)
	assert.Equal(t, "WRAPPED_ERROR", wrapped.Code)
	assert.Equal(t, "Wrapped message", wrapped.Message)
	assert.Equal(t, ErrorSeverityCritical, wrapped.Severity) // Изменилась серьезность
}

func TestWrap_NilError(t *testing.T) {
	result := Wrap(nil, ErrorTypeValidation, "TEST", "Test")
	assert.Nil(t, result)
}

func TestFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")
	ctx = context.WithValue(ctx, "session_id", "session-789")

	errorCtx := FromContext(ctx)

	assert.Equal(t, "req-123", errorCtx.RequestID)
	assert.Equal(t, "user-456", errorCtx.UserID)
	assert.Equal(t, "session-789", errorCtx.SessionID)
}

func TestFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	errorCtx := FromContext(ctx)

	assert.Empty(t, errorCtx.RequestID)
	assert.Empty(t, errorCtx.UserID)
	assert.Empty(t, errorCtx.SessionID)
}

func TestGetDefaultSeverity(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  ErrorSeverity
	}{
		{ErrorTypeValidation, ErrorSeverityLow},
		{ErrorTypeNotFound, ErrorSeverityLow},
		{ErrorTypeUnauthorized, ErrorSeverityMedium},
		{ErrorTypeForbidden, ErrorSeverityMedium},
		{ErrorTypeConflict, ErrorSeverityMedium},
		{ErrorTypeTooManyRequests, ErrorSeverityMedium},
		{ErrorTypeExternalService, ErrorSeverityHigh},
		{ErrorTypeNetwork, ErrorSeverityHigh},
		{ErrorTypeTimeout, ErrorSeverityHigh},
		{ErrorTypeDatabase, ErrorSeverityHigh},
		{ErrorTypeStorage, ErrorSeverityHigh},
		{ErrorTypeInternal, ErrorSeverityCritical},
		{ErrorTypeServiceUnavailable, ErrorSeverityCritical},
		{ErrorTypeGatewayTimeout, ErrorSeverityCritical},
	}

	for _, test := range tests {
		t.Run(string(test.errorType), func(t *testing.T) {
			result := getDefaultSeverity(test.errorType)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetDefaultHTTPStatus(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  int
	}{
		{ErrorTypeValidation, http.StatusBadRequest},
		{ErrorTypeUnauthorized, http.StatusUnauthorized},
		{ErrorTypeForbidden, http.StatusForbidden},
		{ErrorTypeNotFound, http.StatusNotFound},
		{ErrorTypeConflict, http.StatusConflict},
		{ErrorTypeTooManyRequests, http.StatusTooManyRequests},
		{ErrorTypeInternal, http.StatusInternalServerError},
		{ErrorTypeServiceUnavailable, http.StatusServiceUnavailable},
		{ErrorTypeGatewayTimeout, http.StatusGatewayTimeout},
		{ErrorTypeExternalService, http.StatusBadGateway},
		{ErrorTypeNetwork, http.StatusGatewayTimeout},
		{ErrorTypeTimeout, http.StatusGatewayTimeout},
		{ErrorTypeDatabase, http.StatusServiceUnavailable},
		{ErrorTypeStorage, http.StatusServiceUnavailable},
	}

	for _, test := range tests {
		t.Run(string(test.errorType), func(t *testing.T) {
			result := getDefaultHTTPStatus(test.errorType)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCaptureStack(t *testing.T) {
	stack := captureStack()

	assert.NotEmpty(t, stack)
	assert.Greater(t, len(stack), 0)

	// Проверяем структуру первого кадра
	firstFrame := stack[0]
	assert.NotEmpty(t, firstFrame.Function)
	assert.NotEmpty(t, firstFrame.File)
	assert.Greater(t, firstFrame.Line, 0)
}

func TestBusinessError(t *testing.T) {
	businessErr := NewBusinessError("INVALID_OPERATION", "Operation not allowed", false)

	assert.Equal(t, ErrorTypeBusinessLogic, businessErr.Type)
	assert.Equal(t, "BUSINESS_LOGIC_ERROR", businessErr.Code)
	assert.Equal(t, "Operation not allowed", businessErr.Message)
	assert.Equal(t, "INVALID_OPERATION", businessErr.BusinessCode)
	assert.False(t, businessErr.Retryable)
}

func TestWhatsAppErrors(t *testing.T) {
	// Test WhatsAppNotConfiguredError
	err := NewWhatsAppNotConfiguredError()
	assert.Equal(t, ErrorTypeConfiguration, err.Type)
	assert.Equal(t, "WHATSAPP_NOT_CONFIGURED", err.Code)
	assert.Contains(t, err.Message, "not configured")

	// Test WhatsAppInvalidPhoneError
	err = NewWhatsAppInvalidPhoneError("+123")
	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, "WHATSAPP_INVALID_PHONE", err.Code)
	assert.Contains(t, err.Message, "+123")

	// Test WhatsAppRateLimitError
	err = NewWhatsAppRateLimitError(5 * time.Minute)
	assert.Equal(t, ErrorTypeTooManyRequests, err.Type)
	assert.Equal(t, "WHATSAPP_RATE_LIMIT", err.Code)
	assert.Contains(t, err.Description, "5m")
}

func TestBulkErrors(t *testing.T) {
	// Test BulkCampaignNotFoundError
	err := NewBulkCampaignNotFoundError("campaign-123")
	assert.Equal(t, ErrorTypeNotFound, err.Type)
	assert.Equal(t, "BULK_CAMPAIGN_NOT_FOUND", err.Code)
	assert.Contains(t, err.Message, "campaign-123")

	// Test BulkFileParseError
	err = NewBulkFileParseError("excel", "invalid format")
	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, "BULK_FILE_PARSE_ERROR", err.Code)
	assert.Contains(t, err.Message, "excel")
	assert.Contains(t, err.Message, "invalid format")

	// Test BulkNoValidNumbersError
	err = NewBulkNoValidNumbersError()
	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, "BULK_NO_VALID_NUMBERS", err.Code)
	assert.Contains(t, err.Message, "valid phone numbers")
}

func TestDatabaseErrors(t *testing.T) {
	cause := errors.New("connection refused")

	// Test DatabaseConnectionError
	err := NewDatabaseConnectionError(cause)
	assert.Equal(t, ErrorTypeDatabase, err.Type)
	assert.Equal(t, "DATABASE_CONNECTION_ERROR", err.Code)
	assert.Equal(t, cause, err.Cause)

	// Test DatabaseQueryError
	err = NewDatabaseQueryError("SELECT users", cause)
	assert.Equal(t, ErrorTypeDatabase, err.Type)
	assert.Equal(t, "DATABASE_QUERY_ERROR", err.Code)
	assert.Contains(t, err.Message, "SELECT users")

	// Test DatabaseTimeoutError
	err = NewDatabaseTimeoutError("INSERT", 30*time.Second)
	assert.Equal(t, ErrorTypeDatabase, err.Type)
	assert.Equal(t, "DATABASE_TIMEOUT_ERROR", err.Code)
	assert.Contains(t, err.Message, "INSERT")
	assert.Contains(t, err.Message, "30s")
}

func TestErrorSystem_GlobalInstances(t *testing.T) {
	// Тестируем глобальные экземпляры
	InitErrorSystem()

	validator := GetErrorValidator()
	assert.NotNil(t, validator)

	utils := GetErrorUtils()
	assert.NotNil(t, utils)
}

func TestErrorSystem_NewWithValidation(t *testing.T) {
	InitErrorSystem()

	// Тестируем создание ошибки с валидацией
	err, validationErr := NewWithValidation(ErrorTypeValidation, "TEST_ERROR", "Test error", nil)
	assert.NoError(t, validationErr)
	assert.NotNil(t, err)
	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, "TEST_ERROR", err.Code)

	// Тестируем создание некорректной ошибки
	_, validationErr = NewWithValidation("", "", "", nil)
	assert.Error(t, validationErr)
	assert.Contains(t, validationErr.Error(), "validation failed")
}

func TestErrorSystem_NewValidationErrorWithValidation(t *testing.T) {
	InitErrorSystem()

	// Тестируем создание ошибки валидации с проверкой
	err, validationErr := NewValidationErrorWithValidation("Test validation error", "detail1", "detail2")
	assert.NoError(t, validationErr)
	assert.NotNil(t, err)
	assert.Equal(t, ErrorTypeValidation, err.Type)
	assert.Equal(t, "VALIDATION_ERROR", err.Code)
	assert.Contains(t, err.Message, "Test validation error")
}

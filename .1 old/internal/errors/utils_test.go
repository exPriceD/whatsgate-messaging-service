package errors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorUtils_IsAppError(t *testing.T) {
	utils := NewErrorUtils()

	// Test with AppError
	appErr := NewValidationError("Test error")
	assert.True(t, utils.IsAppError(appErr))

	// Test with regular error
	regularErr := errors.New("regular error")
	assert.False(t, utils.IsAppError(regularErr))

	// Test with nil
	assert.False(t, utils.IsAppError(nil))
}

func TestErrorUtils_GetAppError(t *testing.T) {
	utils := NewErrorUtils()

	// Test with AppError
	appErr := NewValidationError("Test error")
	result, ok := utils.GetAppError(appErr)
	assert.True(t, ok)
	assert.Equal(t, appErr, result)

	// Test with regular error
	regularErr := errors.New("regular error")
	result, ok = utils.GetAppError(regularErr)
	assert.False(t, ok)
	assert.Nil(t, result)

	// Test with nil
	result, ok = utils.GetAppError(nil)
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestErrorUtils_IsErrorType(t *testing.T) {
	utils := NewErrorUtils()

	appErr := NewValidationError("Test error")

	assert.True(t, utils.IsErrorType(appErr, ErrorTypeValidation))
	assert.False(t, utils.IsErrorType(appErr, ErrorTypeInternal))

	// Test with regular error
	regularErr := errors.New("regular error")
	assert.False(t, utils.IsErrorType(regularErr, ErrorTypeValidation))
}

func TestErrorUtils_IsErrorCode(t *testing.T) {
	utils := NewErrorUtils()

	appErr := NewValidationError("Test error")

	assert.True(t, utils.IsErrorCode(appErr, "VALIDATION_ERROR"))
	assert.False(t, utils.IsErrorCode(appErr, "INTERNAL_ERROR"))

	// Test with regular error
	regularErr := errors.New("regular error")
	assert.False(t, utils.IsErrorCode(regularErr, "VALIDATION_ERROR"))
}

func TestErrorUtils_IsSeverity(t *testing.T) {
	utils := NewErrorUtils()

	appErr := NewValidationError("Test error")

	assert.True(t, utils.IsSeverity(appErr, ErrorSeverityMedium))
	assert.False(t, utils.IsSeverity(appErr, ErrorSeverityCritical))

	// Test with regular error
	regularErr := errors.New("regular error")
	assert.False(t, utils.IsSeverity(regularErr, ErrorSeverityMedium))
}

func TestErrorUtils_WrapWithContext(t *testing.T) {
	utils := NewErrorUtils()

	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	originalErr := errors.New("original error")

	result := utils.WrapWithContext(originalErr, ctx, ErrorTypeExternalService, "TEST_ERROR", "Test message")

	assert.Equal(t, ErrorTypeExternalService, result.Type)
	assert.Equal(t, "TEST_ERROR", result.Code)
	assert.Equal(t, "Test message", result.Message)
	assert.Equal(t, originalErr, result.Cause)
	assert.Equal(t, "req-123", result.Context.RequestID)
}

func TestErrorUtils_ChainErrors(t *testing.T) {
	utils := NewErrorUtils()

	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err3 := errors.New("error 3")

	// Test with multiple errors
	chain := utils.ChainErrors(err1, err2, err3)
	assert.NotNil(t, chain)
	assert.Contains(t, chain.Error(), "error 1")
	assert.Contains(t, chain.Error(), "error 2")
	assert.Contains(t, chain.Error(), "error 3")

	// Test with single error
	chain = utils.ChainErrors(err1)
	assert.Equal(t, err1, chain)

	// Test with no errors
	chain = utils.ChainErrors()
	assert.Nil(t, chain)

	// Test with nil errors
	chain = utils.ChainErrors(nil, nil, nil)
	assert.Nil(t, chain)

	// Test with mixed nil and non-nil errors
	chain = utils.ChainErrors(nil, err1, nil, err2)
	assert.NotNil(t, chain)
	assert.Contains(t, chain.Error(), "error 1")
	assert.Contains(t, chain.Error(), "error 2")
}

func TestErrorUtils_ExtractErrorChain(t *testing.T) {
	utils := NewErrorUtils()

	// Создаем цепочку ошибок
	base := errors.New("error 1")
	wrapped := Wrap(base, ErrorTypeInternal, "ERR2", "error 2")
	wrapped2 := Wrap(wrapped, ErrorTypeInternal, "ERR3", "error 3")

	chain := utils.ExtractErrorChain(wrapped2)
	assert.GreaterOrEqual(t, len(chain), 2)
	assert.Contains(t, chain[0].Error(), "error 3")
	assert.Contains(t, chain[len(chain)-1].Error(), "error 1")
}

func TestErrorUtils_GetRootCause(t *testing.T) {
	utils := NewErrorUtils()

	err1 := errors.New("root cause")
	err2 := errors.New("middle error")
	err3 := errors.New("top error")

	// Create a chain
	chain := utils.ChainErrors(err1, err2, err3)

	rootCause := utils.GetRootCause(chain)
	assert.ErrorIs(t, rootCause, err1)

	// Test with single error
	rootCause = utils.GetRootCause(err1)
	assert.ErrorIs(t, rootCause, err1)

	// Test with nil
	rootCause = utils.GetRootCause(nil)
	assert.Nil(t, rootCause)
}

func TestErrorUtils_SanitizeError(t *testing.T) {
	utils := NewErrorUtils()

	appErr := NewValidationError("Test error with api_key=secret123").
		WithMetadata("password", "mypassword").
		WithMetadata("token", "jwt_token_here").
		WithMetadata("phone", "+1234567890")

	sanitized := utils.SanitizeError(appErr)

	sanitizedAppErr, ok := sanitized.(*AppError)
	assert.True(t, ok)

	// Check that sensitive data is masked in message
	assert.Contains(t, sanitizedAppErr.Message, "***")

	// Test with regular error
	regularErr := errors.New("regular error")
	sanitized = utils.SanitizeError(regularErr)
	assert.Equal(t, regularErr, sanitized)
}

func TestErrorUtils_SanitizeMessage(t *testing.T) {
	utils := NewErrorUtils()

	tests := []struct {
		input    string
		expected string
	}{
		{
			"Error with api_key=secret123",
			"Error with ***",
		},
		{
			"Error with token=jwt_token_here",
			"Error with ***",
		},
		{
			"Error with password=mypassword",
			"Error with ***",
		},
		{
			"Error with phone=+1234567890",
			"Error with phone=+***-***-7890",
		},
		{
			"Regular error message",
			"Regular error message",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := utils.sanitizeMessage(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestErrorUtils_MaskPattern(t *testing.T) {
	utils := NewErrorUtils()

	tests := []struct {
		text        string
		pattern     string
		replacement string
		expected    string
	}{
		{
			"api_key=secret123",
			`api[_-]?key["\s]*[:=]["\s]*["']?[a-zA-Z0-9_-]+["']?`,
			"***",
			"***",
		},
		{
			"token=jwt_token_here",
			`token["\s]*[:=]["\s]*["']?[a-zA-Z0-9_-]+["']?`,
			"***",
			"***",
		},
		{
			"password=mypassword",
			`password["\s]*[:=]["\s]*["']?[^"'\s]+["']?`,
			"***",
			"***",
		},
		{
			"phone=+1234567890",
			`(\+?[0-9]{1,3}[-\s]?)?([0-9]{3,4})[-\s]?([0-9]{3,4})[-\s]?([0-9]{4})`,
			"***-***-$4",
			"phone=+***-***-7890",
		},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			result := utils.maskPattern(test.text, test.pattern, test.replacement)
			assert.Equal(t, test.expected, result)
		})
	}
}

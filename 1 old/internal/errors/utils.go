package errors

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrorUtils предоставляет утилиты для работы с ошибками
type ErrorUtils struct{}

// NewErrorUtils создает новый экземпляр ErrorUtils
func NewErrorUtils() *ErrorUtils {
	return &ErrorUtils{}
}

// IsAppError проверяет, является ли ошибка AppError
func (eu *ErrorUtils) IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError извлекает AppError из ошибки
func (eu *ErrorUtils) GetAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// IsErrorType проверяет, является ли ошибка определенного типа
func (eu *ErrorUtils) IsErrorType(err error, errorType ErrorType) bool {
	if appErr, ok := eu.GetAppError(err); ok {
		return appErr.Type == errorType
	}
	return false
}

// IsErrorCode проверяет, имеет ли ошибка определенный код
func (eu *ErrorUtils) IsErrorCode(err error, code string) bool {
	if appErr, ok := eu.GetAppError(err); ok {
		return appErr.Code == code
	}
	return false
}

// IsSeverity проверяет, имеет ли ошибка определенную серьезность
func (eu *ErrorUtils) IsSeverity(err error, severity ErrorSeverity) bool {
	if appErr, ok := eu.GetAppError(err); ok {
		return appErr.Severity == severity
	}
	return false
}

// WrapWithContext оборачивает ошибку с контекстом
func (eu *ErrorUtils) WrapWithContext(err error, ctx context.Context, errorType ErrorType, code, message string) *AppError {
	appErr := Wrap(err, errorType, code, message)
	if appErr != nil {
		errorContext := FromContext(ctx)
		appErr = appErr.WithContext(errorContext)
	}
	return appErr
}

// ChainErrors создает цепочку ошибок
func (eu *ErrorUtils) ChainErrors(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	var result error
	for i := len(errs) - 1; i >= 0; i-- {
		if errs[i] != nil {
			if result == nil {
				result = errs[i]
			} else {
				result = fmt.Errorf("%w: %w", errs[i], result)
			}
		}
	}

	return result
}

// ExtractErrorChain извлекает цепочку ошибок
func (eu *ErrorUtils) ExtractErrorChain(err error) []error {
	var chain []error
	current := err

	for current != nil {
		chain = append(chain, current)
		current = errors.Unwrap(current)
	}

	return chain
}

// GetRootCause возвращает корневую причину ошибки
func (eu *ErrorUtils) GetRootCause(err error) error {
	if err == nil {
		return nil
	}

	root := err
	for {
		unwrapped := errors.Unwrap(root)
		if unwrapped == nil {
			break
		}
		root = unwrapped
	}

	return root
}

// SanitizeError очищает ошибку от чувствительной информации
func (eu *ErrorUtils) SanitizeError(err error) error {
	if appErr, ok := eu.GetAppError(err); ok {
		// Создаем копию ошибки без чувствительных данных
		sanitized := &AppError{
			Type:        appErr.Type,
			Code:        appErr.Code,
			Message:     eu.sanitizeMessage(appErr.Message),
			Description: appErr.Description,
			Severity:    appErr.Severity,
			Timestamp:   appErr.Timestamp,
			HTTPStatus:  appErr.HTTPStatus,
			Version:     appErr.Version,
		}

		// Очищаем контекст от чувствительных данных
		if appErr.Context != nil {
			sanitized.Context = &ErrorContext{
				RequestID: appErr.Context.RequestID,
				Operation: appErr.Context.Operation,
				Component: appErr.Context.Component,
				Method:    appErr.Context.Method,
				Path:      appErr.Context.Path,
				Metadata:  eu.sanitizeMetadata(appErr.Context.Metadata),
			}
		}

		return sanitized
	}

	return err
}

// sanitizeMessage очищает сообщение от чувствительной информации
func (eu *ErrorUtils) sanitizeMessage(message string) string {
	if message == "" {
		return message
	}

	// Маскируем API ключи
	message = eu.maskPattern(message, `api[_-]?key["\s]*[:=]["\s]*["']?[a-zA-Z0-9_-]+["']?`, "***")

	// Маскируем токены
	message = eu.maskPattern(message, `token["\s]*[:=]["\s]*["']?[a-zA-Z0-9_-]+["']?`, "***")

	// Маскируем пароли
	message = eu.maskPattern(message, `password["\s]*[:=]["\s]*["']?[^"'\s]+["']?`, "***")

	// Маскируем секреты
	message = eu.maskPattern(message, `secret["\s]*[:=]["\s]*["']?[a-zA-Z0-9_-]+["']?`, "***")

	// Маскируем номера телефонов (частично)
	message = eu.maskPattern(message, `(\+?[0-9]{1,3}[-\s]?)?([0-9]{3,4})[-\s]?([0-9]{3,4})[-\s]?([0-9]{4})`, "***-***-$4")

	return message
}

// sanitizeMetadata очищает метаданные от чувствительной информации
func (eu *ErrorUtils) sanitizeMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return nil
	}

	sanitized := make(map[string]interface{})
	sensitiveKeys := []string{"password", "token", "secret", "api_key", "apikey", "auth", "key"}

	for key, value := range metadata {
		keyLower := strings.ToLower(key)
		isSensitive := false

		// Проверяем, является ли ключ чувствительным
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(keyLower, sensitiveKey) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			sanitized[key] = "***"
		} else {
			// Рекурсивно очищаем вложенные структуры
			if nestedMap, ok := value.(map[string]interface{}); ok {
				sanitized[key] = eu.sanitizeMetadata(nestedMap)
			} else {
				sanitized[key] = value
			}
		}
	}

	return sanitized
}

// maskPattern маскирует паттерн в строке с использованием regexp
func (eu *ErrorUtils) maskPattern(text, pattern, replacement string) string {
	if text == "" || pattern == "" {
		return text
	}

	// Компилируем регулярное выражение
	re, err := regexp.Compile(pattern)
	if err != nil {
		// В случае ошибки компиляции возвращаем исходный текст
		return text
	}

	// Заменяем все совпадения
	return re.ReplaceAllString(text, replacement)
}

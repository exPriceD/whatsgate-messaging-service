package errors

import (
	"fmt"
	"net/http"
	"time"
)

// BusinessError представляет ошибку бизнес-логики
type BusinessError struct {
	*AppError
	BusinessCode string `json:"business_code,omitempty"`
	Retryable    bool   `json:"retryable,omitempty"`
}

// NewBusinessError создает новую ошибку бизнес-логики
func NewBusinessError(businessCode, message string, retryable bool) *BusinessError {
	return &BusinessError{
		AppError: &AppError{
			Type:       ErrorTypeBusinessLogic,
			Code:       "BUSINESS_LOGIC_ERROR",
			Message:    message,
			Severity:   ErrorSeverityMedium,
			Stack:      captureStack(),
			Timestamp:  time.Now(),
			HTTPStatus: http.StatusBadRequest,
			Version:    "1.0.0",
		},
		BusinessCode: businessCode,
		Retryable:    retryable,
	}
}

// WhatsAppErrors - специализированные ошибки для WhatsApp функциональности

// NewWhatsAppNotConfiguredError создает ошибку для не настроенного WhatsApp
func NewWhatsAppNotConfiguredError() *AppError {
	return &AppError{
		Type:        ErrorTypeConfiguration,
		Code:        "WHATSAPP_NOT_CONFIGURED",
		Message:     "WhatsApp service is not configured",
		Description: "Please configure WhatsApp settings before sending messages",
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// NewWhatsAppConnectionError создает ошибку подключения к WhatsApp
func NewWhatsAppConnectionError(cause error) *AppError {
	return &AppError{
		Type:        ErrorTypeExternalService,
		Code:        "WHATSAPP_CONNECTION_ERROR",
		Message:     "Failed to connect to WhatsApp service",
		Description: "Unable to establish connection with WhatsApp API",
		Severity:    ErrorSeverityHigh,
		Cause:       cause,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadGateway,
		Version:     "1.0.0",
	}
}

// NewWhatsAppRateLimitError создает ошибку превышения лимита запросов
func NewWhatsAppRateLimitError(retryAfter time.Duration) *AppError {
	return &AppError{
		Type:        ErrorTypeTooManyRequests,
		Code:        "WHATSAPP_RATE_LIMIT",
		Message:     "WhatsApp rate limit exceeded",
		Description: fmt.Sprintf("Rate limit exceeded. Retry after %v", retryAfter),
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusTooManyRequests,
		Version:     "1.0.0",
	}
}

// NewWhatsAppInvalidPhoneError создает ошибку неверного номера телефона
func NewWhatsAppInvalidPhoneError(phone string) *AppError {
	return &AppError{
		Type:        ErrorTypeValidation,
		Code:        "WHATSAPP_INVALID_PHONE",
		Message:     fmt.Sprintf("Invalid phone number: %s", phone),
		Description: "Phone number must be in international format (e.g., +1234567890)",
		Severity:    ErrorSeverityLow,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// NewWhatsAppMessageTooLongError создает ошибку слишком длинного сообщения
func NewWhatsAppMessageTooLongError(maxLength int) *AppError {
	return &AppError{
		Type:        ErrorTypeValidation,
		Code:        "WHATSAPP_MESSAGE_TOO_LONG",
		Message:     "Message is too long",
		Description: fmt.Sprintf("Message length exceeds maximum allowed length of %d characters", maxLength),
		Severity:    ErrorSeverityLow,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// NewWhatsAppMediaError создает ошибку медиа-файла
func NewWhatsAppMediaError(mediaType, reason string) *AppError {
	return &AppError{
		Type:        ErrorTypeValidation,
		Code:        "WHATSAPP_MEDIA_ERROR",
		Message:     fmt.Sprintf("Media error for type %s: %s", mediaType, reason),
		Description: "Media file is invalid or unsupported",
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// BulkErrors - специализированные ошибки для массовых рассылок

// NewBulkCampaignNotFoundError создает ошибку для не найденной кампании
func NewBulkCampaignNotFoundError(campaignID string) *AppError {
	return &AppError{
		Type:        ErrorTypeNotFound,
		Code:        "BULK_CAMPAIGN_NOT_FOUND",
		Message:     fmt.Sprintf("Bulk campaign with id '%s' not found", campaignID),
		Description: "The specified bulk campaign does not exist",
		Severity:    ErrorSeverityLow,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusNotFound,
		Version:     "1.0.0",
	}
}

// NewBulkCampaignAlreadyRunningError создает ошибку для уже запущенной кампании
func NewBulkCampaignAlreadyRunningError(campaignID string) *AppError {
	return &AppError{
		Type:        ErrorTypeConflict,
		Code:        "BULK_CAMPAIGN_ALREADY_RUNNING",
		Message:     fmt.Sprintf("Bulk campaign '%s' is already running", campaignID),
		Description: "Cannot start a campaign that is already in progress",
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusConflict,
		Version:     "1.0.0",
	}
}

// NewBulkCampaignCompletedError создает ошибку для завершенной кампании
func NewBulkCampaignCompletedError(campaignID string) *AppError {
	return &AppError{
		Type:        ErrorTypeBusinessLogic,
		Code:        "BULK_CAMPAIGN_COMPLETED",
		Message:     fmt.Sprintf("Bulk campaign '%s' is already completed", campaignID),
		Description: "Cannot modify a completed campaign",
		Severity:    ErrorSeverityLow,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// NewBulkFileParseError создает ошибку парсинга файла
func NewBulkFileParseError(fileType, reason string) *AppError {
	return &AppError{
		Type:        ErrorTypeValidation,
		Code:        "BULK_FILE_PARSE_ERROR",
		Message:     fmt.Sprintf("Failed to parse %s file: %s", fileType, reason),
		Description: "The uploaded file format is invalid or corrupted",
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// NewBulkNoValidNumbersError создает ошибку отсутствия валидных номеров
func NewBulkNoValidNumbersError() *AppError {
	return &AppError{
		Type:        ErrorTypeValidation,
		Code:        "BULK_NO_VALID_NUMBERS",
		Message:     "No valid phone numbers found in the file",
		Description: "The uploaded file must contain at least one valid phone number",
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// NewBulkRateLimitExceededError создает ошибку превышения лимита рассылок
func NewBulkRateLimitExceededError(maxPerHour int) *AppError {
	return &AppError{
		Type:        ErrorTypeTooManyRequests,
		Code:        "BULK_RATE_LIMIT_EXCEEDED",
		Message:     "Bulk sending rate limit exceeded",
		Description: fmt.Sprintf("Maximum %d messages per hour allowed", maxPerHour),
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusTooManyRequests,
		Version:     "1.0.0",
	}
}

// NewBulkFileTooLargeError создает ошибку слишком большого файла
func NewBulkFileTooLargeError(maxSizeMB int) *AppError {
	return &AppError{
		Type:        ErrorTypeValidation,
		Code:        "BULK_FILE_TOO_LARGE",
		Message:     "Uploaded file is too large",
		Description: fmt.Sprintf("File size exceeds maximum allowed size of %d MB", maxSizeMB),
		Severity:    ErrorSeverityMedium,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusBadRequest,
		Version:     "1.0.0",
	}
}

// DatabaseErrors - специализированные ошибки для базы данных

// NewDatabaseConnectionError создает ошибку подключения к БД
func NewDatabaseConnectionError(cause error) *AppError {
	return &AppError{
		Type:        ErrorTypeDatabase,
		Code:        "DATABASE_CONNECTION_ERROR",
		Message:     "Failed to connect to database",
		Description: "Unable to establish connection with the database",
		Severity:    ErrorSeverityHigh,
		Cause:       cause,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusServiceUnavailable,
		Version:     "1.0.0",
	}
}

// NewDatabaseQueryError создает ошибку запроса к БД
func NewDatabaseQueryError(operation string, cause error) *AppError {
	return &AppError{
		Type:        ErrorTypeDatabase,
		Code:        "DATABASE_QUERY_ERROR",
		Message:     fmt.Sprintf("Database query failed: %s", operation),
		Description: "Failed to execute database operation",
		Severity:    ErrorSeverityHigh,
		Cause:       cause,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusServiceUnavailable,
		Version:     "1.0.0",
	}
}

// NewDatabaseTimeoutError создает ошибку таймаута БД
func NewDatabaseTimeoutError(operation string, timeout time.Duration) *AppError {
	return &AppError{
		Type:        ErrorTypeDatabase,
		Code:        "DATABASE_TIMEOUT_ERROR",
		Message:     fmt.Sprintf("Database operation '%s' timed out after %v", operation, timeout),
		Description: "Database operation exceeded the maximum allowed time",
		Severity:    ErrorSeverityHigh,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusServiceUnavailable,
		Version:     "1.0.0",
	}
}

// ConfigurationErrors - специализированные ошибки конфигурации

// NewConfigurationMissingError создает ошибку отсутствующей конфигурации
func NewConfigurationMissingError(configKey string) *AppError {
	return &AppError{
		Type:        ErrorTypeConfiguration,
		Code:        "CONFIGURATION_MISSING",
		Message:     fmt.Sprintf("Required configuration '%s' is missing", configKey),
		Description: "Application configuration is incomplete",
		Severity:    ErrorSeverityCritical,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusInternalServerError,
		Version:     "1.0.0",
	}
}

// NewConfigurationInvalidError создает ошибку неверной конфигурации
func NewConfigurationInvalidError(configKey, reason string) *AppError {
	return &AppError{
		Type:        ErrorTypeConfiguration,
		Code:        "CONFIGURATION_INVALID",
		Message:     fmt.Sprintf("Invalid configuration '%s': %s", configKey, reason),
		Description: "Application configuration contains invalid values",
		Severity:    ErrorSeverityCritical,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusInternalServerError,
		Version:     "1.0.0",
	}
}

// StorageErrors - специализированные ошибки хранилища

// NewStorageConnectionError создает ошибку подключения к хранилищу
func NewStorageConnectionError(storageType string, cause error) *AppError {
	return &AppError{
		Type:        ErrorTypeStorage,
		Code:        "STORAGE_CONNECTION_ERROR",
		Message:     fmt.Sprintf("Failed to connect to %s storage", storageType),
		Description: "Unable to establish connection with the storage service",
		Severity:    ErrorSeverityHigh,
		Cause:       cause,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusServiceUnavailable,
		Version:     "1.0.0",
	}
}

// NewStorageOperationError создает ошибку операции с хранилищем
func NewStorageOperationError(operation, storageType string, cause error) *AppError {
	return &AppError{
		Type:        ErrorTypeStorage,
		Code:        "STORAGE_OPERATION_ERROR",
		Message:     fmt.Sprintf("Storage operation '%s' failed for %s", operation, storageType),
		Description: "Failed to perform storage operation",
		Severity:    ErrorSeverityHigh,
		Cause:       cause,
		Stack:       captureStack(),
		Timestamp:   time.Now(),
		HTTPStatus:  http.StatusServiceUnavailable,
		Version:     "1.0.0",
	}
}

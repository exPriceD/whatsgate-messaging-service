package types

import "time"

// --- Common ---
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// AppErrorResponse представляет структурированную ошибку приложения для API
type AppErrorResponse struct {
	// Основная информация
	Type        string `json:"type" example:"VALIDATION_ERROR"`
	Code        string `json:"code" example:"INVALID_PHONE"`
	Message     string `json:"message" example:"Invalid phone number format"`
	Description string `json:"description,omitempty" example:"Phone number must be exactly 11 digits and start with 7"`

	// Метаданные
	Severity string `json:"severity" example:"MEDIUM"`

	// Контекстная информация
	Context *ErrorContext `json:"context,omitempty"`

	// Техническая информация
	Stack     []StackFrame `json:"stack,omitempty"`
	Timestamp time.Time    `json:"timestamp" example:"2023-01-01T12:00:00Z"`

	// HTTP информация
	HTTPStatus int `json:"http_status,omitempty" example:"400"`

	// Версионирование
	Version string `json:"version" example:"1.0.0"`
}

// ClientErrorResponse представляет упрощенную ошибку для клиента
// без технических деталей и чувствительной информации
type ClientErrorResponse struct {
	// Основная информация для пользователя
	Message     string `json:"message" example:"Invalid phone number format"`
	Description string `json:"description,omitempty" example:"Phone number must be exactly 11 digits and start with 7"`

	// Код ошибки для клиентской логики (без технических деталей)
	Code string `json:"code" example:"INVALID_PHONE"`

	// HTTP статус
	HTTPStatus int `json:"http_status,omitempty" example:"400"`

	// Временная метка
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T12:00:00Z"`
}

// ErrorContext содержит контекстную информацию об ошибке
type ErrorContext struct {
	RequestID  string                 `json:"request_id,omitempty" example:"req-123"`
	UserID     string                 `json:"user_id,omitempty" example:"user-456"`
	SessionID  string                 `json:"session_id,omitempty" example:"sess-789"`
	ResourceID string                 `json:"resource_id,omitempty" example:"campaign-123"`
	Operation  string                 `json:"operation,omitempty" example:"send_message"`
	Component  string                 `json:"component,omitempty" example:"send_message_handler"`
	Method     string                 `json:"method,omitempty" example:"POST"`
	Path       string                 `json:"path,omitempty" example:"/messages/send"`
	IP         string                 `json:"ip,omitempty" example:"192.168.1.1"`
	UserAgent  string                 `json:"user_agent,omitempty" example:"Mozilla/5.0..."`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// StackFrame представляет кадр стека вызовов
type StackFrame struct {
	Function string `json:"function" example:"main.main"`
	File     string `json:"file" example:"main.go"`
	Line     int    `json:"line" example:"42"`
}

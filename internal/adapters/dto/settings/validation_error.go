package settings

// ValidationErrorResponse представляет детальную ошибку валидации
type ValidationErrorResponse struct {
	Message string                 `json:"message"`
	Errors  []FieldValidationError `json:"errors"`
}

// FieldValidationError представляет ошибку валидации конкретного поля
type FieldValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

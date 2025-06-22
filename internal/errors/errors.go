package appErr

import "fmt"

// AppError — базовая структура для ошибок приложения.
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New создает новую AppError.
func New(code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

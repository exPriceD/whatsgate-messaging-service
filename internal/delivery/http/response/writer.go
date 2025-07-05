package response

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse представляет стандартный формат ошибки
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse представляет стандартный формат успешного ответа
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// HealthResponse представляет ответ health check
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// WriteJSON отправляет JSON ответ
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Fallback в случае ошибки кодирования
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}
}

// WriteError отправляет стандартную ошибку
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSON(w, statusCode, ErrorResponse{Error: message})
}

// WriteErrorWithCode отправляет ошибку с кодом
func WriteErrorWithCode(w http.ResponseWriter, statusCode int, message, code string) {
	WriteJSON(w, statusCode, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// WriteSuccess отправляет успешный ответ
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, SuccessResponse{Data: data})
}

// WriteSuccessWithMessage отправляет успешный ответ с сообщением
func WriteSuccessWithMessage(w http.ResponseWriter, data interface{}, message string) {
	WriteJSON(w, http.StatusOK, SuccessResponse{
		Data:    data,
		Message: message,
	})
}

// WriteMethodNotAllowed отправляет ошибку 405
func WriteMethodNotAllowed(w http.ResponseWriter) {
	WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// WriteNotFound отправляет ошибку 404
func WriteNotFound(w http.ResponseWriter) {
	WriteError(w, http.StatusNotFound, "Endpoint not found")
}

// WriteHealth отправляет health check ответ
func WriteHealth(w http.ResponseWriter) {
	WriteJSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: "whatsapp-service",
	})
}

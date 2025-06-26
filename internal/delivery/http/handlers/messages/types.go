package messages

// SendMessageRequest представляет запрос на отправку текстового сообщения.
type SendMessageRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required" example:"79991234567"`
	Message     string `json:"message" binding:"required" example:"Привет! Это тестовое сообщение"`
	Async       bool   `json:"async" example:"true"`
}

// SendMediaMessageRequest представляет запрос на отправку медиа-сообщения.
type SendMediaMessageRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required" example:"79991234567"`
	Message     string `json:"message" example:"Сообщение с медиа"`
	MessageType string `json:"message_type" binding:"required" example:"image"`
	Filename    string `json:"filename" binding:"required" example:"image.png"`
	MimeType    string `json:"mime_type" binding:"required" example:"image/png"`
	FileData    string `json:"file_data" binding:"required" example:"base64_encoded_data"`
	Async       bool   `json:"async" example:"true"`
}

// SendMessageResponse представляет ответ на отправку сообщения.
type SendMessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Message sent successfully"`
	ID      string `json:"id,omitempty" example:"message_id_123"`
	Status  string `json:"status,omitempty" example:"sent"`
}

// BulkSendRequest представляет запрос на массовую рассылку сообщений.
type BulkSendRequest struct {
	PhoneNumbers []string       `json:"phone_numbers" binding:"required" example:"79991234567,79998765432"`
	Message      string         `json:"message" binding:"required" example:"Массовое сообщение"`
	Async        bool           `json:"async" example:"true"`
	Media        *BulkSendMedia `json:"media,omitempty"`
}

// BulkSendMedia представляет медиа-файл для массовой рассылки.
type BulkSendMedia struct {
	MessageType string `json:"message_type" binding:"required" example:"image"`
	Filename    string `json:"filename" binding:"required" example:"image.png"`
	MimeType    string `json:"mime_type" binding:"required" example:"image/png"`
	FileData    string `json:"file_data" binding:"required" example:"base64_encoded_data"`
}

// BulkSendResult представляет результат отправки сообщения на один номер.
type BulkSendResult struct {
	PhoneNumber string `json:"phone_number" example:"79991234567"`
	Success     bool   `json:"success" example:"true"`
	MessageID   string `json:"message_id,omitempty" example:"message_id_123"`
	Status      string `json:"status,omitempty" example:"sent"`
	Error       string `json:"error,omitempty" example:"Invalid phone number"`
}

// BulkSendResponse представляет ответ на массовую рассылку.
type BulkSendResponse struct {
	Success      bool             `json:"success" example:"true"`
	TotalCount   int              `json:"total_count" example:"10"`
	SuccessCount int              `json:"success_count" example:"8"`
	FailedCount  int              `json:"failed_count" example:"2"`
	Results      []BulkSendResult `json:"results"`
}

// ErrorResponse представляет стандартный ответ об ошибке для всех эндпоинтов.
type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Error   string `json:"error" example:"Validation error"`
	Message string `json:"message" example:"Подробное описание ошибки"`
}

// BulkSendStartResponse — быстрый ответ после запуска массовой рассылки
type BulkSendStartResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Bulk send started in background"`
	Total   int    `json:"total" example:"123"`
}

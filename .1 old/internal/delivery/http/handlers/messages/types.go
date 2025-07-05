package messages

// Этот файл больше не используется, так как все структуры перенесены в types пакет
// Оставляем пустой файл для совместимости

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

// BulkSendMultipartRequest представляет запрос на массовую рассылку через multipart/form-data
type BulkSendMultipartRequest struct {
	Name            string `form:"name" binding:"required" example:"Летняя акция"`
	Message         string `form:"message" binding:"required" example:"Текст сообщения"`
	MessagesPerHour int    `form:"messages_per_hour" binding:"required,min=1" example:"20"`
	Async           bool   `form:"async" example:"false"`
	// Новые поля для дополнительных и исключаемых номеров
	AdditionalNumbers string `form:"additional_numbers" example:"+7(123)456-78-90\n+7(987)654-32-10"`
	ExcludeNumbers    string `form:"exclude_numbers" example:"+7(555)123-45-67\n+7(777)888-99-00"`
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

// BulkSendStartResponse — быстрый ответ после запуска массовой рассылки
type BulkSendStartResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"Bulk send started in background"`
	Total   int    `json:"total" example:"123"`
}

// BulkCampaignResponse — информация о рассылке для истории
type BulkCampaignResponse struct {
	ID              string  `json:"id" example:"uuid"`
	Name            string  `json:"name" example:"Летняя акция"`
	CreatedAt       string  `json:"created_at" example:"2023-01-01T12:00:00Z"`
	Message         string  `json:"message" example:"Текст сообщения"`
	Total           int     `json:"total" example:"100"`
	ProcessedCount  int     `json:"processed_count" example:"45"`
	Status          string  `json:"status" example:"started"`
	MediaFilename   *string `json:"media_filename,omitempty" example:"image.jpg"`
	MediaMime       *string `json:"media_mime,omitempty" example:"image/jpeg"`
	MediaType       *string `json:"media_type,omitempty" example:"image"`
	MessagesPerHour int     `json:"messages_per_hour" example:"20"`
	Initiator       *string `json:"initiator,omitempty" example:"user@example.com"`
	ErrorCount      int     `json:"error_count" example:"3"`
}

// SentNumbersResponse — список отправленных номеров для рассылки
type SentNumbersResponse struct {
	CampaignID  string   `json:"campaign_id" example:"uuid"`
	SentNumbers []string `json:"sent_numbers" example:"79991234567,79998765432"`
	TotalSent   int      `json:"total_sent" example:"45"`
}

// CountFileRowsResponse представляет ответ на запрос подсчета строк в файле
type CountFileRowsResponse struct {
	Success bool `json:"success" example:"true"`
	Rows    int  `json:"rows" example:"150"`
}

// CampaignError представляет ошибку для конкретного номера телефона
type CampaignError struct {
	PhoneNumber string `json:"phone_number" example:"79991234567"`
	Error       string `json:"error" example:"Invalid phone number format"`
}

// GetBulkCampaignErrorsResponse представляет ответ с ошибками кампании
type GetBulkCampaignErrorsResponse struct {
	Success bool            `json:"success" example:"true"`
	Errors  []CampaignError `json:"errors"`
	Total   int             `json:"total" example:"5"`
}

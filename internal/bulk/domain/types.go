package domain

// BulkSendParams содержит параметры для bulk-рассылки
// NumbersFile и MediaFile — абстракции для файлов (например, *multipart.FileHeader)
type BulkSendParams struct {
	Name              string
	Message           string
	Async             bool
	MessagesPerHour   int
	NumbersFile       any
	MediaFile         any
	AdditionalNumbers []string // Дополнительные номера к файлу
	ExcludeNumbers    []string // Номера для исключения из файла
}

// BulkMedia описывает медиа-файл для рассылки
// FileData — base64
// MessageType: image, video, audio, document
// MimeType: MIME type файла
// Filename: имя файла
type BulkMedia struct {
	MessageType string
	Filename    string
	MimeType    string
	FileData    string
}

// BulkSendResult — результат bulk-рассылки
type BulkSendResult struct {
	Started bool
	Message string
	Total   int
	Results []SingleSendResult
}

// SingleSendResult — результат отправки одного сообщения
type SingleSendResult struct {
	PhoneNumber string
	Success     bool
	Status      string
	Error       string
}

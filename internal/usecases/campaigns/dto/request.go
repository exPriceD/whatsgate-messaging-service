package dto

import "mime/multipart"

// CreateCampaignRequest представляет запрос на создание кампании
type CreateCampaignRequest struct {
	Name              string                // Название кампании
	Message           string                // Текст сообщения
	PhoneFile         *multipart.FileHeader // Excel файл с номерами
	MediaFile         *multipart.FileHeader // Медиа-файл (опционально)
	AdditionalNumbers []string              // Дополнительные номера
	ExcludeNumbers    []string              // Номера для исключения
	MessagesPerHour   int                   // Лимит сообщений в час
	Initiator         string                // Инициатор кампании
	Async             bool                  // Асинхронное выполнение
}

// StartCampaignRequest представляет запрос на запуск кампании
type StartCampaignRequest struct {
	CampaignID string // ID кампании для запуска
}

// CancelCampaignRequest представляет запрос на отмену кампании
type CancelCampaignRequest struct {
	CampaignID string // ID кампании для отмены
	Reason     string // Причина отмены (опционально)
}

// GetCampaignByIDRequest представляет запрос на получения кампании по ID
type GetCampaignByIDRequest struct {
	CampaignID string
}

// ListCampaignsRequest представляет запрос на получение списка кампаний
type ListCampaignsRequest struct {
	Limit  int    // Лимит количества кампаний (опционально)
	Offset int    // Смещение для пагинации (опционально)
	Status string // Фильтр по статусу (опционально)
}

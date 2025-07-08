package dto

import "whatsapp-service/internal/entities/campaign"

// CreateCampaignResponse представляет ответ на создание кампании
type CreateCampaignResponse struct {
	Campaign       *campaign.Campaign // Созданная кампания
	ValidPhones    int                // Количество валидных номеров
	InvalidPhones  int                // Количество невалидных номеров
	DuplicateCount int                // Количество дубликатов
	TotalNumbers   int                // Общее количество номеров после обработки
	Warnings       []string           // Предупреждения
}

// CancelCampaignResponse представляет ответ на отмену кампании
type CancelCampaignResponse struct {
	CampaignID         string                  // ID кампании
	Status             campaign.CampaignStatus // Новый статус кампании
	CancelledNumbers   int                     // Количество отмененных номеров
	AlreadySentNumbers int                     // Количество уже отправленных
	TotalNumbers       int                     // Общее количество номеров
	WorkerStopped      bool                    // Остановлен ли background worker
	Reason             string                  // Причина отмены
}

// StartCampaignResponse представляет ответ на запуск кампании
type StartCampaignResponse struct {
	CampaignID          string                  // ID кампании
	Status              campaign.CampaignStatus // Новый статус кампании
	TotalNumbers        int                     // Общее количество номеров для отправки
	EstimatedCompletion string                  // Ориентировочное время завершения
	WorkerStarted       bool                    // Запущен ли background worker
}

// PhoneNumberStatus представляет информацию о номере телефона и его статусе
type PhoneNumberStatus struct {
	ID                string
	PhoneNumber       string
	Status            string
	Error             string
	WhatsappMessageID string
	SentAt            string
	DeliveredAt       string
	ReadAt            string
	CreatedAt         string
}

// MediaInfo представляет информацию о медиафайле в кампании
type MediaInfo struct {
	ID          string
	Filename    string
	MimeType    string
	MessageType string
	Size        int64
	StoragePath string
	ChecksumMD5 string
	CreatedAt   string
}

type GetCampaignByIDResponse struct {
	ID              string
	Name            string
	Message         string
	Status          campaign.CampaignStatus
	TotalCount      int
	ProcessedCount  int
	ErrorCount      int
	MessagesPerHour int
	CreatedAt       string
	SentNumbers     []PhoneNumberStatus
	FailedNumbers   []PhoneNumberStatus
	Media           *MediaInfo
}

// CampaignSummary представляет краткую информацию о кампании для списка
type CampaignSummary struct {
	ID              string
	Name            string
	Status          campaign.CampaignStatus
	TotalCount      int
	ProcessedCount  int
	ErrorCount      int
	MessagesPerHour int
	CreatedAt       string
}

// ListCampaignsResponse представляет ответ на запрос списка кампаний
type ListCampaignsResponse struct {
	Campaigns []CampaignSummary
	Total     int
	Limit     int
	Offset    int
}

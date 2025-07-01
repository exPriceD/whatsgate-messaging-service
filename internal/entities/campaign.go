package entities

import (
	"time"
	"whatsapp-service/internal/entities/errors"
)

// CampaignStatus представляет статус кампании
type CampaignStatus string

const (
	CampaignStatusPending   CampaignStatus = "pending"
	CampaignStatusStarted   CampaignStatus = "started"
	CampaignStatusFinished  CampaignStatus = "finished"
	CampaignStatusFailed    CampaignStatus = "failed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

// Campaign представляет кампанию массовой рассылки
type Campaign struct {
	id                string
	name              string
	message           string
	phoneNumbers      []*PhoneNumber
	additionalNumbers []*PhoneNumber
	excludeNumbers    []*PhoneNumber
	status            CampaignStatus
	media             *Media
	messagesPerHour   int
	initiator         string
	errorCount        int
	createdAt         time.Time
	statuses          []*CampaignPhoneStatus
}

// NewCampaign создает новую кампанию
func NewCampaign(id, name, message string) *Campaign {
	return &Campaign{
		id:                id,
		name:              name,
		message:           message,
		phoneNumbers:      make([]*PhoneNumber, 0),
		additionalNumbers: make([]*PhoneNumber, 0),
		excludeNumbers:    make([]*PhoneNumber, 0),
		status:            CampaignStatusPending,
		messagesPerHour:   20,
		createdAt:         time.Now(),
		statuses:          make([]*CampaignPhoneStatus, 0),
	}
}

// CanBeCancelled проверяет, может ли кампания быть отменена
func (c *Campaign) CanBeCancelled() bool {
	return c.status == CampaignStatusPending || c.status == CampaignStatusStarted
}

// CanBeStarted проверяет, может ли кампания быть запущена
func (c *Campaign) CanBeStarted() bool {
	return c.status == CampaignStatusPending && c.TotalCount() > 0
}

// CanBeModified проверяет, может ли кампания быть изменена
func (c *Campaign) CanBeModified() bool {
	return c.status == CampaignStatusPending
}

// Start запускает кампанию
func (c *Campaign) Start() error {
	if !c.CanBeStarted() {
		return errors.ErrCampaignNotPending
	}

	c.status = CampaignStatusStarted

	targetNumbers := c.GetAllTargetNumbers()
	for _, phone := range targetNumbers {
		status := NewCampaignStatus(c.id, phone.Value())
		c.statuses = append(c.statuses, status)
	}

	return nil
}

// Cancel отменяет кампанию
func (c *Campaign) Cancel() error {
	if !c.CanBeCancelled() {
		return errors.ErrCannotCancelCampaign
	}

	c.status = CampaignStatusCancelled

	for _, status := range c.statuses {
		if status.Status() == CampaignStatusTypePending {
			status.Cancel()
		}
	}

	return nil
}

// Finish завершает кампанию
func (c *Campaign) Finish() {
	if c.status == CampaignStatusStarted {
		c.status = CampaignStatusFinished
	}
}

// Fail помечает кампанию как проваленную
func (c *Campaign) Fail() {
	c.status = CampaignStatusFailed
}

// AddPhoneNumbers добавляет основные номера телефонов
func (c *Campaign) AddPhoneNumbers(phones []*PhoneNumber) error {
	if len(phones) == 0 {
		return errors.ErrNoPhoneNumbers
	}
	c.phoneNumbers = append(c.phoneNumbers, phones...)
	return nil
}

// AddAdditionalNumbers добавляет дополнительные номера
func (c *Campaign) AddAdditionalNumbers(phones []*PhoneNumber) {
	c.additionalNumbers = append(c.additionalNumbers, phones...)
}

// AddExcludeNumbers добавляет номера для исключения
func (c *Campaign) AddExcludeNumbers(phones []*PhoneNumber) {
	c.excludeNumbers = append(c.excludeNumbers, phones...)
}

// RemovePhoneNumber удаляет номер телефона из кампании
func (c *Campaign) RemovePhoneNumber(phone PhoneNumber) error {
	if !c.CanBeModified() {
		return errors.ErrCannotModifyRunningCampaign
	}

	for i, p := range c.phoneNumbers {
		if p.Equal(phone) {
			c.phoneNumbers = append(c.phoneNumbers[:i], c.phoneNumbers[i+1:]...)
			return nil
		}
	}

	return errors.ErrPhoneNumberNotFound
}

// SetMedia устанавливает медиа-файл для кампании
func (c *Campaign) SetMedia(filename, mimeType string, data []byte) {
	c.media = NewMedia(filename, mimeType, data)
}

// SetMessagesPerHour устанавливает лимит сообщений в час
func (c *Campaign) SetMessagesPerHour(rate int) error {
	if rate <= 0 || rate > 3600 {
		return errors.ErrInvalidMessagesPerHour
	}
	c.messagesPerHour = rate
	return nil
}

// IncrementProcessedCount увеличивает счетчик обработанных сообщений
func (c *Campaign) IncrementProcessedCount() {
	processed := 0
	for _, status := range c.statuses {
		if status.IsProcessed() {
			processed++
		}
	}
}

// IncrementErrorCount увеличивает счетчик ошибок
func (c *Campaign) IncrementErrorCount() {
	c.errorCount++
}

// GetProgress возвращает прогресс выполнения кампании (0.0 - 1.0)
func (c *Campaign) GetProgress() float64 {
	total := c.TotalCount()
	if total == 0 {
		return 0.0
	}
	return float64(c.ProcessedCount()) / float64(total)
}

// IsCompleted checks if all messages have been processed
func (c *Campaign) IsCompleted() bool {
	return c.ProcessedCount() >= c.TotalCount()
}

// Getters (read-only access)

// ID возвращает идентификатор кампании
func (c *Campaign) ID() string {
	return c.id
}

// Name возвращает название кампании
func (c *Campaign) Name() string {
	return c.name
}

// Message возвращает текст сообщения
func (c *Campaign) Message() string {
	return c.message
}

// Status возвращает статус кампании
func (c *Campaign) Status() CampaignStatus {
	return c.status
}

// PhoneNumbers возвращает основные номера телефонов
func (c *Campaign) PhoneNumbers() []*PhoneNumber {
	return c.phoneNumbers
}

// AdditionalNumbers возвращает дополнительные номера
func (c *Campaign) AdditionalNumbers() []*PhoneNumber {
	return c.additionalNumbers
}

// ExcludeNumbers возвращает исключаемые номера
func (c *Campaign) ExcludeNumbers() []*PhoneNumber {
	return c.excludeNumbers
}

// CreatedAt возвращает время создания
func (c *Campaign) CreatedAt() time.Time {
	return c.createdAt
}

// ProcessedCount возвращает количество обработанных номеров
func (c *Campaign) ProcessedCount() int {
	processed := 0
	for _, status := range c.statuses {
		if status.IsProcessed() {
			processed++
		}
	}
	return processed
}

// ErrorCount возвращает количество ошибок
func (c *Campaign) ErrorCount() int {
	return c.errorCount
}

// TotalCount возвращает общее количество целевых номеров
func (c *Campaign) TotalCount() int {
	return len(c.GetAllTargetNumbers())
}

// Media возвращает медиа-файл кампании
func (c *Campaign) Media() *Media {
	return c.media
}

// Initiator возвращает инициатора кампании
func (c *Campaign) Initiator() string {
	return c.initiator
}

// SetInitiator устанавливает инициатора кампании
func (c *Campaign) SetInitiator(initiator string) {
	c.initiator = initiator
}

// MessagesPerHour возвращает лимит сообщений в час
func (c *Campaign) MessagesPerHour() int {
	return c.messagesPerHour
}

// Statuses возвращает статусы отправки по номерам
func (c *Campaign) Statuses() []*CampaignPhoneStatus {
	return c.statuses
}

// GetAllTargetNumbers возвращает все целевые номера (основные + дополнительные - исключаемые)
func (c *Campaign) GetAllTargetNumbers() []*PhoneNumber {
	// Создаем map для быстрого поиска исключаемых номеров
	excludeMap := make(map[string]struct{})
	for _, phone := range c.excludeNumbers {
		excludeMap[phone.Value()] = struct{}{}
	}

	// Объединяем основные и дополнительные номера
	allNumbers := make([]*PhoneNumber, 0, len(c.phoneNumbers)+len(c.additionalNumbers))
	allNumbers = append(allNumbers, c.phoneNumbers...)
	allNumbers = append(allNumbers, c.additionalNumbers...)

	// Исключаем дубликаты и исключаемые номера
	seen := make(map[string]struct{})
	result := make([]*PhoneNumber, 0, len(allNumbers))

	for _, phone := range allNumbers {
		phoneValue := phone.Value()
		// Пропускаем если номер в списке исключений
		if _, excluded := excludeMap[phoneValue]; excluded {
			continue
		}
		// Пропускаем дубликаты
		if _, exists := seen[phoneValue]; exists {
			continue
		}

		seen[phoneValue] = struct{}{}
		result = append(result, phone)
	}

	return result
}

// IsActive проверяет, является ли кампания активной
func (c *Campaign) IsActive() bool {
	return c.status == CampaignStatusPending || c.status == CampaignStatusStarted
}

// AddCampaignStatus добавляет статус отправки для номера
func (c *Campaign) AddCampaignStatus(status *CampaignPhoneStatus) {
	c.statuses = append(c.statuses, status)
}

// GetStatusForPhone возвращает статус для конкретного номера
func (c *Campaign) GetStatusForPhone(phoneNumber string) *CampaignPhoneStatus {
	for _, status := range c.statuses {
		if status.PhoneNumber() == phoneNumber {
			return status
		}
	}
	return nil
}

// GetFailedStatuses возвращает все неудачные статусы
func (c *Campaign) GetFailedStatuses() []*CampaignPhoneStatus {
	var failed []*CampaignPhoneStatus
	for _, status := range c.statuses {
		if status.Status() == CampaignStatusTypeFailed {
			failed = append(failed, status)
		}
	}
	return failed
}

// GetSentStatuses возвращает все успешно отправленные статусы
func (c *Campaign) GetSentStatuses() []*CampaignPhoneStatus {
	var sent []*CampaignPhoneStatus
	for _, status := range c.statuses {
		if status.Status() == CampaignStatusTypeSent {
			sent = append(sent, status)
		}
	}
	return sent
}

// Helper function to generate unique ID
func generateID() string {
	// TODO: implement proper ID generation
	return "campaign_" + time.Now().Format("20060102_150405")
}

package entities

import (
	"time"
)

// CampaignStatusType представляет статус отправки для конкретного номера
type CampaignStatusType string

const (
	CampaignStatusTypePending   CampaignStatusType = "pending"
	CampaignStatusTypeSent      CampaignStatusType = "sent"
	CampaignStatusTypeFailed    CampaignStatusType = "failed"
	CampaignStatusTypeCancelled CampaignStatusType = "cancelled"
)

// CampaignPhoneStatus представляет статус отправки сообщения на конкретный номер
type CampaignPhoneStatus struct {
	id          string
	campaignID  string
	phoneNumber string
	status      CampaignStatusType
	error       string
	sentAt      *time.Time
	createdAt   time.Time
}

// NewCampaignStatus создает новый статус кампании для номера
func NewCampaignStatus(campaignID, phoneNumber string) *CampaignPhoneStatus {
	return &CampaignPhoneStatus{
		id:          generateID(),
		campaignID:  campaignID,
		phoneNumber: phoneNumber,
		status:      CampaignStatusTypePending,
		createdAt:   time.Now(),
	}
}

// ID возвращает идентификатор статуса
func (cs *CampaignPhoneStatus) ID() string {
	return cs.id
}

// CampaignID возвращает идентификатор кампании
func (cs *CampaignPhoneStatus) CampaignID() string {
	return cs.campaignID
}

// PhoneNumber возвращает номер телефона
func (cs *CampaignPhoneStatus) PhoneNumber() string {
	return cs.phoneNumber
}

// Status возвращает текущий статус
func (cs *CampaignPhoneStatus) Status() CampaignStatusType {
	return cs.status
}

// Error возвращает сообщение об ошибке (если есть)
func (cs *CampaignPhoneStatus) Error() string {
	return cs.error
}

// SentAt возвращает время отправки сообщения
func (cs *CampaignPhoneStatus) SentAt() *time.Time {
	return cs.sentAt
}

// CreatedAt возвращает время создания статуса
func (cs *CampaignPhoneStatus) CreatedAt() time.Time {
	return cs.createdAt
}

// MarkAsSent помечает сообщение как отправленное
func (cs *CampaignPhoneStatus) MarkAsSent() {
	cs.status = CampaignStatusTypeSent
	now := time.Now()
	cs.sentAt = &now
	cs.error = "" // Очищаем ошибку при успешной отправке
}

// MarkAsFailed помечает сообщение как неудачное
func (cs *CampaignPhoneStatus) MarkAsFailed(errorMsg string) {
	cs.status = CampaignStatusTypeFailed
	cs.error = errorMsg
}

// Cancel отменяет отправку сообщения
func (cs *CampaignPhoneStatus) Cancel() {
	if cs.status == CampaignStatusTypePending {
		cs.status = CampaignStatusTypeCancelled
	}
}

// IsProcessed проверяет, был ли номер обработан (отправлен, провален или отменен)
func (cs *CampaignPhoneStatus) IsProcessed() bool {
	return cs.status != CampaignStatusTypePending
}

// IsSuccessful проверяет, была ли отправка успешной
func (cs *CampaignPhoneStatus) IsSuccessful() bool {
	return cs.status == CampaignStatusTypeSent
}

// IsFailed проверяет, была ли отправка неуспешной
func (cs *CampaignPhoneStatus) IsFailed() bool {
	return cs.status == CampaignStatusTypeFailed
}

// IsCancelled проверяет, была ли отправка отменена
func (cs *CampaignPhoneStatus) IsCancelled() bool {
	return cs.status == CampaignStatusTypeCancelled
}

// CanBeRetried проверяет, можно ли повторить отправку
func (cs *CampaignPhoneStatus) CanBeRetried() bool {
	return cs.status == CampaignStatusTypeFailed
}

// Retry сбрасывает статус для повторной попытки отправки
func (cs *CampaignPhoneStatus) Retry() {
	if cs.CanBeRetried() {
		cs.status = CampaignStatusTypePending
		cs.error = ""
		cs.sentAt = nil
	}
}

// SetID устанавливает идентификатор (для восстановления из БД)
func (cs *CampaignPhoneStatus) SetID(id string) {
	cs.id = id
}

// SetSentAt устанавливает время отправки (для восстановления из БД)
func (cs *CampaignPhoneStatus) SetSentAt(sentAt *time.Time) {
	cs.sentAt = sentAt
}

// SetCreatedAt устанавливает время создания (для восстановления из БД)
func (cs *CampaignPhoneStatus) SetCreatedAt(createdAt time.Time) {
	cs.createdAt = createdAt
}

// RestoreFromDB восстанавливает статус из данных БД
func (cs *CampaignPhoneStatus) RestoreFromDB(id, campaignID, phoneNumber string, status CampaignStatusType, errorMsg string, sentAt *time.Time, createdAt time.Time) {
	cs.id = id
	cs.campaignID = campaignID
	cs.phoneNumber = phoneNumber
	cs.status = status
	cs.error = errorMsg
	cs.sentAt = sentAt
	cs.createdAt = createdAt
}

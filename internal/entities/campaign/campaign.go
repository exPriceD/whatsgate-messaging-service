package campaign

import (
	"time"

	"github.com/google/uuid"
)

// generateID создает новый UUID для сущности
func generateID() string {
	return uuid.New().String()
}

type CampaignStatus string

const (
	CampaignStatusPending   CampaignStatus = "pending"
	CampaignStatusFiltering CampaignStatus = "filtering"
	CampaignStatusStarted   CampaignStatus = "started"
	CampaignStatusFinished  CampaignStatus = "finished"
	CampaignStatusFailed    CampaignStatus = "failed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type TargetAudience struct {
	Primary    []*PhoneNumber
	Additional []*PhoneNumber
	Excluded   []*PhoneNumber
}

func (a *TargetAudience) AllTargets() []*PhoneNumber {
	excludeMap := make(map[string]struct{})
	for _, phone := range a.Excluded {
		excludeMap[phone.Value()] = struct{}{}
	}

	allNumbers := append([]*PhoneNumber{}, a.Primary...)
	allNumbers = append(allNumbers, a.Additional...)

	seen := make(map[string]struct{})
	result := make([]*PhoneNumber, 0, len(allNumbers))
	for _, phone := range allNumbers {
		v := phone.Value()
		if _, excluded := excludeMap[v]; excluded {
			continue
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, phone)
	}
	return result
}

type CampaignMetrics struct {
	Total     int
	Processed int
	Errors    int
}

func (m *CampaignMetrics) Progress() float64 {
	if m.Total == 0 {
		return 0.0
	}
	return float64(m.Processed) / float64(m.Total)
}

func (m *CampaignMetrics) MarkProcessed() { m.Processed++ }
func (m *CampaignMetrics) MarkError()     { m.Errors++ }
func (m *CampaignMetrics) IsCompleted() bool {
	return m.Processed >= m.Total
}

type DeliveryStatus struct {
	records []*CampaignPhoneStatus
}

func (d *DeliveryStatus) Add(status *CampaignPhoneStatus) {
	d.records = append(d.records, status)
}

func (d *DeliveryStatus) CancelPending() {
	for _, status := range d.records {
		if status.Status() == CampaignStatusTypePending {
			status.Cancel()
		}
	}
}

func (d *DeliveryStatus) ForPhone(phoneNumber string) *CampaignPhoneStatus {
	for _, status := range d.records {
		if status.PhoneNumber() == phoneNumber {
			return status
		}
	}
	return nil
}

func (d *DeliveryStatus) Failed() []*CampaignPhoneStatus {
	var failed []*CampaignPhoneStatus
	for _, status := range d.records {
		if status.Status() == CampaignStatusTypeFailed {
			failed = append(failed, status)
		}
	}
	return failed
}

func (d *DeliveryStatus) Sent() []*CampaignPhoneStatus {
	var sent []*CampaignPhoneStatus
	for _, status := range d.records {
		if status.Status() == CampaignStatusTypeSent {
			sent = append(sent, status)
		}
	}
	return sent
}

func (d *DeliveryStatus) All() []*CampaignPhoneStatus {
	return d.records
}

type Campaign struct {
	id              string
	name            string
	message         string
	status          CampaignStatus
	media           *Media
	messagesPerHour int
	initiator       string
	categoryName    string
	createdAt       time.Time
	audience        *TargetAudience
	metrics         *CampaignMetrics
	delivery        *DeliveryStatus
}

func NewCampaign(name, message string, messagesPerHour int, categoryName string) *Campaign {
	return &Campaign{
		id:              generateID(),
		name:            name,
		message:         message,
		status:          CampaignStatusPending,
		messagesPerHour: messagesPerHour,
		categoryName:    categoryName,
		createdAt:       time.Now(),
		initiator:       "",
		audience:        &TargetAudience{},
		metrics:         &CampaignMetrics{},
		delivery:        &DeliveryStatus{},
	}
}

func RestoreCampaign(
	id, name, message, initiator string,
	status CampaignStatus,
	media *Media,
	messagesPerHour int,
	categoryName string,
	createdAt time.Time,
	audience *TargetAudience,
	metrics *CampaignMetrics,
	delivery *DeliveryStatus,
) *Campaign {
	return &Campaign{
		id:              id,
		name:            name,
		message:         message,
		status:          status,
		media:           media,
		messagesPerHour: messagesPerHour,
		categoryName:    categoryName,
		createdAt:       createdAt,
		initiator:       initiator,
		audience:        audience,
		metrics:         metrics,
		delivery:        delivery,
	}
}

func (c *Campaign) ID() string             { return c.id }
func (c *Campaign) Name() string           { return c.name }
func (c *Campaign) Message() string        { return c.message }
func (c *Campaign) Status() CampaignStatus { return c.status }
func (c *Campaign) CreatedAt() time.Time   { return c.createdAt }
func (c *Campaign) Initiator() string      { return c.initiator }
func (c *Campaign) Media() *Media          { return c.media }
func (c *Campaign) MessagesPerHour() int   { return c.messagesPerHour }
func (c *Campaign) CategoryName() string   { return c.categoryName }

func (c *Campaign) Audience() *TargetAudience { return c.audience }
func (c *Campaign) Metrics() *CampaignMetrics { return c.metrics }
func (c *Campaign) Delivery() *DeliveryStatus { return c.delivery }

func (c *Campaign) AddPhoneNumbers(numbers []*PhoneNumber) error {
	if len(numbers) == 0 {
		return ErrNoPhoneNumbers
	}
	c.audience.Primary = append(c.audience.Primary, numbers...)
	return nil
}

func (c *Campaign) AddAdditionalNumbers(numbers []*PhoneNumber) {
	c.audience.Additional = append(c.audience.Additional, numbers...)
}

func (c *Campaign) AddExcludedNumbers(numbers []*PhoneNumber) {
	c.audience.Excluded = append(c.audience.Excluded, numbers...)
}

// SetInitiator устанавливает инициатора кампании
func (c *Campaign) SetInitiator(initiator string) {
	c.initiator = initiator
}

// SetMedia устанавливает медиа-файл кампании
func (c *Campaign) SetMedia(media *Media) {
	c.media = media
}

// SetStatus устанавливает статус кампании
func (c *Campaign) SetStatus(status CampaignStatus) {
	c.status = status
}

func (c *Campaign) CanBeCancelled() bool {
	return c.status == CampaignStatusPending || c.status == CampaignStatusStarted || c.status == CampaignStatusFiltering
}

func (c *Campaign) CanBeStarted() bool {
	return (c.status == CampaignStatusPending || c.status == CampaignStatusFiltering) && c.metrics.Total > 0
}

func (c *Campaign) CanBeModified() bool {
	return c.status == CampaignStatusPending
}

func (c *Campaign) Start() error {
	if !c.CanBeStarted() {
		return ErrCampaignNotPending
	}
	c.status = CampaignStatusStarted
	for _, phone := range c.audience.AllTargets() {
		c.delivery.Add(NewCampaignStatus(c.id, phone.Value()))
	}
	return nil
}

func (c *Campaign) Cancel() error {
	if !c.CanBeCancelled() {
		return ErrCannotCancelCampaign
	}
	c.status = CampaignStatusCancelled
	c.delivery.CancelPending()
	return nil
}

func (c *Campaign) Finish() {
	if c.status == CampaignStatusStarted {
		c.status = CampaignStatusFinished
	}
}
func (c *Campaign) Fail() { c.status = CampaignStatusFailed }

func (c *Campaign) MarkProcessed()    { c.metrics.MarkProcessed() }
func (c *Campaign) MarkError()        { c.metrics.MarkError() }
func (c *Campaign) Progress() float64 { return c.metrics.Progress() }
func (c *Campaign) IsCompleted() bool { return c.metrics.IsCompleted() }
func (c *Campaign) IsActive() bool {
	return c.status == CampaignStatusPending || c.status == CampaignStatusStarted || c.status == CampaignStatusFiltering
}

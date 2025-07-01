package errors

import "errors"

// Бизнес-ошибки кампаний
var (
	ErrCannotStartCampaign         = errors.New("campaign cannot be started")
	ErrCannotCancelCampaign        = errors.New("campaign cannot be cancelled")
	ErrCannotFinishCampaign        = errors.New("campaign cannot be finished")
	ErrCannotFailCampaign          = errors.New("campaign cannot be failed")
	ErrCannotCompleteCampaign      = errors.New("campaign cannot be completed")
	ErrCannotModifyRunningCampaign = errors.New("cannot modify running campaign")
	ErrCampaignNotPending          = errors.New("campaign is not in pending status")
	ErrNoPhoneNumbers              = errors.New("no phone numbers provided")
	ErrInvalidPhoneNumber          = errors.New("invalid phone number")
	ErrPhoneNumberNotFound         = errors.New("phone number not found in campaign")
	ErrInvalidMessagesPerHour      = errors.New("invalid messages per hour rate")
	ErrCampaignNotFound            = errors.New("campaign not found")
	ErrRepositoryError             = errors.New("repository error")
	ErrCampaignAlreadyRunning      = errors.New("уже есть запущенная кампания")
	ErrCampaignNameRequired        = errors.New("название кампании обязательно")
	ErrCampaignMessageRequired     = errors.New("сообщение кампании обязательно")
)

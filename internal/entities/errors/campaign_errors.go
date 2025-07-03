package errors

import (
	"errors"
)

var (
	ErrCannotStartCampaign         = errors.New("campaign cannot be started")
	ErrCannotCancelCampaign        = errors.New("campaign cannot be cancelled")
	ErrCannotModifyRunningCampaign = errors.New("cannot modify running campaign")
	ErrCampaignNotPending          = errors.New("campaign is not in pending status")
	ErrNoPhoneNumbers              = errors.New("no phone numbers provided")
	ErrInvalidPhoneNumber          = errors.New("invalid phone number")
	ErrPhoneNumberNotFound         = errors.New("phone number not found in campaign")
	ErrInvalidMessagesPerHour      = errors.New("invalid messages per hour rate")
	ErrCampaignNotFound            = errors.New("campaign not found")
	ErrRepositoryError             = errors.New("repository error")
	ErrCampaignAlreadyRunning      = errors.New("campaign is already running")
	ErrCampaignNameRequired        = errors.New("campaign name is required")
	ErrCampaignMessageRequired     = errors.New("campaign message is required")
	ErrCampaignCannotBeStarted     = errors.New("campaign cannot be started in its current state")
)

package messaging

import "context"

// GlobalRateLimiter defines the contract for a component that enforces a single,
// shared rate limit across the entire application.
type GlobalRateLimiter interface {
	// SetRate sets the global rate for all campaigns, typically in messages per hour.
	// This determines the size of the token bucket for the shared limit.
	SetRate(messagesPerHour int)

	// SetRateForCampaign sets the rate limit for a specific campaign.
	// This will be used when a new campaign is added to the dispatcher.
	SetRateForCampaign(campaignID string, messagesPerHour int)

	// Wait blocks until a message can be sent according to the global limit,
	// or until the context is canceled.
	Wait(ctx context.Context) error

	// WaitForCampaign blocks until a message can be sent for a specific campaign,
	// according to that campaign's rate limit.
	WaitForCampaign(ctx context.Context, campaignID string) error

	// Reset clears the global rate limiter state, effectively ending the current
	// batch and pause period. This should be called when all campaigns are finished.
	Reset()
}

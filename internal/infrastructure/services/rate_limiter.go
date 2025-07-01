package services

import (
	"sync"
	"time"
	"whatsapp-service/internal/usecases/interfaces"
)

// MemoryRateLimiter простая in-memory реализация RateLimiter
type MemoryRateLimiter struct {
	campaigns map[string]*campaignLimiter
	mutex     sync.RWMutex
}

type campaignLimiter struct {
	messagesPerHour int
	sentCount       int
	lastReset       time.Time
	lastSent        time.Time
}

// NewMemoryRateLimiter создает новый in-memory rate limiter
func NewMemoryRateLimiter() interfaces.RateLimiter {
	return &MemoryRateLimiter{
		campaigns: make(map[string]*campaignLimiter),
	}
}

// CanSend проверяет, можно ли отправить сообщение для данной кампании
func (rl *MemoryRateLimiter) CanSend(campaignID string) bool {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	limiter, exists := rl.campaigns[campaignID]
	if !exists {
		return true
	}

	now := time.Now()

	// Сбрасываем счетчик раз в час
	if now.Sub(limiter.lastReset) >= time.Hour {
		limiter.sentCount = 0
		limiter.lastReset = now
	}

	return limiter.sentCount < limiter.messagesPerHour
}

// MessageSent уведомляет лимитер о том, что сообщение было отправлено
func (rl *MemoryRateLimiter) MessageSent(campaignID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	limiter, exists := rl.campaigns[campaignID]
	if !exists {
		limiter = &campaignLimiter{
			messagesPerHour: 60,
			lastReset:       now,
		}
		rl.campaigns[campaignID] = limiter
	}

	limiter.sentCount++
	limiter.lastSent = now
}

// SetRate устанавливает лимит сообщений в час для кампании
func (rl *MemoryRateLimiter) SetRate(campaignID string, messagesPerHour int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.campaigns[campaignID]
	if !exists {
		limiter = &campaignLimiter{
			lastReset: time.Now(),
		}
		rl.campaigns[campaignID] = limiter
	}

	limiter.messagesPerHour = messagesPerHour
}

// TimeToNext возвращает время в секундах до следующей возможной отправки
func (rl *MemoryRateLimiter) TimeToNext(campaignID string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	limiter, exists := rl.campaigns[campaignID]
	if !exists {
		return 0
	}

	if limiter.messagesPerHour <= 0 {
		return 0
	}

	// Интервал между сообщениями
	interval := 3600 / limiter.messagesPerHour // секунды
	timeSinceLastSent := int(time.Since(limiter.lastSent).Seconds())

	if timeSinceLastSent >= interval {
		return 0
	}

	return interval - timeSinceLastSent
}

// GetWaitTime возвращает время ожидания до следующей отправки
func (rl *MemoryRateLimiter) GetWaitTime(campaignID string) int {
	return rl.TimeToNext(campaignID)
}

// Reset сбрасывает счетчики для кампании
func (rl *MemoryRateLimiter) Reset(campaignID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	delete(rl.campaigns, campaignID)
}

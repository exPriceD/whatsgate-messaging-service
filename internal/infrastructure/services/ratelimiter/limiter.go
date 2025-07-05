package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// GlobalMemoryRateLimiter реализует GlobalRateLimiter с состоянием в памяти.
// Он обеспечивает единый лимит скорости для всех операций.
type GlobalMemoryRateLimiter struct {
	mutex sync.Mutex

	// Настройки
	batchSize    int
	pause        time.Duration
	sendInterval time.Duration

	// Состояние
	sentInBatch int
	batchStart  time.Time

	// Поля для переопределения в тестах
	testPauseOverride    time.Duration
	testIntervalOverride time.Duration
}

// NewGlobalMemoryRateLimiter создает новый глобальный rate limiter.
// Он приводится к интерфейсу, чтобы скрыть поля для тестов.
func NewGlobalMemoryRateLimiter() *GlobalMemoryRateLimiter {
	return &GlobalMemoryRateLimiter{
		batchSize:    20,
		pause:        time.Hour,
		sendInterval: time.Second,
		batchStart:   time.Now(),
	}
}

// SetRate устанавливает глобальный лимит для всех кампаний.
func (rl *GlobalMemoryRateLimiter) SetRate(messagesPerHour int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if messagesPerHour <= 0 {
		messagesPerHour = 20 // Значение по умолчанию, если передано некорректное
	}
	rl.batchSize = messagesPerHour
}

// Wait блокирует выполнение до тех пор, пока отправка не будет разрешена глобальным лимитом.
func (rl *GlobalMemoryRateLimiter) Wait(ctx context.Context) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	pause := rl.pause
	if rl.testPauseOverride > 0 {
		pause = rl.testPauseOverride
	}

	interval := rl.sendInterval
	if rl.testIntervalOverride > 0 {
		interval = rl.testIntervalOverride
	}

	if rl.sentInBatch >= rl.batchSize {
		elapsed := time.Since(rl.batchStart)
		if elapsed < pause {
			waitTime := pause - elapsed
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		// Сбрасываем пакет
		rl.sentInBatch = 0
		rl.batchStart = time.Now()
	}

	if rl.sentInBatch > 0 {
		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	rl.sentInBatch++
	return nil
}

// Reset сбрасывает состояние глобального лимитера.
func (rl *GlobalMemoryRateLimiter) Reset() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.sentInBatch = 0
	rl.batchStart = time.Now()
}

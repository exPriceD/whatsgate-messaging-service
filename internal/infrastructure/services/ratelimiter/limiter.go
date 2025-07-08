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

	// Настройки кампаний
	campaignLimits map[string]*campaignRateState

	// Поля для переопределения в тестах
	testPauseOverride    time.Duration
	testIntervalOverride time.Duration
}

// campaignRateState хранит состояние rate limiting для конкретной кампании
type campaignRateState struct {
	batchSize    int
	pause        time.Duration
	sendInterval time.Duration
	sentInBatch  int
	batchStart   time.Time
}

// NewGlobalMemoryRateLimiter создает новый глобальный rate limiter.
// Он приводится к интерфейсу, чтобы скрыть поля для тестов.
func NewGlobalMemoryRateLimiter() *GlobalMemoryRateLimiter {
	return &GlobalMemoryRateLimiter{
		batchSize:      20,
		pause:          time.Hour,
		sendInterval:   time.Second,
		batchStart:     time.Now(),
		campaignLimits: make(map[string]*campaignRateState),
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

// SetRateForCampaign устанавливает лимит для конкретной кампании.
func (rl *GlobalMemoryRateLimiter) SetRateForCampaign(campaignID string, messagesPerHour int) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if messagesPerHour <= 0 {
		messagesPerHour = 20 // Значение по умолчанию, если передано некорректное
	}

	// Настройки для пакетной отправки:
	// - batchSize = количество сообщений в пакете (= messagesPerHour)
	// - sendInterval = 2 секунды между сообщениями в пакете
	// - pause = 1 час между пакетами

	rl.campaignLimits[campaignID] = &campaignRateState{
		batchSize:    messagesPerHour,
		pause:        time.Hour,
		sendInterval: 2 * time.Second, // 2 секунды между сообщениями в пакете
		sentInBatch:  0,
		batchStart:   time.Now(),
	}
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

// WaitForCampaign блокирует выполнение до тех пор, пока отправка не будет разрешена лимитом кампании.
func (rl *GlobalMemoryRateLimiter) WaitForCampaign(ctx context.Context, campaignID string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Проверяем есть ли настройки для кампании
	campaignState, exists := rl.campaignLimits[campaignID]
	if !exists {
		// Если настроек нет, используем глобальные
		return rl.waitGlobal(ctx)
	}

	// Используем настройки кампании
	pause := campaignState.pause
	interval := campaignState.sendInterval

	// Переопределения для тестов
	if rl.testPauseOverride > 0 {
		pause = rl.testPauseOverride
	}
	if rl.testIntervalOverride > 0 {
		interval = rl.testIntervalOverride
	}

	if campaignState.sentInBatch >= campaignState.batchSize {
		elapsed := time.Since(campaignState.batchStart)
		if elapsed < pause {
			waitTime := pause - elapsed
			select {
			case <-time.After(waitTime):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		// Сбрасываем пакет кампании
		campaignState.sentInBatch = 0
		campaignState.batchStart = time.Now()
	}

	if campaignState.sentInBatch > 0 {
		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	campaignState.sentInBatch++
	return nil
}

// waitGlobal - вспомогательный метод для применения глобального лимита
func (rl *GlobalMemoryRateLimiter) waitGlobal(ctx context.Context) error {
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

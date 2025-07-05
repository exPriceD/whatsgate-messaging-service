package ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setup создает и настраивает GlobalMemoryRateLimiter для тестов.
func setupGlobalRateLimiter(t *testing.T, interval, pause time.Duration) *GlobalMemoryRateLimiter {
	t.Helper()
	rl, ok := NewGlobalMemoryRateLimiter().(*GlobalMemoryRateLimiter)
	require.True(t, ok, "NewGlobalMemoryRateLimiter should return a *GlobalMemoryRateLimiter")
	rl.testIntervalOverride = interval
	rl.testPauseOverride = pause
	return rl
}

func TestGlobalMemoryRateLimiter_SetRate(t *testing.T) {
	rl := setupGlobalRateLimiter(t, 0, 0)
	batchSize := 50

	rl.SetRate(batchSize)
	assert.Equal(t, batchSize, rl.batchSize, "Batch size should be set correctly")
}

func TestGlobalMemoryRateLimiter_SetRate_Zero(t *testing.T) {
	rl := setupGlobalRateLimiter(t, 0, 0)
	rl.SetRate(0)
	assert.Equal(t, 20, rl.batchSize, "Batch size should default to 20 when set to 0")
}

func TestGlobalMemoryRateLimiter_BatchExecution(t *testing.T) {
	interval := 10 * time.Millisecond
	pause := 100 * time.Millisecond
	rl := setupGlobalRateLimiter(t, interval, pause)

	batchSize := 5
	rl.SetRate(batchSize)

	start := time.Now()
	for i := 0; i < batchSize; i++ {
		err := rl.Wait(context.Background())
		require.NoError(t, err)
	}
	duration := time.Since(start)

	expectedMinDuration := time.Duration(batchSize-1) * interval
	assert.GreaterOrEqual(t, duration, expectedMinDuration, "Execution time should be at least the sum of intervals")
	assert.Less(t, duration, pause, "Execution time should be less than the long pause")
}

func TestGlobalMemoryRateLimiter_PauseAfterBatch(t *testing.T) {
	interval := 5 * time.Millisecond
	pause := 50 * time.Millisecond
	rl := setupGlobalRateLimiter(t, interval, pause)

	batchSize := 3
	rl.SetRate(batchSize)

	for i := 0; i < batchSize; i++ {
		require.NoError(t, rl.Wait(context.Background()))
	}

	start := time.Now()
	require.NoError(t, rl.Wait(context.Background()))
	duration := time.Since(start)

	totalIntervalTime := time.Duration(batchSize-1) * interval
	expectedMinPause := pause - totalIntervalTime

	assert.GreaterOrEqual(t, duration, expectedMinPause)
	assert.Less(t, duration, expectedMinPause+20*time.Millisecond, "The wait should be very close to the calculated pause")
}

func TestGlobalMemoryRateLimiter_Reset(t *testing.T) {
	interval := 5 * time.Millisecond
	pause := 100 * time.Millisecond
	rl := setupGlobalRateLimiter(t, interval, pause)

	batchSize := 2
	rl.SetRate(batchSize)

	require.NoError(t, rl.Wait(context.Background()))
	require.NoError(t, rl.Wait(context.Background()))

	rl.Reset()

	start := time.Now()
	require.NoError(t, rl.Wait(context.Background()))
	duration := time.Since(start)

	assert.Less(t, duration, interval, "Wait should be immediate after reset")
}

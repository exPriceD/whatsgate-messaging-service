package bulk_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"whatsapp-service/internal/bulk/domain"
	"whatsapp-service/internal/bulk/mocks"
)

func TestMockWhatsGateClient_SendTextMessage(t *testing.T) {
	client := &mocks.MockWhatsGateClient{}

	// Тест с дефолтным поведением
	result, err := client.SendTextMessage(context.Background(), "71234567890", "Test message", false)
	require.NoError(t, err)
	assert.Equal(t, "71234567890", result.PhoneNumber)
	assert.True(t, result.Success)
	assert.Equal(t, "sent", result.Status)

	// Тест с кастомной функцией
	client.SendTextMessageFunc = func(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
		return domain.SingleSendResult{
			PhoneNumber: phoneNumber,
			Success:     true,
			Status:      "delivered",
			Error:       "",
		}, nil
	}

	result, err = client.SendTextMessage(context.Background(), "79876543210", "Custom message", true)
	require.NoError(t, err)
	assert.Equal(t, "79876543210", result.PhoneNumber)
	assert.True(t, result.Success)
	assert.Equal(t, "delivered", result.Status)

	// Тест с ошибкой
	client.SendTextMessageFunc = func(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
		return domain.SingleSendResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Status:      "failed",
			Error:       "network error",
		}, errors.New("network error")
	}

	result, err = client.SendTextMessage(context.Background(), "71234567890", "Error message", false)
	require.Error(t, err)
	assert.Equal(t, "71234567890", result.PhoneNumber)
	assert.False(t, result.Success)
	assert.Equal(t, "failed", result.Status)
	assert.Equal(t, "network error", result.Error)
	assert.Equal(t, "network error", err.Error())
}

func TestMockWhatsGateClient_SendMediaMessage(t *testing.T) {
	client := &mocks.MockWhatsGateClient{}

	// Тест с дефолтным поведением
	fileData := []byte("fake-image-data")
	result, err := client.SendMediaMessage(context.Background(), "71234567890", "image", "Test image", "test.jpg", fileData, "image/jpeg", false)
	require.NoError(t, err)
	assert.Equal(t, "71234567890", result.PhoneNumber)
	assert.True(t, result.Success)
	assert.Equal(t, "sent", result.Status)

	// Тест с кастомной функцией
	client.SendMediaMessageFunc = func(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (domain.SingleSendResult, error) {
		return domain.SingleSendResult{
			PhoneNumber: phoneNumber,
			Success:     true,
			Status:      "delivered",
			Error:       "",
		}, nil
	}

	result, err = client.SendMediaMessage(context.Background(), "79876543210", "video", "Test video", "test.mp4", fileData, "video/mp4", true)
	require.NoError(t, err)
	assert.Equal(t, "79876543210", result.PhoneNumber)
	assert.True(t, result.Success)
	assert.Equal(t, "delivered", result.Status)

	// Тест с ошибкой
	client.SendMediaMessageFunc = func(ctx context.Context, phoneNumber, messageType, text, filename string, fileData []byte, mimeType string, async bool) (domain.SingleSendResult, error) {
		return domain.SingleSendResult{
			PhoneNumber: phoneNumber,
			Success:     false,
			Status:      "failed",
			Error:       "file too large",
		}, errors.New("file too large")
	}

	result, err = client.SendMediaMessage(context.Background(), "71234567890", "image", "Error image", "large.jpg", fileData, "image/jpeg", false)
	require.Error(t, err)
	assert.Equal(t, "71234567890", result.PhoneNumber)
	assert.False(t, result.Success)
	assert.Equal(t, "failed", result.Status)
	assert.Equal(t, "file too large", result.Error)
	assert.Equal(t, "file too large", err.Error())
}

func TestMockWhatsGateClient_ContextHandling(t *testing.T) {
	client := &mocks.MockWhatsGateClient{}

	// Тест с контекстом
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.SendTextMessageFunc = func(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
		select {
		case <-ctx.Done():
			return domain.SingleSendResult{}, ctx.Err()
		default:
			return domain.SingleSendResult{
				PhoneNumber: phoneNumber,
				Success:     true,
				Status:      "sent",
			}, nil
		}
	}

	// Тест с активным контекстом
	result, err := client.SendTextMessage(ctx, "71234567890", "Test message", false)
	require.NoError(t, err)
	assert.True(t, result.Success)

	// Тест с отмененным контекстом
	cancel()
	result, err = client.SendTextMessage(ctx, "71234567890", "Test message", false)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestMockWhatsGateClient_AsyncHandling(t *testing.T) {
	client := &mocks.MockWhatsGateClient{}

	// Тест синхронной отправки
	client.SendTextMessageFunc = func(ctx context.Context, phoneNumber, text string, async bool) (domain.SingleSendResult, error) {
		if async {
			return domain.SingleSendResult{
				PhoneNumber: phoneNumber,
				Success:     true,
				Status:      "queued",
			}, nil
		}
		return domain.SingleSendResult{
			PhoneNumber: phoneNumber,
			Success:     true,
			Status:      "sent",
		}, nil
	}

	// Синхронная отправка
	result, err := client.SendTextMessage(context.Background(), "71234567890", "Sync message", false)
	require.NoError(t, err)
	assert.Equal(t, "sent", result.Status)

	// Асинхронная отправка
	result, err = client.SendTextMessage(context.Background(), "71234567890", "Async message", true)
	require.NoError(t, err)
	assert.Equal(t, "queued", result.Status)
}

//go:build integration_test

package integration

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate/types"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"whatsapp-service/internal/entities"
)

// cfgFromEnv собирает конфиг WhatsGate из переменных окружения.
// Если переменные не заданы, тесты помечаются как пропущенные.
func cfgFromEnv(t *testing.T) (*types.WhatsGateConfig, string) {
	err := godotenv.Load("../../../../.env")
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to load .env file: %v", err)
	}
	baseURL := os.Getenv("TEST_WHATSGATE_BASE_URL")
	apiKey := os.Getenv("TEST_WHATSGATE_API_KEY")
	waID := os.Getenv("TEST_WHATSGATE_WHATSAPP_ID")
	phoneNumber := os.Getenv("TEST_WHATSGATE_PHONE_NUMBER")

	if baseURL == "" || apiKey == "" || waID == "" || phoneNumber == "" {
		t.Skip("integration env vars TEST_WHATSGATE_BASE_URL, TEST_WHATSGATE_API_KEY, TEST_WHATSGATE_WHATSAPP_ID, TEST_WHATSGATE_PHONE_NUMBER not set")
	}

	return &types.WhatsGateConfig{
		BaseURL:       baseURL,
		APIKey:        apiKey,
		WhatsappID:    waID,
		Timeout:       10 * time.Second,
		RetryAttempts: 1,
		RetryDelay:    1 * time.Second,
		MaxFileSize:   types.MaxFileSizeBytes,
	}, phoneNumber
}

// TestWhatsGate_SendTextMessage_Live проверяет фактическую отправку текстового сообщения через живой WhatsGate API.
func TestWhatsGate_SendTextMessage_Live(t *testing.T) {
	cfg, phoneNumber := cfgFromEnv(t)
	gw := whatsgate.NewWhatsGateGateway(cfg)

	res, err := gw.SendTextMessage(context.Background(), phoneNumber, "Ping from integration test", false)

	require.NoError(t, err, "SendTextMessage returned an unexpected error")
	require.NotNil(t, res)
	require.True(t, res.Success, "gateway returned failure: %+v", res)
}

// TestWhatsGate_TestConnection_Live проверяет эндпоинт /check реального API.
func TestWhatsGate_TestConnection_Live(t *testing.T) {
	cfg, _ := cfgFromEnv(t)
	gw := whatsgate.NewWhatsGateGateway(cfg)

	res, err := gw.TestConnection(context.Background())

	require.NoError(t, err, "TestConnection returned an unexpected error")
	require.NotNil(t, res)
	require.True(t, res.Success, "connection failed: %+v", res)
}

func TestWhatsGate_SendMediaMessage_Live(t *testing.T) {
	cfg, phoneNumber := cfgFromEnv(t)
	gw := whatsgate.NewWhatsGateGateway(cfg)

	data, err := os.ReadFile("../testdata/test.jpg")
	if os.IsNotExist(err) {
		t.Skip("testdata/test.jpg not found, skipping media test")
	}
	require.NoError(t, err, "failed to read test image")

	res, err := gw.SendMediaMessage(context.Background(), phoneNumber, entities.MessageTypeImage, "Integration photo", "test.jpg", bytes.NewReader(data), "image/jpeg", false)

	require.NoError(t, err, "SendMediaMessage returned an unexpected error")
	require.NotNil(t, res)
	require.True(t, res.Success, "media send failed: %+v", res)
}

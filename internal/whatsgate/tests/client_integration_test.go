package whatsgate_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"whatsapp-service/internal/logger"
	domain "whatsapp-service/internal/whatsgate/domain"
	infra "whatsapp-service/internal/whatsgate/infra"
	usecase "whatsapp-service/internal/whatsgate/usecase"
)

// Для запуска:
// WG_BASE_URL=https://whatsgate.ru/api/v1 TEST_WG_WHATSAPP_ID=... TEST_WG_API_KEY=... TEST_WG_PHONE=... go test -v -tags=integration ./internal/whatsgate

// testEnv содержит параметры окружения для интеграционных тестов WhatsGate API.
type testEnv struct {
	baseURL    string
	whatsappID string
	apiKey     string
	phone      string
}

// loadTestEnv загружает .env и необходимые переменные окружения для интеграционных тестов WhatGate API.
func loadTestEnv(t *testing.T) testEnv {
	err := godotenv.Load("../../.env")
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to load .env file: %v", err)
	}

	env := testEnv{
		baseURL:    os.Getenv("TEST_WG_BASE_URL"),
		whatsappID: os.Getenv("TEST_WG_WHATSAPP_ID"),
		apiKey:     os.Getenv("TEST_WG_API_KEY"),
		phone:      os.Getenv("TEST_WG_PHONE"),
	}

	if env.baseURL == "" || env.whatsappID == "" || env.apiKey == "" || env.phone == "" {
		t.Skip("TEST_WG_BASE_URL, TEST_WG_WHATSAPP_ID, TEST_WG_API_KEY, TEST_WG_PHONE env vars required")
	}
	return env
}

// TestClient_SendTextMessage проверяет отправку текстового сообщения через WhatsGate API.
func TestClient_SendTextMessage(t *testing.T) {
	env := loadTestEnv(t)
	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})
	client := infra.NewClient(env.baseURL, env.whatsappID, env.apiKey, log)
	resp, err := client.SendTextMessage(context.Background(), env.phone, "Integration test text", false)
	if err != nil {
		t.Fatalf("SendTextMessage failed: %v", err)
	}
	t.Logf("Response: %+v", resp)
}

// TestClient_SendMediaMessage проверяет отправку медиа-сообщения через WhatsGate API.
func TestClient_SendMediaMessage(t *testing.T) {
	env := loadTestEnv(t)
	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})
	client := infra.NewClient(env.baseURL, env.whatsappID, env.apiKey, log)

	imageData := []byte("fake-image-data-for-integration-test")
	resp, err := client.SendMediaMessage(context.Background(), env.phone, "image", "Integration test image", "test.jpg", imageData, "image/jpeg", false)
	if err != nil {
		t.Fatalf("SendMediaMessage failed: %v", err)
	}
	t.Logf("Response: %+v", resp)
}

// TestClient_Validation проверяет валидацию входных данных.
func TestClient_Validation(t *testing.T) {
	// Тест валидации номера телефона
	err := infra.ValidatePhoneNumber("invalid-phone")
	require.Error(t, err)

	err = infra.ValidatePhoneNumber("71234567890")
	require.NoError(t, err)

	// Тест валидации типа сообщения
	err = infra.ValidateMessageType("invalid-type")
	require.Error(t, err)

	err = infra.ValidateMessageType("text")
	require.NoError(t, err)
}

// TestClient_WithDatabase проверяет работу клиента с настройками из БД.
func TestClient_WithDatabase(t *testing.T) {
	// Этот тест требует реальной БД PostgreSQL
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:5432/whatsapp_test?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available for testing")
	}
	defer pool.Close()

	repo := infra.NewSettingsRepository(pool, log)
	usecase := usecase.NewSettingsUsecase(repo, log)

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	testSettings := &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://whatsgate.ru/api/v1",
	}

	err = usecase.UpdateSettings(testSettings)
	require.NoError(t, err)

	client, err := usecase.GetClient()
	require.NoError(t, err)

	_, err = client.SendTextMessage(context.Background(), "71234567890", "Integration test message", false)
	// Ожидаем ошибку, так как это тестовые данные
	require.Error(t, err)
}

func TestClientWithContext(t *testing.T) {
	// Этот тест требует реальной БД PostgreSQL
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:5432/whatsapp_test?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available for testing")
	}
	defer pool.Close()

	repo := infra.NewSettingsRepository(pool, log)
	usecase := usecase.NewSettingsUsecase(repo, log)

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	settings := &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://whatsgate.ru/api/v1",
	}

	err = usecase.UpdateSettings(settings)
	require.NoError(t, err)

	client, err := usecase.GetClient()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.SendTextMessage(ctx, "71234567890", "Context test message", false)
	require.Error(t, err)
}

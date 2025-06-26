//go:build integration

package whatsgate_test

import (
	"context"
	"testing"
	"whatsapp-service/internal/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"whatsapp-service/internal/whatsgate/domain"
	whatsgateInfra "whatsapp-service/internal/whatsgate/infra"
	"whatsapp-service/internal/whatsgate/mocks"
	usecase "whatsapp-service/internal/whatsgate/usecase"
)

func TestSettingsService_UpdateSettings_GetSettings(t *testing.T) {
	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})
	repo := &mocks.MockSettingsRepository{}
	usecase := usecase.NewSettingsUsecase(repo, log)

	// Тест 1: Получение настроек по умолчанию
	settings := usecase.GetSettings()
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)

	// Тест 2: Обновление настроек
	testSettings := &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err := usecase.UpdateSettings(testSettings)
	require.NoError(t, err)

	updatedSettings := usecase.GetSettings()
	require.Equal(t, testSettings.WhatsappID, updatedSettings.WhatsappID)
	require.Equal(t, testSettings.APIKey, updatedSettings.APIKey)
	require.Equal(t, testSettings.BaseURL, updatedSettings.BaseURL)

	// Тест 3: Проверка IsConfigured
	require.True(t, usecase.IsConfigured())
}

func TestSettingsService_UpdateSettings_Validation(t *testing.T) {
	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})
	repo := &mocks.MockSettingsRepository{}
	usecase := usecase.NewSettingsUsecase(repo, log)

	// Тест 1: Пустой WhatsappID
	settings := &domain.Settings{
		WhatsappID: "",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err := usecase.UpdateSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "whatsapp_id is required")

	// Тест 2: Пустой APIKey
	settings = &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "",
		BaseURL:    "https://test-api.example.com",
	}

	err = usecase.UpdateSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "api_key is required")
}

func TestSettingsService_GetClient(t *testing.T) {
	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})
	repo := &mocks.MockSettingsRepository{}
	usecase := usecase.NewSettingsUsecase(repo, log)

	// Тест 1: Попытка получить клиент без настроек
	_, err := usecase.GetClient()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured")

	// Тест 2: Получение клиента с настройками
	testSettings := &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err = usecase.UpdateSettings(testSettings)
	require.NoError(t, err)

	client, err := usecase.GetClient()
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestSettingsService_ResetSettings(t *testing.T) {
	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})
	repo := &mocks.MockSettingsRepository{}
	usecase := usecase.NewSettingsUsecase(repo, log)

	testSettings := &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err := usecase.UpdateSettings(testSettings)
	require.NoError(t, err)

	require.True(t, usecase.IsConfigured())

	err = usecase.ResetSettings()
	require.NoError(t, err)

	require.False(t, usecase.IsConfigured())

	settings := usecase.GetSettings()
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)
}

func TestSettingsService_Integration(t *testing.T) {
	// Этот тест требует реальной БД PostgreSQL
	// Для запуска: go test -v -tags=integration ./internal/whatsgate/

	// Пропускаем, если не интеграционный тест
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	log, _ := logger.NewZapLogger(logger.Config{Level: "debug", Format: "console", OutputPath: "stdout"})

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5433/whatsapp_service?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available for testing")
	}
	defer pool.Close()

	repo := whatsgateInfra.NewSettingsRepository(pool, log)
	usecase := usecase.NewSettingsUsecase(repo, log)

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	// Тест 1: Получение настроек по умолчанию
	settings := usecase.GetSettings()
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)

	// Тест 2: Обновление настроек
	testSettings := &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err = usecase.UpdateSettings(testSettings)
	require.NoError(t, err)

	updatedSettings := usecase.GetSettings()
	require.Equal(t, testSettings.WhatsappID, updatedSettings.WhatsappID)
	require.Equal(t, testSettings.APIKey, updatedSettings.APIKey)
	require.Equal(t, testSettings.BaseURL, updatedSettings.BaseURL)

	// Тест 3: Проверка IsConfigured
	require.True(t, usecase.IsConfigured())

	// Тест 4: Сброс настроек
	err = usecase.ResetSettings()
	require.NoError(t, err)

	require.False(t, usecase.IsConfigured())

	settings = usecase.GetSettings()
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)
}

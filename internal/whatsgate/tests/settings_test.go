//go:build integration

package whatsgate_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"
	whatsgateInfra "whatsapp-service/internal/whatsgate/infra"
	"whatsapp-service/internal/whatsgate/interfaces"
	"whatsapp-service/internal/whatsgate/mocks"
)

func TestSettingsService_UpdateSettings_GetSettings(t *testing.T) {
	repo := &mocks.MockSettingsRepository{}
	service := whatsgateDomain.NewSettingsService(repo)

	// Тест 1: Получение настроек по умолчанию
	settings := service.GetSettings()
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)

	// Тест 2: Обновление настроек
	testSettings := &interfaces.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err := service.UpdateSettings(testSettings)
	require.NoError(t, err)

	updatedSettings := service.GetSettings()
	require.Equal(t, testSettings.WhatsappID, updatedSettings.WhatsappID)
	require.Equal(t, testSettings.APIKey, updatedSettings.APIKey)
	require.Equal(t, testSettings.BaseURL, updatedSettings.BaseURL)

	// Тест 3: Проверка IsConfigured
	require.True(t, service.IsConfigured())
}

func TestSettingsService_UpdateSettings_Validation(t *testing.T) {
	repo := &mocks.MockSettingsRepository{}
	service := whatsgateDomain.NewSettingsService(repo)

	// Тест 1: Пустой WhatsappID
	settings := &interfaces.Settings{
		WhatsappID: "",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err := service.UpdateSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "whatsapp_id is required")

	// Тест 2: Пустой APIKey
	settings = &interfaces.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "",
		BaseURL:    "https://test-api.example.com",
	}

	err = service.UpdateSettings(settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "api_key is required")
}

func TestSettingsService_GetClient(t *testing.T) {
	repo := &mocks.MockSettingsRepository{}
	service := whatsgateDomain.NewSettingsService(repo)

	// Тест 1: Попытка получить клиент без настроек
	_, err := service.GetClient()
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured")

	// Тест 2: Получение клиента с настройками
	testSettings := &interfaces.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err = service.UpdateSettings(testSettings)
	require.NoError(t, err)

	client, err := service.GetClient()
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestSettingsService_ResetSettings(t *testing.T) {
	repo := &mocks.MockSettingsRepository{}
	service := whatsgateDomain.NewSettingsService(repo)

	testSettings := &interfaces.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err := service.UpdateSettings(testSettings)
	require.NoError(t, err)

	require.True(t, service.IsConfigured())

	err = service.ResetSettings()
	require.NoError(t, err)

	require.False(t, service.IsConfigured())

	settings := service.GetSettings()
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

	// Подключение к тестовой БД
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5433/whatsapp_service?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available for testing")
	}
	defer pool.Close()

	repo := whatsgateInfra.NewSettingsRepository(pool)
	service := whatsgateDomain.NewSettingsService(repo)

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	// Тест 1: Получение настроек по умолчанию
	settings := service.GetSettings()
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)

	// Тест 2: Обновление настроек
	testSettings := &interfaces.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err = service.UpdateSettings(testSettings)
	require.NoError(t, err)

	updatedSettings := service.GetSettings()
	require.Equal(t, testSettings.WhatsappID, updatedSettings.WhatsappID)
	require.Equal(t, testSettings.APIKey, updatedSettings.APIKey)
	require.Equal(t, testSettings.BaseURL, updatedSettings.BaseURL)

	// Тест 3: Проверка IsConfigured
	require.True(t, service.IsConfigured())

	// Тест 4: Сброс настроек
	err = service.ResetSettings()
	require.NoError(t, err)

	require.False(t, service.IsConfigured())

	settings = service.GetSettings()
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)
}

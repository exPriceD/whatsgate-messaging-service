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
)

// MockPoolDB создает мок для тестирования
type MockPoolDB struct {
	pool *pgxpool.Pool
}

func (m *MockPoolDB) GetPool() *pgxpool.Pool {
	return m.pool
}

func TestDatabaseSettingsStorage_Integration(t *testing.T) {
	// Этот тест требует реальной БД PostgreSQL
	// Для запуска: go test -v -tags=integration ./internal/whatsgate/
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

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	storage := whatsgateInfra.NewDatabaseSettingsStorage(repo)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	// Тест 1: Загрузка настроек из пустой БД
	settings, err := storage.Load()
	require.NoError(t, err)
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)

	// Тест 2: Сохранение настроек
	testSettings := &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err = storage.Save(testSettings)
	require.NoError(t, err)

	require.True(t, storage.IsConfigured())

	loadedSettings, err := storage.Load()
	require.NoError(t, err)
	require.Equal(t, testSettings.WhatsappID, loadedSettings.WhatsappID)
	require.Equal(t, testSettings.APIKey, loadedSettings.APIKey)
	require.Equal(t, testSettings.BaseURL, loadedSettings.BaseURL)

	// Тест 3: Обновление настроек
	updatedSettings := &domain.Settings{
		WhatsappID: "updated-whatsapp-id",
		APIKey:     "updated-api-key",
		BaseURL:    "https://updated-api.example.com",
	}

	err = storage.Save(updatedSettings)
	require.NoError(t, err)

	loadedSettings, err = storage.Load()
	require.NoError(t, err)
	require.Equal(t, updatedSettings.WhatsappID, loadedSettings.WhatsappID)
	require.Equal(t, updatedSettings.APIKey, loadedSettings.APIKey)
	require.Equal(t, updatedSettings.BaseURL, loadedSettings.BaseURL)

	// Тест 4: Удаление настроек
	err = storage.Delete()
	require.NoError(t, err)

	require.False(t, storage.IsConfigured())

	settings, err = storage.Load()
	require.NoError(t, err)
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)
}

func TestDatabaseSettingsStorage_EmptyDatabase(t *testing.T) {
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

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	storage := whatsgateInfra.NewDatabaseSettingsStorage(repo)

	// Тест: Загрузка из пустой БД
	settings, err := storage.Load()
	require.NoError(t, err)
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)
	require.False(t, storage.IsConfigured())
}

func TestDatabaseSettingsStorage_ConcurrentAccess(t *testing.T) {
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

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	storage := whatsgateInfra.NewDatabaseSettingsStorage(repo)

	testSettings := &domain.Settings{
		WhatsappID: "concurrent-test-id",
		APIKey:     "concurrent-test-key",
		BaseURL:    "https://concurrent-test.example.com",
	}

	err = storage.Save(testSettings)
	require.NoError(t, err)

	// Тест: Конкурентный доступ
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			settings, err := storage.Load()
			require.NoError(t, err)
			require.Equal(t, testSettings.WhatsappID, settings.WhatsappID)
			done <- true
		}()
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSettingsRepository_Load_EmptyDB(t *testing.T) {
	// Этот тест требует реальной БД PostgreSQL
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

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	settings, err := repo.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, "", settings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", settings.BaseURL)

	require.False(t, repo.IsConfigured(ctx))
}

func TestSettingsRepository_Save_Update(t *testing.T) {
	// Этот тест требует реальной БД PostgreSQL
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

	err = repo.InitTable(ctx)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "DELETE FROM whatsgate_settings")
	require.NoError(t, err)

	testSettings := &domain.Settings{
		WhatsappID: "test-whatsapp-id",
		APIKey:     "test-api-key",
		BaseURL:    "https://test-api.example.com",
	}

	err = repo.Save(ctx, testSettings)
	require.NoError(t, err)

	require.True(t, repo.IsConfigured(ctx))

	loadedSettings, err := repo.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, testSettings.WhatsappID, loadedSettings.WhatsappID)
	require.Equal(t, testSettings.APIKey, loadedSettings.APIKey)
	require.Equal(t, testSettings.BaseURL, loadedSettings.BaseURL)

	updatedSettings := &domain.Settings{
		WhatsappID: "new-whatsapp-id",
		APIKey:     "new-api-key",
		BaseURL:    "https://new-api.example.com",
	}

	err = repo.Save(ctx, updatedSettings)
	require.NoError(t, err)

	reloadedSettings, err := repo.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, updatedSettings.WhatsappID, reloadedSettings.WhatsappID)
	require.Equal(t, updatedSettings.APIKey, reloadedSettings.APIKey)
	require.Equal(t, updatedSettings.BaseURL, reloadedSettings.BaseURL)

	history, err := repo.GetSettingsHistory(ctx)
	require.NoError(t, err)
	require.Len(t, history, 2) // Два сохранения

	err = repo.Delete(ctx)
	require.NoError(t, err)

	require.False(t, repo.IsConfigured(ctx))

	deletedSettings, err := repo.Load(ctx)
	require.NoError(t, err)
	require.Equal(t, "", deletedSettings.WhatsappID)
	require.Equal(t, "https://whatsgate.ru/api/v1", deletedSettings.BaseURL)
}

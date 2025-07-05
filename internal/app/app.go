package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"whatsapp-service/internal/adapters/converter"
	"whatsapp-service/internal/adapters/presenters"
	"whatsapp-service/internal/config"
	"whatsapp-service/internal/delivery/http"
	"whatsapp-service/internal/delivery/http/handlers"
	"whatsapp-service/internal/infrastructure/database/postgres"
	"whatsapp-service/internal/infrastructure/dispatcher/messaging"
	"whatsapp-service/internal/infrastructure/gateways/whatsapp/dynamic/whatsgate"
	gtypes "whatsapp-service/internal/infrastructure/gateways/whatsapp/whatsgate/types"
	zaplogger "whatsapp-service/internal/infrastructure/logger/zap"
	"whatsapp-service/internal/infrastructure/parsers/excel"
	"whatsapp-service/internal/infrastructure/registry"
	campaignRepository "whatsapp-service/internal/infrastructure/repositories/campaign"
	settingsRepository "whatsapp-service/internal/infrastructure/repositories/settings"
	"whatsapp-service/internal/infrastructure/services/ratelimiter"
	"whatsapp-service/internal/shared/logger"
	campaignInteractor "whatsapp-service/internal/usecases/campaigns/interactor"
	"whatsapp-service/internal/usecases/campaigns/ports"
	settingsInteractor "whatsapp-service/internal/usecases/settings/interactor"
)

// App инкапсулирует все зависимости и умеет запускаться/останавливаться.
type App struct {
	cfg        *config.Config
	logger     logger.Logger
	db         *pgxpool.Pool
	server     *http.HTTPServer
	dispatcher ports.Dispatcher
}

// New собирает приложение из конфигурации.
func New(cfg *config.Config) (*App, error) {
	// 1. Логгер
	sharedLogger, err := zaplogger.New(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	// 2. Создаем отдельный zap.Logger для компонентов, которые требуют именно zap
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("init zap logger: %w", err)
	}

	// 3. БД
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := postgres.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	// 4. Репозитории
	campaignRepo := campaignRepository.NewPostgresCampaignRepository(pool)
	campaignStatusRepo := campaignRepository.NewPostgresCampaignStatusRepository(pool)
	settingsRepo := settingsRepository.NewPostgresWhatsGateSettingsRepository(pool)

	// 5. Утилитарные сервисы
	globalRateLimiter := ratelimiter.NewGlobalMemoryRateLimiter()
	fileParser := excel.NewExcelParser()
	messageGateway := whatsgate.NewSettingsAwareGateway(settingsRepo)
	dispatcherSvc := messaging.NewDispatcher(messageGateway, globalRateLimiter, zapLogger)
	campaignRegistry := registry.NewInMemoryCampaignRegistry()

	// 6. Настройка WhatsGate
	_ = initWhatsGateConfig(ctx, settingsRepo) // используем для инициализации

	// 7. Use Cases
	campaignUseCase := campaignInteractor.NewCampaignInteractor(
		campaignRepo,
		campaignStatusRepo,
		dispatcherSvc,
		campaignRegistry,
		fileParser,
		sharedLogger,
	)

	settingsUseCase := settingsInteractor.NewService(settingsRepo)

	// 8. Конвертеры
	campaignConverter := converter.NewCampaignConverter()
	settingsConverter := converter.NewSettingsConverter()

	// 9. Presenters
	campaignPresenter := presenters.NewCampaignPresenter(campaignConverter)
	settingsPresenter := presenters.NewSettingsPresenter(settingsConverter)

	// 10. Handlers
	campaignHandler := handlers.NewCampaignsHandler(
		campaignUseCase,
		campaignPresenter,
		campaignConverter,
	)

	settingsHandler := handlers.NewSettingsHandler(
		settingsUseCase,
		settingsPresenter,
		settingsConverter,
	)

	// 11. Health Handler
	healthHandler := handlers.NewHealthHandler(
		sharedLogger,
		campaignRepo,
		dispatcherSvc,
	)

	// 12. HTTP Server
	httpSrv := createHTTPServer(cfg.HTTP.Port, campaignHandler, settingsHandler, healthHandler, sharedLogger)

	return &App{
		cfg:        cfg,
		logger:     sharedLogger,
		db:         pool,
		server:     httpSrv,
		dispatcher: dispatcherSvc,
	}, nil
}

// createHTTPServer создает HTTP сервер с новыми handlers
func createHTTPServer(
	port int,
	campaignHandler *handlers.CampaignsHandler,
	settingsHandler *handlers.SettingsHandler,
	healthHandler *handlers.HealthHandler,
	logger logger.Logger,
) *http.HTTPServer {
	return http.NewHTTPServer(
		port,
		campaignHandler,
		settingsHandler,
		healthHandler,
		logger,
	)
}

// Start запускает HTTP сервер и фоновые процессы.
func (a *App) Start(ctx context.Context) error {
	a.dispatcher.Start(ctx)
	a.logger.Info("Dispatcher started")

	a.logger.Info("HTTP server starting", "port", a.cfg.HTTP.Port)
	return a.server.Start()
}

// Stop останавливает HTTP сервер и закрывает ресурсы.
func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("Stopping application...")

	if err := a.dispatcher.Stop(ctx); err != nil {
		a.logger.Error("failed to stop dispatcher gracefully", "error", err)
		// Не возвращаем ошибку, чтобы продолжить остановку
	} else {
		a.logger.Info("Dispatcher stopped")
	}

	if err := a.server.Stop(ctx); err != nil {
		return err
	}
	a.logger.Info("HTTP server stopped")

	postgres.Close(a.db)
	a.logger.Info("Database pool closed")

	return nil
}

// initWhatsGateConfig инициализирует конфигурацию WhatsGate
func initWhatsGateConfig(ctx context.Context, settingsRepo *settingsRepository.PostgresWhatsGateSettingsRepository) *gtypes.WhatsGateConfig {
	// Получаем сохраненные настройки, если есть
	stored, _ := settingsRepo.Get(ctx) // игнорируем ошибку -> считаем как отсутствующие

	config := &gtypes.WhatsGateConfig{
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		MaxFileSize:   gtypes.MaxFileSizeBytes,
	}

	if stored != nil {
		config.BaseURL = stored.BaseURL()
		config.APIKey = stored.APIKey()
		config.WhatsappID = stored.WhatsappID()
	} else {
		// Fallback к переменным окружения
		config.BaseURL = getenv("WHATSGATE_URL", "http://localhost:3000")
		config.APIKey = getenv("WHATSGATE_API_KEY", "demo-key")
		config.WhatsappID = getenv("WHATSAPP_ID", "demo-whatsapp")
	}

	return config
}

// getenv возвращает значение переменной окружения или значение по умолчанию
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

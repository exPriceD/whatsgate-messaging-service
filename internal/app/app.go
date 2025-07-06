package app

import (
	"context"
	"fmt"
	"time"

	"whatsapp-service/internal/adapters/converter"
	"whatsapp-service/internal/adapters/presenters"
	"whatsapp-service/internal/config"
	"whatsapp-service/internal/delivery/http"
	"whatsapp-service/internal/delivery/http/handlers"
	"whatsapp-service/internal/infrastructure/database/postgres"
	"whatsapp-service/internal/infrastructure/dispatcher/messaging"
	messagingPorts "whatsapp-service/internal/infrastructure/dispatcher/messaging/ports"
	"whatsapp-service/internal/infrastructure/gateways/whatsapp/dynamic/whatsgate"
	zaplogger "whatsapp-service/internal/infrastructure/logger/zap"
	"whatsapp-service/internal/infrastructure/parsers/excel"
	"whatsapp-service/internal/infrastructure/registry"
	campaignRepository "whatsapp-service/internal/infrastructure/repositories/campaign"
	settingsRepository "whatsapp-service/internal/infrastructure/repositories/settings"
	"whatsapp-service/internal/infrastructure/services/ratelimiter"
	"whatsapp-service/internal/shared/logger"
	campaignInteractor "whatsapp-service/internal/usecases/campaigns/interactor"
	campaignInterfaces "whatsapp-service/internal/usecases/campaigns/interfaces"
	"whatsapp-service/internal/usecases/campaigns/ports"
	settingsInteractor "whatsapp-service/internal/usecases/settings/interactor"
	settingsInterfaces "whatsapp-service/internal/usecases/settings/interfaces"
	settingsPorts "whatsapp-service/internal/usecases/settings/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Infrastructure содержит все инфраструктурные зависимости
type Infrastructure struct {
	Database           *pgxpool.Pool
	Logger             logger.Logger
	CampaignRepo       ports.CampaignRepository
	CampaignStatusRepo ports.CampaignStatusRepository
	SettingsRepo       settingsPorts.WhatsGateSettingsRepository
	FileParser         ports.FileParser
	MessageGateway     messagingPorts.MessageGateway
	GlobalRateLimiter  messagingPorts.GlobalRateLimiter
	Dispatcher         ports.Dispatcher
	CampaignRegistry   ports.CampaignRegistry
}

// UseCases содержит все use case зависимости
type UseCases struct {
	Campaign campaignInterfaces.CampaignUseCase
	Settings settingsInterfaces.SettingsUseCase
}

// Adapters содержит все адаптеры (конвертеры и презентеры)
type Adapters struct {
	CampaignConverter converter.CampaignConverter
	SettingsConverter converter.SettingsConverter
	CampaignPresenter presenters.CampaignPresenterInterface
	SettingsPresenter presenters.SettingsPresenterInterface
}

// Handlers содержит все HTTP обработчики
type Handlers struct {
	Campaign *handlers.CampaignsHandler
	Settings *handlers.SettingsHandler
	Health   *handlers.HealthHandler
}

// App инкапсулирует все зависимости и умеет запускаться/останавливаться.
type App struct {
	cfg            *config.Config
	infrastructure *Infrastructure
	server         *http.HTTPServer
}

// NewInfrastructure создает все инфраструктурные зависимости
func NewInfrastructure(cfg *config.Config) (*Infrastructure, error) {
	// Логгер
	sharedLogger, err := zaplogger.New(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	// БД
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := postgres.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	// Репозитории
	var campaignRepo ports.CampaignRepository = campaignRepository.NewPostgresCampaignRepository(pool, sharedLogger)
	var campaignStatusRepo ports.CampaignStatusRepository = campaignRepository.NewPostgresCampaignStatusRepository(pool, sharedLogger)
	var settingsRepo settingsPorts.WhatsGateSettingsRepository = settingsRepository.NewPostgresWhatsGateSettingsRepository(pool, sharedLogger)

	// Утилитарные сервисы
	var globalRateLimiter messagingPorts.GlobalRateLimiter = ratelimiter.NewGlobalMemoryRateLimiter()
	var fileParser ports.FileParser = excel.NewExcelParser()
	var messageGateway messagingPorts.MessageGateway = whatsgate.NewSettingsAwareGateway(settingsRepo)
	var dispatcherSvc ports.Dispatcher = messaging.NewDispatcher(messageGateway, globalRateLimiter, sharedLogger)
	var campaignRegistry ports.CampaignRegistry = registry.NewInMemoryCampaignRegistry()

	return &Infrastructure{
		Database:           pool,
		Logger:             sharedLogger,
		CampaignRepo:       campaignRepo,
		CampaignStatusRepo: campaignStatusRepo,
		SettingsRepo:       settingsRepo,
		FileParser:         fileParser,
		MessageGateway:     messageGateway,
		GlobalRateLimiter:  globalRateLimiter,
		Dispatcher:         dispatcherSvc,
		CampaignRegistry:   campaignRegistry,
	}, nil
}

// NewUseCases создает все use case зависимости
func NewUseCases(infra *Infrastructure) *UseCases {
	// Use Cases
	var campaignUseCase campaignInterfaces.CampaignUseCase = campaignInteractor.NewCampaignInteractor(
		infra.CampaignRepo,
		infra.CampaignStatusRepo,
		infra.Dispatcher,
		infra.CampaignRegistry,
		infra.FileParser,
		infra.Logger,
	)

	var settingsUseCase settingsInterfaces.SettingsUseCase = settingsInteractor.NewService(
		infra.SettingsRepo,
		infra.Logger,
	)

	return &UseCases{
		Campaign: campaignUseCase,
		Settings: settingsUseCase,
	}
}

// NewAdapters создает все адаптеры (конвертеры и презентеры)
func NewAdapters() *Adapters {
	// Конвертеры
	var campaignConverter converter.CampaignConverter = converter.NewCampaignConverter()
	var settingsConverter converter.SettingsConverter = converter.NewSettingsConverter()

	// Presenters
	var campaignPresenter presenters.CampaignPresenterInterface = presenters.NewCampaignPresenter(campaignConverter)
	var settingsPresenter presenters.SettingsPresenterInterface = presenters.NewSettingsPresenter(settingsConverter)

	return &Adapters{
		CampaignConverter: campaignConverter,
		SettingsConverter: settingsConverter,
		CampaignPresenter: campaignPresenter,
		SettingsPresenter: settingsPresenter,
	}
}

// NewHandlers создает все HTTP обработчики
func NewHandlers(useCases *UseCases, adapters *Adapters, infra *Infrastructure) *Handlers {
	// Handlers
	campaignHandler := handlers.NewCampaignsHandler(
		useCases.Campaign,
		adapters.CampaignPresenter,
		adapters.CampaignConverter,
		infra.Logger,
	)

	settingsHandler := handlers.NewSettingsHandler(
		useCases.Settings,
		adapters.SettingsPresenter,
		adapters.SettingsConverter,
		infra.Logger,
	)

	// Health Handler
	healthHandler := handlers.NewHealthHandler(
		infra.Logger,
		infra.CampaignRepo,
		infra.Dispatcher,
	)

	return &Handlers{
		Campaign: campaignHandler,
		Settings: settingsHandler,
		Health:   healthHandler,
	}
}

// New собирает приложение из конфигурации.
func New(cfg *config.Config) (*App, error) {
	// Инфраструктура
	infra, err := NewInfrastructure(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Use cases
	useCases := NewUseCases(infra)

	// Адаптеры
	adapters := NewAdapters()

	// Handlers
	handlerSet := NewHandlers(useCases, adapters, infra)

	// HTTP Server
	httpSrv := createHTTPServer(cfg.HTTP.Port, handlerSet.Campaign, handlerSet.Settings, handlerSet.Health, infra.Logger)

	return &App{
		cfg:            cfg,
		infrastructure: infra,
		server:         httpSrv,
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
	a.infrastructure.Dispatcher.Start(ctx)
	a.infrastructure.Logger.Info("Dispatcher started")

	a.infrastructure.Logger.Info("HTTP server starting", "port", a.cfg.HTTP.Port)
	return a.server.Start()
}

// Stop останавливает HTTP сервер и закрывает ресурсы.
func (a *App) Stop(ctx context.Context) error {
	a.infrastructure.Logger.Info("Stopping application...")

	if err := a.infrastructure.Dispatcher.Stop(ctx); err != nil {
		a.infrastructure.Logger.Error("failed to stop dispatcher gracefully", "error", err)
	} else {
		a.infrastructure.Logger.Info("Dispatcher stopped")
	}

	if err := a.server.Stop(ctx); err != nil {
		return err
	}
	a.infrastructure.Logger.Info("HTTP server stopped")

	postgres.Close(a.infrastructure.Database)
	a.infrastructure.Logger.Info("Database pool closed")

	return nil
}

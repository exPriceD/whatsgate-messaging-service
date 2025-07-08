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
	Database          *pgxpool.Pool
	Logger            logger.Logger
	CampaignRepo      ports.CampaignRepository
	SettingsRepo      settingsPorts.WhatsGateSettingsRepository
	FileParser        ports.FileParser
	MessageGateway    messagingPorts.MessageGateway
	GlobalRateLimiter messagingPorts.GlobalRateLimiter
	Dispatcher        ports.Dispatcher
	CampaignRegistry  ports.CampaignRegistry
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
	var settingsRepo settingsPorts.WhatsGateSettingsRepository = settingsRepository.NewPostgresWhatsGateSettingsRepository(pool, sharedLogger)
	var _ = settingsRepository.NewPostgresRetailCRMSettingsRepository(pool, sharedLogger)

	// Утилитарные сервисы
	var globalRateLimiter messagingPorts.GlobalRateLimiter = ratelimiter.NewGlobalMemoryRateLimiter()
	var fileParser ports.FileParser = excel.NewExcelParser()
	var messageGateway messagingPorts.MessageGateway = whatsgate.NewSettingsAwareGateway(settingsRepo)
	var dispatcherSvc ports.Dispatcher = messaging.NewDispatcher(messageGateway, globalRateLimiter, sharedLogger)
	var campaignRegistry ports.CampaignRegistry = registry.NewInMemoryCampaignRegistry()

	return &Infrastructure{
		Database:          pool,
		Logger:            sharedLogger,
		CampaignRepo:      campaignRepo,
		SettingsRepo:      settingsRepo,
		FileParser:        fileParser,
		MessageGateway:    messageGateway,
		GlobalRateLimiter: globalRateLimiter,
		Dispatcher:        dispatcherSvc,
		CampaignRegistry:  campaignRegistry,
	}, nil
}

// NewUseCases создает все use case зависимости
func NewUseCases(infra *Infrastructure) *UseCases {
	// Use Cases
	var campaignUseCase campaignInterfaces.CampaignUseCase = campaignInteractor.NewCampaignInteractor(
		infra.CampaignRepo,
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

	// Use Cases
	useCases := NewUseCases(infra)

	// Adapters
	adapters := NewAdapters()

	// Handlers
	handlers := NewHandlers(useCases, adapters, infra)

	// HTTP сервер
	httpSrv := createHTTPServer(
		cfg.HTTP.Port,
		handlers.Campaign,
		handlers.Settings,
		handlers.Health,
		infra.Logger,
	)

	return &App{
		cfg:            cfg,
		infrastructure: infra,
		server:         httpSrv,
	}, nil
}

// createHTTPServer создает HTTP сервер
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

// Start запускает приложение
func (a *App) Start(ctx context.Context) error {
	// Запускаем диспетчер
	a.infrastructure.Logger.Info("starting dispatcher")
	a.infrastructure.Dispatcher.Start(ctx)

	// Восстанавливаем "orphaned" кампании при старте
	if err := a.recoverOrphanedCampaigns(ctx); err != nil {
		a.infrastructure.Logger.Error("failed to recover orphaned campaigns", "error", err)
		// Не останавливаем приложение из-за этой ошибки
	}

	a.infrastructure.Logger.Info("HTTP server starting", "port", a.cfg.HTTP.Port)
	return a.server.Start()
}

// Stop останавливает приложение
func (a *App) Stop(ctx context.Context) error {
	a.infrastructure.Logger.Info("stopping application")

	// Graceful shutdown кампаний
	if err := a.gracefulShutdownCampaigns(ctx); err != nil {
		a.infrastructure.Logger.Error("failed to gracefully shutdown campaigns", "error", err)
	}

	// Останавливаем диспетчер
	a.infrastructure.Logger.Info("stopping dispatcher")
	if err := a.infrastructure.Dispatcher.Stop(ctx); err != nil {
		a.infrastructure.Logger.Error("failed to stop dispatcher", "error", err)
	}

	if err := a.server.Stop(ctx); err != nil {
		return err
	}

	// Закрываем пул соединений
	if a.infrastructure.Database != nil {
		a.infrastructure.Database.Close()
	}

	a.infrastructure.Logger.Info("application stopped")
	return nil
}

// gracefulShutdownCampaigns останавливает активные кампании
func (a *App) gracefulShutdownCampaigns(ctx context.Context) error {
	a.infrastructure.Logger.Info("gracefully shutting down campaigns")

	// Получаем все активные кампании
	activeCampaigns, err := a.infrastructure.CampaignRepo.GetActiveCampaigns(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active campaigns: %w", err)
	}

	// Останавливаем каждую кампанию
	for _, campaign := range activeCampaigns {
		if err := a.updateCampaignStatusOnShutdown(ctx, campaign.ID()); err != nil {
			a.infrastructure.Logger.Error("failed to update campaign status on shutdown",
				"campaign_id", campaign.ID(), "error", err)
		}
	}

	a.infrastructure.Logger.Info("campaigns shutdown completed", "count", len(activeCampaigns))
	return nil
}

// updateCampaignStatusOnShutdown обновляет статус кампании при остановке
func (a *App) updateCampaignStatusOnShutdown(ctx context.Context, campaignID string) error {
	// Получаем текущую кампанию
	campaign, err := a.infrastructure.CampaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}

	// Если кампания активна, помечаем ее как остановленную
	if campaign.Status() == "started" {
		if err := a.infrastructure.CampaignRepo.UpdateStatus(ctx, campaignID, "stopped"); err != nil {
			return fmt.Errorf("failed to update campaign status: %w", err)
		}
		a.infrastructure.Logger.Info("campaign marked as stopped", "campaign_id", campaignID)
	}

	return nil
}

// recoverOrphanedCampaigns восстанавливает "orphaned" кампании
func (a *App) recoverOrphanedCampaigns(ctx context.Context) error {
	a.infrastructure.Logger.Info("recovering orphaned campaigns")

	// Получаем все кампании со статусом "started"
	startedCampaigns, err := a.infrastructure.CampaignRepo.ListByStatus(ctx, "started", 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get started campaigns: %w", err)
	}

	// Помечаем их как "stopped" так как приложение перезапускается
	for _, campaign := range startedCampaigns {
		if err := a.infrastructure.CampaignRepo.UpdateStatus(ctx, campaign.ID(), "stopped"); err != nil {
			a.infrastructure.Logger.Error("failed to mark orphaned campaign as stopped",
				"campaign_id", campaign.ID(), "error", err)
		} else {
			a.infrastructure.Logger.Info("orphaned campaign marked as stopped", "campaign_id", campaign.ID())
		}
	}

	a.infrastructure.Logger.Info("orphaned campaigns recovery completed", "count", len(startedCampaigns))
	return nil
}

package app

import (
	"context"
	"fmt"
	"time"
	campaignRepository "whatsapp-service/internal/entities/campaign/repository"
	settingsRepository "whatsapp-service/internal/entities/settings/repository"
	"whatsapp-service/internal/interfaces"

	"whatsapp-service/internal/adapters/converter"
	"whatsapp-service/internal/adapters/presenters"
	"whatsapp-service/internal/config"
	"whatsapp-service/internal/delivery/http"
	"whatsapp-service/internal/delivery/http/handlers"
	"whatsapp-service/internal/infrastructure/database/postgres"
	"whatsapp-service/internal/infrastructure/dispatcher/messaging"
	"whatsapp-service/internal/infrastructure/gateways/retailcrm/client"
	retailcrmPorts "whatsapp-service/internal/infrastructure/gateways/retailcrm/ports"
	retailcrmService "whatsapp-service/internal/infrastructure/gateways/retailcrm/service"
	"whatsapp-service/internal/infrastructure/gateways/whatsapp/dynamic/whatsgate"
	zaplogger "whatsapp-service/internal/infrastructure/logger/zap"
	"whatsapp-service/internal/infrastructure/parsers/excel"
	"whatsapp-service/internal/infrastructure/registry"
	campaignRepositoryImpl "whatsapp-service/internal/infrastructure/repositories/campaign"
	settingsRepositoryImpl "whatsapp-service/internal/infrastructure/repositories/settings"
	"whatsapp-service/internal/infrastructure/services/ratelimiter"
	campaignInteractor "whatsapp-service/internal/usecases/campaigns/interactor"
	campaignInterfaces "whatsapp-service/internal/usecases/campaigns/interfaces"
	campaignPorts "whatsapp-service/internal/usecases/campaigns/ports"
	messagingInteractor "whatsapp-service/internal/usecases/messaging/interactor"
	messagingInterfaces "whatsapp-service/internal/usecases/messaging/interfaces"
	retailcrmInteractor "whatsapp-service/internal/usecases/retailcrm/interactor"
	retailcrmInterfaces "whatsapp-service/internal/usecases/retailcrm/interfaces"
	settingsInteractor "whatsapp-service/internal/usecases/settings/interactor"
	settingsInterfaces "whatsapp-service/internal/usecases/settings/interfaces"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Infrastructure содержит все инфраструктурные зависимости
type Infrastructure struct {
	Database              *pgxpool.Pool
	Logger                interfaces.Logger
	CampaignRepo          campaignRepository.CampaignRepository
	WhatsgateSettingsRepo settingsRepository.WhatsGateSettingsRepository
	RetailCRMSettingsRepo settingsRepository.RetailCRMSettingsRepository
	FileParser            campaignPorts.FileParser
	MessageGateway        interfaces.MessageGateway
	GlobalRateLimiter     messaging.GlobalRateLimiter
	Dispatcher            campaignPorts.Dispatcher
	CampaignRegistry      campaignPorts.CampaignRegistry
	RetailCRMGateway      retailcrmPorts.RetailCRMGateway
}

// UseCases содержит все use case зависимости
type UseCases struct {
	Campaign          campaignInterfaces.CampaignUseCase
	WhatsgateSettings settingsInterfaces.WhatsgateSettingsUseCase
	RetailCRMSettings settingsInterfaces.RetailCRMSettingsUseCase
	Message           messagingInterfaces.MessageUseCase
	RetailCRM         retailcrmInterfaces.RetailCRMUseCase
}

// Adapters содержит все адаптеры (конвертеры и презентеры)
type Adapters struct {
	CampaignConverter          converter.CampaignConverter
	WhatsgateSettingsConverter converter.WhatsgateSettingsConverter
	RetailCRMSettingsConverter converter.RetailCRMSettingsConverter
	MessagingConverter         converter.MessagingConverter
	RetailCRMConverter         converter.RetailCRMConverter
	CampaignPresenter          presenters.CampaignPresenterInterface
	WhatsgateSettingsPresenter presenters.WhatsgateSettingsPresenterInterface
	RetailCRMSettingsPresenter presenters.RetailCRMSettingsPresenterInterface
	MessagingPresenter         presenters.MessagingPresenterInterface
	RetailCRMPresenter         presenters.RetailCRMPresenterInterface
}

// Handlers содержит все HTTP обработчики
type Handlers struct {
	Campaign          *handlers.CampaignsHandler
	WhatsgateSettings *handlers.WhatsgateSettingsHandler
	RetailCRMSettings *handlers.RetailCRMSettingsHandler
	Messaging         *handlers.MessagingHandler
	Health            *handlers.HealthHandler
	RetailCRM         *handlers.RetailCRMHandler
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
	var campaignRepo campaignRepository.CampaignRepository = campaignRepositoryImpl.NewPostgresCampaignRepository(pool, sharedLogger)
	var whatsgateSettingsRepo settingsRepository.WhatsGateSettingsRepository = settingsRepositoryImpl.NewPostgresWhatsGateSettingsRepository(pool, sharedLogger)
	var retailCRMSettingsRepo settingsRepository.RetailCRMSettingsRepository = settingsRepositoryImpl.NewPostgresRetailCRMSettingsRepository(pool, sharedLogger)

	// Утилитарные сервисы
	var globalRateLimiter messaging.GlobalRateLimiter = ratelimiter.NewGlobalMemoryRateLimiter()
	var fileParser campaignPorts.FileParser = excel.NewExcelParser()
	var messageGateway interfaces.MessageGateway = whatsgate.NewSettingsAwareGateway(whatsgateSettingsRepo)
	var dispatcherSvc campaignPorts.Dispatcher = messaging.NewDispatcher(messageGateway, globalRateLimiter, sharedLogger)
	var campaignRegistry campaignPorts.CampaignRegistry = registry.NewInMemoryCampaignRegistry()

	// RetailCRM сервис
	var retailCRMClient client.RetailCRMClientInterface = client.NewSettingsAwareRetailCRMClient(retailCRMSettingsRepo, sharedLogger)
	var retailCRMGateway retailcrmPorts.RetailCRMGateway = retailcrmService.NewRetailCRMService(retailCRMClient, sharedLogger, &cfg.RetailCRM)

	return &Infrastructure{
		Database:              pool,
		Logger:                sharedLogger,
		CampaignRepo:          campaignRepo,
		WhatsgateSettingsRepo: whatsgateSettingsRepo,
		RetailCRMSettingsRepo: retailCRMSettingsRepo,
		FileParser:            fileParser,
		MessageGateway:        messageGateway,
		GlobalRateLimiter:     globalRateLimiter,
		Dispatcher:            dispatcherSvc,
		CampaignRegistry:      campaignRegistry,
		RetailCRMGateway:      retailCRMGateway,
	}, nil
}

// NewUseCases создает все use case зависимости
func NewUseCases(infra *Infrastructure) *UseCases {
	// Сначала создаем RetailCRM usecase
	var retailCRMUseCase retailcrmInterfaces.RetailCRMUseCase = retailcrmInteractor.NewRetailCRMInteractor(
		infra.RetailCRMGateway,
		infra.Logger,
	)

	// Use Cases
	var campaignUseCase campaignInterfaces.CampaignUseCase = campaignInteractor.NewCampaignInteractor(
		infra.CampaignRepo,
		infra.Dispatcher,
		infra.CampaignRegistry,
		infra.FileParser,
		retailCRMUseCase, // Используем RetailCRM usecase
		infra.Logger,
	)

	var whatsgateSettingsUseCase settingsInterfaces.WhatsgateSettingsUseCase = settingsInteractor.NewWhatsgateSettingsInteractor(
		infra.WhatsgateSettingsRepo,
		infra.Logger,
	)

	var retailCRMSettingsUseCase settingsInterfaces.RetailCRMSettingsUseCase = settingsInteractor.NewRetailCRMSettingsInteractor(
		infra.RetailCRMSettingsRepo,
		infra.Logger,
	)

	var testMessageUseCase messagingInterfaces.MessageUseCase = messagingInteractor.NewMessageInteractor(
		infra.MessageGateway,
		infra.Logger,
	)

	return &UseCases{
		Campaign:          campaignUseCase,
		WhatsgateSettings: whatsgateSettingsUseCase,
		RetailCRMSettings: retailCRMSettingsUseCase,
		Message:           testMessageUseCase,
		RetailCRM:         retailCRMUseCase,
	}
}

// NewAdapters создает все адаптеры (конвертеры и презентеры)
func NewAdapters() *Adapters {
	// Конвертеры
	var campaignConverter converter.CampaignConverter = converter.NewCampaignConverter()
	var whatsgateSettingsConverter converter.WhatsgateSettingsConverter = converter.NewWhatsgateSettingsConverter()
	var retailCRMSettingsConverter converter.RetailCRMSettingsConverter = converter.NewRetailCRMSettingsConverter()
	var messagingConverter converter.MessagingConverter = converter.NewMessagingConverter()
	var retailCRMConverter converter.RetailCRMConverter = converter.NewRetailCRMConverter()

	// Presenters
	var campaignPresenter presenters.CampaignPresenterInterface = presenters.NewCampaignPresenter(campaignConverter)
	var whatsgateSettingsPresenter presenters.WhatsgateSettingsPresenterInterface = presenters.NewWhatsgateSettingsPresenter(whatsgateSettingsConverter)
	var retailCRMSettingsPresenter presenters.RetailCRMSettingsPresenterInterface = presenters.NewRetailCRMSettingsPresenter(retailCRMSettingsConverter)
	var messagingPresenter presenters.MessagingPresenterInterface = presenters.NewMessagingPresenter(messagingConverter)
	var retailCRMPresenter presenters.RetailCRMPresenterInterface = presenters.NewRetailCRMPresenter(retailCRMConverter)

	return &Adapters{
		CampaignConverter:          campaignConverter,
		WhatsgateSettingsConverter: whatsgateSettingsConverter,
		RetailCRMSettingsConverter: retailCRMSettingsConverter,
		MessagingConverter:         messagingConverter,
		RetailCRMConverter:         retailCRMConverter,
		CampaignPresenter:          campaignPresenter,
		WhatsgateSettingsPresenter: whatsgateSettingsPresenter,
		RetailCRMSettingsPresenter: retailCRMSettingsPresenter,
		MessagingPresenter:         messagingPresenter,
		RetailCRMPresenter:         retailCRMPresenter,
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

	whatsgateSettingsHandler := handlers.NewWhatsgateSettingsHandler(
		useCases.WhatsgateSettings,
		adapters.WhatsgateSettingsPresenter,
		adapters.WhatsgateSettingsConverter,
		infra.Logger,
	)

	retailCRMSettingsHandler := handlers.NewRetailCRMSettingsHandler(
		useCases.RetailCRMSettings,
		adapters.RetailCRMSettingsPresenter,
		adapters.RetailCRMSettingsConverter,
		infra.Logger,
	)

	messagingHandler := handlers.NewMessagingHandler(
		useCases.Message,
		adapters.MessagingPresenter,
		adapters.MessagingConverter,
		infra.Logger,
	)

	retailCRMHandler := handlers.NewRetailCRMHandler(
		useCases.RetailCRM,
		adapters.RetailCRMConverter,
		adapters.RetailCRMPresenter,
		infra.Logger,
	)

	// Health Handler
	healthHandler := handlers.NewHealthHandler(
		infra.Logger,
		infra.CampaignRepo,
		infra.Dispatcher,
	)

	return &Handlers{
		Campaign:          campaignHandler,
		WhatsgateSettings: whatsgateSettingsHandler,
		RetailCRMSettings: retailCRMSettingsHandler,
		Messaging:         messagingHandler,
		RetailCRM:         retailCRMHandler,
		Health:            healthHandler,
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
	h := NewHandlers(useCases, adapters, infra)

	// HTTP сервер
	httpSrv := createHTTPServer(
		cfg.HTTP.Port,
		h.Campaign,
		h.Messaging,
		h.WhatsgateSettings,
		h.RetailCRMSettings,
		h.RetailCRM,
		h.Health,
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
	messagingHandler *handlers.MessagingHandler,
	whatsgateSettingsHandler *handlers.WhatsgateSettingsHandler,
	retailCRMSettingsHandler *handlers.RetailCRMSettingsHandler,
	retailCRMHandler *handlers.RetailCRMHandler,
	healthHandler *handlers.HealthHandler,
	logger interfaces.Logger,
) *http.HTTPServer {
	return http.NewHTTPServer(
		port,
		campaignHandler,
		messagingHandler,
		whatsgateSettingsHandler,
		retailCRMSettingsHandler,
		retailCRMHandler,
		healthHandler,
		logger,
	)
}

// Start запускает приложение
func (a *App) Start(ctx context.Context) error {
	a.infrastructure.Logger.Info("starting dispatcher")
	a.infrastructure.Dispatcher.Start(ctx)

	if err := a.recoverOrphanedCampaigns(ctx); err != nil {
		a.infrastructure.Logger.Error("failed to recover orphaned campaigns", "error", err)
	}

	a.infrastructure.Logger.Info("HTTP server starting", "port", a.cfg.HTTP.Port)
	return a.server.Start()
}

// Stop останавливает приложение
func (a *App) Stop(ctx context.Context) error {
	a.infrastructure.Logger.Info("stopping application")

	if err := a.gracefulShutdownCampaigns(ctx); err != nil {
		a.infrastructure.Logger.Error("failed to gracefully shutdown campaigns", "error", err)
	}

	a.infrastructure.Logger.Info("stopping dispatcher")
	if err := a.infrastructure.Dispatcher.Stop(ctx); err != nil {
		a.infrastructure.Logger.Error("failed to stop dispatcher", "error", err)
	}

	if err := a.server.Stop(ctx); err != nil {
		return err
	}

	if a.infrastructure.Database != nil {
		a.infrastructure.Database.Close()
	}

	a.infrastructure.Logger.Info("application stopped")
	return nil
}

// gracefulShutdownCampaigns останавливает активные кампании
func (a *App) gracefulShutdownCampaigns(ctx context.Context) error {
	a.infrastructure.Logger.Info("gracefully shutting down campaigns")

	activeCampaigns, err := a.infrastructure.CampaignRepo.GetActiveCampaigns(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active campaigns: %w", err)
	}

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
	campaign, err := a.infrastructure.CampaignRepo.GetByID(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}

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

	startedCampaigns, err := a.infrastructure.CampaignRepo.ListByStatus(ctx, "started", 100, 0)
	if err != nil {
		return fmt.Errorf("failed to get started campaigns: %w", err)
	}

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

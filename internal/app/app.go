package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"whatsapp-service/internal/adapters/controllers"
	"whatsapp-service/internal/config"
	"whatsapp-service/internal/infrastructure/database"
	"whatsapp-service/internal/infrastructure/gateways/whatsgate"
	gtypes "whatsapp-service/internal/infrastructure/gateways/whatsgate/types"
	"whatsapp-service/internal/infrastructure/logger"
	"whatsapp-service/internal/infrastructure/parsers"
	"whatsapp-service/internal/infrastructure/repositories"
	"whatsapp-service/internal/infrastructure/server"
	"whatsapp-service/internal/infrastructure/services"
	"whatsapp-service/internal/usecases/campaigns"
)

// App инкапсулирует все зависимости и умеет запускаться/останавливаться.
type App struct {
	cfg    *config.Config
	logger *zap.Logger
	db     *pgxpool.Pool
	server *server.HTTPServer
}

// New собирает приложение из конфигурации.
func New(cfg *config.Config) (*App, error) {
	// 1. Логгер
	lg, err := logger.New(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	// 2. БД
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := database.NewPostgresPool(ctx, cfg.Database)
	if err != nil {
		logger.Sync()
		return nil, fmt.Errorf("init db: %w", err)
	}

	// 3. Репозитории
	campaignRepo := repositories.NewPostgresCampaignRepository(pool)
	campaignStatusRepo := repositories.NewPostgresCampaignStatusRepository(pool)

	// 4. Внешний gateway WhatsGate
	gwCfg := &gtypes.WhatsGateConfig{
		BaseURL:       getenv("WHATSGATE_URL", "http://localhost:3000"),
		APIKey:        getenv("WHATSGATE_API_KEY", "demo-key"),
		WhatsappID:    getenv("WHATSAPP_ID", "demo-whatsapp"),
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		MaxFileSize:   gtypes.MaxFileSizeBytes,
	}
	messageGateway := whatsgate.NewWhatsGateGateway(gwCfg)

	// 5. Утилитарные сервисы
	rateLimiter := services.NewMemoryRateLimiter()
	fileParser := parsers.NewExcelParser()

	// 6. Use cases
	createUC := campaigns.NewCreateCampaignUseCase(campaignRepo, campaignStatusRepo, fileParser)
	startUC := campaigns.NewStartCampaignUseCase(campaignRepo, campaignStatusRepo, messageGateway, rateLimiter)
	cancelUC := campaigns.NewCancelCampaignUseCase(campaignRepo, campaignStatusRepo, startUC)

	// 7. Controllers
	controller := controllers.NewCampaignController(createUC, startUC, cancelUC, campaignRepo)

	httpSrv := server.NewHTTPServer(cfg.HTTP.Port, controller) // HTTP config currently uses only port.

	return &App{
		cfg:    cfg,
		logger: lg,
		db:     pool,
		server: httpSrv,
	}, nil
}

// Start запускает HTTP сервер (блокирующе).
func (a *App) Start() error {
	return a.server.Start()
}

// Stop останавливает HTTP сервер и закрывает ресурсы.
func (a *App) Stop(ctx context.Context) error {
	if err := a.server.Stop(ctx); err != nil {
		return err
	}
	database.Close(a.db)
	logger.Sync()
	return nil
}

// helper for env with default
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

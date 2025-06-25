package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"whatsapp-service/internal/config"
	"whatsapp-service/internal/database"
	"whatsapp-service/internal/delivery/http"
	appErr "whatsapp-service/internal/errors"
	"whatsapp-service/internal/logger"
	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"
	whatsgateInfra "whatsapp-service/internal/whatsgate/infra"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// App агрегирует все основные зависимости приложения.
type App struct {
	Config           *config.Config
	Logger           logger.Logger
	DB               database.DB
	DBPool           *pgxpool.Pool // Пул для прямого доступа к БД
	WhatsGateService *whatsgateDomain.SettingsService
	Server           *http.Server
}

// InitConfig инициализирует конфиг приложения.
func InitConfig(path string) (*config.Config, error) {
	cfg, err := config.LoadConfig(path)
	if err != nil {
		return nil, appErr.New("CONFIG_LOAD_ERROR", "failed to load config", err)
	}
	return cfg, nil
}

// InitLogger инициализирует логгер приложения.
func InitLogger(cfg config.LoggingConfig) (logger.Logger, error) {
	logCfg := logger.NewConfigFromAppConfig(cfg)
	log, err := logger.NewZapLogger(logCfg)
	if err != nil {
		return nil, appErr.New("LOGGER_INIT_ERROR", "failed to init logger", err)
	}
	return log, nil
}

// InitDB инициализирует пул соединений к базе данных.
func InitDB(ctx context.Context, cfg config.DatabaseConfig) (database.DB, *pgxpool.Pool, error) {
	dbCfg := database.NewConfigFromAppConfig(cfg)
	db, err := database.NewPostgresPool(ctx, dbCfg)
	if err != nil {
		return nil, nil, appErr.New("DB_INIT_ERROR", "failed to init database", err)
	}

	// Получаем прямой доступ к пулу для WhatGate
	poolDB, ok := db.(*database.PoolDB)
	if !ok {
		return nil, nil, appErr.New("DB_POOL_ERROR", "failed to get database pool", nil)
	}

	// Получаем pgxpool.Pool из адаптера
	pool := poolDB.GetPool()

	return db, pool, nil
}

// InitWhatsGateService инициализирует сервис настроек WhatGate
func InitWhatsGateService(pool *pgxpool.Pool, log logger.Logger) (*whatsgateDomain.SettingsService, error) {
	// Создаем репозиторий
	repo := whatsgateInfra.NewSettingsRepository(pool)

	// Создаем сервис с БД хранилищем
	service := whatsgateDomain.NewSettingsService(repo)

	// Инициализируем таблицу
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := service.InitDatabase(ctx); err != nil {
		log.Error("failed to initialize WhatGate database table", zap.Error(err))
		return nil, appErr.New("WHATSGATE_DB_INIT_ERROR", "failed to initialize WhatGate database", err)
	}

	log.Info("WhatGate database table initialized successfully")
	return service, nil
}

// InitServer инициализирует HTTP-сервер приложения.
func InitServer(cfg config.HTTPConfig, log logger.Logger, whatsgateService *whatsgateDomain.SettingsService) *http.Server {
	return http.NewServer(cfg, log, whatsgateService)
}

// BuildApp инициализирует все зависимости приложения.
func BuildApp(ctx context.Context, configPath string) (*App, error) {
	cfg, err := InitConfig(configPath)
	if err != nil {
		return nil, err
	}

	log, err := InitLogger(cfg.Logging)
	if err != nil {
		return nil, err
	}

	db, pool, err := InitDB(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	whatsgateService, err := InitWhatsGateService(pool, log)
	if err != nil {
		return nil, err
	}

	server := InitServer(cfg.HTTP, log, whatsgateService)

	return &App{
		Config:           cfg,
		Logger:           log,
		DB:               db,
		DBPool:           pool,
		WhatsGateService: whatsgateService,
		Server:           server,
	}, nil
}

// GracefulShutdown обеспечивает корректное завершение приложения.
func GracefulShutdown(cancel context.CancelFunc, timeout time.Duration, cleanup func()) {
	// Ожидаем сигналы завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Блокируемся до получения сигнала
	<-sigChan

	// Отменяем контекст
	cancel()

	// Ждем завершения с таймаутом
	time.Sleep(timeout)

	// Выполняем cleanup
	cleanup()
}

// Shutdown корректно завершает работу приложения.
func (a *App) Shutdown(ctx context.Context) error {
	// Закрываем HTTP-сервер
	if a.Server != nil {
		if err := a.Server.Shutdown(ctx); err != nil {
			a.Logger.Error("failed to shutdown HTTP server", zap.Error(err))
		}
	}

	// Закрываем соединения с БД
	if a.DB != nil {
		a.DB.Close()
	}

	return nil
}

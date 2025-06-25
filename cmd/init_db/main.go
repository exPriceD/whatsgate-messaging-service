package main

import (
	"context"
	"fmt"
	"log"

	"whatsapp-service/internal/config"
	"whatsapp-service/internal/database"
	"whatsapp-service/internal/logger"
	whatsgateDomain "whatsapp-service/internal/whatsgate/domain"
	whatsgateInfra "whatsapp-service/internal/whatsgate/infra"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Инициализируем логгер
	logCfg := logger.NewConfigFromAppConfig(cfg.Logging)
	logger, err := logger.NewZapLogger(logCfg)
	if err != nil {
		log.Fatal("Failed to init logger:", err)
	}

	// Создаем контекст
	ctx := context.Background()

	// Подключаемся к PostgreSQL
	dbCfg := database.NewConfigFromAppConfig(cfg.Database)

	// Сначала подключаемся к postgres для создания базы данных
	createDBCfg := database.Config{
		Host:     dbCfg.Host,
		Port:     dbCfg.Port,
		Name:     "postgres", // Подключаемся к системной БД
		User:     dbCfg.User,
		Password: dbCfg.Password,
		SSLMode:  dbCfg.SSLMode,
	}

	// Создаем строку подключения для создания БД
	createDBURL := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=%s",
		createDBCfg.User, createDBCfg.Password, createDBCfg.Host, createDBCfg.Port, createDBCfg.SSLMode)

	// Подключаемся к postgres
	pool, err := pgxpool.New(ctx, createDBURL)
	if err != nil {
		log.Fatal("Failed to connect to postgres:", err)
	}
	defer pool.Close()

	// Создаем базу данных
	_, err = pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbCfg.Name))
	if err != nil {
		// Если база уже существует, это нормально
		logger.Info("Database might already exist, continuing...")
	} else {
		logger.Info("Database created successfully")
	}

	// Закрываем соединение с postgres
	pool.Close()

	// Теперь подключаемся к нашей базе данных
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode)

	pool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	logger.Info("Connected to database successfully")

	// Инициализируем WhatGate сервис
	repo := whatsgateInfra.NewSettingsRepository(pool)
	service := whatsgateDomain.NewSettingsService(repo)

	// Создаем таблицу WhatGate
	err = service.InitDatabase(ctx)
	if err != nil {
		log.Fatal("Failed to initialize WhatGate database:", err)
	}

	logger.Info("WhatGate database initialized successfully")

	// Проверяем, что все работает
	settings := service.GetSettings()
	logger.Info("Default settings loaded",
		zap.String("whatsapp_id", settings.WhatsappID),
		zap.String("base_url", settings.BaseURL))

	logger.Info("Database initialization completed successfully!")
}

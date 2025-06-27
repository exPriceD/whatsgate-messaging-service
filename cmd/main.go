// Package main WhatsApp Service API
//
// Сервис для отправки сообщений через WhatsApp с использованием API WhatGate.
// Поддерживает массовую рассылку и автоматические уведомления о подписке.
//
//	Schemes: http, https
//	Host: localhost:8080
//	BasePath: /api/v1
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Security:
//	- bearer
//
// swagger:meta
package main

import (
	"context"
	"log"
	"os"
	"whatsapp-service/internal/logger"

	"whatsapp-service/internal/app"

	"go.uber.org/zap"
)

// @title WhatsApp Service API
// @version 1.0
// @description Сервис для отправки сообщений через WhatsApp с использованием API WhatGate
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {token}

func main() {
	os.Setenv("TZ", "Europe/Moscow")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// Сборка приложения
	application, err := app.BuildApp(ctx, configPath)
	if err != nil {
		log.Fatalf("failed to build app: %v", err)
	}
	defer func(Logger logger.Logger) {
		_ = Logger.Sync()
	}(application.Logger)

	application.Logger.Info("service started")

	// Запуск HTTP-сервера
	go func() {
		if err := application.Server.Start(); err != nil {
			application.Logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Ожидание сигнала завершения
	app.GracefulShutdown(cancel, application.Config.HTTP.ShutdownTimeout, func() {
		application.Logger.Info("service stopped")
		if err := application.Shutdown(context.Background()); err != nil {
			application.Logger.Error("failed to shutdown application", zap.Error(err))
		}
	})
}

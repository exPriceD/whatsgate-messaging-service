package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"whatsapp-service/internal/app"
	"whatsapp-service/internal/config"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("app init error: %v", err)
	}

	// Запускаем в goroutine, чтобы main получил сигнал
	go func() {
		if err := application.Start(); err != nil {
			log.Fatalf("app run error: %v", err)
		}
	}()

	// Ждём SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := application.Stop(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}

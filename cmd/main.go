// WhatsApp Service API
// @title WhatsApp Service API
// @version 1.0
// @description Сервис для массовой рассылки сообщений через WhatsApp
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@whatsapp-service.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @schemes http https
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
	cfg, err := config.Load("./config/config.yaml")
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("app init error: %v", err)
	}

	go func() {
		if err := application.Start(context.Background()); err != nil {
			log.Fatalf("app run error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := application.Stop(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}

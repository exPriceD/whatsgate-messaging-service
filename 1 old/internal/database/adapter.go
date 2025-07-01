package database

import "whatsapp-service/internal/config"

// NewConfigFromAppConfig преобразует config.DatabaseConfig в database.Config.
func NewConfigFromAppConfig(appCfg config.DatabaseConfig) Config {
	return Config{
		Host:            appCfg.Host,
		Port:            appCfg.Port,
		Name:            appCfg.Name,
		User:            appCfg.User,
		Password:        appCfg.Password,
		SSLMode:         appCfg.SSLMode,
		MaxOpenConns:    appCfg.MaxOpenConns,
		MaxIdleConns:    appCfg.MaxIdleConns,
		ConnMaxLifetime: appCfg.ConnMaxLifetime,
	}
}

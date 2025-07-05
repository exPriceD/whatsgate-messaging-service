package logger

import "whatsapp-service/internal/config"

// NewConfigFromAppConfig преобразует config.LoggingConfig в logger.Config.
func NewConfigFromAppConfig(appCfg config.LoggingConfig) Config {
	return Config{
		Level:      appCfg.Level,
		Format:     appCfg.Format,
		OutputPath: appCfg.OutputPath,
	}
}

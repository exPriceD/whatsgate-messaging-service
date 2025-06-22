package logger

import "fmt"

// Константы уровней логирования.
const (
	LevelDebug  = "debug"
	LevelInfo   = "info"
	LevelWarn   = "warn"
	LevelError  = "error"
	LevelDPanic = "dpanic"
	LevelPanic  = "panic"
	LevelFatal  = "fatal"
)

// Константы форматов логирования.
const (
	FormatJSON    = "json"
	FormatConsole = "console"
)

// Config содержит параметры логгера.
type Config struct {
	Level      string
	Format     string
	OutputPath string
}

// Validate проверяет корректность конфигурации логгера и возвращает подробные ошибки.
func (c *Config) Validate() error {
	switch c.Level {
	case LevelDebug, LevelInfo, LevelWarn, LevelError, LevelDPanic, LevelPanic, LevelFatal:
		// ok
	default:
		return fmt.Errorf("logger: level must be one of debug, info, warn, error, dpanic, panic, fatal")
	}

	switch c.Format {
	case FormatJSON, FormatConsole:
		// ok
	default:
		return fmt.Errorf("logger: format must be one of json, console")
	}
	if c.OutputPath == "" {
		return fmt.Errorf("logger: output_path is required")
	}

	return nil
}

package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var l *zap.Logger

// Init инициализирует zap.Logger по настройкам.
func Init(cfg Config) (*zap.Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	var encoderCfg zapcore.EncoderConfig
	if strings.ToLower(cfg.Format) == FormatJSON {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}

	var ws zapcore.WriteSyncer
	if cfg.OutputPath == "" || cfg.OutputPath == "stdout" {
		ws = zapcore.AddSync(os.Stdout)
	} else {
		f, err := os.OpenFile(cfg.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// fallback на stdout, если не удалось открыть файл
			ws = zapcore.AddSync(os.Stdout)
		} else {
			ws = zapcore.AddSync(f)
		}
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		ws,
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	l = logger
	return logger, nil
}

// L возвращает глобальный логгер.
func L() *zap.Logger {
	if l == nil {
		panic("logger not initialized")
	}
	return l
}

// Sync завершает работу логгера корректно.
func Sync() error {
	if l != nil {
		return l.Sync()
	}
	return nil
}

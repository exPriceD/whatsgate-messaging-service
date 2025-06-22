package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var l *zap.Logger

// Logger — интерфейс для логгирования.
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Sync() error
}

// ZapLogger — адаптер для zap.Logger, реализующий интерфейс Logger.
type ZapLogger struct {
	l *zap.Logger
}

func (z *ZapLogger) Info(msg string, fields ...zap.Field) {
	z.l.Info(msg, fields...)
}

func (z *ZapLogger) Error(msg string, fields ...zap.Field) {
	z.l.Error(msg, fields...)
}

func (z *ZapLogger) Sync() error {
	return z.l.Sync()
}

// NewZapLogger создает ZapLogger по настройкам.
func NewZapLogger(cfg Config) (Logger, error) {
	zl, err := Init(cfg)
	if err != nil {
		return nil, err
	}
	return &ZapLogger{l: zl}, nil
}

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

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel), zap.AddCallerSkip(1))
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

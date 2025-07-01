package logger

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"whatsapp-service/internal/config"
)

var (
	global      *zap.Logger
	atomicLevel zap.AtomicLevel
)

// New создает zap.Logger согласно конфигурации.
// Формат: json | console; output_path: stdout | stderr | filename.
func New(cfg config.LoggingConfig) (*zap.Logger, error) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var enc zapcore.Encoder
	if strings.ToLower(cfg.Format) == "console" {
		enc = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		enc = zapcore.NewJSONEncoder(encoderCfg)
	}

	atomicLevel = zap.NewAtomicLevel()

	switch strings.ToLower(cfg.Level) {
	case "debug":
		atomicLevel.SetLevel(zap.DebugLevel)
	case "warn":
		atomicLevel.SetLevel(zap.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zap.ErrorLevel)
	case "dpanic":
		atomicLevel.SetLevel(zap.DPanicLevel)
	case "panic":
		atomicLevel.SetLevel(zap.PanicLevel)
	case "fatal":
		atomicLevel.SetLevel(zap.FatalLevel)
	default:
		atomicLevel.SetLevel(zap.InfoLevel)
	}

	var ws zapcore.WriteSyncer
	switch cfg.OutputPath {
	case "", "stdout":
		ws = zapcore.AddSync(os.Stdout)
	case "stderr":
		ws = zapcore.AddSync(os.Stderr)
	default:
		f, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, err
		}
		ws = zapcore.AddSync(f)
	}

	ws = zapcore.Lock(ws)

	core := zapcore.NewCore(enc, ws, atomicLevel)

	core = zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)

	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.ErrorOutput(ws),
	}

	logger := zap.New(core, opts...)

	if cfg.Service != "" {
		logger = logger.With(zap.String("service", cfg.Service))
	}
	if cfg.Env != "" {
		logger = logger.With(zap.String("env", cfg.Env))
	}

	global = logger
	return logger, nil
}

// L возвращает глобальный логгер. Если не инициализирован — production logger.
func L() *zap.Logger {
	if global == nil {
		global, _ = zap.NewProduction()
	}
	return global
}

// Sync вызывает Sync у глобального логгера.
func Sync() {
	if global != nil {
		_ = global.Sync()
	}
}

// SetLevel меняет уровень логирования на лету (debug, info ...)
func SetLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		atomicLevel.SetLevel(zap.DebugLevel)
	case "info":
		atomicLevel.SetLevel(zap.InfoLevel)
	case "warn":
		atomicLevel.SetLevel(zap.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zap.ErrorLevel)
	case "dpanic":
		atomicLevel.SetLevel(zap.DPanicLevel)
	case "panic":
		atomicLevel.SetLevel(zap.PanicLevel)
	case "fatal":
		atomicLevel.SetLevel(zap.FatalLevel)
	}
}

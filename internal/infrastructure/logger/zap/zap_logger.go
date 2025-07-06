package zaplogger

import (
	"os"
	"strings"
	"time"
	"whatsapp-service/internal/shared/logger"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"whatsapp-service/internal/config"
)

type zapAdapter struct {
	sugar *zap.SugaredLogger
}

var (
	globalLogger *zap.Logger
	atomicLevel  zap.AtomicLevel
)

// New возвращает обёртку zap над интерфейсом logger.Logger
func New(cfg config.LoggingConfig) (logger.Logger, error) {
	coreLogger, err := buildZap(cfg)
	if err != nil {
		return nil, err
	}

	sugar := coreLogger.Sugar()
	return &zapAdapter{sugar: sugar}, nil
}

// Info logs a message at InfoLevel.
func (l *zapAdapter) Info(msg string, fields ...any) {
	l.sugar.Infow(msg, l.convertFields(fields)...)
}

// Warn logs a message at WarnLevel.
func (l *zapAdapter) Warn(msg string, fields ...any) {
	l.sugar.Warnw(msg, l.convertFields(fields)...)
}

// Error logs a message at ErrorLevel.
func (l *zapAdapter) Error(msg string, fields ...any) {
	l.sugar.Errorw(msg, l.convertFields(fields)...)
}

// Debug logs a message at DebugLevel.
func (l *zapAdapter) Debug(msg string, fields ...any) {
	l.sugar.Debugw(msg, l.convertFields(fields)...)
}

// With returns a child logger with structured context.
func (l *zapAdapter) With(fields ...any) logger.Logger {
	return &zapAdapter{sugar: l.sugar.With(l.convertFields(fields)...)}
}

// convertFields преобразует различные форматы полей в формат zap
func (l *zapAdapter) convertFields(fields []any) []any {
	if len(fields) == 0 {
		return fields
	}

	if len(fields) == 1 {
		if fieldMap, ok := fields[0].(map[string]interface{}); ok {
			result := make([]any, 0, len(fieldMap)*2)
			for key, value := range fieldMap {
				result = append(result, key, value)
			}
			return result
		}
	}

	return fields
}

// Sync flushes any buffered log entries.
func (l *zapAdapter) Sync() {
	_ = l.sugar.Sync()
}

// SetLevel changes log level at runtime.
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

// buildZap создаёт zap.Logger согласно конфигурации
func buildZap(cfg config.LoggingConfig) (*zap.Logger, error) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	switch strings.ToLower(cfg.Format) {
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	default:
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	atomicLevel = zap.NewAtomicLevel()
	SetLevel(cfg.Level)

	var syncer zapcore.WriteSyncer
	switch cfg.OutputPath {
	case "", "stdout":
		syncer = zapcore.AddSync(os.Stdout)
	case "stderr":
		syncer = zapcore.AddSync(os.Stderr)
	default:
		f, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, err
		}
		syncer = zapcore.AddSync(f)
	}

	syncer = zapcore.Lock(syncer)

	core := zapcore.NewCore(encoder, syncer, atomicLevel)
	core = zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)

	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.ErrorOutput(syncer),
	}

	coreLogger := zap.New(core, options...)

	if cfg.Service != "" {
		coreLogger = coreLogger.With(zap.String("service", cfg.Service))
	}
	if cfg.Env != "" {
		coreLogger = coreLogger.With(zap.String("env", cfg.Env))
	}

	globalLogger = coreLogger
	return coreLogger, nil
}

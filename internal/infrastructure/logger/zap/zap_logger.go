package zaplogger

import (
	"os"
	"strings"
	"time"
	"whatsapp-service/internal/interfaces"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"whatsapp-service/internal/config"
)

type zapAdapter struct {
	logger *zap.Logger
}

var (
	globalLogger *zap.Logger
	atomicLevel  zap.AtomicLevel
)

// New возвращает обёртку zap над интерфейсом logger.Logger
func New(cfg config.LoggingConfig) (interfaces.Logger, error) {
	coreLogger, err := buildZap(cfg)
	if err != nil {
		return nil, err
	}

	return &zapAdapter{logger: coreLogger}, nil
}

// Info logs a message at InfoLevel.
func (l *zapAdapter) Info(msg string, fields ...any) {
	l.logger.WithOptions(zap.AddCallerSkip(1)).Info(msg, l.convertFields(fields)...)
}

// Warn logs a message at WarnLevel.
func (l *zapAdapter) Warn(msg string, fields ...any) {
	l.logger.WithOptions(zap.AddCallerSkip(1)).Warn(msg, l.convertFields(fields)...)
}

// Error logs a message at ErrorLevel.
func (l *zapAdapter) Error(msg string, fields ...any) {
	l.logger.WithOptions(zap.AddCallerSkip(1)).Error(msg, l.convertFields(fields)...)
}

// Debug logs a message at DebugLevel.
func (l *zapAdapter) Debug(msg string, fields ...any) {
	l.logger.WithOptions(zap.AddCallerSkip(1)).Debug(msg, l.convertFields(fields)...)
}

// With returns a child logger with structured context.
func (l *zapAdapter) With(fields ...any) interfaces.Logger {
	return &zapAdapter{logger: l.logger.With(l.convertFields(fields)...)}
}

// convertFields преобразует различные форматы полей в формат zap
func (l *zapAdapter) convertFields(fields []any) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	var zapFields []zap.Field

	if len(fields) == 1 {
		if fieldMap, ok := fields[0].(map[string]interface{}); ok {
			zapFields = make([]zap.Field, 0, len(fieldMap))
			for key, value := range fieldMap {
				zapFields = append(zapFields, zap.Any(key, value))
			}
			return zapFields
		}
	}

	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			if key, ok := fields[i].(string); ok {
				zapFields = append(zapFields, zap.Any(key, fields[i+1]))
			}
		}
	}

	return zapFields
}

// Sync flushes any buffered log entries.
func (l *zapAdapter) Sync() {
	_ = l.logger.Sync()
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

// isDevelopment проверяет является ли окружение разработческим
func isDevelopment(env string) bool {
	devEnvs := []string{"dev", "development", "local", "debug"}
	envLower := strings.ToLower(env)
	for _, devEnv := range devEnvs {
		if envLower == devEnv {
			return true
		}
	}
	return false
}

// buildDevelopmentConfig создает конфигурацию для разработки с красивым форматированием
func buildDevelopmentConfig() zapcore.EncoderConfig {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder           // Цветные уровни логов
	config.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000") // Короткое время
	config.EncodeCaller = zapcore.FullCallerEncoder                 // Полный путь к файлу
	config.ConsoleSeparator = " | "
	return config
}

// buildProductionConfig создает конфигурацию для продакшена
func buildProductionConfig() zapcore.EncoderConfig {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "timestamp"
	config.EncodeTime = zapcore.RFC3339TimeEncoder
	config.MessageKey = "message"
	config.LevelKey = "level"
	config.CallerKey = "caller"
	config.StacktraceKey = "stacktrace"
	config.EncodeCaller = zapcore.FullCallerEncoder // Полный путь к файлу
	return config
}

// buildZap создаёт zap.Logger согласно конфигурации с учетом окружения
func buildZap(cfg config.LoggingConfig) (*zap.Logger, error) {
	var encoderCfg zapcore.EncoderConfig
	var encoder zapcore.Encoder

	isDev := isDevelopment(cfg.Env)

	if isDev {
		// Development окружение - красивые цветные логи
		encoderCfg = buildDevelopmentConfig()
		if cfg.Format == "json" {
			encoder = zapcore.NewJSONEncoder(encoderCfg)
		} else {
			encoder = zapcore.NewConsoleEncoder(encoderCfg)
		}
	} else {
		// Production окружение - структурированные JSON логи
		encoderCfg = buildProductionConfig()
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	atomicLevel = zap.NewAtomicLevel()

	if cfg.Level == "" {
		if isDev {
			cfg.Level = "debug"
		} else {
			cfg.Level = "info"
		}
	}
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

	if !isDev {
		core = zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
	}

	options := []zap.Option{
		zap.AddCaller(),
	}

	if isDev {
		options = append(options, zap.AddStacktrace(zapcore.WarnLevel))
		options = append(options, zap.Development())
	} else {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	options = append(options, zap.ErrorOutput(syncer))

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

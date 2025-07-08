package zaplogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"
	"whatsapp-service/internal/interfaces"

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
func New(cfg config.LoggingConfig) (interfaces.Logger, error) {
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
func (l *zapAdapter) With(fields ...any) interfaces.Logger {
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
	config.EncodeCaller = zapcore.ShortCallerEncoder                // Короткий путь к файлу
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

	// Автоматическая настройка уровня логирования по окружению
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

	// Для продакшена добавляем сэмплинг для высоконагруженных приложений
	if !isDev {
		core = zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
	}

	options := []zap.Option{
		zap.AddCaller(),
	}

	// Стэктрейсы только для Error+ в продакшене, для всех уровней в dev
	if isDev {
		options = append(options, zap.AddStacktrace(zapcore.WarnLevel))
		options = append(options, zap.Development())
	} else {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	options = append(options, zap.ErrorOutput(syncer))

	coreLogger := zap.New(core, options...)

	// Добавляем контекстные поля
	if cfg.Service != "" {
		coreLogger = coreLogger.With(zap.String("service", cfg.Service))
	}
	if cfg.Env != "" {
		coreLogger = coreLogger.With(zap.String("env", cfg.Env))
	}

	globalLogger = coreLogger
	return coreLogger, nil
}

package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
)

// zapPGXLogger удовлетворяет интерфейсу tracelog.Logger и пишет в zap.
// Если запрос выполнялся дольше slowQuery, пишем уровень Warn.
// Иначе Info.

type zapPGXLogger struct {
	logger    *zap.Logger
	slowQuery time.Duration
}

func newZapPGXLogger(l *zap.Logger, slow time.Duration) tracelog.Logger {
	return &zapPGXLogger{logger: l, slowQuery: slow}
}

func (z *zapPGXLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	dur, _ := data["time"].(time.Duration)
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	if dur >= z.slowQuery {
		z.logger.Warn(msg, fields...)
		return
	}

	switch level {
	case tracelog.LogLevelError:
		z.logger.Error(msg, fields...)
	case tracelog.LogLevelWarn:
		z.logger.Warn(msg, fields...)
	case tracelog.LogLevelInfo:
		z.logger.Info(msg, fields...)
	case tracelog.LogLevelDebug, tracelog.LogLevelTrace:
		z.logger.Debug(msg, fields...)
	default:
		z.logger.Info(msg, fields...)
	}
}

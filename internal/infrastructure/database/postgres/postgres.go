package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
	"net/url"
	"time"

	"whatsapp-service/internal/config"
)

// NewPostgresPool создает пул соединений pgxpool.Pool согласно настройкам из конфигурации.
// Он сразу проверяет соединение вызовом Ping.
func NewPostgresPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	tz := url.QueryEscape(cfg.Timezone)
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode, tz)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns >= 0 {
		poolCfg.MinConns = int32(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime
	}
	if cfg.ConnMaxIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime
	}
	if cfg.HealthCheckPeriod > 0 {
		poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	}

	poolCfg.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   newZapPGXLogger(zap.L(), 500*time.Millisecond),
		LogLevel: tracelog.LogLevelInfo,
	}

	var pool *pgxpool.Pool
	for attempt := 1; attempt <= cfg.MaxAttemptConnection; attempt++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
		if err != nil {
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		if pingErr := pool.Ping(ctx); pingErr != nil {
			pool.Close()
			err = pingErr
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("connect retries exceeded: %w", err)
	}

	return pool, nil
}

// Close безопасно закрывает пул, если он не nil.
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

// HealthCheck пытается выполнить Ping к базе в течение timeout.
func HealthCheck(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) error {
	if pool == nil {
		return fmt.Errorf("nil pool")
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return pool.Ping(ctx)
}

// WithTx оборачивает выполнение fn в транзакцию. При ошибке делает Rollback.
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

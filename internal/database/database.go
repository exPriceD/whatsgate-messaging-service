package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool создает пул соединений к PostgreSQL через pgxpool.
func NewPostgresPool(ctx context.Context, dbCfg Config) (*pgxpool.Pool, error) {
	if err := dbCfg.Validate(); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	poolCfg.MaxConns = int32(dbCfg.MaxOpenConns)
	poolCfg.MinConns = int32(dbCfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = dbCfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return pool, nil
}

// HealthCheck проверяет доступность базы данных.
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return pool.Ping(ctx)
}

package database

import (
	"context"
	"fmt"
	"time"

	appErr "whatsapp-service/internal/errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB — интерфейс для работы с базой данных.
type DB interface {
	Close()
	Ping(ctx context.Context) error
}

// PoolDB — адаптер для pgxpool.Pool, реализующий интерфейс DB.
type PoolDB struct {
	pool *pgxpool.Pool
}

func (p *PoolDB) Close() {
	p.pool.Close()
}

func (p *PoolDB) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// GetPool возвращает прямой доступ к pgxpool.Pool
func (p *PoolDB) GetPool() *pgxpool.Pool {
	return p.pool
}

// NewPostgresPool создает PoolDB (адаптер).
func NewPostgresPool(ctx context.Context, dbCfg Config) (DB, error) {
	if err := dbCfg.Validate(); err != nil {
		return nil, appErr.New("DB_CONFIG_INVALID", "invalid database config", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=Europe/Moscow",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, appErr.New("DB_POOL_PARSE_ERROR", "failed to parse pool config", err)
	}

	poolCfg.MaxConns = int32(dbCfg.MaxOpenConns)
	poolCfg.MinConns = int32(dbCfg.MaxIdleConns)
	poolCfg.MaxConnLifetime = dbCfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, appErr.New("DB_POOL_CREATE_ERROR", "failed to create pool", err)
	}

	return &PoolDB{pool: pool}, nil
}

// HealthCheck проверяет доступность базы данных.
func HealthCheck(ctx context.Context, db DB) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return db.Ping(ctx)
}

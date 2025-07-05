package database

import (
	"fmt"
	"time"
)

// Config содержит параметры подключения к базе данных.
type Config struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Validate проверяет корректность конфигурации и возвращает подробные ошибки.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("database: host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("database: port must be in 1..65535")
	}
	if c.Name == "" {
		return fmt.Errorf("database: name is required")
	}
	if c.User == "" {
		return fmt.Errorf("database: user is required")
	}
	if c.Password == "" {
		return fmt.Errorf("database: password is required")
	}
	if c.SSLMode != "disable" && c.SSLMode != "require" && c.SSLMode != "verify-ca" && c.SSLMode != "verify-full" {
		return fmt.Errorf("database: ssl_mode must be one of disable, require, verify-ca, verify-full")
	}
	if c.MaxOpenConns < 1 {
		return fmt.Errorf("database: max_open_conns must be >= 1")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("database: max_idle_conns must be >= 0")
	}
	if c.ConnMaxLifetime <= 0 {
		return fmt.Errorf("database: conn_max_lifetime must be > 0")
	}
	return nil
}

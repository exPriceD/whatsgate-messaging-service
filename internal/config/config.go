package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	appErr "whatsapp-service/internal/errors"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http" validate:"required"`
	Database DatabaseConfig `yaml:"database" validate:"required"`
	Logging  LoggingConfig  `yaml:"logging" validate:"required"`
}

type HTTPConfig struct {
	Host            string        `yaml:"host" validate:"required,hostname|ip"`
	Port            int           `yaml:"port" validate:"required,gt=0,lte=65535"`
	ReadTimeout     time.Duration `yaml:"read_timeout" validate:"required,gt=0"`
	WriteTimeout    time.Duration `yaml:"write_timeout" validate:"required,gt=0"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" validate:"required,gt=0"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" validate:"required,gt=0"`
	CORS            CORSConfig    `yaml:"cors" validate:"required"`
}

type CORSConfig struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins" validate:"dive,required"`
	AllowedMethods   []string `yaml:"allowed_methods" validate:"dive,required"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" validate:"gte=0"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host" validate:"required,hostname|ip"`
	Port            int           `yaml:"port" validate:"required,gt=0,lte=65535"`
	Name            string        `yaml:"name" validate:"required"`
	User            string        `yaml:"user" validate:"required"`
	Password        string        `yaml:"password" validate:"required"`
	SSLMode         string        `yaml:"ssl_mode" validate:"oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int           `yaml:"max_open_conns" validate:"gte=1"`
	MaxIdleConns    int           `yaml:"max_idle_conns" validate:"gte=0"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" validate:"required,gt=0"`
}

type LoggingConfig struct {
	Level      string `yaml:"level" validate:"required,oneof=debug info warn error dpanic panic fatal"`
	Format     string `yaml:"format" validate:"required,oneof=json console"`
	OutputPath string `yaml:"output_path" validate:"required"`
}

// LoadConfig читает конфиг из yaml-файла, применяет дефолты, перекрывает env и валидирует.
func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, appErr.New("CONFIG_FILE_OPEN_ERROR", "failed to open config file", err)
	}

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, appErr.New("CONFIG_DECODE_ERROR", "failed to decode yaml config", err)
	}

	cfg.setDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, appErr.New("CONFIG_VALIDATE_ERROR", "config validation failed", err)
	}

	return &cfg, nil
}

// Validate валидирует конфигурацию согласно тегам validate.
func (c *Config) Validate() error {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("yaml")
	})

	if err := validate.Struct(c); err != nil {
		var errs []string
		for _, err := range err.(validator.ValidationErrors) {
			errs = append(errs, fmt.Sprintf("field '%s' failed validation: %s", err.Field(), err.Tag()))
		}
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// setDefaults применяет дефолтные значения для некоторых полей.
func (c *Config) setDefaults() {
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 10
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 5
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
}

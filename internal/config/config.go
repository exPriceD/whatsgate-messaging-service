package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

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
	Host                 string        `yaml:"host" validate:"required,hostname|ip"`
	Port                 int           `yaml:"port" validate:"required,gt=0,lte=65535"`
	Name                 string        `yaml:"name" validate:"required"`
	User                 string        `yaml:"user" validate:"required"`
	Password             string        `yaml:"password" validate:"required"`
	SSLMode              string        `yaml:"ssl_mode" validate:"oneof=disable require verify-ca verify-full"`
	MaxOpenConns         int           `yaml:"max_open_conns" validate:"gte=1"`
	MaxIdleConns         int           `yaml:"max_idle_conns" validate:"gte=0"`
	ConnMaxLifetime      time.Duration `yaml:"conn_max_lifetime" validate:"required,gt=0"`
	ConnMaxIdleTime      time.Duration `yaml:"conn_max_idle_time" validate:"gte=0"`
	HealthCheckPeriod    time.Duration `yaml:"health_check_period" validate:"gte=0"`
	Timezone             string        `yaml:"timezone" validate:"required"`
	MaxAttemptConnection int           `yaml:"max_attempt_connection" validate:"required,gte=1"`
}

type LoggingConfig struct {
	Level      string `yaml:"level" validate:"required,oneof=debug info warn error dpanic panic fatal"`
	Format     string `yaml:"format" validate:"required,oneof=json console"`
	OutputPath string `yaml:"output_path" validate:"required"`
	Service    string `yaml:"service,omitempty"`
	Env        string `yaml:"env,omitempty"`
}

// LoadConfig читает файл YAML, применяет дефолтные значения, перекрывает часть
// настроек переменными окружения и валидирует итоговую структуру.
// Если path пустой, пытается взять CONFIG_PATH, иначе "config.dev.yaml".
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		if env := os.Getenv("CONFIG_PATH"); env != "" {
			path = env
		} else {
			path = "config.dev.yaml"
		}
	}

	// Файл может отсутствовать – например, когда всё задаётся через окружение.
	var cfg Config
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
			return nil, fmt.Errorf("decode yaml: %w", err)
		}
	} else if !os.IsNotExist(err) { // если ошибка отлична от "нет файла"
		return nil, fmt.Errorf("open config file: %w", err)
	}

	// Переопределяем отдельные поля через переменные окружения.
	overrideWithEnv(&cfg)

	cfg.setDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Load – короткий алиас для LoadConfig.
func Load(path string) (*Config, error) { return LoadConfig(path) }

// overrideWithEnv переопределяет часть конфигурации переменными окружения.
// Используем плоские переменные вида HTTP_HOST, DATABASE_PASSWORD и т.д.
func overrideWithEnv(cfg *Config) {
	if v := os.Getenv("HTTP_HOST"); v != "" {
		cfg.HTTP.Host = v
	}
	if v := os.Getenv("HTTP_PORT"); v != "" {
		if p, _ := strconv.Atoi(v); p > 0 {
			cfg.HTTP.Port = p
		}
	}

	if v := os.Getenv("DATABASE_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DATABASE_PORT"); v != "" {
		if p, _ := strconv.Atoi(v); p > 0 {
			cfg.Database.Port = p
		}
	}
	if v := os.Getenv("DATABASE_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("DATABASE_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DATABASE_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}

	// Настройки логгера
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = strings.ToLower(v)
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.Logging.Format = strings.ToLower(v)
	}
	if v := os.Getenv("LOG_OUTPUT"); v != "" {
		cfg.Logging.OutputPath = v
	}
	if v := os.Getenv("LOG_SERVICE"); v != "" {
		cfg.Logging.Service = v
	}

	// Автоматическое определение окружения
	if v := os.Getenv("ENV"); v != "" {
		cfg.Logging.Env = strings.ToLower(v)
	} else if v := os.Getenv("ENVIRONMENT"); v != "" {
		cfg.Logging.Env = strings.ToLower(v)
	} else if v := os.Getenv("APP_ENV"); v != "" {
		cfg.Logging.Env = strings.ToLower(v)
	}
}

// Validate валидирует конфигурацию согласно тегам validate.
func (c *Config) Validate() error {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("yaml")
		if tag == "" {
			tag = fld.Name
		}
		return tag
	})

	if err := validate.Struct(c); err != nil {
		var sb strings.Builder
		sb.WriteString("config validation errors:")
		for _, ve := range err.(validator.ValidationErrors) {
			sb.WriteString(fmt.Sprintf(" %s[%s]", ve.Field(), ve.Tag()))
		}
		return fmt.Errorf(sb.String())
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
	if c.Database.Timezone == "" {
		c.Database.Timezone = "Europe/Moscow"
	}
	if c.Database.HealthCheckPeriod == 0 {
		c.Database.HealthCheckPeriod = 30 * time.Second
	}
	if c.Logging.Service == "" {
		c.Logging.Service = "whatsapp-service"
	}
}

// HTTPListenAddress возвращает host:port строку.
func (c *Config) HTTPListenAddress() string {
	return fmt.Sprintf("%s:%d", c.HTTP.Host, c.HTTP.Port)
}

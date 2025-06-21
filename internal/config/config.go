package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Database DatabaseConfig `yaml:"database"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host" default:"localhost"`
	Port            int           `yaml:"port" default:"5432"`
	Name            string        `yaml:"name"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode" default:"disable"`
	MaxOpenConns    int           `yaml:"max_open_conns" default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" default:"25"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" default:"5m"`
}

type HTTPConfig struct {
	Host            string        `yaml:"host" default:"0.0.0.0"`
	Port            int           `yaml:"port" default:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" default:"10s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" default:"30s"`
	CORS            CORSConfig    `yaml:"cors"`
}

type CORSConfig struct {
	Enabled          bool     `yaml:"enabled" default:"false"`
	AllowedOrigins   []string `yaml:"allowed_origins" default:"*"`
	AllowedMethods   []string `yaml:"allowed_methods" default:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowedHeaders   []string `yaml:"allowed_headers" default:"Content-Type,Authorization"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" default:"false"`
	MaxAge           int      `yaml:"max_age" default:"86400"` // 24 hours in seconds
}

type KafkaConfig struct {
	Brokers       []string      `yaml:"brokers"`
	Topic         string        `yaml:"topic" default:"email_events"`
	ConsumerGroup string        `yaml:"consumer_group" default:"email_service_group"`
	MaxRetries    int           `yaml:"max_retries" default:"3"`
	RetryBackoff  time.Duration `yaml:"retry_backoff" default:"2s"`
	EnableTLS     bool          `yaml:"enable_tls" default:"false"`
	TLSCACert     string        `yaml:"tls_ca_cert"`
	TLSClientCert string        `yaml:"tls_client_cert"`
	TLSClientKey  string        `yaml:"tls_client_key"`
}

type LoggingConfig struct {
	Level      string `yaml:"level" default:"info"`
	Format     string `yaml:"format" default:"json"`
	OutputPath string `yaml:"output_path" default:"stdout"`
}

func LoadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения YAML файла: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка распаковки YAML файла: %v", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("ошибка валидации: %v", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("POSTGRES_HOST is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("POSTGRES_DBNAME is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("POSTGRES_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required")
	}
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}
	return nil
}

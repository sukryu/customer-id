package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config holds the configuration for the customer-id service.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	HTTPPort int           `mapstructure:"http_port"`
	GRPCPort int           `mapstructure:"grpc_port"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

type PostgresConfig struct {
	Host               string `mapstructure:"host"`
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
	Database           string `mapstructure:"database"`
	MaxConnections     int    `mapstructure:"max_connections"`
	MinIdleConnections int    `mapstructure:"min_idle_connections"`
}

type JWTConfig struct {
	PrivateKeyPath string `mapstructure:"private_key"`
	PublicKeyPath  string `mapstructure:"public_key"`
	Expiration     int64  `mapstructure:"expiration"`
}

type KafkaConfig struct {
	Broker       string        `mapstructure:"broker"`
	Topic        string        `mapstructure:"topic"`
	Partition    int           `mapstructure:"partition"`
	RetryBackoff time.Duration `mapstructure:"retry_backoff"`
}

type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// Load loads the configuration from file and environment variables.
func Load(logger *zap.Logger) (*Config, error) {
	if logger == nil {
		// Fallback to a default logger if none provided
		var err error
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, fmt.Errorf("failed to create default logger: %w", err)
		}
		defer logger.Sync()
	}

	v := viper.New()

	// Set environment variable prefix and enable automatic env override
	v.SetEnvPrefix("TASTESYNC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Determine environment (default to "dev")
	env := v.GetString("ENV")
	if env == "" {
		env = "dev"
		logger.Info("No TASTESYNC_ENV specified, defaulting to 'dev'")
	}

	// Use absolute path to config directory
	configPath, err := filepath.Abs("internal/config")
	if err != nil {
		logger.Error("Failed to resolve config path", zap.Error(err))
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}
	v.AddConfigPath(configPath)
	v.SetConfigName(fmt.Sprintf("config-%s", env))
	v.SetConfigType("yaml")

	// Attempt to read config file, falling back to defaults if not found
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Warn("Config file not found, using defaults and environment variables", zap.String("env", env))
		} else {
			logger.Error("Failed to read config file", zap.Error(err))
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Error("Failed to unmarshal config", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&cfg, logger); err != nil {
		return nil, err
	}

	logger.Info("Configuration loaded successfully",
		zap.String("env", env),
		zap.Int("http_port", cfg.Server.HTTPPort),
		zap.Int("grpc_port", cfg.Server.GRPCPort))
	return &cfg, nil
}

// validateConfig checks for required fields and sets sensible defaults.
func validateConfig(cfg *Config, logger *zap.Logger) error {
	if cfg.Server.HTTPPort <= 0 || cfg.Server.HTTPPort > 65535 {
		logger.Error("Invalid HTTP port", zap.Int("http_port", cfg.Server.HTTPPort))
		return fmt.Errorf("http_port must be between 1 and 65535")
	}
	if cfg.Server.GRPCPort <= 0 || cfg.Server.GRPCPort > 65535 {
		logger.Error("Invalid gRPC port", zap.Int("grpc_port", cfg.Server.GRPCPort))
		return fmt.Errorf("grpc_port must be between 1 and 65535")
	}
	if cfg.Server.Timeout <= 0 {
		logger.Warn("Invalid timeout, setting default", zap.Duration("timeout", cfg.Server.Timeout))
		cfg.Server.Timeout = 5 * time.Second
	}
	if cfg.Redis.Host == "" {
		logger.Error("Redis host is required")
		return fmt.Errorf("redis.host is required")
	}
	if cfg.Postgres.Host == "" || cfg.Postgres.User == "" || cfg.Postgres.Database == "" {
		logger.Error("PostgreSQL configuration incomplete",
			zap.String("host", cfg.Postgres.Host),
			zap.String("user", cfg.Postgres.User),
			zap.String("database", cfg.Postgres.Database))
		return fmt.Errorf("postgres.host, user, and database are required")
	}
	if cfg.JWT.PrivateKeyPath == "" || cfg.JWT.PublicKeyPath == "" {
		logger.Error("JWT key paths are required")
		return fmt.Errorf("jwt.private_key and public_key are required")
	}
	if cfg.Kafka.Broker == "" || cfg.Kafka.Topic == "" {
		logger.Error("Kafka configuration incomplete",
			zap.String("broker", cfg.Kafka.Broker),
			zap.String("topic", cfg.Kafka.Topic))
		return fmt.Errorf("kafka.broker and topic are required")
	}
	if cfg.Logging.Level == "" {
		logger.Warn("Log level not specified, defaulting to 'info'")
		cfg.Logging.Level = "info"
	}
	return nil
}

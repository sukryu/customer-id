package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/customer-id.git/internal/config"
	"go.uber.org/zap/zaptest"
)

func TestLoadConfig(t *testing.T) {
	// Setup test logger
	logger := zaptest.NewLogger(t)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "customer-id-test-")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Create internal/config structure in temp directory
	configDir := filepath.Join(tempDir, "internal", "config")
	err = os.MkdirAll(configDir, 0755)
	assert.NoError(t, err, "Failed to create config directory")

	// Create a test config file
	testConfig := `
server:
  http_port: 3000
  grpc_port: 50051
  timeout: 5s
redis:
  host: "localhost:6379"
  password: ""
  db: 0
  max_retries: 3
  pool_size: 10
postgres:
  host: "localhost:5432"
  user: "tastesync"
  password: "secret"
  database: "tastesync"
  max_connections: 20
  min_idle_connections: 5
jwt:
  private_key: "keys/dev/private.pem"
  public_key: "keys/dev/public.pem"
  expiration: 3600
kafka:
  broker: "localhost:9092"
  topic: "customer-events"
  partition: 3
  retry_backoff: 500ms
logging:
  level: "info"
  output: "stdout"
`
	configPath := filepath.Join(configDir, "config-test.yaml")
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	assert.NoError(t, err, "Failed to write test config file")

	// Change working directory to temp directory
	err = os.Chdir(tempDir)
	assert.NoError(t, err, "Failed to change to temp directory")

	// Set environment variables for testing
	os.Setenv("TASTESYNC_ENV", "test")
	os.Setenv("TASTESYNC_POSTGRES_PASSWORD", "testsecret")
	defer func() {
		os.Unsetenv("TASTESYNC_ENV")
		os.Unsetenv("TASTESYNC_POSTGRES_PASSWORD")
	}()

	// Load config
	cfg, err := config.Load(logger)
	if !assert.NoError(t, err, "Expected no error loading config") {
		t.Logf("Config load failed: %v", err)
		return
	}
	assert.NotNil(t, cfg, "Config should not be nil")

	// Validate config values
	assert.Equal(t, 3000, cfg.Server.HTTPPort, "HTTP port mismatch")
	assert.Equal(t, 50051, cfg.Server.GRPCPort, "gRPC port mismatch")
	assert.Equal(t, 5*time.Second, cfg.Server.Timeout, "Timeout mismatch")
	assert.Equal(t, "localhost:6379", cfg.Redis.Host, "Redis host mismatch")
	assert.Equal(t, "testsecret", cfg.Postgres.Password, "Postgres password should be overridden by env")
	assert.Equal(t, "keys/dev/private.pem", cfg.JWT.PrivateKeyPath, "JWT private key path mismatch")
	assert.Equal(t, "keys/dev/public.pem", cfg.JWT.PublicKeyPath, "JWT public key path mismatch")
	assert.Equal(t, "customer-events", cfg.Kafka.Topic, "Kafka topic mismatch")
	assert.Equal(t, "info", cfg.Logging.Level, "Logging level mismatch")
}

func TestLoadConfigMissingRequired(t *testing.T) {
	// Setup test logger
	logger := zaptest.NewLogger(t)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "customer-id-test-")
	assert.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Create internal/config structure in temp directory
	configDir := filepath.Join(tempDir, "internal", "config")
	err = os.MkdirAll(configDir, 0755)
	assert.NoError(t, err, "Failed to create config directory")

	// Create an invalid test config file with missing required fields
	invalidConfig := `
server:
  http_port: -1  # Invalid port
redis:
  host: ""       # Missing required field
`
	configPath := filepath.Join(configDir, "config-missing.yaml")
	err = os.WriteFile(configPath, []byte(invalidConfig), 0644)
	assert.NoError(t, err, "Failed to write missing config file")

	// Change working directory to temp directory
	err = os.Chdir(tempDir)
	assert.NoError(t, err, "Failed to change to temp directory")

	// Set environment to use the invalid config
	os.Setenv("TASTESYNC_ENV", "missing")
	defer os.Unsetenv("TASTESYNC_ENV")

	// Expect validation failure
	_, err = config.Load(logger)
	assert.Error(t, err, "Expected error when required fields are missing")
	assert.Contains(t, err.Error(), "http_port must be between 1 and 65535", "Error should indicate invalid port")
}

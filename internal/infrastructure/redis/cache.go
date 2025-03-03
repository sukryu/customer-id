package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sukryu/customer-id.git/internal/domain/aggregates"
)

// Cache provides methods to interact with Redis for caching customer identities.
// It supports setting and retrieving CustomerIdentity data with TTL expiration.
type Cache interface {
	SetCustomerIdentity(ctx context.Context, identity *aggregates.CustomerIdentity) error
	GetCustomerIdentity(ctx context.Context, customerID string) (*aggregates.CustomerIdentity, error)
	Close() error
}

// cache implements the Cache interface using Redis as the underlying store.
type cache struct {
	client *redis.Client // Redis client instance for connection pooling
}

// NewCache creates a new Cache instance with the provided Redis configuration.
// It establishes a connection to Redis and verifies connectivity with a ping.
// Returns an error if the connection fails or configuration is invalid.
func NewCache(addr, password string, db int) (Cache, error) {
	if addr == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,     // e.g., "localhost:6379"
		Password: password, // Redis password, empty if not set
		DB:       db,       // Database number, typically 0
		PoolSize: 10,       // Connection pool size for concurrency
	})

	// Verify connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis at %s: %w", addr, err)
	}

	return &cache{
		client: client,
	}, nil
}

// SetCustomerIdentity stores a CustomerIdentity in Redis with a TTL of 1 hour.
// It serializes the identity to JSON and uses the customer ID as the key prefix.
// Returns an error if serialization or storage fails.
func (c *cache) SetCustomerIdentity(ctx context.Context, identity *aggregates.CustomerIdentity) error {
	if identity == nil {
		return fmt.Errorf("identity is required")
	}

	// Validate identity before caching
	if err := identity.Validate(); err != nil {
		return fmt.Errorf("invalid identity: %w", err)
	}

	// Serialize to JSON
	data, err := json.Marshal(identity)
	if err != nil {
		return fmt.Errorf("failed to marshal identity to JSON: %w", err)
	}

	// Define cache key (e.g., "customer:cust123")
	key := fmt.Sprintf("customer:%s", identity.GetCustomerID())

	// Set with 1-hour TTL
	err = c.client.Set(ctx, key, data, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set identity in redis for key %s: %w", key, err)
	}

	return nil
}

// GetCustomerIdentity retrieves a CustomerIdentity from Redis by customer ID.
// It returns the deserialized identity or nil if not found.
// Returns an error if retrieval or deserialization fails.
func (c *cache) GetCustomerIdentity(ctx context.Context, customerID string) (*aggregates.CustomerIdentity, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customerID is required")
	}

	// Define cache key
	key := fmt.Sprintf("customer:%s", customerID)

	// Retrieve from Redis
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// Cache miss, not an error
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get identity from redis for key %s: %w", key, err)
	}

	// Deserialize from JSON
	var identity aggregates.CustomerIdentity
	if err = json.Unmarshal(data, &identity); err != nil {
		return nil, fmt.Errorf("failed to unmarshal identity from JSON for key %s: %w", key, err)
	}

	return &identity, nil
}

// Close terminates the Redis client connection.
// It should be called when the cache is no longer needed to free resources.
// Returns an error if closing fails.
func (c *cache) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("failed to close redis client: %w", err)
	}
	return nil
}

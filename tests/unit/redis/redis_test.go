package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/customer-id.git/internal/domain/aggregates"
	"github.com/sukryu/customer-id.git/internal/domain/entities"
	"github.com/sukryu/customer-id.git/internal/infrastructure/redis"
)

func TestCache(t *testing.T) {
	// Setup Redis cache (assuming Redis is running locally via docker-compose)
	cache, err := redis.NewCache("localhost:6379", "redisecret", 0)
	assert.NoError(t, err, "Failed to create Redis cache")
	defer cache.Close()

	// Create test data
	cust, _ := entities.NewCustomer("cust123", nil)
	cust.LastSeen = time.Now().UTC().Add(-2 * time.Minute)
	beacon, _ := entities.NewBeacon("550e8400-e29b-41d4-a716-446655440000", "store100", 100, 3, "Table 3", entities.StatusActive)
	identity, _ := aggregates.NewCustomerIdentity(cust, beacon, 0.95, time.Now().UTC())

	// Test SetCustomerIdentity
	ctx := context.Background()
	err = cache.SetCustomerIdentity(ctx, identity)
	assert.NoError(t, err, "Failed to set CustomerIdentity in Redis")

	// Test GetCustomerIdentity
	retrieved, err := cache.GetCustomerIdentity(ctx, "cust123")
	assert.NoError(t, err, "Failed to get CustomerIdentity from Redis")
	assert.NotNil(t, retrieved, "Retrieved identity should not be nil")
	assert.Equal(t, "cust123", retrieved.GetCustomerID(), "CustomerID mismatch")
	assert.Equal(t, "Table 3", retrieved.GetLocation(), "Location mismatch")
}

func TestGetCustomerIdentityNotFound(t *testing.T) {
	cache, err := redis.NewCache("localhost:6379", "redisecret", 0)
	assert.NoError(t, err)
	defer cache.Close()

	ctx := context.Background()
	retrieved, err := cache.GetCustomerIdentity(ctx, "nonexistent")
	assert.NoError(t, err, "Expected no error for cache miss")
	assert.Nil(t, retrieved, "Expected nil for non-existent key")
}

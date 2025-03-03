package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/customer-id.git/internal/domain/entities"
	"github.com/sukryu/customer-id.git/internal/infrastructure/db"
)

func TestPostgresStorage(t *testing.T) {
	// Setup PostgreSQL storage (assuming local Docker PostgreSQL is running)
	ctx := context.Background()
	connString := "postgres://tastesync:secret@localhost:5432/tastesync?sslmode=disable"
	storage, err := db.NewPostgresStorage(ctx, connString)
	assert.NoError(t, err, "Failed to create PostgresStorage")
	defer storage.Close()

	// Test Save and FindByID for Customer
	cust, _ := entities.NewCustomer("cust123", map[string]string{"drink": "coffee"})
	cust.LastSeen = time.Now().UTC().Add(-2 * time.Minute)
	err = storage.Save(ctx, cust)
	assert.NoError(t, err, "Failed to save customer")

	retrievedCust, err := storage.FindByID(ctx, "cust123")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCust)
	assert.Equal(t, "cust123", retrievedCust.CustomerID)
	assert.Equal(t, "coffee", retrievedCust.Preferences["drink"])

	// Test Save and FindByUUID for Beacon
	beacon, err := entities.NewBeacon("550e8400-e29b-41d4-a716-446655440000", "store100", 100, 3, "Table 3", entities.StatusActive)
	assert.NoError(t, err, "Failed to create beacon")
	err = storage.SaveBeacon(ctx, beacon) // Save beacon to database
	assert.NoError(t, err, "Failed to save beacon")

	retrievedBeacon, err := storage.FindByUUID(ctx, "550e8400-e29b-41d4-a716-446655440000")
	assert.NoError(t, err, "Failed to retrieve beacon")
	assert.NotNil(t, retrievedBeacon, "Retrieved beacon should not be nil")
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", retrievedBeacon.BeaconID, "BeaconID mismatch")
	assert.Equal(t, "Table 3", retrievedBeacon.Location, "Location mismatch")
}

func TestFindByIDNotFound(t *testing.T) {
	ctx := context.Background()
	connString := "postgres://tastesync:secret@localhost:5432/tastesync?sslmode=disable"
	storage, err := db.NewPostgresStorage(ctx, connString)
	assert.NoError(t, err)
	defer storage.Close()

	retrieved, err := storage.FindByID(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

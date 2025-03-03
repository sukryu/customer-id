package aggregates_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/customer-id.git/internal/domain/aggregates"
	"github.com/sukryu/customer-id.git/internal/domain/entities"
)

func TestNewCustomerIdentity(t *testing.T) {
	// Setup test data
	cust, err := entities.NewCustomer("cust123", nil)
	assert.NoError(t, err, "Failed to create customer")

	// Set LastSeen to a time more than 1 minute ago to avoid duplicate check
	cust.LastSeen = time.Now().UTC().Add(-2 * time.Minute)

	beacon, err := entities.NewBeacon("550e8400-e29b-41d4-a716-446655440000", "store100", 100, 3, "Table 3", entities.StatusActive)
	assert.NoError(t, err, "Failed to create beacon")

	// Create CustomerIdentity
	detectedAt := time.Now().UTC()
	ci, err := aggregates.NewCustomerIdentity(cust, beacon, 0.95, detectedAt)
	if !assert.NoError(t, err, "Expected no error creating CustomerIdentity") {
		t.Logf("CustomerIdentity creation failed: %v", err)
		return
	}
	if !assert.NotNil(t, ci, "CustomerIdentity should not be nil") {
		return
	}

	// Validate config values
	assert.Equal(t, "cust123", ci.CustomerID(), "CustomerID mismatch")
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", ci.BeaconID(), "BeaconID mismatch")
	assert.Equal(t, "Table 3", ci.Location(), "Location mismatch")
	assert.Equal(t, float32(0.95), ci.Confidence(), "Confidence mismatch")
	assert.Equal(t, detectedAt, ci.DetectedAt(), "DetectedAt mismatch")

	// Validate
	err = ci.Validate()
	assert.NoError(t, err, "Validation should pass")
}

func TestNewCustomerIdentityLowConfidence(t *testing.T) {
	cust, _ := entities.NewCustomer("cust123", nil)
	cust.LastSeen = time.Now().UTC().Add(-2 * time.Minute) // Avoid duplicate check
	beacon, _ := entities.NewBeacon("550e8400-e29b-41d4-a716-446655440000", "store100", 100, 3, "Table 3", entities.StatusActive)
	_, err := aggregates.NewCustomerIdentity(cust, beacon, 0.7, time.Now().UTC())
	assert.Error(t, err, "Expected error for low confidence")
	assert.Contains(t, err.Error(), "confidence must be at least 0.8", "Error should indicate low confidence")
}

func TestNewCustomerIdentityDuplicate(t *testing.T) {
	cust, _ := entities.NewCustomer("cust123", nil)
	beacon, _ := entities.NewBeacon("550e8400-e29b-41d4-a716-446655440000", "store100", 100, 3, "Table 3", entities.StatusActive)
	now := time.Now().UTC()
	cust.LastSeen = now.Add(-30 * time.Second) // Within 1 minute
	_, err := aggregates.NewCustomerIdentity(cust, beacon, 0.95, now)
	assert.Error(t, err, "Expected error for duplicate identification")
	assert.Contains(t, err.Error(), "duplicate identification within 1 minute", "Error should indicate duplicate")
}

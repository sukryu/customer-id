package services_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/sukryu/customer-id.git/internal/domain/entities"
	"github.com/sukryu/customer-id.git/internal/domain/services"
)

type mockCustomerRepo struct {
	customers map[string]*entities.Customer
}

func (r *mockCustomerRepo) FindByID(customerID string) (*entities.Customer, error) {
	cust, exists := r.customers[customerID]
	if !exists {
		return nil, nil
	}
	return cust, nil
}

func (r *mockCustomerRepo) Save(customer *entities.Customer) error {
	r.customers[customer.CustomerID] = customer
	return nil
}

type mockBeaconRepo struct {
	beacons map[string]*entities.Beacon
}

func (r *mockBeaconRepo) FindByUUID(uuid string) (*entities.Beacon, error) {
	beacon, exists := r.beacons[uuid]
	if !exists {
		return nil, nil
	}
	return beacon, nil
}

func TestIdentifyCustomer(t *testing.T) {
	// Setup mock repositories
	customerRepo := &mockCustomerRepo{customers: make(map[string]*entities.Customer)}
	beaconRepo := &mockBeaconRepo{
		beacons: map[string]*entities.Beacon{
			"550e8400-e29b-41d4-a716-446655440000": {
				BeaconID: "550e8400-e29b-41d4-a716-446655440000",
				StoreID:  "store100",
				Major:    100,
				Minor:    3,
				Location: "Table 3",
				Status:   entities.StatusActive,
			},
		},
	}

	// Create service
	svc, err := services.NewIdentificationService(customerRepo, beaconRepo)
	assert.NoError(t, err, "Failed to create IdentificationService")

	// Test identification
	beaconData, err := entities.NewBeaconData("550e8400-e29b-41d4-a716-446655440000", 100, 3, -20)
	assert.NoError(t, err, "Failed to create BeaconData")

	// Ensure customer.LastSeen is at least 1 minute ago to avoid duplicate check
	customerID := services.GenerateCustomerID(beaconData)
	cust, err := entities.NewCustomer(customerID, nil)
	assert.NoError(t, err)
	cust.LastSeen = time.Now().UTC().Add(-2 * time.Minute) // 2 minutes ago
	err = customerRepo.Save(cust)
	assert.NoError(t, err)

	identity, err := svc.IdentifyCustomer(beaconData)
	if !assert.NoError(t, err, "Expected no error identifying customer") {
		t.Logf("IdentifyCustomer failed: %v", err)
		return
	}
	if !assert.NotNil(t, identity, "CustomerIdentity should not be nil") {
		return
	}

	// Validate results
	assert.Equal(t, customerID, identity.GetCustomerID(), "CustomerID mismatch")
	assert.Equal(t, "Table 3", identity.GetLocation(), "Location mismatch")
	assert.True(t, identity.GetConfidence() >= 0.8, "Confidence should be >= 0.8")
	assert.WithinDuration(t, time.Now().UTC(), identity.GetDetectedAt(), time.Second, "DetectedAt mismatch")
}

func TestIdentifyCustomerInactiveBeacon(t *testing.T) {
	customerRepo := &mockCustomerRepo{customers: make(map[string]*entities.Customer)}
	beaconRepo := &mockBeaconRepo{
		beacons: map[string]*entities.Beacon{
			"550e8400-e29b-41d4-a716-446655440000": {
				BeaconID: "550e8400-e29b-41d4-a716-446655440000",
				StoreID:  "store100",
				Major:    100,
				Minor:    3,
				Location: "Table 3",
				Status:   entities.StatusInactive,
			},
		},
	}

	svc, err := services.NewIdentificationService(customerRepo, beaconRepo)
	assert.NoError(t, err)

	beaconData, err := entities.NewBeaconData("550e8400-e29b-41d4-a716-446655440000", 100, 3, -20)
	assert.NoError(t, err)

	_, err = svc.IdentifyCustomer(beaconData)
	assert.Error(t, err, "Expected error for inactive beacon")
	// Check if error message contains the relevant substring
	assert.Contains(t, err.Error(), "not active", "Error should indicate inactive beacon")
}

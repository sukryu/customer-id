package services

import (
	"fmt"
	"time"

	"github.com/sukryu/customer-id.git/internal/domain/aggregates"
	"github.com/sukryu/customer-id.git/internal/domain/entities"
)

// IdentificationService defines the interface for customer identification logic.
// It provides methods to identify customers based on beacon data.
type IdentificationService interface {
	IdentifyCustomer(beaconData entities.BeaconData) (*aggregates.CustomerIdentity, error)
}

// identificationService implements the IdentificationService interface.
// It orchestrates customer identification by validating beacon data, retrieving or creating
// customer and beacon entities, and enforcing domain rules.
type identificationService struct {
	customerRepo CustomerRepository // Repository for customer data access
	beaconRepo   BeaconRepository   // Repository for beacon data access
}

// CustomerRepository defines the interface for customer data operations.
// This abstraction allows decoupling from specific storage implementations.
type CustomerRepository interface {
	FindByID(customerID string) (*entities.Customer, error)
	Save(customer *entities.Customer) error
}

// BeaconRepository defines the interface for beacon data operations.
// This abstraction supports querying and updating beacon entities.
type BeaconRepository interface {
	FindByUUID(uuid string) (*entities.Beacon, error)
}

// NewIdentificationService creates a new instance of identificationService.
// It requires customer and beacon repositories to perform identification.
// Returns an error if dependencies are invalid.
func NewIdentificationService(customerRepo CustomerRepository, beaconRepo BeaconRepository) (IdentificationService, error) {
	if customerRepo == nil {
		return nil, fmt.Errorf("customer repository is required")
	}
	if beaconRepo == nil {
		return nil, fmt.Errorf("beacon repository is required")
	}
	return &identificationService{
		customerRepo: customerRepo,
		beaconRepo:   beaconRepo,
	}, nil
}

// IdentifyCustomer identifies a customer based on the provided beacon data.
// It retrieves or creates the associated customer and beacon entities, calculates
// identification confidence, and enforces domain rules (e.g., minimum confidence, no duplicates).
// Returns a CustomerIdentity instance or an error if identification fails.
func (s *identificationService) IdentifyCustomer(beaconData entities.BeaconData) (*aggregates.CustomerIdentity, error) {
	// Validate beacon data
	if err := beaconData.Validate(); err != nil {
		return nil, fmt.Errorf("invalid beacon data: %w", err)
	}

	// Retrieve beacon entity
	beacon, err := s.beaconRepo.FindByUUID(beaconData.UUID())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve beacon: %w", err)
	}
	if beacon == nil {
		return nil, fmt.Errorf("beacon not found for UUID: %s", beaconData.UUID())
	}
	if beacon.Status != entities.StatusActive {
		return nil, fmt.Errorf("beacon %s is not active, current status: %s", beacon.BeaconID, beacon.Status)
	}

	// Simple confidence calculation based on RSSI (production would use more sophisticated logic)
	confidence := calculateConfidence(beaconData.RSSI())
	if confidence < 0.8 {
		return nil, fmt.Errorf("identification confidence %f below minimum threshold of 0.8", confidence)
	}

	// Retrieve or create customer (simplified logic for initial implementation)
	customerID := GenerateCustomerID(beaconData) // Placeholder for actual logic
	customer, err := s.customerRepo.FindByID(customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve customer: %w", err)
	}
	if customer == nil {
		customer, err = entities.NewCustomer(customerID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create new customer: %w", err)
		}
		if err = s.customerRepo.Save(customer); err != nil {
			return nil, fmt.Errorf("failed to save new customer: %w", err)
		}
	}

	// Create CustomerIdentity with current timestamp
	detectedAt := time.Now().UTC()
	identity, err := aggregates.NewCustomerIdentity(customer, beacon, confidence, detectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create customer identity: %w", err)
	}

	// Update customer's LastSeen timestamp
	customer.UpdateLastSeen()
	if err = s.customerRepo.Save(customer); err != nil {
		return nil, fmt.Errorf("failed to update customer last seen: %w", err)
	}

	return identity, nil
}

// calculateConfidence computes a simple confidence score based on RSSI.
// For production, this could integrate more complex algorithms (e.g., distance, signal quality).
// Returns a value between 0.0 and 1.0, where higher RSSI (closer to 0) yields higher confidence.
func calculateConfidence(rssi int32) float32 {
	// Normalize RSSI (-100 to 0) to confidence (0.0 to 1.0)
	// Example: -100 -> 0.0, -50 -> 0.5, 0 -> 1.0
	normalized := float32(rssi+100) / 100.0
	if normalized < 0.0 {
		return 0.0
	}
	if normalized > 1.0 {
		return 1.0
	}
	return normalized
}

// generateCustomerID generates a simple customer ID based on beacon data.
// This is a placeholder; production would use a more robust ID generation strategy (e.g., UUID).
func GenerateCustomerID(beaconData entities.BeaconData) string {
	return fmt.Sprintf("cust-%s-%d-%d", beaconData.UUID(), beaconData.Major(), beaconData.Minor())
}

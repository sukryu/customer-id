package aggregates

import (
	"fmt"
	"time"

	"github.com/sukryu/customer-id.git/internal/domain/entities"
)

// CustomerIdentity represents the aggregate root for customer identification.
// It combines Customer and Beacon entities with additional identification data,
// enforcing domain rules such as confidence threshold and duplicate detection prevention.
type CustomerIdentity struct {
	customerID string    // Unique identifier of the customer (references Customer).
	beaconID   string    // Unique identifier of the beacon (references Beacon).
	location   string    // Identified location (e.g., "Table 3").
	confidence float32   // Confidence score of identification (0.0 to 1.0).
	detectedAt time.Time // Timestamp of identification (UTC).
}

// NewCustomerIdentity creates a new CustomerIdentity instance.
// It validates the provided customer and beacon data against domain rules,
// including a minimum confidence threshold of 0.8.
// Returns an error if validation fails.
func NewCustomerIdentity(customer *entities.Customer, beacon *entities.Beacon, confidence float32, detectedAt time.Time) (*CustomerIdentity, error) {
	// Validate input entities
	if customer == nil {
		return nil, fmt.Errorf("customer entity is required")
	}
	if err := customer.Validate(); err != nil {
		return nil, fmt.Errorf("invalid customer: %w", err)
	}
	if beacon == nil {
		return nil, fmt.Errorf("beacon entity is required")
	}
	if err := beacon.Validate(); err != nil {
		return nil, fmt.Errorf("invalid beacon: %w", err)
	}

	// Validate confidence score
	if confidence < 0.0 || confidence > 1.0 {
		return nil, fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", confidence)
	}
	const minConfidence = 0.8
	if confidence < minConfidence {
		return nil, fmt.Errorf("confidence must be at least %f, got %f", minConfidence, confidence)
	}

	// Validate detectedAt
	if detectedAt.IsZero() {
		return nil, fmt.Errorf("detectedAt must be set")
	}
	// Prevent duplicate identification within 1 minute
	timeSinceLastSeen := detectedAt.Sub(customer.LastSeen)
	if timeSinceLastSeen < time.Minute {
		return nil, fmt.Errorf("duplicate identification within 1 minute: last seen %v, detected at %v", customer.LastSeen, detectedAt)
	}

	return &CustomerIdentity{
		customerID: customer.CustomerID,
		beaconID:   beacon.BeaconID,
		location:   beacon.Location,
		confidence: confidence,
		detectedAt: detectedAt,
	}, nil
}

// CustomerID returns the customer identifier.
// This method provides read-only access to the customerID field.
func (ci *CustomerIdentity) CustomerID() string {
	return ci.customerID
}

// BeaconID returns the beacon identifier.
// This method provides read-only access to the beaconID field.
func (ci *CustomerIdentity) BeaconID() string {
	return ci.beaconID
}

// Location returns the identified location.
// This method provides read-only access to the location field.
func (ci *CustomerIdentity) Location() string {
	return ci.location
}

// Confidence returns the identification confidence score.
// This method provides read-only access to the confidence field.
func (ci *CustomerIdentity) Confidence() float32 {
	return ci.confidence
}

// DetectedAt returns the timestamp of identification.
// This method provides read-only access to the detectedAt field.
func (ci *CustomerIdentity) DetectedAt() time.Time {
	return ci.detectedAt
}

// Validate ensures the CustomerIdentity meets all domain constraints.
// Returns an error if any constraint is violated.
func (ci *CustomerIdentity) Validate() error {
	if ci.customerID == "" {
		return fmt.Errorf("customerID is required")
	}
	if len(ci.customerID) > 64 {
		return fmt.Errorf("customerID exceeds maximum length of 64 characters")
	}
	if ci.beaconID == "" {
		return fmt.Errorf("beaconID is required")
	}
	if len(ci.beaconID) != 36 { // UUID format check
		return fmt.Errorf("beaconID must be a valid UUID (36 characters)")
	}
	if len(ci.location) > 32 {
		return fmt.Errorf("location exceeds maximum length of 32 characters")
	}
	if ci.confidence < 0.0 || ci.confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", ci.confidence)
	}
	const minConfidence = 0.8
	if ci.confidence < minConfidence {
		return fmt.Errorf("confidence must be at least %f, got %f", minConfidence, ci.confidence)
	}
	if ci.detectedAt.IsZero() {
		return fmt.Errorf("detectedAt must be set")
	}
	return nil
}

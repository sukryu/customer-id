package aggregates

import (
	"fmt"
	"time"

	"github.com/sukryu/customer-id.git/internal/domain/entities"
)

// CustomerIdentity represents the aggregate root for customer identification.
type CustomerIdentity struct {
	CustomerID string    // Unique identifier of the customer (references Customer).
	BeaconID   string    // Unique identifier of the beacon (references Beacon).
	Location   string    // Identified location (e.g., "Table 3").
	Confidence float32   // Confidence score of identification (0.0 to 1.0).
	DetectedAt time.Time // Timestamp of identification (UTC).
}

// NewCustomerIdentity creates a new CustomerIdentity instance.
func NewCustomerIdentity(customer *entities.Customer, beacon *entities.Beacon, confidence float32, detectedAt time.Time) (*CustomerIdentity, error) {
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

	if confidence < 0.0 || confidence > 1.0 {
		return nil, fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", confidence)
	}
	const minConfidence = 0.8
	if confidence < minConfidence {
		return nil, fmt.Errorf("confidence must be at least %f, got %f", minConfidence, confidence)
	}

	if detectedAt.IsZero() {
		return nil, fmt.Errorf("detectedAt must be set")
	}
	timeSinceLastSeen := detectedAt.Sub(customer.LastSeen)
	if timeSinceLastSeen < time.Minute {
		return nil, fmt.Errorf("duplicate identification within 1 minute: last seen %v, detected at %v", customer.LastSeen, detectedAt)
	}

	return &CustomerIdentity{
		CustomerID: customer.CustomerID,
		BeaconID:   beacon.BeaconID,
		Location:   beacon.Location,
		Confidence: confidence,
		DetectedAt: detectedAt,
	}, nil
}

// (기존 접근자 및 Validate 메서드 유지, 필드명만 대문자로 변경 반영)
// CustomerID returns the customer identifier.
func (ci *CustomerIdentity) GetCustomerID() string {
	return ci.CustomerID
}

// BeaconID returns the beacon identifier.
func (ci *CustomerIdentity) GetBeaconID() string {
	return ci.BeaconID
}

// Location returns the identified location.
func (ci *CustomerIdentity) GetLocation() string {
	return ci.Location
}

// Confidence returns the identification confidence score.
func (ci *CustomerIdentity) GetConfidence() float32 {
	return ci.Confidence
}

// DetectedAt returns the timestamp of identification.
func (ci *CustomerIdentity) GetDetectedAt() time.Time {
	return ci.DetectedAt
}

// Validate ensures the CustomerIdentity meets all domain constraints.
func (ci *CustomerIdentity) Validate() error {
	if ci.CustomerID == "" {
		return fmt.Errorf("customerID is required")
	}
	if len(ci.CustomerID) > 64 {
		return fmt.Errorf("customerID exceeds maximum length of 64 characters")
	}
	if ci.BeaconID == "" {
		return fmt.Errorf("beaconID is required")
	}
	if len(ci.BeaconID) != 36 {
		return fmt.Errorf("beaconID must be a valid UUID (36 characters)")
	}
	if len(ci.Location) > 32 {
		return fmt.Errorf("location exceeds maximum length of 32 characters")
	}
	if ci.Confidence < 0.0 || ci.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", ci.Confidence)
	}
	const minConfidence = 0.8
	if ci.Confidence < minConfidence {
		return fmt.Errorf("confidence must be at least %f, got %f", minConfidence, ci.Confidence)
	}
	if ci.DetectedAt.IsZero() {
		return fmt.Errorf("detectedAt must be set")
	}
	return nil
}

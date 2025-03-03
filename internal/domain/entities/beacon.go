package entities

import (
	"fmt"
)

// BeaconStatus defines the possible states of a beacon device.
type BeaconStatus string

const (
	// StatusActive indicates the beacon is operational.
	StatusActive BeaconStatus = "active"
	// StatusInactive indicates the beacon is not operational.
	StatusInactive BeaconStatus = "inactive"
	// StatusMaintenance indicates the beacon is under maintenance.
	StatusMaintenance BeaconStatus = "maintenance"
)

// Beacon represents a physical beacon device used for customer identification.
// It includes unique identification, store association, and operational status.
type Beacon struct {
	BeaconID string       // Unique identifier (UUID, e.g., "550e8400-e29b-41d4-a716-446655440000").
	StoreID  string       // Store identifier (e.g., "store100").
	Major    int32        // Major group identifier (0-65535, e.g., store section).
	Minor    int32        // Minor location identifier (0-65535, e.g., table number).
	Location string       // Physical location description (e.g., "Table 3").
	Status   BeaconStatus // Operational status (active, inactive, maintenance).
}

// NewBeacon creates a new Beacon instance with the given parameters.
// It validates required fields and constraints, returning an error if invalid.
// Default status is set to "active" if not specified.
func NewBeacon(beaconID, storeID string, major, minor int32, location string, status BeaconStatus) (*Beacon, error) {
	if beaconID == "" {
		return nil, fmt.Errorf("beaconID is required")
	}
	if len(beaconID) != 36 { // UUID format check (e.g., 8-4-4-4-12)
		return nil, fmt.Errorf("beaconID must be a valid UUID (36 characters)")
	}
	if storeID == "" {
		return nil, fmt.Errorf("storeID is required")
	}
	if major < 0 || major > 65535 {
		return nil, fmt.Errorf("major must be between 0 and 65535")
	}
	if minor < 0 || minor > 65535 {
		return nil, fmt.Errorf("minor must be between 0 and 65535")
	}
	if len(location) > 32 {
		return nil, fmt.Errorf("location exceeds maximum length of 32 characters")
	}
	if status == "" {
		status = StatusActive // Default to active
	}
	switch status {
	case StatusActive, StatusInactive, StatusMaintenance:
		// Valid status
	default:
		return nil, fmt.Errorf("invalid status: %s, must be one of active, inactive, maintenance", status)
	}

	return &Beacon{
		BeaconID: beaconID,
		StoreID:  storeID,
		Major:    major,
		Minor:    minor,
		Location: location,
		Status:   status,
	}, nil
}

// SetStatus updates the beacon's operational status.
// It ensures the new status is valid before applying the change.
func (b *Beacon) SetStatus(status BeaconStatus) error {
	switch status {
	case StatusActive, StatusInactive, StatusMaintenance:
		b.Status = status
		return nil
	default:
		return fmt.Errorf("invalid status: %s, must be one of active, inactive, maintenance", status)
	}
}

// Validate ensures the Beacon entity meets all domain constraints.
// Returns an error if any constraint is violated.
func (b *Beacon) Validate() error {
	if b.BeaconID == "" {
		return fmt.Errorf("beaconID is required")
	}
	if len(b.BeaconID) != 36 {
		return fmt.Errorf("beaconID must be a valid UUID (36 characters)")
	}
	if b.StoreID == "" {
		return fmt.Errorf("storeID is required")
	}
	if b.Major < 0 || b.Major > 65535 {
		return fmt.Errorf("major must be between 0 and 65535")
	}
	if b.Minor < 0 || b.Minor > 65535 {
		return fmt.Errorf("minor must be between 0 and 65535")
	}
	if len(b.Location) > 32 {
		return fmt.Errorf("location exceeds maximum length of 32 characters")
	}
	switch b.Status {
	case StatusActive, StatusInactive, StatusMaintenance:
		return nil
	default:
		return fmt.Errorf("invalid status: %s", b.Status)
	}
}

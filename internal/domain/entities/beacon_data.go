package entities

import (
	"fmt"
)

// BeaconData represents the raw data received from a beacon device.
// As a value object, it is immutable and encapsulates the beacon's identification
// and signal strength data used for customer identification.
type BeaconData struct {
	uuid  string // Unique identifier of the beacon (UUID format, e.g., "550e8400-e29b-41d4-a716-446655440000").
	major int32  // Major group identifier (0-65535, e.g., store section).
	minor int32  // Minor location identifier (0-65535, e.g., table number).
	rssi  int32  // Received Signal Strength Indicator (-100 to 0 dBm).
}

// NewBeaconData creates a new BeaconData instance with the provided values.
// It enforces domain constraints such as UUID format and valid ranges for major, minor, and RSSI.
// Returns an error if any constraint is violated.
func NewBeaconData(uuid string, major, minor, rssi int32) (BeaconData, error) {
	// Validate UUID
	if uuid == "" {
		return BeaconData{}, fmt.Errorf("uuid is required")
	}
	if len(uuid) != 36 { // UUID format: 8-4-4-4-12
		return BeaconData{}, fmt.Errorf("uuid must be a valid UUID (36 characters), got %d", len(uuid))
	}

	// Validate major and minor
	if major < 0 || major > 65535 {
		return BeaconData{}, fmt.Errorf("major must be between 0 and 65535, got %d", major)
	}
	if minor < 0 || minor > 65535 {
		return BeaconData{}, fmt.Errorf("minor must be between 0 and 65535, got %d", minor)
	}

	// Validate RSSI
	if rssi < -100 || rssi > 0 {
		return BeaconData{}, fmt.Errorf("rssi must be between -100 and 0 dBm, got %d", rssi)
	}

	return BeaconData{
		uuid:  uuid,
		major: major,
		minor: minor,
		rssi:  rssi,
	}, nil
}

// UUID returns the beacon's unique identifier.
// This method provides read-only access to the uuid field.
func (bd BeaconData) UUID() string {
	return bd.uuid
}

// Major returns the beacon's major group identifier.
// This method provides read-only access to the major field.
func (bd BeaconData) Major() int32 {
	return bd.major
}

// Minor returns the beacon's minor location identifier.
// This method provides read-only access to the minor field.
func (bd BeaconData) Minor() int32 {
	return bd.minor
}

// RSSI returns the beacon's signal strength.
// This method provides read-only access to the rssi field.
func (bd BeaconData) RSSI() int32 {
	return bd.rssi
}

// Validate ensures the BeaconData meets all domain constraints.
// Returns an error if any constraint is violated.
func (bd BeaconData) Validate() error {
	if bd.uuid == "" {
		return fmt.Errorf("uuid is required")
	}
	if len(bd.uuid) != 36 {
		return fmt.Errorf("uuid must be a valid UUID (36 characters), got %d", len(bd.uuid))
	}
	if bd.major < 0 || bd.major > 65535 {
		return fmt.Errorf("major must be between 0 and 65535, got %d", bd.major)
	}
	if bd.minor < 0 || bd.minor > 65535 {
		return fmt.Errorf("minor must be between 0 and 65535, got %d", bd.minor)
	}
	if bd.rssi < -100 || bd.rssi > 0 {
		return fmt.Errorf("rssi must be between -100 and 0 dBm, got %d", bd.rssi)
	}
	return nil
}

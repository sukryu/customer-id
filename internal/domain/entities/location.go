package entities

import (
	"fmt"
)

// LocationType defines the possible types of a location within a store.
type LocationType string

const (
	// LocationTypeEntrance represents an entrance area.
	LocationTypeEntrance LocationType = "entrance"
	// LocationTypeTable represents a table or seating area.
	LocationTypeTable LocationType = "table"
	// LocationTypeCounter represents a counter or service area.
	LocationTypeCounter LocationType = "counter"
)

// Location represents a physical location within a store as a value object.
// It is immutable and used to describe where a customer is identified.
type Location struct {
	name  string       // Name of the location (e.g., "Table 3").
	type_ LocationType // Type of the location (entrance, table, counter).
}

// NewLocation creates a new Location instance with the given name and type.
// It enforces domain constraints such as maximum length and valid location types.
// Returns an error if any constraint is violated.
func NewLocation(name string, locationType LocationType) (Location, error) {
	// Validate name
	if name == "" {
		return Location{}, fmt.Errorf("name is required")
	}
	if len(name) > 32 {
		return Location{}, fmt.Errorf("name exceeds maximum length of 32 characters, got %d", len(name))
	}

	// Validate location type
	switch locationType {
	case LocationTypeEntrance, LocationTypeTable, LocationTypeCounter:
		// Valid type
	default:
		return Location{}, fmt.Errorf("invalid location type: %s, must be one of entrance, table, counter", locationType)
	}

	return Location{
		name:  name,
		type_: locationType,
	}, nil
}

// Name returns the location's name.
// This method provides read-only access to the name field.
func (l Location) Name() string {
	return l.name
}

// Type returns the location's type.
// This method provides read-only access to the type_ field.
func (l Location) Type() LocationType {
	return l.type_
}

// Validate ensures the Location meets all domain constraints.
// Returns an error if any constraint is violated.
func (l Location) Validate() error {
	if l.name == "" {
		return fmt.Errorf("name is required")
	}
	if len(l.name) > 32 {
		return fmt.Errorf("name exceeds maximum length of 32 characters, got %d", len(l.name))
	}
	switch l.type_ {
	case LocationTypeEntrance, LocationTypeTable, LocationTypeCounter:
		return nil
	default:
		return fmt.Errorf("invalid location type: %s", l.type_)
	}
}

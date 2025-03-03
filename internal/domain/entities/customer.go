package entities

import (
	"fmt"
	"time"
)

// Customer represents an identified customer in the TasteSync system.
// This entity encapsulates the customer's unique identity, last seen time,
// and preferences, serving as a core component for customer identification.
type Customer struct {
	CustomerID  string            // Unique identifier for the customer (e.g., "cust123").
	LastSeen    time.Time         // Timestamp of the customer's last identification (UTC).
	Preferences map[string]string // Customer preferences (e.g., {"drink": "coffee"}).
}

// NewCustomer creates a new Customer instance with the given ID and preferences.
// It initializes LastSeen with the current UTC time and validates required fields.
// Returns an error if validation fails.
func NewCustomer(customerID string, preferences map[string]string) (*Customer, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customerID is required")
	}
	if len(customerID) > 64 {
		return nil, fmt.Errorf("customerID exceeds maximum length of 64 characters")
	}
	if preferences == nil {
		preferences = make(map[string]string) // Initialize empty map to avoid nil
	}

	return &Customer{
		CustomerID:  customerID,
		LastSeen:    time.Now().UTC(),
		Preferences: preferences,
	}, nil
}

// UpdateLastSeen updates the customer's LastSeen timestamp to the current UTC time.
// This method ensures the customer's activity is tracked accurately.
func (c *Customer) UpdateLastSeen() {
	c.LastSeen = time.Now().UTC()
}

// AddPreference adds or updates a preference key-value pair for the customer.
// It ensures that preferences remain mutable and extensible.
func (c *Customer) AddPreference(key, value string) error {
	if key == "" {
		return fmt.Errorf("preference key cannot be empty")
	}
	if c.Preferences == nil {
		c.Preferences = make(map[string]string)
	}
	c.Preferences[key] = value
	return nil
}

// Validate ensures the Customer entity meets all domain constraints.
// Returns an error if any constraint is violated.
func (c *Customer) Validate() error {
	if c.CustomerID == "" {
		return fmt.Errorf("customerID is required")
	}
	if len(c.CustomerID) > 64 {
		return fmt.Errorf("customerID exceeds maximum length of 64 characters")
	}
	if c.LastSeen.IsZero() {
		return fmt.Errorf("lastSeen must be set")
	}
	return nil
}

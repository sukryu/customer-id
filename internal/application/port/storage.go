package ports

import (
	"context"

	"github.com/sukryu/customer-id.git/internal/domain/entities"
)

// CustomerRepository defines the interface for customer data operations.
// It provides methods to find and save customer entities in a persistent store.
type CustomerRepository interface {
	// FindByID retrieves a customer by its unique identifier.
	// Returns nil if not found, or an error if the operation fails.
	FindByID(ctx context.Context, customerID string) (*entities.Customer, error)

	// Save persists a customer entity to the store.
	// Returns an error if the operation fails.
	Save(ctx context.Context, customer *entities.Customer) error
}

// BeaconRepository defines the interface for beacon data operations.
// It provides methods to find beacon entities in a persistent store.
type BeaconRepository interface {
	// FindByUUID retrieves a beacon by its unique UUID.
	// Returns nil if not found, or an error if the operation fails.
	FindByUUID(ctx context.Context, uuid string) (*entities.Beacon, error)
}

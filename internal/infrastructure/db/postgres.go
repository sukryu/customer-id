package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	ports "github.com/sukryu/customer-id.git/internal/application/port"
	"github.com/sukryu/customer-id.git/internal/domain/entities"
)

// PostgresStorage implements CustomerRepository and BeaconRepository for PostgreSQL.
// It uses pgxpool for connection pooling and efficient database operations.
type PostgresStorage struct {
	pool *pgxpool.Pool // Connection pool for PostgreSQL
}

// NewPostgresStorage creates a new PostgresStorage instance with the provided configuration.
// It establishes a connection pool to PostgreSQL and verifies connectivity.
// Returns an error if the connection fails or configuration is invalid.
func NewPostgresStorage(ctx context.Context, connString string) (*PostgresStorage, error) {
	if connString == "" {
		return nil, fmt.Errorf("connection string is required")
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx connection pool: %w", err)
	}

	// Verify connection
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return &PostgresStorage{
		pool: pool,
	}, nil
}

// FindByID retrieves a customer by its unique identifier from PostgreSQL.
// Returns nil if not found, or an error if the query fails.
func (s *PostgresStorage) FindByID(ctx context.Context, customerID string) (*entities.Customer, error) {
	if customerID == "" {
		return nil, fmt.Errorf("customerID is required")
	}

	query := `
		SELECT customer_id, last_seen, preferences
		FROM customers
		WHERE customer_id = $1
	`
	var cust entities.Customer
	var preferencesJSON []byte

	err := s.pool.QueryRow(ctx, query, customerID).Scan(&cust.CustomerID, &cust.LastSeen, &preferencesJSON)
	if err == pgx.ErrNoRows {
		return nil, nil // Not found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query customer by ID %s: %w", customerID, err)
	}

	// Unmarshal preferences JSON
	if err = json.Unmarshal(preferencesJSON, &cust.Preferences); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferences for customer %s: %w", customerID, err)
	}

	return &cust, nil
}

// Save persists a customer entity to PostgreSQL.
// It inserts or updates the record based on the customer_id using ON CONFLICT.
// Returns an error if the operation fails.
func (s *PostgresStorage) Save(ctx context.Context, customer *entities.Customer) error {
	if customer == nil {
		return fmt.Errorf("customer is required")
	}
	if err := customer.Validate(); err != nil {
		return fmt.Errorf("invalid customer: %w", err)
	}

	// Marshal preferences to JSON
	preferencesJSON, err := json.Marshal(customer.Preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences for customer %s: %w", customer.CustomerID, err)
	}

	query := `
		INSERT INTO customers (customer_id, last_seen, preferences)
		VALUES ($1, $2, $3)
		ON CONFLICT (customer_id)
		DO UPDATE SET last_seen = EXCLUDED.last_seen, preferences = EXCLUDED.preferences
	`
	_, err = s.pool.Exec(ctx, query, customer.CustomerID, customer.LastSeen, preferencesJSON)
	if err != nil {
		return fmt.Errorf("failed to save customer %s: %w", customer.CustomerID, err)
	}

	return nil
}

// SaveBeacon persists a beacon entity to PostgreSQL.
// It inserts or updates the record based on the beacon_id using ON CONFLICT.
// Returns an error if the operation fails.
func (s *PostgresStorage) SaveBeacon(ctx context.Context, beacon *entities.Beacon) error {
	if beacon == nil {
		return fmt.Errorf("beacon is required")
	}
	if err := beacon.Validate(); err != nil {
		return fmt.Errorf("invalid beacon: %w", err)
	}

	query := `
		INSERT INTO beacons (beacon_id, store_id, major, minor, location, status, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (beacon_id)
		DO UPDATE SET store_id = EXCLUDED.store_id, major = EXCLUDED.major, minor = EXCLUDED.minor, 
		              location = EXCLUDED.location, status = EXCLUDED.status, updated_at = EXCLUDED.updated_at
	`
	_, err := s.pool.Exec(ctx, query,
		beacon.BeaconID,
		beacon.StoreID,
		beacon.Major,
		beacon.Minor,
		beacon.Location,
		beacon.Status,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to save beacon %s: %w", beacon.BeaconID, err)
	}

	return nil
}

// FindByUUID retrieves a beacon by its unique UUID from PostgreSQL.
// Returns nil if not found, or an error if the query fails.
func (s *PostgresStorage) FindByUUID(ctx context.Context, uuid string) (*entities.Beacon, error) {
	if uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}

	query := `
		SELECT beacon_id, store_id, major, minor, location, status
		FROM beacons
		WHERE beacon_id = $1
	`
	var beacon entities.Beacon
	err := s.pool.QueryRow(ctx, query, uuid).Scan(
		&beacon.BeaconID,
		&beacon.StoreID,
		&beacon.Major,
		&beacon.Minor,
		&beacon.Location,
		&beacon.Status,
	)
	if err == pgx.ErrNoRows {
		return nil, nil // Not found, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query beacon by UUID %s: %w", uuid, err)
	}

	return &beacon, nil
}

// Close terminates the PostgreSQL connection pool.
// It should be called when the storage is no longer needed to free resources.
// Returns an error if closing fails.
func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}

// Verify interfaces are implemented
var _ ports.CustomerRepository = (*PostgresStorage)(nil)
var _ ports.BeaconRepository = (*PostgresStorage)(nil)

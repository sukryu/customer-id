-- schema.sql defines the PostgreSQL schema for the TasteSync customer-id service.
-- It includes tables for customers, beacons, and customer identities with
-- appropriate constraints, indexes, and partitions for performance and scalability.

-- Customers table stores customer entities with their unique IDs and preferences.
CREATE TABLE customers (
    customer_id VARCHAR(64) PRIMARY KEY,          -- Unique customer identifier (e.g., "cust123")
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL,  -- Last identification timestamp (UTC)
    preferences JSONB DEFAULT '{}'::jsonb,        -- Customer preferences as JSON (e.g., {"drink": "coffee"})
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP  -- Record creation timestamp
);

-- Index on last_seen for efficient querying of recent customer activity.
CREATE INDEX idx_customers_last_seen ON customers (last_seen);

-- Beacons table stores beacon device entities with their unique IDs and metadata.
CREATE TABLE beacons (
    beacon_id VARCHAR(36) PRIMARY KEY,            -- Unique beacon identifier (UUID, e.g., "550e8400-e29b-41d4-a716-446655440000")
    store_id VARCHAR(64) NOT NULL,                -- Store identifier (e.g., "store100")
    major INT NOT NULL CHECK (major >= 0 AND major <= 65535),  -- Major group (0-65535)
    minor INT NOT NULL CHECK (minor >= 0 AND minor <= 65535),  -- Minor location (0-65535)
    location VARCHAR(32),                         -- Physical location (e.g., "Table 3")
    status VARCHAR(16) NOT NULL DEFAULT 'active', -- Operational status (active, inactive, maintenance)
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,  -- Last update timestamp
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'maintenance'))
);

-- Indexes for efficient querying by store_id and filtering by status.
CREATE INDEX idx_beacons_store_id ON beacons (store_id);
CREATE INDEX idx_beacons_status ON beacons (status);

-- Customer_identities table stores customer identification events as an aggregate.
-- Partitioned by detected_at for scalability with large datasets (e.g., 10M users).
CREATE TABLE customer_identities (
    id BIGSERIAL,                                 -- Unique record ID (auto-incremented)
    customer_id VARCHAR(64) NOT NULL,             -- References customers(customer_id)
    beacon_id VARCHAR(36) NOT NULL,               -- References beacons(beacon_id)
    location VARCHAR(32),                         -- Identified location (e.g., "Table 3")
    confidence REAL NOT NULL CHECK (confidence >= 0 AND confidence <= 1),  -- Confidence score (0.0-1.0)
    detected_at TIMESTAMP WITH TIME ZONE NOT NULL, -- Timestamp of identification (UTC)
    PRIMARY KEY (id, detected_at),                -- Composite primary key with partitioning
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id) ON DELETE CASCADE,
    FOREIGN KEY (beacon_id) REFERENCES beacons(beacon_id) ON DELETE RESTRICT
) PARTITION BY RANGE (detected_at);

-- Partition for 2025 data (extendable for future years).
CREATE TABLE customer_identities_2025 PARTITION OF customer_identities
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

-- Indexes for efficient querying by customer_id and detected_at.
CREATE INDEX idx_customer_identities_customer_id ON customer_identities (customer_id);
CREATE INDEX idx_customer_identities_detected_at ON customer_identities (detected_at);

-- Unique constraint to prevent duplicate identifications within a short time frame.
CREATE UNIQUE INDEX idx_customer_identities_unique ON customer_identities (customer_id, detected_at);
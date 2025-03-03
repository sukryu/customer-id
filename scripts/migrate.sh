#!/bin/bash
# migrate.sh initializes the PostgreSQL database schema for the TasteSync customer-id service.
# It applies the schema.sql file to the specified database using environment variables for configuration.

# 꼭 권한 설정 해줘야 함. chmod +x scripts/migrate.sh

# Default values if environment variables are not set
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-tastesync}
DB_PASSWORD=${DB_PASSWORD:-secret}
DB_NAME=${DB_NAME:-tastesync}

# Export password to avoid interactive prompt
export PGPASSWORD=$DB_PASSWORD

# Apply schema.sql to the database
echo "Applying schema to PostgreSQL database: $DB_NAME at $DB_HOST:$DB_PORT"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f internal/infrastructure/db/schema.sql

# Check if migration was successful
if [ $? -eq 0 ]; then
    echo "Database schema initialized successfully."
else
    echo "Failed to initialize database schema." >&2
    exit 1
fi
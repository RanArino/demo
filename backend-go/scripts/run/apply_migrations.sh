#!/bin/bash
set -e

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "Error: migrate command is not installed."
    echo "Please install golang-migrate: https://github.com/golang-migrate/migrate"
    exit 1
fi

# Get database URL from environment or use default
DB_URL=${DATABASE_URL:-"postgres://postgres:postgres@localhost:5432/document_service?sslmode=disable"}

echo "Using database URL: $DB_URL"
echo "Running migrations..."

# Run migrations
migrate -path ../migrations -database "$DB_URL" up

echo "Migrations completed successfully." 
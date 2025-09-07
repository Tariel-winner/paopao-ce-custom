#!/bin/bash

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "migrate tool is not installed. Please install it first."
    exit 1
fi

# Run migrations
echo "Applying migrations..."
migrate -path ./scripts/migration/postgres -database "postgres://paopao:paopao@localhost:5432/paopao?sslmode=disable" up

echo "Migrations applied successfully." 
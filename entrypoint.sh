#!/bin/bash
set -e  # Exit on error

# Load environment variables
set -a
source .env
set +a

# Extract database host and port from DATABASE_URL
DB_HOST=$(echo $DATABASE_URL | sed -E 's/^.*@([^:/]+):([0-9]+).*$/\1/')
DB_PORT=$(echo $DATABASE_URL | sed -E 's/^.*@([^:/]+):([0-9]+).*$/\2/')

echo "Waiting for database at $DB_HOST:$DB_PORT..."

# Wait for database to be ready
until nc -z $DB_HOST $DB_PORT; do
  echo "Database is unavailable - sleeping"
  sleep 2
done

echo "Database is up! Running migrations..."

# Run migrations
goose up

echo "Starting application..."
exec "$@"


#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Run database migrations
make goose-up

# Execute the main application exec "$@"

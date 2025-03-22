DEFAULT_GOAL := run
ENV_FILE := .env
include $(ENV_FILE)
export

# Build the application
## build: build the application
build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

## swag
swag:
	swag init -g ./cmd/api/main.go --parseDependency -o ./docs/swagger/

## run: run the application
run: 
	@go run cmd/api/main.go
	 

reset-db: goose-down goose-up

## goose-create: create a new goose migration
goose-create:
	goose -s create init -dir ./internal/database/migrations/ sql

## goose-up: apply goose migrations
goose-up:
	goose up

## goose-status: show goose migration status
goose-status:
	goose status

## goose-down: rollback all goose migrations
goose-down:
	goose down-to 0

# Manage DB containers
## up: create and start the database container
up:
	docker compose up -d

## down: stop and remove the database container
down:
	docker compose down -v

## sqlc: generate SQL code using sqlc
sqlc:
	sqlc generate

# Manage Docker containers
## docker-down: shutdown DB container (fallback to Docker Compose V1 if needed)
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Testing
## test: run all tests
test:
	@echo "Testing..."
	@go test -cover ./... -v 

## itest: run integration tests
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the application
## clean: clean the compiled binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
## watch: live reload using air (installs air if not available)
watch:
	@if command -v air > /dev/null; then \
		air; \
		echo "Watching...";\
	else \
		read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/air-verse/air@latest; \
			air; \
			echo "Watching...";\
		else \
			echo "You chose not to install air. Exiting..."; \
			exit 1; \
		fi; \
	fi


.PHONY: all build test clean watch docker-run docker-down itest run run-prod watch docker-up docker-down up down sqlc

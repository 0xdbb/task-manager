DEFAULT_GOAL := run
ENV_FILE := .env
include $(ENV_FILE)
export

# Build the application
## build: build the application
build:
	@echo "Building..."
	@go build -o ./bin/main cmd/api/main.go

## swag
swag:
	swag init -g ./cmd/api/main.go --parseDependency -o ./docs/swagger/

## run: run the application
run: 
	@go run cmd/api/main.go
	 
run-worker:
	@go run cmd/worker/main.go

sqlc:
	sqlc generate

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
	docker compose up -d --build
	

## down: stop and remove the database container
down:
	docker compose down -v

test:
	go test -v ./...

# watch: live reload using air (installs air if not available)
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

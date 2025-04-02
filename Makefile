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



# Replace "your-dockerhub-username" with your actual Docker Hub username
docker-build:
	docker build -t dennislazy/task-manager-api:latest -t dennislazy/task-manager:v1 .

docker-push:
	docker push dennislazy/task-manager:latest

## sqlc: generate SQL code using sqlc
sqlc:
	sqlc generate

apply-depl:
	kubectl apply -f ./k8s/api-deployment.yaml
	kubectl apply -f ./k8s/worker-deployment.yaml
	kubectl apply -f ./k8s/postgres.yaml
	kubectl apply -f ./k8s/rabbitmq.yaml
	kubectl apply -f ./k8s/config-secrets.yaml
	kubectl apply -f ./k8s/hpa.yaml

# Testing
## test: run all tests
test:
	@echo "Testing..."
	@go test ./... -v 

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

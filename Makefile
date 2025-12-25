.PHONY: run build test clean migrate seed dev

# Variables
APP_NAME=investify-api
BUILD_DIR=bin
MAIN_FILE=cmd/api/main.go

# Run the application
run:
	go run $(MAIN_FILE)

# Build the application
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy

# Install development tools
dev-tools:
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Run with hot reload (requires air)
dev:
	air

# Generate Swagger docs
swagger:
	swag init -g $(MAIN_FILE) -o docs/swagger

# Copy .env.example to .env
env:
	cp .env.example .env

# Docker commands
docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker run -p 8080:8080 --env-file .env $(APP_NAME)

# Database commands (if using golang-migrate)
migrate-up:
	migrate -path migrations -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path migrations -database "$${DATABASE_URL}" down

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Seed the database with test data
seed:
	go run cmd/seed/main.go

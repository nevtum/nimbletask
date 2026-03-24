.PHONY: all test build clean

# Default target
all: test

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

# Build the application
build:
	@echo "Building todo_cli..."
	go build -o bin/todo ./cmd/*.go

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Run tests in short mode
test-short:
	go test ./... -short

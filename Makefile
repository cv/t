.PHONY: all build test test-cover test-race lint clean install generate

# Default target
all: lint test build

# Generate code (downloads fresh IATA data)
generate:
	go generate ./...

# Build the binary
build:
	go build -o t ./cmd/t

# Install the binary
install:
	go install ./cmd/t

# Run tests
test:
	go test ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run tests with race detector
test-race:
	go test -race ./...

# Run linters
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run
	go vet ./...

# Format code
fmt:
	gofmt -w .
	@which goimports > /dev/null && goimports -w . || true

# Clean build artifacts
clean:
	rm -f t coverage.out coverage.html

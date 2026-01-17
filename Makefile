.PHONY: build test clean install dev lint

# Variables
BINARY_NAME=proj
VERSION?=dev
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
GO?=$(shell which go || echo /home/cjennings/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.12.linux-amd64/bin/go)
GOFLAGS=-trimpath

# Build for current platform
build:
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/proj

# Build for all supported platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/proj
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/proj
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/proj
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/proj

# Run tests
test:
	$(GO) test -v -race -cover ./...

# Run tests with coverage
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html

# Install to $GOPATH/bin
install:
	$(GO) install $(GOFLAGS) $(LDFLAGS) ./cmd/proj

# Run in development mode (use: make dev ARGS="--list" or make dev ARGS="--set-path ~/code")
dev:
	$(GO) run ./cmd/proj $(ARGS)

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	$(GO) fmt ./...
	gofmt -s -w .

# Tidy dependencies
tidy:
	$(GO) mod tidy

# Download dependencies
deps:
	$(GO) mod download

.PHONY: all build build-all clean install help

# Binary name
BINARY_NAME=worldclock

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Version and build info (optional)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default target
all: build

# Build for current platform
build:
	@echo "Building for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v

# Build for all common platforms
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

# Build for Linux AMD64
build-linux-amd64:
	@echo "Building for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v

# Build for Linux ARM64
build-linux-arm64:
	@echo "Building for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 -v

# Build for macOS AMD64 (Intel)
build-darwin-amd64:
	@echo "Building for macOS AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v

# Build for macOS ARM64 (Apple Silicon)
build-darwin-arm64:
	@echo "Building for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -v

# Build for Windows AMD64
build-windows-amd64:
	@echo "Building for Windows AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -v

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Install to GOPATH/bin
install:
	@echo "Installing to GOPATH/bin..."
	$(GOCMD) install

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run the application
run:
	@echo "Running application..."
	$(GOCMD) run .

# Show help
help:
	@echo "Available targets:"
	@echo "  make build              - Build for current platform (output: bin/worldclock)"
	@echo "  make build-all          - Build for all common platforms"
	@echo "  make build-linux-amd64  - Build for Linux AMD64"
	@echo "  make build-linux-arm64  - Build for Linux ARM64"
	@echo "  make build-darwin-amd64 - Build for macOS Intel"
	@echo "  make build-darwin-arm64 - Build for macOS Apple Silicon"
	@echo "  make build-windows-amd64- Build for Windows AMD64"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make install            - Install to GOPATH/bin"
	@echo "  make test               - Run tests"
	@echo "  make run                - Run without building"
	@echo "  make help               - Show this help message"

# Get the version from the latest git tag
VERSION := $(shell git describe --tags --abbrev=0)

# Set the output binary name
BINARY_NAME := llm-data-analyzer
DEV_BINARY_PATH := ./bin/$(BINARY_NAME)

# Default target
all: build

# Build the development binary
build:
	@echo "Building for development..."
	@mkdir -p ./bin
	@go build -o $(DEV_BINARY_PATH) .

# Run the unit tests
test:
	@echo "Running unit tests..."
	@go test -v ./...

# Test the version command
test-version:
	@echo "Testing version command..."
	@go run -ldflags="-X 'main.version=$(VERSION)'" . version

# Build the release packages
release: test-version
	@echo "Building release packages for version $(VERSION)..."
	@rm -rf ./release
	
	# Create release directories
	@mkdir -p ./release/linux_amd64
	@mkdir -p ./release/windows_amd64
	@mkdir -p ./release/darwin_amd64
	@mkdir -p ./release/darwin_arm64

	# Build binaries
	@GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.version=$(VERSION)'" -o ./release/linux_amd64/$(BINARY_NAME) .
	@GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.version=$(VERSION)'" -o ./release/windows_amd64/$(BINARY_NAME).exe .
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.version=$(VERSION)'" -o ./release/darwin_amd64/$(BINARY_NAME) .
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'main.version=$(VERSION)'" -o ./release/darwin_arm64/$(BINARY_NAME) .

	# Copy README and LICENSE
	@cp README.md LICENSE ./release/linux_amd64/
	@cp README.md LICENSE ./release/windows_amd64/
	@cp README.md LICENSE ./release/darwin_amd64/
	@cp README.md LICENSE ./release/darwin_arm64/

	@echo "Creating release archives..."
	@cd ./release/linux_amd64 && tar -czvf ../llm-data-analyzer_$(VERSION)_linux_amd64.tar.gz .
	@cd ./release/windows_amd64 && zip -r ../llm-data-analyzer_$(VERSION)_windows_amd64.zip .
	@cd ./release/darwin_amd64 && tar -czvf ../llm-data-analyzer_$(VERSION)_darwin_amd64.tar.gz .
	@cd ./release/darwin_arm64 && tar -czvf ../llm-data-analyzer_$(VERSION)_darwin_arm64.tar.gz .

	@echo "Release packages are in ./release"

# Clean the build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf ./bin ./release

.PHONY: all build release clean test test-version
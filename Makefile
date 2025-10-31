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

# Build the release packages
release:
	@echo "Building release packages for version $(VERSION)..."
	@rm -rf ./release
	@GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.version=$(VERSION)'" -o ./release/linux_amd64/$(BINARY_NAME) .
	@GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.version=$(VERSION)'" -o ./release/windows_amd64/$(BINARY_NAME).exe .
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.version=$(VERSION)'" -o ./release/darwin_amd64/$(BINARY_NAME) .
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'main.version=$(VERSION)'" -o ./release/darwin_arm64/$(BINARY_NAME) .
	@echo "Release packages are in ./release"

# Clean the build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf ./bin ./release

.PHONY: all build release clean

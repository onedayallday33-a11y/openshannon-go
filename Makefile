# Build configurations
BINARY_NAME=openshannon
BUILD_DIR=./cmd/openshannon
LDFLAGS=-s -w

.PHONY: all build clean help

all: build

## build: Standard build (optimized size)
build:
	@echo "Building optimized binary: $(BINARY_NAME)..."
	go build -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) $(BUILD_DIR)
	@echo "Done!"

## build-debug: Build with debug symbols
build-debug:
	@echo "Building debug binary..."
	go build -o $(BINARY_NAME)_debug $(BUILD_DIR)

## clean: Remove binaries
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)_debug $(BINARY_NAME).exe

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //'

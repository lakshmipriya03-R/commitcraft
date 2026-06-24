BINARY   := commitcraft
BUILD_DIR := ./bin
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

.PHONY: build install clean test lint tidy

## build: compile the binary into ./bin/
build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .
	@echo "built: $(BUILD_DIR)/$(BINARY)"

## install: build and copy to /usr/local/bin
install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "installed: /usr/local/bin/$(BINARY)"

## tidy: download deps and tidy go.mod / go.sum
tidy:
	go mod tidy

## test: run all tests
test:
	go test ./... -v -count=1

## lint: run golangci-lint (must be installed separately)
lint:
	golangci-lint run ./...

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'

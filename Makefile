.PHONY: all build clean test lint run help

APP_NAME    := lhd
BUILD_DIR   := ./build
GO_FLAGS    := -ldflags="-s -w -X github.com/linuxhealthdoctor/lhd/pkg/version.Version=$(VERSION) -X github.com/linuxhealthdoctor/lhd/pkg/version.Commit=$(COMMIT) -X github.com/linuxhealthdoctor/lhd/pkg/version.Date=$(DATE)"
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE        ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

all: clean test build

build:
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(GO_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd/lhd

build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(GO_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/lhd
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(GO_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/lhd
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build $(GO_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-386 ./cmd/lhd

test:
	go test ./... -v -count=1

test-short:
	go test ./... -short -count=1

test-race:
	go test ./... -race -count=1

lint:
	go vet ./...
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

fmt:
	go fmt ./...

clean:
	rm -rf $(BUILD_DIR)

run: build
	./$(BUILD_DIR)/$(APP_NAME)

install:
	go install ./cmd/lhd

tidy:
	go mod tidy
	go mod verify

coverage:
	mkdir -p $(BUILD_DIR)
	go test ./... -coverprofile=$(BUILD_DIR)/coverage.out -count=1
	go tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

help:
	@echo "Usage:"
	@echo "  make build      Build the binary"
	@echo "  make test       Run all tests"
	@echo "  make lint       Run linters"
	@echo "  make clean      Clean build artifacts"
	@echo "  make run        Build and run"
	@echo "  make install    Install to GOPATH"
	@echo "  make coverage   Generate coverage report"

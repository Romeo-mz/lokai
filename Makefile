BINARY  := lokai
MODULE  := github.com/romeo-mz/lokai
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build build-all clean test lint run fmt tidy docker docker-run benchmark coverage ci

## build: Build for current platform
build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/lokai

## run: Build and run
run: build
	./bin/$(BINARY)

## build-all: Cross-compile for all platforms
build-all:
	@mkdir -p bin
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-amd64   ./cmd/lokai
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-arm64   ./cmd/lokai
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-amd64  ./cmd/lokai
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-arm64  ./cmd/lokai
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-windows-amd64.exe ./cmd/lokai
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-windows-arm64.exe ./cmd/lokai

## test: Run tests
test:
	go test -v -race ./...

## lint: Run linters
lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

## fmt: Format code
fmt:
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

## tidy: Tidy dependencies
tidy:
	go mod tidy

## clean: Remove build artifacts
clean:
	rm -rf bin/

## docker: Build Docker image
docker:
	docker build -t lokai:latest --build-arg VERSION=$(VERSION) .

## docker-run: Run lokai in Docker (connects to host Ollama)
docker-run:
	docker run --rm -it -e OLLAMA_HOST=http://host.docker.internal:11434 lokai:latest

## benchmark: Build and run benchmark mode
benchmark: build
	./bin/$(BINARY) --benchmark

## coverage: Run tests with coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## ci: Run all CI checks locally
ci: fmt lint test build

## install: Install lokai to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" ./cmd/lokai

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'

# golangci-lint version — update here when upgrading
GOLANGCI_LINT_VERSION := v2.11.1

BINARY_NAME = t-f
BIN_DIR = .bin
MAIN_PATH = ./cmd/t-f
COMMIT = $(shell git rev-parse --short HEAD)
DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION ?= dev

.PHONY: lint test build check install-hooks

lint:
	golangci-lint run ./...
test:
	go test ./...

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build \
		-ldflags "-s -w -X t-f/internal/version.Version=$(VERSION) -X t-f/internal/version.Commit=$(COMMIT) -X t-f/internal/version.Date=$(DATE)" \
		-o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary created at $(BIN_DIR)/$(BINARY_NAME)"

check: lint test

install-hooks:
	@echo '#!/usr/bin/env bash' > .git/hooks/pre-commit
	@echo 'set -e' >> .git/hooks/pre-commit
	@echo 'make lint' >> .git/hooks/pre-commit
	@echo 'make test' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hook installed."

# GGU (Go Generic Utils) Makefile
# Provides commands for testing and maintaining the project

SHELL := /bin/bash
GO := go
MODULE := github.com/noobtrump/go-generic-utils
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.1")

# Directories
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
SOURCES := $(shell find . -name "*.go" -type f)

.PHONY: all clean test test-coverage lint fmt vet examples update-copyright help update-deps bench

all: lint test ## Run default targets (lint, test)

clean: ## Clean up generated files
	rm -f coverage.out
	rm -rf dist/

test: ## Run unit tests
	$(GO) test -race -cover ./...

test-coverage: ## Run tests with coverage
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

lint: ## Run linters
	golangci-lint run

fmt: ## Format source code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

examples: ## Run example applications
	$(GO) run ./example/basic
	$(GO) run ./example/advanced

update-copyright: ## Update copyright headers
	./update_copyright.sh

bench: ## Run benchmarks
	$(GO) test -benchmem -bench=. ./dataStructures/...
	$(GO) test -benchmem -bench=. ./tree/...
	$(GO) test -benchmem -bench=. ./sliceutils/...

update-deps: ## Update dependencies
	$(GO) get -u ./...
	$(GO) mod tidy

doc: ## Generate and serve Go documentation
	$(GO) install golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060

help: ## Show help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' 
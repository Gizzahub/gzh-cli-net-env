# Makefile - gzh-cli-net-env Library
# Network Environment Management Library

projectname := gzh-cli-net-env
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0-alpha")

export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

CYAN  := \033[36m
GREEN := \033[32m
RESET := \033[0m

.PHONY: help build test fmt lint check clean tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'

build: ## Build the library
	go build ./...

test: ## Run tests
	go test ./...

fmt: ## Format code
	gofmt -w .

lint: ## Run linter
	golangci-lint run ./... 2>/dev/null || go vet ./...

tidy: ## Tidy go modules
	go mod tidy

check: fmt lint test ## Run all checks

clean: ## Clean build artifacts
	go clean ./...

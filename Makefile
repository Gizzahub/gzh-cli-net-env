# Makefile - gzh-cli-net-env Library
# Network Environment Management Library

projectname := gzh-cli-net-env
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0-alpha")

export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

CYAN  := \033[36m
GREEN := \033[32m
RESET := \033[0m

.PHONY: help build test fmt fmt-check lint check clean tidy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-20s$(RESET) %s\n", $$1, $$2}'

build: ## Build the library
	go build ./...

test: ## Run tests
	go test ./...

fmt: ## Format code
	gofmt -w .

fmt-check: ## Check formatting without writing (fails with the files fmt would rewrite)
	@files="$$(gofmt -l .)" || exit 1; \
	if [ -n "$$files" ]; then echo "Code is not formatted. Run: make fmt"; printf '%s\n' "$$files"; exit 1; fi

lint: ## Run linter
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Run: make install-tools" >&2; exit 1; }
	golangci-lint run ./...

tidy: ## Tidy go modules
	go mod tidy

check: fmt lint test ## Run all checks

clean: ## Clean build artifacts
	go clean ./...

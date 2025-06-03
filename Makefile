# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BINARY_NAME=parity-client
MAIN_PATH=cmd/main.go
AIR_VERSION=v1.49.0
GOPATH=$(shell go env GOPATH)
AIR=$(GOPATH)/bin/air

# Formatting and linting tools
GOFUMPT=$(GOPATH)/bin/gofumpt
GOIMPORTS=$(GOPATH)/bin/goimports
GOLANGCI_LINT=$(shell which golangci-lint)

# Test related variables
COVERAGE_DIR=coverage
COVERAGE_PROFILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html
TEST_FLAGS=-race -coverprofile=$(COVERAGE_PROFILE) -covermode=atomic
TEST_PACKAGES=./...  # This will test all packages
TEST_PATH=./test/...

# Linting and formatting parameters
LINT_FLAGS=--timeout=5m
LINT_CONFIG=.golangci.yml
LINT_OUTPUT_FORMAT=colored-line-number
LINT_VERBOSE=false

# Build flags
BUILD_FLAGS=-v

# Installation path
INSTALL_PATH=/usr/local/bin

# Parity config directory
PARITY_CONFIG_DIR=$(HOME)/.parity

.PHONY: all build test run clean deps fmt help docker-up docker-down docker-logs docker-build docker-clean install-air watch install uninstall install-lint-tools lint install-hooks format-lint check-format

all: clean build

deps:
	git submodule update --init --recursive
	go mod tidy
	go mod download

build: ## Build the application
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd
	chmod +x $(BINARY_NAME)

test: setup-coverage ## Run tests with coverage
	$(GOTEST) $(TEST_FLAGS) -v $(TEST_PACKAGES)
	@go tool cover -func=$(COVERAGE_PROFILE)
	@go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)

setup-coverage: ## Create coverage directory
	@mkdir -p $(COVERAGE_DIR)

run:  ## Run the application
	$(GOCMD) run $(MAIN_PATH) 

stake:  ## Stake tokens in the network
	$(GOCMD) run $(MAIN_PATH) stake --amount 10

balance:  ## Check token balances
	$(GOCMD) run $(MAIN_PATH) balance

auth:  ## Authenticate with the network
	$(GOCMD) run $(MAIN_PATH) auth

clean: ## Clean build files
	rm -f $(BINARY_NAME)
	find . -type f -name '*.test' -delete
	find . -type f -name '*.out' -delete
	rm -rf tmp/

fmt: ## Format code using gofumpt (preferred) or gofmt
	@echo "Formatting code..."
	@if command -v $(GOFUMPT) >/dev/null 2>&1; then \
		echo "Using gofumpt for better formatting..."; \
		$(GOFUMPT) -l -w .; \
	else \
		echo "gofumpt not found, using standard gofmt..."; \
		$(GOFMT) ./...; \
		echo "Consider installing gofumpt: go install mvdan.cc/gofumpt@latest"; \
	fi

imports: ## Fix imports formatting and add missing imports
	@echo "Organizing imports..."
	@if command -v $(GOIMPORTS) >/dev/null 2>&1; then \
		$(GOIMPORTS) -w -local github.com/theblitlabs/parity-runner .; \
	else \
		echo "goimports not found. Installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		$(GOIMPORTS) -w -local github.com/theblitlabs/parity-runner .; \
	fi

format: fmt imports ## Run all formatters (gofumpt + goimports)
	@echo "All formatting completed."

lint: ## Run linting with options (make lint VERBOSE=true CONFIG=custom.yml OUTPUT=json)
	@echo "Running linters..."
	$(eval FINAL_LINT_FLAGS := $(LINT_FLAGS))
	@if [ "$(VERBOSE)" = "true" ]; then \
		FINAL_LINT_FLAGS="$(FINAL_LINT_FLAGS) -v"; \
	fi
	@if [ -n "$(CONFIG)" ]; then \
		FINAL_LINT_FLAGS="$(FINAL_LINT_FLAGS) --config=$(CONFIG)"; \
	else \
		FINAL_LINT_FLAGS="$(FINAL_LINT_FLAGS) --config=$(LINT_CONFIG)"; \
	fi
	@if [ -n "$(OUTPUT)" ]; then \
		FINAL_LINT_FLAGS="$(FINAL_LINT_FLAGS) --out-format=$(OUTPUT)"; \
	else \
		FINAL_LINT_FLAGS="$(FINAL_LINT_FLAGS) --out-format=$(LINT_OUTPUT_FORMAT)"; \
	fi
	$(GOLANGCI_LINT) run $(FINAL_LINT_FLAGS)

format-lint: format lint ## Format code and run linters in one step

check-format: ## Check code formatting without applying changes (useful for CI)
	@echo "Checking code formatting..."
	@./scripts/check_format.sh

install-air: ## Install air for hot reloading
	@if ! command -v air > /dev/null; then \
		echo "Installing air..." && \
		go install github.com/air-verse/air@latest; \
	fi

watch: install-air ## Run the application with hot reload
	$(AIR)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install: build ## Install parity command globally
	@echo "Installing parity to $(INSTALL_PATH)..."
	@sudo mv $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Setting up parity config directory..."
	@mkdir -p $(PARITY_CONFIG_DIR)
	@if [ -f .env ]; then \
		if [ -f $(PARITY_CONFIG_DIR)/.env ]; then \
			echo "Config file already exists at $(PARITY_CONFIG_DIR)/.env"; \
			read -p "Do you want to replace it? (Y/n): " -n 1 -r; \
			echo; \
			if [[ $$REPLY =~ ^[Yy]$$ ]] || [[ -z $$REPLY ]]; then \
				cp .env $(PARITY_CONFIG_DIR)/.env; \
				echo "Config file replaced at $(PARITY_CONFIG_DIR)/.env"; \
			else \
				echo "Keeping existing config file at $(PARITY_CONFIG_DIR)/.env"; \
			fi; \
		else \
			cp .env $(PARITY_CONFIG_DIR)/.env; \
			echo "Config file copied to $(PARITY_CONFIG_DIR)/.env"; \
		fi; \
	else \
		echo "No .env file found to copy"; \
	fi
	@echo "Installation complete"

uninstall: ## Remove parity command from system
	@echo "Uninstalling parity from $(INSTALL_PATH)..."
	@sudo rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Removing parity config directory..."
	@rm -rf $(PARITY_CONFIG_DIR)
	@echo "Uninstallation complete"

install-lint-tools: ## Install formatting and linting tools
	@echo "Installing linting and formatting tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installation complete."

install-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@./scripts/hooks/install-hooks.sh

.DEFAULT_GOAL := help

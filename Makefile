# Makefile for ScribesAI
# Cross-platform build system for Linux and macOS

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    OS = linux
    DETECTED_OS = Linux
endif
ifeq ($(UNAME_S),Darwin)
    OS = darwin
    DETECTED_OS = macOS
endif

# Build variables
BINARY_NAME = scribescli
CMD_PATH = ./cmd/scribescli
BUILD_DIR = build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Colors for output
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[1;33m
CYAN = \033[0;36m
NC = \033[0m # No Color

.PHONY: all build clean test deps install run help check-os check-deps dev release

all: check-deps build

help: ## Show this help message
	@echo "$(CYAN)ScribesAI - Meeting Intelligence System$(NC)"
	@echo "$(YELLOW)Detected OS: $(DETECTED_OS)$(NC)"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-15s$(NC) %s\n", $$1, $$2}'

check-os: ## Check operating system compatibility
	@echo "$(CYAN)Checking OS compatibility...$(NC)"
ifeq ($(OS),)
	@echo "$(RED)✗ Unsupported OS. Only Linux and macOS are supported.$(NC)"
	@exit 1
else
	@echo "$(GREEN)✓ Detected: $(DETECTED_OS)$(NC)"
endif

check-deps: check-os ## Check if all dependencies are installed
	@echo "$(CYAN)Checking dependencies...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)✗ Go is not installed$(NC)"; exit 1; }
	@echo "$(GREEN)✓ Go: $$(go version)$(NC)"
	@pkg-config --exists portaudio-2.0 || { echo "$(YELLOW)! PortAudio not found. Run 'make install-portaudio'$(NC)"; exit 1; }
	@echo "$(GREEN)✓ PortAudio installed$(NC)"

install-portaudio: ## Install PortAudio (requires sudo)
	@echo "$(CYAN)Installing PortAudio for $(DETECTED_OS)...$(NC)"
ifeq ($(OS),linux)
	@if command -v apt-get >/dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y portaudio19-dev; \
	elif command -v dnf >/dev/null 2>&1; then \
		sudo dnf install -y portaudio-devel; \
	elif command -v pacman >/dev/null 2>&1; then \
		sudo pacman -S portaudio; \
	else \
		echo "$(RED)✗ Unsupported package manager$(NC)"; \
		exit 1; \
	fi
endif
ifeq ($(OS),darwin)
	@command -v brew >/dev/null 2>&1 || { echo "$(RED)✗ Homebrew required. Install from https://brew.sh/$(NC)"; exit 1; }
	@brew install portaudio
endif
	@echo "$(GREEN)✓ PortAudio installed$(NC)"

deps: ## Download Go dependencies
	@echo "$(CYAN)Downloading Go dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

build: deps ## Build the application for current platform
	@echo "$(CYAN)Building ScribesAI for $(DETECTED_OS)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "$(GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-linux: ## Build for Linux (from any platform)
	@echo "$(CYAN)Cross-compiling for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "$(GREEN)✓ Linux build complete$(NC)"

build-mac: ## Build for macOS (from any platform)
	@echo "$(CYAN)Cross-compiling for macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "$(GREEN)✓ macOS builds complete (Intel + Apple Silicon)$(NC)"

build-all: ## Build for all supported platforms
	@echo "$(CYAN)Building for all platforms...$(NC)"
	@$(MAKE) build-linux
	@$(MAKE) build-mac
	@echo "$(GREEN)✓ All builds complete$(NC)"
	@ls -lh $(BUILD_DIR)

install: build ## Install binary to /usr/local/bin (requires sudo)
	@echo "$(CYAN)Installing $(BINARY_NAME)...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Installed to /usr/local/bin/$(BINARY_NAME)$(NC)"

uninstall: ## Uninstall binary from /usr/local/bin (requires sudo)
	@echo "$(CYAN)Uninstalling $(BINARY_NAME)...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✓ Uninstalled$(NC)"

clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "$(GREEN)✓ Clean complete$(NC)"

test: ## Run tests
	@echo "$(CYAN)Running tests...$(NC)"
	@$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage report
	@echo "$(CYAN)Running tests with coverage...$(NC)"
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

run: build ## Build and run the application
	@echo "$(CYAN)Running ScribesAI...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode (no build)
	@echo "$(CYAN)Running in dev mode...$(NC)"
	@$(GOCMD) run $(CMD_PATH)

setup: ## Complete setup (install deps, create .env, build)
	@echo "$(CYAN)Running complete setup...$(NC)"
	@$(MAKE) check-deps || $(MAKE) install-portaudio
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(YELLOW)! Created .env - please add your ANTHROPIC_API_KEY$(NC)"; \
	fi
	@mkdir -p data models data/exports
	@touch models/.gitkeep
	@$(MAKE) deps
	@$(MAKE) build
	@echo "$(GREEN)✓ Setup complete!$(NC)"

release: clean test build-all ## Build release binaries for all platforms
	@echo "$(CYAN)Creating release artifacts...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	@cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	@cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	@cd $(BUILD_DIR)/release && sha256sum *.tar.gz > checksums.txt
	@echo "$(GREEN)✓ Release artifacts created in $(BUILD_DIR)/release/$(NC)"

format: ## Format Go code
	@echo "$(CYAN)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

lint: ## Run linter
	@echo "$(CYAN)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)! golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

doctor: ## Check system health and configuration
	@echo "$(CYAN)System Health Check$(NC)"
	@echo ""
	@echo "Operating System:"
	@echo "  $(GREEN)✓$(NC) Detected: $(DETECTED_OS) ($(OS))"
	@echo ""
	@echo "Dependencies:"
	@command -v go >/dev/null 2>&1 && echo "  $(GREEN)✓$(NC) Go: $$(go version)" || echo "  $(RED)✗$(NC) Go not installed"
	@pkg-config --exists portaudio-2.0 && echo "  $(GREEN)✓$(NC) PortAudio: $$(pkg-config --modversion portaudio-2.0)" || echo "  $(RED)✗$(NC) PortAudio not found"
	@echo ""
	@echo "Configuration:"
	@if [ -f .env ]; then \
		echo "  $(GREEN)✓$(NC) .env file exists"; \
		if grep -q "ANTHROPIC_API_KEY=sk-ant-" .env; then \
			echo "  $(GREEN)✓$(NC) ANTHROPIC_API_KEY configured"; \
		else \
			echo "  $(YELLOW)!$(NC) ANTHROPIC_API_KEY not set in .env"; \
		fi; \
	else \
		echo "  $(RED)✗$(NC) .env file missing"; \
	fi
	@echo ""
	@echo "Directories:"
	@[ -d data ] && echo "  $(GREEN)✓$(NC) data/" || echo "  $(YELLOW)!$(NC) data/ missing"
	@[ -d models ] && echo "  $(GREEN)✓$(NC) models/" || echo "  $(YELLOW)!$(NC) models/ missing"
	@echo ""

.DEFAULT_GOAL := help

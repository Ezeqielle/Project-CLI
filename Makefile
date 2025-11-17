# Default settings
SHELL := /bin/bash
GO ?= go
GOBIN ?= /usr/local/go/bin
BUILD_DIR = build
APP_NAME ?= pcli

# Force these targets to always run
.PHONY: all help clean build confirm

# Default
all: help

# ------------------------------------------------------------------------------------ #
# HELPERS
# ------------------------------------------------------------------------------------ #
confirm:
	@bash -c 'read -p "Are you sure? [y/N] " ans; if [ "$$ans" != "y" ]; then echo "Cancelled."; exit 1; fi'

# ------------------------------------------------------------------------------------ #
# DEPENDENCIES
# ------------------------------------------------------------------------------------ #
deps-install:
	@echo "Installing prerequisites..."
	@if command -v apt >/dev/null 2>&1; then \
		sudo apt update && sudo apt install -y make golang-go; \
	elif command -v dnf >/dev/null 2>&1; then \
		sudo dnf install -y make golang; \
	elif command -v pacman >/dev/null 2>&1; then \
		sudo pacman -S --noconfirm make go; \
	else \
		echo "⚠️ Please install dependencies manually for your distribution."; \
	fi
	@echo "✅ Dependencies installed."

deps-check:
	@echo "Checking dependencies..."
	@for dep in go make; do \
		if ! command -v $$dep >/dev/null 2>&1; then \
			echo "❌ Missing dependency: $$dep"; \
		else \
			echo "✅ Found: $$dep"; \
		fi; \
	done

# ------------------------------------------------------------------------------------ #
# BUILD / INSTALL
# ------------------------------------------------------------------------------------ #
build:
ifndef APP_NAME
	$(error APP_NAME is required. Usage: make build APP_NAME=yourbinary)
endif
	@echo "🔧 Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)/main.go
	@echo "✅ Build complete. Output: $(BUILD_DIR)/$(APP_NAME)"

clean: confirm
ifndef APP_NAME
	$(error APP_NAME is required. Usage: make clean APP_NAME=yourbinary)
endif
	@echo "Cleaning build directory..."
	@rm $(BUILD_DIR)/$(APP_NAME) || true

install: confirm
ifndef APP_NAME
	$(error APP_NAME is required. Usage: sudo make install APP_NAME=yourbinary)
endif
	@if [[ $$EUID -ne 0 ]]; then echo "❌ Run with sudo."; exit 1; fi
	@echo "Installing $(APP_NAME) to $(GOBIN)..."
	@mkdir -p $(GOBIN)
	@chmod 755 $(BUILD_DIR)/$(APP_NAME)
	@cp $(BUILD_DIR)/$(APP_NAME) $(GOBIN)/$(APP_NAME)
	@ls -l $(GOBIN)/$(APP_NAME)
	@echo "✅ Installed: $(GOBIN)/$(APP_NAME)"

install-clean: confirm
ifndef APP_NAME
	$(error APP_NAME is required. Usage: sudo make install-clean APP_NAME=yourbinary)
endif
	@if [[ $$EUID -ne 0 ]]; then echo "❌ Run with sudo."; exit 1; fi
	@if [ -f $(GOBIN)/$(APP_NAME) ]; then rm -f $(GOBIN)/$(APP_NAME); fi
	@echo "✅ Uninstalled."

# ------------------------------------------------------------------------------------ #
# DEBUG
# ------------------------------------------------------------------------------------ #
debug-PATH:
	@echo "Debugging info:"
	@echo "  GOBIN: $(GOBIN)"
	@echo "  BUILD_DIR: $(BUILD_DIR)"
	@echo "  GO: $(GO)"

# ------------------------------------------------------------------------------------ #
# HELP
# ------------------------------------------------------------------------------------ #
help:
	@echo ""
	@echo "Available commands:"
	@echo ""
	@echo "  BUILD / CODEGEN:"
	@echo "    make build APP_NAME=yourbinary         - Build binary (mandatory APP_NAME)"
	@echo "    make clean APP_NAME=yourbinary         - Clean built binary"
	@echo ""
	@echo "  DEPENDENCIES:"
	@echo "    make deps-install                      - Install required deps (system + Go tools)"
	@echo "    make deps-check                        - Check installed deps"
	@echo ""
	@echo "  SUDO ONLY:"
	@echo "    sudo make install APP_NAME=yourbinary  - Copy built binary to \$$GOBIN"
	@echo "    sudo make install-clean APP_NAME=...   - Remove binary from \$$GOBIN"
	@echo ""
	@echo "  DEBUG:"
	@echo "    make debug-PATH                        - Print important paths and vars"
	@echo ""


# Project variables
APP_NAME    := starter
MIGRATE_CMD := ./cmd/migrate/main.go
BIN_DIR     := bin
MAIN_FILE   := ./cmd/main.go
PKG         := ./...

# Go parameters
GO ?= go
LINTER := golangci-lint

.PHONY: all run build clean lint test tidy deps help \
        migrate-up migrate-down migrate-version migrate-steps

all: build

## Generate templ components
templ-gen:
	@echo ">> Generating templ components..."
	@templ generate ./internal/views/...

## Run the application
run: templ-gen
	@echo ">> Running $(APP_NAME)..."
	@$(GO) run $(MAIN_FILE)

## Build the binary
build:
	@echo ">> Building binary..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BIN_DIR)/$(APP_NAME)"

## Clean build artifacts
clean:
	@echo ">> Cleaning..."
	@rm -rf $(BIN_DIR)
	@$(GO) clean
	@echo "🧹 Clean complete"

## Lint the codebase
lint:
	@echo ">> Running linter..."
	@$(LINTER) run $(PKG)

## Run unit tests with coverage
test:
	@echo ">> Running tests..."
	@$(GO) test -v -cover $(PKG)

## Format and tidy modules
tidy:
	@echo ">> Tidying up..."
	@$(GO) fmt $(PKG)
	@$(GO) mod tidy

## Install dependencies
deps:
	@echo ">> Installing dependencies..."
	@$(GO) mod download

## Run all pending database migrations
migrate-up:
	@echo ">> Migrating up..."
	@$(GO) run $(MIGRATE_CMD) -cmd up

## Roll back all applied database migrations
migrate-down:
	@echo ">> Migrating down..."
	@$(GO) run $(MIGRATE_CMD) -cmd down

## Show current migration version
migrate-version:
	@$(GO) run $(MIGRATE_CMD) -cmd version

## Migrate N steps: make migrate-steps N=2 (negative to roll back)
migrate-steps:
	@echo ">> Migrating $(N) step(s)..."
	@$(GO) run $(MIGRATE_CMD) -cmd steps -steps $(N)

## Force migration version (clears dirty flag, does NOT touch schema): make migrate-force N=7
## Use only after manually inspecting/fixing the DB. Set N to last known-good version.
migrate-force:
	@echo ">> Forcing migration version to $(N)..."
	@$(GO) run $(MIGRATE_CMD) -cmd force -force $(N)

## Help menu
help:
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[32m%-12s\033[0m %s\n", $$1, $$2}'
	@echo ""

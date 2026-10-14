# ==============================================================================
# SMARTREVERSAL PLATFORM MAKEFILE
# ==============================================================================

.PHONY: help dev build test lint clean docker-up docker-down db-migrate

help: ## Display available commands
	@echo "SMARTREVERSAL Enterprise Financial Operations Platform"
	@echo "======================================================="
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ------------------------------------------------------------------------------
# Local Development & Orchestration
# ------------------------------------------------------------------------------
docker-up: ## Start core infrastructure and services with Docker Compose
	docker compose up -d

docker-down: ## Stop all Docker Compose containers
	docker compose down

docker-logs: ## Follow Docker Compose logs
	docker compose logs -f

dev-api: ## Run the Go backend API locally
	cd apps/api && go run cmd/api/main.go

dev-ai: ## Run the Python AI service locally
	cd apps/ai-service && uvicorn app.main:app --reload --port 8000

dev-web: ## Run the Vite web operations console locally
	npm run dev --workspace=apps/web

# ------------------------------------------------------------------------------
# Build & Compilation
# ------------------------------------------------------------------------------
build: build-api build-web ## Build all production assets

build-api: ## Compile Go API binary
	cd apps/api && go build -v -o bin/api cmd/api/main.go

build-web: ## Build Vite frontend production bundle
	npm run build --workspace=apps/web

# ------------------------------------------------------------------------------
# Quality & Verification
# ------------------------------------------------------------------------------
test: test-api test-ai test-web ## Run unit tests across all applications

test-api: ## Run Go tests
	cd apps/api && go test -v -race ./...

test-ai: ## Run Python AI service tests
	cd apps/ai-service && pytest

test-web: ## Run web tests
	npm run test --workspace=apps/web --if-present

lint: lint-api lint-web ## Run linters across projects

lint-api: ## Run golangci-lint if installed
	cd apps/api && go vet ./...

lint-web: ## Run ESLint for web and UI packages
	npm run lint --workspaces --if-present

clean: ## Clean build artifacts and temporary files
	rm -rf apps/api/bin apps/web/dist packages/ui/dist

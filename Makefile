.PHONY: help setup build up down logs clean test migrate migrate-down migrate-create db-seed proto lint format docker-clean

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo '$(BLUE)Library Management System - Development Commands$(NC)'
	@echo ''
	@echo 'Usage:'
	@echo '  $(GREEN)make$(NC) $(BLUE)<target>$(NC)'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

setup: ## Initial project setup
	@echo "$(BLUE)Setting up project...$(NC)"
	@command -v go >/dev/null 2>&1 || { echo "$(RED)Go is not installed$(NC)"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)Docker is not installed$(NC)"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "$(RED)Docker Compose is not installed$(NC)"; exit 1; }
	@echo "Installing Go dependencies..."
	@cd services/api-gateway && go mod download
	@cd services/document-service && go mod download
	@cd services/indexer-service && go mod download
	@echo "$(GREEN)Setup complete!$(NC)"

build: ## Build all services
	@echo "$(BLUE)Building services...$(NC)"
	docker-compose build
	@echo "$(GREEN)Build complete!$(NC)"

up: ## Start all services
	@echo "$(BLUE)Starting services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Services started!$(NC)"
	@echo ""
	@echo "$(BLUE)Service URLs:$(NC)"
	@echo "  API Gateway:    http://localhost:8085"
	@echo "  PostgreSQL:     localhost:5432"
	@echo "  Redis:          localhost:6380"
	@echo "  Elasticsearch:  http://localhost:9200"
	@echo "  Kibana:         http://localhost:5601"
	@echo "  MinIO Console:  http://localhost:9001"
	@echo "  Grafana:        http://localhost:3000"
	@echo "  Prometheus:     http://localhost:9090"
	@echo "  Jaeger:         http://localhost:16686"
	@echo ""
	@echo "$(GREEN)Run 'make logs' to view logs$(NC)"

down: ## Stop all services
	@echo "$(BLUE)Stopping services...$(NC)"
	docker-compose down
	@echo "$(GREEN)Services stopped!$(NC)"

restart: down up ## Restart all services

logs: ## View logs from all services
	docker-compose logs -f

logs-api: ## View API gateway logs
	docker-compose logs -f api-gateway

logs-doc: ## View document service logs
	docker-compose logs -f document-service

logs-indexer: ## View indexer service logs
	docker-compose logs -f indexer-service

ps: ## Show running containers
	docker-compose ps

clean: ## Stop services and remove volumes
	@echo "$(RED)Warning: This will delete all data!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose down -v; \
		echo "$(GREEN)Cleanup complete!$(NC)"; \
	fi

docker-clean: ## Remove all Docker images and rebuild
	@echo "$(BLUE)Cleaning Docker images...$(NC)"
	docker-compose down --rmi all
	@echo "$(GREEN)Docker cleanup complete!$(NC)"

test: ## Run all tests
	@echo "$(BLUE)Running tests...$(NC)"
	@cd services/api-gateway && go test -v -race -coverprofile=coverage.out ./...
	@cd services/document-service && go test -v -race -coverprofile=coverage.out ./...
	@cd services/indexer-service && go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests complete!$(NC)"

test-api: ## Run API gateway tests
	@cd services/api-gateway && go test -v -race -coverprofile=coverage.out ./...

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@cd services/api-gateway && go test -v -race -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: services/api-gateway/coverage.html$(NC)"

migrate: ## Run database migrations
	@echo "$(BLUE)Running migrations...$(NC)"
	docker-compose exec postgres psql -U libsystem -d libsystem -f /docker-entrypoint-initdb.d/001_initial_schema.up.sql
	@echo "$(GREEN)Migrations complete!$(NC)"

migrate-down: ## Rollback last migration
	@echo "$(BLUE)Rolling back migration...$(NC)"
	docker-compose exec postgres psql -U libsystem -d libsystem -f /docker-entrypoint-initdb.d/001_initial_schema.down.sql
	@echo "$(GREEN)Rollback complete!$(NC)"

migrate-create: ## Create a new migration (usage: make migrate-create NAME=add_users)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required$(NC)"; \
		echo "Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="$${timestamp}_$(NAME)"; \
	touch migrations/$${filename}.up.sql migrations/$${filename}.down.sql; \
	echo "$(GREEN)Created migrations/$${filename}.up.sql$(NC)"; \
	echo "$(GREEN)Created migrations/$${filename}.down.sql$(NC)"

db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U libsystem -d libsystem

db-seed: ## Seed database with test data
	@echo "$(BLUE)Seeding database...$(NC)"
	@docker-compose exec postgres psql -U libsystem -d libsystem -c "\
		INSERT INTO collections (name, description, slug, owner_id) VALUES \
		('Research Papers', 'Academic research papers collection', 'research-papers', (SELECT id FROM users WHERE username='admin')), \
		('Historical Documents', 'Historical documents and archives', 'historical-docs', (SELECT id FROM users WHERE username='admin')); \
	"
	@echo "$(GREEN)Database seeded!$(NC)"

redis-cli: ## Open Redis CLI
	docker-compose exec redis redis-cli -a libsystem_redis_pass

proto: ## Generate protobuf files
	@echo "$(BLUE)Generating protobuf files...$(NC)"
	@command -v protoc >/dev/null 2>&1 || { echo "$(RED)protoc is not installed$(NC)"; exit 1; }
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		shared/proto/*.proto
	@echo "$(GREEN)Protobuf generation complete!$(NC)"

lint: ## Run linters
	@echo "$(BLUE)Running linters...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "$(RED)golangci-lint is not installed$(NC)"; exit 1; }
	@cd services/api-gateway && golangci-lint run
	@cd services/document-service && golangci-lint run
	@cd services/indexer-service && golangci-lint run
	@echo "$(GREEN)Linting complete!$(NC)"

format: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	@find . -name "*.go" -not -path "./vendor/*" -exec gofmt -s -w {} \;
	@echo "$(GREEN)Formatting complete!$(NC)"

benchmark: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@cd services/api-gateway && go test -bench=. -benchmem ./...
	@echo "$(GREEN)Benchmarks complete!$(NC)"

health: ## Check service health
	@echo "$(BLUE)Checking service health...$(NC)"
	@curl -s http://localhost:8085/health | jq . || echo "$(RED)API Gateway is down$(NC)"
	@curl -s http://localhost:9200/_cluster/health | jq . || echo "$(RED)Elasticsearch is down$(NC)"
	@echo "$(GREEN)Health check complete!$(NC)"

init-minio: ## Initialize MinIO buckets
	@echo "$(BLUE)Initializing MinIO buckets...$(NC)"
	@docker-compose exec minio mc alias set local http://localhost:9000 minioadmin minioadmin123
	@docker-compose exec minio mc mb local/documents || true
	@docker-compose exec minio mc mb local/thumbnails || true
	@docker-compose exec minio mc policy set public local/thumbnails
	@echo "$(GREEN)MinIO buckets initialized!$(NC)"

init-kafka: ## Initialize Kafka topics
	@echo "$(BLUE)Initializing Kafka topics...$(NC)"
	@docker-compose exec kafka kafka-topics --create --if-not-exists --topic document.uploaded --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
	@docker-compose exec kafka kafka-topics --create --if-not-exists --topic document.indexed --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
	@docker-compose exec kafka kafka-topics --create --if-not-exists --topic document.deleted --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
	@echo "$(GREEN)Kafka topics initialized!$(NC)"

init: up init-minio init-kafka migrate db-seed ## Initialize entire development environment
	@echo "$(GREEN)Development environment ready!$(NC)"

backup-db: ## Backup database
	@timestamp=$$(date +%Y%m%d_%H%M%S); \
	docker-compose exec -T postgres pg_dump -U libsystem libsystem > backups/libsystem_$${timestamp}.sql; \
	echo "$(GREEN)Database backed up to backups/libsystem_$${timestamp}.sql$(NC)"

restore-db: ## Restore database (usage: make restore-db FILE=backup.sql)
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)Error: FILE is required$(NC)"; \
		echo "Usage: make restore-db FILE=backups/libsystem_20240101.sql"; \
		exit 1; \
	fi
	@docker-compose exec -T postgres psql -U libsystem libsystem < $(FILE)
	@echo "$(GREEN)Database restored from $(FILE)$(NC)"

dev-api: ## Run API gateway locally (outside Docker)
	@cd services/api-gateway && \
	export PORT=8085 \
	export ENVIRONMENT=development \
	export DATABASE_URL=postgres://libsystem:libsystem_dev_pass@localhost:5432/libsystem?sslmode=disable \
	export REDIS_URL=redis://:libsystem_redis_pass@localhost:6360/0 \
	export ELASTICSEARCH_URL=http://localhost:9200 \
	export MINIO_ENDPOINT=localhost:9000 \
	export MINIO_ACCESS_KEY=minioadmin \
	export MINIO_SECRET_KEY=minioadmin123 \
	export KAFKA_BROKERS=localhost:9092 \
	export JWT_SECRET=your-super-secret-jwt-key-change-in-production && \
	go run main.go

watch: ## Watch for changes and reload
	@command -v air >/dev/null 2>&1 || { echo "$(RED)air is not installed. Run: go install github.com/cosmtrek/air@latest$(NC)"; exit 1; }
	@cd services/api-gateway && air

doc: ## Generate API documentation
	@echo "$(BLUE)Generating API documentation...$(NC)"
	@command -v swag >/dev/null 2>&1 || { echo "$(RED)swag is not installed$(NC)"; exit 1; }
	@cd services/api-gateway && swag init
	@echo "$(GREEN)Documentation generated at services/api-gateway/docs$(NC)"

k8s-deploy: ## Deploy to Kubernetes
	@echo "$(BLUE)Deploying to Kubernetes...$(NC)"
	kubectl apply -f infrastructure/kubernetes/
	@echo "$(GREEN)Deployment complete!$(NC)"

.DEFAULT_GOAL := help
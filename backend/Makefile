ifneq (,$(wildcard .env.local))
include .env.local
export
endif

POSTGRES_DB ?= agenteur_local
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= postgres
DB_HOST ?= 127.0.0.1
DB_PORT ?= 5432
DATABASE_URL ?= postgres://$(POSTGRES_USER):postgres@$(DB_HOST):$(DB_PORT)/$(POSTGRES_DB)?sslmode=disable
DATABASE_URL := $(subst @localhost:,@127.0.0.1:,$(DATABASE_URL))
MIGRATIONS_DIR := ./migrations
GOOSE_CMD := go run github.com/pressly/goose/v3/cmd/goose@v3.26.0

.PHONY: db-up db-down db-reset db-logs db-wait db-create migrate-up migrate-down migrate-status migrate-create seed-superadmin

db-up:
	docker compose -f docker-compose.yml --env-file .env.local up -d postgres

db-down:
	docker compose -f docker-compose.yml --env-file .env.local down

db-reset:
	docker compose -f docker-compose.yml --env-file .env.local down -v
	docker compose -f docker-compose.yml --env-file .env.local up -d postgres

db-logs:
	docker compose -f docker-compose.yml --env-file .env.local logs -f postgres

db-wait:
	@until docker compose -f docker-compose.yml --env-file .env.local exec -T postgres pg_isready -U "$(POSTGRES_USER)" -d postgres >/dev/null 2>&1; do sleep 1; done

db-create:
	@PGPASSWORD="$(POSTGRES_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(POSTGRES_USER)" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$(POSTGRES_DB)'" | grep -q 1 || \
	PGPASSWORD="$(POSTGRES_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(POSTGRES_USER)" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"$(POSTGRES_DB)\";"
	@PGPASSWORD="$(POSTGRES_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(POSTGRES_USER)" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$(POSTGRES_DB)'" | grep -q 1

migrate-up: db-up db-wait db-create
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	@$(GOOSE_CMD) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	@$(GOOSE_CMD) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is required"; exit 1; fi
	@$(GOOSE_CMD) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-create:
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=create_users_table"; exit 1; fi
	@$(GOOSE_CMD) -dir $(MIGRATIONS_DIR) create "$(name)" sql

seed-superadmin:
	@if [ -z "$(email)" ]; then echo "Usage: make seed-superadmin email=user@example.com"; exit 1; fi
	@PGPASSWORD="$(POSTGRES_PASSWORD)" psql -h "$(DB_HOST)" -p "$(DB_PORT)" -U "$(POSTGRES_USER)" -d "$(POSTGRES_DB)" -c "UPDATE users SET is_superadmin = true WHERE LOWER(email) = LOWER('$(email)');"

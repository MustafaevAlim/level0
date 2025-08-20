include .env
export POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB DB_HOST

DATABASE_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(DB_HOST):5432/$(POSTGRES_DB)?sslmode=disable

MIGRATE=migrate -path ./migrations -database $(DATABASE_URL)

CONTAINER=level0-backend-1

.PHONY: migrate-up migrate-down migrate-version migrate-force migrate-up-docker migrate-down-docker migrate-version-docker

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

migrate-version:
	$(MIGRATE) version

migrate-force:
	$(MIGRATE) force $(ver)

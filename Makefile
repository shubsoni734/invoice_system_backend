APP_NAME=invoice-backend
CMD_PATH=./cmd/api
MIGRATIONS_DIR=./internal/pkg/db/migrations
DB_URL=$(shell grep DATABASE_URL .env.development | cut -d '=' -f2-)

.PHONY: run build migrate-up migrate-down migrate-status tidy sqlc

run:
	go run $(CMD_PATH)/main.go

build:
	go build -o bin/$(APP_NAME) $(CMD_PATH)/main.go

tidy:
	go mod tidy

sqlc:
	sqlc generate

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

.PHONY: help run build test docker-up docker-down migrate-up migrate-down migrate-status

DB_DSN=postgres://postgres:postgres@localhost:5432/org_structure?sslmode=disable

help:
	@echo "Доступные команды:"
	@echo "  make run            - запуск приложения локально"
	@echo "  make build          - сборка приложения"
	@echo "  make test           - запуск тестов"
	@echo "  make docker-up      - запуск через docker-compose"
	@echo "  make docker-down    - остановка docker-compose"
	@echo "  make migrate-up     - применить миграции"
	@echo "  make migrate-down   - откатить миграции"
	@echo "  make migrate-status - статус миграций"

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -v -cover

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

migrate-up:
	goose -dir migrations postgres "$(DB_DSN)" up

migrate-down:
	goose -dir migrations postgres "$(DB_DSN)" down

migrate-status:
	goose -dir migrations postgres "$(DB_DSN)" status
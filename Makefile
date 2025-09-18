.PHONY: run build tidy up down swag migrate logs

APP_NAME=subscription-service
CMD_DIR=./cmd/app

# Сборка бинарника
build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)

# Запуск локально
run:
	go run $(CMD_DIR)/main.go

# Чистим зависимости
tidy:
	go mod tidy

# Поднять контейнеры
up:
	docker compose up -d --build

# Остановить и удалить контейнеры + volume
down:
	docker compose down -v

# Генерация swagger-документации
swag:
	swag init -g $(CMD_DIR)/main.go -o docs

# Применить миграции (пример)
migrate:
	psql $$DATABASE_URL -f migrations/001_create_table.sql

# Логи приложения
logs:
	docker compose logs -f app
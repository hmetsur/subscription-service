# 📦 Subscription Service

Микросервис для управления пользовательскими подписками.  
Реализован на **Go 1.23**, хранение данных в **PostgreSQL**, запуск через **Docker Compose**, документация — **Swagger**.

---

## ✨ Возможности
- CRUDL-операции для подписок (создать, получить по ID, обновить, удалить, получить список)
- Подсчёт суммарной стоимости подписок за выбранный период
- Конфигурация через `.env` или `.yaml`
- Логирование через `slog`
- Автоматические миграции БД
- Swagger-документация (`/swagger/index.html`)
- Полностью контейнеризован через Docker Compose

---

## ⚙️ Запуск

### 🔹 Локально
1. Скопировать `.env.example` в `.env`:  
   cp configs/.env.example configs/.env
2. Убедиться, что PostgreSQL доступен на `localhost:5432`
3. Запустить сервис:  
   go run ./cmd/app/main.go

### 🔹 Через Docker Compose
docker compose up --build

---

## 📖 API примеры

- Health-check:  
  curl http://localhost:8080/healthz

- Создать подписку:  
  curl -X POST http://localhost:8080/api/v1/subscriptions/ -H "Content-Type: application/json" -d '{"service_name":"Netflix","price":899,"user_id":"66061fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"2025-07"}'

- Получить список:  
  curl "http://localhost:8080/api/v1/subscriptions?user_id=66061fee-2bf1-4721-ae6f-7636e79a0cba"

- Получить по ID:  
  curl http://localhost:8080/api/v1/subscriptions/{id}

- Обновить:  
  curl -X PUT http://localhost:8080/api/v1/subscriptions/{id} -H "Content-Type: application/json" -d '{"price":999}'

- Удалить:  
  curl -X DELETE http://localhost:8080/api/v1/subscriptions/{id}

- Подсчёт суммы:  
  curl "http://localhost:8080/api/v1/subscriptions/total?user_id=66061fee-2bf1-4721-ae6f-7636e79a0cba&from=2025-07&to=2025-12"

---

## 📑 Swagger
- Документация доступна: http://localhost:8080/swagger/index.html
- Обновление документации: swag init -g cmd/app/main.go -o docs

---

## 📂 Структура проекта
cmd/app           → main.go — точка входа  
internal/api      → handlers + router  
internal/service  → бизнес-логика  
internal/repo     → работа с БД (pgx)  
internal/model    → модели  
internal/config   → конфигурация (env/yaml)  
internal/log      → логгер (slog)  
migrations        → SQL-миграции  
configs           → .env/.yaml конфиги  
docs              → swagger docs

---

## 🛠 Технологии
Go 1.23  
PostgreSQL 14+  
pgx v5  
chi router  
slog logger  
swaggo (Swagger)  
Docker + Docker Compose

---

## 🚀 Запуск тестового окружения
docker compose down -v   # остановить и удалить тома  
docker compose up --build

После старта сервис доступен:
- API: http://localhost:8080/api/v1/...
- Swagger: http://localhost:8080/swagger/index.html

---


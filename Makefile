.PHONY: run build tidy up

build:
	go build -o bin/app ./cmd/subscription-service

tidy:
	go mod tidy

run:
	go run ./cmd/subscription-service

up:
	docker compose up -d --build

down:
	docker compose down -v
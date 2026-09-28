GO ?= go
BINARY ?= bin/trip-service

# Читаем переменные из .env файла
include .env
export DATABASE_URL
export HTTP_ADDR
export LOG_LEVEL
export SHUTDOWN_TIMEOUT

.PHONY: install generate build run migrate migrate-down test lint help

install: ## Скачать зависимости и установить инструменты
	$(GO) mod download

generate: ## Вся кодогенерация: OpenAPI, mocks, protobuf
	$(GO) tool oapi-codegen \
		-generate types,chi-server \
		-package api \
		-o api/api.gen.go \
		contracts/openapi/trip-service.openapi.yaml
	$(GO) generate ./...

build: ## Собрать исполняемый файл
	$(GO) build -trimpath -o $(BINARY) ./cmd/trip-service

run: ## Запустить сервис локально
	$(GO) run ./cmd/trip-service

migrate: ## Накатить миграции на базу данных
	$(GO) tool goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down: ## Откатить последнюю миграцию
	$(GO) tool goose -dir migrations postgres "$(DATABASE_URL)" down

test: ## Запустить все тесты с race detector
	$(GO) test -race ./...

lint: ## Запустить линтер
	$(GO) tool golangci-lint run ./...

format: ## Отформатировать код
	$(GO) fmt ./...

check: ## Проверить go.mod и форматирование
	$(GO) mod tidy -diff
	$(GO) fmt -d ./...
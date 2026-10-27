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
	$(GO) build -trimpath -o $(BINARY) ./cmd/service

run: ## Запустить сервис локально
	$(GO) run ./cmd/service

migrate: ## Накатить миграции на базу данных
	$(GO) tool goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down: ## Откатить последнюю миграцию
	$(GO) tool goose -dir migrations postgres "$(DATABASE_URL)" down

test-race: ## Запустить все тесты с race detector
	$(GO) test -race ./...

test: ## Запустить тесты и сформировать HTML-отчет покрытия
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

test-integration:
	$(GO) test -tags=integration -race ./...

lint: ## Запустить линтер
	golangci-lint run ./...

format: ## Отформатировать код
	$(GO) fmt ./...

check: ## Проверить go.mod и форматирование
	$(GO) mod tidy -diff
	$(GO) fmt -d ./...

psql: ## Подключиться к базе данных через интерактивную консоль psql
	psql "$(DATABASE_URL)"
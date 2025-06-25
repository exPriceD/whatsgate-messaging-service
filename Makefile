.PHONY: swagger build run test clean deps all init-db build-init-db create-db migrate-up migrate-down migrate-force migrate-version new-migration help

# Определение ОС
ifeq ($(OS),Windows_NT)
    # Windows
    RM := rmdir /s /q
    RMF := del /f /q
    MKDIR := mkdir
    BIN_DIR := bin
    BINARY := whatsapp-service.exe
    INIT_DB_BINARY := init-db.exe
else
    # Unix-like системы (Linux, macOS)
    RM := rm -rf
    RMF := rm -f
    MKDIR := mkdir -p
    BIN_DIR := bin
    BINARY := whatsapp-service
    INIT_DB_BINARY := init-db
endif

# Настройки подключения к БД (можно переопределять через env или make ... VAR=...)
DB_HOST ?= localhost
DB_PORT ?= 5433
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= whatsapp_service

# Миграции через golang-migrate
MIGRATIONS_DIR := ./migrations
MIGRATE_DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATE := migrate -path $(MIGRATIONS_DIR) -database $(MIGRATE_DB_URL)

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

migrate-force:
	$(MIGRATE) force

migrate-version:
	$(MIGRATE) version

# Генерация Swagger документации
swagger:
	swag init -g cmd/main.go -o internal/docs

# Создание директории для бинарных файлов
$(BIN_DIR):
	$(MKDIR) $(BIN_DIR)

# Сборка приложения
build: $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) cmd/main.go

# Инициализация базы данных
init-db:
	$(BIN_DIR)/$(INIT_DB_BINARY)

# Запуск приложения
run:
	go run cmd/main.go

# Запуск тестов
test:
	go test ./...

# Очистка
clean:
ifeq ($(OS),Windows_NT)
	@if exist $(BIN_DIR) $(RM) $(BIN_DIR)
	@if exist internal\docs\docs.go $(RMF) internal\docs\docs.go
	@if exist internal\docs\swagger.json $(RMF) internal\docs\swagger.json
	@if exist internal\docs\swagger.yaml $(RMF) internal\docs\swagger.yaml
else
	$(RM) $(BIN_DIR)
	$(RMF) internal/docs/docs.go
	$(RMF) internal/docs/swagger.json
	$(RMF) internal/docs/swagger.yaml
endif

# Установка зависимостей
deps:
	go mod tidy
	go mod download

# Установка swag (если не установлен)
install-swag:
ifeq ($(OS),Windows_NT)
	@echo "Установка swag для Windows..."
	go install github.com/swaggo/swag/cmd/swag@latest
else
	@echo "Установка swag для Unix-like систем..."
	go install github.com/swaggo/swag/cmd/swag@latest
endif

# Полная сборка с Swagger
all: swagger build

# Проверка установки swag
check-swag:
	@swag --version > /dev/null 2>&1 || (echo "swag не установлен. Выполните: make install-swag" && exit 1)

# Генерация Swagger с проверкой
swagger-safe: check-swag swagger

# Полная сборка с проверкой Swagger
all-safe: check-swag all

# Автоматизация создания миграции: make new-migration name=...
new-migration:
ifeq ($(OS),Windows_NT)
	@if not defined name (echo Укажите имя миграции: make new-migration name=...) else migrate create -ext sql -dir ./migrations -seq $(name)
else
	@if [ -z "$(name)" ]; then \
		echo "Укажите имя миграции: make new-migration name=..."; \
	else \
		migrate create -ext sql -dir ./migrations -seq $(name); \
	fi
endif

# Создание базы данных: make create-db DB_NAME=whatsapp_service
create-db:
ifeq ($(OS),Windows_NT)
	@if not defined DB_NAME (echo Укажите имя базы: make create-db DB_NAME=whatsapp_service) else psql -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) -c "CREATE DATABASE $(DB_NAME);"
else
	@if [ -z "$(DB_NAME)" ]; then \
		echo "Укажите имя базы: make create-db DB_NAME=whatsapp_service"; \
	else \
		psql -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) -c "CREATE DATABASE $(DB_NAME);"; \
	fi
endif

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make build        - Сборка приложения"
	@echo "  make run          - Запуск приложения"
	@echo "  make test         - Запуск тестов"
	@echo "  make clean        - Очистка файлов"
	@echo "  make deps         - Установка зависимостей"
	@echo "  make swagger      - Генерация Swagger документации"
	@echo "  make migrate-up   - Применить все миграции к БД (golang-migrate)"
	@echo "  make migrate-down - Откатить одну миграцию (golang-migrate)"
	@echo "  make migrate-force - Принудительно установить версию миграции (golang-migrate)"
	@echo "  make migrate-version - Показать текущую версию миграции (golang-migrate)"
	@echo "  make new-migration name=... - Создать новую миграцию с заданным именем (golang-migrate)"
	@echo "  make create-db DB_NAME=whatsapp_service - Создать новую базу данных (PostgreSQL)"
	@echo "  make help         - Показать эту справку"
	@echo ""
	@echo "Параметры подключения к БД можно переопределять через переменные окружения или прямо в команде make:"
	@echo "  make migrate-up DB_NAME=mydb DB_USER=admin DB_PASSWORD=secret DB_PORT=5432 DB_HOST=localhost" 
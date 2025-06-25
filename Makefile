.PHONY: swagger build run test clean deps all init-db build-init-db

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

# Генерация Swagger документации
swagger:
	swag init -g cmd/main.go -o internal/docs

# Создание директории для бинарных файлов
$(BIN_DIR):
	$(MKDIR) $(BIN_DIR)

# Сборка приложения
build: $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) cmd/main.go

# Сборка утилиты инициализации БД
build-init-db: $(BIN_DIR)
	go build -o $(BIN_DIR)/$(INIT_DB_BINARY) cmd/init_db/main.go

# Инициализация базы данных
init-db: build-init-db
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

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make build        - Сборка приложения"
	@echo "  make run          - Запуск приложения"
	@echo "  make test         - Запуск тестов"
	@echo "  make clean        - Очистка файлов"
	@echo "  make deps         - Установка зависимостей"
	@echo "  make swagger      - Генерация Swagger документации"
	@echo "  make swagger-safe - Генерация Swagger с проверкой установки"
	@echo "  make all          - Полная сборка с Swagger"
	@echo "  make all-safe     - Полная сборка с проверкой Swagger"
	@echo "  make install-swag - Установка swag"
	@echo "  make init-db      - Инициализация базы данных"
	@echo "  make build-init-db - Сборка утилиты инициализации БД"
	@echo "  make help         - Показать эту справку" 
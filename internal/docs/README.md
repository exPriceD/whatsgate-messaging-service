# Swagger Documentation

Эта папка содержит автоматически сгенерированную Swagger документацию для WhatsApp Service API.

## Файлы

- `docs.go` - Go код с метаданными API (автогенерируется)
- `swagger.json` - OpenAPI спецификация в JSON формате (автогенерируется)
- `swagger.yaml` - OpenAPI спецификация в YAML формате (автогенерируется)

## Генерация

Для обновления документации после изменения аннотаций в коде:

```bash
make swagger
```

## Аннотации

Swagger аннотации добавляются в:
- `cmd/main.go` - общая информация об API
- `internal/delivery/http/handlers.go` - эндпоинты
- `internal/delivery/http/types.go` - типы данных

## Доступ

Swagger UI доступен по адресу: `http://localhost:8080/swagger/index.html` 
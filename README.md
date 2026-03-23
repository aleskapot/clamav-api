# ClamAV API

RESTful API для проверки файлов на вирусы с использованием ClamAV.

## Возможности

- Синхронное сканирование файлов (`/files/scan`)
- Асинхронное сканирование с webhook уведомлениями (`/files/upload`)
- Проверка здоровья сервиса (`/health`, `/ready`)
- Prometheus метрики (`/metrics`)
- Swagger UI (`/swagger`) и OpenAPI спецификация (`/swagger.yaml`)
- Авторизация по API ключу

## Требования

- Go 1.26+
- ClamAV (работающий в режиме TCP)

## Установка

```bash
# Клонирование репозитория
git clone <repository-url>
cd clamav-api

# Запуск
go run ./cmd/server
```

## Конфигурация

Создайте файл `configs/config.yaml`:

```yaml
app:
  host: "0.0.0.0"
  port: 8080
  max_file_size: 100  # MB

clamav:
  host: "localhost"
  port: 3310
  timeout: 60s

auth:
  api_key: "your-secret-api-key"

webhook:
  url: "http://localhost:8081/webhook"
  timeout: 30s
  retry_count: 3

storage:
  temp_dir: "/tmp/clamav-api"
```

### Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `CONFIG_PATH` | Путь к config.yaml | `configs/config.yaml` |

## API Endpoints

### Health Check (без авторизации)

```bash
GET /health
```

Ответ:
```json
{"status": "ok"}
```

### Readiness Check (требуется API ключ)

```bash
GET /ready
Header: API-Key: your-api-key
```

Ответ:
```json
{
  "status": "ok",
  "services": {
    "clamav": "ok"
  }
}
```

### Синхронное сканирование

```bash
POST /files/scan
Header: API-Key: your-api-key
Content-Type: multipart/form-data

file: <binary>
```

Ответ:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440000",
  "file_name": "document.pdf",
  "file_size": 1024000,
  "result": "clean",
  "duration_ms": 150,
  "scanned_at": "2026-03-21T12:00:00Z"
}
```

### Асинхронная загрузка

```bash
POST /files/upload
Header: API-Key: your-api-key
Content-Type: multipart/form-data

file: <binary>
```

Ответ (HTTP 202):
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440002",
  "file_name": "large_file.zip",
  "file_size": 52428800,
  "message": "File uploaded and queued for scanning",
  "received_at": "2026-03-21T12:00:00Z"
}
```

### Webhook Payload

При асинхронном сканировании результат отправляется на настроенный URL:

```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440002",
  "file_name": "large_file.zip",
  "file_size": 52428800,
  "result": "clean",
  "scanned_at": "2026-03-21T12:00:30Z",
  "duration_ms": 500
}
```

## Prometheus метрики

```bash
GET /metrics
```

Доступные метрики:
- `clamav_http_requests_total` - счётчик HTTP запросов
- `clamav_http_request_duration_seconds` - гистограмма времени обработки
- `clamav_files_scanned_total` - счётчик проверенных файлов
- `clamav_file_size_bytes` - гистограмма размеров файлов

## Docker

```bash
# Сборка
docker build -t clamav-api .

# Запуск
docker run -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/etc/clamav-api/config.yaml \
  clamav-api
```

## Docker Compose

Запуск всех сервисов (ClamAV + API + webhook):

```bash
docker-compose up -d
```

Просмотр логов:

```bash
docker-compose logs -f
```

Остановка:

```bash
docker-compose down
```

## Тестирование

```bash
go test ./tests/... -v
```

## Структура проекта

```
.
├── cmd/
│   └── server/
│       └── main.go           # Точка входа
├── internal/
│   ├── config/               # Конфигурация (viper)
│   ├── handler/              # HTTP обработчики
│   ├── middleware/           # Auth, logging, metrics
│   ├── clamscan/            # ClamAV клиент
│   └── model/               # Модели данных
├── tests/                   # Тесты
├── configs/
│   └── config.yaml         # Конфигурация
├── docs/
│   └── openapi.yaml        # OpenAPI спецификация
└── Dockerfile
```

## Лицензия

MIT

# ClamAV API

RESTful API for file virus scanning using ClamAV.

## Features

- Synchronous file scanning (`/files/scan`)
- Asynchronous scanning with webhook notifications (`/files/upload`)
- Health checks (`/health`, `/ready`)
- Prometheus metrics (`/metrics`)
- Swagger UI (`/swagger`) and OpenAPI specification (`/swagger.yaml`)
- API key authorization

## Requirements

- Go 1.26+
- ClamAV (running in TCP mode)

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd clamav-api

# Run
go run ./cmd/server
```

## Configuration

Create a `configs/config.yaml` file:

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

### Environment Variables

| Variable      | Description         | Default Value         |
|---------------|---------------------|-----------------------|
| `CONFIG_PATH` | Path to config.yaml | `configs/config.yaml` |

## API Endpoints

### Health Check (no authorization)

```bash
GET /health
```

Response:
```json
{"status": "ok"}
```

### Readiness Check (no authorization)

```bash
GET /ready
```

Response:
```json
{
  "status": "ok",
  "services": {
    "clamav": "ok"
  }
}
```

### ClamAV Info (no authorization)

```bash
GET /info
```

### Synchronous Scan (API key required)

```bash
POST /files/scan
Header: API-Key: your-api-key
Content-Type: multipart/form-data

file: <binary>
```

Response:
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

### Asynchronous Upload (API key required)

```bash
POST /files/upload
Header: API-Key: your-api-key
Content-Type: multipart/form-data

file: <binary>
```

Response (HTTP 202):
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

For asynchronous scanning, the result is sent to the configured URL:

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

## Prometheus Metrics

```bash
GET /metrics
```

Available metrics:
- `clamav_http_requests_total` - HTTP request counter
- `clamav_http_request_duration_seconds` - request duration histogram
- `clamav_files_scanned_total` - scanned files counter
- `clamav_file_size_bytes` - file size histogram

## Docker

```bash
# Build
docker build -t clamav-api .

# Run
docker run -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/etc/clamav-api/config.yaml \
  clamav-api
```

## Docker Compose

Run all services (ClamAV + API + webhook):

```bash
docker-compose up -d
```

View logs:

```bash
docker-compose logs -f
```

Stop:

```bash
docker-compose down
```

## Testing

```bash
go test ./tests/... -v
```

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── config/               # Configuration (viper)
│   ├── handler/              # HTTP handlers
│   ├── middleware/           # Auth, logging, metrics
│   ├── clamscan/            # ClamAV client
│   └── model/               # Data models
├── tests/                   # Tests
├── configs/
│   └── config.yaml         # Configuration
├── docs/
│   └── openapi.yaml        # OpenAPI specification
└── Dockerfile
```

## License

MIT

# ClamAV API - Agent Instructions

## Project Overview

RESTful API for file scanning using ClamAV antivirus. Supports synchronous and asynchronous scanning modes with webhook notifications.

**Stack:** Go 1.26+, Fiber v2, Zap, Viper, Prometheus

## Project Structure

```
clamav-api/
├── cmd/server/main.go           # Entry point
├── internal/
│   ├── api/server.go            # API server setup, routes, graceful shutdown
│   ├── config/config.go         # Configuration loading with Viper + validation
│   ├── handler/
│   │   ├── health.go            # /health, /ready endpoints
│   │   └── files.go             # /files/scan, /files/upload endpoints + webhook
│   ├── middleware/
│   │   ├── auth.go              # API Key authentication middleware
│   │   └── metrics.go           # Prometheus metrics middleware
│   ├── clamscan/client.go       # ClamAV TCP client (INSTREAM protocol)
│   ├── logger/logger.go         # Zap logger initialization
│   └── model/response.go        # Data models and response types
├── configs/
│   └── config.yaml             # Configuration file
├── docs/
│   └── openapi.yaml             # OpenAPI 3.0 specification
├── tests/                       # Unit tests
├── Dockerfile
├── go.mod / go.sum
└── README.md
```

## Code Conventions

### Go Conventions
- Use `go fmt` before commits
- Run `go mod tidy` when adding/removing dependencies
- All exported functions have documentation comments
- Error wrapping with `fmt.Errorf("...: %w", err)`

### Naming
- Package names: lowercase, single word (e.g., `handler`, `middleware`)
- Struct names: PascalCase
- Variable names: camelCase, short for locals, descriptive for exports
- File names: lowercase with underscores for packages (`auth.go`, `auth_test.go`)

### HTTP Handlers
- Return appropriate HTTP status codes
- JSON responses use models from `internal/model`
- Log errors with context using zap

### Configuration
- All config via `config.yaml` (Viper)
- Required fields validated in `config.go`
- Environment variables supported (uppercase with underscores)

## Commands

### Development
```bash
# Run the server
go run ./cmd/server

# Run with custom config
CONFIG_PATH=/path/to/configs/config.yaml go run ./cmd/server
```

### Testing
```bash
# Run all tests
go test ./tests/... -v

# Run with coverage
go test ./tests/... -cover

# Run specific test
go test ./tests/... -v -run TestAuthMiddleware
```

### Building
```bash
# Build binary
go build -o server ./cmd/server

# Build for Docker
docker build -t clamav-api .
```

### Code Quality
```bash
# Format code
go fmt ./...

# Tidy dependencies
go mod tidy

# Verify build
go build ./...
```

## Configuration

### config.yaml
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
  api_key: "change-me-in-production"

webhook:
  url: "http://localhost:8081/webhook"
  timeout: 30s
  retry_count: 3

storage:
  temp_dir: "/tmp/clamav-api"
```

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_PATH` | Path to config file | `configs/config.yaml` |

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | No | Liveness probe |
| GET | `/ready` | Yes | Readiness check (ClamAV) |
| GET | `/info` | Yes | ClamAV version and stats |
| POST | `/files/scan` | Yes | Synchronous scan |
| POST | `/files/upload` | Yes | Async upload + webhook |
| GET | `/metrics` | No | Prometheus metrics |
| GET | `/swagger` | No | Swagger UI |
| GET | `/swagger.yaml` | No | OpenAPI specification |

## Adding New Endpoints

1. Add route in `internal/api/server.go`
2. Create handler in `internal/handler/`
3. Define response models in `internal/model/`
4. Update `docs/openapi.yaml`
5. Add tests in `tests/`

## Testing Guidelines

- Unit tests in `tests/` package
- Test file naming: `*_test.go`
- Use `github.com/stretchr/testify/assert` and `require`
- Mock external dependencies (ClamAV, HTTP clients)
- Aim for >70% coverage on handlers and business logic

## Adding Dependencies

```bash
# Add dependency
go get github.com/example/package

# Update go.mod and go.sum
go mod tidy
```

## Docker

```bash
# Build image
docker build -t clamav-api .

# Run container
docker run -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/etc/clamav-api/config.yaml \
  clamav-api

# Run with docker-compose (requires docker-compose.test.yaml)
docker-compose up
```

## Common Issues

### ClamAV Connection Failed
- Verify ClamAV is running with `clamd` daemon
- Check `clamav.host` and `clamav.port` in config
- Test connectivity: `telnet localhost 3310`

### File Too Large
- Increase `app.max_file_size` in config
- Verify Fiber body limit matches

### Auth Not Working
- Ensure `API-Key` header is passed (uppercase)
- Check `auth.api_key` matches in config

## Notes for Agents

- This project uses Go 1.26 features (check compatibility if downgrading)
- ClamAV client uses raw TCP socket with INSTREAM protocol
- Webhook uses standard `net/http` client with retries
- Prometheus metrics use `promauto` for automatic registration
- Graceful shutdown waits for in-flight requests (30s timeout)

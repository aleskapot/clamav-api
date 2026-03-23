# Build stage
FROM golang:1.26-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=1.0.0
ARG BUILD_TIME
ARG GIT_COMMIT

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /server \
    ./cmd/server

# Production stage
FROM alpine:latest

# Environment variables
ENV TZ=Europe/Moscow

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    mkdir -p /tmp/clamav-api && \
    echo "${TZ}" && \
    date

# Copy binary from builder
COPY --from=builder /server /server
# Copy configuration files
COPY --from=builder /app/configs/config.yaml /etc/clamav-api/config.yaml
# Copy OpenAPI spec
COPY --from=builder /app/docs/openapi.yaml /app/docs/openapi.yaml

# Create non-root user
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -s /bin/sh -D appuser

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /app

# Expose ports
EXPOSE 8080

ENV CONFIG_PATH=/etc/clamav-api/config.yaml

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/server"]

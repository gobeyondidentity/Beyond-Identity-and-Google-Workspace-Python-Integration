# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION:-dev} -X main.commit=${COMMIT:-unknown} -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o scim-sync ./cmd

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary
COPY --from=builder /app/scim-sync /usr/local/bin/scim-sync

# Create non-root user
RUN adduser -D -s /bin/sh scimuser
USER scimuser

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD scim-sync version || exit 1

# Default command
ENTRYPOINT ["scim-sync"]
CMD ["server"]
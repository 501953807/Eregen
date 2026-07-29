# syntax=docker/dockerfile:1

# ---- stage 1: build ----
FROM golang:1.22-alpine AS builder

RUN apk --no-cache add git ca-certificates

WORKDIR /app

# Cache module download
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

# Copy source
COPY . .

# Build with security flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api-server ./cmd/

# ---- stage 2: run ----
FROM alpine:3.19

# Create a non-root user for running the application
RUN addgroup -S appuser && adduser -S appuser -G appuser

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /api-server .

# Set ownership to non-root user
RUN chown appuser:appuser /api-server

EXPOSE 8080

# Run as non-root user
USER appuser

# Security recommendations for production:
# -- Read-only root filesystem (requires tmpfs for /tmp, /var/run etc.)
# -- No shell access (remove /bin/sh)
# -- AppArmor/SELinux profiles enabled at host level
# -- Network policies restricting egress traffic

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- https://localhost:8080/health || exit 1

ENTRYPOINT ["./api-server"]

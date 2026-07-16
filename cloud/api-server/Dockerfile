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

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api-server ./cmd/

# ---- stage 2: run ----
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /api-server .

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./api-server"]

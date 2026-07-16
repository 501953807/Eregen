# © 2026 Eregen (颐贞). All rights reserved.

# ---- Stage 1: Build ----
FROM golang:1.22-alpine AS builder

RUN apk --no-cache add git ca-certificates tzdata

WORKDIR /build

# Cache module download
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
ARG LDFLAGS="-s -w"
RUN CGO_ENABLED=0 go build \
    -ldflags "$LDFLAGS" \
    -o /usr/local/bin/gateway \
    ./cmd/main.go

# ---- Stage 2: Runtime ----
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 1001 eregen && \
    adduser -D -u 1001 -G eregen eregen

USER eregen:eregen

COPY --from=builder /usr/local/bin/gateway /usr/local/bin/gateway
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

WORKDIR /app

ENTRYPOINT ["gateway"]
CMD []

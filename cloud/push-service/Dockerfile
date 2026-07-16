FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /push-service ./cmd/main.go

FROM alpine:3.19

RUN addgroup -S eregen && adduser -S eregen -G eregen
USER eregen

COPY --from=builder /push-service /usr/local/bin/push-service

EXPOSE 8085
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://localhost:8085/health || exit 1

ENTRYPOINT ["push-service"]

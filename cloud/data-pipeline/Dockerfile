FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /pipeline ./cmd/main.go

FROM alpine:3.19

RUN addgroup -S eregen && adduser -S eregen -G eregen
USER eregen

COPY --from=builder /pipeline /usr/local/bin/pipeline

EXPOSE 8087
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -qO- http://localhost:8087/health || exit 1

ENTRYPOINT ["pipeline"]

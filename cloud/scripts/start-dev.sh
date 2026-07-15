#!/bin/bash
# Eregen Platform - One-click Development Infrastructure Startup
# © 2026 Eregen (颐贞). All rights reserved.

set -e

echo "Starting Eregen development infrastructure..."
cd "$(dirname "$0")/.."

# Start all services
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d

echo "Waiting for services to be healthy..."

# Wait loop checking docker compose ps for health status
MAX_WAIT=180
ELAPSED=0
while [ $ELAPSED -lt $MAX_WAIT ]; do
    HEALTHY=true
    while IFS= read -r line; do
        if echo "$line" | grep -q '"Health":"healthy"'; then
            continue
        elif echo "$line" | grep -q '"Health":"starting"\|"Health":"healthy"'; then
            continue
        else
            if echo "$line" | grep -qE '"Health":"(unhealthy|starting)|\"Status\":\"(created|restarting)"'; then
                HEALTHY=false
            fi
        fi
    done < <(docker compose -f docker-compose.yml -f docker-compose.dev.yml ps --format json 2>/dev/null)

    if [ "$HEALTHY" = true ]; then
        echo "Infrastructure ready!"
        echo "  PostgreSQL:     localhost:5432"
        echo "  InfluxDB:       http://localhost:8086"
        echo "  Redis:          localhost:6379"
        echo "  NATS:           localhost:4222"
        echo "  EMQX Dashboard: http://localhost:18083"
        echo "  Grafana:        http://localhost:3000"
        echo "  Prometheus:     http://localhost:9090"
        echo "  Loki:           http://localhost:3100"
        exit 0
    fi

    sleep 5
    ELAPSED=$((ELAPSED + 5))
    echo "  Waiting... (${ELAPSED}s/${MAX_WAIT}s)"
done

echo "Warning: Not all services became healthy within ${MAX_WAIT}s. Check with:"
echo "  docker compose -f docker-compose.yml -f docker-compose.dev.yml ps"

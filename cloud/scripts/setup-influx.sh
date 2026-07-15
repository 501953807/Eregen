#!/bin/bash
# Eregen Platform - Standalone InfluxDB Setup Script
# For initializing InfluxDB outside Docker (for testing)
# © 2026 Eregen (颐贞). All rights reserved.

set -euo pipefail

INFLUX_URL="${INFLUX_URL:-http://localhost:8086}"
INFLUX_TOKEN="${INFLUX_TOKEN:-eregen_admin_token}"
INFLUX_USER="${INFLUX_USER:-eregen}"
INFLUX_PASSWORD="${INFLUX_PASSWORD:-eregen_dev_password}"

echo "Setting up standalone InfluxDB at ${INFLUX_URL}..."

# Create organization
echo "Creating organization: eregen"
influx org create \
    --name "eregen" \
    --user "$INFLUX_USER" \
    --token "$INFLUX_TOKEN" \
    --url "$INFLUX_URL" 2>/dev/null || echo "Organization 'eregen' may already exist."

# Create buckets with retention policies
echo "Creating buckets..."

influx bucket create \
    --name "health-data" \
    --retention 30 \
    --org "eregen" \
    --token "$INFLUX_TOKEN" \
    --url "$INFLUX_URL" 2>/dev/null || echo "Bucket 'health-data' may already exist."

influx bucket create \
    --name "alerts" \
    --retention 90 \
    --org "eregen" \
    --token "$INFLUX_TOKEN" \
    --url "$INFLUX_URL" 2>/dev/null || echo "Bucket 'alerts' may already exist."

influx bucket create \
    --name "device-metadata" \
    --retention 180 \
    --org "eregen" \
    --token "$INFLUX_TOKEN" \
    --url "$INFLUX_URL" 2>/dev/null || echo "Bucket 'device-metadata' may already exist."

echo "InfluxDB setup complete."
echo "  Organization: eregen"
echo "  Buckets: health-data (30d), alerts (90d), device-metadata (180d)"

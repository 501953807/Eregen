#!/bin/bash
# Test API connectivity for all firmware device types
# Verifies that the backend APIs accept the device message formats
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
ADMIN_API="${EREGEN_ADMIN_API:-http://localhost:8081}"
DEVICE_API="${EREGEN_DEVICE_API:-http://localhost:8080}"

echo "=== Eregen API Connector Test ==="
echo ""

# Health check
echo "[1] Health Check"
for svc in "API Server ($DEVICE_API)" "Admin API ($ADMIN_API)"; do
    url=$(echo "$svc" | grep -oP 'http://\S+')
    status=$(curl -s -o /dev/null -w "%{http_code}" "$url/api/v1/health" 2>/dev/null || echo "000")
    if [ "$status" = "200" ]; then
        echo "  ✓ $svc — OK"
    else
        echo "  ✗ $svc — NOT RUNNING (status=$status)"
    fi
done

echo ""
echo "[2] Auth Test (Admin API)"
login_res=$(curl -s -X POST "$ADMIN_API/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"identifier":"admin@example.com","password":"Admin@123"}' 2>/dev/null)
token=$(echo "$login_res" | grep -oP '"access_token":"\K[^"]+' || echo "")
if [ -n "$token" ]; then
    echo "  ✓ Admin login OK, token obtained"
    AUTH_HEADER="Authorization: Bearer $token"
else
    echo "  ⚠ Admin login failed — using unauthenticated mode"
    AUTH_HEADER=""
fi

echo ""
echo "[3] Device Telemetry Simulation"
# Simulate bracelet heartbeat
curl -s -X POST "$DEVICE_API/api/v1/devices/telemetry" \
    -H "Content-Type: application/json" \
    -d '{"dev_id":"BR-TEST01","type":"heartbeat","bat":95,"ts":'"$(date +%s)"'}' \
    -w "\n  Bracelet heartbeat: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Telemetry endpoint not reachable"

# Simulate pillbox med status
curl -s -X POST "$DEVICE_API/api/v1/devices/telemetry" \
    -H "Content-Type: application/json" \
    -d '{"dev_id":"PX-TEST01","type":"med_status","compartment":3,"taken":true,"ts":'"$(date +%s)"'}' \
    -w "\n  Pillbox med_status: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Telemetry endpoint not reachable"

# Simulate wristband heartbeat
curl -s -X POST "$DEVICE_API/api/v1/devices/telemetry" \
    -H "Content-Type: application/json" \
    -d '{"dev_id":"WB-TEST01","type":"heartbeat","bat":92,"ts":'"$(date +%s)"'}' \
    -w "\n  Wristband heartbeat: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Telemetry endpoint not reachable"

echo ""
echo "[4] Admin API CRUD Check"
# List elderly
curl -s "$ADMIN_API/api/v1/admin/elderly?page=1&page_size=5" \
    -H "$AUTH_HEADER" \
    -w "\n  GET /admin/elderly: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Endpoint not reachable"

# List devices
curl -s "$ADMIN_API/api/v1/admin/devices?page=1&page_size=5" \
    -H "$AUTH_HEADER" \
    -w "\n  GET /admin/devices: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Endpoint not reachable"

# List alerts
curl -s "$ADMIN_API/api/v1/admin/alerts?limit=5" \
    -H "$AUTH_HEADER" \
    -w "\n  GET /admin/alerts: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Endpoint not reachable"

# Medical wristband patients
curl -s "$ADMIN_API/api/v1/admin/medical/patients?page=1&page_size=5" \
    -H "$AUTH_HEADER" \
    -w "\n  GET /admin/medical/patients: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Endpoint not reachable"

# Community elders
curl -s "$ADMIN_API/api/v1/admin/community-wb/elders?page=1&page_size=5" \
    -H "$AUTH_HEADER" \
    -w "\n  GET /admin/community-wb/elders: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Endpoint not reachable"

echo ""
echo "[5] Regulatory Dashboard Check"
curl -s "$ADMIN_API/api/v1/admin/regulatory/dashboard/patient-overview" \
    -H "$AUTH_HEADER" \
    -w "\n  GET /regulatory/overview: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Endpoint not reachable"

curl -s "$ADMIN_API/api/v1/admin/regulatory/rules" \
    -H "$AUTH_HEADER" \
    -w "\n  GET /regulatory/rules: HTTP %{http_code}\n" 2>/dev/null || echo "  ⚠ Endpoint not reachable"

echo ""
echo "=== Test Complete ==="

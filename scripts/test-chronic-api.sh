#!/bin/bash
# Eregen Chronic Care API Test Script
# Tests all 17 chronic care endpoints
# Usage: ./scripts/test-chronic-api.sh

set -euo pipefail

# ── Configuration ──────────────────────────────────────────────────────────────
BASE_URL="http://localhost:8080"
ELDERLY_ID="test-elderly-001"
TEST_USER="test-user-chronic"
TEST_PASS="TestPass123!"

# ── Counters ───────────────────────────────────────────────────────────────────
PASS=0
FAIL=0
SKIP=0
TOTAL=0

# ── Colors ─────────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ── Helper Functions ───────────────────────────────────────────────────────────
log_test() {
    echo -e "${BLUE}▶ ${1}${NC}"
}

log_pass() {
    echo -e "  ${GREEN}✔ PASS${NC} — $1"
    ((PASS++))
    ((TOTAL++))
}

log_fail() {
    echo -e "  ${RED}✗ FAIL${NC} — $1"
    ((FAIL++))
    ((TOTAL++))
}

log_skip() {
    echo -e "  ${YELLOW}○ SKIP${NC} — $1"
    ((SKIP++))
    ((TOTAL++))
}

log_header() {
    echo ""
    echo -e "${CYAN}══════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}══════════════════════════════════════════════════════════${NC}"
}

# Check if server is reachable
check_server() {
    if ! curl -sf "${BASE_URL}/api/v1/health" > /dev/null 2>&1; then
        echo -e "${RED}ERROR: Server not reachable at ${BASE_URL}${NC}"
        echo "  Please start the api-server first: ./scripts/start.sh start api-server"
        exit 1
    fi
}

# Register or login to get JWT token and CSRF token
auth_setup() {
    log_header "Authentication Setup"

    # Step 1: Get CSRF token
    CSRF_RESP=$(curl -sf "${BASE_URL}/api/v1/auth/csrf/get")
    CSRF_TOKEN=$(echo "$CSRF_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('csrf_token',''))" 2>/dev/null || echo "")
    if [ -z "$CSRF_TOKEN" ]; then
        log_skip "Could not get CSRF token from /auth/csrf/get"
    else
        log_pass "Got CSRF token"
    fi

    # Step 2: Try to register
    REGISTER_RESP=$(curl -sf -X POST "${BASE_URL}/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"${TEST_USER}@test.local\",
            \"phone\": \"+8613800000001\",
            \"password\": \"${TEST_PASS}\",
            \"name\": \"Chronic Test User\",
            \"otp_code\": \"000000\"
        }" 2>/dev/null || echo "{}")
    REGISTER_CODE=$(echo "$REGISTER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code','UNKNOWN'))" 2>/dev/null || echo "UNKNOWN")

    if [ "$REGISTER_CODE" = "OK" ]; then
        log_pass "User registered successfully"
    else
        log_skip "User registration skipped (user may already exist): ${REGISTER_CODE}"
    fi

    # Step 3: Login
    LOGIN_RESP=$(curl -sf -X POST "${BASE_URL}/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"identifier\": \"${TEST_USER}@test.local\", \"password\": \"${TEST_PASS}\"}" 2>/dev/null || echo "{}")
    JWT_TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('access_token',''))" 2>/dev/null || echo "")

    if [ -n "$JWT_TOKEN" ]; then
        log_pass "Logged in successfully (JWT token obtained)"
    else
        log_skip "Login failed — chronic endpoints require authentication"
        echo "  Response: ${LOGIN_RESP}"
        JWT_TOKEN=""
    fi

    # Build auth headers
    AUTH_HEADERS=(
        "Authorization: Bearer ${JWT_TOKEN}"
        "Content-Type: application/json"
    )
    if [ -n "$CSRF_TOKEN" ]; then
        AUTH_HEADERS+=("X-CSRF-Token: ${CSRF_TOKEN}")
    fi
}

# Make a GET request
do_get() {
    local desc="$1"
    local url="$2"
    local expected_status="${3:-200}"
    local extra_args=()
    for h in "${AUTH_HEADERS[@]}"; do
        extra_args+=(-H "$h")
    done
    TOTAL=$((TOTAL + 1))
    log_test "$desc"
    RESP=$(curl -sf -w "\n%{http_code}" "${extra_args[@]}" "${url}" 2>/dev/null || echo "000")
    HTTP_CODE=$(echo "$RESP" | tail -1)
    BODY=$(echo "$RESP" | sed '$d')
    if [ "$HTTP_CODE" = "$expected_status" ]; then
        log_pass "HTTP ${HTTP_CODE}"
    else
        log_fail "Expected ${expected_status}, got ${HTTP_CODE}"
        echo "    Response: ${BODY:0:200}"
    fi
}

# Make a POST request
do_post() {
    local desc="$1"
    local url="$2"
    local body="$3"
    local expected_status="${4:-200}"
    local extra_args=()
    for h in "${AUTH_HEADERS[@]}"; do
        extra_args+=(-H "$h")
    done
    TOTAL=$((TOTAL + 1))
    log_test "$desc"
    RESP=$(curl -sf -w "\n%{http_code}" -X POST "${extra_args[@]}" -d "$body" "${url}" 2>/dev/null || echo "000")
    HTTP_CODE=$(echo "$RESP" | tail -1)
    BODY=$(echo "$RESP" | sed '$d')
    if [ "$HTTP_CODE" = "$expected_status" ]; then
        log_pass "HTTP ${HTTP_CODE}"
    else
        log_fail "Expected ${expected_status}, got ${HTTP_CODE}"
        echo "    Body: ${BODY:0:300}"
    fi
    echo "$BODY"
}

# Make a PUT request
do_put() {
    local desc="$1"
    local url="$2"
    local body="$3"
    local expected_status="${4:-200}"
    local extra_args=()
    for h in "${AUTH_HEADERS[@]}"; do
        extra_args+=(-H "$h")
    done
    TOTAL=$((TOTAL + 1))
    log_test "$desc"
    RESP=$(curl -sf -w "\n%{http_code}" -X PUT "${extra_args[@]}" -d "$body" "${url}" 2>/dev/null || echo "000")
    HTTP_CODE=$(echo "$RESP" | tail -1)
    BODY=$(echo "$RESP" | sed '$d')
    if [ "$HTTP_CODE" = "$expected_status" ]; then
        log_pass "HTTP ${HTTP_CODE}"
    else
        log_fail "Expected ${expected_status}, got ${HTTP_CODE}"
        echo "    Body: ${BODY:0:300}"
    fi
    echo "$BODY"
}

# ── Main Test Suite ─────────────────────────────────────────────────────────────
main() {
    log_header "Eregen Chronic Care API Test Suite"
    echo "  Base URL : ${BASE_URL}"
    echo "  Elderly  : ${ELDERLY_ID}"
    echo "  Date     : $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""

    # Pre-flight check
    check_server

    # Auth setup
    auth_setup
    echo ""

    # ── 1. POST /chronic/:elderly_id/glucose (manual entry) ──────────────────
    do_post \
        "POST /chronic/:elderly_id/glucose — manual glucose entry" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/glucose" \
        '{"value": 6.8, "unit": "mmol/L", "test_mode": "fasting", "source": "manual"}' \
        201
    echo ""

    # ── 2. GET /chronic/:elderly_id/glucose ──────────────────────────────────
    do_get \
        "GET /chronic/:elderly_id/glucose — list glucose records" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/glucose?days=7" \
        200
    echo ""

    # ── 3. GET /chronic/:elderly_id/glucose/trend ────────────────────────────
    do_get \
        "GET /chronic/:elderly_id/glucose/trend — glucose trend data" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/glucose/trend?days=7" \
        200
    echo ""

    # ── 4. POST /chronic/:elderly_id/test-strip/read ─────────────────────────
    do_post \
        "POST /chronic/:elderly_id/test-strip/read — test-strip device read" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/test-strip/read" \
        '{"value": 7.2, "unit": "mmol/L", "test_mode": "random", "source": "test_strip", "quality": 0.92}' \
        201
    echo ""

    # ── 5. POST /chronic/:elderly_id/uric-acid ───────────────────────────────
    do_post \
        "POST /chronic/:elderly_id/uric-acid — uric acid manual entry" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/uric-acid" \
        '{"value": 380, "unit": "μmol/L", "source": "manual"}' \
        201
    echo ""

    # ── 6. GET /chronic/:elderly_id/uric-acid ────────────────────────────────
    do_get \
        "GET /chronic/:elderly_id/uric-acid — list uric acid records" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/uric-acid?days=30" \
        200
    echo ""

    # ── 7. POST /chronic/:elderly_id/blood-pressure ──────────────────────────
    do_post \
        "POST /chronic/:elderly_id/blood-pressure — BP manual entry" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/blood-pressure" \
        '{"systolic": 135, "diastolic": 85, "pulse": 72}' \
        201
    echo ""

    # ── 8. GET /chronic/:elderly_id/blood-pressure ───────────────────────────
    do_get \
        "GET /chronic/:elderly_id/blood-pressure — list BP records" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/blood-pressure?days=30" \
        200
    echo ""

    # ── 9. POST /chronic/:elderly_id/bp-device/sync ──────────────────────────
    do_post \
        "POST /chronic/:elderly_id/bp-device/sync — sync from BP device" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/bp-device/sync" \
        '{"systolic": 128, "diastolic": 80, "pulse": 68, "source": "device"}' \
        201
    echo ""

    # ── 10. POST /chronic/:elderly_id/diet ───────────────────────────────────
    do_post \
        "POST /chronic/:elderly_id/diet — diet/meal record" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/diet" \
        '{"meal_type": "breakfast", "food_items": "[\"oatmeal\", \"banana\", \"green tea\"]", "total_carbs": 45.0, "total_calories": 320.0}' \
        201
    echo ""

    # ── 11. GET /chronic/:elderly_id/diet ────────────────────────────────────
    do_get \
        "GET /chronic/:elderly_id/diet — list diet records" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/diet" \
        200
    echo ""

    # ── 12. POST /chronic/:elderly_id/exercise ───────────────────────────────
    do_post \
        "POST /chronic/:elderly_id/exercise — exercise session record" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/exercise" \
        '{"type": "walking", "duration_min": 30, "calories": 120.0, "avg_hr": 85, "max_hr": 105}' \
        201
    echo ""

    # ── 13. GET /chronic/:elderly_id/exercise ────────────────────────────────
    do_get \
        "GET /chronic/:elderly_id/exercise — list exercise records" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/exercise" \
        200
    echo ""

    # ── 14. GET /chronic/:elderly_id/daily-tasks ─────────────────────────────
    do_get \
        "GET /chronic/:elderly_id/daily-tasks — list today's tasks" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/daily-tasks?date=2026-08-08" \
        200
    echo ""

    # ── 15. PUT /chronic/:elderly_id/daily-tasks/:task_id ────────────────────
    # First get a task ID from the daily-tasks list above
    TASK_ID=$(curl -sf "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/daily-tasks?date=2026-08-08" \
        -H "Authorization: Bearer ${JWT_TOKEN}" \
        -H "Content-Type: application/json" 2>/dev/null \
        | python3 -c "
import sys, json
data = json.load(sys.stdin).get('data', [])
if data and len(data) > 0:
    print(data[0].get('id', ''))
" 2>/dev/null || echo "")

    if [ -n "$TASK_ID" ]; then
        do_put \
            "PUT /chronic/:elderly_id/daily-tasks/:task_id — mark task complete" \
            "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/daily-tasks/${TASK_ID}" \
            '{}' \
            200
    else
        log_skip "PUT /daily-tasks/:task_id — no task ID found, skipping update"
    fi
    echo ""

    # ── 16. POST /chronic/:elderly_id/report/generate ────────────────────────
    do_post \
        "POST /chronic/:elderly_id/report/generate — generate weekly report" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/report/generate" \
        '{"report_type": "weekly"}' \
        200
    echo ""

    # ── 17. GET /chronic/:elderly_id/report/weekly ───────────────────────────
    do_get \
        "GET /chronic/:elderly_id/report/weekly — get weekly report" \
        "${BASE_URL}/api/v1/chronic/${ELDERLY_ID}/report/weekly" \
        200
    echo ""

    # ── Summary ───────────────────────────────────────────────────────────────
    log_header "Test Summary"
    echo -e "  ${GREEN}Passed : ${PASS}${NC}"
    echo -e "  ${RED}Failed : ${FAIL}${NC}"
    echo -e "  ${YELLOW}Skipped: ${SKIP}${NC}"
    echo -e "  Total  : ${TOTAL}"
    echo ""

    if [ "$FAIL" -gt 0 ]; then
        echo -e "  ${RED}Some tests failed.${NC}"
        exit 1
    else
        echo -e "  ${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

main "$@"

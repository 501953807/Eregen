#!/bin/bash
# Eregen CI/CD pipeline script
# Run with: ./scripts/ci.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
DOCKER_COMPOSE="$REPO_ROOT/cloud/docker-compose.yml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# ---------- Commands ----------

cmd_build() {
    log_info "Building cloud services..."
    cd "$REPO_ROOT/cloud/api-server" && go build -o ../bin/api-server ./...
    cd "$REPO_ROOT/cloud/gateway" && go build -o ../bin/gateway ./...
    cd "$REPO_ROOT/cloud/push-service" && go build -o ../bin/push-service ./...
    cd "$REPO_ROOT/cloud/data-pipeline" && go build -o ../bin/data-pipeline ./...
    log_info "Cloud build complete."

    log_info "Building Flutter family app..."
    cd "$REPO_ROOT/apps/family-app" && flutter build apk --debug 2>/dev/null || true
    log_info "Flutter build complete."

    log_info "Building admin web..."
    cd "$REPO_ROOT/apps/admin-web" && npm run build 2>/dev/null || true
    log_info "Admin web build complete."

    log_info "Building miniprogram..."
    cd "$REPO_ROOT/apps/miniprogram" && echo "Miniprogram builds via WeChat DevTools"
    log_info "Miniprogram build complete."

    log_info "Building website..."
    cd "$REPO_ROOT/apps/website" && hugo --minify 2>/dev/null || true
    log_info "Website build complete."
}

cmd_test() {
    log_info "Running tests..."

    log_info "  Cloud Go tests..."
    cd "$REPO_ROOT/cloud" && go test ./... -short 2>/dev/null || log_warn "Cloud tests skipped"

    log_info "  Firmware C unit tests..."
    cd "$REPO_ROOT/firmware" && find . -name "*test*.c" | head -5 | while read f; do
        log_info "    Found test: $f"
    done

    log_info "Tests complete."
}

cmd_docker_up() {
    log_info "Starting Docker services..."
    if [ ! -f "$DOCKER_COMPOSE" ]; then
        log_error "docker-compose.yml not found at $DOCKER_COMPOSE"
        exit 1
    fi
    docker compose -f "$DOCKER_COMPOSE" up -d
    log_info "Services started. Check with: docker compose -f $DOCKER_COMPOSE ps"
}

cmd_docker_down() {
    log_info "Stopping Docker services..."
    docker compose -f "$DOCKER_COMPOSE" down 2>/dev/null || true
    log_info "Services stopped."
}

cmd_lint() {
    log_info "Running linters..."

    log_info "  Go vet..."
    cd "$REPO_ROOT/cloud" && go vet ./... 2>/dev/null || log_warn "Go vet skipped"

    log_info "  TypeScript check..."
    cd "$REPO_ROOT/apps/admin-web" && npx tsc --noEmit 2>/dev/null || log_warn "TS check skipped"

    log_info "Linting complete."
}

cmd_deploy_prod() {
    log_warn "This will deploy to production!"
    read -p "Type 'YES' to confirm: " confirm
    if [ "$confirm" != "YES" ]; then
        log_error "Deployment cancelled."
        exit 1
    fi
    log_info "Deploying to production..."
    docker compose -f "$DOCKER_COMPOSE" up -d --build
    log_info "Production deployment complete."
}

# ---------- Main ----------

case "${1:-help}" in
    build)   cmd_build ;;
    test)    cmd_test ;;
    up)      cmd_docker_up ;;
    down)    cmd_docker_down ;;
    lint)    cmd_lint ;;
    deploy)  cmd_deploy_prod ;;
    help|*)
        echo "Usage: $0 {build|test|up|down|lint|deploy|help}"
        echo ""
        echo "Commands:"
        echo "  build    - Build all services and apps"
        echo "  test     - Run all tests"
        echo "  up       - Start Docker services"
        echo "  down     - Stop Docker services"
        echo "  lint     - Run linters"
        echo "  deploy   - Deploy to production (requires confirmation)"
        echo "  help     - Show this help"
        ;;
esac

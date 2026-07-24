# ---------------------------------------------------------------------------
# Load common library — resolve relative to THIS file's directory
# ---------------------------------------------------------------------------
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
source "$_LIB_DIR/common.sh"
unset _LIB_DIR

docker_compose_up() {
  local group="${1:-all}"
  local compose_file="$PROJECT_ROOT/docker-compose.yml"

  if [ ! -f "$compose_file" ]; then
    log_error "docker-compose.yml not found at $compose_file"
    return 1
  fi

  if ! command -v docker &>/dev/null; then
    log_error "Docker not found. Install from https://docker.com"
    return 1
  fi

  if ! docker compose version &>/dev/null 2>&1; then
    log_error "'docker compose' (v2) not found"
    return 1
  fi

  log_info "Starting Docker services (group: $group)..."
  cd "$PROJECT_ROOT"

  case "$group" in
    all)
      docker compose -f docker-compose.yml up -d
      ;;
    cloud)
      docker compose -f docker-compose.yml up -d postgres redis influxdb emqx nats
      docker compose -f docker-compose.yml up -d gateway api-server push-service data-pipeline admin-api
      ;;
    b2b)
      docker compose -f docker-compose.yml up -d postgres
      docker compose -f docker-compose.yml up -d hospital-api community-platform insurance-integration
      ;;
    monitoring)
      docker compose -f docker-compose.yml up -d prometheus grafana loki
      ;;
    frontend)
      docker compose -f docker-compose.yml up -d admin-web website
      ;;
    *)
      log_error "Unknown group: $group"
      log_info "Groups: all, cloud, b2b, monitoring, frontend"
      return 1
      ;;
  esac

  log_success "Docker services started"
  log_info "Check status: docker compose -f $compose_file ps"
}

docker_compose_down() {
  local group="${1:-all}"
  cd "$PROJECT_ROOT"
  docker compose -f docker-compose.yml down
  log_info "Docker services stopped"
}

docker_compose_logs() {
  local service="${1:-}"
  cd "$PROJECT_ROOT"
  if [ -n "$service" ]; then
    docker compose -f docker-compose.yml logs -f "$service"
  else
    docker compose -f docker-compose.yml logs -f
  fi
}

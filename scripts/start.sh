#!/usr/bin/env bash
set -euo pipefail
# Eregen Platform - Main Startup Script
# Ties all lib modules together. POSIX-compatible (bash 3.2 on macOS).
# Usage: ./scripts/start.sh <command> [service] [options]

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/common.sh"

DOCKER_MODE=false
EXTRA_PORT=""
COMMAND=""
SERVICE=""

# Parse arguments
while [ $# -gt 0 ]; do
  case "$1" in
    --docker) DOCKER_MODE=true; shift ;;
    --port) EXTRA_PORT="$2"; shift 2 ;;
    start|stop|restart|status|logs|clean|check-deps|ports-check) COMMAND="$1"; shift ;;
    *) SERVICE="$1"; shift ;;
  esac
done

# If no command, show usage
if [ -z "$COMMAND" ]; then
  echo "Eregen Platform - Startup Script"
  echo ""
  echo "Usage: $0 <command> [service] [options]"
  echo ""
  echo "Commands:"
  echo "  start <service|--all|--group>  Start a service or group"
  echo "  stop <service|--all|--group>   Stop a service or group"
  echo "  restart <service>              Restart a service"
  echo "  status [--all]                 Show service status"
  echo "  logs <service|--all>           View logs"
  echo "  clean                          Remove PID/lock files"
  echo "  check-deps                     Check required dependencies"
  echo "  ports-check                    Check for port conflicts"
  echo ""
  echo "Options:"
  echo "  --docker    Use Docker compose instead of native"
  echo "  --port X    Override service port"
  echo ""
  list_available_services
  exit 0
fi

# Load environment
load_env

case "$COMMAND" in
  start)
    if [ "$DOCKER_MODE" = true ]; then
      source "$SCRIPT_DIR/lib/docker-compose.sh"
      local_group="${SERVICE:-all}"
      [ "$local_group" = "--all" ] || [ -z "$local_group" ] && local_group="all"
      docker_compose_up "$local_group"
      exit 0
    fi

    # Port conflict check before starting
    check_ports_conflict || {
      log_error "Fix port conflicts before starting services"
      exit 1
    }

    # Load all libraries
    source "$SCRIPT_DIR/lib/go-service.sh"
    source "$SCRIPT_DIR/lib/flutter-app.sh"
    source "$SCRIPT_DIR/lib/vue-app.sh"
    source "$SCRIPT_DIR/lib/hugo-site.sh"
    source "$SCRIPT_DIR/lib/firmware.sh"

    case "$SERVICE" in
      --all|"")
        log_header "Starting All Services"
        log_info "Cloud backend..."
        for svc in api-server push-service data-pipeline admin-api gateway; do
          go_start "$svc" "$EXTRA_PORT" || log_warn "Failed to start $svc"
        done
        log_info "B2B services..."
        for svc in hospital-api community-platform insurance-integration; do
          go_start "$svc" "$EXTRA_PORT" || log_warn "Failed to start $svc"
        done
        log_info "Frontend apps..."
        vue_start admin-web "$EXTRA_PORT" || log_warn "Failed to start admin-web"
        hugo_start website "$EXTRA_PORT" || log_warn "Failed to start website"
        flutter_start family-app chrome || log_warn "Failed to start family-app"
        log_success "All services started (check individual logs for errors)"
        ;;
      cloud)
        for svc in api-server push-service data-pipeline admin-api gateway; do
          go_start "$svc" "$EXTRA_PORT"
        done
        ;;
      b2b)
        for svc in hospital-api community-platform insurance-integration; do
          go_start "$svc" "$EXTRA_PORT"
        done
        ;;
      apps)
        vue_start admin-web "$EXTRA_PORT"
        hugo_start website "$EXTRA_PORT"
        flutter_start family-app chrome
        ;;
      firmware)
        firmware_list_targets
        ;;
      api-server|push-service|data-pipeline|admin-api|gateway)
        go_start "$SERVICE" "$EXTRA_PORT"
        ;;
      hospital-api|community-platform|insurance-integration)
        go_start "$SERVICE" "$EXTRA_PORT"
        ;;
      family-app)
        flutter_start "$SERVICE" "$EXTRA_PORT"
        ;;
      admin-web)
        vue_start "$SERVICE" "$EXTRA_PORT"
        ;;
      website)
        hugo_start "$SERVICE" "$EXTRA_PORT"
        ;;
      bracelet|pillbox|medical-wristband)
        firmware_build "$SERVICE" "$EXTRA_PORT"
        ;;
      *)
        log_error "Unknown service: $SERVICE"
        list_available_services
        exit 1
        ;;
    esac
    ;;

  stop)
    if [ "$DOCKER_MODE" = true ]; then
      source "$SCRIPT_DIR/lib/docker-compose.sh"
      local_group="${SERVICE:-all}"
      [ "$local_group" = "--all" ] || [ -z "$local_group" ] && local_group="all"
      docker_compose_down "$local_group"
      exit 0
    fi

    source "$SCRIPT_DIR/lib/go-service.sh"
    source "$SCRIPT_DIR/lib/flutter-app.sh"
    source "$SCRIPT_DIR/lib/vue-app.sh"
    source "$SCRIPT_DIR/lib/hugo-site.sh"

    case "$SERVICE" in
      --all|"")
        for svc in api-server push-service data-pipeline admin-api gateway; do
          go_stop "$svc"
        done
        for svc in hospital-api community-platform insurance-integration; do
          go_stop "$svc"
        done
        vue_stop admin-web
        hugo_stop website
        flutter_stop family-app
        ;;
      cloud)
        for svc in api-server push-service data-pipeline admin-api gateway; do
          go_stop "$svc"
        done
        ;;
      b2b)
        for svc in hospital-api community-platform insurance-integration; do
          go_stop "$svc"
        done
        ;;
      apps)
        vue_stop admin-web
        hugo_stop website
        flutter_stop family-app
        ;;
      api-server|push-service|data-pipeline|admin-api|gateway)
        go_stop "$SERVICE"
        ;;
      hospital-api|community-platform|insurance-integration)
        go_stop "$SERVICE"
        ;;
      family-app)
        flutter_stop "$SERVICE"
        ;;
      admin-web)
        vue_stop "$SERVICE"
        ;;
      website)
        hugo_stop "$SERVICE"
        ;;
      *)
        log_error "Unknown service: $SERVICE"
        exit 1
        ;;
    esac
    ;;

  restart)
    if [ -z "$SERVICE" ]; then
      log_error "Usage: $0 restart <service>"
      exit 1
    fi
    "$SCRIPT_DIR/start.sh" stop "$SERVICE"
    sleep 1
    "$SCRIPT_DIR/start.sh" start "$SERVICE"
    ;;

  status)
    source "$SCRIPT_DIR/lib/go-service.sh"
    source "$SCRIPT_DIR/lib/flutter-app.sh"
    source "$SCRIPT_DIR/lib/vue-app.sh"
    source "$SCRIPT_DIR/lib/hugo-site.sh"

    if [ "$SERVICE" = "--all" ] || [ -z "$SERVICE" ]; then
      log_header "Service Status"
      for svc in api-server push-service data-pipeline admin-api gateway; do
        go_status "$svc" 2>/dev/null || true
      done
      for svc in hospital-api community-platform insurance-integration; do
        go_status "$svc" 2>/dev/null || true
      done
      vue_status admin-web 2>/dev/null || true
      hugo_status website 2>/dev/null || true
      flutter_status family-app 2>/dev/null || true
    else
      go_status "$SERVICE" 2>/dev/null || \
      vue_status "$SERVICE" 2>/dev/null || \
      hugo_status "$SERVICE" 2>/dev/null || \
      flutter_status "$SERVICE" 2>/dev/null || \
      log_error "Unknown service: $SERVICE"
    fi
    ;;

  logs)
    if [ "$DOCKER_MODE" = true ]; then
      source "$SCRIPT_DIR/lib/docker-compose.sh"
      local_svc="$SERVICE"
      [ "$local_svc" = "--all" ] || [ -z "$local_svc" ] && local_svc=""
      docker_compose_logs "$local_svc"
      exit 0
    fi

    if [ "$SERVICE" = "--all" ] || [ -z "$SERVICE" ]; then
      log_info "Showing all logs (Ctrl+C to stop):"
      tail -f "$PID_DIR/"*.log 2>/dev/null | sed "s/^/[all] /"
    else
      local_logfile="$PID_DIR/${SERVICE}.log"
      if [ -f "$local_logfile" ]; then
        tail -f "$local_logfile"
      else
        log_error "No log file found for $SERVICE"
        log_info "Start the service first: $0 start $SERVICE"
      fi
    fi
    ;;

  clean)
    log_info "Cleaning runtime files..."
    if [ -d "$PID_DIR" ]; then
      find "$PID_DIR" -name "*.pid" -delete 2>/dev/null || true
      find "$PID_DIR" -name "*.timestamp" -delete 2>/dev/null || true
      log_info "Removed PID files from $PID_DIR"
    else
      log_info "No PID directory found (nothing to clean)"
    fi
    find "$PROJECT_ROOT" -name "*.lock" -not -path "*/node_modules/*" -not -path "*/.git/*" -mmin +60 -delete 2>/dev/null || true
    log_success "Clean complete"
    ;;

  check-deps)
    log_header "Dependency Check"
    all_ok=true

    # Go
    if command -v go &>/dev/null; then
      go_ver=$(go version 2>/dev/null | sed -n 's/.*go\([0-9][0-9]*\.[0-9][0-9]*\).*/\1/p' | head -1)
      if [ "$(echo "$go_ver >= 1.22" | bc -l 2>/dev/null || echo 0)" = "1" ]; then
        log_success "Go $go_ver OK"
      else
        log_warn "Go $go_ver found (>= 1.22 recommended)"
      fi
    else
      log_error "Go not found -- install from https://golang.org"
      all_ok=false
    fi

    # Node.js
    if command -v node &>/dev/null; then
      node_ver=$(node --version 2>/dev/null | sed -n 's/v\([0-9][0-9]*\).*/\1/p')
      if [ "$node_ver" -ge 18 ] 2>/dev/null; then
        log_success "Node.js $(node --version) OK"
      else
        log_error "Node.js $node_ver found (>= 18 required)"
        all_ok=false
      fi
    else
      log_error "Node.js not found -- install from https://nodejs.org"
      all_ok=false
    fi

    # npm
    if command -v npm &>/dev/null; then
      log_success "npm $(npm --version) OK"
    else
      log_error "npm not found"
      all_ok=false
    fi

    # Flutter
    if command -v flutter &>/dev/null; then
      log_success "Flutter found OK"
    else
      log_warn "Flutter not found -- install from https://flutter.dev"
    fi

    # Hugo
    if command -v hugo &>/dev/null; then
      log_success "Hugo found OK"
    else
      log_warn "Hugo not found -- install from https://gohugo.io"
    fi

    # ESP-IDF
    if command -v idf.py &>/dev/null; then
      log_success "ESP-IDF (idf.py) OK"
    else
      log_warn "ESP-IDF not in PATH -- set IDF_PATH for firmware builds"
    fi

    # Arm GNU Toolchain
    if command -v arm-none-eabi-gcc &>/dev/null; then
      log_success "Arm GNU Toolchain OK"
    else
      log_warn "Arm GNU Toolchain not found -- install for bracelet firmware"
    fi

    # Docker
    if command -v docker &>/dev/null && docker compose version &>/dev/null 2>&1; then
      log_success "Docker Compose OK"
    else
      log_warn "Docker not found -- docker compose mode unavailable"
    fi

    if [ "$all_ok" = true ]; then
      log_success "All critical dependencies OK"
    else
      log_error "Some critical dependencies are missing"
    fi
    ;;

  ports-check)
    log_header "Port Conflict Check"
    check_ports_conflict
    ;;

  *)
    log_error "Unknown command: $COMMAND"
    list_available_services
    exit 1
    ;;
esac

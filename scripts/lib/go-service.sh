#!/usr/bin/env bash
# Eregen Platform - Go Microservice Library
# Manages all Go microservices: start, stop, status.
# Copyright (c) 2026 Eregen (颐贞). All rights reserved.
# Fully compatible with bash 3.2 (macOS default).

# ---------------------------------------------------------------------------
# Load common library — resolve relative to THIS file's directory
# ---------------------------------------------------------------------------
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
source "$_LIB_DIR/common.sh"
unset _LIB_DIR

# ---------------------------------------------------------------------------
# Service configuration — flat string (bash 3.2 compatible, no associative arrays).
# Format: service:dir:port_var:default_port|...
# Gateway is MQTT-only and has no HTTP port (default_port=0).
# ---------------------------------------------------------------------------
GO_SERVICES="api-server|push-service|data-pipeline|admin-api|hospital-api|community-platform|insurance-integration|gateway"

GO_CONFIG="api-server:cloud/api-server:PORT_API_SERVER:8080|\
push-service:cloud/push-service:PORT_PUSH_SERVICE:8085|\
data-pipeline:cloud/data-pipeline:PORT_DATA_PIPELINE:8087|\
admin-api:cloud/admin-api:PORT_ADMIN_API:8089|\
hospital-api:b2b/hospital-api:PORT_HOSPITAL_API:8082|\
community-platform:b2b/community-platform:PORT_COMMUNITY_PLATFORM:8083|\
insurance-integration:b2b/insurance-integration:PORT_INSURANCE_INTEGRATION:8084|\
gateway:cloud/gateway::0"

# ---------------------------------------------------------------------------
# Helpers — lookup functions replacing associative-array indexing.
# ---------------------------------------------------------------------------

_go_dir() {
  echo "$GO_CONFIG" | tr '|' '\n' | grep "^${1}:" | head -1 | cut -d: -f2
}

_go_port_var() {
  echo "$GO_CONFIG" | tr '|' '\n' | grep "^${1}:" | head -1 | cut -d: -f3
}

_go_default_port() {
  echo "$GO_CONFIG" | tr '|' '\n' | grep "^${1}:" | head -1 | cut -d: -f4
}

_go_validate() {
  echo "|${GO_SERVICES}|" | grep -q "|${1}|"
}

# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------

_resolve_port() {
  local service="$1"
  local extra_port="${2:-}"
  local port_var
  port_var=$(_go_port_var "$service")
  local default_port
  default_port=$(_go_default_port "$service")

  # 1. Explicit extra_port argument takes highest priority
  if [ -n "$extra_port" ]; then
    echo "$extra_port"
    return
  fi

  # 2. Shell environment variable (exported before calling this script)
  if [ -n "$port_var" ] && [ -n "${!port_var:-}" ]; then
    echo "${!port_var}"
    return
  fi

  # 3. Value from .env file (already sourced by load_env)
  if [ -n "$port_var" ] && [ -n "${!port_var:-}" ]; then
    echo "${!port_var}"
    return
  fi

  # 4. Hardcoded default
  echo "$default_port"
}

_build_extra_env() {
  # Emit additional KEY=VALUE pairs for the given service.
  local service="$1"
  case "$service" in
    api-server)
      echo "JWT_SECRET=${JWT_SECRET:-changeme}"
      echo "DB_URL=${DB_URL:-postgres://eregen:eregen@localhost:5432/eregen?sslmode=disable}"
      ;;
    push-service)
      echo "PUSH_SERVICE_PORT=${PUSH_SERVICE_PORT:-8085}"
      ;;
    data-pipeline)
      echo "PIPELINE_PORT=${PIPELINE_PORT:-8087}"
      ;;
    admin-api)
      echo "JWT_SECRET=${JWT_SECRET:-changeme}"
      echo "SQLITE_PATH=${SQLITE_PATH:-$PROJECT_ROOT/data/admin.db}"
      ;;
    hospital-api)
      echo "DATABASE_URL=${DATABASE_URL:-postgres://eregen:eregen@localhost:5432/hospital?sslmode=disable}"
      ;;
    community-platform)
      echo "DATABASE_URL=${DATABASE_URL:-postgres://eregen:eregen@localhost:5432/community?sslmode=disable}"
      ;;
    insurance-integration)
      echo "DATABASE_URL=${DATABASE_URL:-postgres://eregen:eregen@localhost:5432/insurance?sslmode=disable}"
      ;;
    gateway)
      echo "MQTT_BROKER=${MQTT_BROKER:-localhost:1883}"
      echo "NATS_URL=${NATS_URL:-nats://localhost:4222}"
      ;;
  esac
}

go_start() {
  local service="$1"
  local extra_port="${2:-}"

  if ! _go_validate "$service"; then
    return 1
  fi
  ensure_pid_dir

  # Check if already running
  local existing_pid
  if existing_pid=$(read_pid "$service"); then
    log_warn "$service is already running (PID $existing_pid)"
    return 0
  fi

  # Resolve port
  local port
  port=$(_resolve_port "$service" "$extra_port")

  # Gateway is MQTT-only — no HTTP port to wait on
  if [ "$service" = "gateway" ]; then
    log_warn "Gateway is MQTT-only (no HTTP port). Starting without port check."
  elif [ "$port" -eq 0 ] 2>/dev/null; then
    log_warn "Service '$service' has no HTTP port configured (port=0)."
  fi

  # Build the command
  local svc_dir="$PROJECT_ROOT/$(_go_dir "$service")"
  if [ ! -d "$svc_dir" ]; then
    log_error "Directory not found: $svc_dir"
    return 1
  fi

  local log_file="$PID_DIR/${service}.log"
  local env_lines
  env_lines=$(_build_extra_env "$service")

  # Launch: cd into service dir, set PORT + extra env vars, run go, capture output
  (
    cd "$svc_dir"
    export PORT="$port"
    if [ -n "$env_lines" ]; then
      eval "$env_lines"
    fi
    exec go run ./cmd 2>&1
  ) > "$log_file" 2>&1 &
  local pid=$!

  write_pid "$service" "$pid"
  log_info "Started $service (PID $pid) on port $port"

  # Wait for the port to become available (skip for gateway / port=0)
  if [ "$service" != "gateway" ] && [ "$port" -gt 0 ] 2>/dev/null; then
    log_info "Waiting for $service to be ready on port $port..."
    local waited=0
    while [ $waited -lt 30 ]; do
      if check_ports_conflict >/dev/null 2>&1 && check_process_running "$port"; then
        log_success "$service is ready on port $port (PID $pid)"
        log_info "Log: $log_file"
        return 0
      fi
      sleep 1
      waited=$((waited + 1))
    done

    # Last check — maybe it started but lsof/ss wasn't fast enough
    if kill -0 "$pid" 2>/dev/null; then
      log_warn "$service may still be starting (PID $pid alive, log at $log_file)"
      return 0
    fi

    log_error "$service failed to start on port $port within 30s"
    log_info "Check log: $log_file"
    return 1
  else
    # Non-HTTP or no-port service — just confirm process is alive
    sleep 2
    if kill -0 "$pid" 2>/dev/null; then
      log_success "$service is running (PID $pid), log: $log_file"
      return 0
    else
      log_error "$service failed to start (check $log_file)"
      rm -f "$PID_DIR/${service}.pid" "$PID_DIR/${service}.timestamp"
      return 1
    fi
  fi
}

go_stop() {
  local service="$1"

  if ! _go_validate "$service"; then
    return 1
  fi

  kill_pid "$service"
  log_success "$service stopped"
}

go_status() {
  local service="$1"

  if ! _go_validate "$service"; then
    return 1
  fi

  local port
  port=$(_go_default_port "$service")
  local port_var
  port_var=$(_go_port_var "$service")

  # Try to resolve actual port
  if [ -n "$port_var" ] && [ -n "${!port_var:-}" ]; then
    port="${!port_var}"
  fi

  local pid
  if pid=$(read_pid "$service"); then
    if [ "$service" = "gateway" ] || [ "$port" -eq 0 ] 2>/dev/null; then
      echo -e "${GREEN}$service${NC} is running (PID $pid) — MQTT only"
    else
      echo -e "${GREEN}$service${NC} is running (PID $pid) on port $port"
    fi
    return 0
  fi

  echo -e "${RED}$service${NC} is not running"
  return 1
}

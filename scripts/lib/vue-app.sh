#!/usr/bin/env bash
# Eregen Platform - Vue App Manager
# Manages Vue frontend apps (admin-web, family-app, etc.)
# © 2026 Eregen (颐贞). All rights reserved.
# Fully compatible with bash 3.2 (macOS default).

# ---------------------------------------------------------------------------
# Load common library — resolve relative to THIS file's directory
# ---------------------------------------------------------------------------
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
source "$_LIB_DIR/common.sh"
unset _LIB_DIR

# --- App Registry (bash 3.2 compatible — no associative arrays) ---
VUE_APPS_LIST="admin-web|family-app"

_vue_app_dir() {
  case "$1" in
    admin-web) echo "apps/admin-web" ;;
    family-app) echo "apps/family-app" ;;
  esac
}

# --- Start ---
vue_start() {
  local app_name="${1:-}"
  if [ -z "$app_name" ]; then
    log_error "Usage: vue_start <app_name> [extra_port]"
    return 1
  fi

  # Validate app exists
  if ! echo "$VUE_APPS_LIST" | grep -qw "$app_name"; then
    log_error "Unknown Vue app: $app_name"
    log_info "Available: $(echo "$VUE_APPS_LIST" | tr '|' ' ')"
    return 1
  fi

  local app_dir
  app_dir=$(_vue_app_dir "$app_name")
  local full_path="$PROJECT_ROOT/$app_dir"

  # Check if already running
  local existing_pid
  if existing_pid=$(read_pid "$app_name"); then
    log_warn "$app_name is already running (PID $existing_pid)"
    return 0
  fi

  # Check npm is installed
  if ! command -v npm &>/dev/null; then
    log_error "npm is not installed. Install Node.js first."
    return 1
  fi

  # Determine port
  local port=""
  if [ -n "${2:-}" ]; then
    port="$2"
  else
    port=$(get_port "ADMIN_WEB" "3001")
  fi

  # Auto-install node_modules if missing
  if [ ! -d "$full_path/node_modules" ]; then
    log_info "Installing dependencies for $app_name..."
    (cd "$full_path" && npm install)
  fi

  # Ensure PID directory exists
  ensure_pid_dir

  # Start the dev server (Vite uses --port, not PORT env var)
  log_info "Starting $app_name on port $port..."
  (cd "$full_path" && npx vite --host 0.0.0.0 --port "$port" > "$PID_DIR/${app_name}.log" 2>&1 &)
  local pid=$!
  write_pid "$app_name" "$pid"

  # Wait for port to become available (up to 30s)
  local waited=0
  while [ $waited -lt 30 ]; do
    if check_process_running "$port"; then
      log_success "$app_name started on http://localhost:$port (PID $pid)"
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
  done

  log_error "$app_name failed to start within 30s. Check $PID_DIR/${app_name}.log"
  return 1
}

# --- Stop ---
vue_stop() {
  local app_name="${1:-}"
  if [ -z "$app_name" ]; then
    log_error "Usage: vue_stop <app_name>"
    return 1
  fi

  if ! echo "$VUE_APPS_LIST" | grep -qw "$app_name"; then
    log_error "Unknown Vue app: $app_name"
    return 1
  fi

  kill_pid "$app_name"
}

# --- Status ---
vue_status() {
  local app_name="${1:-}"
  if [ -z "$app_name" ]; then
    log_error "Usage: vue_status <app_name>"
    return 1
  fi

  if ! echo "$VUE_APPS_LIST" | grep -qw "$app_name"; then
    log_error "Unknown Vue app: $app_name"
    return 1
  fi

  local app_dir
  app_dir=$(_vue_app_dir "$app_name")
  local port=""
  case "$app_name" in
    admin-web)
      port=$(get_port "ADMIN_WEB" "3001")
      ;;
    family-app)
      port=$(get_port "FAMILY_APP" "5173")
      ;;
  esac

  local pid
  if pid=$(read_pid "$app_name"); then
    if check_process_running "$port"; then
      log_success "$app_name is running (PID $pid) — http://localhost:$port"
    else
      log_warn "$app_name has stale PID ($pid) but process not listening on port $port"
    fi
  else
    log_info "$app_name is not running"
  fi
}

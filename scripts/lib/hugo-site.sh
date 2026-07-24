#!/usr/bin/env bash
# Eregen Platform - Hugo Site Manager
# Manages Hugo static sites (website, etc.)
# © 2026 Eregen (颐贞). All rights reserved.
# Fully compatible with bash 3.2 (macOS default).

# ---------------------------------------------------------------------------
# Load common library — resolve relative to THIS file's directory
# ---------------------------------------------------------------------------
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
source "$_LIB_DIR/common.sh"
unset _LIB_DIR

# --- Site Registry (bash 3.2 compatible — no associative arrays) ---
HUGO_SITES_LIST="website"

_hugo_site_dir() {
  case "$1" in
    website) echo "apps/website" ;;
  esac
}

# --- Start ---
hugo_start() {
  local site_name="${1:-}"
  if [ -z "$site_name" ]; then
    log_error "Usage: hugo_start <site_name> [extra_port]"
    return 1
  fi

  # Validate site exists
  if ! echo "$HUGO_SITES_LIST" | grep -qw "$site_name"; then
    log_error "Unknown Hugo site: $site_name"
    log_info "Available: $HUGO_SITES_LIST"
    return 1
  fi

  local site_dir
  site_dir=$(_hugo_site_dir "$site_name")
  local full_path="$PROJECT_ROOT/$site_dir"

  # Check if already running
  local existing_pid
  if existing_pid=$(read_pid "$site_name"); then
    log_warn "$site_name is already running (PID $existing_pid)"
    return 0
  fi

  # Check hugo is installed
  if ! command -v hugo &>/dev/null; then
    log_error "Hugo is not installed. Install it first:"
    log_error "  brew install hugo  (macOS)"
    log_error "  or https://gohugo.io/getting-started/installing/"
    return 1
  fi

  # Determine port
  local port=""
  if [ -n "${2:-}" ]; then
    port="$2"
  else
    port=$(get_port "WEBSITE" "1313")
  fi

  # Ensure PID directory exists
  ensure_pid_dir

  # Start Hugo server
  log_info "Starting $site_name on port $port..."
  (cd "$full_path" && hugo server --bind 0.0.0.0 --port "$port" > "$PID_DIR/${site_name}.log" 2>&1 &)
  local pid=$!
  write_pid "$site_name" "$pid"

  # Wait for port to become available (up to 30s)
  local waited=0
  while [ $waited -lt 30 ]; do
    if check_process_running "$port"; then
      log_success "$site_name started on http://localhost:$port (PID $pid)"
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
  done

  log_error "$site_name failed to start within 30s. Check $PID_DIR/${site_name}.log"
  return 1
}

# --- Stop ---
hugo_stop() {
  local site_name="${1:-}"
  if [ -z "$site_name" ]; then
    log_error "Usage: hugo_stop <site_name>"
    return 1
  fi

  if ! echo "$HUGO_SITES_LIST" | grep -qw "$site_name"; then
    log_error "Unknown Hugo site: $site_name"
    return 1
  fi

  kill_pid "$site_name"
}

# --- Status ---
hugo_status() {
  local site_name="${1:-}"
  if [ -z "$site_name" ]; then
    log_error "Usage: hugo_status <site_name>"
    return 1
  fi

  if ! echo "$HUGO_SITES_LIST" | grep -qw "$site_name"; then
    log_error "Unknown Hugo site: $site_name"
    return 1
  fi

  local port=""
  case "$site_name" in
    website)
      port=$(get_port "WEBSITE" "1313")
      ;;
  esac

  local pid
  if pid=$(read_pid "$site_name"); then
    if check_process_running "$port"; then
      log_success "$site_name is running (PID $pid) — http://localhost:$port"
    else
      log_warn "$site_name has stale PID ($pid) but process not listening on port $port"
    fi
  else
    log_info "$site_name is not running"
  fi
}

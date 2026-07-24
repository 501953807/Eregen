#!/usr/bin/env bash
# Eregen Platform - Common Library
# Shared functions for all startup scripts.
# © 2026 Eregen (颐贞). All rights reserved.
# Fully compatible with bash 3.2 (macOS default).

set -euo pipefail

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log_info()    { echo -e "${GREEN}[INFO]${NC} $*"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $*"; }
log_success() { echo -e "${CYAN}[OK]${NC} $*"; }
log_header()  { echo -e "\n${BOLD}=== $* ===${NC}\n"; }

# --- Environment ---
: "${PROJECT_ROOT:=}"
load_env() {
  local search_dir="$PWD"
  while [ "$search_dir" != "/" ]; do
    if [ -f "$search_dir/.env" ]; then
      PROJECT_ROOT="$search_dir"
      set -a; source "$search_dir/.env"; set +a
      log_info "Loaded .env from $PROJECT_ROOT"
      return 0
    fi
    search_dir="$(dirname "$search_dir")"
  done
  log_warn "No .env found — using hardcoded defaults"
  PROJECT_ROOT="${PWD}"
  return 0
}

get_port() {
  local service="$1"
  local default="${2:-}"
  local upper
  upper=$(echo "$service" | tr '[:lower:]' '[:upper:]')
  local var_name="PORT_${upper}"
  # bash 3.2 compatible indirect expansion (no ${(P)var_name} which is zsh-only)
  local val="${!var_name:-}"
  if [ -n "$val" ]; then
    echo "$val"
  else
    echo "$default"
  fi
}

# --- Port Conflict Detection ---
check_ports_conflict() {
  # Collect all PORT_* values from environment and .env file.
  # bash 3.2 does not support local -a, so we use a temp file approach.
  local port_file
  port_file=$(mktemp)
  trap "rm -f '$port_file'" RETURN

  # Check .env file directly
  if [ -n "$PROJECT_ROOT" ] && [ -f "$PROJECT_ROOT/.env" ]; then
    while IFS='=' read -r key value || [ -n "$key" ]; do
      key=$(echo "$key" | tr -d ' ')
      value=$(echo "$value" | tr -d ' ' | tr -d '"')
      case "$key" in
        PORT_[A-Z_]*)
          # Validate value is numeric
          case "$value" in
            *[!0-9]*) ;; # skip non-numeric
            *) echo "${value}=${key}" >> "$port_file" ;;
          esac
          ;;
      esac
    done < "$PROJECT_ROOT/.env"
  fi

  # Check for duplicates by comparing each pair
  local i j
  local total_lines
  total_lines=$(wc -l < "$port_file")
  if [ "$total_lines" -lt 2 ]; then
    log_success "No port conflicts detected"
    return 0
  fi

  local line_num=0
  while [ $line_num -lt "$total_lines" ]; do
    local port_i svc_i
    port_i=$(sed -n "$((line_num + 1))p" "$port_file" | cut -d= -f1)
    svc_i=$(sed -n "$((line_num + 1))p" "$port_file" | cut -d= -f2-)
    local inner=0
    while [ $inner -lt "$total_lines" ]; do
      if [ $inner -ne $line_num ]; then
        local port_j svc_j
        port_j=$(sed -n "$((inner + 1))p" "$port_file" | cut -d= -f1)
        svc_j=$(sed -n "$((inner + 1))p" "$port_file" | cut -d= -f2-)
        if [ "$port_i" = "$port_j" ]; then
          log_error "Port conflict: $svc_i & $svc_j on port $port_i"
          rm -f "$port_file"
          return 1
        fi
      fi
      inner=$((inner + 1))
    done
    line_num=$((line_num + 1))
  done

  log_success "No port conflicts detected"
  return 0
}

# --- PID Management ---
PID_DIR="$HOME/.eregen/pids"

ensure_pid_dir() {
  mkdir -p "$PID_DIR"
}

write_pid() {
  local service="$1"
  local pid="$2"
  ensure_pid_dir
  echo "${pid}" > "$PID_DIR/${service}.pid"
  echo "$(date +%s)" > "$PID_DIR/${service}.timestamp"
}

read_pid() {
  local service="$1"
  local pid_file="$PID_DIR/${service}.pid"
  [ -f "$pid_file" ] || return 1
  local pid
  pid=$(cat "$pid_file")
  if kill -0 "$pid" 2>/dev/null; then
    echo "$pid"
    return 0
  else
    rm -f "$pid_file" "$PID_DIR/${service}.timestamp"
    return 1
  fi
}

kill_pid() {
  local service="$1"
  local pid_file="$PID_DIR/${service}.pid"
  [ -f "$pid_file" ] || return 0
  local pid
  pid=$(cat "$pid_file")
  if kill -0 "$pid" 2>/dev/null; then
    log_info "Stopping $service (PID $pid)..."
    kill "$pid" 2>/dev/null || true
    local waited=0
    while [ $waited -lt 5 ]; do
      if ! kill -0 "$pid" 2>/dev/null; then break; fi
      sleep 1
      waited=$((waited + 1))
    done
    if kill -0 "$pid" 2>/dev/null; then
      kill -9 "$pid" 2>/dev/null || true
    fi
    log_info "$service stopped"
  fi
  rm -f "$pid_file" "$PID_DIR/${service}.timestamp"
}

check_process_running() {
  local port="$1"
  if command -v lsof &>/dev/null; then
    lsof -i :"$port" &>/dev/null && return 0
  elif command -v ss &>/dev/null; then
    ss -tlnp &>/dev/null | grep -q ":$port " && return 0
  elif command -v netstat &>/dev/null; then
    netstat -tlnp &>/dev/null | grep -q ":$port " && return 0
  fi
  return 1
}

# --- Service Discovery ---
list_available_services() {
  log_header "Available Services"
  echo "  Cloud:    api-server push-service data-pipeline admin-api gateway"
  echo "  B2B:      hospital-api community-platform insurance-integration"
  echo "  Apps:     family-app admin-web website miniprogram"
  echo "  Firmware: bracelet pillbox medical-wristband"
  echo ""
  echo "  Groups:   cloud b2b apps firmware all"
}

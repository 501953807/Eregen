#!/usr/bin/env bash
# Flutter apps management library
# Usage: source scripts/lib/flutter-app.sh
#        flutter_start family-app [chrome|ios|android|device:*]
#        flutter_stop family-app
#        flutter_status family-app
# Fully compatible with bash 3.2 (macOS default).

# ---------------------------------------------------------------------------
# Load common library — resolve relative to THIS file's directory
# ---------------------------------------------------------------------------
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
source "$_LIB_DIR/common.sh"
unset _LIB_DIR

# --- App Registry (bash 3.2 compatible — no associative arrays) ---
FLUTTER_APPS_LIST="family-app"

flutter_app_dir() {
  case "$1" in
    family-app) echo "apps/family-app" ;;
  esac
}

flutter_start() {
  local app="$1"
  local target="${2:-chrome}"  # chrome, ios, android, or device ID

  # Validate app name
  if ! echo "$FLUTTER_APPS_LIST" | grep -qw "$app"; then
    log_error "Unknown Flutter app: $app"
    log_info "Available: $FLUTTER_APPS_LIST"
    return 1
  fi

  local app_dir
  app_dir=$(flutter_app_dir "$app")
  local full_dir="$PROJECT_ROOT/$app_dir"
  if [ ! -d "$full_dir" ]; then
    log_error "Directory not found: $full_dir"
    return 1
  fi

  # Check if already running
  if read_pid "$app" >/dev/null 2>&1; then
    log_warn "$app is already running (PID $(read_pid "$app"))"
    return 0
  fi

  # Check Flutter SDK
  if ! command -v flutter &>/dev/null; then
    log_error "Flutter SDK not found. Install from https://flutter.dev"
    return 1
  fi

  log_info "Starting $app ($target)..."
  cd "$full_dir"

  # Build device flag
  local device_flag=""
  case "$target" in
    device:*)
      device_flag="-d ${target#device:}"
      ;;
    *)
      device_flag="-d $target"
      ;;
  esac

  # Build dart-define args from .env
  local dart_defines=""
  if [ -n "${AMAP_KEY:-}" ]; then
    dart_defines+=" --dart-define=AMAP_KEY=$AMAP_KEY"
  fi

  # Run in background
  flutter run $device_flag$dart_defines > "$PID_DIR/${app}.log" 2>&1 &
  local pid=$!
  write_pid "$app" "$pid"

  log_success "$app starting on $target (PID $pid)"
  log_info "Logs: tail -f $PID_DIR/${app}.log"
}

flutter_stop() {
  local app="$1"
  if ! echo "$FLUTTER_APPS_LIST" | grep -qw "$app"; then
    log_error "Unknown Flutter app: $app"
    return 1
  fi
  kill_pid "$app"
}

flutter_status() {
  local app="$1"
  if ! echo "$FLUTTER_APPS_LIST" | grep -qw "$app"; then
    log_error "Unknown Flutter app: $app"
    return 1
  fi
  local pid
  if pid=$(read_pid "$app" 2>/dev/null); then
    log_success "$app is running (PID $pid)"
  else
    log_error "$app is not running"
  fi
}

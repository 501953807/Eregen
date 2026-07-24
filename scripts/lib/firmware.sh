#!/usr/bin/env bash
# Firmware build management library
# Fully compatible with bash 3.2 (macOS default).

# ---------------------------------------------------------------------------
# Load common library — resolve relative to THIS file's directory
# ---------------------------------------------------------------------------
_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
source "$_LIB_DIR/common.sh"
unset _LIB_DIR

# --- Target Registry (bash 3.2 compatible — no associative arrays) ---
FIRMWARE_TARGETS_LIST="bracelet|pillbox|medical-wristband"

_firmware_dir() {
  case "$1" in
    bracelet) echo "firmware/bracelet" ;;
    pillbox) echo "firmware/pillbox" ;;
    medical-wristband) echo "firmware/medical-wristband" ;;
  esac
}

firmware_build() {
  local target="$1"
  local variant="${2:-}"

  if ! echo "$FIRMWARE_TARGETS_LIST" | grep -qw "$target"; then
    log_error "Unknown firmware target: $target"
    log_info "Available: $(echo "$FIRMWARE_TARGETS_LIST" | tr '|' ' ')"
    return 1
  fi

  local full_dir="$PROJECT_ROOT/$(_firmware_dir "$target")"
  if [ ! -d "$full_dir" ]; then
    log_error "Directory not found: $full_dir"
    return 1
  fi

  log_info "Building firmware: $target${variant:+/$variant}..."

  case "$target" in
    bracelet)
      if ! command -v arm-none-eabi-gcc &>/dev/null; then
        log_error "Arm GNU Toolchain not found"
        log_info "Install: brew install arm-none-eabi-gcc"
        return 1
      fi
      cd "$full_dir"
      [ -n "$variant" ] && cd "$variant"
      cmake -B build -DCMAKE_BUILD_TYPE=Release
      cmake --build build -j$(sysctl -n hw.ncpu 2>/dev/null || nproc)
      ;;
    pillbox|medical-wristband)
      if ! command -v idf.py &>/dev/null; then
        log_error "ESP-IDF not found"
        log_info "Set IDF_PATH or install: brew install espressif-tool"
        return 1
      fi
      cd "$full_dir"
      [ "$target" = "pillbox" ] && [ -n "$variant" ] && cd "$variant"
      idf.py build
      ;;
  esac

  log_success "$target built successfully"
  log_info "Flash: cd $full_dir && idf.py flash (ESP32) or openocd + arm-none-eabi-size (GD32)"
}

firmware_clean() {
  local target="$1"
  local variant="${2:-}"
  local dir
  dir=$(_firmware_dir "$target")
  local full_dir="$PROJECT_ROOT/${dir:-$target}"

  case "$target" in
    bracelet)
      cd "$full_dir" && rm -rf build
      ;;
    pillbox|medical-wristband)
      cd "$full_dir" && idf.py fullclean
      ;;
  esac
  log_info "Cleaned $target build artifacts"
}

firmware_list_targets() {
  log_header "Firmware Targets"
  echo "  bracelet    -- GD32E230C8T3, FreeRTOS, CMake"
  echo "              Variants: common, entry, plus, pro"
  echo "  pillbox     -- ESP32-C3, ESP-IDF v5.3, C"
  echo "              Variants: basic, smart, auto"
  echo "  medical-wristband -- ESP32-S3, ESP-IDF v5.3, C"
  echo "              Variants: esp32s3"
}

#!/usr/bin/env bash
# Eregen Platform - Enhanced Dependency Checker

_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
source "$_LIB_DIR/common.sh"
unset _LIB_DIR

compare_versions() {
  local v1="$1"
  local v2="$2"
  set -- ${v1//./ }
  local v1_a=$1 v1_b=$2 v1_c=${3:-0}
  set -- ${v2//./ }
  local v2_a=$1 v2_b=$2 v2_c=${3:-0}
  [ "$v1_a" -lt "$v2_a" ] && { echo -1; return; }
  [ "$v1_a" -gt "$v2_a" ] && { echo 1; return; }
  [ "$v1_b" -lt "$v2_b" ] && { echo -1; return; }
  [ "$v1_b" -gt "$v2_b" ] && { echo 1; return; }
  [ "$v1_c" -lt "$v2_c" ] && { echo -1; return; }
  [ "$v1_c" -gt "$v2_c" ] && { echo 1; return; }
  echo 0
}

check_pass() { log_success "$1"; return 0; }
check_fail() { log_error "$1"; return 1; }

run_all_deps_check() {
  log_header "Dependency Check (Enhanced Mode B)"

  # Go
  if ! command -v go &>/dev/null; then
    check_fail "Go not found -- install from https://golang.org"
    return 1
  fi
  local go_ver=$(go version 2>/dev/null | sed -n 's/.*go\([0-9][0-9]*\.[0-9][0-9]*\).*/\1/p' | head -1)
  if [ -z "$go_ver" ]; then
    log_warn "Go found but version unclear"
  else
    local cmp=$(compare_versions "$go_ver" "1.22")
    [ "$cmp" -lt 0 ] && { check_fail "Go $go_ver (need >= 1.22)"; return 1; }
    check_pass "Go $go_ver OK"
  fi

  # Node
  if ! command -v node &>/dev/null; then
    check_fail "Node.js not found -- install from https://nodejs.org"
    return 1
  fi
  local node_ver=$(node --version 2>/dev/null | sed 's/v//')
  local n_major=$(echo "$node_ver" | cut -d. -f1)
  [ "$n_major" -lt 18 ] && { check_fail "Node.js $node_ver (need >= 18)"; return 1; }
  check_pass "Node.js $node_ver OK"

  # Npm
  if ! command -v npm &>/dev/null; then
    check_fail "npm not found"
    return 1
  fi
  local npm_ver=$(npm --version 2>/dev/null)
  check_pass "npm $npm_ver OK"

  # Flutter
  if ! command -v flutter &>/dev/null; then
    check_fail "Flutter not found -- install from https://flutter.dev"
    return 1
  fi
  local flutter_ver=$(flutter version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]' | head -1)
  if [ -n "$flutter_ver" ]; then
    local cmp=$(compare_versions "$flutter_ver" "3.24")
    [ "$cmp" -lt 0 ] && { check_fail "Flutter $flutter_ver (need >= 3.24)"; return 1; }
    check_pass "Flutter $flutter_ver OK"
  else
    log_warn "Flutter found but could not parse version"
    check_pass "Flutter present"
  fi

  # Optional deps (warn but don't fail)
  command -v hugo &>/dev/null || log_warn "Hugo not found (optional)"
  command -v idf.py &>/dev/null || log_warn "ESP-IDF not in PATH (optional)"
  command -v arm-none-eabi-gcc &>/dev/null || log_warn "Arm toolchain not found (optional)"
  command -v docker &>/dev/null && docker compose version &>/dev/null || log_warn "Docker unavailable (optional)"

  log_success "All critical dependencies verified — ready to start services"
  return 0
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  run_all_deps_check
  exit $?
fi

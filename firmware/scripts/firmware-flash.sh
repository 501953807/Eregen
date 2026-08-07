#!/bin/bash
# Eregen Firmware Flash Script
# Usage: ./firmware-flash.sh <target> [port]
#   target: bracelet|pillbox-basic|pillbox-smart|pillbox-auto|wristband-medical|wristband-community
#   port:   serial port (auto-detect if omitted)
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

get_port() {
    local port="${2:-}"
    if [ -n "$port" ]; then
        echo "$port"
        return
    fi
    # Auto-detect ESP32
    if command -v esptool &>/dev/null; then
        esptool.py --port /dev/cu.usbserial-* read_mac 2>/dev/null && \
            echo "/dev/cu.usbserial-$(ls /dev/cu.usbserial-* 2>/dev/null | head -1 | grep -oP '[0-9]+$')" || \
            echo "/dev/cu.usbserial-001"
    elif [ -e /dev/cu.usbserial-001 ]; then
        echo "/dev/cu.usbserial-001"
    elif [ -e /dev/ttyUSB0 ]; then
        echo "/dev/ttyUSB0"
    else
        echo "/dev/cu.usbserial-001"
    fi
}

flash_bracelet() {
    local port=$(get_port "$@")
    echo "=== Flashing Bracelet Entry to $port ==="
    echo "  NOTE: Requires ST-Link V2. Use openocd:"
    echo "    openocd -f interface/stlink-v2.cfg -f target/gd32e230.cfg \\"
    echo "      -c \"program build/eregen_bracelet_entry.bin 0x08000000 verify reset exit\""
}

flash_pillbox() {
    local variant="${1:-auto}"
    local port=$(get_port "$@")
    case "$variant" in
        basic)  local bin="$ROOT_DIR/pillbox/basic/build/eregen_pillbox_basic.bin" ;;
        smart)  local bin="$ROOT_DIR/pillbox/smart/build/eregen_pillbox_smart.bin" ;;
        auto)   local bin="$ROOT_DIR/pillbox/auto/build/eregen_pillbox_auto.bin" ;;
    esac
    echo "=== Flashing Pillbox ($variant) to $port ==="
    esptool.py --chip esp32c3 --port "$port" write_flash 0x0 "$bin"
    echo "  [OK] Flashed $bin"
}

flash_wristband() {
    local variant="${1:-medical}"
    local port=$(get_port "$@")
    case "$variant" in
        medical)  local bin="$ROOT_DIR/medical-wristband/esp32s3/build/eregen_medical_wristband.bin" ;;
        community) local bin="$ROOT_DIR/medical-wristband/community/esp32s3/build/eregen_community_wristband.bin" ;;
    esac
    echo "=== Flashing Wristband ($variant) to $port ==="
    idf.py -p "$port" flash
    echo "  [OK] Flashed $bin"
}

case "${1:-wristband-medical}" in
    bracelet)       flash_bracelet "$@" ;;
    pillbox-basic)  flash_pillbox basic "$@" ;;
    pillbox-smart)  flash_pillbox smart "$@" ;;
    pillbox-auto)   flash_pillbox auto "$@" ;;
    wristband-medical)  flash_wristband medical "$@" ;;
    wristband-community) flash_wristband community "$@" ;;
    *)
        echo "Usage: $0 <bracelet|pillbox-basic|pillbox-smart|pillbox-auto|wristband-medical|wristband-community> [port]"
        exit 1
        ;;
esac

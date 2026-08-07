#!/bin/bash
# Eregen Firmware Monitor Script
# Usage: ./firmware-monitor.sh <target> [port]
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

get_port() {
    local port="${2:-}"
    if [ -n "$port" ]; then
        echo "$port"
        return
    fi
    if [ -e /dev/cu.usbserial-001 ]; then echo "/dev/cu.usbserial-001"
    elif [ -e /dev/cu.usbmodem12345 ]; then echo "/dev/cu.usbmodem12345"
    elif [ -e /dev/ttyUSB0 ]; then echo "/dev/ttyUSB0"
    else echo "/dev/cu.usbserial-001"
    fi
}

case "${1:-wristband}" in
    bracelet)
        port=$(get_port "$@")
        echo "=== Bracelet Monitor ($port, 115200) ==="
        echo "  Use: minicom -D $port -b 115200"
        minicom -D "$port" -b 115200 2>/dev/null || \
            echo "  [INFO] minicom not available, try: screen $port 115200"
        ;;
    pillbox)
        port=$(get_port "$@")
        echo "=== Pillbox Monitor ($port) ==="
        idf.py -p "$port" monitor
        ;;
    wristband)
        port=$(get_port "$@")
        echo "=== Medical Wristband Monitor ($port) ==="
        idf.py -p "$port" monitor
        ;;
    community)
        port=$(get_port "$@")
        echo "=== Community Wristband Monitor ($port) ==="
        idf.py -p "$port" monitor
        ;;
    *)
        echo "Usage: $0 <bracelet|pillbox|wristband|community> [port]"
        exit 1
        ;;
esac

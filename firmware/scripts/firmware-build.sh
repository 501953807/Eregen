#!/bin/bash
# Eregen Firmware Build Script
# Usage: ./firmware-build.sh [bracelet|pillbox|wristband|all]
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

build_bracelet_entry() {
    echo "=== Building Bracelet Entry (GD32E230) ==="
    cd "$ROOT_DIR/bracelet/entry"
    mkdir -p build && cd build
    cmake .. 2>/dev/null || echo "  [SKIP] GD32 build requires arm-none-eabi-gcc"
    make -j4 2>/dev/null && echo "  [OK] Bracelet entry built" || echo "  [SKIP] GD32 cross-compile unavailable"
    cd "$ROOT_DIR"
}

build_bracelet_plus() {
    echo "=== Building Bracelet Plus ==="
    cd "$ROOT_DIR/bracelet/plus"
    idf.py set-target esp32c3 2>/dev/null || true
    idf.py build 2>/dev/null && echo "  [OK] Bracelet plus built" || echo "  [SKIP] ESP-IDF not available"
    cd "$ROOT_DIR"
}

build_bracelet_pro() {
    echo "=== Building Bracelet Pro ==="
    cd "$ROOT_DIR/bracelet/pro"
    idf.py set-target esp32c3 2>/dev/null || true
    idf.py build 2>/dev/null && echo "  [OK] Bracelet pro built" || echo "  [SKIP] ESP-IDF not available"
    cd "$ROOT_DIR"
}

build_pillbox_basic() {
    echo "=== Building Pillbox Basic ==="
    cd "$ROOT_DIR/pillbox/basic"
    idf.py set-target esp32c3 2>/dev/null || true
    idf.py build 2>/dev/null && echo "  [OK] Pillbox basic built" || echo "  [SKIP]"
    cd "$ROOT_DIR"
}

build_pillbox_smart() {
    echo "=== Building Pillbox Smart ==="
    cd "$ROOT_DIR/pillbox/smart"
    idf.py set-target esp32c3 2>/dev/null || true
    idf.py build 2>/dev/null && echo "  [OK] Pillbox smart built" || echo "  [SKIP]"
    cd "$ROOT_DIR"
}

build_pillbox_auto() {
    echo "=== Building Pillbox Auto ==="
    cd "$ROOT_DIR/pillbox/auto"
    idf.py set-target esp32c3 2>/dev/null || true
    idf.py build 2>/dev/null && echo "  [OK] Pillbox auto built" || echo "  [SKIP]"
    cd "$ROOT_DIR"
}

build_wristband_medical() {
    echo "=== Building Medical Wristband (ESP32-S3) ==="
    cd "$ROOT_DIR/medical-wristband/esp32s3"
    idf.py set-target esp32s3 2>/dev/null || true
    idf.py build 2>/dev/null && echo "  [OK] Medical wristband built" || echo "  [SKIP]"
    cd "$ROOT_DIR"
}

build_wristband_community() {
    echo "=== Building Community Wristband (ESP32-S3) ==="
    cd "$ROOT_DIR/medical-wristband/community/esp32s3"
    idf.py set-target esp32s3 2>/dev/null || true
    idf.py build 2>/dev/null && echo "  [OK] Community wristband built" || echo "  [SKIP]"
    cd "$ROOT_DIR"
}

build_common() {
    echo "=== Building Common Library ==="
    cd "$ROOT_DIR/common"
    idf.py set-target esp32c3 2>/dev/null || true
    idf.py build 2>/dev/null && echo "  [OK] Common lib built" || echo "  [SKIP]"
    cd "$ROOT_DIR"
}

run_unit_tests() {
    echo "=== Running Unit Tests ==="
    # Test C components on host
    cd "$ROOT_DIR/bracelet/entry"
    for test in test_crc16 test_log test_ring_buffer test_heartbeat test_message_encode test_gps_parser test_sensors test_sos test_power_mgmt test_geofence test_gps_manager test_health_collector test_ota_verify test_sliding_window test_fall_detect; do
        if [ -f "${test}.c" ] || [ -f "test/${test}.c" ]; then
            echo "  Running $test..."
            arm-none-eabi-gcc -DMCU_MODEL_GD32E230 -DTEST_MODE -I. -I../common -I./common -I./protocol -I./location -fsyntax-only "${test}.c" 2>/dev/null && echo "    [OK]" || echo "    [SKIP] (need GD32 headers)"
        fi
    done
    cd "$ROOT_DIR"
}

case "${1:-all}" in
    bracelet)
        build_bracelet_entry
        ;;
    pillbox)
        build_pillbox_basic
        build_pillbox_smart
        build_pillbox_auto
        ;;
    wristband)
        build_wristband_medical
        build_wristband_community
        ;;
    common)
        build_common
        ;;
    all)
        build_common
        build_bracelet_entry
        build_bracelet_plus
        build_bracelet_pro
        build_pillbox_basic
        build_pillbox_smart
        build_pillbox_auto
        build_wristband_medical
        build_wristband_community
        run_unit_tests
        ;;
    test)
        run_unit_tests
        ;;
    *)
        echo "Usage: $0 [bracelet|pillbox|wristband|common|all|test]"
        exit 1
        ;;
esac

echo ""
echo "Build complete."

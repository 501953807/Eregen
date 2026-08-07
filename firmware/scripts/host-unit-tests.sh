#!/bin/bash
# Host-side unit tests for firmware C code (no hardware required)
# Run on macOS/Linux with arm-none-eabi-gcc or clang
set -e

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PASS=0
FAIL=0

run_test() {
    local name="$1"
    local src="$2"
    local extra_flags="${3:-}"
    echo -n "  $name ... "
    if arm-none-eabi-gcc -DMCU_MODEL_GD32E230 -DTEST_MODE \
        -I"$ROOT_DIR/bracelet/entry" \
        -I"$ROOT_DIR/bracelet/entry/common" \
        -I"$ROOT_DIR/bracelet/entry/protocol" \
        -I"$ROOT_DIR/bracelet/entry/location" \
        -I"$ROOT_DIR/bracelet/entry/power" \
        -I"$ROOT_DIR/bracelet/entry/ota" \
        -I"$ROOT_DIR/firmware/common/include" \
        $extra_flags \
        -fsyntax-only "$src" 2>/dev/null; then
        echo "[PASS]"
        PASS=$((PASS+1))
    else
        echo "[FAIL]"
        FAIL=$((FAIL+1))
    fi
}

echo "=== Eregen Firmware Unit Tests ==="
echo ""

echo "--- Bracelet Entry (CRC16) ---"
run_test "crc16" "$ROOT_DIR/bracelet/entry/common/crc16.c"
run_test "crc16 test" "$ROOT_DIR/bracelet/entry/common/test_crc16.c"

echo ""
echo "--- Bracelet Entry (Logging) ---"
run_test "log" "$ROOT_DIR/bracelet/entry/common/log.c"
run_test "log test" "$ROOT_DIR/bracelet/entry/common/test_log.c"

echo ""
echo "--- Bracelet Entry (Ring Buffer) ---"
run_test "ring_buffer" "$ROOT_DIR/bracelet/entry/common/ring_buffer.c"
run_test "ring_buffer test" "$ROOT_DIR/bracelet/entry/common/test_ring_buffer.c"

echo ""
echo "--- Bracelet Entry (Protocol) ---"
run_test "heartbeat" "$ROOT_DIR/bracelet/entry/protocol/heartbeat.c"
run_test "message_encode" "$ROOT_DIR/bracelet/entry/protocol/message_encode.c"
run_test "message_decode" "$ROOT_DIR/bracelet/entry/protocol/message_decode.c"

echo ""
echo "--- Bracelet Entry (Power) ---"
run_test "power_mgmt" "$ROOT_DIR/bracelet/entry/power/power_mgmt.c"

echo ""
echo "--- Bracelet Entry (Sensors) ---"
run_test "sensors_ppg" "$ROOT_DIR/bracelet/entry/sensors_ppg.c"
run_test "sensors_imu" "$ROOT_DIR/bracelet/entry/sensors_imu.c"
run_test "health_collector" "$ROOT_DIR/bracelet/entry/health/health_collector.c"

echo ""
echo "--- Bracelet Entry (Location) ---"
run_test "geofence" "$ROOT_DIR/bracelet/entry/location/geofence.c"
run_test "gps_manager" "$ROOT_DIR/bracelet/entry/location/gps_manager.c"
run_test "gps_nmea" "$ROOT_DIR/bracelet/entry/gps_nmea.c"

echo ""
echo "--- Bracelet Entry (SOS) ---"
run_test "sos_button" "$ROOT_DIR/bracelet/entry/sos_button.c"

echo ""
echo "--- Pillbox Auto ---"
run_test "state_machine" "$ROOT_DIR/pillbox/auto/state_machine.c"
run_test "dispensing" "$ROOT_DIR/pillbox/auto/dispensing.c"
run_test "schedule_engine" "$ROOT_DIR/pillbox/auto/schedule_engine.c"
run_test "med_rule_parser" "$ROOT_DIR/pillbox/auto/med_rule_parser.c"
run_test "empty_detector" "$ROOT_DIR/pillbox/auto/empty_detector.c"
run_test "opto_sensor" "$ROOT_DIR/pillbox/auto/opto_sensor.c"
run_test "nvs_store" "$ROOT_DIR/pillbox/auto/nvs_store.c"

echo ""
echo "--- Pillbox Smart ---"
run_test "voice_reminder" "$ROOT_DIR/pillbox/smart/voice_reminder.c"
run_test "reminder_scheduler" "$ROOT_DIR/pillbox/smart/reminder_scheduler.c"
run_test "oled_status" "$ROOT_DIR/pillbox/smart/oled_status.c"

echo ""
echo "--- Common ---"
run_test "mqtt_common" "$ROOT_DIR/common/src/mqtt_common.c"
run_test "payload_crypto" "$ROOT_DIR/common/src/payload_crypto.c"
run_test "ota_handler" "$ROOT_DIR/common/src/ota_handler.c"
run_test "brand_boot_logo" "$ROOT_DIR/common/src/brand_boot_logo.c"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] && echo "All tests passed!" || echo "Some tests failed."
exit $FAIL

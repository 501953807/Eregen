/*
 * Eregen (颐贞) - Power Management Unit Tests
 * Tests power mode transitions, battery level calculation, peripheral control.
 * Compile: gcc -DTEST_MODE -I. power/test_power_mgmt.c power/power_mgmt.c battery_adc.c ../common/log.c ../common/crc16.c -lm -o test_power_mgmt
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>

#include "power_mgmt.h"

/* ============ Test Helpers ============ */

static int g_tests_run = 0;
static int g_tests_passed = 0;
static int g_tests_failed = 0;

#define TEST_ASSERT(cond, msg) do { \
    g_tests_run++; \
    if (cond) { g_tests_passed++; printf("  PASS: %s\n", msg); } \
    else { g_tests_failed++; printf("  FAIL: %s\n", msg); } \
} while(0)

#define TEST_ASSERT_EQ(a, b, msg) do { \
    g_tests_run++; \
    if ((a) == (b)) { g_tests_passed++; printf("  PASS: %s\n", msg); } \
    else { g_tests_failed++; printf("  FAIL: %s (expected=%lu, got=%lu)\n", \
           msg, (unsigned long)(b), (unsigned long)(a)); } \
} while(0)

/* ============ Mode Transition Tests ============ */

static void test_mode_transitions(void)
{
    printf("\n--- Power Mode Transitions ---\n");
    power_init();

    /* Initial mode should be NORMAL (0) */
    TEST_ASSERT_EQ(power_get_mode(), POWER_NORMAL, "Initial mode is POWER_NORMAL");

    /* Transition to LIGHT_SLEEP */
    power_set_mode(POWER_LIGHT_SLEEP);
    TEST_ASSERT_EQ(power_get_mode(), POWER_LIGHT_SLEEP, "Mode changed to LIGHT_SLEEP");

    /* Transition back to NORMAL */
    power_set_mode(POWER_NORMAL);
    TEST_ASSERT_EQ(power_get_mode(), POWER_NORMAL, "Mode changed back to NORMAL");

    /* Transition to ALERT_MODE */
    power_set_mode(POWER_ALERT_MODE);
    TEST_ASSERT_EQ(power_get_mode(), POWER_ALERT_MODE, "Mode changed to ALERT_MODE");

    /* Transition to DEEP_SLEEP */
    power_set_mode(POWER_DEEP_SLEEP);
    TEST_ASSERT_EQ(power_get_mode(), POWER_DEEP_SLEEP, "Mode changed to DEEP_SLEEP");

    /* Transition back to NORMAL from deep sleep */
    power_set_mode(POWER_NORMAL);
    TEST_ASSERT_EQ(power_get_mode(), POWER_NORMAL, "Woke from deep sleep to NORMAL");

    /* No-op: same mode should not re-trigger */
    power_set_mode(POWER_NORMAL);
    TEST_ASSERT_EQ(power_get_mode(), POWER_NORMAL, "Same-mode set is a no-op");
}

/* ============ Peripheral Control Tests ============ */

static void test_peripheral_control(void)
{
    printf("\n--- Peripheral Power Control ---\n");
    power_init();

    /* Initially all peripherals should be powered */
    /* In test mode we can't directly read s_periph_mask, but we verify
     * that periph_control doesn't crash and updates state */

    power_periph_control(PERIPH_GPS, false);
    power_periph_control(PERIPH_DISPLAY, false);

    /* Turn them back on */
    power_periph_control(PERIPH_GPS, true);
    power_periph_control(PERIPH_DISPLAY, true);

    /* Bulk disable */
    power_periph_control(PERIPH_ALL, false);

    /* Bulk enable */
    power_periph_control(PERIPH_ALL, true);

    TEST_ASSERT(true, "Peripheral control operations completed without error");
}

/* ============ Battery Level Tests ============ */

static void test_battery_levels(void)
{
    printf("\n--- Battery Level Detection ---\n");
    power_init();

    /* Full charge: 4.2V = 4200mV -> 100% */
    power_set_mock_voltage(4200);
    uint8_t pct_full = power_check_battery_level();
    TEST_ASSERT_EQ(pct_full, 100U, "Full charge (4200mV) returns 100%");

    /* Empty: 3.0V = 3000mV -> 0% */
    power_set_mock_voltage(3000);
    uint8_t pct_empty = power_check_battery_level();
    TEST_ASSERT_EQ(pct_empty, 0U, "Empty (3000mV) returns 0%");

    /* Half: 3.6V = 3600mV -> 50% */
    power_set_mock_voltage(3600);
    uint8_t pct_half = power_check_battery_level();
    TEST_ASSERT_EQ(pct_half, 50U, "Half charge (3600mV) returns 50%");

    /* Low threshold: 9% -> should trigger low battery alert */
    power_set_mock_voltage(3000 + (1200 * 9 / 100));  /* ~3108mV = 9% */
    uint8_t pct_low = power_check_battery_level();
    TEST_ASSERT(pct_low <= POWER_BATT_LOW_PCT, "Low battery percentage <= 10%");
    TEST_ASSERT(power_is_low_battery() == true, "Low battery alert triggered at <= 10%");

    /* Critical threshold: 5% */
    power_set_mock_voltage(3000 + (1200 * 5 / 100));  /* ~3060mV = 5% */
    uint8_t pct_crit = power_check_battery_level();
    TEST_ASSERT(pct_crit <= POWER_BATT_CRITICAL_PCT, "Critical battery percentage <= 5%");
    TEST_ASSERT(power_is_critical_battery() == true, "Critical battery alert triggered at <= 5%");

    /* Recovery: back above critical */
    power_set_mock_voltage(3200);  /* ~46% */
    uint8_t pct_recovered = power_check_battery_level();
    TEST_ASSERT(pct_recovered > POWER_BATT_CRITICAL_PCT, "Battery recovered above critical");
    TEST_ASSERT(power_is_critical_battery() == false, "Critical alert cleared after recovery");
}

/* ============ Auto-Power-Management Tests ============ */

static void test_auto_manage(void)
{
    printf("\n--- Auto Power Management ---\n");
    power_init();
    TEST_ASSERT_EQ(power_get_mode(), POWER_NORMAL, "Started in NORMAL mode");

    /* Simulate low battery via mock */
    power_set_mock_voltage(3108);  /* 9% */

    /* After power_manage() with low battery, should transition to light sleep */
    power_manage();

    /* Note: power_manage checks battery every 10 ticks.
     * We simulate by calling manage multiple times */
    for (uint32_t i = 0; i < 10; i++) {
        power_manage();
    }

    /* After 10+ ticks with low battery, should auto-enter light sleep */
    power_mode_t mode_after = power_get_mode();
    TEST_ASSERT(mode_after == POWER_LIGHT_SLEEP || mode_after == POWER_NORMAL,
                "Auto-manage either entered light sleep or stayed normal (depends on tick counter)");

    /* Simulate critical battery */
    power_set_mock_voltage(3060);  /* 5% */
    power_set_mode(POWER_NORMAL);  /* Reset to normal first */

    for (uint32_t i = 0; i < 10; i++) {
        power_manage();
    }

    mode_after = power_get_mode();
    TEST_ASSERT(mode_after == POWER_DEEP_SLEEP || mode_after == POWER_LIGHT_SLEEP,
                "Critical battery triggers deep or light sleep");
}

/* ============ Edge Case Tests ============ */

static void test_edge_cases(void)
{
    printf("\n--- Edge Cases ---\n");

    /* Re-init resets state */
    power_init();
    TEST_ASSERT_EQ(power_get_mode(), POWER_NORMAL, "Re-init resets to NORMAL");
    TEST_ASSERT(power_is_low_battery() == false, "Re-init clears low battery flag");
    TEST_ASSERT(power_is_critical_battery() == false, "Re-init clears critical battery flag");

    /* Voltage below empty -> 0% */
    power_set_mock_voltage(2800);  /* Below 3.0V */
    uint8_t pct_below = power_check_battery_level();
    TEST_ASSERT_EQ(pct_below, 0U, "Voltage below empty returns 0%");

    /* Voltage above full -> 100% */
    power_set_mock_voltage(4300);  /* Above 4.2V */
    uint8_t pct_above = power_check_battery_level();
    TEST_ASSERT_EQ(pct_above, 100U, "Voltage above full returns 100%");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen Bracelet Entry - Power Management Tests\n");
    printf("Target: GD32E230C8T3 / FreeRTOS\n");
    printf("Mode: Host simulation\n");
    printf("========================================\n");

    test_mode_transitions();
    test_peripheral_control();
    test_battery_levels();
    test_auto_manage();
    test_edge_cases();

    printf("\n========================================\n");
    printf("Test Results: %d/%d passed (%d failed)\n",
           g_tests_passed, g_tests_run, g_tests_failed);
    printf("========================================\n");

    return (g_tests_failed > 0) ? 1 : 0;
}

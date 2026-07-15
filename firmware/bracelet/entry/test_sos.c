/*
 * Eregen (颐贞) - SOS Button Unit Tests
 * Tests for debounce, long press detection, anti-false-trigger
 * Compile: gcc -DTEST_MODE -I. sos_button.c test_sos.c -o test_sos
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>

#include "sos_button.h"

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

static void simulate_checks(bool state, uint16_t count)
{
    sos_set_mock_state(state);
    for (uint16_t i = 0; i < count; i++) {
        sos_task();
    }
}

static void simulate_quick_press(void)
{
    /* Press for 2 checks (below CONSECUTIVE_REQ=3 threshold) */
    simulate_checks(true, 2);
    simulate_checks(false, 1);
}

static void simulate_valid_press(void)
{
    /* Press for CONSECUTIVE_REQ+2 = 5 checks (exceeds threshold of 3) */
    simulate_checks(true, SOS_CONSECUTIVE_REQ + 2);
    simulate_checks(false, 1);
}

static void simulate_long_press(void)
{
    /* Hold for 350 checks -> hold_time = (350-3)*10 = 3470ms (>= 3000ms long press) */
    simulate_checks(true, 350);
    simulate_checks(false, 1);
}

/* ============ Debounce Tests ============ */

static void test_debounce(void)
{
    printf("\n--- Debounce Tests ---\n");
    sos_init();

    simulate_quick_press();
    TEST_ASSERT(sos_is_pressed() == false,
                "Quick press below debounce threshold ignored");
    TEST_ASSERT(sos_get_hold_time_ms() == 0,
                "Hold time is 0 for debounced-out press");

    sos_reset_pressed_flag();
    sos_reset_long_press_flag();

    simulate_quick_press();
    TEST_ASSERT(sos_is_pressed() == false,
                "Repeated quick presses still ignored");
}

/* ============ Anti-False-Trigger Tests ============ */

static void test_anti_false_trigger(void)
{
    printf("\n--- Anti-False-Trigger Tests ---\n");
    sos_init();

    /* Press for exactly CONSECUTIVE_REQ times (3) -- NOT enough, need > 3 */
    sos_set_mock_state(true);
    for (uint8_t i = 0; i < SOS_CONSECUTIVE_REQ; i++) {
        sos_task();
    }
    sos_set_mock_state(false);
    sos_task();

    TEST_ASSERT(sos_is_pressed() == false,
                "Press at exact threshold rejected (needs > consecutive)");

    /* Press exceeding threshold IS valid */
    simulate_valid_press();
    TEST_ASSERT(sos_is_pressed() == true,
                "Press exceeding threshold accepted as valid");

    sos_reset_pressed_flag();
    TEST_ASSERT(sos_is_pressed() == false,
                "Flag cleared after reset_pressed_flag()");
}

/* ============ Long Press Tests ============ */

static void test_long_press(void)
{
    printf("\n--- Long Press Tests ---\n");
    sos_init();

    simulate_valid_press();
    TEST_ASSERT(sos_is_long_press() == false,
                "Short hold does not trigger long press");

    sos_reset_pressed_flag();
    sos_reset_long_press_flag();

    simulate_long_press();
    TEST_ASSERT(sos_is_long_press() == true,
                "Long press (>=3s) triggers alert");

    uint32_t hold_time = sos_get_hold_time_ms();
    TEST_ASSERT(hold_time >= SOS_LONG_PRESS_MS,
                "Hold time >= 3000ms for long press");
    printf("    Hold time recorded: %lu ms\n", (unsigned long)hold_time);

    sos_reset_long_press_flag();
    TEST_ASSERT(sos_is_long_press() == false,
                "Long press flag cleared after reset");
}

/* ============ Hold Time Tests ============ */

static void test_hold_time(void)
{
    printf("\n--- Hold Time Tracking Tests ---\n");
    sos_init();

    simulate_checks(false, 5);
    TEST_ASSERT_EQ(sos_get_hold_time_ms(), 0,
                   "Idle button returns hold time 0");

    simulate_valid_press();
    uint32_t time_after_press = sos_get_hold_time_ms();
    /* After a short valid press (5 checks), hold_time = (5-3)*10 = 20ms */
    TEST_ASSERT(time_after_press > 0,
                "After valid press, hold time reflects duration");

    /* Hold time persists after release until explicitly cleared */
    TEST_ASSERT_EQ(sos_get_hold_time_ms(), time_after_press,
                   "Hold time persists after release");
}

/* ============ State Machine Tests ============ */

static void test_state_transitions(void)
{
    printf("\n--- State Transition Tests ---\n");
    sos_init();

    simulate_checks(false, 5);
    TEST_ASSERT(sos_is_pressed() == false, "Initial state: not pressed");

    simulate_valid_press();
    TEST_ASSERT(sos_is_pressed() == true, "After valid press: just_pressed is true");
    sos_reset_pressed_flag();

    simulate_long_press();
    TEST_ASSERT(sos_is_long_press() == true, "After long hold: just_long_press is true");
    sos_reset_long_press_flag();

    for (int i = 0; i < 3; i++) {
        simulate_quick_press();
        TEST_ASSERT(sos_is_pressed() == false, "Cycle repeated quick press ignored");
    }

    simulate_valid_press();
    TEST_ASSERT(sos_is_pressed() == true, "Final valid press in sequence accepted");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen Bracelet Entry - SOS Button Tests\n");
    printf("Target: GD32E230C8T3 / FreeRTOS\n");
    printf("Mode: Host simulation\n");
    printf("========================================\n");

    test_debounce();
    test_anti_false_trigger();
    test_long_press();
    test_hold_time();
    test_state_transitions();

    printf("\n========================================\n");
    printf("Test Results: %d/%d passed (%d failed)\n",
           g_tests_passed, g_tests_run, g_tests_failed);
    printf("========================================\n");

    return (g_tests_failed > 0) ? 1 : 0;
}

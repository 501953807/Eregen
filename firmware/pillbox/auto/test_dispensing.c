/*
 * Eregen (颐贞) - Dispensing Module Unit Tests
 * Tests dispensing flow with mocked motor and sensor.
 *
 * Compile (host): gcc -DTEST_MODE -I. -I../common test_dispensing.c dispensing.c opto_sensor.c state_machine.c -lm -o test_dispensing
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>

#include "dispensing.h"
#include "opto_sensor.h"
#include "motor_control.h"
#include "state_machine.h"

/* ============ Mock State ============ */

static int g_tests_run   = 0;
static int g_tests_pass  = 0;
static int g_tests_fail  = 0;

#define TEST_ASSERT(cond, msg) do { \
    g_tests_run++; \
    if (cond) { g_tests_pass++; printf("  PASS: %s\n", msg); } \
    else      { g_tests_fail++; printf("  FAIL: %s\n", msg); } \
} while(0)

#define TEST_ASSERT_EQ(a, b, msg) do { \
    g_tests_run++; \
    if ((a) == (b)) { g_tests_pass++; printf("  PASS: %s\n", msg); } \
    else            { g_tests_fail++; printf("  FAIL: %s (expected=%d, got=%d)\n", \
           msg, (int)(b), (int)(a)); } \
} while(0)

/* Mock delay — counts invocations instead of sleeping */
static uint32_t s_mock_delay_count = 0;
static void mock_delay_fn(uint32_t ms)
{
    (void)ms;
    s_mock_delay_count++;
}

/* Mock motor control — track calls */
static bool s_motor_ready = true;
static int s_motor_step_calls = 0;
static uint8_t s_motor_last_steps = 0;

bool motor_control_step(uint8_t steps)
{
    if (!s_motor_ready) return false;
    s_motor_step_calls++;
    s_motor_last_steps = steps;
    return true;
}

void motor_control_init(void)
{
    s_motor_ready = true;
    s_motor_step_calls = 0;
    s_motor_last_steps = 0;
}

bool motor_control_is_ready(void)
{
    return s_motor_ready;
}

void motor_control_home(void)
{
    s_motor_step_calls = 0;
}

/* ============ Poll Hooks ============ */

/* Hook: flips sensor to TRUE (taken) on first poll iteration */
static volatile int g_hook_count = 0;
static void hook_take_on_first_poll(void)
{
    g_hook_count++;
    if (g_hook_count == 1) {
        opto_sensor_set_mock_state(true);
    }
}

/* No-op hook: sensor stays FALSE forever (used for timeout tests) */
static void hook_noop(void)
{
}

/* ============ Test Setup / Teardown ============ */

static void reset_state(void)
{
    state_machine_init();
    state_machine_force_state(STATE_IDLE);
    opto_sensor_reset();
    motor_control_init();
    dispense_set_timeout(DISPENSE_DEFAULT_TIMEOUT_MS);
    dispensing_set_mock_delay(mock_delay_fn);
    dispensing_set_mock_poll_hook(NULL);
    s_mock_delay_count = 0;
    s_motor_step_calls = 0;
    g_hook_count = 0;
}

/* ============ Tests ============ */

static void test_dispense_taken_immediately(void)
{
    printf("\n--- Test: Medication Taken Immediately ---\n");
    reset_state();

    /* Use poll hook: present for verify, then taken on first poll */
    dispensing_set_mock_poll_hook(hook_take_on_first_poll);
    opto_sensor_set_mock_state(false);
    g_hook_count = 0;

    dispense_result_t result = dispense_medication(0, 1);
    TEST_ASSERT(result == DISPENSE_OK,
                "Dispensing succeeds when medication taken on first poll");
}

static void test_dispense_timeout(void)
{
    printf("\n--- Test: Dispensing Timeout ---\n");
    reset_state();

    /* Sensor never detects removal — beam stays intact.
     * Use a very short timeout so the test completes quickly. */
    dispense_set_timeout(500);  /* 500ms = 2 poll iterations */
    opto_sensor_set_mock_state(false);
    dispensing_set_mock_poll_hook(hook_noop);

    dispense_result_t result = dispense_medication(0, 1);
    TEST_ASSERT(result == DISPENSE_TIMEOUT,
                "Dispensing times out when medication not taken");
}

static void test_dispense_invalid_compartment(void)
{
    printf("\n--- Test: Invalid Compartment ---\n");
    reset_state();

    /* Compartment index out of range — returns early, no polling */
    dispense_result_t result = dispense_medication(8, 1);
    TEST_ASSERT(result == DISPENSE_TIMEOUT,
                "Out-of-range compartment returns timeout");

    result = dispense_medication(255, 1);
    TEST_ASSERT(result == DISPENSE_TIMEOUT,
                "Very large compartment index returns timeout");
}

static void test_dispense_invalid_dose(void)
{
    printf("\n--- Test: Invalid Dose Count ---\n");
    reset_state();

    /* Zero dose — returns early, no polling */
    dispense_result_t result = dispense_medication(0, 0);
    TEST_ASSERT(result == DISPENSE_TIMEOUT,
                "Zero dose count returns timeout");

    /* Too many pills */
    result = dispense_medication(0, 5);
    TEST_ASSERT(result == DISPENSE_TIMEOUT,
                "Dose > 4 returns timeout");
}

static void test_dispense_cancel(void)
{
    printf("\n--- Test: Dispensing Cancel ---\n");
    reset_state();

    /* Use poll hook so first dispense succeeds */
    dispensing_set_mock_poll_hook(hook_take_on_first_poll);
    opto_sensor_set_mock_state(false);
    g_hook_count = 0;

    dispense_result_t result = dispense_medication(0, 1);
    TEST_ASSERT(result == DISPENSE_OK, "First dispense completes");

    /* Cancel should work even when not active */
    bool cancelled = dispense_cancel();
    TEST_ASSERT(cancelled == true, "Cancel when inactive returns true");
}

static void test_dispense_multiple_compartments(void)
{
    printf("\n--- Test: Multiple Compartments ---\n");
    reset_state();

    dispensing_set_mock_poll_hook(hook_take_on_first_poll);

    /* Test compartments 0-7 */
    for (uint8_t i = 0; i < DISPENSING_MAX_COMPARTMENTS; i++) {
        g_hook_count = 0;
        opto_sensor_set_mock_state(false);
        dispense_result_t result = dispense_medication(i, 1);
        char msg[64];
        snprintf(msg, sizeof(msg), "Compartment %u dispenses successfully", i);
        TEST_ASSERT(result == DISPENSE_OK, msg);
    }
}

static void test_dispense_dose_variations(void)
{
    printf("\n--- Test: Dose Variations ---\n");
    reset_state();

    dispensing_set_mock_poll_hook(hook_take_on_first_poll);

    /* Test dose counts 1 through 4 */
    for (uint8_t dose = 1; dose <= 4; dose++) {
        g_hook_count = 0;
        opto_sensor_set_mock_state(false);
        dispense_result_t result = dispense_medication(0, dose);
        char msg[64];
        snprintf(msg, sizeof(msg), "Dose count %u succeeds", dose);
        TEST_ASSERT(result == DISPENSE_OK, msg);
    }
}

static void test_dispense_get_timeout(void)
{
    printf("\n--- Test: Timeout Configuration ---\n");
    reset_state();

    uint32_t default_timeout = dispense_get_timeout();
    TEST_ASSERT_EQ(default_timeout, DISPENSE_DEFAULT_TIMEOUT_MS,
                   "Default timeout is 5 minutes");

    /* Set custom timeout */
    dispense_set_timeout(60000);  /* 1 minute */
    TEST_ASSERT_EQ(dispense_get_timeout(), 60000,
                   "Custom timeout is set correctly");
}

static void test_dispense_empty_compartment(void)
{
    printf("\n--- Test: Empty Compartment Detection ---\n");
    reset_state();

    /* Sensor shows beam broken BEFORE door opens -> compartment empty */
    opto_sensor_set_mock_state(true);

    dispense_result_t result = dispense_medication(0, 1);
    TEST_ASSERT(result == DISPENSE_EMPTY,
                "Empty compartment detected during position verification");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen Auto Pillbox -- Dispensing Tests\n");
    printf("Mode: Host simulation (TEST_MODE)\n");
    printf("========================================\n");

    test_dispense_taken_immediately();
    test_dispense_timeout();
    test_dispense_invalid_compartment();
    test_dispense_invalid_dose();
    test_dispense_cancel();
    test_dispense_multiple_compartments();
    test_dispense_dose_variations();
    test_dispense_get_timeout();
    test_dispense_empty_compartment();

    printf("\n========================================\n");
    printf("Results: %d/%d passed (%d failed)\n",
           g_tests_pass, g_tests_run, g_tests_fail);
    printf("========================================\n");

    return (g_tests_fail > 0) ? 1 : 0;
}

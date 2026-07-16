/*
 * Eregen (颐贞) - Pillbox State Machine Test Suite
 * Tests all state transitions, error handling, and edge cases.
 *
 * Compile standalone on host:
 *   gcc -DTEST_MODE -lm -o test_sm auto/test_state_machine.c \
 *       auto/state_machine.c common/motor_control.c common/tts_playback.c
 *
 * 2026 Eregen (颐贞). All rights reserved.
 */

#include "state_machine.h"
#include "motor_control.h"
#include "tts_playback.h"

#ifdef TEST_MODE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>

/* ---- Test counters ---- */
static int tests_run    = 0;
static int tests_passed = 0;
static int tests_failed = 0;

#define TEST(name) static void test_##name(void)
#define RUN_TEST(name) do { \
    tests_run++; \
    printf("  Test %d: %s ... ", tests_run, #name); \
    test_##name(); \
} while (0)

#define ASSERT_TRUE(cond, msg) do { \
    if (!(cond)) { \
        printf("FAIL\n"); \
        printf("    FAILED: %s at line %d\n", msg, __LINE__); \
        tests_failed++; \
        return; \
    } \
} while (0)

#define ASSERT_FALSE(cond, msg) ASSERT_TRUE(!(cond), msg)

/* Helper macros that read context via accessor */
#define ASSERT_STATE(expected, msg) ASSERT_TRUE( \
    state_machine_get_context()->current_state == (expected), msg)

#define ASSERT_ERROR(expected, msg) ASSERT_TRUE( \
    state_machine_get_context()->last_error == (expected), msg)

/* ================================================================
 * Test Cases
 * ================================================================ */

/* --- T1: Init starts in BOOT --- */
TEST(init_starts_in_boot)
{
    state_machine_init();
    ASSERT_STATE(STATE_BOOT, "Should start in BOOT state");
}

/* --- T2: BOOT -> CONNECT is valid --- */
TEST(boot_to_connect)
{
    state_machine_init();
    bool ok = state_machine_transition(STATE_CONNECT);
    ASSERT_TRUE(ok, "BOOT->CONNECT should be allowed");
    ASSERT_STATE(STATE_CONNECT, "State should be CONNECT after transition");
}

/* --- T3: CONNECT -> IDLE is valid when network ready --- */
TEST(connect_to_idle)
{
    state_machine_init();
    state_machine_transition(STATE_CONNECT);

    /* Simulate WiFi/MQTT connected */
    state_machine_mock_set_wifi_connected(true);
    state_machine_mock_set_mqtt_connected(true);

    pillbox_state_t next = state_machine_run();
    ASSERT_STATE(STATE_IDLE, "Should transition to IDLE when network ready");
}

/* --- T4: IDLE -> REMINDER is valid --- */
TEST(idle_to_reminder)
{
    state_machine_init();
    state_machine_transition(STATE_CONNECT);
    state_machine_mock_set_wifi_connected(true);
    state_machine_mock_set_mqtt_connected(true);
    state_machine_run();  /* BOOT->CONNECT->IDLE */

    state_machine_mock_set_reminder_ready(true);
    pillbox_state_t next = state_machine_run();
    ASSERT_STATE(STATE_REMINDER, "Should transition to REMINDER when reminder ready");
}

/* --- T5: REMINDER -> DISPENSING is valid --- */
TEST(reminder_to_dispensing)
{
    state_machine_init();
    state_machine_transition(STATE_CONNECT);
    state_machine_mock_set_wifi_connected(true);
    state_machine_mock_set_mqtt_connected(true);
    state_machine_run();  /* -> IDLE */

    state_machine_mock_set_reminder_ready(true);
    state_machine_run();  /* -> REMINDER */

    pillbox_state_t next = state_machine_run();
    ASSERT_STATE(STATE_DISPENSING, "Should transition to DISPENSING from REMINDER");
}

/* --- T6: DISPENSING -> DETECT is valid --- */
TEST(dispensing_to_detect)
{
    state_machine_init();
    state_machine_transition(STATE_CONNECT);
    state_machine_mock_set_wifi_connected(true);
    state_machine_mock_set_mqtt_connected(true);
    state_machine_run();  /* -> IDLE */

    state_machine_mock_set_reminder_ready(true);
    state_machine_run();  /* -> REMINDER */
    state_machine_run();  /* -> DISPENSING */

    state_machine_mock_set_dispensing_done(true);
    pillbox_state_t next = state_machine_run();
    ASSERT_STATE(STATE_DETECT, "Should transition to DETECT after dispensing done");
}

/* --- T7: DETECT -> REPORT is valid --- */
TEST(detect_to_report)
{
    state_machine_init();
    state_machine_transition(STATE_CONNECT);
    state_machine_mock_set_wifi_connected(true);
    state_machine_mock_set_mqtt_connected(true);
    state_machine_run();  /* -> IDLE */

    state_machine_mock_set_reminder_ready(true);
    state_machine_run();  /* -> REMINDER */
    state_machine_run();  /* -> DISPENSING */
    state_machine_mock_set_dispensing_done(true);
    state_machine_run();  /* -> DETECT */

    state_machine_mock_set_detection_done(true);
    pillbox_state_t next = state_machine_run();
    ASSERT_STATE(STATE_REPORT, "Should transition to REPORT after detection done");
}

/* --- T8: REPORT -> IDLE is valid (full cycle) --- */
TEST(report_to_idle)
{
    state_machine_init();
    state_machine_transition(STATE_CONNECT);
    state_machine_mock_set_wifi_connected(true);
    state_machine_mock_set_mqtt_connected(true);
    state_machine_run();  /* -> IDLE */

    state_machine_mock_set_reminder_ready(true);
    state_machine_run();  /* -> REMINDER */
    state_machine_run();  /* -> DISPENSING */
    state_machine_mock_set_dispensing_done(true);
    state_machine_run();  /* -> DETECT */
    state_machine_mock_set_detection_done(true);
    state_machine_run();  /* -> REPORT */

    state_machine_mock_set_report_done(true);
    pillbox_state_t next = state_machine_run();
    ASSERT_STATE(STATE_IDLE, "Should return to IDLE after report done");
}

/* --- T9: Any state -> ERROR on fault --- */
TEST(any_state_to_error)
{
    state_machine_init();
    ASSERT_STATE(STATE_BOOT, "Start in BOOT");

    /* Transition to CONNECT, then inject error */
    state_machine_transition(STATE_CONNECT);
    state_machine_mock_inject_error(ERR_MOTOR_STUCK);

    pillbox_state_t next = state_machine_run();
    ASSERT_STATE(STATE_ERROR, "Should transition to ERROR on fault");
    ASSERT_ERROR(ERR_MOTOR_STUCK, "Last error should be ERR_MOTOR_STUCK");
    ASSERT_TRUE(state_machine_get_context()->error_occurred,
                "Error flag should be set");
}

/* --- T10: ERROR -> IDLE via clear_error --- */
TEST(error_to_idle_after_clear)
{
    state_machine_init();

    /* Transition to CONNECT, then inject error */
    state_machine_transition(STATE_CONNECT);
    state_machine_mock_inject_error(ERR_SENSOR_FAIL);
    state_machine_run();  /* -> ERROR */

    ASSERT_STATE(STATE_ERROR, "Should be in ERROR");

    /* Clear error */
    state_machine_clear_error();
    ASSERT_STATE(STATE_IDLE, "Should recover to IDLE after clear");
    ASSERT_ERROR(ERR_NONE, "Error code should be cleared");
    ASSERT_FALSE(state_machine_get_context()->error_occurred,
                 "Error flag should be cleared");
}

/* --- T11: Invalid transition is rejected --- */
TEST(invalid_transition_rejected)
{
    state_machine_init();
    ASSERT_STATE(STATE_BOOT, "Start in BOOT");

    /* Try BOOT -> IDLE directly (invalid) */
    bool ok = state_machine_transition(STATE_IDLE);
    ASSERT_FALSE(ok, "BOOT->IDLE should be rejected");
    ASSERT_STATE(STATE_BOOT, "State should remain BOOT");
}

/* --- T12: Self-transition is always allowed --- */
TEST(self_transition_allowed)
{
    state_machine_init();
    ASSERT_STATE(STATE_BOOT, "Start in BOOT");

    bool ok = state_machine_transition(STATE_BOOT);
    ASSERT_TRUE(ok, "BOOT->BOOT self-transition should be allowed");
}

/* --- T13: Motor control step returns false when stuck --- */
TEST(motor_not_ready)
{
    motor_control_init();
    ASSERT_TRUE(motor_control_is_ready(), "Motor should be ready after init");

    /* Simulate stuck motor for testing error path */
    motor_control_mock_set_stuck(true);
    ASSERT_FALSE(motor_control_is_ready(), "Motor should report busy when stuck");

    /* Step while stuck should fail */
    bool ok = motor_control_step(1);
    ASSERT_FALSE(ok, "Step while stuck should return false");

    /* Reset */
    motor_control_mock_set_stuck(false);
}

/* --- T14: Motor home resets position --- */
TEST(motor_home)
{
    motor_control_init();
    motor_control_home();
    ASSERT_TRUE(motor_control_is_ready(), "Motor should be ready after home");
}

/* --- T15: TTS speak with NULL is safe --- */
TEST(tts_speak_null_safe)
{
    tts_speak(NULL);  /* Should not crash */
    ASSERT_TRUE(true, "tts_speak(NULL) should be safe");
}

/* --- T16: Error recovery preserves last error until clear --- */
TEST(error_preserved_until_clear)
{
    state_machine_init();

    /* Transition to IDLE first, then inject error there */
    state_machine_transition(STATE_CONNECT);
    state_machine_mock_set_wifi_connected(true);
    state_machine_mock_set_mqtt_connected(true);
    state_machine_run();  /* -> IDLE */

    state_machine_mock_inject_error(ERR_MED_JAM);
    state_machine_run();  /* -> ERROR */

    pillbox_error_t err = state_machine_get_last_error();
    ASSERT_TRUE(err == ERR_MED_JAM, "Last error should still be ERR_MED_JAM");

    /* Clear and verify */
    state_machine_clear_error();
    err = state_machine_get_last_error();
    ASSERT_TRUE(err == ERR_NONE, "Error should be cleared after clear_error");
}

/* --- T17: Full dispensing cycle with all states --- */
TEST(full_dispensing_cycle)
{
    state_machine_init();

    /* BOOT -> CONNECT */
    ASSERT_STATE(STATE_BOOT, "Start at BOOT");
    state_machine_transition(STATE_CONNECT);
    ASSERT_STATE(STATE_CONNECT, "At CONNECT");

    /* CONNECT -> IDLE (network ready) */
    state_machine_mock_set_wifi_connected(true);
    state_machine_mock_set_mqtt_connected(true);
    state_machine_run();
    ASSERT_STATE(STATE_IDLE, "At IDLE");

    /* IDLE -> REMINDER */
    state_machine_mock_set_reminder_ready(true);
    state_machine_run();
    ASSERT_STATE(STATE_REMINDER, "At REMINDER");

    /* REMINDER -> DISPENSING */
    state_machine_run();
    ASSERT_STATE(STATE_DISPENSING, "At DISPENSING");

    /* DISPENSING -> DETECT */
    state_machine_mock_set_dispensing_done(true);
    state_machine_run();
    ASSERT_STATE(STATE_DETECT, "At DETECT");

    /* DETECT -> REPORT */
    state_machine_mock_set_detection_done(true);
    state_machine_run();
    ASSERT_STATE(STATE_REPORT, "At REPORT");

    /* REPORT -> IDLE */
    state_machine_mock_set_report_done(true);
    state_machine_run();
    ASSERT_STATE(STATE_IDLE, "Back at IDLE");
}

/* ================================================================
 * Main
 * ================================================================ */

int main(void)
{
    printf("\n=== Pillbox State Machine Test Suite ===\n\n");

    /* State machine transition tests */
    RUN_TEST(init_starts_in_boot);
    RUN_TEST(boot_to_connect);
    RUN_TEST(connect_to_idle);
    RUN_TEST(idle_to_reminder);
    RUN_TEST(reminder_to_dispensing);
    RUN_TEST(dispensing_to_detect);
    RUN_TEST(detect_to_report);
    RUN_TEST(report_to_idle);
    RUN_TEST(any_state_to_error);
    RUN_TEST(error_to_idle_after_clear);
    RUN_TEST(invalid_transition_rejected);
    RUN_TEST(self_transition_allowed);
    RUN_TEST(error_preserved_until_clear);
    RUN_TEST(full_dispensing_cycle);

    /* Motor control tests */
    RUN_TEST(motor_not_ready);
    RUN_TEST(motor_home);

    /* TTS tests */
    RUN_TEST(tts_speak_null_safe);

    /* Summary */
    printf("\n=== Results ===\n");
    printf("  Total : %d\n", tests_run);
    printf("  Passed: %d\n", tests_run - tests_failed);
    printf("  Failed: %d\n", tests_failed);

    if (tests_failed > 0) {
        printf("\nSOME TESTS FAILED!\n");
        return 1;
    }
    printf("\nALL TESTS PASSED.\n");
    return 0;
}
#endif /* TEST_MODE */

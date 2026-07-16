/*
 * Eregen (颐贞) - Auto Dispensing Implementation
 * Complete dispensing flow with motor control, sensor verification,
 * and timeout-based detection.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "dispensing.h"
#include "opto_sensor.h"
#include "motor_control.h"

#include <stdio.h>
#include <string.h>

#ifdef TEST_MODE
#include <unistd.h>
#else
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_log.h"
#endif

static const char *TAG = "dispense";

/* Internal state */
static bool s_dispensing_active = false;
static uint32_t s_timeout_ms = DISPENSE_DEFAULT_TIMEOUT_MS;

/* Mock delay function — overridden by tests */
#ifdef TEST_MODE
static void (*s_mock_delay_fn)(uint32_t ms) = NULL;
static void (*s_mock_poll_hook)(void) = NULL;  /* called during wait_for_removal polls only */

void dispensing_set_mock_delay(void (*fn)(uint32_t))
{
    s_mock_delay_fn = fn;
}

/**
 * Mock: set a callback invoked on each poll iteration during
 * wait_for_removal(). Allows tests to inject state changes
 * (e.g., simulate medication being taken after N polls).
 * Note: this hook is ONLY called inside the wait_for_removal polling loop,
 * not during position verification or door operations.
 */
void dispensing_set_mock_poll_hook(void (*fn)(void))
{
    s_mock_poll_hook = fn;
}

static void mock_delay(uint32_t ms)
{
    if (s_mock_delay_fn) {
        s_mock_delay_fn(ms);
    }
    /* Default: no actual sleep needed in test mode.
     * Tests control opto_sensor_set_mock_state() directly. */
    (void)ms;
}
#else
#define mock_delay(ms) vTaskDelay(pdMS_TO_TICKS(ms))
#endif

/* ---- Forward declarations ---- */
static bool verify_compartment_position(void);
static void open_dispensing_door(void);
static void close_dispensing_door(void);
static dispense_result_t wait_for_removal(uint32_t timeout_ms);

/**
 * Dispense medication from a specific compartment.
 *
 * Complete dispensing flow:
 *   1. Validate inputs and transition state machine to DISPENSING
 *   2. Rotate motor to target compartment (iterative motor_control_step)
 *   3. Verify position via opto_sensor (compartment must have medication)
 *   4. Open dispensing door (reset sensor to waiting state)
 *   5. Poll opto_sensor until beam broken (medication removed) or timeout
 *   6. If timeout: retry once with half timeout, else report DISPENSE_TIMEOUT
 *   7. Close door, transition to REPORT state
 */
dispense_result_t dispense_medication(uint8_t compartment, uint8_t dose_count)
{
    /* Validate inputs */
    if (compartment >= DISPENSING_MAX_COMPARTMENTS) {
#ifdef TEST_MODE
        printf("[dispense] ERROR: compartment %u out of range\n",
               compartment);
#else
        ESP_LOGE(TAG, "Compartment %u out of range", compartment);
#endif
        return DISPENSE_TIMEOUT;
    }

    if (dose_count == 0 || dose_count > 4) {
#ifdef TEST_MODE
        printf("[dispense] ERROR: invalid dose count %u\n", dose_count);
#else
        ESP_LOGE(TAG, "Invalid dose count: %u", dose_count);
#endif
        return DISPENSE_TIMEOUT;
    }

    /* If already dispensing, cancel first */
    if (s_dispensing_active) {
        dispense_cancel();
    }

    s_dispensing_active = true;

    /* Step 1: Transition state machine to DISPENSING */
    state_machine_transition(STATE_DISPENSING);

    /* Step 2: Rotate motor to target compartment.
     * motor_control_step accepts uint8_t (max 255), so we call iteratively. */
#ifdef TEST_MODE
    printf("[dispense] Rotating to compartment %u\n", compartment);
#else
    ESP_LOGI(TAG, "Rotating to compartment %u", compartment);
#endif

    {
        uint32_t remaining = (uint32_t)compartment * STEPS_PER_COMPARTMENT;
        while (remaining > 0) {
            uint8_t chunk = (remaining > MOTOR_MAX_STEPS)
                                ? (uint8_t)(MOTOR_MAX_STEPS / 8)
                                : (uint8_t)(remaining / 8);
            if (!motor_control_step(chunk)) {
                state_machine_transition(STATE_ERROR);
                s_dispensing_active = false;
                return DISPENSE_JAM;
            }
            remaining -= ((uint32_t)chunk * 8);
        }
    }

    /* Step 3: Verify compartment position via sensor.
     * Sensor should indicate medication is PRESENT (beam intact = read returns false). */
    if (!verify_compartment_position()) {
#ifdef TEST_MODE
        printf("[dispense] Position verification failed\n");
#else
        ESP_LOGE(TAG, "Compartment position verification failed");
#endif
        state_machine_transition(STATE_ERROR);
        s_dispensing_active = false;
        return DISPENSE_EMPTY;
    }

    /* Step 4: Open dispensing door */
    open_dispensing_door();

#ifdef TEST_MODE
    printf("[dispense] Door opened, waiting up to %lus\n",
           (unsigned long)(s_timeout_ms / 1000));
#else
    ESP_LOGI(TAG, "Door opened, waiting up to %lus",
             (unsigned long)(s_timeout_ms / 1000));
#endif

    /* Step 5: Wait for user to take medication (with retry) */
    dispense_result_t result = wait_for_removal(s_timeout_ms);

    /* Step 6: Close door and clean up */
    close_dispensing_door();

    /* Transition to REPORT state */
    state_machine_transition(STATE_REPORT);

    s_dispensing_active = false;
    return result;
}

/**
 * Cancel an ongoing dispensing operation.
 */
bool dispense_cancel(void)
{
    if (!s_dispensing_active) {
        return true;
    }

    close_dispensing_door();
    state_machine_transition(STATE_IDLE);
    s_dispensing_active = false;

#ifdef TEST_MODE
    printf("[dispense] Cancelled\n");
#else
    ESP_LOGI(TAG, "Dispensing cancelled");
#endif

    return true;
}

/**
 * Get the configured timeout value.
 */
uint32_t dispense_get_timeout(void)
{
    return s_timeout_ms;
}

/**
 * Set the dispensing timeout value.
 */
void dispense_set_timeout(uint32_t ms)
{
    s_timeout_ms = ms;
}

/* ---- Internal helpers ---- */

/**
 * Verify that the correct compartment is aligned using the opto sensor.
 * Returns true if medication is detected as PRESENT (beam intact).
 */
static bool verify_compartment_position(void)
{
    mock_delay(200);

    /* Sensor returns true when beam is BROKEN (medication removed).
     * We want it present, so we expect false. */
    bool beam_broken = opto_sensor_read();
    return !beam_broken;
}

/**
 * Open the dispensing door.
 * Simulated by resetting the sensor to "waiting" state.
 */
static void open_dispensing_door(void)
{
    opto_sensor_reset();
    mock_delay(50);
}

/**
 * Close the dispensing door.
 */
static void close_dispensing_door(void)
{
    opto_sensor_reset();
    mock_delay(50);
}

/**
 * Wait for the user to remove medication from the open compartment.
 * Implements two-attempt strategy:
 *   Attempt 1: full timeout — poll sensor every 200ms
 *   Attempt 2: half timeout — reopen door briefly, then poll again
 *   If both fail: DISPENSE_TIMEOUT
 */
static dispense_result_t wait_for_removal(uint32_t timeout_ms)
{
    for (uint8_t attempt = 0; attempt < 2; attempt++) {
        uint32_t this_timeout = (attempt == 0) ? timeout_ms : timeout_ms / 2;

#ifdef TEST_MODE
        printf("[dispense] Waiting for removal (attempt %d, timeout=%lus)\n",
               attempt + 1, (unsigned long)(this_timeout / 1000));
#else
        ESP_LOGI(TAG, "Waiting for removal, attempt %d", attempt + 1);
#endif

        /* Poll sensor until beam is broken (medication removed) or timeout */
        uint32_t elapsed_ticks = 0;
        uint32_t tick_timeout = this_timeout / 200;  /* 200ms poll interval */

        while (elapsed_ticks < tick_timeout) {
            mock_delay(200);

#ifdef TEST_MODE
            if (s_mock_poll_hook) {
                s_mock_poll_hook();
            }
#endif

            if (opto_sensor_read()) {
                /* Beam broken — medication removed! */
#ifdef TEST_MODE
                printf("[dispense] Medication taken on attempt %d\n", attempt + 1);
#else
                ESP_LOGI(TAG, "Medication taken on attempt %d", attempt + 1);
#endif
                return DISPENSE_OK;
            }

            elapsed_ticks++;
        }

        /* Timeout on this attempt — retry once */
        if (attempt < 1) {
#ifdef TEST_MODE
            printf("[dispense] Timeout on attempt %d, reopening door\n",
                   attempt + 1);
#else
            ESP_LOGW(TAG, "Timeout on attempt %d, reopening", attempt + 1);
#endif
            open_dispensing_door();
        }
    }

#ifdef TEST_MODE
    printf("[dispense] Both attempts timed out — medication not taken\n");
#else
    ESP_LOGW(TAG, "Both attempts timed out");
#endif
    return DISPENSE_TIMEOUT;
}

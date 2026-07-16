/*
 * Eregen (颐贞) - Empty Compartment Detector Implementation
 * Detects empty compartments by rotating to each position and
 * checking the photoelectric sensor.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "empty_detector.h"
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

static const char *TAG = "empty_det";

/* Internal state */
static uint8_t s_empty_bitmap = 0;
static uint8_t s_empty_count  = 0;

#ifdef TEST_MODE
/* Mock status array — tracks which compartments are simulated as empty */
static bool s_mock_empty[EMPTY_DETECTOR_MAX_COMPARTMENTS];
static bool s_mock_initialized = false;

void empty_detector_set_mock_status(uint8_t compartment, bool empty)
{
    if (compartment >= EMPTY_DETECTOR_MAX_COMPARTMENTS) return;
    s_mock_empty[compartment] = empty;
    s_mock_initialized = true;
}

void empty_detector_clear_mock(void)
{
    memset(s_mock_empty, 0, sizeof(s_mock_empty));
    s_mock_initialized = false;
}
#endif

/* ---- Forward declarations ---- */
static void mock_delay(uint32_t ms);

/**
 * Check all compartments and build an empty-compartment bitmap.
 */
uint8_t empty_check_all_compartments(void)
{
    s_empty_bitmap = 0;
    s_empty_count  = 0;

#ifdef TEST_MODE
    printf("[empty_det] Checking all %d compartments...\n",
           EMPTY_DETECTOR_MAX_COMPARTMENTS);
#else
    ESP_LOGI(TAG, "Checking all %d compartments",
             EMPTY_DETECTOR_MAX_COMPARTMENTS);
#endif

    for (uint8_t i = 0; i < EMPTY_DETECTOR_MAX_COMPARTMENTS; i++) {
        bool is_empty = empty_check_single(i);
        if (is_empty) {
            s_empty_bitmap |= (1 << i);
            s_empty_count++;
        }
    }

#ifdef TEST_MODE
    if (s_empty_count > 0) {
        printf("[empty_det] Found %d empty compartment(s): bitmap=0x%02X\n",
               s_empty_count, s_empty_bitmap);
    } else {
        printf("[empty_det] All compartments have medication\n");
    }
#else
    if (s_empty_count > 0) {
        ESP_LOGW(TAG, "%d empty compartment(s) detected, bitmap=0x%02X",
                 s_empty_count, s_empty_bitmap);
    } else {
        ESP_LOGI(TAG, "All compartments filled");
    }
#endif

    return s_empty_bitmap;
}

/**
 * Check a single compartment for medication presence.
 * Rotates motor to the compartment, reads the sensor.
 *
 * Returns true if EMPTY (no medication detected).
 */
bool empty_check_single(uint8_t compartment)
{
    if (compartment >= EMPTY_DETECTOR_MAX_COMPARTMENTS) {
        return true;  /* Treat out-of-range as "empty" (error condition) */
    }

    /* Rotate motor to target compartment */
    uint32_t steps = (uint32_t)compartment * EMPTY_DETECTOR_STEPS_PER_COMPARTMENT;

    while (steps > 0) {
        uint8_t chunk = (steps > MOTOR_MAX_STEPS)
                            ? (uint8_t)(MOTOR_MAX_STEPS / 8)
                            : (uint8_t)(steps / 8);
        motor_control_step(chunk);
        steps -= ((uint32_t)chunk * 8);
    }

    mock_delay(200);  /* Let motor settle */

    /* Read sensor:
     * opto_sensor_read() returns true if beam is broken (medication removed).
     * In TEST_MODE with mock active, use per-compartment mock status instead. */
    bool beam_broken;
#ifdef TEST_MODE
    if (s_mock_initialized && s_mock_empty[compartment]) {
        beam_broken = true;  /* mock says empty */
    } else {
        beam_broken = opto_sensor_read();
    }
#else
    beam_broken = opto_sensor_read();
#endif

    /* If beam is broken, compartment is empty */
    if (beam_broken) {
#ifdef TEST_MODE
        printf("[empty_det] Compartment %u: EMPTY (beam broken)\n", compartment);
#else
        ESP_LOGI(TAG, "Compartment %u: EMPTY", compartment);
#endif
        return true;
    }

#ifdef TEST_MODE
    printf("[empty_det] Compartment %u: FILLED (beam intact)\n", compartment);
#else
    ESP_LOGI(TAG, "Compartment %u: FILLED", compartment);
#endif
    return false;
}

/**
 * Get the number of empty compartments.
 */
uint8_t empty_check_count(void)
{
    return s_empty_count;
}

/* ---- Internal helpers ---- */

static void mock_delay(uint32_t ms)
{
#ifdef TEST_MODE
    /* In test mode, check mock status instead of reading real sensor */
    if (s_mock_initialized) {
        /* When mock is active, sensor behavior is determined by
         * the compartment's mock status, not by opto_sensor_read().
         * This is handled in empty_check_single via direct mock access. */
    }
    (void)ms;
#else
    vTaskDelay(pdMS_TO_TICKS(ms));
#endif
}

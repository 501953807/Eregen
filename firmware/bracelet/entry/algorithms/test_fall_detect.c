/*
 * Eregen (颐贞) - Fall Detection Unit Tests
 * Tests for the rule-based fall detection algorithm using pre-populated windows.
 * Compile: gcc -DTEST_MODE -I. algorithms/sliding_window.c algorithms/fall_detect.c algorithms/test_fall_detect.c -lm -o test_fall_detect
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <stdint.h>
#include <stdbool.h>

#include "algorithms/sliding_window.h"
#include "algorithms/fall_detect.h"

/* ============ Test Helpers ============ */

static int g_tests_run = 0;
static int g_tests_passed = 0;
static int g_tests_failed = 0;

#define TEST_ASSERT(cond, msg) do { \
    g_tests_run++; \
    if (cond) { \
        g_tests_passed++; \
        printf("  PASS: %s\n", msg); \
    } else { \
        g_tests_failed++; \
        printf("  FAIL: %s\n", msg); \
    } \
} while(0)

#define TEST_ASSERT_NEAR(a, b, epsilon, msg) do { \
    g_tests_run++; \
    double diff = fabs((double)(a) - (double)(b)); \
    if (diff < (epsilon)) { \
        g_tests_passed++; \
        printf("  PASS: %s\n", msg); \
    } else { \
        g_tests_failed++; \
        printf("  FAIL: %s (expected~%.3f, got=%.3f, diff=%.3f)\n", \
               msg, (double)(b), (double)(a), diff); \
    } \
} while(0)

#define TEST_ASSERT_EQ(a, b, msg) do { \
    g_tests_run++; \
    if ((a) == (b)) { \
        g_tests_passed++; \
        printf("  PASS: %s\n", msg); \
    } else { \
        g_tests_failed++; \
        printf("  FAIL: %s (expected=%d, got=%d)\n", msg, (int)(b), (int)(a)); \
    } \
} while(0)

/*
 * Helper: populate a sliding window with repeated sample values.
 * @param sw Pointer to uninitialized window.
 * @param count Number of samples to push.
 * @param ax ay az gx gy gz Values to repeat for each sample.
 */
static void fill_window(sliding_window_t *sw, uint16_t count,
                        float ax, float ay, float az,
                        float gx, float gy, float gz)
{
    sw_init(sw);
    for (uint16_t i = 0; i < count; i++) {
        sw_push(sw, ax, ay, az, gx, gy, gz);
    }
}

/*
 * Helper: populate a window with a sequence of different values.
 */
static void fill_window_sequence(sliding_window_t *sw,
                                 const float *ax_arr,
                                 const float *ay_arr,
                                 const float *az_arr,
                                 const float *gx_arr,
                                 const float *gy_arr,
                                 const float *gz_arr,
                                 uint16_t count)
{
    sw_init(sw);
    for (uint16_t i = 0; i < count; i++) {
        sw_push(sw, ax_arr[i], ay_arr[i], az_arr[i],
                gx_arr[i], gy_arr[i], gz_arr[i]);
    }
}

/* ============ Test Cases ============ */

static void test_fall_detect_init(void)
{
    printf("\n--- Init Tests ---\n");

    fall_detect_init();

    /* After init, process should return NO_FALL for empty window */
    sliding_window_t sw;
    sw_init(&sw);
    fall_event_t event = fall_detect_process(&sw);

    TEST_ASSERT_EQ(event.result, NO_FALL, "Empty window returns NO_FALL");
    TEST_ASSERT_EQ(event.consecutive_detections, 0, "Counter starts at zero");
}

static void test_normal_walking(void)
{
    printf("\n--- Normal Walking Simulation ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* Walking: ~1g vertical (gravity) + small horizontal oscillation
     * + moderate gyro from arm swing. At 100Hz ODR, simulate 50 samples. */
    float ax[50], ay[50], az[50], gx[50], gy[50], gz[50];
    for (int i = 0; i < 50; i++) {
        float t = (float)i * 0.01f;  /* 10ms per sample */
        ax[i] = 0.1f * sinf(t * 4.0f * 3.14159f);   /* ±0.1g lateral sway */
        ay[i] = 0.05f * sinf(t * 4.0f * 3.14159f);  /* ±0.05g forward sway */
        az[i] = 1.0f + 0.05f * sinf(t * 8.0f * 3.14159f);  /* ~1g + bounce */
        gx[i] = 20.0f * sinf(t * 4.0f * 3.14159f);  /* arm swing gyro */
        gy[i] = 10.0f * sinf(t * 2.0f * 3.14159f);
        gz[i] = 5.0f * sinf(t * 4.0f * 3.14159f);
    }
    fill_window_sequence(&sw, ax, ay, az, gx, gy, gz, 50);

    fall_event_t event = fall_detect_process(&sw);
    TEST_ASSERT_EQ(event.result, NO_FALL,
                   "Normal walking returns NO_FALL");
    TEST_ASSERT(event.confidence < 0.7f,
                "Walking confidence below suspect threshold");
    TEST_ASSERT_EQ(event.consecutive_detections, 0,
                   "No consecutive detections during normal motion");
}

static void test_running(void)
{
    printf("\n--- Running Simulation ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* Running: higher magnitude oscillations but still within normal range.
     * Impact peaks reach ~2-3g briefly but no free-fall or post-impact static. */
    float ax[100], ay[100], az[100], gx[100], gy[100], gz[100];
    for (int i = 0; i < 100; i++) {
        float t = (float)i * 0.01f;
        /* Higher amplitude than walking */
        ax[i] = 0.3f * sinf(t * 6.0f * 3.14159f);
        ay[i] = 0.2f * sinf(t * 6.0f * 3.14159f);
        az[i] = 1.0f + 0.3f * sinf(t * 6.0f * 3.14159f);
        gx[i] = 50.0f * sinf(t * 6.0f * 3.14159f);
        gy[i] = 30.0f * sinf(t * 3.0f * 3.14159f);
        gz[i] = 15.0f * sinf(t * 6.0f * 3.14159f);
    }
    fill_window_sequence(&sw, ax, ay, az, gx, gy, gz, 100);

    fall_event_t event = fall_detect_process(&sw);
    TEST_ASSERT_EQ(event.result, NO_FALL,
                   "Running returns NO_FALL (no free-fall + static pattern)");
    TEST_ASSERT(event.confidence < 0.7f,
                "Running confidence below suspect threshold");
}

static void test_simulated_fall(void)
{
    printf("\n--- Simulated Fall (Impact -> Free-fall -> Static) ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* Fall simulation at 100Hz ODR:
     * Samples 0-9:   Normal standing (~1g on Z axis)
     * Samples 10-14: Impact spike (>1.5g combined)
     * Samples 15-24: Free-fall (<0.3g) for ~100ms minimum
     * Samples 25-74: Post-impact static (<0.2g) for ~500ms minimum
     * Samples 75-99: Still static
     */
    float ax[100] = {0};
    float ay[100] = {0};
    float az[100] = {0};
    float gx[100] = {0};
    float gy[100] = {0};
    float gz[100] = {0};

    /* Phase 1: Standing (samples 0-9) */
    for (int i = 0; i < 10; i++) {
        ax[i] = 0.0f;
        ay[i] = 0.0f;
        az[i] = 1.0f;  /* gravity */
    }

    /* Phase 2: Impact spike (samples 10-14) — sudden high acceleration */
    float impact_ax[] = { 1.8f, 2.0f, 1.5f, 1.2f, 0.8f };
    float impact_ay[] = { 0.5f, 0.8f, 1.0f, 0.6f, 0.3f };
    float impact_az[] = { 0.3f, 0.1f, 0.2f, 0.4f, 0.6f };
    for (int i = 0; i < 5; i++) {
        ax[10 + i] = impact_ax[i];
        ay[10 + i] = impact_ay[i];
        az[10 + i] = impact_az[i];
        /* High gyro during impact rotation */
        gx[10 + i] = 200.0f * (1.0f - (float)i / 5.0f);
        gy[10 + i] = 150.0f * (1.0f - (float)i / 5.0f);
        gz[10 + i] = 100.0f * (1.0f - (float)i / 5.0f);
    }

    /* Phase 3: Free-fall (samples 15-24) — near-zero acceleration */
    for (int i = 15; i < 25; i++) {
        ax[i] = 0.05f * sinf((float)i * 0.1f);
        ay[i] = 0.03f * cosf((float)i * 0.1f);
        az[i] = 0.1f * sinf((float)i * 0.15f);
        gx[i] = 50.0f * sinf((float)i * 0.2f);
        gy[i] = 30.0f * cosf((float)i * 0.2f);
        gz[i] = 20.0f * sinf((float)i * 0.25f);
    }

    /* Phase 4: Post-impact static (samples 25-99) — very low movement */
    for (int i = 25; i < 100; i++) {
        ax[i] = 0.02f * sinf((float)i * 0.05f);
        ay[i] = 0.01f * cosf((float)i * 0.05f);
        az[i] = 0.01f * sinf((float)i * 0.03f);
        gx[i] = 2.0f * sinf((float)i * 0.1f);
        gy[i] = 1.0f * cosf((float)i * 0.1f);
        gz[i] = 1.5f * sinf((float)i * 0.12f);
    }

    fill_window_sequence(&sw, ax, ay, az, gx, gy, gz, 100);

    fall_event_t event = fall_detect_process(&sw);

    TEST_ASSERT(event.result == FALL_DETECTED || event.result == FALL_SUSPECT,
                "Fall simulation produces SUSPECT or DETECTED result");
    TEST_ASSERT(event.confidence > 0.5f,
                "Fall simulation has meaningful confidence score");
    printf("  INFO: Fall confidence = %.3f\n", event.confidence);
}

static void test_rapid_head_bob_anti_misfire(void)
{
    printf("\n--- Rapid Head Bob (Anti-Misfire Test) ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* Rapid head nodding: creates brief accel spikes but NO free-fall
     * and NO prolonged static afterward. Should NOT trigger fall. */
    float ax[100], ay[100], az[100], gx[100], gy[100], gz[100];
    for (int i = 0; i < 100; i++) {
        float t = (float)i * 0.01f;
        /* Fast oscillation with occasional larger spikes */
        float spike = (i >= 40 && i <= 45) ? 1.8f : 0.0f;
        ax[i] = 0.2f * sinf(t * 10.0f * 3.14159f) + spike;
        ay[i] = 0.1f * cosf(t * 10.0f * 3.14159f);
        az[i] = 1.0f + 0.15f * sinf(t * 10.0f * 3.14159f);
        gx[i] = 80.0f * sinf(t * 10.0f * 3.14159f);
        gy[i] = 40.0f * cosf(t * 10.0f * 3.14159f);
        gz[i] = 30.0f * sinf(t * 10.0f * 3.14159f);
    }
    fill_window_sequence(&sw, ax, ay, az, gx, gy, gz, 100);

    fall_event_t event = fall_detect_process(&sw);
    TEST_ASSERT_EQ(event.result, NO_FALL,
                   "Rapid head bob returns NO_FALL (no free-fall + static)");
    TEST_ASSERT(event.confidence < 0.7f,
                "Head bob confidence below suspect threshold");
}

static void test_consecutive_detection_counter(void)
{
    printf("\n--- Consecutive Detection Counter ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* Create three separate fall-like windows and verify counter increments */
    float ax[100], ay[100], az[100], gx[100], gy[100], gz[100];

    /* Build a fall-like pattern */
    for (int i = 0; i < 100; i++) {
        if (i < 10) {
            /* Standing */
            ax[i] = 0.0f; ay[i] = 0.0f; az[i] = 1.0f;
            gx[i] = 0.0f; gy[i] = 0.0f; gz[i] = 0.0f;
        } else if (i < 15) {
            /* Impact */
            ax[i] = 1.8f; ay[i] = 0.5f; az[i] = 0.3f;
            gx[i] = 150.0f; gy[i] = 100.0f; gz[i] = 80.0f;
        } else if (i < 25) {
            /* Free-fall */
            ax[i] = 0.05f; ay[i] = 0.03f; az[i] = 0.1f;
            gx[i] = 30.0f; gy[i] = 20.0f; gz[i] = 15.0f;
        } else {
            /* Static on ground */
            ax[i] = 0.01f; ay[i] = 0.01f; az[i] = 0.02f;
            gx[i] = 1.0f; gy[i] = 0.5f; gz[i] = 0.8f;
        }
    }

    /* First detection */
    fill_window_sequence(&sw, ax, ay, az, gx, gy, gz, 100);
    fall_event_t e1 = fall_detect_process(&sw);
    TEST_ASSERT(e1.consecutive_detections >= 1,
                "First fall detection increments counter to >= 1");

    /* Second detection (same data, fresh call) */
    fall_event_t e2 = fall_detect_process(&sw);
    TEST_ASSERT(e2.consecutive_detections >= e1.consecutive_detections,
                "Second fall detection does not decrement counter");

    /* Third detection */
    fall_event_t e3 = fall_detect_process(&sw);
    TEST_ASSERT(e3.consecutive_detections >= 3,
                "Third consecutive detection reaches requirement of 3");
}

static void test_no_fall_resets_counter(void)
{
    printf("\n--- NO_FALL Resets Counter ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* Populate with normal walking data */
    fill_window(&sw, 50, 0.1f, 0.05f, 1.0f, 20.0f, 10.0f, 5.0f);

    /* Process multiple times with normal data */
    for (int i = 0; i < 10; i++) {
        fall_event_t event = fall_detect_process(&sw);
        TEST_ASSERT_EQ(event.result, NO_FALL,
                       "Normal data always returns NO_FALL");
        TEST_ASSERT_EQ(event.consecutive_detections, 0,
                       "NO_FALL resets counter to zero");
    }
}

static void test_insufficient_samples(void)
{
    printf("\n--- Insufficient Samples Test ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* Push only 2 samples — below minimum threshold */
    sw_init(&sw);
    sw_push(&sw, 1.8f, 0.5f, 0.3f, 150.0f, 100.0f, 80.0f);
    sw_push(&sw, 0.05f, 0.03f, 0.1f, 30.0f, 20.0f, 15.0f);

    fall_event_t event = fall_detect_process(&sw);
    TEST_ASSERT_EQ(event.result, NO_FALL,
                   "Window with < 3 samples returns NO_FALL");
    TEST_ASSERT_EQ(event.consecutive_detections, 0,
                   "Insufficient samples produce zero counter");
}

static void test_null_window(void)
{
    printf("\n--- Null Window Test ---\n");

    fall_detect_init();
    fall_event_t event = fall_detect_process(NULL);

    TEST_ASSERT_EQ(event.result, NO_FALL,
                   "NULL window returns NO_FALL");
    TEST_ASSERT_EQ(event.confidence, 0.0f,
                   "NULL window has zero confidence");
}

static void test_edge_case_zero_acceleration(void)
{
    printf("\n--- Zero Acceleration Edge Case ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* All zeros — could be sensor disconnected */
    fill_window(&sw, 50, 0.0f, 0.0f, 0.0f, 0.0f, 0.0f, 0.0f);

    fall_event_t event = fall_detect_process(&sw);
    /* With all zeros, acc mag is 0 which is < STATIC_THRESHOLD,
     * so it may register as "static after impact." But there's no
     * impact (no peak > 1.5g), so confidence should be low. */
    TEST_ASSERT(event.confidence < 0.7f || event.result != FALL_DETECTED,
                "All-zero data does not produce confirmed fall");
}

static void test_edge_case_high_gyro_only(void)
{
    printf("\n--- High Gyro Only Edge Case ---\n");

    fall_detect_init();
    sliding_window_t sw;

    /* High gyro but normal accel — spinning in chair */
    fill_window(&sw, 50, 0.0f, 0.0f, 1.0f, 500.0f, 500.0f, 500.0f);

    fall_event_t event = fall_detect_process(&sw);
    TEST_ASSERT_EQ(event.result, NO_FALL,
                   "High gyro alone (no impact+freefall+static) returns NO_FALL");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen - Fall Detection Algorithm Tests\n");
    printf("Target: GD32E230C8T3 / FreeRTOS\n");
    printf("Mode: Host simulation (pre-populated windows)\n");
    printf("========================================\n");

    test_fall_detect_init();
    test_normal_walking();
    test_running();
    test_simulated_fall();
    test_rapid_head_bob_anti_misfire();
    test_consecutive_detection_counter();
    test_no_fall_resets_counter();
    test_insufficient_samples();
    test_null_window();
    test_edge_case_zero_acceleration();
    test_edge_case_high_gyro_only();

    printf("\n========================================\n");
    printf("Test Results: %d/%d passed (%d failed)\n",
           g_tests_passed, g_tests_run, g_tests_failed);
    printf("========================================\n");

    return (g_tests_failed > 0) ? 1 : 0;
}

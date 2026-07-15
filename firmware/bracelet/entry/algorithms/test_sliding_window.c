/*
 * Eregen (颐贞) - Sliding Window Unit Tests
 * Tests for circular buffer behavior with known values
 * Compile: gcc -DTEST_MODE -I. algorithms/sliding_window.c algorithms/test_sliding_window.c -o test_sliding_window
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

/* ============ Test Cases ============ */

static void test_sw_init(void)
{
    printf("\n--- Sliding Window Init Tests ---\n");

    sliding_window_t sw;
    sw_init(&sw);

    TEST_ASSERT_EQ(sw.count, 0, "New window has zero count");
    TEST_ASSERT_EQ(sw.head, 0, "New window has zero head");
    TEST_ASSERT_NEAR(sw.ax[0], 0.0f, 0.001f, "Initial ax[0] is zero");
    TEST_ASSERT_NEAR(sw.ay[0], 0.0f, 0.001f, "Initial ay[0] is zero");
    TEST_ASSERT_NEAR(sw.az[0], 0.0f, 0.001f, "Initial az[0] is zero");
}

static void test_sw_push_basic(void)
{
    printf("\n--- Basic Push Tests ---\n");

    sliding_window_t sw;
    sw_init(&sw);

    bool ok = sw_push(&sw, 1.0f, 2.0f, 3.0f, 10.0f, 20.0f, 30.0f);
    TEST_ASSERT(ok == true, "Push returns true on success");
    TEST_ASSERT_EQ(sw_count(&sw), 1, "Count is 1 after one push");
    TEST_ASSERT_NEAR(sw.ax[0], 1.0f, 0.001f, "ax[0] stores correct value");
    TEST_ASSERT_NEAR(sw.ay[0], 2.0f, 0.001f, "ay[0] stores correct value");
    TEST_ASSERT_NEAR(sw.az[0], 3.0f, 0.001f, "az[0] stores correct value");
    TEST_ASSERT_NEAR(sw.gx[0], 10.0f, 0.001f, "gx[0] stores correct value");
    TEST_ASSERT_NEAR(sw.gy[0], 20.0f, 0.001f, "gy[0] stores correct value");
    TEST_ASSERT_NEAR(sw.gz[0], 30.0f, 0.001f, "gz[0] stores correct value");

    sw_push(&sw, 4.0f, 5.0f, 6.0f, 40.0f, 50.0f, 60.0f);
    TEST_ASSERT_EQ(sw_count(&sw), 2, "Count is 2 after two pushes");
    TEST_ASSERT_NEAR(sw.ax[1], 4.0f, 0.001f, "ax[1] stores second value");
    TEST_ASSERT_NEAR(sw.ax[0], 1.0f, 0.001f, "ax[0] still holds first value");
}

static void test_sw_circular_wrap(void)
{
    printf("\n--- Circular Wrap Tests ---\n");

    sliding_window_t sw;
    sw_init(&sw);

    /* Fill exactly to capacity */
    for (uint16_t i = 0; i < SW_MAX_SAMPLES; i++) {
        sw_push(&sw, (float)i, (float)i*2, (float)i*3,
                (float)i*10, (float)i*20, (float)i*30);
    }
    TEST_ASSERT_EQ(sw_count(&sw), SW_MAX_SAMPLES,
                   "Count equals MAX_SAMPLES when full");

    /* Push one more — should overwrite index 0 */
    sw_push(&sw, 999.0f, 888.0f, 777.0f, 111.0f, 222.0f, 333.0f);
    TEST_ASSERT_EQ(sw_count(&sw), SW_MAX_SAMPLES,
                   "Count stays at MAX_SAMPLES after overflow");
    TEST_ASSERT_NEAR(sw.ax[0], 999.0f, 0.001f,
                     "First slot overwritten by newest sample");
    TEST_ASSERT_NEAR(sw.ax[1], 1.0f, 0.001f,
                     "Second slot now holds what was previously at index 1");
    TEST_ASSERT_NEAR(sw.ax[SW_MAX_SAMPLES - 1], 99.0f, 0.001f,
                     "Last slot holds value from 100th push (never overwritten)");
}

static void test_sw_null_safety(void)
{
    printf("\n--- Null Safety Tests ---\n");

    TEST_ASSERT(sw_push(NULL, 1.0f, 2.0f, 3.0f, 4.0f, 5.0f, 6.0f) == false,
                "Push with NULL returns false");
    TEST_ASSERT_EQ(sw_count(NULL), 0, "Count with NULL returns 0");
    TEST_ASSERT(sw_get_ax(NULL) == NULL, "Get AX with NULL returns NULL");
}

static void test_sw_get_ax(void)
{
    printf("\n--- sw_get_ax Tests ---\n");

    sliding_window_t sw;
    sw_init(&sw);

    sw_push(&sw, 5.5f, 0.0f, 0.0f, 0.0f, 0.0f, 0.0f);
    sw_push(&sw, 6.6f, 0.0f, 0.0f, 0.0f, 0.0f, 0.0f);

    const float* ax = sw_get_ax(&sw);
    TEST_ASSERT(ax != NULL, "sw_get_ax returns non-NULL pointer");
    TEST_ASSERT_NEAR(ax[0], 5.5f, 0.001f, "ax[0] accessible via pointer");
    TEST_ASSERT_NEAR(ax[1], 6.6f, 0.001f, "ax[1] accessible via pointer");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen - Sliding Window Tests\n");
    printf("Mode: Host simulation\n");
    printf("========================================\n");

    test_sw_init();
    test_sw_push_basic();
    test_sw_circular_wrap();
    test_sw_null_safety();
    test_sw_get_ax();

    printf("\n========================================\n");
    printf("Test Results: %d/%d passed (%d failed)\n",
           g_tests_passed, g_tests_run, g_tests_failed);
    printf("========================================\n");

    return (g_tests_failed > 0) ? 1 : 0;
}

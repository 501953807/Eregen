/*
 * Eregen (颐贞) - Ring Buffer Test Harness
 * Host-compiled test driver for ring_buffer module.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>
#include "../common/ring_buffer.h"

static int passed = 0;
static int failed = 0;

static void check(bool cond, const char *label)
{
    if (cond) {
        printf("  PASS: %s\n", label);
        passed++;
    } else {
        printf("  FAIL: %s\n", label);
        failed++;
    }
}

int main(void)
{
    printf("Ring buffer tests:\n");

    ring_buf_t rb;
    ring_buf_init(&rb);

    /* Empty buffer */
    check(ring_buf_available(&rb) == 0, "initial count is 0");
    check(!ring_buf_pop(&rb, NULL), "pop on empty returns false");

    /* Fill one item */
    check(ring_buf_push(&rb, 0x42), "push single byte");
    check(ring_buf_available(&rb) == 1, "count is 1 after push");

    uint8_t val;
    check(ring_buf_pop(&rb, &val) && val == 0x42, "pop returns correct value");
    check(ring_buf_available(&rb) == 0, "count is 0 after pop");

    /* Fill and drain multiple items */
    for (uint16_t i = 0; i < 100; i++) {
        check(ring_buf_push(&rb, (uint8_t)i), "push 100 items");
    }
    check(ring_buf_available(&rb) == 100, "count is 100");

    for (uint16_t i = 0; i < 100; i++) {
        uint8_t v;
        check(ring_buf_pop(&rb, &v) && v == (uint8_t)i, "drain 100 items in order");
    }
    check(ring_buf_available(&rb) == 0, "count is 0 after drain");

    /* Fill to capacity */
    for (uint16_t i = 0; i < RB_MAX_SIZE; i++) {
        check(ring_buf_push(&rb, (uint8_t)i), "push to capacity");
    }
    check(!ring_buf_push(&rb, 0xFF), "push on full returns false");

    /* Wrap-around test: drain half, push more, wrap */
    for (uint16_t i = 0; i < RB_MAX_SIZE / 2; i++) {
        ring_buf_pop(&rb, NULL);
    }
    for (uint16_t i = 0; i < RB_MAX_SIZE / 2; i++) {
        check(ring_buf_push(&rb, (uint8_t)(i + 128)), "wrap-around push");
    }
    check(!ring_buf_push(&rb, 0xFF), "wrap-around full returns false");

    /* Drain all remaining */
    uint8_t drain_val;
    for (int i = 0; i < RB_MAX_SIZE / 2; i++) {
        check(ring_buf_pop(&rb, &drain_val), "drain wrapped items");
        check(drain_val == (uint8_t)(i + 128), "wrapped values match");
    }

    printf("\nResults: %d passed, %d failed\n", passed, failed);
    return failed > 0 ? 1 : 0;
}

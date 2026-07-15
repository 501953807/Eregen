/*
 * Eregen (颐贞) - Ring Buffer Implementation
 * Simple circular buffer with volatile access for interrupt safety.
 * Not fully lock-free (no atomic ops needed for single-producer/single-consumer).
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "ring_buffer.h"

#ifdef TEST_MODE
int ring_buffer_test(void);
int main(void) { return ring_buffer_test(); }
#endif

void ring_buf_init(ring_buf_t *rb)
{
    rb->head = 0;
    rb->tail = 0;
    rb->count = 0;
}

bool ring_buf_push(ring_buf_t *rb, uint8_t data)
{
    if (rb->count >= RB_MAX_SIZE) {
        return false;  /* Buffer full */
    }

    rb->buf[rb->head] = data;
    rb->head = (rb->head + 1) % RB_MAX_SIZE;
    rb->count++;
    return true;
}

bool ring_buf_pop(ring_buf_t *rb, uint8_t *data)
{
    if (rb->count == 0) {
        return false;  /* Buffer empty */
    }

    if (data) {
        *data = rb->buf[rb->tail];
    }
    rb->tail = (rb->tail + 1) % RB_MAX_SIZE;
    rb->count--;
    return true;
}

uint16_t ring_buf_available(const ring_buf_t *rb)
{
    return rb->count;
}

#ifdef TEST_MODE
#include <stdio.h>
#include <string.h>

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

int ring_buffer_test(void)
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
#endif

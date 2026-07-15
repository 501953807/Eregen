/*
 * Eregen (颐贞) - Ring Buffer Header
 * Thread-safe lock-free ring buffer for embedded use.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef RING_BUFFER_H
#define RING_BUFFER_H

#include <stdint.h>
#include <stdbool.h>

#define RB_MAX_SIZE 256

/**
 * Ring buffer structure.
 * head points to next write position, tail to next read position.
 * count tracks current number of items (for O(1) available check).
 */
typedef struct {
    uint8_t buf[RB_MAX_SIZE];
    volatile uint16_t head;
    volatile uint16_t tail;
    volatile uint16_t count;
} ring_buf_t;

/**
 * Initialize a ring buffer to empty state.
 * @param rb Pointer to ring buffer instance.
 */
void ring_buf_init(ring_buf_t *rb);

/**
 * Push a single byte into the ring buffer.
 * @param rb   Pointer to ring buffer.
 * @param data Byte to push.
 * @return true if pushed successfully, false if full.
 */
bool ring_buf_push(ring_buf_t *rb, uint8_t data);

/**
 * Pop a single byte from the ring buffer.
 * @param rb   Pointer to ring buffer.
 * @param data Output pointer for popped byte.
 * @return true if popped successfully, false if empty.
 */
bool ring_buf_pop(ring_buf_t *rb, uint8_t *data);

/**
 * Get the number of available bytes in the buffer.
 * @param rb Pointer to ring buffer.
 * @return Current count of items in buffer.
 */
uint16_t ring_buf_available(const ring_buf_t *rb);

#endif /* RING_BUFFER_H */

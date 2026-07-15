/*
 * Eregen (颐贞) - Sliding Window Data Structure
 * Fixed-size circular buffer for IMU sample accumulation
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef SLIDING_WINDOW_H
#define SLIDING_WINDOW_H

#include <stdint.h>
#include <stdbool.h>

#define SW_MAX_SAMPLES 100

typedef struct {
    float ax[SW_MAX_SAMPLES];
    float ay[SW_MAX_SAMPLES];
    float az[SW_MAX_SAMPLES];
    float gx[SW_MAX_SAMPLES];
    float gy[SW_MAX_SAMPLES];
    float gz[SW_MAX_SAMPLES];
    volatile uint16_t count;
    volatile uint16_t head;
} sliding_window_t;

/*
 * Initialize a sliding window to empty state.
 * @param sw Pointer to the sliding window instance.
 */
void sw_init(sliding_window_t *sw);

/*
 * Push a new IMU sample into the sliding window.
 * Overwrites oldest data when full (circular buffer behavior).
 * @return true on success, false if sw is NULL.
 */
bool sw_push(sliding_window_t *sw, float ax, float ay, float az,
             float gx, float gy, float gz);

/*
 * Get the current number of samples in the window.
 * Returns min(count, SW_MAX_SAMPLES).
 */
uint16_t sw_count(const sliding_window_t *sw);

/*
 * Get read-only pointer to acceleration X array.
 * Indices 0..sw_count()-1 are valid.
 */
const float* sw_get_ax(const sliding_window_t *sw);

#endif /* SLIDING_WINDOW_H */

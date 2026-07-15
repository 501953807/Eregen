/*
 * Eregen (颐贞) - Sliding Window Implementation
 * Fixed-size circular buffer for IMU sample accumulation
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "algorithms/sliding_window.h"
#include <stddef.h>

void sw_init(sliding_window_t *sw)
{
    if (sw == NULL) return;
    sw->count = 0;
    sw->head = 0;
    /* Zero-fill arrays for clean initial state */
    for (uint16_t i = 0; i < SW_MAX_SAMPLES; i++) {
        sw->ax[i] = 0.0f;
        sw->ay[i] = 0.0f;
        sw->az[i] = 0.0f;
        sw->gx[i] = 0.0f;
        sw->gy[i] = 0.0f;
        sw->gz[i] = 0.0f;
    }
}

bool sw_push(sliding_window_t *sw, float ax, float ay, float az,
             float gx, float gy, float gz)
{
    if (sw == NULL) return false;

    uint16_t idx = sw->head % SW_MAX_SAMPLES;

    sw->ax[idx] = ax;
    sw->ay[idx] = ay;
    sw->az[idx] = az;
    sw->gx[idx] = gx;
    sw->gy[idx] = gy;
    sw->gz[idx] = gz;

    sw->head++;
    if (sw->head > sw->count) {
        sw->count = sw->head;
        if (sw->count > SW_MAX_SAMPLES) {
            sw->count = SW_MAX_SAMPLES;
        }
    }

    return true;
}

uint16_t sw_count(const sliding_window_t *sw)
{
    if (sw == NULL) return 0;
    return sw->count;
}

const float* sw_get_ax(const sliding_window_t *sw)
{
    if (sw == NULL) return NULL;
    return sw->ax;
}

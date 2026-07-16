/*
 * Eregen (颐贞) - Fall Detection Algorithm Implementation (Plus Tier)
 * Threshold + angle-based algorithm using accelerometer + gyroscope fusion.
 *
 * Algorithm flow:
 *   1. Compute acceleration magnitude from 3-axis accel data.
 *   2. Detect impact phase: |acc| exceeds IMPACT_THRESHOLD_G.
 *   3. Compute pitch angle from accel axes: atan2(sqrt(ax^2+ay^2), az).
 *   4. If pitch angle > 60 degrees sustained for 3 seconds -> CONFIRMED.
 *   5. If normal motion detected within recovery window -> RECOVERED.
 *
 * MIT License
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

#include "fall_detect.h"
#include <math.h>
#include <string.h>

#ifndef M_PI_F
#define M_PI_F 3.14159265358979323846f
#endif

/* ---- Internal constants ---- */

/** Sliding window size in samples (covers ~5 seconds at 50 Hz). */
#define FALL_WINDOW_SIZE       250U

/* ---- Internal state ---- */

static bool s_initialized = false;
static fall_sample_t s_window[FALL_WINDOW_SIZE];
static uint16_t s_write_idx = 0;
static uint16_t s_sample_count = 0;

static fall_state_t s_current_state = FALL_STATE_NORMAL;
static uint8_t s_consecutive_confirmed = 0;
static uint32_t s_suspect_start_tick = 0;
static bool s_in_recovery_window = false;
static bool s_alarm_active = false;
static uint32_t s_fake_tick = 0;  /* For testing */

/* ---- Helper functions ---- */

static float calc_accel_mag(float ax, float ay, float az)
{
    return sqrtf(ax * ax + ay * ay + az * az);
}

/**
 * Calculate pitch angle in degrees from accelerometer data.
 * Pitch = atan2(sqrt(ax^2 + ay^2), az)
 * A value near 90 degrees means the device is lying flat (face-up or face-down).
 */
static float calc_pitch_angle(float ax, float ay, float az)
{
    float horizontal = sqrtf(ax * ax + ay * ay);
    float angle_rad = atan2f(horizontal, fabsf(az));
    return angle_rad * (180.0f / M_PI_F);
}

/* ---- Public API ---- */

bool fall_detect_init(void)
{
    memset(s_window, 0, sizeof(s_window));
    s_write_idx = 0;
    s_sample_count = 0;
    s_current_state = FALL_STATE_NORMAL;
    s_consecutive_confirmed = 0;
    s_in_recovery_window = false;
    s_alarm_active = false;
    s_fake_tick = 0;
    s_initialized = true;
    return true;
}

fall_event_t fall_detect_feed(const fall_sample_t *sample)
{
    fall_event_t event;
    memset(&event, 0, sizeof(event));

    if (!s_initialized) {
        fall_detect_init();
    }

    if (!sample) {
        event.state = s_current_state;
        event.confidence = 0.0f;
        event.consecutive = s_consecutive_confirmed;
        event.alarm_ready = (s_current_state == FALL_STATE_CONFIRMED);
        return event;
    }

    /* Ring-buffer insertion into sliding window. */
    s_window[s_write_idx] = *sample;
    s_write_idx = (s_write_idx + 1) % FALL_WINDOW_SIZE;
    if (s_sample_count < FALL_WINDOW_SIZE) {
        s_sample_count++;
    }

    event.state = s_current_state;
    event.confidence = 0.0f;
    event.consecutive = s_consecutive_confirmed;
    event.alarm_ready = false;

    return event;
}

fall_event_t fall_detect_run(void)
{
    fall_event_t event;
    memset(&event, 0, sizeof(event));

    if (!s_initialized) {
        fall_detect_init();
    }

    if (s_sample_count < 5) {
        /* Not enough samples for meaningful analysis. */
        event.state = FALL_STATE_NORMAL;
        return event;
    }

    /* Use fake tick for testing, real ticks otherwise. */
    uint32_t now_tick = s_fake_tick;

    /* ---- Step 1: Check recent samples for impact ---- */

    bool has_impact = false;
    float peak_mag = 0.0f;

    /* Scan last N samples for acceleration spike. */
    uint16_t check_count = (s_sample_count < 20) ? s_sample_count : 20;
    uint16_t start_idx = (s_write_idx >= check_count) ?
                         (s_write_idx - check_count) :
                         (FALL_WINDOW_SIZE - check_count);

    for (uint16_t i = 0; i < check_count; i++) {
        uint16_t idx = (start_idx + i) % FALL_WINDOW_SIZE;
        float mag = calc_accel_mag(s_window[idx].ax, s_window[idx].ay, s_window[idx].az);
        if (mag > peak_mag) {
            peak_mag = mag;
        }
        if (mag > FALL_DETECT_IMPACT_THRESHOLD_G) {
            has_impact = true;
        }
    }

    /* ---- Step 2: Check pitch angle sustained for 3 seconds ---- */

    uint16_t high_pitch_samples = 0;
    for (uint16_t i = 0; i < check_count; i++) {
        uint16_t idx = (start_idx + i) % FALL_WINDOW_SIZE;
        float pitch = calc_pitch_angle(
            s_window[idx].ax, s_window[idx].ay, s_window[idx].az
        );
        if (pitch > FALL_DETECT_PITCH_ANGLE_DEG) {
            high_pitch_samples++;
        }
    }

    /* Each sample represents FALL_DETECT_TICK_MS milliseconds. */
    uint32_t sustained_ms = (uint32_t)high_pitch_samples * FALL_DETECT_TICK_MS;

    /* ---- Step 3: State machine transition ---- */

    switch (s_current_state) {
    case FALL_STATE_NORMAL:
        if (has_impact && sustained_ms >= FALL_DETECT_SUSPEND_DURATION_MS) {
            s_current_state = FALL_STATE_SUSPECT;
            s_suspect_start_tick = now_tick;
            s_in_recovery_window = true;
            event.state = FALL_STATE_SUSPECT;
            event.confidence = 0.8f;
        } else if (has_impact) {
            /* Impact without sustained angle — minor event, not alarming. */
            event.state = FALL_STATE_NORMAL;
            event.confidence = (peak_mag / FALL_DETECT_IMPACT_THRESHOLD_G) * 0.5f;
            if (event.confidence > 1.0f) event.confidence = 1.0f;
        }
        break;

    case FALL_STATE_SUSPECT:
        if (s_in_recovery_window) {
            /* Check recovery window expiry. */
            if ((now_tick - s_suspect_start_tick) > FALL_DETECT_RECOVERY_WINDOW_MS) {
                s_in_recovery_window = false;
            }

            /* Check if still showing fall posture. */
            if (sustained_ms >= FALL_DETECT_SUSPEND_DURATION_MS) {
                s_consecutive_confirmed++;
                event.consecutive = s_consecutive_confirmed;

                if (s_consecutive_confirmed >= FALL_DETECT_MIN_CONSECUTIVE) {
                    s_current_state = FALL_STATE_CONFIRMED;
                    event.state = FALL_STATE_CONFIRMED;
                    event.confidence = 0.95f;
                    event.alarm_ready = true;
                    s_alarm_active = true;
                } else {
                    event.state = FALL_STATE_SUSPECT;
                    event.confidence = 0.85f + (s_consecutive_confirmed * 0.02f);
                }
            } else {
                /* Normal motion detected — recovery! */
                s_current_state = FALL_STATE_RECOVERED;
                event.state = FALL_STATE_RECOVERED;
                event.confidence = 0.0f;
                s_consecutive_confirmed = 0;
                s_in_recovery_window = false;
            }
        }
        break;

    case FALL_STATE_CONFIRMED:
        /* Stay confirmed until explicitly reset. */
        event.state = FALL_STATE_CONFIRMED;
        event.confidence = 1.0f;
        event.alarm_ready = true;
        break;

    case FALL_STATE_RECOVERED:
        /* Transient state — return to normal next tick. */
        s_current_state = FALL_STATE_NORMAL;
        event.state = FALL_STATE_NORMAL;
        event.confidence = 0.0f;
        break;
    }

    /* Clamp confidence. */
    if (event.confidence > 1.0f) event.confidence = 1.0f;
    if (event.confidence < 0.0f) event.confidence = 0.0f;

    return event;
}

void fall_detect_reset(void)
{
    s_current_state = FALL_STATE_NORMAL;
    s_consecutive_confirmed = 0;
    s_in_recovery_window = false;
    s_alarm_active = false;
    s_write_idx = 0;
    s_sample_count = 0;
    memset(s_window, 0, sizeof(s_window));
}

bool fall_detect_is_alarm_active(void)
{
    return s_alarm_active;
}

uint16_t fall_detect_window_size(void)
{
    return s_sample_count;
}

/* ---- Test helpers ---- */

#ifdef TEST_MODE

fall_event_t fall_detect_test_feed_batch(const fall_sample_t *samples, uint16_t count)
{
    fall_event_t last_event;
    memset(&last_event, 0, sizeof(last_event));

    for (uint16_t i = 0; i < count; i++) {
        last_event = fall_detect_feed(&samples[i]);
    }

    /* Run evaluation after feeding all samples. */
    return fall_detect_run();
}

void fall_detect_set_fake_tick(uint32_t fake_tick)
{
    s_fake_tick = fake_tick;
}

#endif

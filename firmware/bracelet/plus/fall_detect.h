/*
 * Eregen (颐贞) - Fall Detection Algorithm Header (Plus Tier)
 * Enhanced IMU-based fall detection using accelerometer + gyroscope fusion.
 * Uses threshold-based algorithm:
 *   - Acceleration magnitude > 1.5g (impact phase)
 *   - Pitch angle > 60 degrees sustained for 3 seconds (post-fall static)
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
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

#ifndef FALL_DETECT_PLUS_H
#define FALL_DETECT_PLUS_H

#include <stdint.h>
#include <stdbool.h>

/* ---- Algorithm thresholds (tuned for elderly fall patterns) ---- */

/** Impact threshold: acceleration magnitude must exceed this (in g). */
#define FALL_DETECT_IMPACT_THRESHOLD_G    1.5f

/** Pitch angle threshold: body must be tilted beyond this (degrees). */
#define FALL_DETECT_PITCH_ANGLE_DEG       60.0f

/** Sustained duration: angle must remain above threshold for this long (ms). */
#define FALL_DETECT_SUSPEND_DURATION_MS   3000U

/** Free-fall threshold: acceleration drops below this during the fall (in g). */
#define FALL_DETECT_FREEFALL_THRESHOLD_G  0.3f

/** Recovery window: if normal motion detected within this period after suspect, cancel. */
#define FALL_DETECT_RECOVERY_WINDOW_MS    8000U

/** Minimum consecutive detections before raising an alarm. */
#define FALL_DETECT_MIN_CONSECUTIVE       3U

/** IMU sample rate assumed by the algorithm (Hz). */
#define FALL_DETECT_IMU_ODR_HZ            50U

/** Tick interval used by fall_detect_run() (milliseconds). */
#define FALL_DETECT_TICK_MS               100U

/* ---- Result enumeration ---- */

typedef enum {
    FALL_STATE_NORMAL    = 0,   /* No fall indicators */
    FALL_STATE_SUSPECT,        /* Fall indicators present, waiting */
    FALL_STATE_CONFIRMED,      /* Confirmed fall — alarm ready */
    FALL_STATE_RECOVERED,      /* Normal motion detected after suspect */
} fall_state_t;

/**
 * Single IMU sample as consumed by the algorithm.
 */
typedef struct {
    float ax;       /* Acceleration X axis, in g */
    float ay;       /* Acceleration Y axis, in g */
    float az;       /* Acceleration Z axis, in g */
    float gx;       /* Gyro X axis, in dps */
    float gy;       /* Gyro Y axis, in dps */
    float gz;       /* Gyro Z axis, in dps */
    uint32_t tick;  /* FreeRTOS tick count at time of sample */
} fall_sample_t;

/**
 * Fall detection event output.
 */
typedef struct {
    fall_state_t  state;        /* Current classification */
    float         confidence;   /* 0.0 – 1.0 weighted score */
    uint8_t       consecutive;  /* Consecutive confirmed frames */
    bool          alarm_ready;  /* True if a fall alarm should be raised */
} fall_event_t;

/*
 * Initialize the fall detection engine.
 * Resets internal state and allocates the sliding window buffer.
 * @return true on success.
 */
bool fall_detect_init(void);

/*
 * Feed a single IMU sample into the fall detection pipeline.
 * This function is called once per IMU ODR tick from the sensor task.
 * @param sample Pointer to calibrated IMU sample.
 * @return fall_event_t with current classification.
 */
fall_event_t fall_detect_feed(const fall_sample_t *sample);

/*
 * Periodic tick function — call every FALL_DETECT_TICK_MS from RTOS task.
 * Evaluates the accumulated samples and advances state machine.
 * @return fall_event_t with current classification (same as feed but
 *         triggered by time rather than sample arrival).
 */
fall_event_t fall_detect_run(void);

/*
 * Reset the fall detection state machine.
 * Call after a confirmed false positive or when the device is picked up.
 */
void fall_detect_reset(void);

/*
 * Check whether a fall alarm has been raised since last reset.
 * @return true if a confirmed fall was detected.
 */
bool fall_detect_is_alarm_active(void);

/*
 * Get the number of samples currently buffered.
 * Useful for diagnostics.
 * @return Sample count in the sliding window.
 */
uint16_t fall_detect_window_size(void);

/* ---- Test helpers ---- */

#ifdef TEST_MODE
/* Feed a batch of pre-crafted samples and return the last event. */
fall_event_t fall_detect_test_feed_batch(const fall_sample_t *samples, uint16_t count);

/* Set a fake tick counter for time-based evaluation. */
void fall_detect_set_fake_tick(uint32_t fake_tick);
#endif

#endif /* FALL_DETECT_PLUS_H */

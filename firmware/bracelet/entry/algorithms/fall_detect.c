/*
 * Eregen (颐贞) - Fall Detection Algorithm Implementation
 * Rule-based fall detection using acceleration magnitude, angular velocity,
 * impact detection, free-fall detection, and post-impact static analysis.
 *
 * Algorithm flow:
 *   1. Extract features from sliding window (magnitude, impulse, free-fall, static)
 *   2. Compute weighted confidence score
 *   3. Apply anti-misfire: require 3 consecutive detections before alarming
 *   4. Recovery window: if normal motion detected after SUSPECT, cancel
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "algorithms/fall_detect.h"
#include <math.h>
#include <string.h>

/* Internal state */
static bool s_initialized = false;
static uint8_t s_consecutive_counter;
static fall_result_t s_last_result;
static uint32_t s_suspect_start_ms;
static bool s_in_recovery_window;

/*
 * Calculate acceleration magnitude from three axes.
 */
static float calc_accel_magnitude(float ax, float ay, float az)
{
    return sqrtf(ax * ax + ay * ay + az * az);
}

/*
 * Calculate angular velocity magnitude from three axes.
 */
static float calc_gyro_magnitude(float gx, float gy, float gz)
{
    return sqrtf(gx * gx + gy * gy + gz * gz);
}

/*
 * Detect if any sample in the window shows an impact event.
 * Impact = acceleration magnitude spikes above IMPACT_THRESHOLD.
 * Returns true if found, fills *peak_mag with the maximum magnitude seen.
 */
static bool detect_impact(const sliding_window_t *sw, float *peak_mag)
{
    uint16_t n = sw_count(sw);
    if (n == 0) return false;

    const float *ax = sw_get_ax(sw);
    float max_mag = 0.0f;

    for (uint16_t i = 0; i < n; i++) {
        /* We only have ax array directly accessible; reconstruct magnitude
         * from stored values. For a complete implementation we'd access all
         * axes. Use ax as proxy when only one axis is available. */
        float mag = fabsf(ax[i]);
        if (mag > max_mag) max_mag = mag;
    }

    /* Also scan the full window if all axes are accessible via the struct */
    for (uint16_t i = 0; i < n; i++) {
        float m = calc_accel_magnitude(sw->ax[i], sw->ay[i], sw->az[i]);
        if (m > max_mag) max_mag = m;
    }

    *peak_mag = max_mag;
    return max_mag > IMPACT_THRESHOLD;
}

/*
 * Detect free-fall phase: acceleration magnitude drops below threshold
 * for a sustained period (IMPULSE_DURATION_MS worth of samples).
 * Assumes ODR = 100 Hz, so each sample = 10ms.
 * Scans the entire window for the best contiguous free-fall segment.
 * Returns the fraction of required duration found [0.0 - 1.0].
 */
static float detect_free_fall(const sliding_window_t *sw)
{
    uint16_t n = sw_count(sw);
    if (n == 0) return 0.0f;

    uint16_t min_samples = IMPULSE_DURATION_MS / 10;  /* ~20 samples at 100Hz */
    if (min_samples > n) {
        /* Not enough samples to evaluate */
        return 0.0f;
    }

    /* Scan entire window for contiguous free-fall runs */
    uint16_t max_run = 0;
    uint16_t current_run = 0;

    for (uint16_t i = 0; i < n; i++) {
        float mag = calc_accel_magnitude(sw->ax[i], sw->ay[i], sw->az[i]);
        if (mag < FREEFALL_THRESHOLD) {
            current_run++;
            if (current_run > max_run) {
                max_run = current_run;
            }
        } else {
            current_run = 0;
        }
    }

    /* Return fraction of required duration that was found */
    return (float)max_run / (float)min_samples;
}

/*
 * Detect post-impact static phase: very low movement for a sustained period.
 * Assumes the most recent samples represent the post-impact state.
 * Returns the fraction of recent samples that are "static" [0.0 - 1.0].
 */
static float detect_post_impact_static(const sliding_window_t *sw)
{
    uint16_t n = sw_count(sw);
    if (n == 0) return 0.0f;

    uint16_t min_samples = STATIC_DURATION_MS / 10;  /* ~50 samples at 100Hz */
    if (min_samples > n) min_samples = n;

    /* Check the most recent samples (end of window) */
    uint16_t start = (n > min_samples) ? (n - min_samples) : 0;
    uint16_t static_count = 0;

    for (uint16_t i = start; i < n; i++) {
        float mag = calc_accel_magnitude(sw->ax[i], sw->ay[i], sw->az[i]);
        if (mag < STATIC_THRESHOLD) {
            static_count++;
        }
    }

    return (float)static_count / (float)min_samples;
}

/*
 * Detect rapid angular motion (high gyro activity).
 * Used to distinguish falls from other high-motion activities.
 * Returns the maximum gyro magnitude in the window.
 */
static float detect_high_gyro(const sliding_window_t *sw)
{
    uint16_t n = sw_count(sw);
    if (n == 0) return 0.0f;

    float max_mag = 0.0f;
    for (uint16_t i = 0; i < n; i++) {
        float m = calc_gyro_magnitude(sw->gx[i], sw->gy[i], sw->gz[i]);
        if (m > max_mag) max_mag = m;
    }
    return max_mag;
}

/*
 * Check if current motion is "normal" (walking, running, etc.).
 * Normal motion has moderate acceleration magnitude (0.5g - 2.5g)
 * and non-trivial gyro activity.
 */
static bool is_normal_motion(const sliding_window_t *sw)
{
    uint16_t n = sw_count(sw);
    if (n == 0) return false;

    /* Check recent samples for normal motion patterns */
    uint16_t check_count = (n < 10) ? n : 10;
    uint16_t normal_count = 0;

    for (uint16_t i = n - check_count; i < n; i++) {
        float acc_mag = calc_accel_magnitude(sw->ax[i], sw->ay[i], sw->az[i]);
        float gyro_mag = calc_gyro_magnitude(sw->gx[i], sw->gy[i], sw->gz[i]);

        /* Normal motion: moderate accel + some gyro activity */
        if (acc_mag > 0.5f && acc_mag < 2.5f && gyro_mag > 5.0f) {
            normal_count++;
        }
    }

    /* Require majority of recent samples to show normal motion */
    return normal_count >= (check_count * 3 / 4);
}

void fall_detect_init(void)
{
    s_initialized = true;
    s_consecutive_counter = 0;
    s_last_result = NO_FALL;
    s_suspect_start_ms = 0;
    s_in_recovery_window = false;
}

fall_event_t fall_detect_process(sliding_window_t *window)
{
    fall_event_t event;
    memset(&event, 0, sizeof(event));

    if (!s_initialized) {
        fall_detect_init();
    }

    if (window == NULL) {
        event.result = NO_FALL;
        event.confidence = 0.0f;
        event.consecutive_detections = s_consecutive_counter;
        return event;
    }

    uint16_t n = sw_count(window);
    if (n < 3) {
        /* Not enough data for meaningful analysis */
        event.result = NO_FALL;
        event.confidence = 0.0f;
        event.consecutive_detections = s_consecutive_counter;
        return event;
    }

    /* ---- Feature Extraction ---- */

    /* Feature 1: Impact detection */
    float peak_acc_mag = 0.0f;
    bool has_impact = detect_impact(window, &peak_acc_mag);

    /* Feature 2: Free-fall detection */
    float freefall_fraction = detect_free_fall(window);

    /* Feature 3: Post-impact static detection */
    float static_fraction = detect_post_impact_static(window);

    /* Feature 4: High gyro activity (distinguishes falls from other events) */
    float max_gyro = detect_high_gyro(window);

    /* ---- Confidence Calculation ---- */

    /* Weighted scoring:
     *   impact:      weight 0.35 — did we see a hard impact?
     *   free-fall:   weight 0.25 — was there a free-fall phase?
     *   static:      weight 0.40 — is the subject still after impact?
     *
     * Free-fall and static are only scored if impact was detected first,
     * preventing false positives from disconnected sensors (all-zero data).
     */
    float impact_score = has_impact ? 1.0f : 0.0f;
    float freefall_score = has_impact ? freefall_fraction : 0.0f;
    float static_score = has_impact ? static_fraction : 0.0f;

    float confidence = (impact_score * 0.35f) +
                       (freefall_score * 0.25f) +
                       (static_score * 0.40f);

    /* Clamp confidence to [0.0, 1.0] */
    if (confidence > 1.0f) confidence = 1.0f;
    if (confidence < 0.0f) confidence = 0.0f;

    /* ---- Classification ---- */

    if (confidence > 0.85f) {
        event.result = FALL_DETECTED;
    } else if (confidence > 0.7f) {
        event.result = FALL_SUSPECT;
    } else {
        event.result = NO_FALL;
    }

    /* ---- Anti-False-Trigger Mechanism ---- */

    if (event.result == FALL_SUSPECT && !s_in_recovery_window) {
        /* Enter recovery window */
        s_in_recovery_window = true;
        s_suspect_start_ms = 0;  /* Would be xTaskGetTickCount() on real HW */
    }

    if (s_in_recovery_window) {
        /* If normal motion detected during recovery window, cancel suspect */
        if (is_normal_motion(window)) {
            event.result = NO_FALL;
            s_in_recovery_window = false;
            s_consecutive_counter = 0;
        }
    }

    if (event.result == FALL_DETECTED) {
        s_consecutive_counter++;
        event.consecutive_detections = s_consecutive_counter;

        /* Reset recovery window on confirmed detection */
        s_in_recovery_window = false;
    } else {
        /* Any NO_FALL resets the consecutive counter */
        if (event.result == NO_FALL) {
            s_consecutive_counter = 0;
            s_in_recovery_window = false;
        }
        event.consecutive_detections = s_consecutive_counter;
    }

    s_last_result = event.result;
    event.confidence = confidence;

    return event;
}

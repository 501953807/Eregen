/*
 * Eregen (颐贞) - Fall Detection Algorithm Header
 * Rule-based fall detection using acceleration + gyroscope features
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef FALL_DETECT_H
#define FALL_DETECT_H

#include <stdint.h>
#include <stdbool.h>
#include "algorithms/sliding_window.h"

/* Result levels */
typedef enum {
    NO_FALL = 0,        /* Normal motion, no fall indicators */
    FALL_SUSPECT,       /* Fall indicators present, needs confirmation */
    FALL_DETECTED       /* Confirmed fall, ready to alarm */
} fall_result_t;

/* Output event with confidence and anti-misfire state */
typedef struct {
    fall_result_t result;         /* Current classification */
    float confidence;             /* 0.0 - 1.0 weighted score */
    uint8_t consecutive_detections;  /* Count of consecutive FALL_DETECTED */
} fall_event_t;

/* Configurable thresholds */
#define IMPACT_THRESHOLD     1.5f   /* g: |acc| exceeds this on impact */
#define FREEFALL_THRESHOLD   0.3f   /* g: |acc| drops below this in free-fall */
#define STATIC_THRESHOLD     0.2f   /* g: post-impact movement below this */
#define IMPULSE_DURATION_MS  200    /* ms: minimum free-fall duration */
#define STATIC_DURATION_MS   500    /* ms: minimum static period after impact */
#define RECOVERY_WINDOW_MS   5000   /* ms: window to cancel false positive */
#define CONSECUTIVE_REQ      3      /* # of consecutive detections for alarm */

/* Initialize the fall detection engine. Call once at startup. */
void fall_detect_init(void);

/*
 * Process a sliding window of IMU samples and classify fall state.
 * @param window Pointer to populated sliding window (must have >= 3 samples).
 * @return fall_event_t with classification, confidence, and counter state.
 */
fall_event_t fall_detect_process(sliding_window_t *window);

#endif /* FALL_DETECT_H */

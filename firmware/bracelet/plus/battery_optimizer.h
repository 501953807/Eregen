/*
 * Eregen (颐贞) - Battery Optimizer Header (Plus Tier)
 * Adaptive sampling-rate controller that adjusts GPS and PPG
 * measurement intervals based on remaining battery percentage.
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

#ifndef BATTERY_OPTIMIZER_H
#define BATTERY_OPTIMIZER_H

#include <stdint.h>
#include <stdbool.h>

/* ---- Sampling rate tiers ---- */

/** GPS sampling interval when battery > 50% (seconds). */
#define OPTIMIZER_GPS_HIGH_BATT_S      10U

/** GPS sampling interval when battery 20-50% (seconds). */
#define OPTIMIZER_GPS_MED_BATT_S       30U

/** GPS sampling interval when battery < 20% (seconds). */
#define OPTIMIZER_GPS_LOW_BATT_S       60U

/** PPG (heart-rate / SpO2) sampling interval when battery > 50% (seconds). */
#define OPTIMIZER_PPG_HIGH_BATT_S      1U

/** PPG sampling interval when battery <= 50% (seconds). */
#define OPTIMIZER_PPG_LOW_BATT_S       5U

/** Battery level thresholds for tier transitions. */
#define OPTIMIZER_BATT_HIGH_PCT        50U
#define OPTIMIZER_BATT_MED_PCT         20U

/** How often to re-evaluate battery level (seconds). */
#define OPTIMIZER_EVAL_INTERVAL_S      15U

/** Minimum GPS interval to prevent excessive polling. */
#define OPTIMIZER_GPS_MIN_INTERVAL_S   5U

/** Maximum PPG interval to keep health monitoring useful. */
#define OPTIMIZER_PPG_MAX_INTERVAL_S   10U

/* ---- Configuration ---- */

/**
 * Adaptive sampling rates for different subsystems.
 */
typedef struct {
    uint16_t gps_interval_s;   /* GPS fix interval in seconds */
    uint16_t ppg_interval_s;   /* PPG health reading interval in seconds */
    uint8_t  batt_pct;         /* Current battery percentage driving these rates */
    uint8_t  tier;             /* 0=high, 1=medium, 2=low power */
} optimizer_config_t;

/**
 * Battery optimization state tracked across ticks.
 */
typedef struct {
    optimizer_config_t config;       /* Current adaptive rates */
    optimizer_config_t prev_config;  /* Previous rates (for change detection) */
    uint32_t eval_counter_s;         /* Seconds since last battery re-evaluation */
    bool     config_changed;         /* True if rates were adjusted this tick */
} battery_optimizer_t;

/*
 * Initialize the battery optimizer.
 * Reads initial battery level and sets default (high-power) sampling rates.
 * @return true on success.
 */
bool battery_optimizer_init(void);

/*
 * Periodic tick function — call every second from the main loop.
 * Re-evaluates battery level periodically and adjusts sampling rates.
 * @return true if sampling rates were changed during this tick.
 */
bool battery_optimizer_tick(void);

/*
 * Get the current optimized sampling configuration.
 * Callers should use these intervals instead of hardcoded values.
 * @param[out] out Output buffer for current config (must not be NULL).
 * @return true on success.
 */
bool battery_optimizer_get_config(optimizer_config_t *out);

/*
 * Force a battery-level re-evaluation immediately.
 * Normally done automatically every OPTIMIZER_EVAL_INTERVAL_S seconds.
 */
void battery_optimizer_force_eval(void);

/*
 * Notify the optimizer that an external event has occurred
 * (e.g. SOS pressed, fall detected) that requires immediate high-power mode.
 * Temporarily overrides adaptive rates until the event clears.
 * @param active true to enter high-power mode, false to restore adaptive.
 */
void battery_optimizer_set_event_mode(bool active);

/*
 * Get the current optimization tier.
 * @return 0=high, 1=medium, 2=low power.
 */
uint8_t battery_optimizer_get_tier(void);

/* ---- Test helpers ---- */

#ifdef TEST_MODE
/* Override the battery voltage reading for testing. */
void battery_optimizer_set_mock_percent(uint8_t pct);
#endif

#endif /* BATTERY_OPTIMIZER_H */

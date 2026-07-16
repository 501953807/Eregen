/*
 * Eregen (颐贞) - Battery Optimizer Implementation (Plus Tier)
 * Adaptive sampling-rate controller for GPS and PPG subsystems.
 * Adjusts intervals based on battery percentage to extend device runtime.
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

#include "battery_optimizer.h"
#include <string.h>

/* ---- External sensor/battery interfaces (provided by entry tier) ---- */

/* Forward declarations for battery reading — links to battery_adc.h / power_mgmt.h */
extern uint8_t power_check_battery_level(void);
extern bool battery_optimizer_link_tick_callback(void (*cb)(void));

/* ---- Internal state ---- */

static battery_optimizer_t s_optimizer;
static bool s_initialized = false;
static bool s_event_mode = false;
static optimizer_config_t s_prev_config_during_event;

/* Mock battery percentage for testing. */
#ifdef TEST_MODE
static uint8_t s_mock_batt_pct = 80;
#endif

/* ---- Sampling rate calculation ---- */

/**
 * Determine optimization tier and corresponding intervals from battery level.
 */
static void calculate_rates(uint8_t batt_pct, optimizer_config_t *cfg)
{
    cfg->batt_pct = batt_pct;

    if (batt_pct > OPTIMIZER_BATT_HIGH_PCT) {
        /* High battery: aggressive sampling. */
        cfg->tier = 0;
        cfg->gps_interval_s = OPTIMIZER_GPS_HIGH_BATT_S;
        cfg->ppg_interval_s = OPTIMIZER_PPG_HIGH_BATT_S;
    } else if (batt_pct > OPTIMIZER_BATT_MED_PCT) {
        /* Medium battery: moderate sampling. */
        cfg->tier = 1;
        cfg->gps_interval_s = OPTIMIZER_GPS_MED_BATT_S;
        cfg->ppg_interval_s = OPTIMIZER_PPG_LOW_BATT_S;
    } else {
        /* Low battery: conservative sampling. */
        cfg->tier = 2;
        cfg->gps_interval_s = OPTIMIZER_GPS_LOW_BATT_S;
        cfg->ppg_interval_s = OPTIMIZER_PPG_LOW_BATT_S;
    }

    /* Enforce hard limits. */
    if (cfg->gps_interval_s < OPTIMIZER_GPS_MIN_INTERVAL_S) {
        cfg->gps_interval_s = OPTIMIZER_GPS_MIN_INTERVAL_S;
    }
    if (cfg->ppg_interval_s > OPTIMIZER_PPG_MAX_INTERVAL_S) {
        cfg->ppg_interval_s = OPTIMIZER_PPG_MAX_INTERVAL_S;
    }
}

/**
 * Check whether two configs differ.
 */
static bool config_changed(const optimizer_config_t *a, const optimizer_config_t *b)
{
    return (a->gps_interval_s != b->gps_interval_s ||
            a->ppg_interval_s != b->ppg_interval_s ||
            a->tier != b->tier);
}

/* ---- Public API ---- */

bool battery_optimizer_init(void)
{
    optimizer_config_t default_cfg;
    default_cfg.gps_interval_s = OPTIMIZER_GPS_HIGH_BATT_S;
    default_cfg.ppg_interval_s = OPTIMIZER_PPG_HIGH_BATT_S;
    default_cfg.batt_pct = 100;
    default_cfg.tier = 0;

    s_optimizer.config = default_cfg;
    s_optimizer.prev_config = default_cfg;
    s_optimizer.eval_counter_s = 0;
    s_optimizer.config_changed = false;
    s_initialized = true;
    s_event_mode = false;

    return true;
}

bool battery_optimizer_tick(void)
{
    if (!s_initialized) {
        return false;
    }

    s_optimizer.eval_counter_s++;

    /* During event mode, restore high-power rates immediately. */
    if (s_event_mode) {
        /* Restore previous config if event just cleared. */
        if (s_optimizer.config_changed) {
            s_optimizer.config_changed = false;
            return true;
        }
        return false;
    }

    /* Re-evaluate battery level at configured interval. */
    if (s_optimizer.eval_counter_s >= OPTIMIZER_EVAL_INTERVAL_S) {
        s_optimizer.eval_counter_s = 0;

#ifdef TEST_MODE
        uint8_t batt_pct = s_mock_batt_pct;
#else
        uint8_t batt_pct = power_check_battery_level();
#endif

        optimizer_config_t new_cfg;
        calculate_rates(batt_pct, &new_cfg);

        if (config_changed(&new_cfg, &s_optimizer.config)) {
            s_optimizer.prev_config = s_optimizer.config;
            s_optimizer.config = new_cfg;
            s_optimizer.config_changed = true;
            return true;
        }
    }

    return false;
}

bool battery_optimizer_get_config(optimizer_config_t *out)
{
    if (!out) {
        return false;
    }
    *out = s_optimizer.config;
    return true;
}

void battery_optimizer_force_eval(void)
{
    s_optimizer.eval_counter_s = OPTIMIZER_EVAL_INTERVAL_S;
}

void battery_optimizer_set_event_mode(bool active)
{
    if (active && !s_event_mode) {
        /* Save current config and switch to high power. */
        s_prev_config_during_event = s_optimizer.config;
        s_optimizer.config.gps_interval_s = OPTIMIZER_GPS_HIGH_BATT_S;
        s_optimizer.config.ppg_interval_s = OPTIMIZER_PPG_HIGH_BATT_S;
        s_optimizer.config.tier = 0;
        s_event_mode = true;
    } else if (!active && s_event_mode) {
        /* Restore previous config. */
        s_optimizer.config = s_prev_config_during_event;
        s_event_mode = false;
    }
}

uint8_t battery_optimizer_get_tier(void)
{
    return s_optimizer.config.tier;
}

/* ---- Test helpers ---- */

#ifdef TEST_MODE

void battery_optimizer_set_mock_percent(uint8_t pct)
{
    s_mock_batt_pct = pct;
}

#endif

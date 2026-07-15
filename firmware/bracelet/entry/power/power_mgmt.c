/*
 * Eregen (颐贞) - Power Management Module Implementation
 * Deep sleep, per-peripheral power control, battery monitoring, mode transitions.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "power_mgmt.h"
#include "../battery_adc.h"
#include "../common/log.h"
#include <string.h>

/* Internal state */
static power_mode_t s_current_mode = POWER_NORMAL;
static uint32_t s_periph_mask = PERIPH_ALL;     /* Currently powered peripherals */
static bool s_low_battery_alert = false;
static bool s_critical_battery_alert = false;
static uint8_t s_last_percent = 255;            /* Track last known percentage */
static uint32_t s_manage_tick_counter = 0;       /* Ticks since last manage() call */

#ifdef TEST_MODE
/* Direct mock voltage override — stored here to avoid cross-module state issues */
static uint16_t s_mock_voltage_mv = 0;
static bool s_mock_voltage_active = false;
#endif

/*
 * Initialize the power management subsystem.
 */
void power_init(void)
{
    memset(&s_current_mode, 0, sizeof(power_mode_t));
    s_periph_mask = PERIPH_ALL;
    s_low_battery_alert = false;
    s_critical_battery_alert = false;
    s_last_percent = 255;
    s_manage_tick_counter = 0;

    log_info("Power management initialized");
}

/*
 * Enter deep sleep mode. Only RTC alarm or external reset can wake up.
 * On embedded target this calls system_enter_stop_mode().
 * In test mode, this is a no-op that logs the intent.
 */
void power_enter_deep_sleep(void)
{
    log_warn("Entering deep sleep — all peripherals off, only RTC wakeup enabled");
    s_current_mode = POWER_DEEP_SLEEP;

    /* Power off all peripherals */
    s_periph_mask = 0U;
    power_periph_control(PERIPH_ALL, false);

#ifdef __EMBEDDED__
    /* Call GD32 system-level stop mode: only RTC can wake */
    /* system_enter_stop_mode(); */
#else
    /* In test mode, just log */
    log_info("Deep sleep entered (test mode — no-op)");
#endif
}

/*
 * Set the current power mode.
 */
void power_set_mode(power_mode_t mode)
{
    if (mode == s_current_mode) {
        return;
    }

    log_info("Power mode change: %u -> %u", (unsigned)s_current_mode, (unsigned)mode);

    switch (mode) {
    case POWER_NORMAL:
        /* Restore all peripherals */
        power_periph_control(PERIPH_ALL, true);
        break;

    case POWER_LIGHT_SLEEP:
        /* Turn off display and PPG (non-critical) */
        power_periph_control(PERIPH_DISPLAY | PERIPH_PPG, false);
        break;

    case POWER_ALERT_MODE:
        /* Keep IMU and GPS on for fall detection / geofence */
        /* Already running in normal — just note the mode */
        break;

    case POWER_DEEP_SLEEP:
        power_enter_deep_sleep();
        return;  /* power_enter_deep_sleep already set the mode */
    }

    s_current_mode = mode;
}

/*
 * Get the current power mode.
 */
power_mode_t power_get_mode(void)
{
    return s_current_mode;
}

/*
 * Check battery level and return percentage.
 * Returns 255 on error (ADC failure).
 */
uint8_t power_check_battery_level(void)
{
    uint16_t voltage_mv;
    uint8_t percent;

#ifdef TEST_MODE
    /* In test mode, use direct mock voltage */
    if (s_mock_voltage_active) {
        voltage_mv = s_mock_voltage_mv;
    } else {
        voltage_mv = battery_read_voltage_mv();
    }
#else
    voltage_mv = battery_read_voltage_mv();
#endif

    if (voltage_mv == 0 && !s_mock_voltage_active) {
        log_error("Battery ADC read failed");
        return 255;
    }

    /* Calculate percentage directly from voltage (linear interpolation) */
    uint16_t empty_mv = (uint16_t)(BATT_VOLTAGE_EMPTY * 1000.0f);
    uint16_t full_mv = (uint16_t)(BATT_VOLTAGE_FULL * 1000.0f);

    if (voltage_mv <= empty_mv) {
        percent = 0U;
    } else if (voltage_mv >= full_mv) {
        percent = 100U;
    } else {
        uint16_t range = full_mv - empty_mv;
        uint16_t above_empty = voltage_mv - empty_mv;
        percent = (uint8_t)((above_empty * 100U) / range);
    }

    /* Update alert flags */
    if (percent <= POWER_BATT_LOW_PCT && !s_low_battery_alert) {
        s_low_battery_alert = true;
        log_warn("Low battery alert: %u%%", (unsigned)percent);
    } else if (percent > POWER_BATT_LOW_PCT && s_low_battery_alert) {
        s_low_battery_alert = false;
    }

    if (percent <= POWER_BATT_CRITICAL_PCT && !s_critical_battery_alert) {
        s_critical_battery_alert = true;
        log_error("CRITICAL battery: %u%% — prepare to shut down", (unsigned)percent);
    } else if (percent > POWER_BATT_CRITICAL_PCT && s_critical_battery_alert) {
        s_critical_battery_alert = false;
    }

    /* Report once when percentage changes by at least 1% */
    if (s_last_percent != percent) {
        log_info("BATTERY: %umV, %u%%", (unsigned)voltage_mv, (unsigned)percent);
        s_last_percent = percent;
    }

    return percent;
}

/*
 * Control peripheral power state.
 * @param mask Bitmask of peripherials to affect.
 * @param enable true to power on, false to power off.
 *
 * On embedded target, this toggles GPIO regulators or I2C power rails.
 * In test mode, it updates the internal state mask only.
 */
void power_periph_control(uint32_t mask, bool enable)
{
    if (enable) {
        s_periph_mask |= mask;
    } else {
        s_periph_mask &= ~mask;
    }

    /* Log which peripherals are affected */
    char buf[64] = "Peripherals:";
    uint8_t idx = 10;
    if (mask & PERIPH_GPS)   { buf[idx++] = 'G'; buf[idx++] = ','; }
    if (mask & PERIPH_CAT1)  { buf[idx++] = 'C'; buf[idx++] = ','; }
    if (mask & PERIPH_DISPLAY) { buf[idx++] = 'D'; buf[idx++] = ','; }
    if (mask & PERIPH_PPG)   { buf[idx++] = 'P'; buf[idx++] = ','; }
    if (mask & PERIPH_IMU)   { buf[idx++] = 'I'; buf[idx++] = ','; }
    buf[idx - 1] = '\0';  /* Remove trailing comma */
    buf[idx] = '\0';

    log_debug("%s %s", buf, enable ? "ON" : "OFF");
}

/*
 * Check if a low-battery alert has been triggered.
 */
bool power_is_low_battery(void)
{
    return s_low_battery_alert;
}

/*
 * Check if a critical-battery alert has been triggered.
 */
bool power_is_critical_battery(void)
{
    return s_critical_battery_alert;
}

/*
 * Main power management tick function.
 * Call periodically (every ~1 second recommended).
 * Handles: battery monitoring, mode transitions, idle detection.
 */
void power_manage(void)
{
    s_manage_tick_counter++;

    /* Check battery every 10 ticks (~10 seconds) */
    if (s_manage_tick_counter % 10 == 0) {
        uint8_t pct = power_check_battery_level();

        /* Auto-enter light sleep if battery is low and no alert pending */
        if (pct <= POWER_BATT_LOW_PCT && s_current_mode == POWER_NORMAL) {
            log_warn("Auto-entering light sleep due to low battery (%u%%)", (unsigned)pct);
            power_set_mode(POWER_LIGHT_SLEEP);
        }

        /* Enter deep sleep if critically low */
        if (pct <= POWER_BATT_CRITICAL_PCT && s_current_mode != POWER_DEEP_SLEEP) {
            log_error("Battery critical (%u%%) — entering deep sleep", (unsigned)pct);
            power_set_mode(POWER_DEEP_SLEEP);
        }
    }

    /* Auto-wake from light sleep after 60 ticks (~60 seconds) */
    if (s_current_mode == POWER_LIGHT_SLEEP && s_manage_tick_counter % 60 == 0) {
        uint8_t pct = power_check_battery_level();
        if (pct > POWER_BATT_LOW_PCT) {
            log_info("Battery recovered (%u%%) — waking from light sleep", (unsigned)pct);
            power_set_mode(POWER_NORMAL);
        }
    }
}

/* ============ Test-mode helpers ============ */

#ifdef TEST_MODE
/* Set a direct mock voltage value in millivolts */
void power_set_mock_voltage(uint16_t mv)
{
    s_mock_voltage_mv = mv;
    s_mock_voltage_active = true;
}
#endif

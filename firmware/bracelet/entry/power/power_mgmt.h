/*
 * Eregen (颐贞) - Power Management Module Header
 * Deep sleep, per-peripheral power control, battery monitoring, mode transitions.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef POWER_MGMT_H
#define POWER_MGMT_H

#include <stdint.h>
#include <stdbool.h>

/* Power modes */
typedef enum {
    POWER_NORMAL = 0,     /* Full operation, all peripherals active */
    POWER_DEEP_SLEEP,     /* MCU stops, only RTC wakeup */
    POWER_LIGHT_SLEEP,    /* Reduced clock, some peripherals off */
    POWER_ALERT_MODE      /* High-frequency sampling for SOS/fall detection */
} power_mode_t;

/* Peripheral power control bits */
typedef enum {
    PERIPH_GPS   = (1U << 0),
    PERIPH_CAT1  = (1U << 1),
    PERIPH_DISPLAY = (1U << 2),
    PERIPH_PPG   = (1U << 3),
    PERIPH_IMU   = (1U << 4),
    PERIPH_ALL   = 0xFFFFU
} periph_id_t;

/* Battery alert thresholds */
#define POWER_BATT_CRITICAL_PCT    5U   /* Below this: immediate shutdown warning */
#define POWER_BATT_LOW_PCT         10U  /* Below this: enter low-power mode */
#define POWER_BATT_FULL_PCT        95U  /* Above this: charging indicator */

/* Sleep configuration */
#define POWER_DEEP_SLEEP_TICKS     300UL  /* Minimum deep-sleep duration in ticks */
#define POWER_WAKEUP_INTERVAL_MS   5000UL /* Normal wakeup polling interval */

/**
 * Initialize the power management subsystem.
 * Must be called once at system startup before any power mode changes.
 */
void power_init(void);

/**
 * Enter deep sleep mode. Only RTC alarm or external reset can wake up.
 * All peripheral clocks are disabled. Calls system-level stop/sleep.
 */
void power_enter_deep_sleep(void);

/**
 * Set the current power mode.
 * @param mode Target power mode.
 */
void power_set_mode(power_mode_t mode);

/**
 * Get the current power mode.
 * @return Current power mode.
 */
power_mode_t power_get_mode(void);

/**
 * Check battery level and return percentage.
 * @return Battery percentage 0-100, or 255 on error.
 */
uint8_t power_check_battery_level(void);

/**
 * Main power management tick function.
 * Call periodically (every ~1 second recommended).
 * Handles: battery monitoring, mode transitions, peripheral power control.
 */
void power_manage(void);

/**
 * Control peripherial power state.
 * @param mask Bitmask of periph_id_t to disable (or enable if enable=true).
 * @param enable true to power on, false to power off.
 */
void power_periph_control(uint32_t mask, bool enable);

/**
 * Check if a low-battery alert has been triggered.
 * @return true if battery is below LOW_PCT threshold.
 */
bool power_is_low_battery(void);

/**
 * Check if a critical-battery alert has been triggered.
 * @return true if battery is below CRITICAL_PCT threshold.
 */
bool power_is_critical_battery(void);

/* Test-mode helpers */
#ifdef TEST_MODE
void power_set_mock_voltage(uint16_t mv);
#endif

#endif /* POWER_MGMT_H */

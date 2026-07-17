/*
 * Eregen (颐贞) - Power Optimizer Header (Pro Tier)
 * Adaptive power management that switches between active, light-sleep,
 * and deep-sleep modes based on activity state and battery level.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef POWER_OPTIMIZER_H
#define POWER_OPTIMIZER_H

#include <stdint.h>
#include <stdbool.h>

/** Power operating modes. */
typedef enum {
    POWER_MODE_ACTIVE,      /* Full operation, all peripherals enabled */
    POWER_MODE_LIGHT_SLEEP, /* CPU halted, RTC/WKUP active, ~50uA */
    POWER_MODE_DEEP_SLEEP,  /* Only RTC runs, ~10uA */
} PowerMode_t;

/**
 * Initialize power management subsystem.
 * Reads initial battery level and sets default mode.
 */
void power_optimizer_init(void);

/** Enter specified power mode. */
void power_set_mode(PowerMode_t mode);

/** Get current power mode. */
PowerMode_t power_get_mode(void);

/** Read battery percentage (0-100). */
int power_get_battery_pct(void);

/** Enter deep sleep for specified interval (milliseconds). */
void power_enter_deep_sleep(uint32_t wake_interval_ms);

/**
 * Check if battery is critically low (< 10%).
 * @return true if battery is critically low.
 */
bool power_is_critical(void);

/**
 * Enter light sleep until GPIO wakeup event.
 * @param gpio_num GPIO pin to monitor for wakeup.
 * @param high true for high-level wakeup, false for low-level.
 * @return true if woke up from external event.
 */
bool power_light_sleep_on_gpio(int gpio_num, bool high_active);

#endif /* POWER_OPTIMIZER_H */

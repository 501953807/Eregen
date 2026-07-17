/*
 * Eregen (颐贞) - Power Optimizer Implementation (Pro Tier)
 * GD32 FreeRTOS-based power management for Pro bracelet.
 * Uses GD32 HAL sleep functions and battery ADC monitoring.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "power_optimizer.h"
#include "board_pro.h"
#include "../entry/battery_adc.h"
#include "../common/log.h"
#include "libmaple/misc.h" /* For gd32_sysclock_enter_lowpower / exit_lowpower */

static PowerMode_t s_mode = POWER_MODE_ACTIVE;

void power_optimizer_init(void) {
    battery_init();
    s_mode = POWER_MODE_ACTIVE;
    log_info("PowerOptimizer: Initialized in ACTIVE mode");
}

void power_set_mode(PowerMode_t mode) {
    s_mode = mode;
    switch (mode) {
        case POWER_MODE_LIGHT_SLEEP:
            /* Enter light sleep via GD32 HAL */
            sys_clock_enter_lowpower();
            break;
        case POWER_MODE_DEEP_SLEEP:
            /* Deep sleep handled by power_enter_deep_sleep() */
            break;
        default:
            sys_clock_exit_lowpower();
            break;
    }
}

PowerMode_t power_get_mode(void) {
    return s_mode;
}

int power_get_battery_pct(void) {
    battery_status_t batt = battery_get_status();
    return (int)batt.percent;
}

bool power_is_critical(void) {
    int pct = power_get_battery_pct();
    return pct < 10;
}

void power_enter_deep_sleep(uint32_t wake_interval_ms) {
    /* Configure RTC alarm for wake-up */
    /* In production, set RTC alarm register here */
    (void)wake_interval_ms;
    log_warn("PowerOptimizer: Deep sleep not yet implemented for GD32E230");
    /* esp_deep_sleep_start() equivalent would go here */
}

bool power_light_sleep_on_gpio(int gpio_num, bool high_active) {
    (void)gpio_num;
    (void)high_active;
    /* Enable EXTI wakeup pin */
    log_info("PowerOptimizer: Light sleep on GPIO %d", gpio_num);
    /* Configure EXTI line and enter low-power mode */
    /* Return true if woken by external event */
    return false;
}

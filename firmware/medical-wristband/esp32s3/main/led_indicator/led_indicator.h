/*
 * Eregen Medical Wristband - LED Indicator Header
 * RGB LED for patient condition visualization
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef LED_INDICATOR_H
#define LED_INDICATOR_H

#include "esp_err.h"
#include <stdint.h>

typedef enum {
    LED_OFF = 0,
    LED_GREEN,
    LED_RED,
    LED_YELLOW,
    LED_BLUE,
    LED_WHITE,
} led_color_t;

esp_err_t led_indicator_init(void);
esp_err_t led_indicator_set_color(led_color_t color);
esp_err_t led_indicator_blink(led_color_t color, int times, int interval_ms);
void led_indicator_set_alert(const char *alert_type);

#endif /* LED_INDICATOR_H */

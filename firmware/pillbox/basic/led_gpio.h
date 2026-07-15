/*
 * Eregen (颐贞) - LED Status Indicator Module
 * GPIO/LEDC control for RGB status LED on ESP32-C3
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef LED_GPIO_H
#define LED_GPIO_H

#include "esp_err.h"

/* LED color states */
typedef enum {
    LED_COLOR_GREEN,   /* Normal operation */
    LED_COLOR_RED,     /* Alert / error */
    LED_COLOR_BLUE,    /* Pairing mode */
    LED_COLOR_OFF      /* LED off */
} led_color_t;

/* LED blink patterns */
typedef enum {
    LED_PATTERN_SOLID,       /* Continuously on */
    LED_PATTERN_SLOW_BLINK,  /* 1 Hz: 500ms on / 500ms off */
    LED_PATTERN_FAST_BLINK,  /* 4 Hz: 125ms on / 125ms off */
    LED_PATTERN_PULSE        /* Gradual brightness ramp */
} led_pattern_t;

/**
 * Initialize the LED peripheral (LEDC on ESP32-C3).
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t led_init(void);

/**
 * Set LED to a solid color.
 *
 * @param color The color to set
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t led_set_color(led_color_t color);

/**
 * Set LED to a blinking pattern.
 *
 * @param color The base color for blinking
 * @param pattern The blink pattern
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t led_blink(led_color_t color, led_pattern_t pattern);

/**
 * Turn the LED off.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t led_off(void);

#endif /* LED_GPIO_H */

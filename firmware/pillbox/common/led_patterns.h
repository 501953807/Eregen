/*
 * Eregen (颐贞) - LED Blink Pattern Library
 * Non-blocking LED pattern engine using ESP32-C3 LEDC.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef LED_PATTERNS_H
#define LED_PATTERNS_H

#include "esp_err.h"
#include <stdint.h>

/* LED pattern types */
typedef enum {
    PATTERN_GREEN_SOLID,    /* Solid green = normal operation */
    PATTERN_RED_BLINK_FAST, /* Fast red blink = alert */
    PATTERN_BLUE_BLINK_SLOW,/* Slow blue blink = pairing mode */
    PATTERN_AMBER_PULSE,    /* Amber pulse = medication due soon */
    PATTERN_OFF,            /* All LEDs off */
    PATTERN_COUNT           /* Number of patterns */
} led_pattern_t;

/**
 * Initialize the LED pattern subsystem.
 * Sets up LEDC channels for RGB LED on ESP32-C3.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t led_pattern_init(void);

/**
 * Start a specific LED pattern.
 * Pattern runs non-blocking in background.
 *
 * @param pattern The pattern to start
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t led_pattern_start(led_pattern_t pattern);

/**
 * Stop the current LED pattern and turn off LEDs.
 *
 * @return ESP_OK on success
 */
esp_err_t led_pattern_stop(void);

/**
 * Set the active LED pattern (alias for start).
 *
 * @param pattern The pattern to set
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t led_pattern_set(led_pattern_t pattern);

#endif /* LED_PATTERNS_H */

/*
 * Eregen (颐贞) - Button Input Module
 * Physical button polling with debounce and press detection
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef BUTTON_INPUT_H
#define BUTTON_INPUT_H

#include "esp_err.h"

/* Button events */
typedef enum {
    BUTTON_NONE,
    BUTTON_SHORT_PRESS,
    BUTTON_LONG_PRESS,
    BUTTON_DOUBLE_PRESS
} button_event_t;

/* Button identifiers */
typedef enum {
    BUTTON_ENTER = 0,
    BUTTON_RIGHT,
    BUTTON_COUNT
} button_id_t;

/* Timing constants (milliseconds) */
#define BUTTON_DEBOUNCE_MS      50
#define BUTTON_LONG_PRESS_MS    2000
#define BUTTON_DOUBLE_GAP_MS    300

/**
 * Initialize button GPIO pins with pull-up resistors.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t buttons_init(void);

/**
 * Poll all buttons and return the latest button event.
 * Call this periodically (e.g., every 10-20ms) from the main loop.
 *
 * @return The latest button event (BUTTON_NONE if no event)
 */
button_event_t buttons_get_event(void);

/**
 * Clear the last consumed button event.
 * Call after handling an event to acknowledge it.
 */
void buttons_clear_event(void);

#endif /* BUTTON_INPUT_H */

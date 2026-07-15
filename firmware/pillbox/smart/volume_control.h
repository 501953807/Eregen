/*
 * Eregen (颐贞) - Volume Control Module
 * Smart pillbox tier — TTS volume adjustment via buttons + NVS persistence
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef VOLUME_CONTROL_H
#define VOLUME_CONTROL_H

#include "esp_err.h"
#include <stdint.h>
#include <stdbool.h>

/* Default volume percentage */
#define VOLUME_DEFAULT          80

/**
 * Initialize volume control. Loads saved volume from NVS.
 *
 * @return ESP_OK on success
 */
esp_err_t volume_init(void);

/**
 * Get current volume setting.
 *
 * @return Volume percentage (0-100)
 */
uint8_t volume_get(void);

/**
 * Set volume to a specific percentage.
 * Persists to NVS so it survives reboots.
 *
 * @param percent Volume percentage (0-100)
 * @return ESP_OK on success
 */
esp_err_t volume_set(uint8_t percent);

/**
 * Increase volume by 10% steps (clamped to 100).
 *
 * @return New volume percentage
 */
uint8_t volume_increase(void);

/**
 * Decrease volume by 10% steps (clamped to 0).
 *
 * @return New volume percentage
 */
uint8_t volume_decrease(void);

/**
 * Handle a volume-adjust button event.
 * Called from button task when long-press right button is detected.
 *
 * @param increase True for volume up, false for volume down
 * @return New volume percentage after adjustment
 */
uint8_t volume_handle_button(bool increase);

#endif /* VOLUME_CONTROL_H */

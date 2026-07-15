/*
 * Eregen (颐贞) - Device ID Management
 * Generates and persists PX-XXXXXXXX device identifier from ESP32-C3 MAC.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef DEVICE_ID_H
#define DEVICE_ID_H

#include "esp_err.h"
#include <stddef.h>
#include <stdint.h>

#define DEVICE_ID_PREFIX      "PX-"
#define DEVICE_ID_LEN         8       /* hex digits from MAC */
#define DEVICE_ID_FULL_LEN    11      /* "PX-" + 8 hex chars + NUL */

/**
 * Initialize the device ID subsystem.
 * Loads existing ID from NVS; if none found, generates a new one.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t device_id_init(void);

/**
 * Retrieve the current device ID string (with prefix).
 *
 * @param buf   Output buffer (must be at least DEVICE_ID_FULL_LEN bytes)
 * @param len   Size of output buffer
 * @return ESP_OK on success, ESP_ERR_INVALID_SIZE if buffer too small
 */
esp_err_t device_id_get(char *buf, size_t len);

/**
 * Generate a new device ID from the ESP32-C3 WiFi MAC address.
 * This function does NOT save to NVS — call device_id_save_to_nvs() separately.
 */
void device_id_generate(void);

/**
 * Save the current device ID to NVS flash.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t device_id_save_to_nvs(void);

/**
 * Load the device ID from NVS flash into the internal buffer.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t device_id_load_from_nvs(void);

#endif /* DEVICE_ID_H */

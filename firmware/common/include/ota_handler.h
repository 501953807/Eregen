/*
 * Eregen (颐贞) - OTA Firmware Update Handler
 * ESP32-C3 based pillbox OTA receiver: parses MQTT ota commands,
 * downloads firmware via HTTPS, verifies SHA-256, writes to OTA
 * partition, reports progress, reboots into new firmware.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef OTA_HANDLER_H
#define OTA_HANDLER_H

#include "esp_err.h"
#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/**
 * Initialize the OTA update subsystem.
 * Must be called before ota_handle_command().
 *
 * @return ESP_OK on success
 */
esp_err_t ota_init(void);

/**
 * Handle an incoming OTA command from MQTT.
 * Parses the JSON payload and starts the OTA update process
 * as a background task.
 *
 * @param topic     MQTT topic (unused, for dispatch consistency)
 * @param payload   JSON payload: {"type":"ota","url":"...","hash":"sha256:...","ver":"...",force:false}
 * @param len       Payload length
 * @return ESP_OK if command accepted (update started in background),
 *         ESP_ERR_INVALID_ARG if JSON is malformed
 */
esp_err_t ota_handle_command(const char* topic, const uint8_t* payload, uint16_t len);

/**
 * Report OTA progress to cloud via MQTT.
 * Format: {"type":"ota_progress","dev_id":"PX-AUTO-XXXX","job_id":"...","progress":NN,"status":"downloading|verifying|flashing|complete|failed","error":"..."}
 *
 * @param progress  0-100 percentage
 * @param status    Status string: "downloading", "verifying", "flashing", "complete", "failed"
 * @param error     Error description (empty on success)
 * @return ESP_OK on success
 */
esp_err_t ota_report_progress(int progress, const char* status, const char* error);

/**
 * Check if an OTA update is in progress.
 *
 * @return true if OTA task is running
 */
bool ota_is_active(void);

/**
 * Get the current OTA status string.
 *
 * @return Static string: "idle", "downloading", "verifying", "flashing", "rebooting"
 */
const char* ota_get_status(void);

#ifdef __cplusplus
}
#endif

#endif /* OTA_HANDLER_H */

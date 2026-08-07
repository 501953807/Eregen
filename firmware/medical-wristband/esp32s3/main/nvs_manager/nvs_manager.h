/*
 * Eregen Medical Wristband - NVS Manager Header
 * Non-volatile storage management
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef NVS_MANAGER_H
#define NVS_MANAGER_H

#include "esp_err.h"
#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

esp_err_t nvs_manager_init(void);
esp_err_t nvs_manager_save_device_id(const char *device_id);
esp_err_t nvs_manager_load_device_id(char *device_id, size_t *len);
esp_err_t nvs_manager_save_firmware_version(const char *fw_version);
esp_err_t nvs_manager_load_firmware_version(char *fw_version, size_t *len);
bool nvs_manager_validate_integrity(void);

#endif /* NVS_MANAGER_H */

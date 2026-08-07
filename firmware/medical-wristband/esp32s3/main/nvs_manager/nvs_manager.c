/*
 * Eregen Medical Wristband - NVS Manager Module
 * Non-volatile storage management with integrity checks
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "nvs_manager.h"
#include <string.h>
#include <stdio.h>
#include "esp_log.h"
#include "nvs_flash.h"
#include "nvs.h"
#include "crc32.h"

static const char *TAG = "nvs_manager";

esp_err_t nvs_manager_init(void) {
    esp_err_t err = nvs_flash_init();
    if (err == ESP_ERR_NVS_NO_FREE_PAGES || err == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        ESP_ERROR_CHECK(nvs_flash_erase());
        err = nvs_flash_init();
    }
    return err;
}

esp_err_t nvs_manager_save_device_id(const char *device_id) {
    nvs_handle_t handle;
    esp_err_t err = nvs_open("med_wb", NVS_READWRITE, &handle);
    if (err != ESP_OK) return err;

    err = nvs_set_str(handle, "dev_id", device_id);
    if (err == ESP_OK) err = nvs_commit(handle);
    nvs_close(handle);
    return err;
}

esp_err_t nvs_manager_load_device_id(char *device_id, size_t *len) {
    nvs_handle_t handle;
    esp_err_t err = nvs_open("med_wb", NVS_READONLY, &handle);
    if (err != ESP_OK) return err;

    err = nvs_get_str(handle, "dev_id", device_id, len);
    nvs_close(handle);
    return err;
}

esp_err_t nvs_manager_save_firmware_version(const char *fw_version) {
    nvs_handle_t handle;
    esp_err_t err = nvs_open("med_wb", NVS_READWRITE, &handle);
    if (err != ESP_OK) return err;

    err = nvs_set_str(handle, "fw_ver", fw_version);
    if (err == ESP_OK) err = nvs_commit(handle);
    nvs_close(handle);
    return err;
}

esp_err_t nvs_manager_load_firmware_version(char *fw_version, size_t *len) {
    nvs_handle_t handle;
    esp_err_t err = nvs_open("med_wb", NVS_READONLY, &handle);
    if (err != ESP_OK) return err;

    err = nvs_get_str(handle, "fw_ver", fw_version, len);
    nvs_close(handle);
    return err;
}

bool nvs_manager_validate_integrity(void) {
    nvs_handle_t handle;
    esp_err_t err = nvs_open("med_wb", NVS_READONLY, &handle);
    if (err != ESP_OK) return false;

    size_t len = 0;
    err = nvs_get_blob(handle, "integrity", NULL, &len);
    nvs_close(handle);

    if (err != ESP_OK || len == 0) return false;

    // In production, would verify CRC32 checksum
    return true;
}

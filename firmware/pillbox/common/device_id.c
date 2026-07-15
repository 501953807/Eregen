/*
 * Eregen (颐贞) - Device ID Management Implementation
 * Generates and persists PX-XXXXXXXX device identifier from ESP32-C3 MAC.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "device_id.h"

#include "esp_log.h"
#include "esp_wifi.h"
#include "nvs_flash.h"

static const char *TAG = "device_id";

/* Internal storage for device ID (with prefix) */
static char s_device_id[DEVICE_ID_FULL_LEN] = "";

#define NVS_NAMESPACE       "eregen_pillbox"
#define NVS_KEY             "dev_id"

/**
 * Initialize the device ID subsystem.
 * Loads existing ID from NVS; if none found, generates a new one.
 */
esp_err_t device_id_init(void)
{
    esp_err_t ret = device_id_load_from_nvs();
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "No saved device ID in NVS, generating new one");
        device_id_generate();
        device_id_save_to_nvs();
    }

    ESP_LOGI(TAG, "Device ID: %s", s_device_id);
    return ESP_OK;
}

/**
 * Retrieve the current device ID string (with prefix).
 */
esp_err_t device_id_get(char *buf, size_t len)
{
    if (len < DEVICE_ID_FULL_LEN) {
        return ESP_ERR_INVALID_SIZE;
    }
    strncpy(buf, s_device_id, len - 1);
    buf[len - 1] = '\0';
    return ESP_OK;
}

/**
 * Generate a new device ID from the ESP32-C3 WiFi MAC address.
 */
void device_id_generate(void)
{
    uint8_t mac[6];
    esp_read_mac(mac, ESP_MAC_WIFI_STA);

    snprintf(s_device_id, sizeof(s_device_id),
             "%02X%02X%02X%02X%02X%02X",
             mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);
}

/**
 * Save the current device ID to NVS flash.
 */
esp_err_t device_id_save_to_nvs(void)
{
    nvs_handle_t handle;
    esp_err_t ret = nvs_open(NVS_NAMESPACE, NVS_READWRITE, &handle);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS open failed: %s", esp_err_to_name(ret));
        return ret;
    }

    ret = nvs_set_str(handle, NVS_KEY, s_device_id);
    if (ret == ESP_OK) {
        ret = nvs_commit(handle);
    }
    nvs_close(handle);

    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS save failed: %s", esp_err_to_name(ret));
    }
    return ret;
}

/**
 * Load the device ID from NVS flash into the internal buffer.
 */
esp_err_t device_id_load_from_nvs(void)
{
    nvs_handle_t handle;
    esp_err_t ret = nvs_open(NVS_NAMESPACE, NVS_READONLY, &handle);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS open failed: %s", esp_err_to_name(ret));
        return ret;
    }

    size_t len = sizeof(s_device_id);
    ret = nvs_get_str(handle, NVS_KEY, s_device_id, &len);
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "No device ID in NVS: %s", esp_err_to_name(ret));
    } else {
        ESP_LOGI(TAG, "Loaded device ID from NVS: %s", s_device_id);
    }

    nvs_close(handle);
    return ret;
}

/*
 * Eregen Medical Wristband - Patient Store Module
 * Secure NVS-backed patient data storage with encryption
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "patient_store.h"
#include <string.h>
#include <stdio.h>
#include "nvs.h"
#include "esp_log.h"
#include "payload_crypto.h"

static const char *TAG = "patient_store";
static nvs_handle_t s_nvs_handle;

esp_err_t patient_store_init(void) {
    esp_err_t err = nvs_open("med_wb_patient", NVS_READWRITE, &s_nvs_handle);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "Failed to open NVS namespace: %s", esp_err_to_name(err));
        return err;
    }
    ESP_LOGI(TAG, "Patient store initialized");
    return ESP_OK;
}

esp_err_t patient_store_save(const patient_info_t *info) {
    if (!info || !info->patient_id[0]) {
        return ESP_ERR_INVALID_ARG;
    }

    // Encrypt patient data before storing
    uint8_t encrypted[256];
    uint16_t encrypted_len = sizeof(encrypted);
    payload_encrypt((const uint8_t *)info, sizeof(patient_info_t), encrypted, &encrypted_len);

    esp_err_t err = nvs_set_blob(s_nvs_handle, "patient_data", encrypted, encrypted_len);
    if (err == ESP_OK) {
        nvs_commit(s_nvs_handle);
        ESP_LOGI(TAG, "Patient %s stored", info->patient_id);
    }
    return err;
}

esp_err_t patient_store_load(patient_info_t *info) {
    if (!info) return ESP_ERR_INVALID_ARG;

    size_t required_size = 0;
    esp_err_t err = nvs_get_blob(s_nvs_handle, "patient_data", NULL, &required_size);
    if (err != ESP_OK || required_size == 0) {
        memset(info, 0, sizeof(patient_info_t));
        strncpy(info->patient_id, "UNBOUND", sizeof(info->patient_id));
        return ESP_ERR_NOT_FOUND;
    }

    uint8_t *encrypted = malloc(required_size);
    if (!encrypted) return ESP_ERR_NO_MEM;

    err = nvs_get_blob(s_nvs_handle, "patient_data", encrypted, &required_size);
    if (err == ESP_OK) {
        uint8_t decrypted[256];
        uint16_t decrypted_len = sizeof(decrypted);
        if (payload_decrypt(encrypted, required_size, decrypted, &decrypted_len) == ESP_OK) {
            memcpy(info, decrypted, sizeof(patient_info_t));
        }
    }

    free(encrypted);
    return err;
}

esp_err_t patient_store_clear(void) {
    esp_err_t err = nvs_erase_key(s_nvs_handle, "patient_data");
    if (err == ESP_OK) {
        nvs_commit(s_nvs_handle);
        ESP_LOGI(TAG, "Patient data cleared");
    }
    return err;
}

void patient_store_deinit(void) {
    nvs_close(s_nvs_handle);
}

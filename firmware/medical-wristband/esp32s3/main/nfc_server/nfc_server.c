/*
 * Eregen Medical Wristband - NFC Server Module
 * Handles NFC A protocol for nurse terminal verification
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "nfc_server.h"
#include <string.h>
#include <stdio.h>
#include "esp_log.h"
#include "nfc_reader.h"
#include "nfc_llcp.h"
#include "nfc_pal.h"

static const char *TAG = "nfc_server";
static nfc_callback_t s_callback = NULL;
static void *s_user_data = NULL;

esp_err_t nfc_server_init(nfc_callback_t callback, void *user_data) {
    s_callback = callback;
    s_user_data = user_data;

    // Initialize NFC reader
    nfc_reader_config_t config = {
        .mode = NFC_READ_WRITE,
        .timeout_ms = 5000,
    };

    esp_err_t err = nfc_reader_init(&config);
    if (err == ESP_OK) {
        ESP_LOGI(TAG, "NFC server initialized");
    }
    return err;
}

esp_err_t nfc_server_start(void) {
    return nfc_reader_start();
}

void nfc_server_stop(void) {
    nfc_reader_stop();
}

esp_err_t nfc_server_read_tag(nfc_tag_info_t *tag_info) {
    if (!tag_info) return ESP_ERR_INVALID_ARG;

    esp_err_t err = nfc_reader_read_tag(tag_info);
    if (err == ESP_OK && s_callback) {
        s_callback(NFC_EVENT_TAG_READ, tag_info, s_user_data);
    }
    return err;
}

esp_err_t nfc_server_write_tag(const nfc_tag_info_t *tag_info, const uint8_t *data, size_t len) {
    if (!tag_info || !data) return ESP_ERR_INVALID_ARG;

    esp_err_t err = nfc_reader_write_tag(tag_info, data, len);
    if (err == ESP_OK && s_callback) {
        s_callback(NFC_EVENT_TAG_WRITE, tag_info, s_user_data);
    }
    return err;
}

void nfc_server_set_callback(nfc_callback_t callback, void *user_data) {
    s_callback = callback;
    s_user_data = user_data;
}

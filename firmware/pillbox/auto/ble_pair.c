/*
 * Eregen (颐贞) - BLE Pairing Implementation
 * Smart pillbox tier — receive home WiFi credentials from smartphone APP via BLE GATT.
 * Uses ESP-IDF BLE stack on ESP32-C3.
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 */

#include "ble_pair.h"

#include <string.h>
#include <stdio.h>
#include <stdlib.h>

#include "esp_log.h"
#include "esp_bt.h"
#include "esp_bt_main.h"
#include "esp_gap_ble_api.h"
#include "esp_gatts_api.h"
#include "esp_gatt_defs.h"
#include "esp_bt_defs.h"
#include "nvs_flash.h"

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/event_groups.h"

static const char *TAG = "ble_pair";

/* ---- BLE constants ---- */

#define SERVICE_UUID        0xFF00   /* Custom service UUID */
#define CREDENTIAL_CHAR_UUID 0xFF01  /* Write-only characteristic */

/* Device name suffix length */
#define NAME_SUFFIX_LEN     4

/* NVS keys */
#define NVS_KEY_WIFI_SSID   "wifi_ssid"
#define NVS_KEY_WIFI_PASS   "wifi_pass"

/* ---- Internal state ---- */

static bool s_ble_active = false;
static bool s_credentials_valid = false;
static char s_saved_ssid[33] = {0};
static char s_saved_password[65] = {0};

/* GATT handles */
static uint16_t s_service_handle = 0;
static uint16_t s_cred_char_handle = 0;

/* Event group for BLE callbacks */
static EventGroupHandle_t s_ble_events;
#define BLE_CRED_RECEIVED_BIT    BIT0
#define BLE_PAIRING_DONE_BIT     BIT1

/* ---- Forward declarations ---- */

static void gap_event_handler(esp_gap_ble_cb_event_t event,
                              esp_ble_gap_cb_param_t *param);
static void gatts_event_handler(esp_gatts_cb_event_t event,
                                esp_gatt_if_t gatts_if,
                                esp_ble_gatts_cb_param_t *param);
static void extract_name_suffix(char *suffix, size_t len);
static esp_err_t parse_credentials(const uint8_t *data, uint16_t len,
                                   char *ssid_out, char *pass_out);

/**
 * Start BLE advertising for credential pairing.
 * Runs in its own task to avoid blocking app_main().
 */
void ble_pair_start(void)
{
    if (s_ble_active) return;

    s_ble_events = xEventGroupCreate();
    if (!s_ble_events) {
        ESP_LOGE(TAG, "Failed to create BLE event group");
        return;
    }

    xTaskCreate(ble_pair_task, "ble_pair", 4096, NULL, 4, NULL);
}

/**
 * Stop BLE advertising and free resources.
 */
void ble_pair_stop(void)
{
    s_ble_active = false;

    esp_ble_gap_stop_advertising();

    if (s_ble_events) {
        vEventGroupDelete(s_ble_events);
        s_ble_events = NULL;
    }

    ESP_LOGI(TAG, "BLE pairing stopped");
}

bool ble_pair_has_credentials(void)
{
    return s_credentials_valid;
}

bool ble_pair_get_credentials(char *ssid, char *password, size_t max_len)
{
    if (!s_credentials_valid) return false;

    if (ssid) {
        strncpy(ssid, s_saved_ssid, max_len - 1);
        ssid[max_len - 1] = '\0';
    }
    if (password) {
        strncpy(password, s_saved_password, max_len - 1);
        password[max_len - 1] = '\0';
    }
    return true;
}

bool ble_pair_save_to_nvs(void)
{
    if (!s_credentials_valid) return false;

    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READWRITE, &handle);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS open failed: %s", esp_err_to_name(ret));
        return false;
    }

    ret = nvs_set_str(handle, NVS_KEY_WIFI_SSID, s_saved_ssid);
    if (ret == ESP_OK) {
        ret = nvs_set_str(handle, NVS_KEY_WIFI_PASS, s_saved_password);
    }
    if (ret == ESP_OK) {
        ret = nvs_commit(handle);
    }
    nvs_close(handle);

    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS save failed: %s", esp_err_to_name(ret));
        return false;
    }

    ESP_LOGI(TAG, "Credentials saved to NVS via BLE");
    return true;
}

/* ---- BLE task ---- */

static void ble_pair_task(void *pvParameter)
{
    (void)pvParameter;

    /* Initialize NVS (may already be done by app_main) */
    esp_err_t ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES ||
        ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        nvs_flash_erase();
        ret = nvs_flash_init();
    }
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS init failed: %s", esp_err_to_name(ret));
        vTaskDelete(NULL);
        return;
    }

    /* Initialize Bluetooth controller */
    esp_bt_controller_config_t bt_cfg = BT_CONTROLLER_INIT_CONFIG_DEFAULT();
    ret = esp_bt_controller_init(&bt_cfg);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "BT controller init failed: %s", esp_err_to_name(ret));
        vTaskDelete(NULL);
        return;
    }

    ret = esp_bt_controller_enable(ESP_BT_MODE_BLE);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "BT controller enable failed: %s", esp_err_to_name(ret));
        vTaskDelete(NULL);
        return;
    }

    ret = esp_bluedroid_init();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Bluedroid init failed: %s", esp_err_to_name(ret));
        vTaskDelete(NULL);
        return;
    }

    ret = esp_bluedroid_enable();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Bluedroid enable failed: %s", esp_err_to_name(ret));
        vTaskDelete(NULL);
        return;
    }

    /* Register GAP and GATTS callbacks */
    esp_ble_gap_register_callback(gap_event_handler);
    esp_ble_gatts_register_callback(gatts_event_handler);

    /* Build device name with MAC suffix */
    char device_name[24];
    char suffix[NAME_SUFFIX_LEN + 1];
    extract_name_suffix(suffix, sizeof(suffix));
    snprintf(device_name, sizeof(device_name), "eregen-pair-%s", suffix);
    esp_bt_dev_set_device_name(device_name);

    /* Configure advertisement parameters */
    esp_ble_adv_params_t adv_params = {
        .adv_int_min    = 0x20,
        .adv_int_max    = 0x40,
        .adv_type       = ADV_TYPE_IND,
        .own_addr_type  = BLE_ADDR_TYPE_PUBLIC,
        .channel_map    = ADV_CHNL_ALL,
        .filter_policy  = ADV_FILTER_ALLOW_SCAN_ANY_CON_ANY,
    };

    /* Configure advertisement data */
    uint8_t adv_data[31] = { 0 };
    int offset = 0;

    /* Flags */
    adv_data[offset++] = 3;
    adv_data[offset++] = 0x01;
    adv_data[offset++] = ESP_BLE_FLAG_GEN_DISC_STD;

    /* Complete Name */
    int name_len = strlen(device_name);
    adv_data[offset++] = name_len + 1;
    adv_data[offset++] = 0x09;
    memcpy(&adv_data[offset], device_name, name_len);
    offset += name_len;

    /* Set adv data */
    esp_ble_gap_config_adv_data((uint8_t *)adv_data, offset);

    /* Start advertising */
    esp_ble_gap_start_advertising(&adv_params);

    s_ble_active = true;
    ESP_LOGI(TAG, "BLE advertising started: %s", device_name);

    /* Wait for credentials to be received */
    EventBits_t bits = xEventGroupWaitBits(s_ble_events,
                                            BLE_CRED_RECEIVED_BIT,
                                            pdFALSE, pdFALSE,
                                            portMAX_DELAY);

    if (bits & BLE_CRED_RECEIVED_BIT) {
        ESP_LOGI(TAG, "Credentials received via BLE");
        xEventGroupSetBits(s_ble_events, BLE_PAIRING_DONE_BIT);
    }

    /* Stop advertising after credentials received */
    ble_pair_stop();
    vTaskDelete(NULL);
}

/* ---- GAP callback ---- */

static void gap_event_handler(esp_gap_ble_cb_event_t event,
                              esp_ble_gap_cb_param_t *param)
{
    switch (event) {
    case ESP_GAP_BLE_ADV_DATA_SET_COMPLETE_EVT:
        break;

    case ESP_GAP_BLE_ADV_START_COMPLETE_EVT:
        if (param->adv_start_cmpl.status != ESP_BT_STATUS_SUCCESS) {
            ESP_LOGE(TAG, "Advertising start failed: %d",
                     param->adv_start_cmpl.status);
        }
        break;

    default:
        break;
    }
}

/* ---- GATTS callback ---- */

static void gatts_event_handler(esp_gatts_cb_event_t event,
                                esp_gatt_if_t gatts_if,
                                esp_ble_gatts_cb_param_t *param)
{
    switch (event) {
    case ESP_GATTS_REG_EVT: {
        if (param->reg.status == ESP_BT_STATUS_SUCCESS) {
            s_service_handle = param->reg.service_handle;
            /* Create the service */
            esp_ble_gatts_start_service(s_service_handle);
        }
        break;
    }

    case ESP_GATTS_CREAT_ATTR_TAB_EVT: {
        if (param->add_attr_tab.status != ESP_BT_STATUS_SUCCESS) {
            ESP_LOGE(TAG, "Service table creation failed");
            break;
        }

        uint16_t handle = param->add_attr_tab.srvc_handle;
        uint16_t start_handle = param->add_attr_tab.start_handle;

        /* Add credential characteristic attribute table */
        uint16_t attr_tab_handle[start_handle + 1];
        for (int i = 0; i < start_handle; i++) {
            attr_tab_handle[i] = start_handle + i;
        }
        attr_tab_handle[start_handle] = start_handle + start_handle;

        esp_ble_gatts_add_attr_table(handle, attr_tab_handle, 1, SERVICE_UUID);
        break;
    }

    case ESP_GATTS_ADD_INCL_SRVC_EVT:
        break;

    case ESP_GATTS_ADD_CHAR_EVT: {
        if (param->add_char.status != ESP_BT_STATUS_SUCCESS) break;

        uint16_t uuid = param->add_char.attr_desc.att_uuid.uuid.uuid16;
        if (uuid == CREDENTIAL_CHAR_UUID) {
            s_cred_char_handle = param->add_char.handle;
            ESP_LOGI(TAG, "Credential characteristic added: handle=%d",
                     s_cred_char_handle);

            /* Now configure read/write permissions on the characteristic */
            uint16_t uuid_arr[] = {CREDENTIAL_CHAR_UUID, 0};
            esp_gatt_perm_t perm = (ESP_GATT_PERM_READ |
                                    ESP_GATT_PERM_WRITE_ENCRYPTED);
            esp_gatt_char_prop_t prop = ESP_GATT_CHAR_PROP_BIT_WRITE;

            esp_ble_gatts_add_char(s_service_handle,
                                   &uuid_arr[0], perm, prop,
                                   NULL, NULL);
        }
        break;
    }

    case ESP_GATTS_ADD_CHAR_DESCR_EVT:
        break;

    case ESP_GATTS_WRITE_EVT: {
        if (param->write.handle != s_cred_char_handle) {
            break;
        }

        if (!param->write.is_prep) {
            /* Handle credential write */
            char ssid[33] = {0};
            char pass[65] = {0};

            esp_err_t ret = parse_credentials(param->write.value,
                                              param->write.len,
                                              ssid, pass);
            if (ret == ESP_OK && strlen(ssid) > 0) {
                strncpy(s_saved_ssid, ssid, sizeof(s_saved_ssid) - 1);
                s_saved_ssid[sizeof(s_saved_ssid) - 1] = '\0';
                strncpy(s_saved_password, pass, sizeof(s_saved_password) - 1);
                s_saved_password[sizeof(s_saved_password) - 1] = '\0';
                s_credentials_valid = true;

                ESP_LOGI(TAG, "BLE credentials: SSID=%s", s_saved_ssid);

                /* Send success response */
                uint8_t resp[] = {0x01};  /* status: success */
                esp_ble_gatts_send_response(gatts_if,
                                            param->write.trans_id,
                                            param->write.handle,
                                            ESP_GATT_OK);

                xEventGroupSetBits(s_ble_events, BLE_CRED_RECEIVED_BIT);
            } else {
                /* Send error response */
                esp_ble_gatts_send_response(gatts_if,
                                            param->write.trans_id,
                                            param->write.handle,
                                            ESP_GATT_INVALID_OFFSET);
            }
        }
        break;
    }

    case ESP_GATTS_CONNECT_EVT: {
        esp_bd_addr_t bd_addr;
        memcpy(bd_addr, param->connect.remote_bda, sizeof(esp_bd_addr_t));
        ESP_LOGI(TAG, "BLE connected to %02X:%02X:%02X:%02X:%02X:%02X",
                 bd_addr[0], bd_addr[1], bd_addr[2],
                 bd_addr[3], bd_addr[4], bd_addr[5]);
        break;
    }

    case ESP_GATTS_DISCONNECT_EVT: {
        ESP_LOGI(TAG, "BLE disconnected, restarting advertising");
        /* Restart advertising on disconnect */
        esp_ble_gap_start_advertising();
        break;
    }

    default:
        break;
    }
}

/* ---- Helpers ---- */

/**
 * Parse credential data from BLE write.
 * Accepts JSON format: {"ssid":"...","password":"..."}
 * or simple format: SSID<TAB>Password
 */
static esp_err_t parse_credentials(const uint8_t *data, uint16_t len,
                                   char *ssid_out, char *pass_out)
{
    if (!data || len == 0 || !ssid_out || !pass_out) {
        return ESP_ERR_INVALID_ARG;
    }

    /* Try JSON format first */
    if (data[0] == '{') {
        char json_buf[128];
        if (len >= sizeof(json_buf)) return ESP_ERR_NO_MEM;
        memcpy(json_buf, data, len);
        json_buf[len] = '\0';

        char *ssid_pos = strstr(json_buf, "\"ssid\"");
        char *pass_pos = strstr(json_buf, "\"password\"");

        if (ssid_pos) {
            ssid_pos = strchr(ssid_pos, ':');
            if (ssid_pos) {
                ssid_pos++;
                while (*ssid_pos == ' ' || *ssid_pos == '"') ssid_pos++;
                char *end = strchr(ssid_pos, '"');
                if (end) {
                    size_t slen = end - ssid_pos;
                    if (slen > 32) slen = 32;
                    strncpy(ssid_out, ssid_pos, slen);
                    ssid_out[slen] = '\0';
                }
            }
        }
        if (pass_pos) {
            pass_pos = strchr(pass_pos, ':');
            if (pass_pos) {
                pass_pos++;
                while (*pass_pos == ' ' || *pass_pos == '"') pass_pos++;
                char *end = strchr(pass_pos, '"');
                if (end) {
                    size_t plen = end - pass_pos;
                    if (plen > 64) plen = 64;
                    strncpy(pass_out, pass_pos, plen);
                    pass_out[plen] = '\0';
                }
            }
        }

        if (strlen(ssid_out) > 0) return ESP_OK;
    }

    /* Fallback: tab-separated format */
    char buf[98];
    if (len >= sizeof(buf)) return ESP_ERR_NO_MEM;
    memcpy(buf, data, len);
    buf[len] = '\0';

    char *tab = strchr(buf, '\t');
    if (tab) {
        *tab = '\0';
        strncpy(ssid_out, buf, 32);
        ssid_out[32] = '\0';
        strncpy(pass_out, tab + 1, 64);
        pass_out[64] = '\0';
        return ESP_OK;
    }

    /* Single string = SSID only, empty password */
    strncpy(ssid_out, buf, 32);
    ssid_out[32] = '\0';
    pass_out[0] = '\0';

    return strlen(ssid_out) > 0 ? ESP_OK : ESP_FAIL;
}

static void extract_name_suffix(char *suffix, size_t len)
{
    uint8_t mac[6];
    esp_read_mac(mac, ESP_MAC_BLUEDROID);
    snprintf(suffix, len, "%02X%02X", mac[4], mac[5]);
}

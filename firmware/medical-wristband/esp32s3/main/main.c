/*
 * Eregen Medical Wristband - ESP32-S3 Firmware
 * NFC身份核验 + Cat1上报双模式医用电子腕带
 *
 * 功能:
 *   1. BLE GATT Server — 护士终端 NFC/蓝牙核验
 *   2. Cat1 蜂窝上报 — 心跳/定位/告警上报云端
 *   3. NVS持久化 — 患者ID/腕带ID/固件版本
 *   4. SOS 物理按钮 — 紧急呼叫
 *   5. 电池监测 — 低功耗管理
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/event_groups.h"
#include "esp_system.h"
#include "esp_log.h"
#include "nvs_flash.h"
#include "nvs.h"

/* BLE */
#include "esp_gap_ble_api.h"
#include "esp_gatts_api.h"
#include "esp_bt_main.h"
#include "esp_bt_device.h"

/* MQTT + Wi-Fi (Cat1 fallback via AT commands on UART) */
#include "mqtt_common.h"
#include "wifi_mqtt_bridge.h"
#include "payload_crypto.h"

/* Common */
#include "ota_handler.h"
#include "brand_boot_logo.h"

static const char *TAG = "eregen_wb";

/* ---- NVS keys ---- */
#define NVS_NAMESPACE "med_wb"
#define NVS_KEY_DEV_ID    "dev_id"
#define NVS_KEY_PATIENT   "patient"
#define NVS_KEY_FW_VER    "fw_ver"

/* ---- BLE GATT service UUIDs ---- */
#define WB_SERVICE_UUID     0x1812  /* Heart Rate Service (standard) */
#define WB_CHAR_STATUS      0x2A38  /* Heart Rate Measurement */
#define WB_CHAR_PATIENT     0xFF01  /* Custom: patient info (NFC read) */
#define WB_CHAR_SOS         0xFF02  /* Custom: SOS trigger */
#define WB_CHAR_CONFIG      0xFF03  /* Custom: config write (nurse push) */

/* ---- Event group bits ---- */
#define GOT_IP_BIT          (1 << 0)
#define BLE_INIT_BIT        (1 << 1)
#define MQTT_CONNECTED_BIT  (1 << 2)
#define SOS_TRIGGERED_BIT   (1 << 3)

static EventGroupHandle_t s_event_group;
static payload_crypto_ctx_t s_crypto_ctx;
static bool s_mqtt_ready = false;

/* Forward declarations */
static void ble_gap_event_handler(esp_gap_ble_cb_event_t event, esp_ble_gap_cb_param_t *param);
static void ble_gatts_event_handler(esp_gatts_cb_event_t event, esp_gatt_cb_param_t *param);
static void ble_start_advertising(void);
static esp_err_t init_profile_db(void);
static void wifi_mqtt_task(void *pvParameters);
static void ble_task(void *pvParameters);
static void sntp_time_task(void *pvParameters);
static void log_device_info(void);

/* ---- GATT table (simplified) ---- */
enum {
    WB_IDX_SVC,
    WB_IDX_STATUS_CHR,
    WB_IDX_STATUS_VAL,
    WB_IDX_PATIENT_CHR,
    WB_IDX_PATIENT_VAL,
    WB_IDX_SOS_CHR,
    WB_IDX_SOS_VAL,
    WB_IDX_CONFIG_CHR,
    WB_IDX_CONFIG_VAL,
    WB_IDX_NB
};

/* Service UUID */
static esp_gatt_srvc_id_t wb_srvc_id = {
    .id = {.uuid = {{ESP_UUID_LEN_16, WB_SERVICE_UUID}}, .inst_id = 0x00},
    .start_handle = 0, .end_handle = 0,
};

/* Characteristic descriptors */
static const uint16_t wb_char_declARATION = ESP_GATT_PERM_READ;
static uint8_t s_heart_rate_val = 0;     /* placeholder, updated by health sensor */
static char  s_patient_id[64] = {0};
static uint8_t s_sos_flag = 0;
static uint8_t s_config_buf[128] = {0};

/* GATT table */
static const uint16_t wb_primary_service_start = 0x0001;
static const uint16_t wb_primary_service_end   = 0x000A;

static esp_gatts_db_t wb_gatts_db = {
    .attr_tbl = NULL,
    .tbl_size  = 0,
};

/* ---- NVS helper ---- */
static esp_err_t nvs_set_str(const char *key, const char *val) {
    nvs_handle_t handle;
    esp_err_t err = nvs_open(NVS_NAMESPACE, NVS_READWRITE, &handle);
    if (err != ESP_OK) return err;
    err = nvs_set_str(handle, key, val);
    nvs_commit(handle);
    nvs_close(handle);
    return err;
}

static esp_err_t nvs_get_str(const char *key, char *out, size_t *len) {
    nvs_handle_t handle;
    esp_err_t err = nvs_open(NVS_NAMESPACE, NVS_READONLY, &handle);
    if (err != ESP_OK) return err;
    err = nvs_get_str(handle, key, out, len);
    nvs_close(handle);
    return err;
}

/* ---- GATT Handler ---- */
static void ble_gatts_event_handler(esp_gatts_cb_event_t event, esp_gatt_cb_param_t *param) {
    switch (event) {
        case ESP_GATTS_REG_EVT: {
            esp_ble_gatts_create_table(wb_gatts_db.attr_tbl, &wb_gatts_db);
            esp_ble_gatts_start_service(wb_srvc_id.id.inst_id);
            break;
        }
        case ESP_GATTS_CREATE_EVT:
            break;
        case ESP_GATTS_ADD_CHAR_EVT: {
            uint16_t handle = param->add_char.result.handle;
            switch (handle) {
                case WB_IDX_STATUS_VAL:
                    ESP_LOGI(TAG, "HR char handle=%d", handle);
                    break;
                case WB_IDX_PATIENT_VAL:
                    ESP_LOGI(TAG, "Patient char handle=%d", handle);
                    break;
                case WB_IDX_SOS_VAL:
                    ESP_LOGI(TAG, "SOS char handle=%d", handle);
                    break;
                case WB_IDX_CONFIG_VAL:
                    ESP_LOGI(TAG, "Config char handle=%d", handle);
                    break;
            }
            break;
        }
        case ESP_GATTS_START_EVT:
            ble_start_advertising();
            break;
        case ESP_GATTS_READ_EVT: {
            if (param->read.handle == WB_IDX_STATUS_VAL) {
                /* Return current heart rate */
                esp_ble_gatts_send_response(param->read.conn_id,
                    param->read.trans_id, ESP_GATT_OK, &s_heart_rate_val, 1);
            } else if (param->read.handle == WB_IDX_PATIENT_VAL) {
                uint16_t len = strlen(s_patient_id);
                esp_ble_gatts_send_response(param->read.conn_id,
                    param->read.trans_id, ESP_GATT_OK,
                    (uint8_t *)s_patient_id, len);
            } else if (param->read.handle == WB_IDX_SOS_VAL) {
                esp_ble_gatts_send_response(param->read.conn_id,
                    param->read.trans_id, ESP_GATT_OK, &s_sos_flag, 1);
            }
            break;
        }
        case ESP_GATTS_WRITE_EVT: {
            if (param->write.handle == WB_IDX_CONFIG_VAL) {
                uint16_t len = param->write.len;
                if (len > sizeof(s_config_buf)) len = sizeof(s_config_buf);
                memcpy(s_config_buf, param->write.value, len);
                s_config_buf[len] = '\0';
                ESP_LOGI(TAG, "Config received: %.64s...", s_config_buf);
                /* Handle config: update MQTT broker, re-join network, etc. */
            }
            break;
        }
        case ESP_GATTS_CONNECT_EVT:
            ESP_LOGI(TAG, "BLE connected from %s",
                     param->connect.bda);
            break;
        case ESP_GATTS_DISCONNECT_EVT:
            ESP_LOGI(TAG, "BLE disconnected, restarting advertising");
            ble_start_advertising();
            break;
        default:
            break;
    }
}

/* ---- GAP Handler ---- */
static void ble_gap_event_handler(esp_gap_ble_cb_event_t event, esp_gap_ble_cb_param_t *param) {
    switch (event) {
        case ESP_GAP_BLE_ADV_DATA_SET_COMPLETE_EVT:
            esp_ble_gap_start_advertising(&param->adv_cmpl.params);
            break;
        case ESP_GAP_BLE_ADV_START_COMPLETE_EVT:
            xEventGroupSetBits(s_event_group, BLE_INIT_BIT);
            break;
        default:
            break;
    }
}

/* ---- Build and start BLE advertising ---- */
static void ble_start_advertising(void) {
    esp_ble_adv_data_t adv_data = {0};
    uint8_t manuf_data[] = {0x0E, 0xFF, 'E', 'R', 'E', 'G', 'E', 'N', 0x01}; /* company ID + type */
    adv_data.set_scan_rsp = false;
    adv_data.include_name = true;
    adv_data.include_txpower = true;
    adv_data.min_interval = 0x20;
    adv_data.max_interval = 0x40;
    adv_data.manuf_data_len = sizeof(manuf_data);
    adv_data.p_manuf_data = manuf_data;

    esp_ble_gap_config_adv_data(&adv_data);
}

/* ---- BLE profile initialization ---- */
static esp_err_t init_profile_db(void) {
    /* Create GATT service and characteristics */
    esp_ble_gatts_cb_param_t create_param = {0};

    /* Primary service */
    esp_ble_gatts_create_service(gattc_if, &wb_srvc_id, 16); /* 16 handles */

    /* Heart Rate char */
    uint16_t hr_char_uuid = WB_CHAR_STATUS;
    uint16_t hr_val_handle;
    esp_ble_gatts_add_char(wb_srvc_id.id.inst_id,
        &hr_char_uuid, ESP_GATT_PERM_READ,
        ESP_GATT_AVAIL_ANY, &hr_val_handle, NULL);

    /* Patient ID char (read by nurse terminal) */
    uint16_t patient_char_uuid = WB_CHAR_PATIENT;
    uint16_t patient_val_handle;
    esp_ble_gatts_add_char(wb_srvc_id.id.inst_id,
        &patient_char_uuid, ESP_GATT_PERM_READ,
        ESP_GATT_AVAIL_ANY, &patient_val_handle, NULL);

    /* SOS char */
    uint16_t sos_char_uuid = WB_CHAR_SOS;
    uint16_t sos_val_handle;
    esp_ble_gatts_add_char(wb_srvc_id.id.inst_id,
        &sos_char_uuid, ESP_GATT_PERM_READ,
        ESP_GATT_AVAIL_ANY, &sos_val_handle, NULL);

    /* Config char (write by nurse terminal) */
    uint16_t config_char_uuid = WB_CHAR_CONFIG;
    uint16_t config_val_handle;
    esp_ble_gatts_add_char(wb_srvc_id.id.inst_id,
        &config_char_uuid, ESP_GATT_PERM_READ | ESP_GATT_PERM_WRITE,
        ESP_GATT_AVAIL_ANY, &config_val_handle, NULL);

    /* Add char declaration */
    esp_ble_gatts_add_char_descr(wb_srvc_id.id.inst_id,
        &hr_char_uuid, ESP_GATT_PERM_READ, &hr_val_handle, NULL);
    esp_ble_gatts_add_char_descr(wb_srvc_id.id.inst_id,
        &patient_char_uuid, ESP_GATT_PERM_READ, &patient_val_handle, NULL);
    esp_ble_gatts_add_char_descr(wb_srvc_id.id.inst_id,
        &sos_char_uuid, ESP_GATT_PERM_READ, &sos_val_handle, NULL);
    esp_ble_gatts_add_char_descr(wb_srvc_id.id.inst_id,
        &config_char_uuid, ESP_GATT_PERM_READ | ESP_GATT_PERM_WRITE,
        &config_val_handle, NULL);

    return ESP_OK;
}

/* ---- BLE task ---- */
static void ble_task(void *pvParameters) {
    ESP_LOGI(TAG, "BLE task started");

    /* Register GAP callback */
    esp_ble_gap_register_cb(ble_gap_event_handler);

    /* Register GATTS callback */
    esp_ble_gatts_register_callback(ble_gatts_event_handler);

    /* Set BLE device name */
    esp_ble_set_device_name("Eregen-WB");

    /* Create GATT service */
    init_profile_db();

    /* Load patient ID from NVS */
    size_t len = sizeof(s_patient_id);
    nvs_get_str(NVS_KEY_PATIENT, s_patient_id, &len);
    ESP_LOGI(TAG, "Patient ID: %s", s_patient_id);

    xEventGroupSetBits(s_event_group, BLE_INIT_BIT);

    /* Start advertising */
    ble_start_advertising();

    /* BLE task runs forever — esp_ble_gatts* calls are event-driven */
    while (1) {
        vTaskDelay(pdMS_TO_TICKS(5000));
        /* Periodic status log */
        ESP_LOGI(TAG, "BLE active, connected=%d",
                 esp_ble_get_conn_count());
    }
}

/* ---- Wi-Fi + MQTT task ---- */
static void wifi_mqtt_task(void *pvParameters) {
    ESP_LOGI(TAG, "WiFi/MQTT task started");

    /* Initialize payload crypto */
    payload_crypto_init(&s_crypto_ctx);

    /* Wait for BLE to initialize first */
    xEventGroupWaitBits(s_event_group, BLE_INIT_BIT, false, true, pdMS_TO_TICKS(10000));

    /* Initialize MQTT common (uses Wi-Fi or fallback to Cat1) */
    const char *client_id = "eregen-wb-0001";
    nvs_get_str(NVS_KEY_DEV_ID, (char *)&client_id, &(uint16_t){0});

    esp_err_t ret = mqtt_common_connect(
        "mqtt.eregen.dev", 8883,
        client_id, NULL, NULL, NULL);  /* no TLS pinning in dev */

    if (ret == 0) {
        s_mqtt_ready = true;
        xEventGroupSetBits(s_event_group, MQTT_CONNECTED_BIT);
        ESP_LOGI(TAG, "MQTT connected");

        /* Subscribe to device command topic */
        char subscribe_topic[64];
        snprintf(subscribe_topic, sizeof(subscribe_topic),
                 "eregen/device/wb/%s/cmd", client_id);
        mqtt_common_subscribe(subscribe_topic, NULL);

        /* Publish heartbeat */
        uint32_t tick = xTaskGetTickCount();
        char heartbeat[128];
        snprintf(heartbeat, sizeof(heartbeat),
                 "{\"type\":\"heartbeat\",\"dev_id\":\"%s\",\"bat\":100,\"ts\":%lu}",
                 client_id, (unsigned long)tick);
        mqtt_common_publish("eregen/device/wb/heartbeat", heartbeat,
                            strlen(heartbeat), 1);
        ESP_LOGI(TAG, "Heartbeat published: %s", heartbeat);
    } else {
        ESP_LOGE(TAG, "MQTT connect failed: %d", ret);
    }

    /* Main MQTT event loop — esp-mqtt handles this internally */
    vTaskDelete(NULL);
}

/* ---- SNTP time sync task ---- */
static void sntp_time_task(void *pvParameters) {
    ESP_LOGI(TAG, "SNTP task started");
    sntp_setservername(0, "ntp.aliyun.com");
    sntp_setservername(1, "ntp.tencent.com");
    sntp_init();

    /* Wait for time sync */
    while (!sntp_get_sync_status() == SNTP_SYNC_STATUS_COMPLETED) {
        vTaskDelay(pdMS_TO_TICKS(10000));
    }
    ESP_LOGI(TAG, "Time synced via SNTP");
    vTaskDelete(NULL);
}

/* ---- Main entry ---- */
void app_main(void) {
    ESP_LOGI(TAG, "Eregen Medical Wristband v1.0.0");
    ESP_LOGI(TAG, "Target: ESP32-S3, ESP-IDF v5.3");

    /* Show brand boot logo */
    brand_boot_logo_show();

    /* Initialize NVS */
    esp_err_t ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES ||
        ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        ESP_ERROR_CHECK(nvs_flash_erase());
        ret = nvs_flash_init();
    }
    ESP_ERROR_CHECK(ret);

    /* Generate or load device ID */
    uint8_t mac[6];
    esp_read_mac(mac, ESP_MAC_BT);
    char dev_id[17];
    snprintf(dev_id, sizeof(dev_id), "WB-%02X%02X%02X",
             mac[3], mac[4], mac[5]);
    nvs_set_str(NVS_KEY_DEV_ID, dev_id);
    ESP_LOGI(TAG, "Device ID: %s", dev_id);

    /* Initialize event group */
    s_event_group = xEventGroupCreate();
    if (s_event_group == NULL) {
        ESP_LOGE(TAG, "Failed to create event group");
        abort();
    }

    /* Load patient ID (set by nurse terminal during binding) */
    size_t plen = sizeof(s_patient_id);
    nvs_get_str(NVS_KEY_PATIENT, s_patient_id, &plen);
    if (strlen(s_patient_id) == 0) {
        strncpy(s_patient_id, "UNBOUND", sizeof(s_patient_id));
    }
    ESP_LOGI(TAG, "Patient: %s", s_patient_id);

    /* Start BLE task */
    xTaskCreate(ble_task, "ble", 4096, NULL, 5, NULL);

    /* Start Wi-Fi/MQTT task */
    xTaskCreate(wifi_mqtt_task, "mqtt", 6144, NULL, 4, NULL);

    /* Start SNTP time sync */
    xTaskCreate(sntp_time_task, "sntp", 2048, NULL, 3, NULL);

    /* Log device info */
    log_device_info();

    /* Main loop */
    while (1) {
        EventBits_t bits = xEventGroupGetBits(s_event_group);

        if (bits & MQTT_CONNECTED_BIT) {
            /* Publish periodic heartbeat */
            static uint32_t s_last_heartbeat = 0;
            uint32_t now = xTaskGetTickCount();
            if (now - s_last_heartbeat > pdMS_TO_TICKS(60000)) {
                s_last_heartbeat = now;
                char msg[128];
                snprintf(msg, sizeof(msg),
                         "{\"type\":\"heartbeat\",\"dev_id\":\"%s\","
                         "\"bat\":%d,\"ts\":%lu}",
                         dev_id, 95, (unsigned long)now);
                mqtt_common_publish("eregen/device/wb/heartbeat",
                                    msg, strlen(msg), 1);
                ESP_LOGI(TAG, "Heartbeat: %s", msg);
            }
        }

        /* Check for SOS (via GPIO or BLE write) */
        if (xEventGroupGetBits(s_event_group) & SOS_TRIGGERED_BIT) {
            xEventGroupClearBits(s_event_group, SOS_TRIGGERED_BIT);
            s_sos_flag = 1;
            ESP_LOGW(TAG, "SOS ALERT triggered!");
            /* Publish SOS to MQTT */
            char sos_msg[256];
            snprintf(sos_msg, sizeof(sos_msg),
                     "{\"type\":\"sos\",\"dev_id\":\"%s\","
                     "\"patient_id\":\"%s\",\"ts\":%lu}",
                     dev_id, s_patient_id, (unsigned long)xTaskGetTickCount());
            mqtt_common_publish("eregen/device/wb/alert",
                                sos_msg, strlen(sos_msg), 1);
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

static void log_device_info(void) {
    ESP_LOGI(TAG, "Chip revision: %d", ESP_getEfuseRev());
    ESP_LOGI(TAG, "Free heap: %lu bytes",
             (unsigned long)esp_get_free_heap_size());
    ESP_LOGI(TAG, "Min free heap: %lu bytes",
             (unsigned long)esp_get_minimum_free_heap_size());
}

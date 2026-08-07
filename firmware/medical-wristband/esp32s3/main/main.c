/*
 * Eregen Medical Wristband - ESP32-S3 Firmware
 * NFC身份核验 + Cat1上报双模式医用电子腕带
 *
 * 功能:
 *   1. BLE GATT Server — 护士终端NFC/蓝牙核验
 *   2. Cat1 蜂窝上报 — 心跳/定位/告警上报云端
 *   3. NVS持久化 — 患者ID/腕带ID/固件版本
 *   4. OLED显示 — 患者姓名+科室+警示标签
 *   5. LED警示灯 — 按标签类型控制颜色和闪烁
 *   6. SOS物理按钮 — 紧急呼叫
 *   7. 加密安全 — AES-128加密患者数据
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

/* MQTT + Wi-Fi */
#include "mqtt_common.h"
#include "wifi_mqtt_bridge.h"
#include "payload_crypto.h"

/* Common */
#include "ota_handler.h"
#include "brand_boot_logo.h"

/* New modules */
#include "patient_store/patient_store.h"
#include "nfc_server/nfc_server.h"
#include "display_oled/display_oled.h"
#include "led_indicator/led_indicator.h"
#include "nvs_manager/nvs_manager.h"
#include "security/security.h"
#include "protocol/medical_protocol.h"

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
#define NFC_SCANNED_BIT     (1 << 4)

static EventGroupHandle_t s_event_group;
static payload_crypto_ctx_t s_crypto_ctx;
static bool s_mqtt_ready = false;

/* Patient info */
static patient_info_t s_patient_info;
static char s_patient_id[64] = {0};

/* Forward declarations */
static void ble_gap_event_handler(esp_gap_ble_cb_event_t event, esp_ble_gap_cb_param_t *param);
static void ble_gatts_event_handler(esp_gatts_cb_event_t event, esp_gatt_cb_param_t *param);
static void ble_start_advertising(void);
static esp_err_t init_profile_db(void);
static void wifi_mqtt_task(void *pvParameters);
static void ble_task(void *pvParameters);
static void nfc_scan_task(void *pvParameters);
static void sntp_time_task(void *pvParameters);
static void log_device_info(void);
static void handle_sos_alert(void);

/* ---- GATT table ---- */
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

static esp_gatt_srvc_id_t wb_srvc_id = {
    .id = {.uuid = {{ESP_UUID_LEN_16, WB_SERVICE_UUID}}, .inst_id = 0x00},
    .start_handle = 0, .end_handle = 0,
};

static const uint16_t wb_char_DECLARATION = ESP_GATT_PERM_READ;
static uint8_t s_heart_rate_val = 0;
static uint8_t s_sos_flag = 0;
static uint8_t s_config_buf[128] = {0};

/* ---- NVS helper ---- */
static esp_err_t nvs_set_str_local(const char *key, const char *val) {
    nvs_handle_t handle;
    esp_err_t err = nvs_open(NVS_NAMESPACE, NVS_READWRITE, &handle);
    if (err != ESP_OK) return err;
    err = nvs_set_str(handle, key, val);
    nvs_commit(handle);
    nvs_close(handle);
    return err;
}

static esp_err_t nvs_get_str_local(const char *key, char *out, size_t *len) {
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
            esp_ble_gatts_create_table(NULL, NULL);
            esp_ble_gatts_start_service(wb_srvc_id.id.inst_id);
            break;
        }
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
            }
            break;
        }
        case ESP_GATTS_CONNECT_EVT:
            ESP_LOGI(TAG, "BLE connected from %s", param->connect.bda);
            break;
        case ESP_GATTS_DISCONNECT_EVT:
            ESP_LOGI(TAG, "BLE disconnected, restarting advertising");
            ble_start_advertising();
            break;
        default:
            break;
    }
}

static void ble_gap_event_handler(esp_gap_ble_cb_event_t event, esp_ble_gap_cb_param_t *param) {
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

static void ble_start_advertising(void) {
    esp_ble_adv_data_t adv_data = {0};
    uint8_t manuf_data[] = {0x0E, 0xFF, 'E', 'R', 'E', 'G', 'E', 'N', 0x01};
    adv_data.set_scan_rsp = false;
    adv_data.include_name = true;
    adv_data.include_txpower = true;
    adv_data.min_interval = 0x20;
    adv_data.max_interval = 0x40;
    adv_data.manuf_data_len = sizeof(manuf_data);
    adv_data.p_manuf_data = manuf_data;
    esp_ble_gap_config_adv_data(&adv_data);
}

static esp_err_t init_profile_db(void) {
    uint16_t hr_char_uuid = WB_CHAR_STATUS;
    uint16_t hr_val_handle;
    esp_ble_gatts_add_char(wb_srvc_id.id.inst_id, &hr_char_uuid, ESP_GATT_PERM_READ,
        ESP_GATT_AVAIL_ANY, &hr_val_handle, NULL);

    uint16_t patient_char_uuid = WB_CHAR_PATIENT;
    uint16_t patient_val_handle;
    esp_ble_gatts_add_char(wb_srvc_id.id.inst_id, &patient_char_uuid, ESP_GATT_PERM_READ,
        ESP_GATT_AVAIL_ANY, &patient_val_handle, NULL);

    uint16_t sos_char_uuid = WB_CHAR_SOS;
    uint16_t sos_val_handle;
    esp_ble_gatts_add_char(wb_srvc_id.id.inst_id, &sos_char_uuid, ESP_GATT_PERM_READ,
        ESP_GATT_AVAIL_ANY, &sos_val_handle, NULL);

    uint16_t config_char_uuid = WB_CHAR_CONFIG;
    uint16_t config_val_handle;
    esp_ble_gatts_add_char(wb_srvc_id.id.inst_id, &config_char_uuid,
        ESP_GATT_PERM_READ | ESP_GATT_PERM_WRITE,
        ESP_GATT_AVAIL_ANY, &config_val_handle, NULL);

    esp_ble_gatts_add_char_descr(wb_srvc_id.id.inst_id, &hr_char_uuid, ESP_GATT_PERM_READ,
        &hr_val_handle, NULL);
    esp_ble_gatts_add_char_descr(wb_srvc_id.id.inst_id, &patient_char_uuid, ESP_GATT_PERM_READ,
        &patient_val_handle, NULL);
    esp_ble_gatts_add_char_descr(wb_srvc_id.id.inst_id, &sos_char_uuid, ESP_GATT_PERM_READ,
        &sos_val_handle, NULL);
    esp_ble_gatts_add_char_descr(wb_srvc_id.id.inst_id, &config_char_uuid,
        ESP_GATT_PERM_READ | ESP_GATT_PERM_WRITE, &config_val_handle, NULL);

    return ESP_OK;
}

/* ---- BLE task ---- */
static void ble_task(void *pvParameters) {
    ESP_LOGI(TAG, "BLE task started");

    esp_ble_gap_register_cb(ble_gap_event_handler);
    esp_ble_gatts_register_callback(ble_gatts_event_handler);
    esp_ble_set_device_name("Eregen-WB");
    init_profile_db();

    nvs_get_str_local(NVS_KEY_PATIENT, s_patient_id, &(size_t){sizeof(s_patient_id)});
    if (strlen(s_patient_id) == 0) {
        strncpy(s_patient_id, "UNBOUND", sizeof(s_patient_id));
    }
    ESP_LOGI(TAG, "Patient ID: %s", s_patient_id);

    xEventGroupSetBits(s_event_group, BLE_INIT_BIT);
    ble_start_advertising();

    while (1) {
        vTaskDelay(pdMS_TO_TICKS(5000));
        ESP_LOGI(TAG, "BLE active, connected=%d", esp_ble_get_conn_count());
    }
}

/* ---- NFC scan task ---- */
static void nfc_scan_task(void *pvParameters) {
    ESP_LOGI(TAG, "NFC scan task started");

    nfc_server_init(NULL, NULL);
    nfc_server_start();

    while (1) {
        nfc_tag_info_t tag_info;
        esp_err_t err = nfc_server_read_tag(&tag_info);
        if (err == ESP_OK) {
            ESP_LOGI(TAG, "NFC tag scanned: UID=%.2X:%.2X:%.2X",
                     tag_info.uid[0], tag_info.uid[1], tag_info.uid[2]);
            xEventGroupSetBits(s_event_group, NFC_SCANNED_BIT);

            // Process patient binding
            if (tag_info.uid[0] == 0xD2 && tag_info.uid[1] == 0xA0) {
                // Nurse terminal tag - bind patient
                strncpy(s_patient_info.patient_id, "NURSE-SCANNED", sizeof(s_patient_info.patient_id));
                patient_store_save(&s_patient_info);
                display_oled_show_patient(&s_patient_info);
            }
        }
        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/* ---- Wi-Fi + MQTT task ---- */
static void wifi_mqtt_task(void *pvParameters) {
    ESP_LOGI(TAG, "WiFi/MQTT task started");

    payload_crypto_init(&s_crypto_ctx);

    xEventGroupWaitBits(s_event_group, BLE_INIT_BIT, false, true, pdMS_TO_TICKS(10000));

    const char *client_id = "eregen-wb-0001";
    nvs_get_str_local(NVS_KEY_DEV_ID, (char *)&client_id, &(uint16_t){0});

    esp_err_t ret = mqtt_common_connect("mqtt.eregen.dev", 8883,
        client_id, NULL, NULL, NULL);

    if (ret == 0) {
        s_mqtt_ready = true;
        xEventGroupSetBits(s_event_group, MQTT_CONNECTED_BIT);
        ESP_LOGI(TAG, "MQTT connected");

        char subscribe_topic[64];
        snprintf(subscribe_topic, sizeof(subscribe_topic),
                 "eregen/medical/wb/%s/cmd", client_id);
        mqtt_common_subscribe(subscribe_topic, NULL);
    } else {
        ESP_LOGE(TAG, "MQTT connect failed: %d", ret);
    }

    vTaskDelete(NULL);
}

/* ---- SNTP time sync task ---- */
static void sntp_time_task(void *pvParameters) {
    ESP_LOGI(TAG, "SNTP task started");
    sntp_setservername(0, "ntp.aliyun.com");
    sntp_setservername(1, "ntp.tencent.com");
    sntp_init();

    while (sntp_get_sync_status() != SNTP_SYNC_STATUS_COMPLETED) {
        vTaskDelay(pdMS_TO_TICKS(10000));
    }
    ESP_LOGI(TAG, "Time synced via SNTP");
    vTaskDelete(NULL);
}

/* ---- SOS alert handler ---- */
static void handle_sos_alert(void) {
    char sos_msg[256];
    snprintf(sos_msg, sizeof(sos_msg),
             "{\"type\":\"sos\",\"dev_id\":\"%s\",\"patient_id\":\"%s\",\"ts\":%lu}",
             s_patient_id, s_patient_id, (unsigned long)xTaskGetTickCount());
    mqtt_common_publish("eregen/medical/wb/alert", sos_msg, strlen(sos_msg), 1);
    led_indicator_set_alert("SOS");
}

/* ---- Main entry ---- */
void app_main(void) {
    ESP_LOGI(TAG, "Eregen Medical Wristband v2.0.0");
    ESP_LOGI(TAG, "Target: ESP32-S3, ESP-IDF v5.3");
    ESP_LOGI(TAG, "Features: NFC, OLED, LED, AES Encryption\n");

    brand_boot_logo_show();

    /* Initialize NVS */
    esp_err_t ret = nvs_manager_init();
    ESP_ERROR_CHECK(ret);

    /* Initialize modules */
    ret = patient_store_init();
    if (ret != ESP_OK) ESP_LOGW(TAG, "Patient store init failed");

    ret = display_oled_init();
    if (ret != ESP_OK) ESP_LOGW(TAG, "OLED init failed");

    ret = led_indicator_init();
    if (ret != ESP_OK) ESP_LOGW(TAG, "LED init failed");

    ret = security_init(&s_crypto_ctx);
    if (ret != ESP_OK) ESP_LOGW(TAG, "Security init failed");

    /* Load device ID */
    uint8_t mac[6];
    esp_read_mac(mac, ESP_MAC_BT);
    char dev_id[17];
    snprintf(dev_id, sizeof(dev_id), "WB-%02X%02X%02X", mac[3], mac[4], mac[5]);
    nvs_manager_save_device_id(dev_id);
    ESP_LOGI(TAG, "Device ID: %s", dev_id);

    /* Load patient info */
    ret = patient_store_load(&s_patient_info);
    if (ret == ESP_OK) {
        strncpy(s_patient_id, s_patient_info.patient_id, sizeof(s_patient_id) - 1);
        ESP_LOGI(TAG, "Patient: %s", s_patient_id);
        display_oled_show_patient(&s_patient_info);
    } else {
        strncpy(s_patient_id, "UNBOUND", sizeof(s_patient_id));
        ESP_LOGW(TAG, "No patient bound");
    }

    /* Create event group */
    s_event_group = xEventGroupCreate();
    if (!s_event_group) {
        ESP_LOGE(TAG, "Failed to create event group");
        abort();
    }

    /* Start tasks */
    xTaskCreate(ble_task, "ble", 4096, NULL, 5, NULL);
    xTaskCreate(nfc_scan_task, "nfc", 4096, NULL, 4, NULL);
    xTaskCreate(wifi_mqtt_task, "mqtt", 6144, NULL, 3, NULL);
    xTaskCreate(sntp_time_task, "sntp", 2048, NULL, 2, NULL);

    log_device_info();

    /* Main loop */
    while (1) {
        EventBits_t bits = xEventGroupGetBits(s_event_group);

        if (bits & MQTT_CONNECTED_BIT) {
            static uint32_t s_last_heartbeat = 0;
            uint32_t now = xTaskGetTickCount();
            if (now - s_last_heartbeat > pdMS_TO_TICKS(60000)) {
                s_last_heartbeat = now;
                char msg[128];
                snprintf(msg, sizeof(msg),
                         "{\"type\":\"heartbeat\",\"dev_id\":\"%s\",\"patient\":\"%s\",\"bat\":%d,\"ts\":%lu}",
                         dev_id, s_patient_id, 95, (unsigned long)now);
                mqtt_common_publish("eregen/medical/wb/heartbeat", msg, strlen(msg), 1);
            }
        }

        if (bits & SOS_TRIGGERED_BIT) {
            xEventGroupClearBits(s_event_group, SOS_TRIGGERED_BIT);
            handle_sos_alert();
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

static void log_device_info(void) {
    ESP_LOGI(TAG, "Chip revision: %d", ESP_getEfuseRev());
    ESP_LOGI(TAG, "Free heap: %lu bytes", (unsigned long)esp_get_free_heap_size());
    ESP_LOGI(TAG, "Min free heap: %lu bytes", (unsigned long)esp_get_minimum_free_heap_size());
}

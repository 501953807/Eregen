/*
 * Eregen (颐贞) - Auto Pillbox Main Entry Point
 * ESP32-C3 based automatic medication dispensing with motor control,
 * optical detection, TTS voice broadcast, and inventory tracking.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_sleep.h"
#include "nvs_flash.h"
#include "esp_log.h"

#include "state_machine.h"
#include "dispensing.h"
#include "schedule_engine.h"
#include "med_rule_parser.h"
#include "opto_sensor.h"
#include "empty_detector.h"
#include "nvs_store.h"
#include "ble_pair.h"
#include "ap_config_mode.h"
#include "ota_handler.h"
#include "wifi_mqtt_bridge.h"

/* MQTT topic for device commands */
#define TOPIC_CMD_FMT       "eregen/device/pillbox/+/cmd"

static const char* TAG = "pillbox_auto";

/**
 * Global MQTT message handler — dispatches incoming messages by type.
 */
static void mqtt_msg_handler(const char* topic, const uint8_t* payload, uint16_t len) {
    cJSON* root = cJSON_ParseWithLength((const char*)payload, len);
    if (!root) return;

    cJSON* type = cJSON_GetObjectItem(root, "type");
    if (!type || !cJSON_IsString(type)) {
        cJSON_Delete(root);
        return;
    }

    if (strcmp(type->valuestring, "ota") == 0) {
        ota_handle_command(topic, payload, len);
    } else if (strcmp(type->valuestring, "med_rule") == 0) {
        /* Parse med_rule and apply to scheduler */
        const cJSON* rules_arr = cJSON_GetObjectItem(root, "rules");
        if (cJSON_IsArray(rules_arr)) {
            int count = cJSON_GetArraySize(rules_arr);
            if (count > 0 && count <= SCHEDULER_MAX_RULES) {
                reminder_rule_t rules[SCHEDULER_MAX_RULES];
                memset(rules, 0, sizeof(rules));
                for (int i = 0; i < count && i < SCHEDULER_MAX_RULES; i++) {
                    const cJSON* rule = cJSON_GetArrayItem(rules_arr, i);
                    if (!cJSON_IsObject(rule)) continue;
                    const cJSON* time_str = cJSON_GetObjectItem(rule, "time");
                    if (cJSON_IsString(time_str)) {
                        int h = 0, m = 0;
                        if (sscanf(time_str->valuestring, "%d:%d", &h, &m) == 2) {
                            rules[i].time.hour = (uint8_t)h;
                            rules[i].time.minute = (uint8_t)m;
                        }
                    }
                    const cJSON* dose = cJSON_GetObjectItem(rule, "dose");
                    rules[i].dose_count = cJSON_IsNumber(dose) ? (uint8_t)dose->valueint : 1;
                    const cJSON* comp = cJSON_GetObjectItem(rule, "compartment");
                    rules[i].compartment_index = cJSON_IsNumber(comp) ? (uint8_t)comp->valueint : i;
                }
                schedule_engine_replace_rules(rules, (uint8_t)count);
            }
        }
    } else if (strcmp(type->valuestring, "tts") == 0) {
        const cJSON* text = cJSON_GetObjectItem(root, "text");
        if (cJSON_IsString(text)) {
            tts_speak(text->valuestring);
        }
    }

    cJSON_Delete(root);
}

void app_main(void) {
    ESP_LOGI(TAG, "Eregen Auto Pillbox starting (v%s)", APP_VERSION);

    /* Initialize NVS for WiFi/stored config */
    esp_err_t ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        nvs_flash_erase();
        nvs_flash_init();
    }

    /* Initialize hardware peripherals */
    opto_sensor_init();
    empty_detector_init();
    state_machine_init();

    /* Initialize NVS store for inventory/rules */
    nvs_store_init();

    /* Initialize OTA handler */
    ota_init();

    /* Connect to Wi-Fi or enter AP pairing mode */
    if (!ap_config_connect_wifi()) {
        ESP_LOGW(TAG, "Wi-Fi not configured, entering AP pairing mode");
        ble_pair_start();
        ap_config_mode_start();
    } else {
        ESP_LOGI(TAG, "Wi-Fi connected");
        ble_pair_stop();
    }

    /* Initialize scheduling engine */
    schedule_engine_init();

    /* Connect to MQTT broker and register global message handler */
    if (wifi_is_connected() || ap_config_is_wifi_configured()) {
        bridge_init();
        bridge_connect();

        /* Wait briefly for MQTT connection */
        vTaskDelay(pdMS_TO_TICKS(2000));

        if (bridge_is_connected()) {
            mqtt_on_message(mqtt_msg_handler);
            bridge_subscribe(TOPIC_CMD_FMT);
            ESP_LOGI(TAG, "MQTT connected and subscribed to commands");
        }
    }

    /* Welcome message via TTS */
    ESP_LOGI(TAG, "Auto pillbox ready");

    /* Main loop — periodic state machine tick */
    while (1) {
        pillbox_state_t state = state_machine_run();

        /* Log state transitions */
        static pillbox_state_t s_prev = STATE_BOOT;
        if (state != s_prev) {
            ESP_LOGI(TAG, "State: %d -> %d", s_prev, state);
            s_prev = state;
        }

        /* Check OTA status during idle ticks */
        if (ota_is_active()) {
            ESP_LOGI(TAG, "OTA in progress: status=%s", ota_get_status());
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

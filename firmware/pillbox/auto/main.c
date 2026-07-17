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

static const char* TAG = "pillbox_auto";
#define APP_VERSION "auto_v1.0.0"
#define DEVICE_ID   "PX-AUTO-XXXX"

void app_main(void) {
    ESP_LOGI(TAG, "Eregen Smart Pillbox Auto starting (v%s)", APP_VERSION);

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

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

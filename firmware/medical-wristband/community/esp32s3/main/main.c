/*
 * Eregen Community Wristband - ESP32-S3 Firmware
 * Community elderly wristband with BLE GATT, OLED display, and WiFi MQTT
 */

#include <stdio.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_system.h"
#include "esp_log.h"
#include "nvs_flash.h"

/* Common bridge (WiFi + MQTT) */
#include "wifi_mqtt_bridge.h"

/* OLED (SSD1306 I2C) */
#include "oled.h"

static const char *TAG = "eregen_community";

/* Community-specific BLE GATT service UUID */
#define COMMUNITY_GATT_SERVICE_UUID  0xFFF0
#define COMMUNITY_GATT_CHAR_SIGNIN   0xFFF1
#define COMMUNITY_GATT_CHAR_WELFARE  0xFFF2
#define COMMUNITY_GATT_CHAR_STATUS   0xFFF3

/* Elder ID stored in NVS */
#define NVS_NAMESPACE "community_wb"

void app_main(void)
{
    ESP_LOGI(TAG, "Eregen Community Wristband starting...");

    // Initialize NVS for persistent storage
    esp_err_t ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        ESP_ERROR_CHECK(nvs_flash_erase());
        ret = nvs_flash_init();
    }
    ESP_ERROR_CHECK(ret);

    // Initialize OLED display
    oled_init();
    oled_print("Eregen", 0);
    oled_print("Community WB", 1);

    // Initialize WiFi + MQTT bridge
    ret = bridge_init();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Bridge init failed: %d", ret);
        oled_print("Bridge ERR", 1);
        vTaskDelay(pdMS_TO_TICKS(5000));
    }
    bridge_connect();

    // TODO: Initialize community BLE GATT server
    // TODO: Load elder_id from NVS
    // TODO: Start periodic signin beacon broadcast
    // TODO: Start main event loop

    while (1) {
        // Blink OLED to indicate alive
        oled_clear_line(2);
        oled_print("Running...", 2);
        vTaskDelay(pdMS_TO_TICKS(2000));
    }
}

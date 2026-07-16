/*
 * Eregen (颐贞) - Pillbox Basic Firmware Entry Point
 * Target: ESP32-C3 (RISC-V) | SDK: ESP-IDF v5.3+
 *
 * Integrates state machine loop with WiFi connection and MQTT heartbeat.
 *
 * 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/event_groups.h"

#include "esp_system.h"
#include "esp_wifi.h"
#include "esp_event.h"
#include "esp_log.h"
#include "nvs_flash.h"

#include "driver/gpio.h"
#include "driver/ledc.h"
#include "driver/adc.h"

#include "wifi_station.h"
#include "led_gpio.h"
#include "battery_manage.h"
#include "button_input.h"
#include "motor_control.h"
#include "tts_playback.h"
#include "state_machine.h"

/* Task priorities */
#define MAIN_TASK_PRIORITY          (tskIDLE_PRIORITY + 2)

/* Task stack sizes (bytes) */
#define MAIN_TASK_STACK_SIZE        (4096)

/* Device ID prefix: PX-XXXX */
#define DEVICE_ID_PREFIX            "PX-"

/* NVS namespace */
#define NVS_NAMESPACE               "pillbox"

/* Heartbeat interval (seconds) */
#define HEARTBEAT_INTERVAL_S        30

/* Button scan interval (milliseconds) */
#define BUTTON_SCAN_INTERVAL_MS     20

/* Battery read interval (seconds) */
#define BATTERY_READ_INTERVAL_S     120

/* State machine tick interval (milliseconds) */
#define STATE_MACHINE_TICK_MS       100

/* Log tag */
static const char *TAG = "pillbox_main";

/* Event group for inter-task signaling */
#define WIFI_CONNECTED_BIT    BIT0
#define HEARTBEAT_BIT         BIT1

static EventGroupHandle_t s_main_events;

/**
 * Application entry point.
 * Initializes NVS, WiFi, peripherals, then creates the main task.
 */
void app_main(void)
{
    esp_err_t ret;

    /* Initialize NVS flash storage */
    ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES ||
        ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        nvs_flash_erase();
        ret = nvs_flash_init();
    }
    ESP_ERROR_CHECK(ret);

    /* Initialize WiFi */
    ret = wifi_init();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "WiFi init failed: %s", esp_err_to_name(ret));
        return;
    }

    /* Connect to configured WiFi network */
    ret = wifi_connect();
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "WiFi connect failed (%s), continuing without network",
                 esp_err_to_name(ret));
    }

    /* Initialize NVS device config */
    {
        nvs_handle_t handle;
        ret = nvs_open(NVS_NAMESPACE, NVS_READWRITE, &handle);
        if (ret == ESP_OK) {
            /* Generate device ID from MAC and store in NVS */
            uint8_t mac[6];
            esp_read_mac(mac, ESP_MAC_WIFI_STA);

            char device_id[16];
            snprintf(device_id, sizeof(device_id), "%02X%02X%02X%02X%02X%02X",
                     mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

            size_t len = strlen(device_id) + 1;
            nvs_set_str(handle, "device_id", device_id);
            nvs_commit(handle);
            nvs_close(handle);

            ESP_LOGI(TAG, "Device ID stored: %s", device_id);
        } else {
            ESP_LOGE(TAG, "NVS open failed: %s", esp_err_to_name(ret));
        }
    }

    /* Initialize LED indicator */
    ret = led_init();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "LED init failed: %s", esp_err_to_name(ret));
        return;
    }

    /* Initialize buttons */
    ret = buttons_init();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Buttons init failed: %s", esp_err_to_name(ret));
        return;
    }

    /* Initialize battery ADC monitoring */
    ret = battery_init();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Battery init failed: %s", esp_err_to_name(ret));
        return;
    }

    /* Initialize motor control */
    motor_control_init();

    /* Create event group */
    s_main_events = xEventGroupCreate();
    if (s_main_events == NULL) {
        ESP_LOGE(TAG, "Failed to create event group");
        return;
    }

    /* Create main application task */
    xTaskCreate(vMainTask, "Main", MAIN_TASK_STACK_SIZE, NULL,
                MAIN_TASK_PRIORITY, NULL);
}

/**
 * Main application task loop.
 * Handles state machine ticks, heartbeat generation, button polling,
 * battery monitoring, and LED status indication based on WiFi state.
 */
static void vMainTask(void *pvParameter)
{
    (void)pvParameter;

    char device_id[16];
    {
        nvs_handle_t handle;
        esp_err_t ret = nvs_open(NVS_NAMESPACE, NVS_READONLY, &handle);
        if (ret == ESP_OK) {
            size_t len = sizeof(device_id);
            nvs_get_str(handle, "device_id", device_id, &len);
            nvs_close(handle);
        } else {
            /* Fallback: generate from MAC */
            uint8_t mac[6];
            esp_read_mac(mac, ESP_MAC_WIFI_STA);
            snprintf(device_id, sizeof(device_id), "%02X%02X%02X%02X%02X%02X",
                     mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);
        }
    }

    ESP_LOGI(TAG, "Device ID: %s%s", DEVICE_ID_PREFIX, device_id);

    uint32_t heartbeat_counter = 0;
    uint32_t battery_counter = 0;
    uint32_t sm_tick_counter = 0;

    for (;;) {
        bool connected = wifi_is_connected();

        /* Update LED based on WiFi state */
        if (connected) {
            led_set_color(LED_COLOR_GREEN);   /* Solid green = connected */
        } else {
            led_blink(LED_COLOR_BLUE, LED_PATTERN_FAST_BLINK);  /* Blue blink = pairing */
        }

        /* Poll buttons */
        button_event_t btn_event = buttons_get_event();
        if (btn_event != BUTTON_NONE) {
            ESP_LOGI(TAG, "Button event: %d", (int)btn_event);

            switch (btn_event) {
            case BUTTON_SHORT_PRESS:
                /* Flash green briefly as feedback */
                led_set_color(LED_COLOR_GREEN);
                vTaskDelay(pdMS_TO_TICKS(200));
                break;

            case BUTTON_LONG_PRESS:
                /* Trigger heartbeat immediately */
                ESP_LOGI(TAG, "Long press: forcing heartbeat");
                break;

            case BUTTON_DOUBLE_PRESS:
                /* Reset WiFi connection attempt */
                ESP_LOGI(TAG, "Double press: resetting WiFi");
                esp_wifi_disconnect();
                esp_wifi_connect();
                break;

            default:
                break;
            }

            buttons_clear_event();
        }

        /* Generate heartbeat every HEARTBEAT_INTERVAL_S seconds */
        heartbeat_counter++;
        if (heartbeat_counter >= HEARTBEAT_INTERVAL_S * 1000 / BUTTON_SCAN_INTERVAL_MS) {
            heartbeat_counter = 0;

            /* Read battery level */
            float voltage = battery_read_voltage();
            float percent = battery_calculate_percent(voltage);
            int bat_int = (int)percent;

            /* Log battery status */
            ESP_LOGI(TAG, "Battery: %.2fV, %d%%", voltage, bat_int);

            /* Low battery warning */
            if (percent < (BATTERY_LOW_THRESHOLD * 100.0f)) {
                ESP_LOGW(TAG, "Low battery warning: %d%%", bat_int);
                led_blink(LED_COLOR_RED, LED_PATTERN_FAST_BLINK);
            }

            /* Format and log heartbeat message */
            char heartbeat_msg[128];
            snprintf(heartbeat_msg, sizeof(heartbeat_msg),
                     "{\"type\":\"heartbeat\",\"dev_id\":\"%s%s\",\"bat\":%d}",
                     DEVICE_ID_PREFIX, device_id, bat_int);
            ESP_LOGI(TAG, "Heartbeat: %s", heartbeat_msg);

            /* Set LED to solid green after heartbeat (if WiFi connected) */
            if (connected) {
                led_set_color(LED_COLOR_GREEN);
            }
        }

        /* Full battery read every BATTERY_READ_INTERVAL_S seconds */
        battery_counter++;
        if (battery_counter >= BATTERY_READ_INTERVAL_S * 1000 / BUTTON_SCAN_INTERVAL_MS) {
            battery_counter = 0;
            float voltage = battery_read_voltage();
            float percent = battery_calculate_percent(voltage);
            ESP_LOGI(TAG, "Battery check: %.2fV (%.1f%%)", voltage, percent);
        }

        /* State machine tick every STATE_MACHINE_TICK_MS */
        sm_tick_counter++;
        if (sm_tick_counter >= STATE_MACHINE_TICK_MS / BUTTON_SCAN_INTERVAL_MS) {
            sm_tick_counter = 0;
            pillbox_state_t new_state = state_machine_run();
            ESP_LOGD(TAG, "SM tick: state=%s",
                     (new_state == STATE_BOOT) ? "BOOT" :
                     (new_state == STATE_CONNECT) ? "CONNECT" :
                     (new_state == STATE_IDLE) ? "IDLE" :
                     (new_state == STATE_REMINDER) ? "REMINDER" :
                     (new_state == STATE_DISPENSING) ? "DISPENSING" :
                     (new_state == STATE_DETECT) ? "DETECT" :
                     (new_state == STATE_REPORT) ? "REPORT" : "ERROR");
        }

        /* Delay for button scan interval */
        vTaskDelay(pdMS_TO_TICKS(BUTTON_SCAN_INTERVAL_MS));
    }
}

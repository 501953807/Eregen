/*
 * Eregen (颐贞) - Pillbox Smart Firmware Entry Point
 * Smart tier: Basic features + TTS voice reminder + OLED display + APP linkage
 * Target: ESP32-C3 (RISC-V) | SDK: ESP-IDF v5.3+
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/event_groups.h"
#include "freertos/queue.h"

#include "esp_system.h"
#include "esp_wifi.h"
#include "esp_event.h"
#include "esp_log.h"
#include "nvs_flash.h"

#include "driver/gpio.h"
#include "driver/ledc.h"
#include "driver/adc.h"
#include "driver/uart.h"

#include "wifi_station.h"
#include "led_gpio.h"
#include "battery_manage.h"
#include "button_input.h"

#include "voice_reminder.h"
#include "reminder_scheduler.h"
#include "app_link.h"
#include "oled_status.h"
#include "volume_control.h"

/* Task priorities */
#define MAIN_TASK_PRIORITY          (tskIDLE_PRIORITY + 2)
#define REMINDER_TASK_PRIORITY      (tskIDLE_PRIORITY + 1)
#define OLED_TASK_PRIORITY          (tskIDLE_PRIORITY + 1)
#define APP_CMD_TASK_PRIORITY       (tskIDLE_PRIORITY + 1)

/* Task stack sizes (bytes) */
#define MAIN_TASK_STACK_SIZE        (4096)
#define REMINDER_TASK_STACK_SIZE    (2048)
#define OLED_TASK_STACK_SIZE        (2048)
#define APP_CMD_TASK_STACK_SIZE     (2048)

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

/* Reminder check interval (seconds) */
#define REMINDER_CHECK_INTERVAL_S   60

/* OLED update interval (seconds) */
#define OLED_UPDATE_INTERVAL_S      5

/* MQTT command topic subscription */
#define MQTT_CMD_TOPIC_FMT          "eregen/device/%s/cmd"

/* Log tag */
static const char *TAG = "pillbox_smart";

/* Event group for inter-task signaling */
#define WIFI_CONNECTED_BIT    BIT0
#define SOS_BIT               BIT1

static EventGroupHandle_t s_main_events;

/* Shared state protected by internal mutexes in each module */
static char s_device_id[20];

/**
 * Send medication status to cloud via MQTT publish placeholder.
 * In production this sends through the MQTT client.
 */
static void send_med_status(uint8_t compartment, bool taken)
{
    char msg[128];
    snprintf(msg, sizeof(msg),
             "{\"type\":\"med_status\",\"dev_id\":\"%s%s\","
             "\"compartment\":%d,\"taken\":%s,\"ts\":%lu}",
             DEVICE_ID_PREFIX, s_device_id, compartment,
             taken ? "true" : "false", (unsigned long)vTaskGetTickCount());
    ESP_LOGI(TAG, "Med status: %s", msg);
    /* TODO: mqtt_publish(MQTT_DATA_TOPIC, msg, strlen(msg)); */
}

/**
 * Reminder task — checks scheduler every minute, triggers TTS + LED.
 */
static void vReminderTask(void *pvParameter)
{
    (void)pvParameter;

    reminder_rule_t pending;
    uint32_t counter = 0;

    ESP_LOGI(TAG, "Reminder task started");

    for (;;) {
        vTaskDelay(pdMS_TO_TICKS(1000));
        counter++;

        /* Check every 60 seconds */
        if (counter % REMINDER_CHECK_INTERVAL_S != 0)
            continue;

        /* Load current volume from NVS in case it changed */
        uint8_t vol = volume_get();
        (void)vol;

        if (scheduler_check_pending(&pending)) {
            ESP_LOGI(TAG, "Reminder due: %02d:%02d, type=%d, comp=%d",
                     pending.time.hour, pending.time.minute,
                     pending.med_type, pending.compartment_index);

            /* Build Chinese reminder text based on medicine type */
            const char *med_name = "药";
            switch (pending.med_type) {
            case MED_TYPE_CAPSULE:  med_name = "胶囊"; break;
            case MED_TYPE_TABLET:   med_name = "片剂"; break;
            case MED_TYPE_SYRUP:    med_name = "糖浆"; break;
            case MED_TYPE_INJECTION: med_name = "注射"; break;
            default:                break;
            }

            char tts_text[64];
            snprintf(tts_text, sizeof(tts_text),
                     "爷爷，该吃%s了", med_name);
            tts_speak(tts_text);

            /* Flash red LED as visual alert */
            led_blink(LED_COLOR_RED, LED_PATTERN_SLOW_BLINK);

            /* Simulate compartment open (hardware trigger would go here) */
            send_med_status(pending.compartment_index, false);

            /* Resume solid green after reminder */
            vTaskDelay(pdMS_TO_TICKS(3000));
            led_set_color(LED_COLOR_GREEN);
        }
    }
}

/**
 * OLED display task — refreshes status screen periodically.
 */
static void vOledTask(void *pvParameter)
{
    (void)pvParameter;

    uint32_t counter = 0;
    uint8_t last_battery = 0;
    bool last_wifi = false;

    ESP_LOGI(TAG, "OLED display task started");

    for (;;) {
        vTaskDelay(pdMS_TO_TICKS(1000));
        counter++;

        /* Refresh every OLED_UPDATE_INTERVAL_S seconds */
        if (counter % OLED_UPDATE_INTERVAL_S != 0)
            continue;

        /* Read battery */
        float voltage = battery_read_voltage();
        float percent = battery_calculate_percent(voltage);
        uint8_t bat_int = (uint8_t)percent;

        /* Only redraw if something changed */
        bool wifi = wifi_is_connected();
        if (bat_int == last_battery && wifi == last_wifi)
            continue;
        last_battery = bat_int;
        last_wifi = wifi;

        /* Clear and redraw */
        oled_clear();
        oled_draw_status_bar(bat_int, wifi);

        /* Draw next reminder time */
        reminder_rule_t next;
        if (scheduler_check_pending(&next)) {
            oled_draw_next_reminder(next.time.hour, next.time.minute);
        } else {
            oled_draw_next_reminder(0, 0);
        }

        oled_refresh();
    }
}

/**
 * APP command task — listens for MQTT commands (placeholder).
 * In production this subscribes to MQTT and dispatches.
 */
static void vAppCmdTask(void *pvParameter)
{
    (void)pvParameter;

    ESP_LOGI(TAG, "APP command task started");

    for (;;) {
        /* TODO: MQTT receive loop
         * while (mqtt_message_available()) {
         *     mqtt_recv(&topic, &payload, &len);
         *     applink_parse_mqtt_message(topic, payload, len);
         * }
         */
        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/**
 * Application entry point.
 * Extends Basic firmware with Smart-tier features.
 */
void app_main(void)
{
    esp_err_t ret;

    /* ---- Basic initialization (same as Basic tier) ---- */

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
            uint8_t mac[6];
            esp_read_mac(mac, ESP_MAC_WIFI_STA);
            snprintf(s_device_id, sizeof(s_device_id),
                     "%02X%02X%02X%02X%02X%02X",
                     mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

            char device_id[16];
            snprintf(device_id, sizeof(device_id), "%02X%02X%02X%02X%02X%02X",
                     mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);
            nvs_set_str(handle, "device_id", device_id);
            nvs_commit(handle);
            nvs_close(handle);

            ESP_LOGI(TAG, "Device ID stored: %s", s_device_id);
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

    /* ---- Smart-tier additions ---- */

    /* Create event group */
    s_main_events = xEventGroupCreate();
    if (s_main_events == NULL) {
        ESP_LOGE(TAG, "Failed to create event group");
        return;
    }

    /* Initialize volume control (loads from NVS) */
    ret = volume_init();
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "Volume init failed, using default");
    }

    /* Initialize TTS voice reminder module */
    ret = tts_init(UART_NUM_1);
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "TTS init failed, voice reminders disabled");
    }

    /* Initialize medication reminder scheduler */
    ret = scheduler_init();
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "Scheduler init failed, no reminders");
    }

    /* Initialize APP linkage command parser */
    ret = applink_init();
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "APP linkage init failed");
    }

    /* Initialize OLED status display */
    ret = oled_init();
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "OLED init failed, display disabled");
    }

    /* Create Smart-tier tasks */
    xTaskCreate(vReminderTask, "Remind", REMINDER_TASK_STACK_SIZE,
                NULL, REMINDER_TASK_PRIORITY, NULL);

    xTaskCreate(vOledTask, "OLED", OLED_TASK_STACK_SIZE,
                NULL, OLED_TASK_PRIORITY, NULL);

    xTaskCreate(vAppCmdTask, "AppCmd", APP_CMD_TASK_STACK_SIZE,
                NULL, APP_CMD_TASK_PRIORITY, NULL);

    /* Create main application task */
    xTaskCreate(vMainTask, "Main", MAIN_TASK_STACK_SIZE, NULL,
                MAIN_TASK_PRIORITY, NULL);
}

/**
 * Main application task loop.
 * Extends Basic main task with Smart-tier button handling.
 */
static void vMainTask(void *pvParameter)
{
    (void)pvParameter;

    ESP_LOGI(TAG, "Device ID: %s%s", DEVICE_ID_PREFIX, s_device_id);

    uint32_t heartbeat_counter = 0;
    uint32_t battery_counter = 0;

    for (;;) {
        bool connected = wifi_is_connected();

        /* Update LED based on WiFi state */
        if (connected) {
            led_set_color(LED_COLOR_GREEN);
        } else {
            led_blink(LED_COLOR_BLUE, LED_PATTERN_FAST_BLINK);
        }

        /* Poll buttons */
        button_event_t btn_event = buttons_get_event();
        if (btn_event != BUTTON_NONE) {
            ESP_LOGI(TAG, "Button event: %d", (int)btn_event);

            switch (btn_event) {
            case BUTTON_SHORT_PRESS:
                led_set_color(LED_COLOR_GREEN);
                vTaskDelay(pdMS_TO_TICKS(200));
                break;

            case BUTTON_LONG_PRESS:
                /* Long press right button = volume adjust */
                if (tts_is_playing()) {
                    tts_stop();
                } else {
                    volume_handle_button(true);
                }
                break;

            case BUTTON_DOUBLE_PRESS:
                esp_wifi_disconnect();
                esp_wifi_connect();
                break;

            default:
                break;
            }

            buttons_clear_event();
        }

        /* Heartbeat every HEARTBEAT_INTERVAL_S seconds */
        heartbeat_counter++;
        if (heartbeat_counter >= HEARTBEAT_INTERVAL_S * 1000 / BUTTON_SCAN_INTERVAL_MS) {
            heartbeat_counter = 0;

            float voltage = battery_read_voltage();
            float percent = battery_calculate_percent(voltage);
            int bat_int = (int)percent;

            ESP_LOGI(TAG, "Battery: %.2fV, %d%%", voltage, bat_int);

            if (percent < (BATTERY_LOW_THRESHOLD * 100.0f)) {
                ESP_LOGW(TAG, "Low battery warning: %d%%", bat_int);
                led_blink(LED_COLOR_RED, LED_PATTERN_FAST_BLINK);
            }

            char heartbeat_msg[128];
            snprintf(heartbeat_msg, sizeof(heartbeat_msg),
                     "{\"type\":\"heartbeat\",\"dev_id\":\"%s%s\",\"bat\":%d}",
                     DEVICE_ID_PREFIX, s_device_id, bat_int);
            ESP_LOGI(TAG, "Heartbeat: %s", heartbeat_msg);

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

        vTaskDelay(pdMS_TO_TICKS(BUTTON_SCAN_INTERVAL_MS));
    }
}

/*
 * Eregen (颐贞) - Pillbox Basic Firmware Entry Point
 * Target: ESP32-C3 (RISC-V)
 * SDK: ESP-IDF v5.3+
 *
 * © 2026 Eregen (颐贞). All rights reserved.
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

/* Task priorities */
#define MAIN_TASK_PRIORITY          (tskIDLE_PRIORITY + 2)
#define WIFI_CONNECT_PRIORITY       (tskIDLE_PRIORITY + 1)

/* Task stack sizes (bytes) */
#define MAIN_TASK_STACK_SIZE        (4096)
#define WIFI_CONNECT_TASK_STACK_SIZE (2048)

/* Device ID prefix: PX-XXXX */
#define DEVICE_ID_PREFIX            "PX-"

/* GPIO pin definitions */
#define PIN_LED_BLUE                GPIO_NUM_8
#define PIN_BUTTON_ENTER            GPIO_NUM_0
#define PIN_BUTTON_RIGHT            GPIO_NUM_9
#define PIN_BATTERY_ADC             ADC1_CHANNEL_0

/* NVS namespace */
#define NVS_NAMESPACE               "pillbox"

/* Log tag */
static const char *TAG = "pillbox_main";

/* Event group for inter-task signaling */
static EventGroupHandle_t s_main_events;

/* Forward declarations */
static void app_wifi_init(void);
static void app_nvs_init(void);
static void app_led_init(void);
static void app_button_init(void);
static void app_battery_monitor_init(void);
static void vMainTask(void *pvParameter);
static esp_err_t app_get_device_id(char *buf, size_t len);

/**
 * Application entry point.
 * Initializes NVS, WiFi, peripherals, then creates the main task.
 */
void app_main(void)
{
    esp_err_t ret;

    /* Initialize NVS flash storage */
    ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        nvs_flash_erase();
        ret = nvs_flash_init();
    }
    ESP_ERROR_CHECK(ret);

    /* Initialize WiFi */
    app_wifi_init();

    /* Initialize non-volatile storage for device config */
    app_nvs_init();

    /* Initialize LED indicator */
    app_led_init();

    /* Initialize buttons */
    app_button_init();

    /* Initialize battery ADC monitoring */
    app_battery_monitor_init();

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
 * WiFi initialization
 * Sets ESP32-C3 as WiFi station mode.
 */
static void app_wifi_init(void)
{
    tcpip_adapter_init();

    wifi_init_config_t cfg = WIFI_INIT_CONFIG_DEFAULT();
    esp_wifi_init(&cfg);

    esp_event_handler_instance_t instance_any_id;
    esp_event_handler_instance_t instance_got_ip;
    esp_event_handler_register(WIFI_EVENT, ESP_EVENT_ANY_ID,
                               &instance_any_id);
    esp_event_handler_register(IP_EVENT, IP_EVENT_STA_GOT_IP,
                               &instance_got_ip);

    wifi_config_t wifi_config = {
        .sta = {
            .threshold.authmode = WIFI_AUTH_WPA2_PSK,
        },
    };

    esp_wifi_set_mode(WIFI_MODE_STA);
    esp_wifi_set_config(WIFI_IF_STA, &wifi_config);
    esp_wifi_start();

    ESP_LOGI(TAG, "WiFi initialized (station mode)");
}

/**
 * NVS initialization
 * Loads or initializes device configuration from non-volatile storage.
 */
static void app_nvs_init(void)
{
    nvs_handle_t handle;
    esp_err_t ret = nvs_open(NVS_NAMESPACE, NVS_READWRITE, &handle);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "NVS open failed: %s", esp_err_to_name(ret));
        return;
    }

    /* TODO: Load device config from NVS */
    /* TODO: Generate and store unique device ID if not present */

    nvs_close(handle);
}

/**
 * LED initialization
 * Configures blue LED on GPIO8 for status indication.
 */
static void app_led_init(void)
{
    ledc_timer_config_t led_timer = {
        .duty_resolution = LEDC_TIMER_8_BIT,
        .freq_hz = 5000,
        .speed_mode = LEDC_LOW_SPEED_MODE,
        .timer_num = LEDC_TIMER_0,
        .clk_cfg = LEDC_AUTO_CLK,
    };
    ledc_timer_config(&led_timer);

    ledc_channel_config_t led_channel = {
        .speed_mode = LEDC_LOW_SPEED_MODE,
        .channel = LEDC_CHANNEL_0,
        .timer_sel = LEDC_TIMER_0,
        .intr_type = LEDC_INTR_DISABLE,
        .gpio_num = PIN_LED_BLUE,
        .duty = 0,
        .hpoint = 0,
    };
    ledc_channel_config(&led_channel);

    ESP_LOGI(TAG, "Blue LED initialized on GPIO%d", PIN_LED_BLUE);
}

/**
 * Button initialization
 * Configures Enter and Right buttons with internal pull-up.
 */
static void app_button_init(void)
{
    gpio_config_t btn_cfg = {
        .pin_bit_mask = (1ULL << PIN_BUTTON_ENTER) | (1ULL << PIN_BUTTON_RIGHT),
        .mode = GPIO_MODE_INPUT,
        .pull_up_en = GPIO_PULLUP_ENABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type = GPIO_INTR_DISABLE,
    };
    gpio_config(&btn_cfg);

    ESP_LOGI(TAG, "Buttons initialized: Enter=GPIO%d, Right=GPIO%d",
             PIN_BUTTON_ENTER, PIN_BUTTON_RIGHT);
}

/**
 * Battery monitoring initialization
 * Configures ADC1 channel for battery voltage reading.
 */
static void app_battery_monitor_init(void)
{
    adc1_config_width(ADC_WIDTH_BIT_12);
    adc1_config_channel_atten(PIN_BATTERY_ADC, ADC_ATTEN_DB_11);

    ESP_LOGI(TAG, "Battery ADC initialized on channel %d", PIN_BATTERY_ADC);
}

/**
 * Main application task loop
 * Handles pillbox logic: timing, reminders, compartment tracking,
 * WiFi/MQTT communication, and LED status indication.
 */
static void vMainTask(void *pvParameter)
{
    (void)pvParameter;

    char device_id[16];
    if (app_get_device_id(device_id, sizeof(device_id)) == ESP_OK) {
        ESP_LOGI(TAG, "Device ID: %s", device_id);
    }

    for (;;) {
        /* TODO: Check medication schedule */
        /* TODO: Trigger voice/LED reminder at scheduled times */
        /* TODO: Monitor compartment sensors for pill removal */
        /* TODO: Track taken/not-taken medication status */
        /* TODO: Read battery level and report if low */
        /* TODO: Connect to MQTT broker and sync data */

        /* Blink blue LED to indicate running */
        ledc_set_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0, 255);
        ledc_update_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0);
        vTaskDelay(pdMS_TO_TICKS(500));
        ledc_set_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0, 0);
        ledc_update_duty(LEDC_LOW_SPEED_MODE, LEDC_CHANNEL_0);
        vTaskDelay(pdMS_TO_TICKS(500));
    }
}

/**
 * Generate or retrieve unique device ID
 * Format: PX-XXXXXXXX (8 hex digits from ESP32 MAC address)
 */
static esp_err_t app_get_device_id(char *buf, size_t len)
{
    uint8_t mac[6];
    esp_read_mac(mac, ESP_MAC_WIFI_STA);

    int ret = snprintf(buf, len, "%02X%02X%02X%02X%02X%02X",
                       mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);
    if (ret < 0 || (size_t)ret >= len) {
        return ESP_ERR_NO_MEM;
    }
    return ESP_OK;
}

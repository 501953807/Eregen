/*
 * Eregen (颐贞) - WiFi Station Module Implementation
 * ESP32-C3 WiFi STA mode connection with auto-reconnect
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "wifi_station.h"

#include <string.h>

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/event_groups.h"

#include "esp_system.h"
#include "esp_wifi.h"
#include "esp_event.h"
#include "esp_log.h"
#include "nvs_flash.h"
#include "esp_netif.h"

#define WIFI_SSID_KEY         "wifi_ssid"
#define WIFI_PASS_KEY         "wifi_pass"
#define WIFI_CONNECT_TIMEOUT  30000   /* 30 seconds max wait for connection */

static const char *TAG = "wifi_station";

/* Event group bits */
#define WIFI_CONNECTED_BIT    BIT0
#define WIFI_FAIL_BIT         BIT1

static EventGroupHandle_t s_wifi_event_group;
static int s_retry_count = 0;
static int s_rssi = 0;

static void wifi_event_handler(void *arg, esp_event_base_t event_base,
                               int32_t event_id, void *event_data);

/**
 * Initialize WiFi in station mode.
 * Creates event group and registers handlers.
 */
esp_err_t wifi_init(void)
{
    s_wifi_event_group = xEventGroupCreate();
    if (s_wifi_event_group == NULL) {
        ESP_LOGE(TAG, "Failed to create event group");
        return ESP_ERR_NO_MEM;
    }

    /* Initialize default TCP/IP stack */
    esp_netif_init();

    /* Create default event loop */
    esp_event_loop_create_default();

    /* Initialize WiFi with default config */
    wifi_init_config_t cfg = WIFI_INIT_CONFIG_DEFAULT();
    esp_err_t ret = esp_wifi_init(&cfg);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "WiFi init failed: %s", esp_err_to_name(ret));
        return ret;
    }

    /* Register event handlers */
    esp_event_handler_instance_t instance_any_id;
    esp_event_handler_instance_t instance_got_ip;

    ret = esp_event_handler_instance_register(WIFI_EVENT,
                                               ESP_EVENT_ANY_ID,
                                               &wifi_event_handler,
                                               NULL, &instance_any_id);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Failed to register WIFI_EVENT handler: %s",
                 esp_err_to_name(ret));
        return ret;
    }

    ret = esp_event_handler_instance_register(IP_EVENT,
                                               IP_EVENT_STA_GOT_IP,
                                               &wifi_event_handler,
                                               NULL, &instance_got_ip);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Failed to register IP_EVENT handler: %s",
                 esp_err_to_name(ret));
        return ret;
    }

    esp_wifi_set_mode(WIFI_MODE_STA);
    esp_wifi_start();

    ESP_LOGI(TAG, "WiFi station initialized");
    return ESP_OK;
}

/**
 * Connect to the configured WiFi network.
 */
esp_err_t wifi_connect(void)
{
    nvs_handle_t handle;
    esp_err_t ret;

    /* Load SSID and password from NVS */
    char ssid[32] = "Eregen-Guest";
    char password[64] = "eregen123456";

    ret = nvs_open("pillbox", NVS_READONLY, &handle);
    if (ret == ESP_OK) {
        size_t len = sizeof(ssid);
        nvs_get_str(handle, WIFI_SSID_KEY, ssid, &len);

        len = sizeof(password);
        nvs_get_str(handle, WIFI_PASS_KEY, password, &len);

        nvs_close(handle);
    } else {
        nvs_close(&handle);
        ESP_LOGW(TAG, "NVS read failed (%s), using defaults",
                 esp_err_to_name(ret));
    }

    wifi_config_t wifi_config = {
        .sta = {
            .ssid = {0},
            .password = {0},
            .threshold.authmode = WIFI_AUTH_WPA2_PSK,
        },
    };

    strlen(ssid) > 0 ? memcpy(wifi_config.sta.ssid, ssid, 32) : NULL;
    strlen(password) > 0 ? memcpy(wifi_config.sta.password, password, 64) : NULL;

    ESP_LOGI(TAG, "Connecting to SSID: %s", ssid);

    ret = esp_wifi_set_config(WIFI_IF_STA, &wifi_config);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "Failed to set WiFi config: %s", esp_err_to_name(ret));
        return ret;
    }

    ret = esp_wifi_connect();
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "WiFi connect failed: %s", esp_err_to_name(ret));
        return ret;
    }

    /* Wait for connection result */
    EventBits_t bits = xEventGroupWaitBits(s_wifi_event_group,
                                            WIFI_CONNECTED_BIT | WIFI_FAIL_BIT,
                                            pdFALSE,
                                            pdFALSE,
                                            pdMS_TO_TICKS(WIFI_CONNECT_TIMEOUT));

    if (bits & WIFI_CONNECTED_BIT) {
        ESP_LOGI(TAG, "Connected to AP. RSSI: %d dBm", s_rssi);
        s_retry_count = 0;
        return ESP_OK;
    }

    ESP_LOGE(TAG, "Failed to connect within timeout");
    return ESP_ERR_TIMEOUT;
}

/**
 * Check if WiFi is currently connected.
 */
bool wifi_is_connected(void)
{
    wifi_sta_info_t info;
    esp_err_t ret = esp_wifi_sta_get_ap_info(&info);
    return (ret == ESP_OK);
}

/**
 * Get the RSSI of the current connection.
 */
int wifi_get_rssi(void)
{
    return s_rssi;
}

/**
 * Internal WiFi event handler.
 * Handles connection state changes and auto-reconnect.
 */
static void wifi_event_handler(void *arg, esp_event_base_t event_base,
                               int32_t event_id, void *event_data)
{
    if (event_base == WIFI_EVENT) {
        switch (event_id) {
        case WIFI_EVENT_STA_START:
            ESP_LOGI(TAG, "WiFi station started");
            break;

        case WIFI_EVENT_STA_DISCONNECTED: {
            ESP_LOGW(TAG, "Disconnected from AP");

            /* Reset connection event */
            xEventGroupClearBits(s_wifi_event_group,
                                 WIFI_CONNECTED_BIT | WIFI_FAIL_BIT);

            /* Auto-reconnect with retry limit */
            if (s_retry_count < 10) {
                esp_err_t ret = esp_wifi_connect();
                if (ret == ESP_OK) {
                    s_retry_count++;
                    ESP_LOGI(TAG, "Reconnecting... (%d/10)", s_retry_count);
                }
            } else {
                xEventGroupSetBits(s_wifi_event_group, WIFI_FAIL_BIT);
                ESP_LOGE(TAG, "Max reconnect attempts reached");
            }
            break;
        }

        default:
            break;
        }
    } else if (event_base == IP_EVENT && event_id == IP_EVENT_STA_GOT_IP) {
        ip_event_got_ip_t *event = (ip_event_got_ip_t *)event_data;
        ESP_LOGI(TAG, "Got IP: " IPSTR, IP2STR(&event->ip_info.ip));

        s_retry_count = 0;
        xEventGroupSetBits(s_wifi_event_group, WIFI_CONNECTED_BIT);
    }

    /* Update RSSI on any WiFi event */
    wifi_sta_info_t sta_info;
    if (esp_wifi_sta_get_ap_info(&sta_info) == ESP_OK) {
        s_rssi = sta_info.rssi;
    }
}

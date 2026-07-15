/*
 * Eregen (颐贞) - WiFi + MQTT Bridge Implementation
 * Unified WiFi connection and MQTT communication layer on ESP32-C3.
 * Uses ESP-IDF WiFi + Eclipse Paho MQTT client.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "wifi_mqtt_bridge.h"

#include "esp_log.h"
#include "esp_wifi.h"
#include "esp_event.h"
#include "esp_netif.h"
#include "mqtt_client.h"

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include <string.h>

static const char *TAG = "bridge";

/* Internal state */
static mqtt_client_handle_t s_mqtt_client = NULL;
static bool s_wifi_connected    = false;
static bool s_mqtt_connected    = false;
static esp_mqtt_client_handle_t s_mqtt = NULL;

/* Device ID buffer for topic construction */
static char s_dev_id[12] = "";

static void event_handler(void *handler_args, esp_event_base_t base,
                          int32_t event_id, void *event_data);

/**
 * Initialize the WiFi+MQTT bridge.
 */
esp_err_t bridge_init(void)
{
    /* Register WiFi event handler */
    esp_event_handler_instance_t wifi_inst;
    esp_event_handler_instance_register(ESP_EVENT_ANY_ID,
                                        event_handler,
                                        NULL, &wifi_inst);

    /* Initialize TCP/IP stack */
    esp_netif_init();
    esp_event_loop_create_default();

    /* Initialize WiFi in station mode */
    esp_netif_t *sta_netif = esp_netif_new_default_from_dt(
        ESP_NETIF_IF_STA_WIFI);
    if (!sta_netif) {
        return ESP_ERR_NO_MEM;
    }

    wifi_init_config_t cfg = WIFI_INIT_CONFIG_DEFAULT();
    esp_err_t ret = esp_wifi_init(&cfg);
    if (ret != ESP_OK) {
        return ret;
    }

    ret = esp_wifi_set_mode(WIFI_MODE_STA);
    if (ret != ESP_OK) return ret;

    ret = esp_wifi_start();
    if (ret != ESP_OK) return ret;

    ESP_LOGI(TAG, "Bridge initialized");
    return ESP_OK;
}

/**
 * Connect to WiFi and MQTT broker.
 */
esp_err_t bridge_connect(void)
{
    wifi_config_t wifi_conf = {
        .sta = {
            .threshold.authmode = WIFI_AUTH_WPA2_PSK,
        },
    };

    /* Load credentials from NVS or use defaults */
    strncpy((char *)wifi_conf.sta.ssid, "eregen", sizeof(wifi_conf.sta.ssid));
    strncpy((char *)wifi_conf.sta.password, "eregen1234",
            sizeof(wifi_conf.sta.password));
    wifi_conf.sta.ssid[sizeof(wifi_conf.sta.ssid) - 1] = '\0';
    wifi_conf.sta.password[sizeof(wifi_conf.sta.password) - 1] = '\0';

    esp_err_t ret = esp_wifi_set_config(WIFI_IF_STA, &wifi_conf);
    if (ret != ESP_OK) return ret;

    ret = esp_wifi_connect();
    if (ret != ESP_OK) return ret;

    /* Wait for WiFi connection (handled by event handler) */
    /* Once connected, we set the flag in event_handler */

    ESP_LOGI(TAG, "Connecting to WiFi...");
    return ESP_OK;
}

/**
 * Disconnect from WiFi and MQTT.
 */
esp_err_t bridge_disconnect(void)
{
    if (s_mqtt) {
        esp_mqtt_stop(s_mqtt);
        esp_mqtt_client_destroy(s_mqtt);
        s_mqtt = NULL;
        s_mqtt_connected = false;
    }

    esp_wifi_disconnect();
    s_wifi_connected = false;

    ESP_LOGI(TAG, "Disconnected");
    return ESP_OK;
}

/**
 * Publish a message to an MQTT topic.
 */
esp_err_t bridge_publish(const char *topic, const uint8_t *payload,
                         size_t len)
{
    if (!s_mqtt || !s_mqtt_connected) {
        ESP_LOGW(TAG, "Not connected, cannot publish to %s", topic);
        return ESP_FAIL;
    }

    int msg_id = esp_mqtt_client_publish(s_mqtt, topic,
                                         (const char *)payload, len,
                                         MQTT_QOS, 0);
    if (msg_id < 0) {
        ESP_LOGE(TAG, "Publish failed: %d", msg_id);
        return ESP_FAIL;
    }

    ESP_LOGD(TAG, "Published to %s (msg_id=%d)", topic, msg_id);
    return ESP_OK;
}

/**
 * Subscribe to an MQTT topic.
 */
esp_err_t bridge_subscribe(const char *topic)
{
    if (!s_mqtt || !s_mqtt_connected) {
        ESP_LOGW(TAG, "Not connected, cannot subscribe to %s", topic);
        return ESP_FAIL;
    }

    int msg_id = esp_mqtt_client_subscribe(s_mqtt, topic, MQTT_QOS);
    if (msg_id < 0) {
        ESP_LOGE(TAG, "Subscribe failed");
        return ESP_FAIL;
    }

    ESP_LOGI(TAG, "Subscribed to %s (msg_id=%d)", topic, msg_id);
    return ESP_OK;
}

/**
 * Check if WiFi and MQTT are both connected.
 */
bool bridge_is_connected(void)
{
    return s_wifi_connected && s_mqtt_connected;
}

/* ---- Event handling ---- */

static void event_handler(void *handler_args, esp_event_base_t base,
                          int32_t event_id, void *event_data)
{
    (void)handler_args;

    if (base == WIFI_EVENT) {
        switch (event_id) {
        case WIFI_EVENT_STA_CONNECTED:
            s_wifi_connected = true;
            ESP_LOGI(TAG, "WiFi connected");
            break;

        case WIFI_EVENT_STA_DISCONNECTED:
            s_wifi_connected = false;
            ESP_LOGI(TAG, "WiFi disconnected, reconnecting...");
            esp_wifi_connect();
            break;

        default:
            break;
        }
    } else if (base == IP_EVENT) {
        switch (event_id) {
        case IP_EVENT_STA_GOT_IP:
            ESP_LOGI(TAG, "Got IP, initializing MQTT...");

            /* Initialize MQTT client */
            esp_mqtt_client_config_t mqtt_cfg = {
                .broker.address.hostname = MQTT_BROKER_HOST,
                .broker.address.port = MQTT_BROKER_PORT,
                .credentials.client_id = s_dev_id,
                .session.keepalive = MQTT_KEEPALIVE_S,
                .session.last_will.topic = TOPIC_BASE "/status",
            };
            strncpy(mqtt_cfg.session.last_will.data,
                    "{\"status\":\"offline\"}", 22);
            mqtt_cfg.session.last_will.qos = MQTT_QOS;
            mqtt_cfg.session.last_will.retain = 0;

            s_mqtt = esp_mqtt_client_init(&mqtt_cfg);
            if (!s_mqtt) {
                ESP_LOGE(TAG, "MQTT init failed");
                return;
            }

            esp_mqtt_client_start(s_mqtt);
            s_mqtt_connected = true;
            ESP_LOGI(TAG, "MQTT connected to %s:%d",
                     MQTT_BROKER_HOST, MQTT_BROKER_PORT);
            break;

        default:
            break;
        }
    }
}

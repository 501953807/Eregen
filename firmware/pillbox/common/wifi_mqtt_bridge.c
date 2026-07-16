/*
 * Eregen (颐贞) - WiFi + MQTT Bridge Implementation
 * Unified WiFi connection and MQTT communication layer on ESP32-C3.
 * Uses ESP-IDF WiFi + Eclipse Paho MQTT client.
 * Features: auto-reconnect with exponential backoff, heartbeat ping,
 *           topic subscription with callbacks, global message handler.
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 */

#include "wifi_mqtt_bridge.h"

#include "esp_log.h"
#include "esp_wifi.h"
#include "esp_event.h"
#include "esp_netif.h"
#include "mqtt_client.h"
#include "nvs_flash.h"

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/timers.h"

#include <string.h>

static const char *TAG = "bridge";

/* ---- Internal state ---- */

/* MQTT client handle */
static esp_mqtt_client_handle_t s_mqtt = NULL;

/* Connection flags */
static bool s_wifi_connected    = false;
static bool s_mqtt_connected    = false;

/* Device ID buffer for topic construction */
static char s_dev_id[DEVICE_ID_FULL_LEN] = "";

/* Global message handler (set by mqtt_on_message) */
static void (*s_global_msg_handler)(const char *topic, const uint8_t *payload, uint16_t len) = NULL;

/* Per-topic subscription list (bounded array) */
#define MAX_SUBSCRIPTIONS  8

typedef struct {
    char topic[64];
    void (*callback)(const char *payload, uint16_t len);
} subscription_entry_t;

static subscription_entry_t s_subscriptions[MAX_SUBSCRIPTIONS];
static int s_subscription_count = 0;

/* Reconnection state machine */
typedef enum {
    RECONNECT_IDLE,
    RECONNECT_WAITING,
    RECONNECT_DONE,
} reconnect_state_t;

static reconnect_state_t s_reconnect_state = RECONNECT_IDLE;
static int s_reconnect_attempts = 0;

/* Heartbeat timer handle */
static TimerHandle_t s_heartbeat_timer = NULL;

/* ---- Forward declarations ---- */

static void wifi_event_handler(void *handler_args, esp_event_base_t base,
                               int32_t event_id, void *event_data);
static void mqtt_event_handler(void *handler_args, esp_event_base_t base,
                               int32_t event_id, void *event_data);
static void mqtt_start_client(void);
static void mqtt_stop_client(void);
static void mqtt_reconnect_callback(TimerHandle_t xTimer);
static void heartbeat_callback(TimerHandle_t xTimer);
static void dispatch_message(const char *topic, const uint8_t *payload, uint16_t len);

/* ---- Public API ---- */

esp_err_t bridge_init(void)
{
    /* Initialize TCP/IP stack */
    esp_netif_init();
    esp_event_loop_create_default();

    /* Initialize WiFi in station mode */
    wifi_init_config_t cfg = WIFI_INIT_CONFIG_DEFAULT();
    esp_err_t ret = esp_wifi_init(&cfg);
    if (ret != ESP_OK) {
        return ret;
    }

    ret = esp_wifi_set_mode(WIFI_MODE_STA);
    if (ret != ESP_OK) return ret;

    ret = esp_wifi_start();
    if (ret != ESP_OK) return ret;

    /* Register WiFi event handler */
    esp_event_handler_instance_t wifi_inst;
    esp_event_handler_instance_register(ESP_EVENT_ANY_ID,
                                        wifi_event_handler,
                                        NULL, &wifi_inst);

    /* Create heartbeat timer (fires every MQTT_HEARTBEAT_INTERVAL_S seconds) */
    s_heartbeat_timer = xTimerCreate("mqtt_heartbeat",
                                     pdMS_TO_TICKS(MQTT_HEARTBEAT_INTERVAL_S * 1000),
                                     pdTRUE,              /* auto-reload */
                                     NULL,
                                     heartbeat_callback);

    ESP_LOGI(TAG, "Bridge initialized");
    return ESP_OK;
}

esp_err_t bridge_connect(void)
{
    /* Load WiFi credentials from NVS or use defaults */
    nvs_handle_t handle;
    char ssid[32] = "eregen";
    char password[64] = "eregen1234";

    esp_err_t ret = nvs_open("pillbox", NVS_READONLY, &handle);
    if (ret == ESP_OK) {
        size_t len = sizeof(ssid);
        nvs_get_str(handle, "wifi_ssid", ssid, &len);
        len = sizeof(password);
        nvs_get_str(handle, "wifi_pass", password, &len);
        nvs_close(handle);
    } else {
        nvs_close(&handle);
        ESP_LOGW(TAG, "NVS read failed, using default SSID");
    }

    wifi_config_t wifi_conf = {
        .sta = {
            .ssid = {0},
            .password = {0},
            .threshold.authmode = WIFI_AUTH_WPA2_PSK,
        },
    };

    memcpy(wifi_conf.sta.ssid, ssid, strlen(ssid));
    memcpy(wifi_conf.sta.password, password, strlen(password));

    ret = esp_wifi_set_config(WIFI_IF_STA, &wifi_conf);
    if (ret != ESP_OK) return ret;

    ret = esp_wifi_connect();
    if (ret != ESP_OK) return ret;

    ESP_LOGI(TAG, "Connecting to WiFi AP: %s", ssid);
    return ESP_OK;
}

esp_err_t bridge_disconnect(void)
{
    mqtt_stop_client();

    esp_wifi_disconnect();
    s_wifi_connected = false;
    s_mqtt_connected = false;

    ESP_LOGI(TAG, "Disconnected");
    return ESP_OK;
}

esp_err_t bridge_publish(const char *topic, const uint8_t *payload, size_t len)
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

bool bridge_is_connected(void)
{
    return s_wifi_connected && s_mqtt_connected;
}

bool mqtt_publish_topic(const char *topic, const uint8_t *data, uint16_t len)
{
    return bridge_publish(topic, data, len) == ESP_OK;
}

bool mqtt_subscribe_topic(const char *topic, void (*callback)(const char *, uint16_t))
{
    if (s_subscription_count >= MAX_SUBSCRIPTIONS) {
        ESP_LOGE(TAG, "Max subscriptions (%d) reached", MAX_SUBSCRIPTIONS);
        return false;
    }

    subscription_entry_t *entry = &s_subscriptions[s_subscription_count];
    strncpy(entry->topic, topic, sizeof(entry->topic) - 1);
    entry->topic[sizeof(entry->topic) - 1] = '\0';
    entry->callback = callback;
    s_subscription_count++;

    /* Also register with the MQTT broker if already connected */
    if (s_mqtt && s_mqtt_connected) {
        bridge_subscribe(topic);
    }

    ESP_LOGI(TAG, "Registered topic callback: %s", topic);
    return true;
}

void mqtt_on_message(void (*handler)(const char *topic, const uint8_t *payload, uint16_t len))
{
    s_global_msg_handler = handler;
}

/* ---- Internal helpers ---- */

/**
 * Start (or restart) the MQTT client and connect to broker.
 */
static void mqtt_start_client(void)
{
    if (s_mqtt) {
        esp_mqtt_client_destroy(s_mqtt);
        s_mqtt = NULL;
    }

    if (!s_wifi_connected) {
        ESP_LOGW(TAG, "Cannot start MQTT: WiFi not connected");
        return;
    }

    esp_mqtt_client_config_t mqtt_cfg = {
        .broker.address.hostname = MQTT_BROKER_HOST,
        .broker.address.port = MQTT_BROKER_PORT,
        .credentials.client_id = s_dev_id,
        .session.keepalive = MQTT_KEEPALIVE_S,
        .session.last_will.topic = TOPIC_BASE "/status",
    };

    /* Build device-specific last will */
    char will_payload[64];
    snprintf(will_payload, sizeof(will_payload),
             "{\"type\":\"status\",\"dev_id\":\"%s%s\",\"status\":\"offline\"}",
             DEVICE_ID_PREFIX, s_dev_id);
    strncpy((char *)mqtt_cfg.session.last_will.data, will_payload, 63);
    mqtt_cfg.session.last_will.data[63] = '\0';
    mqtt_cfg.session.last_will.qos = MQTT_QOS;
    mqtt_cfg.session.last_will.retain = 0;

    s_mqtt = esp_mqtt_client_init(&mqtt_cfg);
    if (!s_mqtt) {
        ESP_LOGE(TAG, "MQTT client init failed");
        return;
    }

    /* Register internal MQTT event handler */
    esp_mqtt_register_event_handler(s_mqtt, ESP_EVENT_ANY_ID,
                                    mqtt_event_handler, NULL);

    esp_err_t ret = esp_mqtt_client_start(s_mqtt);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "MQTT client start failed: %s", esp_err_to_name(ret));
        esp_mqtt_client_destroy(s_mqtt);
        s_mqtt = NULL;
        return;
    }

    ESP_LOGI(TAG, "MQTT client started, connecting to %s:%d",
             MQTT_BROKER_HOST, MQTT_BROKER_PORT);
}

/**
 * Stop and destroy the MQTT client.
 */
static void mqtt_stop_client(void)
{
    if (s_mqtt) {
        esp_mqtt_client_stop(s_mqtt);
        esp_mqtt_client_destroy(s_mqtt);
        s_mqtt = NULL;
    }
    s_mqtt_connected = false;
}

/**
 * Dispatch an incoming MQTT message to per-topic subscribers and the global handler.
 */
static void dispatch_message(const char *topic, const uint8_t *payload, uint16_t len)
{
    /* Call per-topic callbacks that match this topic */
    for (int i = 0; i < s_subscription_count; i++) {
        const subscription_entry_t *sub = &s_subscriptions[i];
        /* Simple prefix match: if subscribed topic is a prefix of received topic,
           or they are exactly equal, call the callback */
        if (strcmp(topic, sub->topic) == 0 ||
            strstr(topic, sub->topic) == topic) {
            if (sub->callback) {
                sub->callback((const char *)payload, len);
            }
        }
    }

    /* Call global message handler if registered */
    if (s_global_msg_handler) {
        s_global_msg_handler(topic, payload, len);
    }
}

/**
 * Attempt MQTT reconnection with exponential backoff.
 */
static void mqtt_reconnect_callback(TimerHandle_t xTimer)
{
    (void)xTimer;

    if (s_mqtt_connected) {
        s_reconnect_state = RECONNECT_IDLE;
        s_reconnect_attempts = 0;
        return;
    }

    if (!s_wifi_connected) {
        return;  /* Wait for WiFi before trying MQTT */
    }

    if (s_reconnect_state == RECONNECT_WAITING) {
        return;  /* Already waiting for MQTT connect */
    }

    if (s_reconnect_attempts >= MQTT_MAX_RECONNECT_ATTEMPTS) {
        ESP_LOGE(TAG, "Max MQTT reconnect attempts (%d) reached",
                 MQTT_MAX_RECONNECT_ATTEMPTS);
        s_reconnect_state = RECONNECT_IDLE;
        return;
    }

    s_reconnect_state = RECONNECT_WAITING;
    s_reconnect_attempts++;

    ESP_LOGI(TAG, "MQTT reconnect attempt %d/%d",
             s_reconnect_attempts, MQTT_MAX_RECONNECT_ATTEMPTS);

    mqtt_start_client();
}

/**
 * Heartbeat: send periodic status publish to keep connection alive.
 * The MQTT broker keepalive handles PINGREQ automatically, but we
 * also publish a retained status message so the cloud knows we're alive.
 */
static void heartbeat_callback(TimerHandle_t xTimer)
{
    (void)xTimer;

    if (!s_mqtt || !s_mqtt_connected) {
        return;
    }

    char status_msg[64];
    snprintf(status_msg, sizeof(status_msg),
             "{\"type\":\"status\",\"dev_id\":\"%s%s\",\"status\":\"online\"}",
             DEVICE_ID_PREFIX, s_dev_id);

    esp_err_t ret = esp_mqtt_client_publish(
        s_mqtt, TOPIC_BASE "/status", status_msg, strlen(status_msg),
        MQTT_QOS, 1);  /* retain = 1 so cloud always sees last status */

    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "MQTT heartbeat publish failed: %s", esp_err_to_name(ret));
    } else {
        ESP_LOGD(TAG, "MQTT heartbeat status published");
    }
}

/* ---- Event handlers ---- */

/**
 * WiFi event handler.
 */
static void wifi_event_handler(void *handler_args, esp_event_base_t base,
                               int32_t event_id, void *event_data)
{
    (void)handler_args;

    if (base == WIFI_EVENT) {
        switch (event_id) {
        case WIFI_EVENT_STA_CONNECTED:
            s_wifi_connected = true;
            s_reconnect_state = RECONNECT_IDLE;
            s_reconnect_attempts = 0;
            ESP_LOGI(TAG, "WiFi connected");
            break;

        case WIFI_EVENT_STA_DISCONNECTED:
            s_wifi_connected = false;
            s_mqtt_connected = false;
            ESP_LOGW(TAG, "WiFi disconnected, reconnecting...");
            esp_wifi_connect();
            break;

        default:
            break;
        }
    } else if (base == IP_EVENT) {
        switch (event_id) {
        case IP_EVENT_STA_GOT_IP:
            ESP_LOGI(TAG, "Got IP, starting MQTT client...");
            mqtt_start_client();

            /* If MQTT wasn't already connected, start reconnect timer */
            if (!s_mqtt_connected) {
                if (s_heartbeat_timer) {
                    xTimerStart(s_heartbeat_timer, 0);
                }
            }
            break;

        default:
            break;
        }
    }
}

/**
 * MQTT event handler.
 * Manages MQTT connection lifecycle, reconnection, and message dispatch.
 */
static void mqtt_event_handler(void *handler_args, esp_event_base_t base,
                               int32_t event_id, void *event_data)
{
    (void)handler_args;

    esp_mqtt_event_handle_t event = *event_data;
    esp_mqtt_client_handle_t client = event->client;

    switch (event_id) {
    case MQTT_EVENT_CONNECTED:
        ESP_LOGI(TAG, "MQTT connected to broker");
        s_mqtt_connected = true;
        s_reconnect_state = RECONNECT_IDLE;
        s_reconnect_attempts = 0;

        /* Resubscribe to all registered topics */
        for (int i = 0; i < s_subscription_count; i++) {
            esp_mqtt_client_subscribe(client, s_subscriptions[i].topic, MQTT_QOS);
        }

        /* Start heartbeat timer */
        if (s_heartbeat_timer) {
            xTimerStart(s_heartbeat_timer, 0);
        }
        break;

    case MQTT_EVENT_DISCONNECTED:
        ESP_LOGW(TAG, "MQTT disconnected from broker");
        s_mqtt_connected = false;
        break;

    case MQTT_EVENT_DATA: {
        ESP_LOGD(TAG, "MQTT message received on topic: %.*s",
                 event->topic_len, event->topic);
        dispatch_message(event->topic, event->data, event->data_len);
        break;
    }

    case MQTT_EVENT_ERROR:
        ESP_LOGE(TAG, "MQTT error occurred");
        s_mqtt_connected = false;
        break;

    default:
        break;
    }
}

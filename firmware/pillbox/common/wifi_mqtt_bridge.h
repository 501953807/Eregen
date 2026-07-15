/*
 * Eregen (颐贞) - WiFi + MQTT Bridge
 * Unified WiFi connection and MQTT communication layer on ESP32-C3.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef WIFI_MQTT_BRIDGE_H
#define WIFI_MQTT_BRIDGE_H

#include "esp_err.h"
#include <stdbool.h>
#include <stddef.h>

/* MQTT configuration */
#define MQTT_BROKER_HOST      "mqtt.eregen.local"
#define MQTT_BROKER_PORT      1883
#define MQTT_KEEPALIVE_S      60
#define MQTT_QOS              1

/* Topic format: eregen/device/pillbox/{dev_id}/up|down */
#define TOPIC_BASE            "eregen/device/pillbox"
#define TOPIC_UP_FMT          TOPIC_BASE "/%s/up"
#define TOPIC_DOWN_FMT        TOPIC_BASE "/%s/down"

/**
 * Initialize the WiFi+MQTT bridge.
 * Sets up WiFi station mode and MQTT client.
 * Must be called before bridge_connect().
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t bridge_init(void);

/**
 * Connect to WiFi and MQTT broker.
 * Performs WiFi STA connection, then MQTT connect with last will.
 * Auto-reconnect is enabled.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t bridge_connect(void);

/**
 * Disconnect from WiFi and MQTT.
 *
 * @return ESP_OK on success
 */
esp_err_t bridge_disconnect(void);

/**
 * Publish a message to an MQTT topic.
 *
 * @param topic   MQTT topic string
 * @param payload Message payload bytes
 * @param len     Payload length in bytes
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t bridge_publish(const char *topic, const uint8_t *payload,
                         size_t len);

/**
 * Subscribe to an MQTT topic.
 *
 * @param topic   MQTT topic string
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t bridge_subscribe(const char *topic);

/**
 * Check if WiFi and MQTT are both connected.
 *
 * @return true if connected, false otherwise
 */
bool bridge_is_connected(void);

#endif /* WIFI_MQTT_BRIDGE_H */

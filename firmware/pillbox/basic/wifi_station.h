/*
 * Eregen (颐贞) - WiFi Station Module
 * ESP32-C3 WiFi STA mode connection with auto-reconnect
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef WIFI_STATION_H
#define WIFI_STATION_H

#include "esp_err.h"

/**
 * Initialize WiFi in station mode.
 * Registers event handlers for connection lifecycle.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t wifi_init(void);

/**
 * Connect to the configured WiFi network.
 * SSID and password are loaded from NVS namespace "pillbox".
 * If not found in NVS, uses default credentials for development.
 *
 * @return ESP_OK on success
 */
esp_err_t wifi_connect(void);

/**
 * Check if WiFi is currently connected.
 *
 * @return true if connected, false otherwise
 */
bool wifi_is_connected(void);

/**
 * Get the RSSI (signal strength) of the current connection.
 *
 * @return RSSI in dBm (typically -30 to -90), or a negative error code
 */
int wifi_get_rssi(void);

#endif /* WIFI_STATION_H */

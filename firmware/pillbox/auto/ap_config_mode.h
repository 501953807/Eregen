/*
 * Eregen (颐贞) - AP Configuration Mode
 * When device boots without saved WiFi credentials, creates an access point
 * so the user can configure home WiFi via a captive portal browser page.
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef AP_CONFIG_MODE_H
#define AP_CONFIG_MODE_H

#include <stdbool.h>

/**
 * Start AP configuration mode.
 * Creates an ESP32-C3 soft-AP named "eregen-pixel-XXXX" where XXXX
 * is the last 4 hex digits of the device MAC address.
 * Starts a lightweight HTTP captive portal on 192.168.4.1.
 *
 * Call this from app_main() when no WiFi credentials are found in NVS.
 */
void ap_config_start(void);

/**
 * Stop AP configuration mode.
 * Tears down soft-AP, stops captive portal server, frees resources.
 */
void ap_config_stop(void);

/**
 * Check whether AP configuration mode is currently active.
 *
 * @return true if AP is broadcasting and captive portal is running
 */
bool ap_config_is_active(void);

/**
 * Retrieve the WiFi credentials that were saved during configuration.
 * Only valid after ap_config_save_credentials() has been called.
 *
 * @param ssid       Output buffer for AP SSID (min 33 bytes)
 * @param password   Output buffer for AP password (min 65 bytes)
 * @param max_len    Size of the larger of the two buffers
 * @return true if credentials are available, false otherwise
 */
bool ap_config_get_credentials(char *ssid, char *password, size_t max_len);

/**
 * Called by the captive portal handler when the user submits WiFi credentials.
 * Saves them to NVS namespace "pillbox" under keys "wifi_ssid" / "wifi_pass".
 *
 * @param ssid     Home network SSID (max 32 bytes)
 * @param password Home network password (max 64 bytes)
 * @return true if saved successfully
 */
bool ap_config_save_credentials(const char *ssid, const char *password);

#endif /* AP_CONFIG_MODE_H */

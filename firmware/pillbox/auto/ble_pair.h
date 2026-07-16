/*
 * Eregen (颐贞) - BLE Pairing for WiFi Credentials
 * Smart pillbox tier — receive home WiFi credentials from smartphone APP via BLE GATT.
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef BLE_PAIR_H
#define BLE_PAIR_H

#include <stdbool.h>
#include <stdint.h>

/**
 * Start BLE advertising and GATT server for credential pairing.
 * The device advertises "eregen-pair-XXXX" where XXXX is derived
 * from the MAC address. A smartphone app connects and writes
 * WiFi SSID/password into a custom GATT characteristic.
 *
 * Call this in app_main() when WiFi credentials are not found in NVS
 * AND the device supports BLE (smart / auto pillbox tiers).
 */
void ble_pair_start(void);

/**
 * Stop BLE advertising and free resources.
 * Call after credentials have been saved and WiFi connection is attempted.
 */
void ble_pair_stop(void);

/**
 * Check whether valid WiFi credentials have been received via BLE.
 *
 * @return true if credentials were set (and can be read via ble_pair_get_credentials)
 */
bool ble_pair_has_credentials(void);

/**
 * Retrieve the WiFi credentials received via BLE pairing.
 *
 * @param ssid       Output buffer for AP SSID (min 33 bytes)
 * @param password   Output buffer for AP password (min 65 bytes)
 * @param max_len    Size of the larger of the two buffers
 * @return true if credentials are available, false otherwise
 */
bool ble_pair_get_credentials(char *ssid, char *password, size_t max_len);

/**
 * Save received BLE credentials to NVS flash.
 * This should be called after confirming the credentials are valid.
 *
 * @return true if saved successfully
 */
bool ble_pair_save_to_nvs(void);

#endif /* BLE_PAIR_H */

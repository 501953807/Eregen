/*
 * Eregen (颐贞) - BLE Pairing Header (Plus Tier)
 * Handles BLE advertising for initial device configuration.
 * Family APP connects via BLE to provision WiFi credentials (SSID + password)
 * and cloud MQTT endpoint. Uses a 6-digit PIN for mutual authentication.
 *
 * MIT License
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

#ifndef BLE_PAIR_H
#define BLE_PAIR_H

#include <stdint.h>
#include <stdbool.h>

/* ---- BLE Constants ---- */

/** BLE service UUID for Eregen provisioning (128-bit). */
#define BLE_PROVISION_SERVICE_UUID \
    {0x71,0x3E,0x00,0x00,0x00,0x00,0x10,0x00, \
     0x80,0x00,0x00,0x80,0x5F,0x9B,0x34,0xFB}

/** Maximum length of WiFi SSID. */
#define BLE_BLE_MAX_SSID_LEN   32U

/** Maximum length of WiFi password. */
#define BLE_BLE_MAX_PASS_LEN   64U

/** PIN code length (6 digits). */
#define BLE_PIN_LENGTH         6U

/** PIN verification timeout (seconds). */
#define BLE_PAIR_TIMEOUT_S     120U

/** Max failed PIN attempts before lockout. */
#define BLE_MAX_PIN_ATTEMPTS   3U

/** Lockout duration after too many failed attempts (seconds). */
#define BLE_LOCKOUT_DURATION_S 300U

/* ---- State Machine ---- */

typedef enum {
    BLE_STATE_IDLE,           /* Not advertising, no connection */
    BLE_STATE_ADVERTISING,    /* Broadcasting provisioning service */
    BLE_STATE_CONNECTED,      /* BLE GATT connected */
    BLE_STATE_PIN_REQUESTED,  /* Waiting for user to enter PIN on phone */
    BLE_STATE_PIN_VERIFIED,   /* PIN matched — provisioning channel open */
    BLE_STATE_PROVISIONING,   /* Transmitting WiFi credentials */
    BLE_STATE_COMPLETE,       /* Provisioning done — stop advertising */
    BLE_STATE_LOCKED,         /* Too many failed PIN attempts */
} ble_pair_state_t;

/* ---- Data Structures ---- */

/**
 * WiFi credential payload received from family APP.
 */
typedef struct {
    char ssid[BLE_BLE_MAX_SSID_LEN + 1];
    char password[BLE_BLE_MAX_PASS_LEN + 1];
    uint8_t security;  /* 0=WPA2, 1=WPA3, 2=Open */
} wifi_credentials_t;

/**
 * Cloud endpoint configuration.
 */
typedef struct {
    char mqtt_host[64];
    uint16_t mqtt_port;
    char client_id[32];
} cloud_config_t;

/**
 * Full provisioning payload.
 */
typedef struct {
    wifi_credentials_t wifi;
    cloud_config_t     cloud;
} provisioning_data_t;

/**
 * Current BLE pairing state.
 */
typedef struct {
    ble_pair_state_t state;
    uint32_t         pin_code;          /* 6-digit PIN for verification */
    uint8_t          pin_attempts;      /* Failed attempts counter */
    uint32_t         lockout_until_s;   /* Epoch seconds until lockout expires */
    bool             provisioned;       /* True if WiFi credentials are stored */
    provisioning_data_t stored_data;    /* Last successful provisioning */
} ble_pair_state_t_data;

/* ---- API ---- */

/*
 * Initialize the BLE pairing subsystem.
 * Generates a random 6-digit PIN and begins advertising.
 * @return true on success.
 */
bool ble_pair_init(void);

/*
 * Start BLE advertising for provisioning.
 * Must be called after ble_pair_init().
 */
void ble_pair_start_advertising(void);

/*
 * Stop BLE advertising.
 * Called when provisioning completes or device is paired.
 */
void ble_pair_stop_advertising(void);

/*
 * Check the current pairing state.
 * @param[out] out Output state (must not be NULL).
 * @return true if state retrieved successfully.
 */
bool ble_pair_get_state(ble_pair_state_t_data *out);

/*
 * Generate a new random 6-digit PIN.
 * Called on init and when a new pairing session starts.
 * @return The generated PIN value.
 */
uint32_t ble_pair_generate_pin(void);

/*
 * Verify a PIN entered by the user on the family APP.
 * @param pin The 6-digit PIN to verify.
 * @return true if the PIN matches.
 */
bool ble_pair_verify_pin(uint32_t pin);

/*
 * Receive and store WiFi + cloud credentials from the family APP.
 * Writes to NVS for persistence across reboots.
 * @param data Pointer to provisioning data (must not be NULL).
 * @return true on success.
 */
bool ble_pair_receive_credentials(const provisioning_data_t *data);

/*
 * Check whether the device has been provisioned (has valid WiFi credentials).
 * @return true if provisioned.
 */
bool ble_pair_is_provisioned(void);

/*
 * Get the stored WiFi credentials.
 * @param[out] out Output buffer (must not be NULL).
 * @return true if credentials are available.
 */
bool ble_pair_get_stored_credentials(provisioning_data_t *out);

/*
 * Clear all stored provisioning data (factory reset).
 * @return true on success.
 */
bool ble_pair_factory_reset(void);

/*
 * Periodic tick function — call every second from RTOS task.
 * Handles: advertising timer, PIN timeout, lockout countdown.
 */
void ble_pair_tick(void);

/* ---- Test helpers ---- */

#ifdef TEST_MODE
/* Inject a known PIN for deterministic testing. */
void ble_pair_set_mock_pin(uint32_t pin);

/* Inject a known credential set for testing receive path. */
void ble_pair_set_mock_credentials(const provisioning_data_t *data);
#endif

#endif /* BLE_PAIR_H */

/*
 * Eregen (颐贞) - BLE Pairing Implementation (Plus Tier)
 * BLE advertising + GATT provisioning service for initial device setup.
 * Family APP discovers the device via BLE, verifies a 6-digit PIN,
 * then receives WiFi credentials and cloud endpoint configuration.
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

#include "ble_pair.h"
#include <stdlib.h>
#include <string.h>

/* ---- Internal state ---- */

static ble_pair_state_t_data s_state;
static bool s_initialized = false;

/* Mock values for testing. */
#ifdef TEST_MODE
static uint32_t s_mock_pin = 0;
static bool s_mock_has_credentials = false;
static provisioning_data_t s_mock_credentials;
#endif

/* ---- PIN generation ---- */

uint32_t ble_pair_generate_pin(void)
{
    /* Generate random 6-digit PIN: [100000, 999999]. */
    uint32_t pin = (uint32_t)(rand() % 900000U) + 100000U;
    s_state.pin_code = pin;
    s_state.pin_attempts = 0;
    return pin;
}

/* ---- Public API ---- */

bool ble_pair_init(void)
{
    s_initialized = true;
    memset(&s_state, 0, sizeof(s_state));
    s_state.state = BLE_STATE_IDLE;
    s_state.provisioned = false;

    srand((unsigned int)0x12345678UL);  /* Seed from device UID in production */
    s_state.pin_code = ble_pair_generate_pin();

    return true;
}

void ble_pair_start_advertising(void)
{
    if (!s_initialized) {
        return;
    }

    s_state.state = BLE_STATE_ADVERTISING;

    /*
     * In production, this calls the BLE stack:
     *   ble_gap_adv_start(...);
     *   ble_gatts_service_register(BLE_PROVISION_SERVICE_UUID);
     *
     * For the GD32 platform with a BLE stack (e.g., NORDIC nRF52-compatible
     * or built-in BLE peripheral), the exact API varies. Here we record
     * the state change and would emit log output.
     */
}

void ble_pair_stop_advertising(void)
{
    s_state.state = BLE_STATE_IDLE;

    /*
     * Production: ble_gap_adv_stop();
     */
}

bool ble_pair_get_state(ble_pair_state_t_data *out)
{
    if (!out) {
        return false;
    }
    *out = s_state;
    return true;
}

bool ble_pair_verify_pin(uint32_t pin)
{
    if (!s_initialized) {
        return false;
    }

    /* Check lockout status. */
    if (s_state.state == BLE_STATE_LOCKED) {
        return false;
    }

    s_state.pin_attempts++;

    if (pin == s_state.pin_code) {
        s_state.pin_attempts = 0;
        s_state.state = BLE_STATE_PIN_VERIFIED;
        return true;
    }

    /* Wrong PIN — track attempts. */
    if (s_state.pin_attempts >= BLE_MAX_PIN_ATTEMPTS) {
        s_state.state = BLE_STATE_LOCKED;
        s_state.lockout_until_s = 300000000U;  /* Arbitrary future epoch */
        return false;
    }

    return false;
}

bool ble_pair_receive_credentials(const provisioning_data_t *data)
{
    if (!data || !s_initialized) {
        return false;
    }

    /* Only accept credentials if PIN has been verified. */
    if (s_state.state != BLE_STATE_PIN_VERIFIED) {
        return false;
    }

    /* Validate inputs. */
    if (strlen(data->wifi.ssid) == 0 || strlen(data->wifi.ssid) > BLE_BLE_MAX_SSID_LEN) {
        return false;
    }
    if (strlen(data->wifi.password) > BLE_BLE_MAX_PASS_LEN) {
        return false;
    }

    /* Store credentials. */
    s_state.stored_data = *data;
    s_state.provisioned = true;
    s_state.state = BLE_STATE_COMPLETE;

    /*
     * In production, persist to NVS:
     *   nvs_write(NVS_KEY_WIFI, &data->wifi, sizeof(data->wifi));
     *   nvs_write(NVS_KEY_CLOUD, &data->cloud, sizeof(data->cloud));
     */

    return true;
}

bool ble_pair_is_provisioned(void)
{
#ifdef TEST_MODE
    return s_mock_has_credentials;
#else
    return s_state.provisioned;
#endif
}

bool ble_pair_get_stored_credentials(provisioning_data_t *out)
{
    if (!out) {
        return false;
    }

#ifdef TEST_MODE
    if (s_mock_has_credentials) {
        *out = s_mock_credentials;
        return true;
    }
    return false;
#else
    if (!s_state.provisioned) {
        return false;
    }
    *out = s_state.stored_data;
    return true;
#endif
}

bool ble_pair_factory_reset(void)
{
    s_state.provisioned = false;
    s_state.state = BLE_STATE_IDLE;
    s_state.pin_attempts = 0;
    memset(&s_state.stored_data, 0, sizeof(s_state.stored_data));

    /*
     * In production, clear NVS:
     *   nvs_erase_key(NVS_KEY_WIFI);
     *   nvs_erase_key(NVS_KEY_CLOUD);
     */

    return true;
}

void ble_pair_tick(void)
{
    if (!s_initialized) {
        return;
    }

    /*
     * In production, this drives the BLE stack timer:
     *   - Check connection status
     *   - Restart advertising if timeout
     *   - Decrement lockout countdown
     */

    if (s_state.state == BLE_STATE_LOCKED) {
        /* Lockout is time-based; in production compare against RTC. */
    }
}

/* ---- Test helpers ---- */

#ifdef TEST_MODE

void ble_pair_set_mock_pin(uint32_t pin)
{
    s_mock_pin = pin;
    s_state.pin_code = pin;
}

void ble_pair_set_mock_credentials(const provisioning_data_t *data)
{
    if (data) {
        s_mock_has_credentials = true;
        s_mock_credentials = *data;
    } else {
        s_mock_has_credentials = false;
    }
}

#endif

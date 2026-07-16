/*
 * Eregen (颐贞) - Photoelectric Sensor Implementation
 * ITR2800E beam-break detection for automatic pillbox.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "opto_sensor.h"

#include <stdio.h>
#include <string.h>

#ifdef TEST_MODE
#else
#include "driver/gpio.h"
#include "esp_log.h"
#endif

static const char *TAG = "opto";

/* Internal state */
static bool s_beam_broken     = false;
static bool s_initialized     = false;

#ifdef TEST_MODE
/* Mock: allow tests to directly set the simulated state */
void opto_sensor_set_mock_state(bool broken)
{
    s_beam_broken = broken;
}

bool opto_sensor_get_mock_state(void)
{
    return s_beam_broken;
}
#endif

/**
 * Initialize the photoelectric sensor GPIO.
 */
bool opto_sensor_init(void)
{
#ifdef TEST_MODE
    s_beam_broken = false;
    s_initialized = true;

    printf("[opto] Initialized (mock mode)\n");
    return true;
#else
    /* GPIO_NUM_4 is used for the ITR2800E sensor */
    gpio_config_t io_conf = {
        .pin_bit_mask = (1ULL << GPIO_NUM_4),
        .mode         = GPIO_MODE_INPUT,
        .pull_up_en   = GPIO_PULLUP_ENABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type    = GPIO_INTR_ANYEDGE,  /* Both edges trigger interrupt */
    };

    esp_err_t ret = gpio_config(&io_conf);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "GPIO config failed: %s", esp_err_to_name(ret));
        return false;
    }

    /* Install ISR handler */
    ret = gpio_install_isr_service(0);
    if (ret != ESP_OK && ret != ESP_ERR_INVALID_STATE) {
        ESP_LOGE(TAG, "ISR install failed: %s", esp_err_to_name(ret));
        return false;
    }

    s_initialized = true;
    s_beam_broken = false;

    ESP_LOGI(TAG, "Opto sensor initialized: GPIO4 (active-low, debounce=%dms)",
             OPTO_DEBOUNCE_MS);
    return true;
#endif
}

/**
 * Read the current sensor state.
 * Returns true if medication has been REMOVED (beam broken).
 */
bool opto_sensor_read(void)
{
    if (!s_initialized) {
        return false;
    }

#ifdef TEST_MODE
    return s_beam_broken;
#else
    bool level = gpio_get_level(GPIO_NUM_4);
    return !level;  /* Active-low: high level = beam broken */
#endif
}

/**
 * Get the last known state.
 */
bool opto_sensor_get_last_state(void)
{
    return s_beam_broken;
}

/**
 * Reset the internal state to simulate "medication present".
 */
void opto_sensor_reset(void)
{
#ifdef TEST_MODE
    /* In test mode, do NOT overwrite mock state set by opto_sensor_set_mock_state().
     * Only reset initialization flag. */
    s_initialized = true;
#else
    s_beam_broken = false;
    s_initialized = true;

    ESP_LOGI(TAG, "Sensor reset: beam assumed intact");
#endif
}

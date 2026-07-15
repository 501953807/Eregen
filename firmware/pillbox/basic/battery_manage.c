/*
 * Eregen (颐贞) - Battery Management Implementation
 * 18650 Li-ion battery voltage measurement via ADC1
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "battery_manage.h"

#include "esp_log.h"
#include "driver/adc.h"
#include "driver/adc_deprecated.h"
#include "esp_adc_cal.h"

static const char *TAG = "battery_manage";

/* ADC pin mapping for ESP32-C3 */
#define BATTERY_ADC_PIN   GPIO_NUM_4

/* Sampling parameters */
#define SAMPLE_COUNT      8
#define SAMPLE_INTERVAL   10   /* ms between samples */

static bool s_initialized = false;

/**
 * Initialize battery monitoring (ADC configuration).
 */
esp_err_t battery_init(void)
{
    /* Configure ADC width (12-bit on ESP32-C3) */
    adc1_config_width(ADC_WIDTH_BIT_12);

    /* Configure attenuation for 0-3.3V range (we use divider for 4.2V max) */
    adc1_config_channel_atten(BATTERY_ADC_PIN, ADC_ATTEN_DB_11);

    s_initialized = true;
    ESP_LOGI(TAG, "Battery ADC initialized on GPIO%d", BATTERY_ADC_PIN);

    return ESP_OK;
}

/**
 * Read the current battery voltage.
 * Takes multiple samples and averages them for stability.
 * Applies voltage divider correction.
 */
float battery_read_voltage(void)
{
    if (!s_initialized) {
        ESP_LOGW(TAG, "Battery not initialized");
        return -1.0f;
    }

    /* Take multiple samples and average */
    uint32_t sum = 0;
    for (int i = 0; i < SAMPLE_COUNT; i++) {
        int raw = adc1_get_reading(BATTERY_ADC_PIN, ADC_WAIT_MAX);
        sum += (uint32_t)raw;
        vTaskDelay(pdMS_TO_TICKS(SAMPLE_INTERVAL));
    }

    uint32_t avg = sum / SAMPLE_COUNT;

    /* Convert raw ADC value to voltage at ADC input */
    float adc_voltage = (avg / BATTERY_ADC_MAX_VALUE) * BATTERY_ADC_REF_VOLTAGE;

    /* Apply voltage divider correction to get actual battery voltage */
    float battery_voltage = adc_voltage / BATTERY_DIVIDER_RATIO;

    /* Clamp to valid range */
    if (battery_voltage > BATTERY_VOLTAGE_FULL) {
        battery_voltage = BATTERY_VOLTAGE_FULL;
    } else if (battery_voltage < BATTERY_VOLTAGE_EMPTY) {
        battery_voltage = BATTERY_VOLTAGE_EMPTY;
    }

    return battery_voltage;
}

/**
 * Calculate battery charge percentage from voltage.
 * Uses linear interpolation between empty and full thresholds.
 */
float battery_calculate_percent(float voltage)
{
    /* Clamp voltage to known range */
    if (voltage >= BATTERY_VOLTAGE_FULL) {
        return 100.0f;
    }
    if (voltage <= BATTERY_VOLTAGE_EMPTY) {
        return 0.0f;
    }

    /* Linear interpolation */
    float percent = ((voltage - BATTERY_VOLTAGE_EMPTY) /
                     (BATTERY_VOLTAGE_FULL - BATTERY_VOLTAGE_EMPTY)) * 100.0f;

    return percent;
}

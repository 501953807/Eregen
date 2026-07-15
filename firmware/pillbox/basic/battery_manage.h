/*
 * Eregen (颐贞) - Battery Management Module
 * 18650 Li-ion battery voltage measurement via ADC
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef BATTERY_MANAGE_H
#define BATTERY_MANAGE_H

#include "esp_err.h"

/* Voltage thresholds for 18650 Li-ion */
#define BATTERY_VOLTAGE_FULL    4.2f   /* 100% */
#define BATTERY_VOLTAGE_MEDIUM  3.7f   /* 50% */
#define BATTERY_VOLTAGE_EMPTY   3.0f   /* 0% */
#define BATTERY_LOW_THRESHOLD   0.20f  /* 20% low-battery warning */

/* ADC voltage divider ratio: R1/(R1+R2) = 1.1V / 5.0V ≈ 0.22 */
/* Actual hardware: 100k + 360k divider => ratio ≈ 0.217 */
#define BATTERY_DIVIDER_RATIO   0.217f

/* ADC reference voltage (ESP32-C3 max input is 1.1V with 11dB attenuation) */
#define BATTERY_ADC_REF_VOLTAGE 1.1f

/* ADC resolution (12-bit) */
#define BATTERY_ADC_MAX_VALUE   4095.0f

/**
 * Initialize battery monitoring (ADC configuration).
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t battery_init(void);

/**
 * Read the current battery voltage.
 * Applies voltage divider correction to return actual cell voltage.
 *
 * @return Battery voltage in volts (3.0-4.2 range), or -1.0f on error
 */
float battery_read_voltage(void);

/**
 * Calculate battery charge percentage from voltage.
 * Linear interpolation between EMPTY and FULL thresholds.
 *
 * @param voltage Current battery voltage in volts
 * @return Charge percentage (0.0-100.0), clamped to valid range
 */
float battery_calculate_percent(float voltage);

#endif /* BATTERY_MANAGE_H */

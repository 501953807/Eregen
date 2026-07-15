/*
 * Eregen (颐贞) - Battery ADC Measurement Header
 * LiPo battery voltage measurement via ADC with voltage divider
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef BATTERY_ADC_H
#define BATTERY_ADC_H

#include <stdint.h>
#include <stdbool.h>

/* Battery voltage range (LiPo) */
#define BATT_VOLTAGE_EMPTY     3.0f   /* Volts, 0% capacity */
#define BATT_VOLTAGE_FULL      4.2f   /* Volts, 100% capacity */

/* Voltage divider: 2×100k resistors (2:1 division) */
#define BATT_DIVIDER_RATIO     2.0f
#define BATT_ADC_REF_VOLTAGE   3.3f   /* ADC reference voltage */
#define BATT_ADC_RESOLUTION    4096U  /* 12-bit ADC */

/* ADC sampling configuration */
#define BATT_SAMPLE_COUNT      8U     /* Number of samples to average */
#define BATT_SAMPLE_DELAY_US   10U    /* Delay between samples in microseconds */

/* Minimum measurable voltage after divider */
#define BATT_MIN_MEASURABLE    (BATT_VOLTAGE_EMPTY / BATT_DIVIDER_RATIO)
#define BATT_MAX_MEASURABLE    (BATT_VOLTAGE_FULL / BATT_DIVIDER_RATIO)

/* Battery status */
typedef struct {
    float voltage_mv;     /* Measured voltage in millivolts */
    uint8_t percent;      /* Battery percentage 0-100 */
    bool charging;        /* True if charging detected */
    bool critical;        /* True if voltage below empty threshold */
} battery_status_t;

/*
 * Initialize the battery ADC peripheral.
 * Configures GPIO and ADC for voltage measurement.
 * @return true on success.
 */
bool battery_init(void);

/*
 * Read the raw battery voltage through the voltage divider.
 * Takes multiple samples and averages for stability.
 * @return Voltage in millivolts (mV), or 0 on error.
 */
uint16_t battery_read_voltage_mv(void);

/*
 * Convert raw ADC reading to voltage in millivolts.
 * @param adc_value Raw 12-bit ADC value
 * @return Voltage in mV
 */
uint16_t battery_adc_to_mv(uint16_t adc_value);

/*
 * Calculate battery percentage from voltage.
 * Uses linear interpolation between empty and full thresholds.
 * @param voltage_mv Voltage in millivolts
 * @return Percentage 0-100
 */
uint8_t battery_calculate_percent(uint16_t voltage_mv);

/*
 * Get complete battery status.
 * @return battery_status_t with voltage, percent, charging state, critical flag.
 */
battery_status_t battery_get_status(void);

#endif /* BATTERY_ADC_H */

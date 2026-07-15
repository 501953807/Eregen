/*
 * Eregen (颐贞) - Battery ADC Measurement Implementation
 * LiPo battery voltage measurement via ADC with voltage divider
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "battery_adc.h"
#include "gd32e230_adc.h"
#include "gd32e230_rcu.h"
#include "gd32e230_gpio.h"

/*
 * Initialize the battery ADC peripheral.
 * Configures PA5 as analog input for voltage measurement.
 */
bool battery_init(void)
{
    rcu_periph_clock_enable(RCU_GPIOA);
    rcu_periph_clock_enable(RCU_ADC0);

    /* Configure PA5 as analog input */
    gpio_init(BATTERY_ADC_GPIO_PORT, GPIO_MODE_AIN,
              GPIO_OSPEED_50MHZ, BATTERY_ADC_GPIO_PIN);

    /* ADC clock configuration: PCLK2 / 8 = 18MHz (max 14MHz recommended) */
    adc_clock_config(ADC_FPCK_DIV_8);

    /* Deinit ADC */
    adc_deinit(ADC0);

    /* ADC mode: independent mode */
    adc_mode_config(ADC_MODE_INDEPENDENT);

    /* Continuous conversion mode */
    adc_special_function_config(ADC0, ADC_CONTINUOUS_MODE, ENABLE);

    /* Trigger: software trigger */
    adc_software_trigger_config(ADC0, ADC_REGULAR_TRIGGER);

    /* Regular channel configuration: 16-bit resolution */
    adc_resolution_config(ADC0, ADC_RESOLUTION_12BIT);
    adc_data_alignment_config(ADC0, ADC_DATAALIGN_RIGHT);

    /* Enable ADC */
    adc_enable(ADC0);

    /* Wait for ADC stabilization */
    for (volatile int32_t i = 0; i < 1000000; i++);

    /* Calibrate ADC */
    adc_calibration_enable(ADC0);

    return true;
}

/*
 * Read a single ADC sample.
 */
static uint16_t adc_read_single(uint8_t channel)
{
    /* Set regular group sequence and sampling time */
    adc_regular_channel_config(ADC0, 0U, channel, ADC_SAMPLETIME_239_5);

    /* Start software conversion */
    adc_software_trigger_enable(ADC0, ADC_REGULAR_CHANNEL);

    /* Wait for conversion to complete */
    while (adc_flag_get(ADC0, ADC_FLAG_EOC) == RESET) {
        /* Wait */
    }

    /* Clear EOC flag */
    adc_flag_clear(ADC0, ADC_FLAG_EOC);

    return adc_regular_data_read(ADC0);
}

/*
 * Read the raw battery voltage through the voltage divider.
 * Takes multiple samples and averages for stability.
 */
uint16_t battery_read_voltage_mv(void)
{
    uint32_t sum = 0;

    for (uint8_t i = 0; i < BATT_SAMPLE_COUNT; i++) {
        sum += (uint32_t)adc_read_single(BATTERY_ADC_CHANNEL);
    }

    uint16_t avg_adc = (uint16_t)(sum / BATT_SAMPLE_COUNT);

    /* Convert ADC value to millivolts:
     * V_measured = (ADC / 4096) * Vref * divider_ratio
     * V_measured_mV = (avg_adc / 4096) * 3300 * 2.0
     */
    uint32_t voltage_mv = (uint32_t)avg_adc * BATT_ADC_REF_VOLTAGE * 1000U /
                          BATT_ADC_RESOLUTION * (uint32_t)BATT_DIVIDER_RATIO;

    return (uint16_t)voltage_mv;
}

/*
 * Convert raw ADC reading to voltage in millivolts.
 */
uint16_t battery_adc_to_mv(uint16_t adc_value)
{
    uint32_t mv = (uint32_t)adc_value * BATT_ADC_REF_VOLTAGE * 1000U /
                  BATT_ADC_RESOLUTION * (uint32_t)BATT_DIVIDER_RATIO;
    return (uint16_t)mv;
}

/*
 * Calculate battery percentage from voltage.
 * Linear interpolation between empty and full thresholds.
 */
uint8_t battery_calculate_percent(uint16_t voltage_mv)
{
    uint16_t empty_mv = (uint16_t)(BATT_VOLTAGE_EMPTY * 1000.0f);
    uint16_t full_mv = (uint16_t)(BATT_VOLTAGE_FULL * 1000.0f);

    if (voltage_mv <= empty_mv) {
        return 0U;
    }
    if (voltage_mv >= full_mv) {
        return 100U;
    }

    /* Linear interpolation */
    uint16_t range = full_mv - empty_mv;
    uint16_t above_empty = voltage_mv - empty_mv;
    uint8_t percent = (uint8_t)((above_empty * 100U) / range);

    return percent;
}

/*
 * Get complete battery status.
 */
battery_status_t battery_get_status(void)
{
    battery_status_t status;

    status.voltage_mv = battery_read_voltage_mv();
    status.percent = battery_calculate_percent(status.voltage_mv);
    status.charging = false;  /* TODO: Detect charging state via GPIO */
    status.critical = (status.voltage_mv < (uint16_t)(BATT_VOLTAGE_EMPTY * 1000.0f * 0.95f));

    return status;
}

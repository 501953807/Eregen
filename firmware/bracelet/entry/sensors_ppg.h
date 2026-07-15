/*
 * Eregen (颐贞) - PPG Sensor Driver Header
 * 汇顶 GT320/GT3x heart rate + SpO2 sensor via I2C
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef SENSORS_PPG_H
#define SENSORS_PPG_H

#include <stdint.h>
#include <stdbool.h>

/* GT320 I2C address (0x98 >> 1 = 0x4C) */
#define PPG_I2C_ADDR             0x4CU

/* GT320 register addresses */
#define PPG_REG_CHIP_ID          0x00U
#define PPG_REG_SYS_CTRL         0x01U
#define PPG_REG_DATA_CTRL        0x02U
#define PPG_REG_INT_CTRL         0x03U
#define PPG_REG_MODE             0x04U
#define PPG_REG_DATA_STATUS      0x2AU
#define PPG_REG_DATA_R           0x2BU
#define PPG_REG_DATA_IR          0x2FU
#define PPG_REG_HR_RESULT        0x50U
#define PPG_REG_SPO2_RESULT      0x51U

/* HR valid range */
#define PPG_HR_MIN               30U
#define PPG_HR_MAX               220U

/* SpO2 valid range */
#define PPG_SPO2_MIN             70U
#define PPG_SPO2_MAX             100U

/* Sampling period in ms */
#define PPG_SAMPLE_INTERVAL_MS   1000U

/* Maximum I2C retry count */
#define PPG_I2C_RETRY_MAX        3U

/* Signal quality thresholds */
#define PPG_QUALITY_GOOD         75U
#define PPG_QUALITY_FAIR         50U

/*
 * Estimate signal quality based on perfusion index consistency.
 * Returns 0-100 quality indicator.
 */
uint8_t ppg_estimate_quality(void);

/* PPG health data output structure */
typedef struct {
    uint16_t hr;       /* Heart rate in BPM, 0 if invalid */
    uint8_t  spo2;     /* SpO2 percentage, 0 if invalid */
    uint8_t  quality;  /* 0-100 signal quality indicator */
} ppg_data_t;

/*
 * Read combined PPG health data (HR + SpO2 + quality).
 * @param data Pointer to ppg_data_t to populate.
 * @return true if sensor was reachable and data is fresh.
 */
bool ppg_read(ppg_data_t *data);

/*
 * Initialize the PPG sensor over I2C.
 * Returns true on success, false on failure.
 */
bool ppg_init(void);

/*
 * Read raw ADC values from the PPG sensor.
 * @param r_val    Pointer to store IR/RED ADC value
 * @param ir_val   Pointer to store IR ADC value
 * @return true if data is fresh and readable
 */
bool ppg_read_raw(uint16_t *r_val, uint16_t *ir_val);

/*
 * Calculate heart rate from sensor data.
 * @return Heart rate in BPM, clamped to [PPG_HR_MIN, PPG_HR_MAX], or 0 if invalid.
 */
uint16_t ppg_calculate_hr(void);

/*
 * Calculate SpO2 from sensor data.
 * @return SpO2 percentage, clamped to [PPG_SPO2_MIN, PPG_SPO2_MAX], or 0 if invalid.
 */
uint8_t ppg_calculate_spo2(void);

#endif /* SENSORS_PPG_H */

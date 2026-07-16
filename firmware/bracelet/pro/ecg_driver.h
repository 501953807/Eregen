/*
 * Eregen (颐贞) - ECG Driver Header
 * ADAS1000-compatible ECG front-end via I2C interface.
 * 200Hz sampling rate with built-in arrhythmia detection.
 *
 * The ADAS1000 family provides 6-channel ECG with integrated
 * lead-off detection, right-leg drive, and configurable filters.
 * This driver supports single-lead ECG (RA-LL differential).
 *
 * Key specs:
 *   - Sampling rate: 200 Hz (configurable 50-500 Hz)
 *   - Resolution: 24-bit ADC
 *   - Input range: +/- 600 mV (programmable PGA)
 *   - Lead-off detection: built-in
 *   - Arrhythmia: AFib detection via RR interval variability
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef ECG_DRIVER_H
#define ECG_DRIVER_H

#include <stdint.h>
#include <stdbool.h>

/* ----------------------------------------------------------------
 * Configuration constants
 * ---------------------------------------------------------------- */

/* Default sampling rate in Hz */
#define ECG_DEFAULT_SAMPLE_RATE_HZ   200U

/* Minimum detectable RR interval change for AFib (ms) */
#define ECG_AFRIB_STDDEV_THRESHOLD   50U

/* Rolling window size for RR std-dev calculation (samples) */
#define ECG_AFRIB_WINDOW_SIZE        10U

/* Number of consecutive AFib detections before alerting */
#define ECG_AFRIB_ALERT_COUNT        3U

/* Lead-off detection threshold (uV) */
#define ECG_LOD_THRESHOLD_UV         30U

/* Maximum ECG amplitude (mV) for clipping detection */
#define ECG_MAX_AMPLITUDE_MV         2.5f

/* I2C register addresses (ADAS1000) */
#define ECG_REG_STATUS1              0x00U
#define ECG_REG_STATUS2              0x01U
#define ECG_REG_LEADOFF              0x02U
#define ECG_REG_DAC_MID              0x04U
#define ECG_REG_RDLEDAC              0x05U
#define ECG_REG_CONFIG1              0x06U
#define ECG_REG_CONFIG2              0x07U
#define ECG_REG_CONFIG3              0x08U
#define ECG_REG_PACE                 0x09U
#define ECG_REG_WCT1                 0x0AU
#define ECG_REG_WCT2                 0x0BU
#define ECG_REG_TSD1                 0x0CU
#define ECG_REG_TSD2                 0x0DU
#define ECG_REG_MIDSENSE             0x0EU
#define ECG_REG_IOCTRL               0x0FU
#define ECG_REG_LEDACSTAT            0x10U
#define ECG_REG_RwaveSTATUS          0x11U
#define ECG_REG_ECGDATA1             0x12U
#define ECG_REG_ECGDATA2             0x13U
#define ECG_REG_ECGDATA3             0x14U

/* ----------------------------------------------------------------
 * ECG data sample (24-bit signed, returned as 32-bit int)
 * ---------------------------------------------------------------- */
typedef struct {
    int32_t raw_ecg_uv;       /* ECG amplitude in microvolts */
    uint32_t timestamp_ms;    /* Sample timestamp in ms */
    bool valid;               /* true if sample is not clipped/noisy */
    bool lead_off;            /* true if lead-off detected */
} ecg_sample_t;

/* ----------------------------------------------------------------
 * Arrhythmia detection result
 * ---------------------------------------------------------------- */
typedef struct {
    bool afib_detected;       /* true if AFib pattern found */
    float rr_stddev_ms;       /* Standard deviation of RR intervals (ms) */
    uint16_t rr_count;        /* Number of RR intervals in window */
    uint16_t alert_counter;   /* Consecutive AFib detection count */
} ecg_arrhythmia_result_t;

/* ----------------------------------------------------------------
 * ECG device state
 * ---------------------------------------------------------------- */
typedef struct {
    uint8_t i2c_addr;         /* I2C address of ECG chip */
    uint16_t sample_rate_hz;  /* Current sampling rate */
    bool measuring;           /* true when ECG acquisition is active */
    ecg_arrhythmia_result_t arrhythmia;
    ecg_sample_t last_sample;
    uint32_t last_rpeak_ms;   /* Timestamp of last R-wave peak */
    uint32_t rpeak_buffer[ECG_AFRIB_WINDOW_SIZE];
    uint8_t rpeak_index;
    uint8_t rpeak_count;
    bool lod_fault;           /* true if any lead-off fault detected */
} ecg_device_t;

/* ----------------------------------------------------------------
 * Public API
 * ---------------------------------------------------------------- */

/**
 * Initialize the ECG chip over I2C.
 * Configures PGA gain, sampling rate, filter settings.
 * @param dev Pointer to device state structure.
 * @return true on success.
 */
bool ecg_init(ecg_device_t *dev);

/**
 * Start continuous ECG acquisition.
 * Triggers the ADAS1000 to begin sampling at configured rate.
 * @param dev ECG device pointer.
 * @return true on success.
 */
bool ecg_start_measure(ecg_device_t *dev);

/**
 * Stop ECG acquisition and enter low-power mode.
 * @param dev ECG device pointer.
 */
void ecg_stop_measure(ecg_device_t *dev);

/**
 * Read a single ECG sample from the chip.
 * Must be called at or near the sampling rate (every 5ms for 200Hz).
 * @param dev ECG device pointer.
 * @param[out] sample Filled with ECG sample data.
 * @return true if sample is valid.
 */
bool ecg_read_sample(ecg_device_t *dev, ecg_sample_t *sample);

/**
 * Read a batch of ECG samples (DMA-style burst read).
 * @param dev ECG device pointer.
 * @param[out] samples Output buffer (must hold at least 'count' elements).
 * @param count Number of samples to read.
 * @return Number of samples actually read.
 */
uint8_t ecg_read_batch(ecg_device_t *dev, ecg_sample_t *samples, uint8_t count);

/**
 * Run arrhythmia detection on collected R-wave peaks.
 * Computes rolling std-dev of RR intervals.
 * @param dev ECG device pointer.
 * @return Result of arrhythmia analysis.
 */
ecg_arrhythmia_result_t ecg_detect_arrhythmia(ecg_device_t *dev);

/**
 * Check if lead-off fault is present on any electrode.
 * @param dev ECG device pointer.
 * @return true if lead-off detected.
 */
bool ecg_lead_off_check(ecg_device_t *dev);

/**
 * Get the current sampling rate.
 * @param dev ECG device pointer.
 * @return Sampling rate in Hz.
 */
uint16_t ecg_get_sample_rate(ecg_device_t *dev);

/**
 * Set the ECG sampling rate (50, 100, 200, 500 Hz supported).
 * @param dev ECG device pointer.
 * @param rate_hz Desired sampling rate.
 * @return true if rate is supported and applied.
 */
bool ecg_set_sample_rate(ecg_device_t *dev, uint16_t rate_hz);

#endif /* ECG_DRIVER_H */

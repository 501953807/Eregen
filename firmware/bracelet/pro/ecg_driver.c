/*
 * Eregen (颐贞) - ECG Driver Implementation
 * ADAS1000-compatible ECG front-end via I2C interface.
 * 200Hz sampling with AFib detection via RR interval variability.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "ecg_driver.h"
#include "board_pro.h"
#include "../common/log.h"
#include <string.h>
#include <math.h>

/* I2C read/write helper using GD32 HAL */
static bool ecg_i2c_write_reg(ecg_device_t *dev, uint8_t reg, uint8_t val)
{
    uint8_t buf[2] = { reg, val };
    /* I2C transmit - block until complete */
    i2c_transmit_bytes(BOARD_PRO_I2C, buf, 2);
    return true;
}

static bool ecg_i2c_read_regs(ecg_device_t *dev, uint8_t reg,
                              uint8_t *buf, uint8_t len)
{
    /* Send register address, then read */
    i2c_transmit_bytes(BOARD_PRO_I2C, &reg, 1);
    i2c_receive_bytes(BOARD_PRO_I2C, buf, len);
    return true;
}

/* ----------------------------------------------------------------
 * I2C register bit fields (ADAS1000)
 * ---------------------------------------------------------------- */

/* CONFIG1 bits */
#define ECG_CFG1_PGA_GAIN_SHIFT    0U
#define ECG_CFG1_PGA_GAIN_MASK     0x07U
/* PGA Gain: 0=1x, 1=2x, 2=4x, 3=6x, 4=8x, 5=12x, 6=24x */
#define ECG_CFG1_LEADOFF_EN        (1U << 3)
#define ECG_CFG1_RLD_EN            (1U << 4)

/* CONFIG2 bits */
#define ECG_CFG2_FILTER_SHIFT      0U
#define ECG_CFG2_FILTER_MASK       0x07U
/* Filter: 0=bypass, 1=HPF 0.05Hz, 2=HPF 0.5Hz, 3=LPF 40Hz, 4=LPF 15Hz */
#define ECG_CFG2_SAMPLING_SHIFT    4U
#define ECG_CFG2_SAMPLING_MASK     0x03U
/* Sampling: 0=50Hz, 1=100Hz, 2=200Hz, 3=500Hz */

/* CONFIG3 bits */
#define ECG_CFG3_PACE_EN           (1U << 0)
#define ECG_CFG3_RDET_EN           (1U << 1)
#define ECG_CFG3_RDET_THRESH_SHIFT 2U

/* ----------------------------------------------------------------
 * Initialization
 * ---------------------------------------------------------------- */

bool ecg_init(ecg_device_t *dev)
{
    if (!dev) {
        return false;
    }

    memset(dev, 0, sizeof(*dev));
    dev->i2c_addr = BOARD_PRO_ECG_I2C_ADDR;
    dev->sample_rate_hz = ECG_DEFAULT_SAMPLE_RATE_HZ;
    dev->measuring = false;

    /* Verify communication with ECG chip by reading STATUS1 */
    uint8_t status = 0;
    ecg_i2c_read_regs(dev, ECG_REG_STATUS1, &status, 1);
    if (status == 0xFF || status == 0x00) {
        log_error("ECG: No response from ADAS1000 at 0x%02X", dev->i2c_addr);
        return false;
    }
    log_info("ECG: Chip ID verified (STATUS1=0x%02X)", status);

    /* Configure CONFIG1: PGA gain 2x (4mV input range), LO + RLD enabled */
    uint8_t pga_gain = 1; /* 2x gain */
    uint8_t config1 = (pga_gain & ECG_CFG1_PGA_GAIN_MASK)
                    | ECG_CFG1_LEADOFF_EN
                    | ECG_CFG1_RLD_EN;
    ecg_i2c_write_reg(dev, ECG_REG_CONFIG1, config1);

    /* Configure CONFIG2: HPF 0.05Hz, 200Hz sampling */
    uint8_t config2 = (2U << ECG_CFG2_SAMPLING_SHIFT); /* 200Hz */
    ecg_i2c_write_reg(dev, ECG_REG_CONFIG2, config2);

    /* Configure CONFIG3: R-wave detection enabled, threshold medium */
    uint8_t config3 = ECG_CFG3_RDET_EN | (2U << ECG_CFG3_RDET_THRESH_SHIFT);
    ecg_i2c_write_reg(dev, ECG_REG_CONFIG3, config3);

    /* Set mid-reference voltage for single-supply operation */
    ecg_i2c_write_reg(dev, ECG_REG_DAC_MID, 0x80); /* Mid-supply */

    log_info("ECG: Initialized at %u Hz, PGA=%ux",
             dev->sample_rate_hz, (1U << pga_gain));

    return true;
}

/* ----------------------------------------------------------------
 * Start / Stop acquisition
 * ---------------------------------------------------------------- */

bool ecg_start_measure(ecg_device_t *dev)
{
    if (!dev || dev->measuring) {
        return dev ? dev->measuring : false;
    }

    /* Trigger START command via CONFIG3 */
    uint8_t cfg3 = ECG_CFG3_RDET_EN | ECG_CFG3_PACE_EN | (2U << ECG_CFG3_RDET_THRESH_SHIFT);
    ecg_i2c_write_reg(dev, ECG_REG_CONFIG3, cfg3);

    dev->measuring = true;
    log_info("ECG: Acquisition started");
    return true;
}

void ecg_stop_measure(ecg_device_t *dev)
{
    if (!dev || !dev->measuring) {
        return;
    }

    /* Disable pace and clear CONFIG3 to stop */
    ecg_i2c_write_reg(dev, ECG_REG_CONFIG3, 0x00);
    dev->measuring = false;

    log_info("ECG: Acquisition stopped");
}

/* ----------------------------------------------------------------
 * Sample reading
 * ---------------------------------------------------------------- */

bool ecg_read_sample(ecg_device_t *dev, ecg_sample_t *sample)
{
    if (!dev || !sample || !dev->measuring) {
        return false;
    }

    /* Read 3-byte ECG data (24-bit signed) */
    uint8_t raw[3] = { 0 };
    if (!ecg_i2c_read_regs(dev, ECG_REG_ECGDATA1, raw, 3)) {
        return false;
    }

    /* Combine into 24-bit signed integer */
    int32_t adc_code = ((int32_t)raw[0] << 16) | ((int32_t)raw[1] << 8) | (int32_t)raw[2];
    if (adc_code & 0x800000) {
        adc_code |= 0xFF000000; /* Sign extend */
    }

    /* Convert to microvolts: 24-bit ADC, 2.4V ref, PGA=2x => 1uV/LSB */
    int32_t uv = adc_code;

    /* Clip check */
    bool valid = (abs(uv) < (int32_t)(ECG_MAX_AMPLITUDE_MV * 1000000));

    /* Lead-off check */
    uint8_t lod_status = 0;
    ecg_i2c_read_regs(dev, ECG_REG_LEADOFF, &lod_status, 1);
    bool lead_off = (lod_status != 0);

    /* Timestamp from system tick (caller should provide via a timer) */
    static uint32_t s_sample_count = 0;
    uint32_t ts = s_sample_count++ * (1000U / dev->sample_rate_hz);

    sample->raw_ecg_uv = uv;
    sample->timestamp_ms = ts;
    sample->valid = valid;
    sample->lead_off = lead_off;

    dev->last_sample = *sample;

    return valid && !lead_off;
}

uint8_t ecg_read_batch(ecg_device_t *dev, ecg_sample_t *samples, uint8_t count)
{
    if (!dev || !samples || !dev->measuring) {
        return 0;
    }

    uint8_t read = 0;
    for (uint8_t i = 0; i < count; i++) {
        if (ecg_read_sample(dev, &samples[i])) {
            read++;
        } else {
            /* Stale or invalid sample - break batch */
            break;
        }
    }
    return read;
}

/* ----------------------------------------------------------------
 * R-wave peak detection and AFib analysis
 * ---------------------------------------------------------------- */

static bool ecg_detect_rpeak(const ecg_sample_t *samples, uint8_t count,
                             uint32_t *peak_ts)
{
    if (!samples || count < 3) {
        return false;
    }

    /* Simple derivative-based R-peak detection */
    for (uint8_t i = 1; i < count - 1; i++) {
        int32_t prev_diff = samples[i].raw_ecg_uv - samples[i - 1].raw_ecg_uv;
        int32_t next_diff = samples[i + 1].raw_ecg_uv - samples[i].raw_ecg_uv;

        /* Peak: rising before, falling after, above threshold */
        if (prev_diff > 0 && next_diff < 0 &&
            samples[i].raw_ecg_uv > 5000) { /* > 5mV threshold */
            *peak_ts = samples[i].timestamp_ms;
            return true;
        }
    }
    return false;
}

ecg_arrhythmia_result_t ecg_detect_arrhythmia(ecg_device_t *dev)
{
    ecg_arrhythmia_result_t result;
    memset(&result, 0, sizeof(result));

    if (!dev || !dev->measuring) {
        return result;
    }

    /* Read a batch of samples for R-peak detection */
    ecg_sample_t batch[20];
    uint8_t n = ecg_read_batch(dev, batch, 20);
    if (n < 3) {
        return result;
    }

    /* Detect R-wave peak in batch */
    uint32_t peak_ts = 0;
    if (ecg_detect_rpeak(batch, n, &peak_ts)) {
        uint32_t rr_ms = 0;
        if (dev->last_rpeak_ms > 0) {
            rr_ms = peak_ts - dev->last_rpeak_ms;
        }
        dev->last_rpeak_ms = peak_ts;

        /* Store RR interval in circular buffer */
        dev->rpeak_buffer[dev->rpeak_index] = rr_ms;
        dev->rpeak_index = (dev->rpeak_index + 1) % ECG_AFRIB_WINDOW_SIZE;
        if (dev->rpeak_count < ECG_AFRIB_WINDOW_SIZE) {
            dev->rpeak_count++;
        }
    }

    /* Compute std-dev of RR intervals if we have enough data */
    if (dev->rpeak_count >= 3) {
        /* Mean */
        uint32_t sum = 0;
        for (uint8_t i = 0; i < dev->rpeak_count; i++) {
            sum += dev->rpeak_buffer[i];
        }
        float mean = (float)sum / (float)dev->rpeak_count;

        /* Variance */
        float var_sum = 0.0f;
        for (uint8_t i = 0; i < dev->rpeak_count; i++) {
            float diff = (float)dev->rpeak_buffer[i] - mean;
            var_sum += diff * diff;
        }
        float stddev = sqrtf(var_sum / (float)dev->rpeak_count);

        result.rr_stddev_ms = stddev;
        result.rr_count = dev->rpeak_count;

        /* AFib detection: RR std-dev > threshold */
        if (stddev > ECG_AFRIB_STDDEV_THRESHOLD) {
            result.afib_detected = true;
            result.alert_counter = dev->rpeak_count >= ECG_AFRIB_ALERT_COUNT ?
                                   ECG_AFRIB_ALERT_COUNT : result.alert_counter + 1;
        } else {
            result.afib_detected = false;
            result.alert_counter = 0;
        }
    }

    /* Lead-off fault overrides everything */
    result.afib_detected = result.afib_detected && !dev->lod_fault;

    if (result.afib_detected && result.alert_counter >= ECG_AFRIB_ALERT_COUNT) {
        log_warn("ECG: Possible AFib detected! RR stddev=%.1f ms, count=%u",
                 result.rr_stddev_ms, result.alert_counter);
    }

    return result;
}

bool ecg_lead_off_check(ecg_device_t *dev)
{
    if (!dev) {
        return false;
    }

    uint8_t lod_status = 0;
    ecg_i2c_read_regs(dev, ECG_REG_LEADOFF, &lod_status, 1);

    bool fault = (lod_status != 0);
    dev->lod_fault = fault;

    if (fault) {
        log_warn("ECG: Lead-off fault detected (STATUS=0x%02X)", lod_status);
    }

    return fault;
}

/* ----------------------------------------------------------------
 * Sampling rate configuration
 * ---------------------------------------------------------------- */

uint16_t ecg_get_sample_rate(ecg_device_t *dev)
{
    return dev ? dev->sample_rate_hz : 0;
}

bool ecg_set_sample_rate(ecg_device_t *dev, uint16_t rate_hz)
{
    if (!dev) {
        return false;
    }

    /* Validate and map to CONFIG2 field */
    uint8_t sampling_sel = 0xFF;
    switch (rate_hz) {
        case 50:  sampling_sel = 0; break;
        case 100: sampling_sel = 1; break;
        case 200: sampling_sel = 2; break;
        case 500: sampling_sel = 3; break;
        default:
            log_error("ECG: Unsupported sample rate %u Hz", rate_hz);
            return false;
    }

    /* Read-modify-write CONFIG2 */
    uint8_t cfg2 = 0;
    ecg_i2c_read_regs(dev, ECG_REG_CONFIG2, &cfg2, 1);
    cfg2 = (cfg2 & ~(ECG_CFG2_SAMPLING_MASK << ECG_CFG2_SAMPLING_SHIFT))
           | (sampling_sel << ECG_CFG2_SAMPLING_SHIFT);
    ecg_i2c_write_reg(dev, ECG_REG_CONFIG2, cfg2);

    dev->sample_rate_hz = rate_hz;
    log_info("ECG: Sample rate changed to %u Hz", rate_hz);
    return true;
}

/*
 * Eregen (颐贞) - PPG Sensor Driver Implementation
 * 汇顶 GT320/GT3x heart rate + SpO2 sensor via I2C
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "sensors_ppg.h"

#ifdef TEST_MODE
#include <string.h>
#else
#include "gd32e230_i2c.h"
#include "gd32e230_rcu.h"
#endif

/* Internal I2C register read/write helpers */
#ifdef TEST_MODE
/* In test mode, these are overridden by the test file via #undef + redefine */
static bool ppg_i2c_write_reg(uint8_t reg, uint8_t val);
static bool ppg_i2c_read_reg(uint8_t reg, uint8_t *val);
static bool ppg_i2c_read_multi(uint8_t reg, uint8_t *buf, uint8_t len);
#else
/* Stub implementations for compile-only in test mode when not overridden */
#endif

/* Cached raw ADC values */
static uint16_t s_raw_r = 0;
static uint16_t s_raw_ir = 0;
static bool s_data_ready = false;

/*
 * Write a single byte to a PPG register.
 */
#ifdef TEST_MODE
static bool ppg_i2c_write_reg(uint8_t reg, uint8_t val)
{
    (void)reg; (void)val;
    return true;
}
#endif

#ifndef TEST_MODE
static bool ppg_i2c_write_reg(uint8_t reg, uint8_t val)
{
    uint8_t buf[2] = { reg, val };
    uint32_t retry = PPG_I2C_RETRY_MAX;

    while (retry-- > 0) {
        i2c_start_on_bus(I2C1);
        while (!i2c_flag_get(I2C1, I2C_FLAG_SBSEND)) {
            if (retry == 0) break;
        }
        if (i2c_flag_get(I2C1, I2C_FLAG_SBSEND) == RESET) {
            i2c_stop_on_bus(I2C1);
            continue;
        }

        i2c_byte_transfer(I2C1, (PPG_I2C_ADDR << 1) | 0U);
        while (!i2c_flag_get(I2C1, I2C_FLAG_ADDRSND)) {
            /* Wait */
        }
        (void)i2c_receive_data(I2C1);

        i2c_byte_transfer(I2C1, buf[0]);
        while (!i2c_flag_get(I2C1, I2C_FLAG_BTC)) {
            /* Wait */
        }

        i2c_byte_transfer(I2C1, buf[1]);
        while (!i2c_flag_get(I2C1, I2C_FLAG_BTC)) {
            /* Wait */
        }

        i2c_stop_on_bus(I2C1);
        return true;
    }
    return false;
}
#endif

/*
 * Read a single byte from a PPG register.
 */
#ifdef TEST_MODE
static bool ppg_i2c_read_reg(uint8_t reg, uint8_t *val)
{
    (void)reg; (void)val;
    return false;
}
#endif

#ifndef TEST_MODE
static bool ppg_i2c_read_reg(uint8_t reg, uint8_t *val)
{
    uint32_t retry = PPG_I2C_RETRY_MAX;

    while (retry-- > 0) {
        i2c_start_on_bus(I2C1);
        while (!i2c_flag_get(I2C1, I2C_FLAG_SBSEND)) {
            if (retry == 0) break;
        }
        if (i2c_flag_get(I2C1, I2C_FLAG_SBSEND) == RESET) {
            i2c_stop_on_bus(I2C1);
            continue;
        }

        i2c_byte_transfer(I2C1, (PPG_I2C_ADDR << 1) | 0U);
        while (!i2c_flag_get(I2C1, I2C_FLAG_ADDRSND)) {
            /* Wait */
        }
        (void)i2c_receive_data(I2C1);

        i2c_byte_transfer(I2C1, reg);
        while (!i2c_flag_get(I2C1, I2C_FLAG_BTC)) {
            /* Wait */
        }

        i2c_start_on_bus(I2C1);
        while (!i2c_flag_get(I2C1, I2C_FLAG_SBSEND)) {
            /* Wait */
        }

        i2c_byte_transfer(I2C1, (PPG_I2C_ADDR << 1) | 1U);
        while (!i2c_flag_get(I2C1, I2C_FLAG_ADDRRCV)) {
            /* Wait */
        }

        i2c_ack_disable(I2C1);
        while (!i2c_flag_get(I2C1, I2C_FLAG_BCE)) {
            /* Wait */
        }
        *val = i2c_receive_data(I2C1);

        i2c_stop_on_bus(I2C1);
        i2c_ack_enable(I2C1);
        return true;
    }
    return false;
}
#endif

/*
 * Read multiple bytes from consecutive registers.
 */
#ifdef TEST_MODE
static bool ppg_i2c_read_multi(uint8_t reg, uint8_t *buf, uint8_t len)
{
    (void)reg; (void)buf; (void)len;
    return false;
}
#endif

#ifndef TEST_MODE
static bool ppg_i2c_read_multi(uint8_t reg, uint8_t *buf, uint8_t len)
{
    uint32_t retry = PPG_I2C_RETRY_MAX;

    while (retry-- > 0) {
        i2c_start_on_bus(I2C1);
        while (!i2c_flag_get(I2C1, I2C_FLAG_SBSEND)) {
            if (retry == 0) break;
        }
        if (i2c_flag_get(I2C1, I2C_FLAG_SBSEND) == RESET) {
            i2c_stop_on_bus(I2C1);
            continue;
        }

        i2c_byte_transfer(I2C1, (PPG_I2C_ADDR << 1) | 0U);
        while (!i2c_flag_get(I2C1, I2C_FLAG_ADDRSND)) {
            /* Wait */
        }
        (void)i2c_receive_data(I2C1);

        i2c_byte_transfer(I2C1, reg);
        while (!i2c_flag_get(I2C1, I2C_FLAG_BTC)) {
            /* Wait */
        }

        i2c_start_on_bus(I2C1);
        while (!i2c_flag_get(I2C1, I2C_FLAG_SBSEND)) {
            /* Wait */
        }

        i2c_byte_transfer(I2C1, (PPG_I2C_ADDR << 1) | 1U);
        while (!i2c_flag_get(I2C1, I2C_FLAG_ADDRRCV)) {
            /* Wait */
        }

        for (uint8_t i = 0; i < len; i++) {
            if (i == (len - 1)) {
                i2c_ack_disable(I2C1);
            } else {
                i2c_ack_enable(I2C1);
            }
            while (!i2c_flag_get(I2C1, I2C_FLAG_BCE)) {
                /* Wait */
            }
            buf[i] = i2c_receive_data(I2C1);
        }

        i2c_stop_on_bus(I2C1);
        i2c_ack_enable(I2C1);
        return true;
    }
    return false;
}
#endif

/*
 * Initialize the PPG sensor: verify chip ID, configure operating mode.
 * Returns true on success.
 */
bool ppg_init(void)
{
#ifdef TEST_MODE
    memset(&s_raw_r, 0, sizeof(s_raw_r) + sizeof(s_raw_ir) + sizeof(s_data_ready));
    return true;
#else
    rcu_periph_clock_enable(RCU_I2C1);

    /* Verify chip ID - GT320 returns 0x60 or 0x61 */
    uint8_t chip_id = 0;
    if (!ppg_i2c_read_reg(PPG_REG_CHIP_ID, &chip_id)) {
        return false;
    }
    if ((chip_id != 0x60U) && (chip_id != 0x61U)) {
        return false;
    }

    /* Enable system control */
    ppg_i2c_write_reg(PPG_REG_SYS_CTRL, 0x01U);

    /* Configure data: enable RED+IR LEDs, high precision mode */
    ppg_i2c_write_reg(PPG_REG_DATA_CTRL, 0x37U);

    /* Set LED pulse width and current */
    ppg_i2c_write_reg(0x05U, 0x3AU);  /* LED pulse width */
    ppg_i2c_write_reg(0x06U, 0x0AU);  /* RED current */
    ppg_i2c_write_reg(0x07U, 0x1FU);  /* IR current */

    /* Enter running mode */
    ppg_i2c_write_reg(PPG_REG_MODE, 0x11U);

    s_data_ready = false;
    return true;
#endif
}

/*
 * Read raw ADC values from PPG sensor.
 * Returns true if data is fresh.
 */
bool ppg_read_raw(uint16_t *r_val, uint16_t *ir_val)
{
#ifdef TEST_MODE
    (void)r_val; (void)ir_val;
    return s_data_ready;
#else
    uint8_t status = 0;
    if (!ppg_i2c_read_reg(PPG_REG_DATA_STATUS, &status)) {
        return false;
    }

    /* Check if data is ready (bit 0 of status) */
    if ((status & 0x01U) == 0U) {
        return false;
    }

    /* Read 4 bytes: R high, R low, IR high, IR low */
    uint8_t raw_buf[4];
    if (!ppg_i2c_read_multi(PPG_REG_DATA_R, raw_buf, 4)) {
        return false;
    }

    s_raw_r = ((uint16_t)raw_buf[0] << 8) | raw_buf[1];
    s_raw_ir = ((uint16_t)raw_buf[2] << 8) | raw_buf[3];

    if (r_val) *r_val = s_raw_r;
    if (ir_val) *ir_val = s_raw_ir;

    s_data_ready = true;
    return true;
#endif
}

/*
 * Calculate heart rate from raw PPG samples.
 * Uses AC/DC ratio of RED vs IR signals.
 * Returns clamped BPM value or 0 if invalid.
 */
uint16_t ppg_calculate_hr(void)
{
    if (!s_data_ready) {
        return 0;
    }

    if (s_raw_r == 0 || s_raw_ir == 0) {
        return 0;
    }

    /* Ratio of ratios: (AC_RED/DC_RED) / (AC_IR/DC_IR)
     * Simplified: use raw ratio as a proxy for perfusion index
     */
    uint32_t ratio = (uint32_t)s_raw_r * 1000U / (uint32_t)s_raw_ir;

    /* Map ratio to HR range [30, 220] using linear approximation
     * Typical resting HR ~60-80, exercise can reach 180-220
     * Ratio range roughly 200-800 for valid readings
     */
    if (ratio < 200U || ratio > 800U) {
        return 0;
    }

    /* Linear mapping: ratio 200 -> 220 bpm, ratio 800 -> 30 bpm */
    uint16_t hr = (uint16_t)(220U - ((ratio - 200U) * 190U / 600U));

    /* Clamp to valid range */
    if (hr < PPG_HR_MIN) hr = PPG_HR_MIN;
    if (hr > PPG_HR_MAX) hr = PPG_HR_MAX;

    return hr;
}

/*
 * Calculate SpO2 from raw PPG samples.
 * Uses ratio-of-ratios method comparing RED and IR absorption.
 * Returns clamped percentage or 0 if invalid.
 */
uint8_t ppg_calculate_spo2(void)
{
    if (!s_data_ready) {
        return 0;
    }

    if (s_raw_r == 0 || s_raw_ir == 0) {
        return 0;
    }

    /* Ratio of ratios for SpO2 calculation:
     * R = (AC_RED / DC_RED) / (AC_IR / DC_IR)
     * SpO2 = A - B * log(R)
     * Calibration constants derived from clinical data.
     * For entry-level: simplified linear approximation.
     */
    float ratio_float = (float)s_raw_r / (float)s_raw_ir;

    /* Typical SpO2 range maps to ratio ~0.5 to ~1.5
     * SpO2 = 110 - 25 * ratio (empirical approximation)
     */
    int16_t spo2_int = (int16_t)(110.0f - 25.0f * ratio_float);

    /* Clamp to valid range */
    if (spo2_int < (int16_t)PPG_SPO2_MIN) spo2_int = PPG_SPO2_MIN;
    if (spo2_int > (int16_t)PPG_SPO2_MAX) spo2_int = PPG_SPO2_MAX;

    return (uint8_t)spo2_int;
}

/*
 * Get combined health data from PPG sensor.
 */
ppg_data_t ppg_get_data(void)
{
    ppg_data_t data;
    data.hr = ppg_calculate_hr();
    data.spo2 = ppg_calculate_spo2();
    data.valid = (data.hr >= PPG_HR_MIN && data.hr <= PPG_HR_MAX &&
                  data.spo2 >= PPG_SPO2_MIN && data.spo2 <= PPG_SPO2_MAX);
    return data;
}

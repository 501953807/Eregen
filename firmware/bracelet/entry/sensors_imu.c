/*
 * Eregen (颐贞) - IMU Sensor Driver Implementation
 * ICM-42670-P accelerometer + gyroscope via SPI
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "sensors_imu.h"
#include <math.h>

#ifdef TEST_MODE
#include <string.h>
#else
#include "gd32e230_spi.h"
#include "gd32e230_rcu.h"
#include "board_init.h"
#endif

/* Internal SPI register access helpers */
#ifdef TEST_MODE
/* Overridden by test file */
static bool imu_spi_write_reg(uint8_t reg, uint8_t val);
static bool imu_spi_read_reg(uint8_t reg, uint8_t *val);
static bool imu_spi_read_multi(uint8_t reg, uint8_t *buf, uint8_t len);
#else
/* Stub for compile-only */
#endif

/* Current configuration state */
static uint8_t s_accel_range = IMU_DEFAULT_ACCEL_RANGE;
static uint8_t s_gyro_range = IMU_DEFAULT_GYRO_RANGE;

/* Scale factors: g per LSB for each accel range */
static const float s_accel_scale[] = { 1.0f/16384.0f, 1.0f/8192.0f,
                                        1.0f/4096.0f, 1.0f/2048.0f };
/* Scale factors: dps per LSB for each gyro range */
static const float s_gyro_scale[] = { 1.0f/131.0f, 1.0f/65.5f,
                                       1.0f/32.8f, 1.0f/16.4f };

/* Mock data storage for TEST_MODE */
#ifdef TEST_MODE
static int16_t s_mock_ax = 0, s_mock_ay = 0, s_mock_az = 16384;
static int16_t s_mock_gx = 0, s_mock_gy = 0, s_mock_gz = 0;
#endif

/*
 * Write a single byte to an IMU SPI register.
 */
#ifdef TEST_MODE
static bool imu_spi_write_reg(uint8_t reg, uint8_t val)
{
    (void)reg; (void)val;
    return true;
}
#endif

#ifndef TEST_MODE
static bool imu_spi_write_reg(uint8_t reg, uint8_t val)
{
    /* Set CS low */
    gpio_bit_reset(IMU_CS_GPIO_PORT, IMU_CS_GPIO_PIN);

    /* Send register address with write bit (0x80 | reg) */
    uint8_t cmd = (uint8_t)(0x80U | reg);
    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
        /* Wait */
    }
    spi_i2s_data_transmit(SPI1, cmd);
    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }

    /* Send data byte */
    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
        /* Wait */
    }
    spi_i2s_data_transmit(SPI1, val);
    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }

    /* Set CS high */
    gpio_bit_set(IMU_CS_GPIO_PORT, IMU_CS_GPIO_PIN);

    return true;
}
#endif

/*
 * Read a single byte from an IMU SPI register.
 */
#ifdef TEST_MODE
static bool imu_spi_read_reg(uint8_t reg, uint8_t *val)
{
    (void)reg; (void)val;
    return false;
}
#endif

#ifndef TEST_MODE
static bool imu_spi_read_reg(uint8_t reg, uint8_t *val)
{
    *val = 0xFFU;

    /* Set CS low */
    gpio_bit_reset(IMU_CS_GPIO_PORT, IMU_CS_GPIO_PIN);

    /* Send register address with read bit (0x7F & reg) */
    uint8_t cmd = (uint8_t)(0x7FU & reg);
    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
        /* Wait */
    }
    spi_i2s_data_transmit(SPI1, cmd);
    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }

    /* Dummy read to get response */
    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
        /* Wait */
    }
    spi_i2s_data_transmit(SPI1, 0xFFU);
    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }
    while (spi_flag_get(SPI1, SPI_FLAG_RBNE) == RESET) {
        /* Wait */
    }
    *val = (uint8_t)spi_i2s_data_receive(SPI1);

    /* Set CS high */
    gpio_bit_set(IMU_CS_GPIO_PORT, IMU_CS_GPIO_PIN);

    return true;
}
#endif

/*
 * Read multiple bytes from consecutive IMU SPI registers.
 */
#ifdef TEST_MODE
static bool imu_spi_read_multi(uint8_t reg, uint8_t *buf, uint8_t len)
{
    (void)reg; (void)buf; (void)len;
    return false;
}
#endif

#ifndef TEST_MODE
static bool imu_spi_read_multi(uint8_t reg, uint8_t *buf, uint8_t len)
{
    /* Set CS low */
    gpio_bit_reset(IMU_CS_GPIO_PORT, IMU_CS_GPIO_PIN);

    /* Send register address with read bit + auto-increment */
    uint8_t cmd = (uint8_t)(0x80U | 0x40U | reg);
    while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
        /* Wait */
    }
    spi_i2s_data_transmit(SPI1, cmd);
    while (spi_flag_get(SPI1, SPI_FLAG_BSY) != RESET) {
        /* Wait */
    }

    /* Read all bytes */
    for (uint8_t i = 0; i < len; i++) {
        while (spi_flag_get(SPI1, SPI_FLAG_TBE) == RESET) {
            /* Wait */
        }
        spi_i2s_data_transmit(SPI1, 0xFFU);
        while (spi_flag_get(SPI1, SPI_FLAG_RBNE) == RESET) {
            /* Wait */
        }
        buf[i] = (uint8_t)spi_i2s_data_receive(SPI1);
    }

    /* Set CS high */
    gpio_bit_set(IMU_CS_GPIO_PORT, IMU_CS_GPIO_PIN);

    return true;
}
#endif

/*
 * Initialize the IMU sensor.
 * Verifies device ID (0xEF), configures default settings.
 */
bool imu_init(void)
{
#ifdef TEST_MODE
    memset(&s_accel_range, 0, sizeof(s_accel_range) + sizeof(s_gyro_range));
    s_accel_range = IMU_DEFAULT_ACCEL_RANGE;
    s_gyro_range = IMU_DEFAULT_GYRO_RANGE;
    return true;
#else
    rcu_periph_clock_enable(RCU_SPI1);

    /* Verify device ID: ICM-42670-P returns 0xEF */
    uint8_t dev_id = 0;
    if (!imu_spi_read_reg(IMU_REG_DEVICE_ID, &dev_id)) {
        return false;
    }
    if (dev_id != 0xEFU) {
        return false;
    }

    /* Put sensor into standby mode for configuration */
    imu_spi_write_reg(IMU_REG_SENSOR_CONFIG0, 0x01U);

    /* Configure accelerometer: 4g range, 100Hz ODR */
    imu_spi_write_reg(IMU_REG_ACCEL_CONFIG, IMU_DEFAULT_ACCEL_RANGE);
    imu_spi_write_reg(IMU_REG_ACCEL_CONFIG2, 0x14U);  /* 100Hz, avg=1 */

    /* Configure gyroscope: 500dps range, 100Hz ODR */
    imu_spi_write_reg(IMU_REG_GYRO_CONFIG, IMU_DEFAULT_GYRO_RANGE);
    imu_spi_write_reg(IMU_REG_GYRO_CONFIG2, 0x14U);   /* 100Hz */

    /* Enable accelerometer and gyroscope */
    imu_spi_write_reg(IMU_REG_PWR_CONFIG, 0x1FU);  /* Temp off, accel+gyro on */

    /* Signal path: enable both accel and gyro data paths */
    imu_spi_write_reg(IMU_REG_SIGNAL_PATH_CFG, 0x00U);

    s_accel_range = IMU_DEFAULT_ACCEL_RANGE;
    s_gyro_range = IMU_DEFAULT_GYRO_RANGE;

    return true;
#endif
}

/*
 * Read raw accelerometer values.
 */
bool imu_read_accel_raw(int16_t *x, int16_t *y, int16_t *z)
{
#ifdef TEST_MODE
    if (x) *x = s_mock_ax;
    if (y) *y = s_mock_ay;
    if (z) *z = s_mock_az;
    return true;
#else
    uint8_t buf[6];
    if (!imu_spi_read_multi(IMU_REG_ACCEL_X0, buf, 6)) {
        return false;
    }

    if (x) *x = (int16_t)((buf[0] << 8) | buf[1]);
    if (y) *y = (int16_t)((buf[2] << 8) | buf[3]);
    if (z) *z = (int16_t)((buf[4] << 8) | buf[5]);

    return true;
#endif
}

/*
 * Read raw gyroscope values.
 */
bool imu_read_gyro_raw(int16_t *x, int16_t *y, int16_t *z)
{
#ifdef TEST_MODE
    if (x) *x = s_mock_gx;
    if (y) *y = s_mock_gy;
    if (z) *z = s_mock_gz;
    return true;
#else
    uint8_t buf[6];
    if (!imu_spi_read_multi(IMU_REG_GYRO_X0, buf, 6)) {
        return false;
    }

    if (x) *x = (int16_t)((buf[0] << 8) | buf[1]);
    if (y) *y = (int16_t)((buf[2] << 8) | buf[3]);
    if (z) *z = (int16_t)((buf[4] << 8) | buf[5]);

    return true;
#endif
}

/*
 * Read calibrated acceleration data in g units.
 */
imu_data_t imu_read_accel(void)
{
    imu_data_t data;
    int16_t raw[3];

    if (!imu_read_accel_raw(&raw[0], &raw[1], &raw[2])) {
        data.ax = data.ay = data.az = 0.0f;
        return data;
    }

    float scale = s_accel_scale[s_accel_range];
    data.ax = (float)raw[0] * scale;
    data.ay = (float)raw[1] * scale;
    data.az = (float)raw[2] * scale;

    return data;
}

/*
 * Read calibrated gyroscope data in dps units.
 */
imu_data_t imu_read_gyro(void)
{
    imu_data_t data;
    int16_t raw[3];

    if (!imu_read_gyro_raw(&raw[0], &raw[1], &raw[2])) {
        data.gx = data.gy = data.gz = 0.0f;
        return data;
    }

    float scale = s_gyro_scale[s_gyro_range];
    data.gx = (float)raw[0] * scale;
    data.gy = (float)raw[1] * scale;
    data.gz = (float)raw[2] * scale;

    return data;
}

/*
 * Read all IMU data at once.
 */
imu_data_t imu_get_data(void)
{
    imu_data_t data;
    int16_t ax_raw, ay_raw, az_raw, gx_raw, gy_raw, gz_raw;

    /* Try combined read first */
#ifdef TEST_MODE
    /* Use individual reads in test mode */
    imu_read_accel_raw(&ax_raw, &ay_raw, &az_raw);
    imu_read_gyro_raw(&gx_raw, &gy_raw, &gz_raw);
#else
    uint8_t buf[12];
    if (!imu_spi_read_multi(IMU_REG_ACCEL_X0, buf, 12)) {
        data.ax = data.ay = data.az = 0.0f;
        data.gx = data.gy = data.gz = 0.0f;
        return data;
    }

    ax_raw = (int16_t)((buf[0] << 8) | buf[1]);
    ay_raw = (int16_t)((buf[2] << 8) | buf[3]);
    az_raw = (int16_t)((buf[4] << 8) | buf[5]);
    gx_raw = (int16_t)((buf[6] << 8) | buf[7]);
    gy_raw = (int16_t)((buf[8] << 8) | buf[9]);
    gz_raw = (int16_t)((buf[10] << 8) | buf[11]);
#endif

    /* Parse accelerometer */
    float accel_scale = s_accel_scale[s_accel_range];
    data.ax = (float)ax_raw * accel_scale;
    data.ay = (float)ay_raw * accel_scale;
    data.az = (float)az_raw * accel_scale;

    /* Parse gyroscope */
    float gyro_scale = s_gyro_scale[s_gyro_range];
    data.gx = (float)gx_raw * gyro_scale;
    data.gy = (float)gy_raw * gyro_scale;
    data.gz = (float)gz_raw * gyro_scale;

    return data;
}

/*
 * Configure accelerometer full-scale range.
 */
void imu_set_accel_range(uint8_t range)
{
    if (range <= IMU_ACCEL_RANGE_16G) {
        s_accel_range = range;
#ifdef TEST_MODE
        (void)imu_spi_write_reg(IMU_REG_ACCEL_CONFIG, range);
#else
        imu_spi_write_reg(IMU_REG_ACCEL_CONFIG, range);
#endif
    }
}

/*
 * Configure gyroscope full-scale range.
 */
void imu_set_gyro_range(uint8_t range)
{
    if (range <= IMU_GYRO_RANGE_2000DPS) {
        s_gyro_range = range;
#ifdef TEST_MODE
        (void)imu_spi_write_reg(IMU_REG_GYRO_CONFIG, range);
#else
        imu_spi_write_reg(IMU_REG_GYRO_CONFIG, range);
#endif
    }
}

/*
 * Calculate acceleration magnitude in g.
 */
float imu_accel_magnitude(const imu_data_t *data)
{
    return sqrtf(data->ax * data->ax +
                 data->ay * data->ay +
                 data->az * data->az);
}

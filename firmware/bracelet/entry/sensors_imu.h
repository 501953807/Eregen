/*
 * Eregen (颐贞) - IMU Sensor Driver Header
 * ICM-42670-P accelerometer + gyroscope via SPI
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef SENSORS_IMU_H
#define SENSORS_IMU_H

#include <stdint.h>
#include <stdbool.h>

/* ICM-42670 SPI address (MSB=0 for read, MSB=1 for write) */
#define IMU_SPI_ADDR_READ      0x30U
#define IMU_SPI_ADDR_WRITE     0x31U

/* ICM-42670 register addresses */
#define IMU_REG_DEVICE_ID      0x00U
#define IMU_REG_INT_CONFIG     0x0BU
#define IMU_REG_PWR_CONFIG     0x1BU
#define IMU_REG_SENSOR_CONFIG0 0x1DU
#define IMU_REG_ACCEL_CONFIG   0x40U
#define IMU_REG_ACCEL_CONFIG2  0x41U
#define IMU_REG_GYRO_CONFIG    0x42U
#define IMU_REG_GYRO_CONFIG2   0x43U
#define IMU_REG_TEMP_CONFIG    0x44U
#define IMU_REG_FIFO_CONFIG    0x4AU
#define IMU_REG_SIGNAL_PATH_CFG 0x50U
#define IMU_REG_STATUS         0x46U
#define IMU_REG_FIFO_LEVEL     0x4CU
#define IMU_REG_FIFO_WR        0x4DU
#define IMU_REG_INT_STATUS     0x49U
#define IMU_REG_TEMP0          0x48U
#define IMU_REG_TEMP1          0x49U
#define IMU_REG_ACCEL_X0       0x4FU
#define IMU_REG_ACCEL_Y0       0x51U
#define IMU_REG_ACCEL_Z0       0x53U
#define IMU_REG_GYRO_X0        0x55U
#define IMU_REG_GYRO_Y0        0x57U
#define IMU_REG_GYRO_Z0        0x59U

/* Accelerometer full-scale ranges */
#define IMU_ACCEL_RANGE_2G     0x00U
#define IMU_ACCEL_RANGE_4G     0x01U
#define IMU_ACCEL_RANGE_8G     0x02U
#define IMU_ACCEL_RANGE_16G    0x03U

/* Gyroscope full-scale ranges */
#define IMU_GYRO_RANGE_250DPS  0x00U
#define IMU_GYRO_RANGE_500DPS  0x01U
#define IMU_GYRO_RANGE_1000DPS 0x02U
#define IMU_GYRO_RANGE_2000DPS 0x03U

/* Default configuration */
#define IMU_DEFAULT_ACCEL_RANGE   IMU_ACCEL_RANGE_4G
#define IMU_DEFAULT_GYRO_RANGE    IMU_GYRO_RANGE_500DPS
#define IMU_DEFAULT_ODR           100U  /* 100 Hz sample rate */

/* IMU measurement data structure */
typedef struct {
    float ax;  /* Acceleration X, in g */
    float ay;  /* Acceleration Y, in g */
    float az;  /* Acceleration Z, in g */
    float gx;  /* Gyro X, in dps */
    float gy;  /* Gyro Y, in dps */
    float gz;  /* Gyro Z, in dps */
} imu_data_t;

/*
 * Initialize the IMU sensor over SPI.
 * Verifies device ID and configures default settings.
 * Returns true on success.
 */
bool imu_init(void);

/*
 * Read raw accelerometer values (signed 16-bit).
 * @param x Out: X axis raw value
 * @param y Out: Y axis raw value
 * @param z Out: Z axis raw value
 * @return true on success
 */
bool imu_read_accel_raw(int16_t *x, int16_t *y, int16_t *z);

/*
 * Read raw gyroscope values (signed 16-bit).
 * @param x Out: X axis raw value
 * @param y Out: Y axis raw value
 * @param z Out: Z axis raw value
 * @return true on success
 */
bool imu_read_gyro_raw(int16_t *x, int16_t *y, int16_t *z);

/*
 * Read calibrated acceleration data in g units.
 * @return imu_data_t with ax, ay, az populated.
 */
imu_data_t imu_read_accel(void);

/*
 * Read calibrated gyroscope data in dps units.
 * @return imu_data_t with gx, gy, gz populated.
 */
imu_data_t imu_read_gyro(void);

/*
 * Read all IMU data at once.
 * @return imu_data_t with all six axes.
 */
imu_data_t imu_get_data(void);

/*
 * Configure accelerometer full-scale range.
 * @param range One of IMU_ACCEL_RANGE_* constants.
 */
void imu_set_accel_range(uint8_t range);

/*
 * Configure gyroscope full-scale range.
 * @param range One of IMU_GYRO_RANGE_* constants.
 */
void imu_set_gyro_range(uint8_t range);

/*
 * Calculate the magnitude of acceleration vector.
 * @param data Pointer to imu_data_t
 * @return Magnitude in g
 */
float imu_accel_magnitude(const imu_data_t *data);

#endif /* SENSORS_IMU_H */

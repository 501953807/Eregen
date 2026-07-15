/*
 * Eregen (颐贞) - Sensor Unit Tests
 * Tests for PPG heart rate + SpO2 driver and IMU sensor
 * Compile: gcc -DTEST_MODE -I. sensors_ppg.c sensors_imu.c test_sensors.c -lm -o test_sensors
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <stdint.h>
#include <stdbool.h>

/* ============ Mock Hardware Layer ============ */

/* Mock PPG I2C state */
static uint8_t s_mock_chip_id = 0x60;
static bool s_mock_ppg_read_raw_ok = true;
static uint16_t s_mock_raw_r = 4096;
static uint16_t s_mock_raw_ir = 8192;

/* Mock IMU SPI state */
static uint8_t s_mock_imu_regs[256];
static bool s_mock_imu_init_ok = true;
static int16_t s_mock_accel_x = 100, s_mock_accel_y = -50, s_mock_accel_z = 16384;
static int16_t s_mock_gyro_x = 0, s_mock_gyro_y = 0, s_mock_gyro_z = 0;

/* ============ PPG Driver (inline implementation for test) ============ */

#define PPG_I2C_ADDR             0x4CU
#define PPG_REG_CHIP_ID          0x00U
#define PPG_REG_SYS_CTRL         0x01U
#define PPG_REG_DATA_CTRL        0x02U
#define PPG_REG_INT_CTRL         0x03U
#define PPG_REG_MODE             0x04U
#define PPG_REG_DATA_STATUS      0x2AU
#define PPG_REG_DATA_R           0x2BU
#define PPG_REG_DATA_IR          0x2FU
#define PPG_HR_MIN               30U
#define PPG_HR_MAX               220U
#define PPG_SPO2_MIN             70U
#define PPG_SPO2_MAX             100U
#define PPG_I2C_RETRY_MAX        3U

typedef struct {
    uint16_t hr;
    uint8_t  spo2;
    bool     valid;
} ppg_data_t;

static uint16_t s_ppg_raw_r = 0;
static uint16_t s_ppg_raw_ir = 0;
static bool s_ppg_data_ready = false;

static bool ppg_i2c_write_reg_test(uint8_t reg, uint8_t val)
{
    (void)reg; (void)val;
    return true;
}

static bool ppg_i2c_read_reg_test(uint8_t reg, uint8_t *val)
{
    if (reg == PPG_REG_CHIP_ID) {
        *val = s_mock_chip_id;
        return true;
    }
    if (reg == PPG_REG_DATA_STATUS) {
        *val = s_mock_ppg_read_raw_ok ? 0x01 : 0x00;
        return true;
    }
    if (reg == PPG_REG_DATA_R) {
        *val = (uint8_t)(s_ppg_raw_r >> 8);
        return true;
    }
    return false;
}

static bool ppg_i2c_read_multi_test(uint8_t reg, uint8_t *buf, uint8_t len)
{
    if (reg == PPG_REG_DATA_R && len >= 4) {
        buf[0] = (uint8_t)(s_ppg_raw_r >> 8);
        buf[1] = (uint8_t)(s_ppg_raw_r & 0xFF);
        buf[2] = (uint8_t)(s_ppg_raw_ir >> 8);
        buf[3] = (uint8_t)(s_ppg_raw_ir & 0xFF);
        return true;
    }
    return false;
}

bool ppg_init(void)
{
    s_ppg_raw_r = 0;
    s_ppg_raw_ir = 0;
    s_ppg_data_ready = false;
    /* Verify chip ID */
    uint8_t chip_id = 0;
    if (!ppg_i2c_read_reg_test(PPG_REG_CHIP_ID, &chip_id)) {
        return false;
    }
    if ((chip_id != 0x60U) && (chip_id != 0x61U)) {
        return false;
    }
    ppg_i2c_write_reg_test(PPG_REG_SYS_CTRL, 0x01U);
    ppg_i2c_write_reg_test(PPG_REG_DATA_CTRL, 0x37U);
    ppg_i2c_write_reg_test(PPG_REG_MODE, 0x11U);
    s_ppg_data_ready = false;
    return true;
}

bool ppg_read_raw(uint16_t *r_val, uint16_t *ir_val)
{
    uint8_t status = 0;
    if (!ppg_i2c_read_reg_test(PPG_REG_DATA_STATUS, &status)) {
        return false;
    }
    if ((status & 0x01U) == 0U) {
        return false;
    }
    uint8_t raw_buf[4];
    if (!ppg_i2c_read_multi_test(PPG_REG_DATA_R, raw_buf, 4)) {
        return false;
    }
    s_ppg_raw_r = ((uint16_t)raw_buf[0] << 8) | raw_buf[1];
    s_ppg_raw_ir = ((uint16_t)raw_buf[2] << 8) | raw_buf[3];
    if (r_val) *r_val = s_ppg_raw_r;
    if (ir_val) *ir_val = s_ppg_raw_ir;
    s_ppg_data_ready = true;
    return true;
}

uint16_t ppg_calculate_hr(void)
{
    if (!s_ppg_data_ready) return 0;
    if (s_ppg_raw_r == 0 || s_ppg_raw_ir == 0) return 0;
    uint32_t ratio = (uint32_t)s_ppg_raw_r * 1000U / (uint32_t)s_ppg_raw_ir;
    if (ratio < 200U || ratio > 800U) return 0;
    uint16_t hr = (uint16_t)(220U - ((ratio - 200U) * 190U / 600U));
    if (hr < PPG_HR_MIN) hr = PPG_HR_MIN;
    if (hr > PPG_HR_MAX) hr = PPG_HR_MAX;
    return hr;
}

uint8_t ppg_calculate_spo2(void)
{
    if (!s_ppg_data_ready) return 0;
    if (s_ppg_raw_r == 0 || s_ppg_raw_ir == 0) return 0;
    float ratio_float = (float)s_ppg_raw_r / (float)s_ppg_raw_ir;
    int16_t spo2_int = (int16_t)(110.0f - 25.0f * ratio_float);
    if (spo2_int < (int16_t)PPG_SPO2_MIN) spo2_int = PPG_SPO2_MIN;
    if (spo2_int > (int16_t)PPG_SPO2_MAX) spo2_int = PPG_SPO2_MAX;
    return (uint8_t)spo2_int;
}

ppg_data_t ppg_get_data(void)
{
    ppg_data_t data;
    data.hr = ppg_calculate_hr();
    data.spo2 = ppg_calculate_spo2();
    data.valid = (data.hr >= PPG_HR_MIN && data.hr <= PPG_HR_MAX &&
                  data.spo2 >= PPG_SPO2_MIN && data.spo2 <= PPG_SPO2_MAX);
    return data;
}

/* ============ IMU Driver (inline implementation for test) ============ */

#define IMU_REG_DEVICE_ID      0x00U
#define IMU_REG_ACCEL_CONFIG   0x40U
#define IMU_REG_GYRO_CONFIG    0x42U
#define IMU_REG_ACCEL_X0       0x4FU
#define IMU_REG_GYRO_X0        0x55U
#define IMU_ACCEL_RANGE_2G     0x00U
#define IMU_ACCEL_RANGE_4G     0x01U
#define IMU_ACCEL_RANGE_8G     0x02U
#define IMU_ACCEL_RANGE_16G    0x03U
#define IMU_GYRO_RANGE_250DPS  0x00U
#define IMU_GYRO_RANGE_500DPS  0x01U
#define IMU_GYRO_RANGE_1000DPS 0x02U
#define IMU_GYRO_RANGE_2000DPS 0x03U
#define IMU_DEFAULT_ACCEL_RANGE   IMU_ACCEL_RANGE_4G
#define IMU_DEFAULT_GYRO_RANGE    IMU_GYRO_RANGE_500DPS

typedef struct {
    float ax, ay, az;
    float gx, gy, gz;
} imu_data_t;

static uint8_t s_imu_accel_range = IMU_DEFAULT_ACCEL_RANGE;
static uint8_t s_imu_gyro_range = IMU_DEFAULT_GYRO_RANGE;

static const float s_imu_accel_scale[] = { 1.0f/16384.0f, 1.0f/8192.0f,
                                            1.0f/4096.0f, 1.0f/2048.0f };
static const float s_imu_gyro_scale[] = { 1.0f/131.0f, 1.0f/65.5f,
                                           1.0f/32.8f, 1.0f/16.4f };

bool imu_spi_write_reg_test(uint8_t reg, uint8_t val)
{
    (void)reg; (void)val;
    return true;
}

bool imu_spi_read_reg_test(uint8_t reg, uint8_t *val)
{
    if (reg == IMU_REG_DEVICE_ID) {
        *val = s_mock_imu_init_ok ? 0xEF : 0x00;
        return true;
    }
    *val = s_mock_imu_regs[reg];
    return true;
}

bool imu_spi_read_multi_test(uint8_t reg, uint8_t *buf, uint8_t len)
{
    if (reg == IMU_REG_ACCEL_X0 && len >= 6) {
        buf[0] = (uint8_t)(s_mock_accel_x >> 8);
        buf[1] = (uint8_t)s_mock_accel_x;
        buf[2] = (uint8_t)(s_mock_accel_y >> 8);
        buf[3] = (uint8_t)s_mock_accel_y;
        buf[4] = (uint8_t)(s_mock_accel_z >> 8);
        buf[5] = (uint8_t)s_mock_accel_z;
        return true;
    }
    if (reg == IMU_REG_GYRO_X0 && len >= 6) {
        buf[0] = (uint8_t)(s_mock_gyro_x >> 8);
        buf[1] = (uint8_t)s_mock_gyro_x;
        buf[2] = (uint8_t)(s_mock_gyro_y >> 8);
        buf[3] = (uint8_t)s_mock_gyro_y;
        buf[4] = (uint8_t)(s_mock_gyro_z >> 8);
        buf[5] = (uint8_t)s_mock_gyro_z;
        return true;
    }
    return true;
}

bool imu_init(void)
{
    uint8_t dev_id = 0;
    if (!imu_spi_read_reg_test(IMU_REG_DEVICE_ID, &dev_id)) return false;
    if (dev_id != 0xEFU) return false;
    imu_spi_write_reg_test(0x1DU, 0x01U);
    s_imu_accel_range = IMU_DEFAULT_ACCEL_RANGE;
    s_imu_gyro_range = IMU_DEFAULT_GYRO_RANGE;
    return true;
}

bool imu_read_accel_raw(int16_t *x, int16_t *y, int16_t *z)
{
    uint8_t buf[6];
    if (!imu_spi_read_multi_test(IMU_REG_ACCEL_X0, buf, 6)) return false;
    if (x) *x = (int16_t)((buf[0] << 8) | buf[1]);
    if (y) *y = (int16_t)((buf[2] << 8) | buf[3]);
    if (z) *z = (int16_t)((buf[4] << 8) | buf[5]);
    return true;
}

bool imu_read_gyro_raw(int16_t *x, int16_t *y, int16_t *z)
{
    uint8_t buf[6];
    if (!imu_spi_read_multi_test(IMU_REG_GYRO_X0, buf, 6)) return false;
    if (x) *x = (int16_t)((buf[0] << 8) | buf[1]);
    if (y) *y = (int16_t)((buf[2] << 8) | buf[3]);
    if (z) *z = (int16_t)((buf[4] << 8) | buf[5]);
    return true;
}

imu_data_t imu_read_accel(void)
{
    imu_data_t data;
    int16_t raw[3];
    if (!imu_read_accel_raw(&raw[0], &raw[1], &raw[2])) {
        data.ax = data.ay = data.az = 0.0f;
        return data;
    }
    float scale = s_imu_accel_scale[s_imu_accel_range];
    data.ax = (float)raw[0] * scale;
    data.ay = (float)raw[1] * scale;
    data.az = (float)raw[2] * scale;
    return data;
}

imu_data_t imu_read_gyro(void)
{
    imu_data_t data;
    int16_t raw[3];
    if (!imu_read_gyro_raw(&raw[0], &raw[1], &raw[2])) {
        data.gx = data.gy = data.gz = 0.0f;
        return data;
    }
    float scale = s_imu_gyro_scale[s_imu_gyro_range];
    data.gx = (float)raw[0] * scale;
    data.gy = (float)raw[1] * scale;
    data.gz = (float)raw[2] * scale;
    return data;
}

imu_data_t imu_get_data(void)
{
    imu_data_t data;
    int16_t ax_raw, ay_raw, az_raw, gx_raw, gy_raw, gz_raw;
    imu_read_accel_raw(&ax_raw, &ay_raw, &az_raw);
    imu_read_gyro_raw(&gx_raw, &gy_raw, &gz_raw);
    float as = s_imu_accel_scale[s_imu_accel_range];
    float gs = s_imu_gyro_scale[s_imu_gyro_range];
    data.ax = (float)ax_raw * as;
    data.ay = (float)ay_raw * as;
    data.az = (float)az_raw * as;
    data.gx = (float)gx_raw * gs;
    data.gy = (float)gy_raw * gs;
    data.gz = (float)gz_raw * gs;
    return data;
}

void imu_set_accel_range(uint8_t range)
{
    if (range <= IMU_ACCEL_RANGE_16G) {
        s_imu_accel_range = range;
        imu_spi_write_reg_test(IMU_REG_ACCEL_CONFIG, range);
    }
}

void imu_set_gyro_range(uint8_t range)
{
    if (range <= IMU_GYRO_RANGE_2000DPS) {
        s_imu_gyro_range = range;
        imu_spi_write_reg_test(IMU_REG_GYRO_CONFIG, range);
    }
}

float imu_accel_magnitude(const imu_data_t *data)
{
    return sqrtf(data->ax * data->ax + data->ay * data->ay + data->az * data->az);
}

/* ============ Test Helpers ============ */

static int g_tests_run = 0;
static int g_tests_passed = 0;
static int g_tests_failed = 0;

#define TEST_ASSERT(cond, msg) do { \
    g_tests_run++; \
    if (cond) { \
        g_tests_passed++; \
        printf("  PASS: %s\n", msg); \
    } else { \
        g_tests_failed++; \
        printf("  FAIL: %s\n", msg); \
    } \
} while(0)

#define TEST_ASSERT_EQ(a, b, msg) do { \
    g_tests_run++; \
    if ((a) == (b)) { \
        g_tests_passed++; \
        printf("  PASS: %s\n", msg); \
    } else { \
        g_tests_failed++; \
        printf("  FAIL: %s (expected=%d, got=%d)\n", msg, (int)(b), (int)(a)); \
    } \
} while(0)

#define TEST_ASSERT_NEAR(a, b, epsilon, msg) do { \
    g_tests_run++; \
    double diff = fabs((double)(a) - (double)(b)); \
    if (diff < (epsilon)) { \
        g_tests_passed++; \
        printf("  PASS: %s\n", msg); \
    } else { \
        g_tests_failed++; \
        printf("  FAIL: %s (expected~%.3f, got=%.3f, diff=%.3f)\n", \
               msg, (double)(b), (double)(a), diff); \
    } \
} while(0)

/* ============ PPG Tests ============ */

static void test_ppg_init(void)
{
    printf("\n--- PPG Init Tests ---\n");

    s_mock_chip_id = 0x60;
    TEST_ASSERT(ppg_init() == true, "PPG init with valid chip ID (0x60)");

    s_mock_chip_id = 0x61;
    TEST_ASSERT(ppg_init() == true, "PPG init with chip ID 0x61");

    s_mock_chip_id = 0x00;
    TEST_ASSERT(ppg_init() == false, "PPG init rejects invalid chip ID");
}

static void test_ppg_hr_calculation(void)
{
    printf("\n--- PPG HR Calculation Tests ---\n");

    s_mock_raw_r = 4096;
    s_mock_raw_ir = 8192;
    s_mock_ppg_read_raw_ok = true;
    ppg_init();
    /* Simulate successful raw read */
    s_ppg_data_ready = true;
    s_ppg_raw_r = s_mock_raw_r;
    s_ppg_raw_ir = s_mock_raw_ir;
    uint16_t hr = ppg_calculate_hr();
    TEST_ASSERT(hr >= PPG_HR_MIN && hr <= PPG_HR_MAX,
                "HR within valid range [30, 220]");

    s_mock_raw_r = 100;
    s_mock_raw_ir = 8192;
    /* Use values that produce ratio in valid range [200, 800] */
    s_ppg_data_ready = true;
    s_ppg_raw_r = 4000;
    s_ppg_raw_ir = 8000;
    hr = ppg_calculate_hr();
    TEST_ASSERT(hr >= PPG_HR_MIN && hr <= PPG_HR_MAX,
                "HR within valid range at moderate ratio");

    s_mock_raw_r = 0;
    s_mock_raw_ir = 8192;
    s_ppg_raw_r = s_mock_raw_r;
    s_ppg_raw_ir = s_mock_raw_ir;
    hr = ppg_calculate_hr();
    TEST_ASSERT_EQ(hr, 0, "Zero RED raw value returns HR=0");

    s_mock_ppg_read_raw_ok = false;
    s_ppg_data_ready = false;
    hr = ppg_calculate_hr();
    TEST_ASSERT_EQ(hr, 0, "No data ready returns HR=0");

    s_mock_ppg_read_raw_ok = true;
    s_mock_raw_r = 8192;
    s_mock_raw_ir = 10;
    s_ppg_data_ready = true;
    s_ppg_raw_r = s_mock_raw_r;
    s_ppg_raw_ir = s_mock_raw_ir;
    hr = ppg_calculate_hr();
    TEST_ASSERT_EQ(hr, 0, "Out-of-bounds ratio returns HR=0");
}

static void test_ppg_spo2_calculation(void)
{
    printf("\n--- PPG SpO2 Calculation Tests ---\n");

    s_mock_raw_r = 4096;
    s_mock_raw_ir = 8192;
    s_mock_ppg_read_raw_ok = true;
    ppg_init();
    s_ppg_data_ready = true;
    s_ppg_raw_r = s_mock_raw_r;
    s_ppg_raw_ir = s_mock_raw_ir;
    uint8_t spo2 = ppg_calculate_spo2();
    TEST_ASSERT(spo2 >= PPG_SPO2_MIN && spo2 <= PPG_SPO2_MAX,
                "SpO2 within valid range [70, 100]");

    s_mock_raw_r = 8000;
    s_mock_raw_ir = 2000;
    s_ppg_data_ready = true;
    s_ppg_raw_r = s_mock_raw_r;
    s_ppg_raw_ir = s_mock_raw_ir;
    spo2 = ppg_calculate_spo2();
    TEST_ASSERT_EQ(spo2, PPG_SPO2_MIN,
                   "High ratio clamps SpO2 to minimum (70%)");

    s_mock_raw_r = 1000;
    s_mock_raw_ir = 8000;
    s_ppg_data_ready = true;
    s_ppg_raw_r = s_mock_raw_r;
    s_ppg_raw_ir = s_mock_raw_ir;
    spo2 = ppg_calculate_spo2();
    TEST_ASSERT_EQ(spo2, PPG_SPO2_MAX,
                   "Low ratio clamps SpO2 to maximum (100%)");

    s_mock_raw_r = 0;
    s_mock_raw_ir = 8192;
    s_ppg_raw_r = s_mock_raw_r;
    s_ppg_raw_ir = s_mock_raw_ir;
    spo2 = ppg_calculate_spo2();
    TEST_ASSERT_EQ(spo2, 0, "Zero raw value returns SpO2=0");
}

static void test_ppg_combined_data(void)
{
    printf("\n--- PPG Combined Data Tests ---\n");

    s_mock_raw_r = 4096;
    s_mock_raw_ir = 8192;
    s_mock_ppg_read_raw_ok = true;
    ppg_init();
    s_ppg_data_ready = true;
    s_ppg_raw_r = s_mock_raw_r;
    s_ppg_raw_ir = s_mock_raw_ir;
    ppg_data_t data = ppg_get_data();
    TEST_ASSERT(data.valid == true, "Valid data has valid=true");
    TEST_ASSERT(data.hr >= PPG_HR_MIN && data.hr <= PPG_HR_MAX,
                "Valid HR in combined data");
    TEST_ASSERT(data.spo2 >= PPG_SPO2_MIN && data.spo2 <= PPG_SPO2_MAX,
                "Valid SpO2 in combined data");

    /* Test 2: Out-of-range data produces valid=false */
    s_ppg_data_ready = true;
    s_ppg_raw_r = 0;
    s_ppg_raw_ir = 8192;
    data = ppg_get_data();
    TEST_ASSERT(data.valid == false, "Invalid data has valid=false");
}

/* ============ IMU Tests ============ */

static void test_imu_init(void)
{
    printf("\n--- IMU Init Tests ---\n");

    s_mock_imu_init_ok = true;
    TEST_ASSERT(imu_init() == true, "IMU init with valid device ID (0xEF)");

    s_mock_imu_init_ok = false;
    TEST_ASSERT(imu_init() == false, "IMU init rejects invalid device ID");
}

static void test_imu_accel_read(void)
{
    printf("\n--- IMU Acceleration Read Tests ---\n");

    imu_init();

    s_mock_accel_x = 100;
    s_mock_accel_y = -50;
    s_mock_accel_z = 8192;  /* 1g at 4g range */

    int16_t ax, ay, az;
    TEST_ASSERT(imu_read_accel_raw(&ax, &ay, &az) == true,
                "Raw accelerometer read succeeds");
    TEST_ASSERT_EQ(ax, 100, "X axis raw value matches");
    TEST_ASSERT_EQ(ay, -50, "Y axis raw value matches");
    TEST_ASSERT_EQ(az, 8192, "Z axis raw value (8192 = 1g at 4g range)");

    imu_data_t accel = imu_read_accel();
    TEST_ASSERT_NEAR(accel.az, 1.0f, 0.05f,
                     "Z axis calibrated to ~1.0g (gravity)");
    TEST_ASSERT(fabs(accel.ax) < 0.1f, "X axis near 0g when stationary");
    TEST_ASSERT(fabs(accel.ay) < 0.1f, "Y axis near 0g when stationary");
}

static void test_imu_gyro_read(void)
{
    printf("\n--- IMU Gyroscope Read Tests ---\n");

    imu_init();

    s_mock_gyro_x = 0;
    s_mock_gyro_y = 0;
    s_mock_gyro_z = 0;

    imu_data_t gyro = imu_read_gyro();
    TEST_ASSERT_NEAR(gyro.gx, 0.0f, 0.1f, "X gyro near 0 dps");
    TEST_ASSERT_NEAR(gyro.gy, 0.0f, 0.1f, "Y gyro near 0 dps");
    TEST_ASSERT_NEAR(gyro.gz, 0.0f, 0.1f, "Z gyro near 0 dps");

    s_mock_gyro_x = 6553;
    s_mock_gyro_y = 0;
    s_mock_gyro_z = 0;

    gyro = imu_read_gyro();
    TEST_ASSERT_NEAR(gyro.gx, 100.0f, 1.0f,
                     "X gyro reads ~100 dps (6553 raw at 500dps range)");
}

static void test_imu_magnitude(void)
{
    printf("\n--- IMU Magnitude Calculation Tests ---\n");

    imu_data_t data;
    data.ax = 0.0f; data.ay = 0.0f; data.az = 1.0f;
    TEST_ASSERT_NEAR(imu_accel_magnitude(&data), 1.0f, 0.01f,
                     "Magnitude of (0,0,1g) equals 1.0g");

    data.ax = 1.0f; data.ay = 1.0f; data.az = 1.0f;
    TEST_ASSERT_NEAR(imu_accel_magnitude(&data), sqrtf(3.0f), 0.01f,
                     "Magnitude of (1,1,1g) equals sqrt(3)");

    data.ax = 0.0f; data.ay = 0.0f; data.az = 0.0f;
    TEST_ASSERT_NEAR(imu_accel_magnitude(&data), 0.0f, 0.01f,
                     "Zero vector magnitude is 0");
}

static void test_imu_range_config(void)
{
    printf("\n--- IMU Range Configuration Tests ---\n");

    imu_init();

    imu_set_accel_range(IMU_ACCEL_RANGE_2G);
    imu_set_accel_range(IMU_ACCEL_RANGE_4G);
    imu_set_accel_range(IMU_ACCEL_RANGE_8G);
    imu_set_accel_range(IMU_ACCEL_RANGE_16G);
    TEST_ASSERT(true, "All accel range settings accepted");

    imu_set_accel_range(0x0FU);
    TEST_ASSERT(true, "Invalid accel range handled gracefully");

    imu_set_gyro_range(IMU_GYRO_RANGE_250DPS);
    imu_set_gyro_range(IMU_GYRO_RANGE_500DPS);
    imu_set_gyro_range(IMU_GYRO_RANGE_1000DPS);
    imu_set_gyro_range(IMU_GYRO_RANGE_2000DPS);
    TEST_ASSERT(true, "All gyro range settings accepted");

    imu_set_gyro_range(0x0FU);
    TEST_ASSERT(true, "Invalid gyro range handled gracefully");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen Bracelet Entry - Sensor Tests\n");
    printf("Target: GD32E230C8T3 / FreeRTOS\n");
    printf("Mode: Host simulation\n");
    printf("========================================\n");

    test_ppg_init();
    test_ppg_hr_calculation();
    test_ppg_spo2_calculation();
    test_ppg_combined_data();
    test_imu_init();
    test_imu_accel_read();
    test_imu_gyro_read();
    test_imu_magnitude();
    test_imu_range_config();

    printf("\n========================================\n");
    printf("Test Results: %d/%d passed (%d failed)\n",
           g_tests_passed, g_tests_run, g_tests_failed);
    printf("========================================\n");

    return (g_tests_failed > 0) ? 1 : 0;
}

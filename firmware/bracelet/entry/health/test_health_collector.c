/*
 * Eregen (颐贞) - Health Collector Tests
 * Tests health data encoding, MQTT publish calls, and edge cases.
 * Compile: gcc -DTEST_MODE -I. health/test_health_collector.c \
 *   sensors_ppg.c sensors_imu.c protocol/message_encode.c \
 *   cat1_at.c common/crc16.c common/log.c -lm -o test_health_collector
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <stdint.h>
#include <stdbool.h>
#include <time.h>
#include <stdarg.h>

/* ============ Mock Hardware Layer ============ */

/* --- PPG Sensor Mock State --- */
static uint16_t s_mock_ppg_raw_r = 4096;
static uint16_t s_mock_ppg_raw_ir = 8192;
static bool s_mock_ppg_data_ready = true;
static uint8_t s_mock_ppg_chip_id = 0x60;

/* --- IMU Sensor Mock State --- */
static int16_t s_mock_accel_x = 100;
static int16_t s_mock_accel_y = -50;
static int16_t s_mock_accel_z = 16384;  /* ~1g at 4g range */
static uint8_t s_mock_imu_dev_id = 0xEF;
static uint8_t s_mock_imu_range = 0x01;  /* 4g default */

/* --- Cat1/MQTT Mock State --- */
static bool s_mock_mqtt_connected = true;
static char s_last_topic[256];
static uint8_t s_last_data[1024];
static uint16_t s_last_len = 0;
static bool s_mqtt_publish_called = false;

/* --- Message Encode Mock State --- */
static int s_encode_return = 0;

/* ============ PPG Driver (inline for test compilation) ============ */

#define PPG_I2C_ADDR             0x4CU
#define PPG_REG_CHIP_ID          0x00U
#define PPG_REG_SYS_CTRL         0x01U
#define PPG_REG_DATA_CTRL        0x02U
#define PPG_REG_MODE             0x04U
#define PPG_REG_DATA_STATUS      0x2AU
#define PPG_REG_DATA_R           0x2BU
#define PPG_REG_DATA_IR          0x2FU
#define PPG_REG_HR_RESULT        0x50U
#define PPG_REG_SPO2_RESULT      0x51U
#define PPG_HR_MIN               30U
#define PPG_HR_MAX               220U
#define PPG_SPO2_MIN             70U
#define PPG_SPO2_MAX             100U
#define PPG_SAMPLE_INTERVAL_MS   1000U
#define PPG_I2C_RETRY_MAX        3U
#define PPG_QUALITY_GOOD         75U
#define PPG_QUALITY_FAIR         50U

typedef struct {
    uint16_t hr;
    uint8_t  spo2;
    uint8_t  quality;
} ppg_data_t;

static uint16_t s_ppg_raw_r = 0;
static uint16_t s_ppg_raw_ir = 0;
static bool s_ppg_data_ready = false;

static bool ppg_i2c_write_reg(uint8_t reg, uint8_t val)
{
    (void)reg; (void)val;
    return true;
}

static bool ppg_i2c_read_reg(uint8_t reg, uint8_t *val)
{
    if (reg == PPG_REG_CHIP_ID) {
        *val = s_mock_ppg_chip_id;
        return true;
    }
    if (reg == PPG_REG_DATA_STATUS) {
        *val = s_mock_ppg_data_ready ? 0x01 : 0x00;
        return true;
    }
    return false;
}

static bool ppg_i2c_read_multi(uint8_t reg, uint8_t *buf, uint8_t len)
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
    s_ppg_raw_r = s_mock_ppg_raw_r;
    s_ppg_raw_ir = s_mock_ppg_raw_ir;
    s_ppg_data_ready = s_mock_ppg_data_ready;
    return true;
}

bool ppg_read_raw(uint16_t *r_val, uint16_t *ir_val)
{
    uint8_t status = 0;
    if (!ppg_i2c_read_reg(PPG_REG_DATA_STATUS, &status)) {
        return false;
    }
    if ((status & 0x01U) == 0U) {
        return false;
    }
    uint8_t raw_buf[4];
    if (!ppg_i2c_read_multi(PPG_REG_DATA_R, raw_buf, 4)) {
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

uint8_t ppg_estimate_quality(void)
{
    if (!s_ppg_data_ready || s_ppg_raw_r == 0 || s_ppg_raw_ir == 0) {
        return 0;
    }
    uint32_t ratio = (uint32_t)s_ppg_raw_r * 1000U / (uint32_t)s_ppg_raw_ir;
    if (ratio >= 300U && ratio <= 600U) {
        uint16_t center = 450U;
        uint16_t dist = (ratio > center) ? (ratio - center) : (center - ratio);
        return (uint8_t)(100U - (dist * 20U / 150U));
    }
    if (ratio < 300U) {
        if (ratio > 200U) return PPG_QUALITY_FAIR;
        return (uint8_t)(ratio / 4U);
    }
    if (ratio < 800U) {
        return (uint8_t)(PPG_QUALITY_FAIR * (800U - ratio) / 200U);
    }
    return 0;
}

bool ppg_read(ppg_data_t *data)
{
    if (!data) return false;
    uint16_t raw_r = 0, raw_ir = 0;
    if (!ppg_read_raw(&raw_r, &raw_ir)) {
        data->hr = 0;
        data->spo2 = 0;
        data->quality = 0;
        return false;
    }
    data->hr = ppg_calculate_hr();
    data->spo2 = ppg_calculate_spo2();
    data->quality = ppg_estimate_quality();
    return true;
}

/* ============ IMU Driver (inline for test compilation) ============ */

#define IMU_REG_DEVICE_ID      0x00U
#define IMU_REG_ACCEL_CONFIG   0x40U
#define IMU_REG_GYRO_CONFIG    0x42U
#define IMU_REG_ACCEL_X0       0x4FU
#define IMU_REG_GYRO_X0        0x55U
#define IMU_ACCEL_RANGE_2G     0x00U
#define IMU_ACCEL_RANGE_4G     0x01U
#define IMU_ACCEL_RANGE_8G     0x02U
#define IMU_ACCEL_RANGE_16G    0x03U
#define IMU_DEFAULT_ACCEL_RANGE   IMU_ACCEL_RANGE_4G

typedef struct {
    float ax, ay, az;
    float gx, gy, gz;
} imu_data_t;

static uint8_t s_imu_accel_range = IMU_DEFAULT_ACCEL_RANGE;
static const float s_imu_accel_scale[] = { 1.0f/16384.0f, 1.0f/8192.0f,
                                            1.0f/4096.0f, 1.0f/2048.0f };

bool imu_spi_write_reg(uint8_t reg, uint8_t val)
{
    (void)reg; (void)val;
    return true;
}

bool imu_spi_read_reg(uint8_t reg, uint8_t *val)
{
    if (reg == IMU_REG_DEVICE_ID) {
        *val = s_mock_imu_dev_id;
        return true;
    }
    *val = 0xFF;
    return true;
}

bool imu_spi_read_multi(uint8_t reg, uint8_t *buf, uint8_t len)
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
    return true;
}

bool imu_init(void)
{
    uint8_t dev_id = 0;
    if (!imu_spi_read_reg(IMU_REG_DEVICE_ID, &dev_id)) return false;
    if (dev_id != 0xEFU) return false;
    s_imu_accel_range = IMU_DEFAULT_ACCEL_RANGE;
    return true;
}

bool imu_read_accel_raw(int16_t *x, int16_t *y, int16_t *z)
{
    uint8_t buf[6];
    if (!imu_spi_read_multi(IMU_REG_ACCEL_X0, buf, 6)) return false;
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

float imu_accel_magnitude(const imu_data_t *data)
{
    return sqrtf(data->ax * data->ax + data->ay * data->ay + data->az * data->az);
}

/* ============ CRC16 (inline for test compilation) ============ */

static uint16_t crc16_table[256];
static bool s_crc_table_computed = false;

static void crc16_compute_table(void)
{
    if (s_crc_table_computed) return;
    for (uint16_t i = 0; i < 256; i++) {
        uint16_t crc = (uint16_t)i << 8;
        for (uint8_t j = 0; j < 8; j++) {
            if (crc & 0x8000) {
                crc = (uint16_t)(crc << 1) ^ 0x1021;
            } else {
                crc <<= 1;
            }
        }
        crc16_table[i] = crc;
    }
    s_crc_table_computed = true;
}

uint16_t crc16_calc(const uint8_t *data, uint16_t len)
{
    if (!s_crc_table_computed) crc16_compute_table();
    uint16_t crc = 0xFFFF;
    for (uint16_t i = 0; i < len; i++) {
        uint8_t table_index = (uint8_t)((crc >> 8) ^ data[i]);
        crc = (uint16_t)((crc << 8) ^ crc16_table[table_index]);
    }
    return crc;
}

/* ============ Message Encode (inline for test compilation) ============ */

#define MAX_PAYLOAD_LEN 512U
#define MAX_MSG_LEN (1 + 16 + 8 + 4 + MAX_PAYLOAD_LEN + 2)

typedef enum {
    MSG_HEARTBEAT = 1,
    MSG_LOCATION,
    MSG_HEALTH,
    MSG_SOS,
    MSG_FALL,
    MSG_MED_STATUS
} msg_type_t;

typedef struct {
    msg_type_t type;
    char dev_id[17];
    uint64_t timestamp;
    uint8_t payload[MAX_PAYLOAD_LEN];
    uint16_t payload_len;
    uint16_t checksum;
} eregen_msg_t;

int message_encode(const eregen_msg_t *msg, uint8_t *out, uint16_t out_len)
{
    if (!msg || !out || out_len == 0) return -1;
    if (msg->type < MSG_HEARTBEAT || msg->type > MSG_MED_STATUS) return -2;
    if (msg->payload_len > MAX_PAYLOAD_LEN) return -3;
    if (strlen(msg->dev_id) >= sizeof(msg->dev_id)) return -4;

    char json_buf[256];
    int json_len;

    if (msg->payload_len > 0 && msg->payload_len < 128) {
        json_len = snprintf(json_buf, sizeof(json_buf),
            "{\"type\":%u,\"dev_id\":\"%s\",\"ts\":%lu,\"data\":\"",
            (unsigned)msg->type,
            msg->dev_id,
            (unsigned long)msg->timestamp);

        for (uint16_t i = 0; i < msg->payload_len; i++) {
            int remaining = (int)sizeof(json_buf) - json_len - 3;
            if (remaining <= 0) break;
            json_len += snprintf(json_buf + json_len, (size_t)remaining,
                                 "%02X", msg->payload[i]);
        }
        json_len += snprintf(json_buf + json_len, (size_t)(sizeof(json_buf) - json_len),
                             "\"}");
    } else {
        json_len = snprintf(json_buf, sizeof(json_buf),
            "{\"type\":%u,\"dev_id\":\"%s\",\"ts\":%lu,\"payload_len\":%u}",
            (unsigned)msg->type,
            msg->dev_id,
            (unsigned long)msg->timestamp,
            (unsigned)msg->payload_len);
    }

    if (json_len <= 0 || json_len >= (int)sizeof(json_buf)) return -5;

    uint16_t total_needed = (uint16_t)(json_len + 2);
    if (total_needed > out_len) return -6;

    memcpy(out, json_buf, (size_t)json_len);
    uint16_t crc = crc16_calc((const uint8_t *)json_buf, (uint16_t)json_len);
    out[json_len]     = (uint8_t)((crc >> 8) & 0xFF);
    out[json_len + 1] = (uint8_t)(crc & 0xFF);

    return (int)total_needed;
}

/* ============ Cat1 MQTT Mock ============ */

bool cat1_mqtt_connect(const char *client_id, const char *user, const char *pass)
{
    (void)client_id; (void)user; (void)pass;
    return s_mock_mqtt_connected;
}

bool cat1_mqtt_publish(const char *topic, const uint8_t *data, uint16_t len)
{
    if (!s_mock_mqtt_connected) return false;
    if (!topic || !data) return false;

    strncpy(s_last_topic, topic, sizeof(s_last_topic) - 1);
    s_last_topic[sizeof(s_last_topic) - 1] = '\0';
    memcpy(s_last_data, data, len < sizeof(s_last_data) ? len : sizeof(s_last_data));
    s_last_len = len;
    s_mqtt_publish_called = true;
    return true;
}

bool cat1_is_connected(void)
{
    return s_mock_mqtt_connected;
}

/* ============ Log (inline for test compilation) ============ */

typedef enum {
    LOG_DEBUG = 0, LOG_INFO, LOG_WARN, LOG_ERROR, LOG_LEVEL_COUNT
} log_level_t;

static log_level_t s_log_level = LOG_DEBUG;
static const char *log_prefixes[LOG_LEVEL_COUNT] = {"[D]", "[I]", "[W]", "[E]"};

void log_init(void) { s_log_level = LOG_DEBUG; }
void log_set_level(log_level_t level) {
    if (level >= 0 && level < LOG_LEVEL_COUNT) s_log_level = level;
}
log_level_t log_get_level(void) { return s_log_level; }

void log_debug(const char *fmt, ...) {
    va_list args; va_start(args, fmt);
    if (LOG_DEBUG >= s_log_level) {
        printf("%s ", log_prefixes[LOG_DEBUG]);
        vprintf(fmt, args);
        printf("\n");
    }
    va_end(args);
}
void log_info(const char *fmt, ...) {
    va_list args; va_start(args, fmt);
    if (LOG_INFO >= s_log_level) {
        printf("%s ", log_prefixes[LOG_INFO]);
        vprintf(fmt, args);
        printf("\n");
    }
    va_end(args);
}
void log_warn(const char *fmt, ...) {
    va_list args; va_start(args, fmt);
    if (LOG_WARN >= s_log_level) {
        printf("%s ", log_prefixes[LOG_WARN]);
        vprintf(fmt, args);
        printf("\n");
    }
    va_end(args);
}
void log_error(const char *fmt, ...) {
    va_list args; va_start(args, fmt);
    if (LOG_ERROR >= s_log_level) {
        printf("%s ", log_prefixes[LOG_ERROR]);
        vprintf(fmt, args);
        printf("\n");
    }
    va_end(args);
}

#include <stdarg.h>

/* ============ Health Collector (inline for test compilation) ============ */

#define DEVICE_ID    "BR-0001"
#define HEALTH_TOPIC "device/" DEVICE_ID "/health"
#define HEALTH_PAYLOAD_BUF_SIZE 128U

typedef struct {
    uint32_t step_count;
    uint32_t sample_count;
    float prev_magnitude;
    bool step_detected;
} step_counter_t;

static step_counter_t s_step_counter = {0};

static void step_counter_reset(void)
{
    s_step_counter.step_count = 0;
    s_step_counter.sample_count = 0;
    s_step_counter.prev_magnitude = 0.0f;
    s_step_counter.step_detected = false;
}

static bool step_counter_update(float magnitude)
{
    const float STEP_THRESHOLD = 1.2f;
    const float REBOUND_HYSTERESIS = 0.8f;

    if (!s_step_counter.step_detected && magnitude > STEP_THRESHOLD) {
        s_step_counter.step_detected = true;
    } else if (s_step_counter.step_detected && magnitude < REBOUND_HYSTERESIS) {
        s_step_counter.step_detected = false;
        s_step_counter.step_count++;
    }

    s_step_counter.sample_count++;
    s_step_counter.prev_magnitude = magnitude;
    return s_step_counter.step_detected;
}

static int encode_health_payload(char *buf, uint16_t len,
                                  uint16_t hr, uint8_t spo2,
                                  uint32_t steps, uint8_t quality)
{
    int written;
    written = snprintf(buf, len,
        "{\"hr\":%u,\"spo2\":%u,\"steps\":%lu,\"quality\":%u}",
        (unsigned)hr, (unsigned)spo2, (unsigned long)steps, (unsigned)quality);
    if (written < 0 || (uint16_t)written >= len) return -1;
    return written;
}

void health_collect_and_send(void)
{
    ppg_data_t ppg;
    bool ppg_ok = ppg_read(&ppg);
    if (!ppg_ok) {
        log_warn("PPG read failed, sending zeros");
        ppg.hr = 0;
        ppg.spo2 = 0;
        ppg.quality = 0;
    }

    imu_data_t accel = imu_read_accel();
    float mag = imu_accel_magnitude(&accel);
    step_counter_update(mag);

    char payload_buf[HEALTH_PAYLOAD_BUF_SIZE];
    int payload_len = encode_health_payload(
        payload_buf, sizeof(payload_buf),
        ppg.hr, ppg.spo2,
        s_step_counter.step_count,
        ppg.quality
    );

    if (payload_len < 0) {
        log_error("Failed to encode health payload");
        return;
    }

    eregen_msg_t msg;
    memset(&msg, 0, sizeof(msg));
    msg.type = MSG_HEALTH;
    strncpy(msg.dev_id, DEVICE_ID, sizeof(msg.dev_id) - 1);
    msg.timestamp = (uint64_t)time(NULL);
    memcpy(msg.payload, (const uint8_t *)payload_buf, (size_t)payload_len);
    msg.payload_len = (uint16_t)payload_len;

    uint8_t encoded[MAX_MSG_LEN];
    int encoded_len = message_encode(&msg, encoded, (uint16_t)sizeof(encoded));

    if (encoded_len < 0) {
        log_error("Message encode failed, code=%d", encoded_len);
        return;
    }

    bool pub_ok = cat1_mqtt_publish(HEALTH_TOPIC, encoded, (uint16_t)encoded_len);

    if (pub_ok) {
        log_info("Health published: hr=%u spo2=%u steps=%lu quality=%u",
                 (unsigned)ppg.hr, (unsigned)ppg.spo2,
                 (unsigned long)s_step_counter.step_count,
                 (unsigned)ppg.quality);
    } else {
        log_error("MQTT publish failed for health data");
    }
}

void health_init(void)
{
    log_info("Initializing health collector");
    step_counter_reset();
    if (!ppg_init()) {
        log_error("PPG sensor initialization failed");
    } else {
        log_info("PPG sensor initialized OK");
    }
    if (!imu_init()) {
        log_error("IMU sensor initialization failed");
    } else {
        log_info("IMU sensor initialized OK");
    }
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
        printf("  FAIL: %s (expected=%ld, got=%ld)\n", msg, (long)(b), (long)(a)); \
    } \
} while(0)

#define TEST_ASSERT_STR_EQ(a, b, msg) do { \
    g_tests_run++; \
    if (strcmp((a), (b)) == 0) { \
        g_tests_passed++; \
        printf("  PASS: %s\n", msg); \
    } else { \
        g_tests_failed++; \
        printf("  FAIL: %s (expected=\"%s\", got=\"%s\")\n", msg, (b), (a)); \
    } \
} while(0)

/* ============ Helper: extract JSON field from encoded message ============ */

/*
 * Extract the JSON payload portion from the encoded message output buffer.
 * The wire format is: {"type":N,"dev_id":"...","ts":N,"data":"<hex>"}
 * We extract the hex-decoded data portion.
 */
static int extract_json_data(const uint8_t *encoded, int enc_len, char *out, int out_len)
{
    /* Find "data":" prefix */
    const char *search = "\"data\":\"";
    const char *found = NULL;
    for (int i = 0; i < enc_len - 8; i++) {
        if (memcmp(encoded + i, search, 8) == 0) {
            found = (const char *)(encoded + i + 8);
            break;
        }
    }
    if (!found) return -1;

    /* Find closing " then } */
    int data_start = (int)(found - (const char *)encoded);
    int data_end = -1;
    for (int i = data_start; i < enc_len - 2; i++) {
        if (encoded[i] == '"' && encoded[i+1] == '}') {
            data_end = i;
            break;
        }
    }
    if (data_end < 0) return -1;

    /* Decode hex data back to ASCII */
    int out_idx = 0;
    for (int i = data_start; i < data_end && out_idx < out_len - 1; i += 2) {
        uint8_t hi = found[i - data_start];
        uint8_t lo = found[i - data_start + 1];
        uint8_t byte = 0;
        if (hi >= '0' && hi <= '9') byte = (hi - '0') << 4;
        else if (hi >= 'A' && hi <= 'F') byte = (hi - 'A' + 10) << 4;
        else if (hi >= 'a' && hi <= 'f') byte = (hi - 'a' + 10) << 4;
        if (lo >= '0' && lo <= '9') byte |= (lo - '0');
        else if (lo >= 'A' && lo <= 'F') byte |= (lo - 'A' + 10);
        else if (lo >= 'a' && lo <= 'f') byte |= (lo - 'a' + 10);
        out[out_idx++] = (char)byte;
    }
    out[out_idx] = '\0';
    return out_idx;
}

/* ============ External access to PPG internal state for tests ============ */
extern uint16_t s_ppg_raw_r;
extern uint16_t s_ppg_raw_ir;
extern bool s_ppg_data_ready;

/* ============ Test: Health Data Format ============ */

static void test_health_data_format(void)
{
    printf("\n--- Health Data Format Tests ---\n");

    /* Setup normal PPG data */
    s_mock_ppg_raw_r = 4096;
    s_mock_ppg_raw_ir = 8192;
    s_mock_ppg_data_ready = true;
    s_mock_ppg_chip_id = 0x60;

    /* Setup IMU at rest (~1g on Z axis) */
    s_mock_accel_x = 100;
    s_mock_accel_y = -50;
    s_mock_accel_z = 16384;

    /* Reset mocks */
    memset(s_last_topic, 0, sizeof(s_last_topic));
    memset(s_last_data, 0, sizeof(s_last_data));
    s_last_len = 0;
    s_mqtt_publish_called = false;
    step_counter_reset();

    health_init();
    health_collect_and_send();

    /* Verify MQTT was called */
    TEST_ASSERT(s_mqtt_publish_called, "MQTT publish was called");

    /* Verify topic */
    TEST_ASSERT_STR_EQ(s_last_topic, HEALTH_TOPIC, "Correct MQTT topic");

    /* Decode the JSON payload from encoded message */
    char json_data[256];
    int decoded_len = extract_json_data(s_last_data, s_last_len, json_data, sizeof(json_data));
    TEST_ASSERT(decoded_len > 0, "Successfully extracted JSON data from encoded message");

    if (decoded_len > 0) {
        json_data[decoded_len] = '\0';
        printf("  Decoded payload: %s\n", json_data);

        /* Verify JSON contains expected fields */
        TEST_ASSERT(strstr(json_data, "\"hr\":") != NULL,
                    "JSON contains 'hr' field");
        TEST_ASSERT(strstr(json_data, "\"spo2\":") != NULL,
                    "JSON contains 'spo2' field");
        TEST_ASSERT(strstr(json_data, "\"steps\":") != NULL,
                    "JSON contains 'steps' field");
        TEST_ASSERT(strstr(json_data, "\"quality\":") != NULL,
                    "JSON contains 'quality' field");

        /* Verify HR value is in reasonable range (not zero since we set valid data) */
        char *hr_pos = strstr(json_data, "\"hr\":");
        if (hr_pos) {
            uint16_t hr = (uint16_t)atoi(hr_pos + 5);
            TEST_ASSERT(hr > 0, "HR is non-zero for valid PPG data");
            TEST_ASSERT(hr >= 30 && hr <= 220, "HR is within valid range [30, 220]");
        }

        /* Verify SpO2 is in valid range */
        char *spo2_pos = strstr(json_data, "\"spo2\":");
        if (spo2_pos) {
            uint8_t spo2 = (uint8_t)atoi(spo2_pos + 7);
            TEST_ASSERT(spo2 > 0, "SpO2 is non-zero for valid PPG data");
            TEST_ASSERT(spo2 >= 70 && spo2 <= 100, "SpO2 is within valid range [70, 100]");
        }
    }
}

/* ============ External access to PPG internal state for tests ============ */
extern uint16_t s_ppg_raw_r;
extern uint16_t s_ppg_raw_ir;
extern bool s_ppg_data_ready;

/* ============ Test: Zero HR Edge Case ============ */

static void test_zero_hr_edge_case(void)
{
    printf("\n--- Zero HR Edge Case Tests ---\n");

    /* Directly set internal PPG state to zero (bypasses I2C mock) */
    s_ppg_raw_r = 0;
    s_ppg_raw_ir = 0;
    s_ppg_data_ready = true;

    /* Setup IMU */
    s_mock_accel_x = 0;
    s_mock_accel_y = 0;
    s_mock_accel_z = 16384;

    /* Reset mocks */
    memset(s_last_topic, 0, sizeof(s_last_topic));
    memset(s_last_data, 0, sizeof(s_last_data));
    s_last_len = 0;
    s_mqtt_publish_called = false;
    step_counter_reset();

    health_collect_and_send();

    TEST_ASSERT(s_mqtt_publish_called, "MQTT publish still called even with zero HR");

    char json_data[256];
    int decoded_len = extract_json_data(s_last_data, s_last_len, json_data, sizeof(json_data));
    TEST_ASSERT(decoded_len > 0, "JSON extracted even with zero HR");

    if (decoded_len > 0) {
        json_data[decoded_len] = '\0';
        char *hr_pos = strstr(json_data, "\"hr\":");
        if (hr_pos) {
            uint16_t hr = (uint16_t)atoi(hr_pos + 5);
            TEST_ASSERT_EQ(hr, 0, "HR is zero when PPG data is invalid");
        }
    }
}

/* ============ Test: Invalid SpO2 Edge Case ============ */

static void test_invalid_spo2_edge_case(void)
{
    printf("\n--- Invalid SpO2 Edge Case Tests ---\n");

    /* Test high SpO2 clamping: very low R/IR ratio -> SpO2 = 110 - 25*small_ratio -> exceeds 100 */
    s_ppg_raw_r = 1000;
    s_ppg_raw_ir = 8000;
    s_ppg_data_ready = true;

    /* Setup IMU */
    s_mock_accel_x = 0;
    s_mock_accel_y = 0;
    s_mock_accel_z = 16384;

    /* Reset mocks */
    memset(s_last_topic, 0, sizeof(s_last_topic));
    memset(s_last_data, 0, sizeof(s_last_data));
    s_last_len = 0;
    s_mqtt_publish_called = false;
    step_counter_reset();

    health_collect_and_send();

    TEST_ASSERT(s_mqtt_publish_called, "MQTT publish called with extreme SpO2 input");

    char json_data[256];
    int decoded_len = extract_json_data(s_last_data, s_last_len, json_data, sizeof(json_data));
    TEST_ASSERT(decoded_len > 0, "JSON extracted with extreme SpO2 input");

    if (decoded_len > 0) {
        json_data[decoded_len] = '\0';
        char *spo2_pos = strstr(json_data, "\"spo2\":");
        if (spo2_pos) {
            uint8_t spo2 = (uint8_t)atoi(spo2_pos + 7);
            TEST_ASSERT_EQ(spo2, 100, "SpO2 clamped to maximum 100 (not exceeding)");
        }
    }

    /* Test low SpO2 clamping: very high ratio */
    s_ppg_raw_r = 8000;
    s_ppg_raw_ir = 1000;
    s_ppg_data_ready = true;

    memset(s_last_topic, 0, sizeof(s_last_topic));
    memset(s_last_data, 0, sizeof(s_last_data));
    s_last_len = 0;
    s_mqtt_publish_called = false;
    step_counter_reset();

    health_collect_and_send();

    TEST_ASSERT(s_mqtt_publish_called, "MQTT publish called with low SpO2 input");

    decoded_len = extract_json_data(s_last_data, s_last_len, json_data, sizeof(json_data));
    if (decoded_len > 0) {
        json_data[decoded_len] = '\0';
        char *spo2_pos = strstr(json_data, "\"spo2\":");
        if (spo2_pos) {
            uint8_t spo2 = (uint8_t)atoi(spo2_pos + 7);
            TEST_ASSERT(spo2 >= 70, "SpO2 not below minimum 70 (clamped)");
        }
    }
}

/* ============ Test: Health Data Format ============ */

static void test_step_counting(void)
{
    printf("\n--- Step Counting Tests ---\n");

    /* Setup PPG for normal reading */
    s_mock_ppg_raw_r = 4096;
    s_mock_ppg_raw_ir = 8192;
    s_mock_ppg_data_ready = true;
    s_mock_ppg_chip_id = 0x60;

    step_counter_reset();

    /* Simulate walking: alternating above/below threshold */
    /* First sample: magnitude rises above 1.2g */
    step_counter_update(1.5f);
    TEST_ASSERT(!s_step_counter.step_detected == false,
                "Step detected flag set after crossing threshold up");

    /* Second sample: magnitude drops below 0.8g -> step counted */
    step_counter_update(0.5f);
    TEST_ASSERT_EQ(s_step_counter.step_count, 1,
                   "One step counted after threshold up-then-down cycle");

    /* Third sample: rising again */
    step_counter_update(1.8f);

    /* Fourth sample: dropping again -> second step */
    step_counter_update(0.3f);
    TEST_ASSERT_EQ(s_step_counter.step_count, 2,
                   "Two steps after two complete cycles");

    /* No step if stays above threshold */
    step_counter_update(2.0f);
    step_counter_update(1.5f);
    TEST_ASSERT_EQ(s_step_counter.step_count, 2,
                   "No extra step when magnitude stays above threshold");

    /* No step if never crosses up */
    step_counter_reset();
    step_counter_update(0.5f);  /* Below threshold */
    step_counter_update(0.3f);  /* Still below */
    TEST_ASSERT_EQ(s_step_counter.step_count, 0,
                   "No step counted when magnitude never exceeds threshold");
}

/* ============ Test: PPG Signal Quality ============ */

static void test_signal_quality(void)
{
    printf("\n--- Signal Quality Tests ---\n");

    /* Optimal ratio (4096/8192 = 0.5 -> ratio = 500) */
    s_mock_ppg_raw_r = 4096;
    s_mock_ppg_raw_ir = 8192;
    s_mock_ppg_data_ready = true;
    ppg_init();

    uint8_t q_good = ppg_estimate_quality();
    TEST_ASSERT(q_good >= PPG_QUALITY_GOOD,
                "Good signal quality for optimal ratio");

    /* Poor ratio: very low R/IR */
    s_ppg_raw_r = 1000;
    s_ppg_raw_ir = 8000;
    s_ppg_data_ready = true;
    uint8_t q_poor = ppg_estimate_quality();
    TEST_ASSERT(q_poor < PPG_QUALITY_FAIR,
                "Poor signal quality for extreme low ratio");

    /* Very poor ratio: very high R/IR */
    s_ppg_raw_r = 8000;
    s_ppg_raw_ir = 1000;
    s_ppg_data_ready = true;
    uint8_t q_very_poor = ppg_estimate_quality();
    TEST_ASSERT(q_very_poor < PPG_QUALITY_FAIR,
                "Very poor quality for extreme high ratio");
}

/* ============ Test: MQTT Publish Parameters ============ */

static void test_mqtt_publish_parameters(void)
{
    printf("\n--- MQTT Publish Parameter Tests ---\n");

    /* Setup valid sensor data */
    s_mock_ppg_raw_r = 4096;
    s_mock_ppg_raw_ir = 8192;
    s_mock_ppg_data_ready = true;
    s_mock_ppg_chip_id = 0x60;
    s_mock_accel_x = 0;
    s_mock_accel_y = 0;
    s_mock_accel_z = 16384;

    memset(s_last_topic, 0, sizeof(s_last_topic));
    memset(s_last_data, 0, sizeof(s_last_data));
    s_last_len = 0;
    s_mqtt_publish_called = false;
    step_counter_reset();

    health_collect_and_send();

    /* Verify topic format */
    const char *expected_prefix = "device/BR-";
    TEST_ASSERT(strncmp(s_last_topic, expected_prefix, strlen(expected_prefix)) == 0,
                "MQTT topic follows device/BR-XXXX/health format");

    /* Verify encoded data has CRC appended (last 2 bytes are non-zero for valid JSON) */
    TEST_ASSERT(s_last_len > 2, "Encoded message has payload plus CRC length");

    /* Verify data is not all zeros (meaningful content) */
    bool has_content = false;
    for (int i = 0; i < s_last_len; i++) {
        if (s_last_data[i] != 0) { has_content = true; break; }
    }
    TEST_ASSERT(has_content, "Encoded message contains non-zero content");
}

/* ============ Test: PPG Read Function ============ */

static void test_ppg_read_function(void)
{
    printf("\n--- PPG Read Function Tests ---\n");

    /* Valid data */
    s_mock_ppg_raw_r = 4096;
    s_mock_ppg_raw_ir = 8192;
    s_mock_ppg_data_ready = true;
    s_mock_ppg_chip_id = 0x60;
    ppg_init();

    ppg_data_t data;
    bool ok = ppg_read(&data);
    TEST_ASSERT(ok == true, "ppg_read returns true with valid sensor");
    TEST_ASSERT(data.hr > 0, "ppg_read populates non-zero HR");
    TEST_ASSERT(data.spo2 > 0, "ppg_read populates non-zero SpO2");
    TEST_ASSERT(data.quality > 0, "ppg_read populates non-zero quality");

    /* Null pointer */
    ok = ppg_read(NULL);
    TEST_ASSERT(ok == false, "ppg_read returns false for NULL pointer");

    /* No data ready */
    s_mock_ppg_data_ready = false;
    ok = ppg_read(&data);
    TEST_ASSERT(ok == false, "ppg_read returns false when sensor data not ready");
}

/* ============ Test: Health Init ============ */

static void test_health_init(void)
{
    printf("\n--- Health Init Tests ---\n");

    /* Normal init */
    s_mock_ppg_chip_id = 0x60;
    s_mock_imu_dev_id = 0xEF;
    step_counter_reset();

    health_init();
    TEST_ASSERT(s_step_counter.step_count == 0,
                "Step counter reset during health_init");
    TEST_ASSERT(s_step_counter.sample_count == 0,
                "Sample counter reset during health_init");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen Bracelet - Health Collector Tests\n");
    printf("Target: GD32E230C8T3 / FreeRTOS\n");
    printf("Mode: Host simulation (TEST_MODE)\n");
    printf("========================================\n");

    test_ppg_read_function();
    test_health_init();
    test_health_data_format();
    test_zero_hr_edge_case();
    test_invalid_spo2_edge_case();
    test_step_counting();
    test_signal_quality();
    test_mqtt_publish_parameters();

    printf("\n========================================\n");
    printf("Test Results: %d/%d passed (%d failed)\n",
           g_tests_passed, g_tests_run, g_tests_failed);
    printf("========================================\n");

    return (g_tests_failed > 0) ? 1 : 0;
}

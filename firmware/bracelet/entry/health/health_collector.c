/*
 * Eregen (颐贞) - Health Data Collector Implementation
 * Collects PPG heart rate + SpO2 and IMU step count, encodes and publishes via MQTT.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "health/health_collector.h"
#include "../sensors_ppg.h"
#include "../sensors_imu.h"
#include "../protocol/message_encode.h"
#include "../cat1_at.h"
#include "../common/log.h"

#ifdef TEST_MODE
#include <string.h>
#include <stdio.h>
#include <time.h>
#else
#include <stdio.h>
#include <string.h>
#include <time.h>
#endif

/* Device ID for this bracelet entry unit */
#define DEVICE_ID    "BR-0001"

/* MQTT topic for health data upstream */
#define HEALTH_TOPIC "device/" DEVICE_ID "/health"

/* Payload buffer for JSON health data before encoding */
#define HEALTH_PAYLOAD_BUF_SIZE 128U

/* Step count accumulator window (samples) */
#define STEP_SAMPLE_WINDOW     512U

/**
 * Internal state for step counting.
 */
typedef struct {
    uint32_t step_count;
    uint32_t sample_count;
    float    prev_magnitude;
    bool     step_detected;
} step_counter_t;

static step_counter_t s_step_counter = {0};

/*
 * Reset the step counter to zero.
 */
static void step_counter_reset(void)
{
    s_step_counter.step_count = 0;
    s_step_counter.sample_count = 0;
    s_step_counter.prev_magnitude = 0.0f;
    s_step_counter.step_detected = false;
}

/*
 * Update step counter with a new acceleration magnitude sample.
 * Simple peak detection: a step is counted when magnitude crosses
 * above a threshold after having been below it.
 * @param magnitude Acceleration magnitude in g.
 * @return true if a step was detected.
 */
static bool step_counter_update(float magnitude)
{
    const float STEP_THRESHOLD = 1.2f;  /* g */
    const float REBOUND_HYSTERESIS = 0.8f;  /* g */

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

/*
 * Encode health data into a JSON payload string.
 * Format: {"hr":N,"spo2":N,"steps":N,"quality":N}
 * @param buf Output buffer.
 * @param len Buffer size.
 * @param hr Heart rate in BPM.
 * @param spo2 Blood oxygen percentage.
 * @param steps Step count since last collection.
 * @param quality Signal quality 0-100.
 * @return Number of bytes written, or -1 on error.
 */
static int encode_health_payload(char *buf, uint16_t len,
                                  uint16_t hr, uint8_t spo2,
                                  uint32_t steps, uint8_t quality)
{
    int written;

    written = snprintf(buf, len,
        "{\"hr\":%u,\"spo2\":%u,\"steps\":%lu,\"quality\":%u}",
        (unsigned)hr,
        (unsigned)spo2,
        (unsigned long)steps,
        (unsigned)quality);

    if (written < 0 || (uint16_t)written >= len) {
        return -1;
    }
    return written;
}

/*
 * Collect health data and publish via MQTT.
 */
void health_collect_and_send(void)
{
    /* Step 1: Read PPG sensor data */
    ppg_data_t ppg;
    bool ppg_ok = ppg_read(&ppg);
    if (!ppg_ok) {
        log_warn("PPG read failed, sending zeros");
        ppg.hr = 0;
        ppg.spo2 = 0;
        ppg.quality = 0;
    }

    /* Step 2: Read IMU acceleration and update step count */
    imu_data_t accel = imu_read_accel();
    float mag = imu_accel_magnitude(&accel);
    step_counter_update(mag);

    /* Step 3: Encode health payload as JSON */
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

    /* Step 4: Build eregen_msg_t */
    eregen_msg_t msg;
    memset(&msg, 0, sizeof(msg));
    msg.type = MSG_HEALTH;
    strncpy(msg.dev_id, DEVICE_ID, sizeof(msg.dev_id) - 1);
    msg.timestamp = (uint64_t)time(NULL);
    memcpy(msg.payload, (const uint8_t *)payload_buf, (size_t)payload_len);
    msg.payload_len = (uint16_t)payload_len;

    /* Step 5: Encode message to wire format */
    uint8_t encoded[MAX_MSG_LEN];
    int encoded_len = message_encode(&msg, encoded, (uint16_t)sizeof(encoded));

    if (encoded_len < 0) {
        log_error("Message encode failed, code=%d", encoded_len);
        return;
    }

    /* Step 6: Publish via MQTT */
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

/*
 * Initialize the health collector subsystem.
 */
void health_init(void)
{
    log_info("Initializing health collector");
    step_counter_reset();

    /* Verify PPG sensor is available */
    if (!ppg_init()) {
        log_error("PPG sensor initialization failed");
    } else {
        log_info("PPG sensor initialized OK");
    }

    /* Verify IMU sensor is available */
    if (!imu_init()) {
        log_error("IMU sensor initialization failed");
    } else {
        log_info("IMU sensor initialized OK");
    }
}

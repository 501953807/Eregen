/*
 * Eregen (颐贞) - Bracelet Firmware Entry Point (Plus Tier)
 * Target: GD32E230C8T3 (ARM Cortex-M4)
 * RTOS: FreeRTOS
 *
 * Plus tier extends entry with: geofence, fall detection, battery optimizer,
 * BLE pairing for family APP provisioning.
 *
 * MIT License
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

#include <stdio.h>
#include <string.h>
#include <time.h>
#include "FreeRTOS.h"
#include "task.h"
#include "queue.h"
#include "board_init.h"
#include "sensors_ppg.h"
#include "sensors_imu.h"
#include "gps_nmea.h"
#include "cat1_at.h"
#include "display_st7789.h"
#include "battery_adc.h"
#include "sos_button.h"
#include "free_rtos_tasks.h"
#include "../common/crc16.h"
#include "../common/ring_buffer.h"
#include "../common/log.h"
#include "../protocol/message_encode.h"
#include "../protocol/heartbeat.h"

/* Plus-tier modules */
#include "geofence_manager.h"
#include "fall_detect.h"
#include "battery_optimizer.h"
#include "ble_pair.h"

/* Device ID: BR-XXXX format */
#define DEVICE_ID_PREFIX          "BR-"
#define DEVICE_ID_SERIAL_LEN      4U

/* Global device serial number (set once at boot) */
char s_device_id[17];

/* Step counter for health data */
static uint32_t s_step_count = 0;
static float s_last_accel_mag = 0.0f;

/* GPS timestamp monotonically increasing counter */
static uint32_t s_gps_timestamp = 0;

/* Fall detection sample accumulator */
static fall_sample_t s_fall_sample;
static uint32_t s_fall_tick_counter = 0;

/* Battery optimization state tracking */
static bool s_batt_opt_applied = false;

/* Forward declarations — base tasks */
static void vSensorTask(void *pvParameters);
static void vGPSTask(void *pvParameters);
static void vCommTask(void *pvParameters);
static void vDisplayTask(void *pvParameters);
static void vSOSTask(void *pvParameters);

/* Forward declarations — plus-tier task functions (declared in free_rtos_tasks.c) */
extern bool tasks_plus_init(void);
extern bool tasks_send_geofence_alert(const geofence_alert_t *alert, uint32_t timeout_ms);
extern bool tasks_send_fall_alert(const fall_alert_t *alert, uint32_t timeout_ms);
extern bool tasks_broadcast_battery_opt(const batt_opt_msg_t *msg, uint32_t timeout_ms);
extern void* tasks_get_comm_handle(void);

/* Forward declarations */
static void generate_device_id(void);
static void uart_tx(const uint8_t *data, uint16_t len);

/*
 * UART transmit callback for the logging subsystem.
 * Writes bytes to USART0 (debug console).
 */
static void uart_tx(const uint8_t *data, uint16_t len)
{
    if (!data || len == 0) {
        return;
    }
    for (uint16_t i = 0; i < len; i++) {
        while (usart_flag_get(USART0, USART_FLAG_TC) == RESET) {
            /* Wait for transmit register empty */
        }
        usart_data_transmit(USART0, data[i]);
    }
}

/*
 * Generate a pseudo-unique device ID in BR-XXXX format.
 * Uses factory-calibrated UID from GD32.
 */
static void generate_device_id(void)
{
    uint32_t uid_low = *(volatile uint32_t*)(0x1FFFF7E8UL);
    uint32_t uid_mid = *(volatile uint32_t*)(0x1FFFF7ECUL);
    snprintf(s_device_id, sizeof(s_device_id),
             "%s%04X", DEVICE_ID_PREFIX,
             (uint16_t)(uid_low ^ uid_mid));
    s_device_id[sizeof(s_device_id) - 1] = '\0';
}

/*
 * Main entry point.
 * Initializes hardware, creates all tasks (base + plus-tier), starts scheduler.
 */
int main(void)
{
    /* Initialize hardware */
    board_init_all();
    generate_device_id();

    /* Initialize logging system with UART callback */
    log_init();
    log_register_uart_tx(uart_tx);
    log_set_level(LOG_INFO);

    log_info("Eregen Bracelet Plus Firmware v1.0");
    log_info("Device ID: %s", s_device_id);
    log_info("Target: GD32E230C8T3 (Cortex-M4) + Plus Features\n");

    /* Verify CRC16 with known test vector */
    const uint8_t crc_test[] = "123456789";
    uint16_t crc_result = crc16_calc(crc_test, sizeof(crc_test) - 1);
    if (crc_result != 0x29B1) {
        log_error("CRC16 verification failed! got 0x%04X", crc_result);
    } else {
        log_debug("CRC16 verification passed (0x%04X)", crc_result);
    }

    /* Initialize base message queues and tasks (entry tier) */
    tasks_init();

    /* ---- Initialize plus-tier modules ---- */

    /* Geofence manager — load zones from NVS */
    /* Note: NVS ops struct would be provided by the platform init layer.
     * For now we pass NULL and the module will handle it gracefully. */
    geofence_init(NULL);
    log_info("Geofence manager initialized, zones loaded from NVS");

    /* Fall detection engine */
    fall_detect_init();
    log_info("Fall detection engine initialized");

    /* Battery optimizer */
    battery_optimizer_init();
    log_info("Battery optimizer initialized");

    /* BLE pairing (only if not yet provisioned) */
    ble_pair_init();
    if (!ble_pair_is_provisioned()) {
        ble_pair_start_advertising();
        log_info("BLE advertising started for initial provisioning");
    } else {
        log_info("Device already provisioned — skipping BLE advertising");
    }

    /* Create base sensor sampling task */
    xTaskCreate(vSensorTask,
                "Sensor",
                TASK_SENSOR_STACK,
                NULL,
                TASK_SENSOR_PRIORITY,
                NULL);

    /* Create GPS positioning task */
    xTaskCreate(vGPSTask,
                "GPS",
                TASK_GPS_STACK,
                NULL,
                TASK_GPS_PRIORITY,
                NULL);

    /* Create communication task (MQTT/Cat1) */
    TaskHandle_t comm_handle = NULL;
    xTaskCreate(vCommTask,
                "Comm",
                TASK_COMM_STACK,
                NULL,
                TASK_COMM_PRIORITY,
                &comm_handle);
    tasks_set_comm_handle(comm_handle);

    /* Create display task */
    xTaskCreate(vDisplayTask,
                "Display",
                TASK_DISPLAY_STACK,
                NULL,
                TASK_DISPLAY_PRIORITY,
                NULL);

    /* Create SOS monitoring task */
    xTaskCreate(vSOSTask,
                "SOS",
                TASK_SOS_STACK,
                NULL,
                TASK_SOS_PRIORITY,
                NULL);

    /* Create plus-tier tasks (geofence, fall_detect, battery_opt, ble_pair) */
    tasks_plus_init();

    /* Start heartbeat publisher (after all tasks are created) */
    heartbeat_start();

    /* Start the scheduler */
    vTaskStartScheduler();

    /* Should never reach here */
    for (;;);
}

/*
 * Sensor Task
 * Reads heart rate (PPG GT320), SpO2, IMU (ICM-42670-P).
 * Feeds IMU data to fall detection engine. Publishes health data via queue.
 */
static void vSensorTask(void *pvParameters)
{
    (void)pvParameters;

    /* Initialize PPG sensor */
    if (!ppg_init()) {
        log_error("PPG sensor initialization failed");
        for (;;) {
            gpio_bit_toggle(GPIOA, GPIO_PIN_1);
            vTaskDelay(pdMS_TO_TICKS(500));
        }
    }
    log_info("PPG sensor initialized");

    /* Initialize IMU sensor */
    if (!imu_init()) {
        log_error("IMU sensor initialization failed");
        for (;;) {
            gpio_bit_toggle(GPIOA, GPIO_PIN_1);
            vTaskDelay(pdMS_TO_TICKS(500));
        }
    }
    log_info("IMU sensor initialized");

    /* Initialize battery ADC */
    battery_init();

    uint32_t step_count = 0;
    uint32_t last_step_count = 0;
    uint32_t imu_tick_counter = 0;

    for (;;) {
        /* Read PPG data */
        ppg_data_t health = ppg_get_data();

        /* Read IMU data */
        imu_data_t imu = imu_get_data();
        float accel_mag = imu_accel_magnitude(&imu);

        /* Feed IMU sample to fall detection engine every tick. */
        s_fall_sample.ax = imu.ax;
        s_fall_sample.ay = imu.ay;
        s_fall_sample.az = imu.az;
        s_fall_sample.gx = imu.gx;
        s_fall_sample.gy = imu.gy;
        s_fall_sample.gz = imu.gz;
        s_fall_sample.tick = xTaskGetTickCount();

        if (++imu_tick_counter >= (1000 / FALL_DETECT_IMU_ODR_HZ)) {
            imu_tick_counter = 0;
            fall_event_t event = fall_detect_feed(&s_fall_sample);

            if (event.alarm_ready && event.confidence > 0.5f) {
                log_error("FALL DETECTED: confidence=%.2f, consecutive=%u",
                          event.confidence, event.consecutive);

                /* Send fall alert via plus-tier queue. */
                fall_alert_t alert;
                alert.lat = 0.0;   /* Filled by GPS task */
                alert.lon = 0.0;
                alert.confidence = event.confidence;
                alert.timestamp = s_gps_timestamp;

                tasks_send_fall_alert(&alert, pdMS_TO_TICKS(200));
            }
        }

        /* Simple step detection */
        if ((accel_mag - s_last_accel_mag) > 0.2f) {
            step_count++;
        }
        s_last_accel_mag = accel_mag;

        /* Send health data via queue every ~1 second */
        if (step_count != last_step_count) {
            health_data_t msg;
            msg.hr = health.hr;
            msg.spo2 = health.spo2;
            msg.step_count = step_count;

            if (tasks_send_health(&msg, pdMS_TO_TICKS(100))) {
                log_info("HEALTH: HR=%u, SpO2=%u, Steps=%lu",
                       (unsigned)health.hr, (unsigned)health.spo2,
                       (unsigned long)step_count);
            }
            last_step_count = step_count;
        }

        /* Send battery status periodically */
        static uint32_t batt_tick = 0;
        if (++batt_tick >= 100) {  /* Every ~10 seconds */
            batt_tick = 0;
            battery_status_t batt = battery_get_status();
            log_info("BATTERY: %umV, %u%%",
                   (unsigned)batt.voltage_mv, (unsigned)batt.percent);
        }

        vTaskDelay(pdMS_TO_TICKS(1000 / FALL_DETECT_IMU_ODR_HZ));
    }
}

/*
 * GPS Task
 * Reads location from GPS module, feeds to geofence checker.
 */
static void vGPSTask(void *pvParameters)
{
    (void)pvParameters;

    gps_init();
    log_info("GPS NMEA parser initialized");

    /* Previous position for change detection. */
    static double s_prev_lat = 0.0;
    static double s_prev_lon = 0.0;
    bool s_first_fix = true;

    for (;;) {
        /* Read characters from GPS UART */
        while (usart_flag_get(USART2, USART_FLAG_RBNE) != RESET) {
            char c = (char)usart_data_receive(USART2);
            gps_parse_char(c);
        }

        /* Check for valid fix */
        if (gps_has_valid_fix()) {
            gps_fix_t fix = gps_get_fix();

            /* Apply battery-optimized GPS interval. */
            optimizer_config_t opt_cfg;
            if (battery_optimizer_get_config(&opt_cfg)) {
                static uint32_t gps_interval_s = OPTIMIZER_GPS_HIGH_BATT_S;
                if (opt_cfg.gps_interval_s != gps_interval_s) {
                    gps_interval_s = opt_cfg.gps_interval_s;
                    log_info("GPS interval updated: %us", gps_interval_s);
                }
            }

            static uint32_t gps_counter = 0;
            gps_counter++;

            if (gps_counter >= gps_interval_s) {
                gps_counter = 0;

                location_data_t msg;
                msg.lat = fix.lat;
                msg.lon = fix.lon;
                msg.accuracy = fix.accuracy;
                msg.timestamp = s_gps_timestamp++;

                /* Check geofence on each GPS fix. */
                geofence_state_t gf_state;
                geofence_result_t gf_ret = geofence_check_position(
                    fix.lat, fix.lon, &gf_state
                );

                if (gf_ret == GEOFENCE_ERR_OUTSIDE && gf_state.exited_zone_id != 0xFF) {
                    /* Elder left a safe zone — send alert. */
                    geofence_alert_t alert;
                    alert.zone_id = gf_state.exited_zone_id;
                    alert.lat = fix.lat;
                    alert.lon = fix.lon;
                    alert.timestamp = msg.timestamp;

                    if (tasks_send_geofence_alert(&alert, pdMS_TO_TICKS(500))) {
                        log_warn("GEOFENCE VIOLATION: exited zone %u at (%.4f, %.4f)",
                                 alert.zone_id, alert.lat, alert.lon);
                    }
                }

                /* Send location to comm task. */
                if (tasks_send_location(&msg, pdMS_TO_TICKS(100))) {
                    log_info("LOCATION: lat=%.6f, lon=%.6f, sats=%u, acc=%um",
                           fix.lat, fix.lon,
                           (unsigned)fix.satellites,
                           (unsigned)fix.accuracy);
                }

                /* Update previous position. */
                s_prev_lat = fix.lat;
                s_prev_lon = fix.lon;
                s_first_fix = false;
            }
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/*
 * Communication Task
 * Handles Cat1 cellular connectivity and MQTT publish/subscribe.
 * Receives data from other tasks via queues and publishes to cloud.
 */
static void vCommTask(void *pvParameters)
{
    (void)pvParameters;

    if (!cat1_init(NULL)) {
        log_error("Cat1 module not responding");
    } else {
        log_info("Cat1 module initialized");
    }

    if (!cat1_connect()) {
        log_warn("APN connection failed, will retry");
    } else {
        log_info("Cat1 connected");
    }

    uint32_t heartbeat_counter = 0;
    uint32_t rssi_counter = 0;

    for (;;) {
        /* Send heartbeat every 30 seconds */
        if (++heartbeat_counter >= 30) {
            heartbeat_counter = 0;

            battery_status_t batt = battery_get_status();
            log_info("HEARTBEAT: dev_id=%s, bat=%u",
                   s_device_id, (unsigned)batt.percent);
        }

        /* Signal strength check */
        if (++rssi_counter >= 10) {
            rssi_counter = 0;
            int16_t rssi = cat1_get_signal_strength();
            log_info("RSSI: %ddBm", (int)rssi);
        }

        /* Process incoming health data from queue */
        health_data_t health_msg;
        void *health_q = tasks_get_health_queue();
        if (health_q &&
            xQueueReceive((QueueHandle_t)health_q, &health_msg, pdMS_TO_TICKS(1)) == pdPASS) {
            /* Data received from sensor task — publish to MQTT */
        }

        /* Process incoming location data from queue */
        location_data_t loc_msg;
        void *loc_q = tasks_get_location_queue();
        if (loc_q &&
            xQueueReceive((QueueHandle_t)loc_q, &loc_msg, pdMS_TO_TICKS(1)) == pdPASS) {
            /* Location data received — publish to MQTT */
        }

        /* Process geofence alerts from plus-tier task */
        geofence_alert_t gf_alert;
        void *gf_q = tasks_get_geofence_queue();
        if (gf_q &&
            xQueueReceive((QueueHandle_t)gf_q, &gf_alert, pdMS_TO_TICKS(1)) == pdPASS) {
            log_warn("Geofence alert forwarded to MQTT: zone %u", gf_alert.zone_id);
        }

        /* Process fall alerts from plus-tier task */
        fall_alert_t fall_alert;
        void *fall_q = tasks_get_fall_queue();
        if (fall_q &&
            xQueueReceive((QueueHandle_t)fall_q, &fall_alert, pdMS_TO_TICKS(1)) == pdPASS) {
            log_error("Fall alert forwarded to MQTT: confidence=%.2f",
                      fall_alert.confidence);
        }

        /* Process battery optimization commands */
        batt_opt_msg_t opt_msg;
        void *batt_q = tasks_get_batt_opt_queue();
        if (batt_q &&
            xQueueReceive((QueueHandle_t)batt_q, &opt_msg, pdMS_TO_TICKS(1)) == pdPASS) {
            log_info("Battery opt applied: GPS=%us PPG=%us tier=%u",
                     opt_msg.gps_interval_s, opt_msg.ppg_interval_s, opt_msg.tier);
            s_batt_opt_applied = true;
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/*
 * Display Task
 * Controls ST7789 LCD display. Shows status, battery, connection indicator.
 * Plus tier adds geofence status and fall detection status indicators.
 */
static void vDisplayTask(void *pvParameters)
{
    (void)pvParameters;

    if (!display_init()) {
        log_error("Display initialization failed");
        for (;;) {
            gpio_bit_toggle(GPIOA, GPIO_PIN_0);
            vTaskDelay(pdMS_TO_TICKS(500));
        }
    }
    log_info("Display initialized");

    for (;;) {
        /* Update battery indicator */
        battery_status_t batt = battery_get_status();

        uint16_t batt_color = DISPLAY_COLOR_RED;
        if (batt.percent > 50U) {
            batt_color = DISPLAY_COLOR_GREEN;
        } else if (batt.percent > 20U) {
            batt_color = DISPLAY_COLOR_YELLOW;
        }
        display_draw_rect_filled(110, 0, 134, 12, DISPLAY_COLOR_WHITE);
        display_draw_rect_filled(112, 2,
                                 (uint16_t)(112 + (20U * batt.percent / 100U)),
                                 10, batt_color);

        char batt_str[8];
        snprintf(batt_str, sizeof(batt_str), "%u%%", (unsigned)batt.percent);
        display_draw_string(2, 2, batt_str, DISPLAY_COLOR_WHITE, DISPLAY_COLOR_BLACK);

        /* Show signal strength bar */
        int16_t rssi = cat1_get_signal_strength();
        uint8_t bars = 0;
        if (rssi > CAT1_RSSI_GOOD) bars = 4;
        else if (rssi > CAT1_RSSI_WEAK) bars = 2;
        else if (rssi > -127) bars = 1;

        for (uint8_t i = 0; i < 4; i++) {
            uint16_t bx = 20U + i * 4U;
            uint16_t by = (i < bars) ? 2U : 8U;
            uint16_t bh = (i < bars) ? 10U : 4U;
            display_draw_rect_filled(bx, by, bx + 2, by + bh,
                                     (i < bars) ? DISPLAY_COLOR_GREEN :
                                                  DISPLAY_COLOR_BLUE);
        }

        /* Show connection status */
        bool connected = cat1_is_connected();
        const char *conn_str = connected ? "CONNECTED" : "DISCONNECTED";
        uint16_t conn_color = connected ? DISPLAY_COLOR_GREEN : DISPLAY_COLOR_RED;
        display_draw_string(2, 14, conn_str, conn_color, DISPLAY_COLOR_BLACK);

        /* Plus tier: show geofence zone count */
        char gf_str[32];
        uint8_t zone_count = geofence_get_zone_count();
        snprintf(gf_str, sizeof(gf_str), "FENCE:%u", zone_count);
        display_draw_string(2, 26, gf_str, DISPLAY_COLOR_CYAN, DISPLAY_COLOR_BLACK);

        /* Plus tier: show fall detection status */
        if (fall_detect_is_alarm_active()) {
            display_draw_string(2, 38, "FALL!", DISPLAY_COLOR_RED, DISPLAY_COLOR_BLACK);
        }

        /* Plus tier: show battery optimization tier */
        uint8_t tier = battery_optimizer_get_tier();
        const char *tier_str[] = {"HIGH", "MED", "LOW"};
        display_draw_string(2, 50, tier_str[tier], DISPLAY_COLOR_MAGENTA, DISPLAY_COLOR_BLACK);

        display_update();
        vTaskDelay(pdMS_TO_TICKS(2000));
    }
}

/*
 * SOS Monitoring Task
 * Watches the physical SOS button and triggers alerts on long press.
 */
static void vSOSTask(void *pvParameters)
{
    (void)pvParameters;

    sos_init();
    log_info("SOS button monitoring started");

    for (;;) {
        sos_task();

        if (sos_is_long_press()) {
            log_error("SOS ALERT TRIGGERED!");

            gps_fix_t fix;
            double lat = 0.0, lon = 0.0;
            if (gps_has_valid_fix()) {
                fix = gps_get_fix();
                lat = fix.lat;
                lon = fix.lon;
            }

            sos_alert_t alert;
            alert.lat = lat;
            alert.lon = lon;
            alert.timestamp = s_gps_timestamp;

            if (tasks_send_sos(&alert, pdMS_TO_TICKS(500))) {
                log_info("SOS alert sent to comm task");
            }

            /* Flash all LEDs red */
            for (uint8_t i = 0; i < 5; i++) {
                gpio_bit_reset(GPIOA, GPIO_PIN_0);
                gpio_bit_reset(GPIOA, GPIO_PIN_1);
                vTaskDelay(pdMS_TO_TICKS(200));
                gpio_bit_set(GPIOA, GPIO_PIN_0);
                gpio_bit_set(GPIOA, GPIO_PIN_1);
                vTaskDelay(pdMS_TO_TICKS(200));
            }

            sos_reset_long_press_flag();
        }

        if (sos_is_pressed()) {
            gpio_bit_reset(GPIOA, GPIO_PIN_1);
            vTaskDelay(pdMS_TO_TICKS(100));
            gpio_bit_set(GPIOA, GPIO_PIN_1);
            sos_reset_pressed_flag();
        }

        vTaskDelay(pdMS_TO_TICKS(SOS_CHECK_INTERVAL_MS));
    }
}

/*
 * Eregen (颐贞) - Bracelet Firmware Entry Point
 * Target: GD32E230C8T3 (ARM Cortex-M4)
 * RTOS: FreeRTOS
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>
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

/* Device ID: BR-XXXX format */
#define DEVICE_ID_PREFIX          "BR-"
#define DEVICE_ID_SERIAL_LEN      4U

/* Prototype MQTT shared secret (Phase 1.2 — upgraded to TLS/cert pinning in Phase 2) */
#define MQTT_SHARED_SECRET        "eregen_dev_prototype"

/* MQTT broker endpoint (EMQX via Cat1 TCP) */
#define MQTT_BROKER_HOST          "mqtt.eregen.dev"
#define MQTT_BROKER_PORT          1883

/* Global device serial number (set once at boot) */
char s_device_id[17];

/* Step counter for health data */
static uint32_t s_step_count = 0;
static float s_last_accel_mag = 0.0f;

/* GPS timestamp monotonically increasing counter */
uint32_t s_gps_timestamp = 0;

/* Forward declarations */
static void vSensorTask(void *pvParameters);
static void vGPSTask(void *pvParameters);
static void vCommTask(void *pvParameters);
static void vDisplayTask(void *pvParameters);
static void vSOSTask(void *pvParameters);
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
    /* Use GD32 unique 96-bit UID as device serial */
    /* Format: BR-ABCD where ABCD is derived from UID */
    uint32_t uid_low = *(volatile uint32_t*)(0x1FFFF7E8UL);
    uint32_t uid_mid = *(volatile uint32_t*)(0x1FFFF7ECUL);
    snprintf(s_device_id, sizeof(s_device_id),
             "%s%04X", DEVICE_ID_PREFIX,
             (uint16_t)(uid_low ^ uid_mid));
    s_device_id[sizeof(s_device_id) - 1] = '\0';
}

/*
 * Main entry point.
 * Initializes the board, creates all tasks, starts the scheduler.
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

    log_info("Eregen Bracelet Firmware v1.0");
    log_info("Device ID: %s", s_device_id);
    log_info("Target: GD32E230C8T3 (Cortex-M4)\n");

    /* Verify CRC16 with known test vector */
    const uint8_t crc_test[] = "123456789";
    uint16_t crc_result = crc16_calc(crc_test, sizeof(crc_test) - 1);
    if (crc_result != 0x29B1) {
        log_error("CRC16 verification failed! got 0x%04X", crc_result);
    } else {
        log_debug("CRC16 verification passed (0x%04X)", crc_result);
    }

    /* Initialize message queues */
    tasks_init();

    /* Create sensor sampling task */
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

    /* Start heartbeat publisher (after all tasks are created) */
    heartbeat_start();

    /* Start the scheduler */
    vTaskStartScheduler();

    /* Should never reach here */
    for (;;);
}

/*
 * Sensor Task
 * Reads heart rate (PPG GT320), SpO2, and accelerometer (ICM-42670-P).
 * Publishes health data via message queue to Comm task.
 */
static void vSensorTask(void *pvParameters)
{
    (void)pvParameters;

    /* Initialize PPG sensor */
    if (!ppg_init()) {
        log_error("PPG sensor initialization failed");
        /* Blink LED blue to indicate error */
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

    for (;;) {
        /* Read PPG data */
        ppg_data_t health = ppg_get_data();

        /* Read IMU for step detection */
        imu_data_t imu = imu_get_data();
        float accel_mag = imu_accel_magnitude(&imu);

        /* Simple step detection: detect peaks in acceleration magnitude */
        float threshold = 0.2f;  /* g change threshold */
        if ((accel_mag - s_last_accel_mag) > threshold) {
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

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/*
 * GPS Task
 * Reads location from GPS module (u-blox NEO-M9N or UGN-7345).
 * Parses NMEA sentences and publishes location data via message queue.
 */
static void vGPSTask(void *pvParameters)
{
    (void)pvParameters;

    /* Initialize GPS parser */
    gps_init();
    log_info("GPS NMEA parser initialized");

    for (;;) {
        /* Read characters from GPS UART and feed to parser.
         * In production, this would be driven by UART RX interrupt.
         * For now, poll the USART receive buffer.
         */
        while (usart_flag_get(USART2, USART_FLAG_RBNE) != RESET) {
            char c = (char)usart_data_receive(USART2);
            gps_parse_char(c);
        }

        /* Check for valid fix and send location data */
        if (gps_has_valid_fix()) {
            gps_fix_t fix = gps_get_fix();

            location_data_t msg;
            msg.lat = fix.lat;
            msg.lon = fix.lon;
            msg.accuracy = fix.accuracy;
            msg.timestamp = s_gps_timestamp++;

            if (tasks_send_location(&msg, pdMS_TO_TICKS(100))) {
                log_info("LOCATION: lat=%.6f, lon=%.6f, sats=%u, acc=%um",
                       fix.lat, fix.lon,
                       (unsigned)fix.satellites,
                       (unsigned)fix.accuracy);
            }
        }

        vTaskDelay(pdMS_TO_TICKS(5000));
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

    /* Initialize Cat1 module */
    if (!cat1_init(NULL)) {
        log_error("Cat1 module not responding");
        /* Continue running but log errors */
    } else {
        log_info("Cat1 module initialized");
    }

    /* Connect to APN */
    if (!cat1_connect()) {
        log_warn("APN connection failed, will retry");
    } else {
        log_info("Cat1 connected");
    }

    /* Establish TCP connection to MQTT broker */
    if (!cat1_tcp_connect(MQTT_BROKER_HOST, MQTT_BROKER_PORT)) {
        log_warn("TCP connect to MQTT broker failed, will retry");
    } else {
        /* Send MQTT CONNECT with device credentials */
        cat1_mqtt_connect(s_device_id, s_device_id, MQTT_SHARED_SECRET);
    }

    /* Heartbeat interval */
    uint32_t heartbeat_counter = 0;

    for (;;) {
        /* Send heartbeat every 30 seconds */
        if (++heartbeat_counter >= 30) {
            heartbeat_counter = 0;

            /* Build and publish heartbeat JSON to MQTT broker */
            battery_status_t batt = battery_get_status();
            char json_buf[80];
            int len = snprintf(json_buf, sizeof(json_buf),
                "{\"type\":\"heartbeat\",\"dev_id\":\"%s\",\"bat\":%u,\"ts\":%lu}",
                s_device_id, (unsigned)batt.percent,
                (unsigned long)s_gps_timestamp);
            if (len > 0 && (uint32_t)len < (uint32_t)sizeof(json_buf)) {
                cat1_mqtt_publish("eregen/device/bracelet/BR-CLOUD/up",
                    (const uint8_t *)json_buf, (uint16_t)len);
            }
        }

        /* Signal strength check */
        static uint32_t rssi_counter = 0;
        if (++rssi_counter >= 10) {
            rssi_counter = 0;
            int16_t rssi = cat1_get_signal_strength();
            log_info("RSSI: %ddBm", (int)rssi);
        }

        /* Process incoming health data from queue */
        {
            static uint16_t s_health_msg_id = 0;
            health_data_t health_msg;
            if (xQueueReceive(tasks_get_health_queue(), &health_msg, pdMS_TO_TICKS(1)) == pdPASS) {
                char json_buf[128];
                int len = snprintf(json_buf, sizeof(json_buf),
                    "{\"type\":\"health\",\"dev_id\":\"%s\",\"hr\":%u,\"spo2\":%u,\"step\":%lu,\"ts\":%lu}",
                    s_device_id,
                    (unsigned)health_msg.hr,
                    (unsigned)health_msg.spo2,
                    (unsigned long)health_msg.step_count,
                    (unsigned long)s_gps_timestamp);
                if (len > 0 && (uint32_t)len < (uint32_t)sizeof(json_buf)) {
                    s_health_msg_id++;
                    cat1_mqtt_publish("eregen/device/bracelet/BR-CLOUD/up",
                        (const uint8_t *)json_buf, (uint16_t)len);
                }
            }
        }

        /* Process incoming location data from queue */
        {
            static uint32_t s_loc_msg_id = 0;
            location_data_t loc_msg;
            if (xQueueReceive(tasks_get_location_queue(), &loc_msg, pdMS_TO_TICKS(1)) == pdPASS) {
                char json_buf[128];
                int len = snprintf(json_buf, sizeof(json_buf),
                    "{\"type\":\"location\",\"dev_id\":\"%s\",\"lat\":%.6f,\"lon\":%.6f,\"acc\":%u,\"ts\":%lu}",
                    s_device_id, loc_msg.lat, loc_msg.lon,
                    (unsigned)loc_msg.accuracy,
                    (unsigned long)loc_msg.timestamp);
                if (len > 0 && (uint32_t)len < (uint32_t)sizeof(json_buf)) {
                    s_loc_msg_id++;
                    cat1_mqtt_publish("eregen/device/bracelet/BR-CLOUD/up",
                        (const uint8_t *)json_buf, (uint16_t)len);
                }
            }
        }

        /* Process incoming SOS alert from queue */
        {
            sos_alert_t sos_msg;
            if (xQueueReceive(tasks_get_sos_queue(), &sos_msg, pdMS_TO_TICKS(1)) == pdPASS) {
                char json_buf[128];
                int len = snprintf(json_buf, sizeof(json_buf),
                    "{\"type\":\"sos\",\"dev_id\":\"%s\",\"lat\":%.6f,\"lon\":%.6f,\"ts\":%lu}",
                    s_device_id, sos_msg.lat, sos_msg.lon,
                    (unsigned long)sos_msg.timestamp);
                if (len > 0 && (uint32_t)len < (uint32_t)sizeof(json_buf)) {
                    cat1_mqtt_publish("eregen/device/bracelet/BR-CLOUD/up",
                        (const uint8_t *)json_buf, (uint16_t)len);
                    log_warn("SOS ALERT PUBLISHED: lat=%.6f, lon=%.6f", sos_msg.lat, sos_msg.lon);
                }
            }
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/*
 * Display Task
 * Controls ST7789 LCD display. Shows status, battery, connection indicator.
 */
static void vDisplayTask(void *pvParameters)
{
    (void)pvParameters;

    /* Initialize display */
    if (!display_init()) {
        log_error("Display initialization failed");
        for (;;) {
            gpio_bit_toggle(GPIOA, GPIO_PIN_0);
            vTaskDelay(pdMS_TO_TICKS(500));
        }
    }
    log_info("Display initialized");

    /* Initial screen: show device ID */
    display_clear(DISPLAY_COLOR_BLACK);
    display_draw_string(10, 10, "Eregen", DISPLAY_COLOR_WHITE, DISPLAY_COLOR_BLACK);
    display_draw_string(10, 25, s_device_id, DISPLAY_COLOR_GREEN, DISPLAY_COLOR_BLACK);
    display_update();

    for (;;) {
        /* Update battery indicator */
        battery_status_t batt = battery_get_status();

        /* Draw battery icon in top-right corner */
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

        /* Show battery percentage */
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

    /* Blink green LED to indicate system is alive */
    for (;;) {
        /* Run the SOS state machine */
        sos_task();

        /* Check for long press (3 seconds) */
        if (sos_is_long_press()) {
            log_error("SOS ALERT TRIGGERED!");

            /* Get current GPS position */
            gps_fix_t fix;
            double lat = 0.0, lon = 0.0;
            if (gps_has_valid_fix()) {
                fix = gps_get_fix();
                lat = fix.lat;
                lon = fix.lon;
            }

            /* Send SOS alert via queue */
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

        /* Brief press indication (single blink of blue LED) */
        if (sos_is_pressed()) {
            gpio_bit_reset(GPIOA, GPIO_PIN_1);
            vTaskDelay(pdMS_TO_TICKS(100));
            gpio_bit_set(GPIOA, GPIO_PIN_1);
            sos_reset_pressed_flag();
        }

        vTaskDelay(pdMS_TO_TICKS(SOS_CHECK_INTERVAL_MS));
    }
}

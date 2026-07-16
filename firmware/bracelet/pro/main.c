/*
 * Eregen (颐贞) - Pro Bracelet Main Entry Point
 * FreeRTOS task creation for Pro tier with ECG, AMOLED, and GNSS.
 *
 * Extends the entry-level main.c with Pro-specific hardware init
 * and additional FreeRTOS tasks for ECG monitoring, AMOLED display,
 * and multi-constellation GNSS positioning.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>
#include "FreeRTOS.h"
#include "task.h"
#include "queue.h"

/* Entry-level modules (shared infrastructure) */
#include "../entry/free_rtos_tasks.h"
#include "../entry/sensors_ppg.h"
#include "../entry/sensors_imu.h"
#include "../entry/cat1_at.h"
#include "../entry/battery_adc.h"
#include "../entry/sos_button.h"
#include "../entry/common/log.h"
#include "../common/crc16.h"
#include "../common/ring_buffer.h"
#include "../protocol/message_encode.h"
#include "../protocol/heartbeat.h"

/* Pro-specific modules */
#include "board_pro.h"
#include "ecg_driver.h"
#include "display_amoled.h"
#include "gps_gnss.h"
#include "free_rtos_tasks.h"  /* Pro task definitions */

/* Device ID: BR-XXXX format */
#define DEVICE_ID_PREFIX    "BR-"
#define DEVICE_ID_SERIAL_LEN 4U

/* Global device serial number */
char s_device_id[17];

/* Forward declarations - entry-level tasks */
static void vSensorTask(void *pvParameters);
static void vCommTask(void *pvParameters);
static void vSOSTask(void *pvParameters);
static void generate_device_id(void);
static void uart_tx(const uint8_t *data, uint16_t len);

/*
 * UART transmit callback for logging (USART0 debug console).
 */
static void uart_tx(const uint8_t *data, uint16_t len)
{
    if (!data || len == 0) return;
    for (uint16_t i = 0; i < len; i++) {
        while (usart_flag_get(USART0, USART_FLAG_TC) == RESET) {}
        usart_data_transmit(USART0, data[i]);
    }
}

/*
 * Generate device ID from GD32 unique 96-bit UID.
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
 * Initializes Pro board, creates all tasks (entry + Pro), starts scheduler.
 */
int main(void)
{
    /* Initialize Pro board hardware (different from entry) */
    board_pro_init_all();
    generate_device_id();

    /* Initialize logging */
    log_init();
    log_register_uart_tx(uart_tx);
    log_set_level(LOG_INFO);

    log_info("Eregen Pro Bracelet Firmware v1.0");
    log_info("Device ID: %s", s_device_id);
    log_info("Target: GD32E230C8T3 (Cortex-M4) - Pro Tier");
    log_info("Features: ECG + AMOLED + Multi-GNSS\n");

    /* Verify CRC16 */
    const uint8_t crc_test[] = "123456789";
    uint16_t crc_result = crc16_calc(crc_test, sizeof(crc_test) - 1);
    if (crc_result != 0x29B1) {
        log_error("CRC16 verification failed! got 0x%04X", crc_result);
    } else {
        log_debug("CRC16 verification passed (0x%04X)", crc_result);
    }

    /* Initialize base message queues */
    tasks_init();

    /* Initialize Pro-specific queues */
    pro_tasks_init();

    /* Create sensor task (PPG + IMU + fall detection) */
    xTaskCreate(vSensorTask,
                "Sensor",
                TASK_SENSOR_STACK,
                NULL,
                TASK_SENSOR_PRIORITY,
                NULL);

    /* Create SOS task */
    xTaskCreate(vSOSTask,
                "SOS",
                TASK_SOS_STACK,
                NULL,
                TASK_SOS_PRIORITY,
                NULL);

    /* Create communication task (Cat1 + MQTT) */
    TaskHandle_t comm_handle = NULL;
    xTaskCreate(vCommTask,
                "Comm",
                TASK_COMM_STACK,
                NULL,
                TASK_COMM_PRIORITY,
                &comm_handle);
    tasks_set_comm_handle(comm_handle);

    /* Start heartbeat publisher */
    heartbeat_start();

    /* Create Pro-specific tasks (ECG, AMOLED, GNSS) */
    if (!pro_tasks_create()) {
        log_error("Pro: Failed to create Pro tasks");
        /* Continue running base tasks even if Pro features fail */
    }

    /* Start the scheduler */
    vTaskStartScheduler();

    /* Should never reach here */
    for (;;);
}

/* ----------------------------------------------------------------
 * Sensor Task (shared with entry)
 * Reads PPG, IMU, battery. Publishes health data.
 * ---------------------------------------------------------------- */

static void vSensorTask(void *pvParameters)
{
    (void)pvParameters;

    if (!ppg_init()) {
        log_error("PPG sensor init failed");
        for (;;) {
            gpio_bit_toggle(GPIOA, GPIO_PIN_1);
            vTaskDelay(pdMS_TO_TICKS(500));
        }
    }
    log_info("PPG sensor initialized");

    if (!imu_init()) {
        log_error("IMU sensor init failed");
        for (;;) {
            gpio_bit_toggle(GPIOA, GPIO_PIN_1);
            vTaskDelay(pdMS_TO_TICKS(500));
        }
    }
    log_info("IMU sensor initialized");

    battery_init();

    uint32_t step_count = 0;
    uint32_t last_step_count = 0;
    static float s_last_accel_mag = 0.0f;

    for (;;) {
        ppg_data_t health = ppg_get_data();
        imu_data_t imu = imu_get_data();
        float accel_mag = imu_accel_magnitude(&imu);

        float threshold = 0.2f;
        if ((accel_mag - s_last_accel_mag) > threshold) {
            step_count++;
        }
        s_last_accel_mag = accel_mag;

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

        /* Send battery status every ~10 seconds */
        static uint32_t batt_tick = 0;
        if (++batt_tick >= 100) {
            batt_tick = 0;
            battery_status_t batt = battery_get_status();
            log_info("BATTERY: %umV, %u%%",
                     (unsigned)batt.voltage_mv, (unsigned)batt.percent);
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/* ----------------------------------------------------------------
 * Communication Task (shared with entry)
 * Handles Cat1 cellular connectivity and MQTT publish/subscribe.
 * ---------------------------------------------------------------- */

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

    for (;;) {
        if (++heartbeat_counter >= 30) {
            heartbeat_counter = 0;
            battery_status_t batt = battery_get_status();
            log_info("HEARTBEAT: dev_id=%s, bat=%u",
                     s_device_id, (unsigned)batt.percent);
        }

        static uint32_t rssi_counter = 0;
        if (++rssi_counter >= 10) {
            rssi_counter = 0;
            int16_t rssi = cat1_get_signal_strength();
            log_info("RSSI: %ddBm", (int)rssi);
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/* ----------------------------------------------------------------
 * SOS Task (shared with entry)
 * ---------------------------------------------------------------- */

static void vSOSTask(void *pvParameters)
{
    (void)pvParameters;

    sos_init();
    log_info("SOS button monitoring started");

    static uint32_t s_gps_timestamp = 0;

    for (;;) {
        sos_task();

        if (sos_is_long_press()) {
            log_error("SOS ALERT TRIGGERED!");

            sos_alert_t alert;
            alert.lat = 0.0;
            alert.lon = 0.0;
            alert.timestamp = s_gps_timestamp;

            if (tasks_send_sos(&alert, pdMS_TO_TICKS(500))) {
                log_info("SOS alert sent to comm task");
            }

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

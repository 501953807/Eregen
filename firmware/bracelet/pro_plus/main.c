/*
 * Eregen (颐贞) - Pro+ Bracelet Main Entry Point
 * 在 Pro 固件基础上增加试纸电化学检测 + BLE 血压计 + 慢病管理任务。
 *
 * 宏 TARGET_PRO_PLUS 通过 CMake 定义，确保 Pro+ 专属代码与 Pro/Entry 隔离。
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
#include "../entry/common/crc16.h"
#include "../entry/common/ring_buffer.h"
#include "../entry/protocol/message_encode.h"
#include "../entry/protocol/heartbeat.h"

/* Pro-specific modules */
#include "../pro/board_pro.h"
#include "../pro/ecg_driver.h"
#include "../pro/display_amoled.h"
#include "../pro/gps_gnss.h"
#include "../pro/free_rtos_tasks.h"

/* Pro+ 专属模块 */
#ifdef TARGET_PRO_PLUS
#include "sensors/electrochemical.h"
#include "bt_peripheral/bp_device.h"
#include "app/chronic_manager.h"
#endif

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
    uint32_t uid_low  = *(volatile uint32_t *)(0x1FFFF7E8UL);
    uint32_t uid_mid  = *(volatile uint32_t *)(0x1FFFF7ECUL);
    snprintf(s_device_id, sizeof(s_device_id),
             "%s%04X", DEVICE_ID_PREFIX,
             (uint16_t)(uid_low ^ uid_mid));
    s_device_id[sizeof(s_device_id) - 1] = '\0';
}

/*
 * Main entry point.
 * Initializes Pro board + Pro+ modules, creates all tasks, starts scheduler.
 */
int main(void)
{
    /* Initialize Pro board hardware */
    board_pro_init_all();
    generate_device_id();

    /* Initialize logging */
    log_init();
    log_register_uart_tx(uart_tx);
    log_set_level(LOG_INFO);

    log_info("Eregen Pro+ Bracelet Firmware v1.0");
    log_info("Device ID: %s", s_device_id);
    log_info("Target: GD32E230C8T3 (Cortex-M4) - Pro+ Tier");
#ifdef TARGET_PRO_PLUS
    log_info("Features: ECG + AMOLED + GNSS + Electrochemical Strip + BLE BP");
#else
    log_info("Features: ECG + AMOLED + GNSS");
#endif
    log_info("");

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

    /* Initialize Pro+ specific modules */
#ifdef TARGET_PRO_PLUS
    if (!chronic_mgr_init()) {
        log_warn("Pro+: Chronic manager init failed (将继续运行基础功能)");
    } else {
        log_info("Pro+: Chronic manager initialized");
    }
#endif

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
        log_error("Pro+: Failed to create Pro tasks");
    }

    /* Create Pro+ chronic management task */
#ifdef TARGET_PRO_PLUS
    if (!chronic_mgr_create_task()) {
        log_warn("Pro+: Failed to create chronic manager task");
    }
#endif

    /* Start the scheduler */
    vTaskStartScheduler();

    /* Should never reach here */
    for (;;);
}

/* ----------------------------------------------------------------
 * Sensor Task
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
 * Communication Task
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
 * SOS Task
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

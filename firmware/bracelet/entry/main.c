/*
 * Eregen (颐贞) - Bracelet Firmware Entry Point
 * Target: GD32E230C8T3 (ARM Cortex-M4)
 * RTOS: FreeRTOS
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include "FreeRTOS.h"
#include "task.h"
#include "gpio.h"
#include "usart.h"
#include "timer.h"

/* Task priorities */
#define SENSOR_TASK_PRIORITY      (tskIDLE_PRIORITY + 3)
#define GPS_TASK_PRIORITY         (tskIDLE_PRIORITY + 2)
#define COMM_TASK_PRIORITY        (tskIDLE_PRIORITY + 4)
#define DISPLAY_TASK_PRIORITY     (tskIDLE_PRIORITY + 1)

/* Task stack sizes (words) */
#define SENSOR_TASK_STACK_SIZE    configMINIMAL_STACK_SIZE * 4
#define GPS_TASK_STACK_SIZE       configMINIMAL_STACK_SIZE * 4
#define COMM_TASK_STACK_SIZE      configMINIMAL_STACK_SIZE * 8
#define DISPLAY_TASK_STACK_SIZE   configMINIMAL_STACK_SIZE * 2

/* Device ID prefix: BR-XXXX */
#define DEVICE_ID_PREFIX          "BR-"

/* UART baud rate for debug output */
#define DEBUG_BAUD_RATE           115200

/* Forward declarations */
void vSensorTask(void *pvParameters);
void vGPSTask(void *pvParameters);
void vCommTask(void *pvParameters);
void vDisplayTask(void *pvParameters);
void board_init(void);
void vApplicationHeapStats(char *pcWriteBuffer, size_t xBufferLen);

/*
 * Main entry point.
 * Initializes the board, creates all tasks, starts the scheduler.
 */
int main(void)
{
    /* Initialize hardware */
    board_init();

    /* Create sensor sampling task */
    xTaskCreate(vSensorTask,
                "Sensor",
                SENSOR_TASK_STACK_SIZE,
                NULL,
                SENSOR_TASK_PRIORITY,
                NULL);

    /* Create GPS positioning task */
    xTaskCreate(vGPSTask,
                "GPS",
                GPS_TASK_STACK_SIZE,
                NULL,
                GPS_TASK_PRIORITY,
                NULL);

    /* Create communication task (MQTT/Cat1) */
    xTaskCreate(vCommTask,
                "Comm",
                COMM_TASK_STACK_SIZE,
                NULL,
                COMM_TASK_PRIORITY,
                NULL);

    /* Create display task (OLED) */
    xTaskCreate(vDisplayTask,
                "Display",
                DISPLAY_TASK_STACK_SIZE,
                NULL,
                DISPLAY_TASK_PRIORITY,
                NULL);

    /* Start the scheduler */
    vTaskStartScheduler();

    /* Should never reach here */
    for (;;);
}

/*
 * Sensor Task
 * Reads heart rate (PPG), SpO2, and accelerometer (ICM-42670-P).
 * Publishes health data via message queue to Comm task.
 */
void vSensorTask(void *pvParameters)
{
    (void)pvParameters;

    for (;;) {
        /* TODO: Initialize PPG sensor (GT320) */
        /* TODO: Initialize IMU sensor (ICM-42670-P) */
        /* TODO: Read heart rate, SpO2, step count */
        /* TODO: Send data to comm queue */

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/*
 * GPS Task
 * Reads location from GPS module (u-blox NEO-M9N or UGN-7345).
 * Publishes location data via message queue to Comm task.
 */
void vGPSTask(void *pvParameters)
{
    (void)pvParameters;

    for (;;) {
        /* TODO: Initialize GPS module */
        /* TODO: Read latitude, longitude, accuracy */
        /* TODO: Send data to comm queue */

        vTaskDelay(pdMS_TO_TICKS(5000));
    }
}

/*
 * Communication Task
 * Handles Cat1 cellular connectivity, MQTT publish/subscribe.
 * Receives data from other tasks via queues and publishes to cloud.
 * Subscribes to cloud commands and forwards to appropriate tasks.
 */
void vCommTask(void *pvParameters)
{
    (void)pvParameters;

    for (;;) {
        /* TODO: Initialize Cat1 module (L610-CM) */
        /* TODO: Connect to EMQX MQTT broker */
        /* TODO: Subscribe to device command topic */
        /* TODO: Publish heartbeat, health, location, SOS data */

        vTaskDelay(pdMS_TO_TICKS(10000));
    }
}

/*
 * Display Task
 * Controls OLED screen (SSD1306). Shows status, alerts, basic UI.
 */
void vDisplayTask(void *pvParameters)
{
    (void)pvParameters;

    for (;;) {
        /* TODO: Initialize OLED display (SSD1306) */
        /* TODO: Draw status bar, battery indicator */
        /* TODO: Render incoming alerts */

        vTaskDelay(pdMS_TO_TICKS(200));
    }
}

/*
 * Board initialization
 * Configures GPIO, UART (debug), I2C (sensors/OLED), SPI (GPS).
 */
void board_init(void)
{
    /* Configure debug UART */
    uart_init(DEBUG_BAUD_RATE);
    printf("Eregen Bracelet Firmware starting...\n");

    /* Configure GPIO for LEDs, buttons, SOS */
    gpio_init();

    /* Configure I2C for sensors and display */
    i2c_init();

    /* Configure SPI for GPS module */
    spi_init();

    /* Configure timers */
    timer_init();

    /* Configure interrupt priorities */
    NVIC_SetPriorityGrouping(4);

    printf("Board initialization complete.\n");
}

/*
 * FreeRTOS hook: heap statistics (called from debug task)
 */
void vApplicationHeapStats(char *pcWriteBuffer, size_t xBufferLen)
{
    (void)xBufferLen;
    /* TODO: Implement heap stats reporting */
    if (pcWriteBuffer) {
        snprintf(pcWriteBuffer, xBufferLen, "Heap stats not implemented");
    }
}

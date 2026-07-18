/*
 * Eregen (颐贞) - FreeRTOS Task Management Header
 * Task creation, message queues, inter-task communication
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef FREERTOS_TASKS_H
#define FREERTOS_TASKS_H

#include <stdint.h>
#include <stdbool.h>

/* Task priorities */
#define TASK_SENSOR_PRIORITY      (tskIDLE_PRIORITY + 3)
#define TASK_GPS_PRIORITY         (tskIDLE_PRIORITY + 2)
#define TASK_COMM_PRIORITY        (tskIDLE_PRIORITY + 4)
#define TASK_DISPLAY_PRIORITY     (tskIDLE_PRIORITY + 1)
#define TASK_SOS_PRIORITY         (tskIDLE_PRIORITY + 5)

/* Task stack sizes (words) */
#define TASK_SENSOR_STACK        configMINIMAL_STACK_SIZE * 4
#define TASK_GPS_STACK           configMINIMAL_STACK_SIZE * 4
#define TASK_COMM_STACK          configMINIMAL_STACK_SIZE * 8
#define TASK_DISPLAY_STACK       configMINIMAL_STACK_SIZE * 2
#define TASK_SOS_STACK           configMINIMAL_STACK_SIZE * 2

/* Message queue sizes */
#define QUEUE_HEALTH_SIZE        8U
#define QUEUE_LOCATION_SIZE      8U
#define QUEUE_SOS_SIZE           4U
#define QUEUE_DISPLAY_SIZE       16U

/* Health data message (matches protocol JSON) */
typedef struct {
    uint16_t hr;          /* Heart rate in BPM */
    uint8_t  spo2;        /* SpO2 percentage */
    uint32_t step_count;  /* Step count since last send */
} health_data_t;

/* Location data message (matches protocol JSON) */
typedef struct {
    double lat;           /* Latitude */
    double lon;           /* Longitude */
    uint8_t  accuracy;    /* Horizontal accuracy in meters */
    uint32_t timestamp;   /* UTC timestamp */
} location_data_t;

/* SOS alert message (matches protocol JSON) */
typedef struct {
    double lat;           /* Current latitude */
    double lon;           /* Current longitude */
    uint32_t timestamp;   /* UTC timestamp */
} sos_alert_t;

/* Display command message */
typedef struct {
    uint8_t  cmd;         /* Command type: 0=clear, 1=status, 2=alert */
    uint16_t color;       /* Text color (RGB565) */
    char     text[32];    /* Text to display */
} display_cmd_t;

/*
 * Initialize all FreeRTOS tasks and message queues.
 * Creates the task handles and queue objects.
 * @return true if all tasks and queues created successfully.
 */
bool tasks_init(void);

/*
 * Send health data to the communication task queue.
 * @param data Pointer to health_data_t to send.
 * @param timeout_ms Block timeout.
 * @return true if sent successfully.
 */
bool tasks_send_health(const health_data_t *data, uint32_t timeout_ms);

/*
 * Send location data to the communication task queue.
 * @param data Pointer to location_data_t to send.
 * @param timeout_ms Block timeout.
 * @return true if sent successfully.
 */
bool tasks_send_location(const location_data_t *data, uint32_t timeout_ms);

/*
 * Send SOS alert to the communication task queue (highest priority).
 * @param data Pointer to sos_alert_t to send.
 * @param timeout_ms Block timeout.
 * @return true if sent successfully.
 */
bool tasks_send_sos(const sos_alert_t *data, uint32_t timeout_ms);

/*
 * Send display command to the display task queue.
 * @param cmd Pointer to display_cmd_t to send.
 * @param timeout_ms Block timeout.
 * @return true if sent successfully.
 */
bool tasks_send_display(const display_cmd_t *cmd, uint32_t timeout_ms);

/*
 * Get the handle to the communication task (for direct access if needed).
 * @return Task handle, or NULL if not initialized.
 */
void* tasks_get_comm_handle(void);

/**
 * Get the handle to the health data message queue.
 * @return Queue handle, or NULL if not initialized.
 */
QueueHandle_t tasks_get_health_queue(void);

/**
 * Get the handle to the location data message queue.
 * @return Queue handle, or NULL if not initialized.
 */
QueueHandle_t tasks_get_location_queue(void);

/**
 * Get the handle to the SOS alert message queue.
 * @return Queue handle, or NULL if not initialized.
 */
QueueHandle_t tasks_get_sos_queue(void);

#endif /* FREERTOS_TASKS_H */

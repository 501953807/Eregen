/*
 * Eregen (颐贞) - FreeRTOS Task Management Implementation
 * Task creation, message queues, inter-task communication
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "free_rtos_tasks.h"
#include "FreeRTOS.h"
#include "task.h"
#include "queue.h"

/* Queue handles */
static QueueHandle_t s_health_queue = NULL;
static QueueHandle_t s_location_queue = NULL;
static QueueHandle_t s_sos_queue = NULL;
static QueueHandle_t s_display_queue = NULL;

/* Task handle for comm task */
static TaskHandle_t s_comm_task_handle = NULL;

/*
 * Initialize all FreeRTOS tasks and message queues.
 */
bool tasks_init(void)
{
    /* Create message queues */
    s_health_queue = xQueueCreate(QUEUE_HEALTH_SIZE, sizeof(health_data_t));
    s_location_queue = xQueueCreate(QUEUE_LOCATION_SIZE, sizeof(location_data_t));
    s_sos_queue = xQueueCreate(QUEUE_SOS_SIZE, sizeof(sos_alert_t));
    s_display_queue = xQueueCreate(QUEUE_DISPLAY_SIZE, sizeof(display_cmd_t));

    if (!s_health_queue || !s_location_queue || !s_sos_queue || !s_display_queue) {
        return false;
    }

    /* Tasks are created in main.c via xTaskCreate.
     * This function only creates the shared communication infrastructure.
     * The actual task creation happens in main().
     */
    return true;
}

/*
 * Send health data to the communication task queue.
 */
bool tasks_send_health(const health_data_t *data, uint32_t timeout_ms)
{
    if (!data || !s_health_queue) {
        return false;
    }

    BaseType_t ret = xQueueSend(s_health_queue, data,
                                pdMS_TO_TICKS(timeout_ms));
    return (ret == pdPASS);
}

/*
 * Send location data to the communication task queue.
 */
bool tasks_send_location(const location_data_t *data, uint32_t timeout_ms)
{
    if (!data || !s_location_queue) {
        return false;
    }

    BaseType_t ret = xQueueSend(s_location_queue, data,
                                pdMS_TO_TICKS(timeout_ms));
    return (ret == pdPASS);
}

/*
 * Send SOS alert to the communication task queue.
 * Uses back-to-back send for highest priority.
 */
bool tasks_send_sos(const sos_alert_t *data, uint32_t timeout_ms)
{
    if (!data || !s_sos_queue) {
        return false;
    }

    /* Use xQueueSendToBack for SOS to ensure immediate delivery */
    BaseType_t ret = xQueueSendToBack(s_sos_queue, data,
                                      pdMS_TO_TICKS(timeout_ms));
    return (ret == pdPASS);
}

/*
 * Send display command to the display task queue.
 */
bool tasks_send_display(const display_cmd_t *cmd, uint32_t timeout_ms)
{
    if (!cmd || !s_display_queue) {
        return false;
    }

    BaseType_t ret = xQueueSend(s_display_queue, cmd,
                                pdMS_TO_TICKS(timeout_ms));
    return (ret == pdPASS);
}

/*
 * Get the handle to the communication task.
 */
void* tasks_get_comm_handle(void)
{
    return (void*)s_comm_task_handle;
}

/*
 * Set the comm task handle (called when task is created).
 */
void tasks_set_comm_handle(TaskHandle_t handle)
{
    s_comm_task_handle = handle;
}

QueueHandle_t tasks_get_health_queue(void)
{
    return s_health_queue;
}

QueueHandle_t tasks_get_location_queue(void)
{
    return s_location_queue;
}

QueueHandle_t tasks_get_sos_queue(void)
{
    return s_sos_queue;
}

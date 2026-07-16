/*
 * Eregen (颐贞) - Pro Tier FreeRTOS Task Definitions
 * Task creation and management for ECG, AMOLED display, and GNSS modules.
 *
 * These tasks run on top of the base entry-level task infrastructure
 * and add Pro-specific functionality: ECG monitoring, AMOLED UI,
 * and multi-constellation GNSS positioning.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef PRO_FREERTOS_TASKS_H
#define PRO_FREERTOS_TASKS_H

#include <stdint.h>
#include <stdbool.h>
#include "FreeRTOS.h"
#include "task.h"

/* ----------------------------------------------------------------
 * Task priorities (higher number = higher priority)
 * ---------------------------------------------------------------- */

/* ECG task: high priority for real-time 200Hz sampling */
#define TASK_PRO_ECG_PRIORITY      (tskIDLE_PRIORITY + 6)
#define TASK_PRO_ECG_STACK         (configMINIMAL_STACK_SIZE * 6)

/* AMOLED display task: medium priority (UI updates are not time-critical) */
#define TASK_PRO_AMOLED_PRIORITY   (tskIDLE_PRIORITY + 2)
#define TASK_PRO_AMOLED_STACK      (configMINIMAL_STACK_SIZE * 4)

/* GNSS task: same priority as entry GPS but with enhanced parsing */
#define TASK_PRO_GNSS_PRIORITY     (tskIDLE_PRIORITY + 3)
#define TASK_PRO_GNSS_STACK        (configMINIMAL_STACK_SIZE * 4)

/* ECG health alert queue size */
#define QUEUE_ECG_SIZE             16U

/* ECG arrhythmia alert queue */
#define QUEUE_ARRITHMIA_SIZE       4U

/* ----------------------------------------------------------------
 * ECG health data message (extends entry health_data_t)
 * ---------------------------------------------------------------- */
typedef struct {
    uint16_t hr;              /* Heart rate from ECG R-R interval (BPM) */
    uint8_t  spo2;            /* SpO2 from PPG (if available) */
    uint32_t step_count;      /* Step count */
    int32_t  ecg_peak_uv;     /* Latest ECG R-wave peak amplitude (uV) */
    float    ecg_rr_interval; /* Current R-R interval in ms */
    bool     ecg_valid;       /* True if ECG data is valid */
} pro_health_data_t;

/* ----------------------------------------------------------------
 * Arrhythmia alert message
 * ---------------------------------------------------------------- */
typedef struct {
    bool afib_detected;       /* True if AFib detected */
    float rr_stddev_ms;       /* RR interval std deviation */
    uint32_t timestamp;       /* Alert timestamp */
    uint8_t severity;         /* 1=warning, 2=alert, 3=critical */
} pro_arrhythmia_alert_t;

/* ----------------------------------------------------------------
 * Pro task initialization
 * ---------------------------------------------------------------- */

/**
 * Initialize Pro-specific FreeRTOS tasks and queues.
 * Creates message queues for ECG data and arrhythmia alerts.
 * @return true if all queues created successfully.
 */
bool pro_tasks_init(void);

/**
 * Send ECG health data to the communication pipeline.
 * @param data Pointer to pro_health_data_t.
 * @param timeout_ms Block timeout.
 * @return true if sent successfully.
 */
bool pro_tasks_send_ecg_health(const pro_health_data_t *data, uint32_t timeout_ms);

/**
 * Send arrhythmia alert to the communication pipeline.
 * @param alert Pointer to pro_arrhythmia_alert_t.
 * @param timeout_ms Block timeout.
 * @return true if sent successfully.
 */
bool pro_tasks_send_arrhythmia_alert(const pro_arrhythmia_alert_t *alert,
                                     uint32_t timeout_ms);

/**
 * Create the Pro-specific tasks (ECG, AMOLED, GNSS).
 * Call after pro_tasks_init() and after base tasks are created.
 * @return true if all tasks created successfully.
 */
bool pro_tasks_create(void);

#endif /* PRO_FREERTOS_TASKS_H */

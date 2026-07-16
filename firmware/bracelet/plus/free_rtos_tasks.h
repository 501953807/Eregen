/*
 * Eregen (颐贞) - FreeRTOS Task Management Header (Plus Tier)
 * Extended task definitions for geofence, fall detection, battery optimization,
 * and BLE pairing — built on top of entry-tier base tasks.
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

#ifndef FREERTOS_TASKS_PLUS_H
#define FREERTOS_TASKS_PLUS_H

#include <stdint.h>
#include <stdbool.h>
#include "free_rtos_tasks.h"  /* Base entry-tier types */

/* ---- Plus-tier task priorities ---- */

#define TASK_GEOFENCE_PRIORITY  (tskIDLE_PRIORITY + 3)
#define TASK_FALL_PRIORITY      (tskIDLE_PRIORITY + 5)  /* High — safety critical */
#define TASK_BATT_OPT_PRIORITY  (tskIDLE_PRIORITY + 1)
#define TASK_BLE_PRIORITY       (tskIDLE_PRIORITY + 2)

/* ---- Plus-tier task stack sizes (words) ---- */

#define TASK_GEOFENCE_STACK     configMINIMAL_STACK_SIZE * 6
#define TASK_FALL_STACK         configMINIMAL_STACK_SIZE * 4
#define TASK_BATT_OPT_STACK     configMINIMAL_STACK_SIZE * 3
#define TASK_BLE_STACK          configMINIMAL_STACK_SIZE * 6

/* ---- Message queue sizes for plus-tier tasks ---- */

#define QUEUE_GEOFENCE_SIZE     4U
#define QUEUE_FALL_SIZE         4U
#define QUEUE_BATT_OPT_SIZE     8U
#define QUEUE_BLE_SIZE          8U

/* ---- Geofence alert message ---- */

typedef struct {
    uint8_t  zone_id;       /* Which zone was exited */
    double   lat;           /* Last known latitude */
    double   lon;           /* Last known longitude */
    uint32_t timestamp;     /* UTC timestamp */
} geofence_alert_t;

/* ---- Fall alert message ---- */

typedef struct {
    double   lat;           /* Last known latitude */
    double   lon;           /* Last known longitude */
    float    confidence;    /* Fall confidence score */
    uint32_t timestamp;     /* UTC timestamp */
} fall_alert_t;

/* ---- Battery optimization message ---- */

typedef struct {
    uint16_t gps_interval_s;   /* New GPS interval */
    uint16_t ppg_interval_s;   /* New PPG interval */
    uint8_t  tier;             /* Optimization tier */
} batt_opt_msg_t;

/*
 * Initialize all plus-tier FreeRTOS tasks and message queues.
 * Must be called AFTER base tasks_init().
 * Creates task handles and queue objects for geofence, fall_detect,
 * battery_opt, and ble_pair tasks.
 * @return true if all plus-tier tasks and queues created successfully.
 */
bool tasks_plus_init(void);

/*
 * Send a geofence alert to the communication task.
 * @param alert Pointer to geofence_alert_t to send.
 * @param timeout_ms Block timeout.
 * @return true if sent successfully.
 */
bool tasks_send_geofence_alert(const geofence_alert_t *alert, uint32_t timeout_ms);

/*
 * Send a fall alert to the communication task.
 * @param alert Pointer to fall_alert_t to send.
 * @param timeout_ms Block timeout.
 * @return true if sent successfully.
 */
bool tasks_send_fall_alert(const fall_alert_t *alert, uint32_t timeout_ms);

/*
 * Broadcast battery optimization config change to interested tasks.
 * @param msg Pointer to batt_opt_msg_t with new intervals.
 * @param timeout_ms Block timeout.
 * @return true if broadcast successfully.
 */
bool tasks_broadcast_battery_opt(const batt_opt_msg_t *msg, uint32_t timeout_ms);

/*
 * Get handles to plus-tier message queues (for direct access if needed).
 */
void* tasks_get_geofence_queue(void);
void* tasks_get_fall_queue(void);
void* tasks_get_batt_opt_queue(void);
void* tasks_get_ble_queue(void);

#endif /* FREERTOS_TASKS_PLUS_H */

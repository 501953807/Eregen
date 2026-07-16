/*
 * Eregen (颐贞) - FreeRTOS Task Management Implementation (Plus Tier)
 * Creates and manages RTOS tasks for geofence monitoring, fall detection,
 * battery optimization, and BLE pairing. Integrates with entry-tier base tasks.
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

#include "free_rtos_tasks.h"
#include "FreeRTOS.h"
#include "task.h"
#include "queue.h"
#include "log.h"

/* ---- Plus-tier message queues ---- */

static QueueHandle_t s_geofence_queue = NULL;
static QueueHandle_t s_fall_queue = NULL;
static QueueHandle_t s_batt_opt_queue = NULL;
static QueueHandle_t s_ble_queue = NULL;

/* ---- Forward declarations for task functions ---- */

static void vGeofenceTask(void *pvParameters);
static void vFallDetectTask(void *pvParameters);
static void vBatteryOptTask(void *pvParameters);
static void vBLEPairTask(void *pvParameters);

/* ---- Base task handle setter (from entry tier) ---- */

extern void tasks_set_comm_handle(void *handle);

/*
 * Initialize all plus-tier tasks and queues.
 * Must be called AFTER base tasks_init().
 */
bool tasks_plus_init(void)
{
    /* Create message queues. */
    s_geofence_queue = xQueueCreate(QUEUE_GEOFENCE_SIZE, sizeof(geofence_alert_t));
    s_fall_queue = xQueueCreate(QUEUE_FALL_SIZE, sizeof(fall_alert_t));
    s_batt_opt_queue = xQueueCreate(QUEUE_BATT_OPT_SIZE, sizeof(batt_opt_msg_t));
    s_ble_queue = xQueueCreate(QUEUE_BLE_SIZE, sizeof(uint32_t));  /* PIN code queue */

    if (!s_geofence_queue || !s_fall_queue ||
        !s_batt_opt_queue || !s_ble_queue) {
        log_error("Failed to create plus-tier message queues");
        return false;
    }

    /* Create geofence monitoring task. */
    TaskHandle_t geofence_handle = NULL;
    xTaskCreate(vGeofenceTask,
                "GeoFence",
                TASK_GEOFENCE_STACK,
                NULL,
                TASK_GEOFENCE_PRIORITY,
                &geofence_handle);
    (void)geofence_handle;

    /* Create fall detection task. */
    TaskHandle_t fall_handle = NULL;
    xTaskCreate(vFallDetectTask,
                "FallDetect",
                TASK_FALL_STACK,
                NULL,
                TASK_FALL_PRIORITY,
                &fall_handle);
    (void)fall_handle;

    /* Create battery optimization task. */
    TaskHandle_t batt_handle = NULL;
    xTaskCreate(vBatteryOptTask,
                "BattOpt",
                TASK_BATT_OPT_STACK,
                NULL,
                TASK_BATT_OPT_PRIORITY,
                &batt_handle);
    (void)batt_handle;

    /* Create BLE pairing task. */
    TaskHandle_t ble_handle = NULL;
    xTaskCreate(vBLEPairTask,
                "BLEPair",
                TASK_BLE_STACK,
                NULL,
                TASK_BLE_PRIORITY,
                &ble_handle);
    (void)ble_handle;

    log_info("Plus-tier tasks initialized: GeoFence, FallDetect, BattOpt, BLEPair");
    return true;
}

bool tasks_send_geofence_alert(const geofence_alert_t *alert, uint32_t timeout_ms)
{
    if (!alert || !s_geofence_queue) {
        return false;
    }
    return (xQueueSend(s_geofence_queue, alert, pdMS_TO_TICKS(timeout_ms)) == pdPASS);
}

bool tasks_send_fall_alert(const fall_alert_t *alert, uint32_t timeout_ms)
{
    if (!alert || !s_fall_queue) {
        return false;
    }
    return (xQueueSend(s_fall_queue, alert, pdMS_TO_TICKS(timeout_ms)) == pdPASS);
}

bool tasks_broadcast_battery_opt(const batt_opt_msg_t *msg, uint32_t timeout_ms)
{
    if (!msg || !s_batt_opt_queue) {
        return false;
    }
    return (xQueueSend(s_batt_opt_queue, msg, pdMS_TO_TICKS(timeout_ms)) == pdPASS);
}

void* tasks_get_geofence_queue(void)
{
    return (void *)s_geofence_queue;
}

void* tasks_get_fall_queue(void)
{
    return (void *)s_fall_queue;
}

void* tasks_get_batt_opt_queue(void)
{
    return (void *)s_batt_opt_queue;
}

void* tasks_get_ble_queue(void)
{
    return (void *)s_ble_queue;
}

/* ---- Task implementations ---- */

/**
 * Geofence Monitoring Task
 * Receives location updates from the base comm task, checks against
 * configured zones, and sends alerts when the elder leaves a safe zone.
 */
static void vGeofenceTask(void *pvParameters)
{
    (void)pvParameters;

    log_info("Geofence task started");

    /* Location data from base task. */
    location_data_t loc_msg;
    geofence_alert_t alert_msg;
    uint32_t last_alert_tick = 0;

    for (;;) {
        /* Wait for location data from the communication task. */
        void *comm_q = tasks_get_comm_handle();
        if (comm_q && xQueueReceive((QueueHandle_t)comm_q, &loc_msg,
                                     pdMS_TO_TICKS(1000)) == pdPASS) {
            /* Check position against geofence zones. */
            /* Note: geofence_check_position is called here; the actual
             * zone data comes from geofence_manager module. */

            /* Placeholder: in production, call geofence_check_position(loc_msg.lat, ...) */
            (void)last_alert_tick;

            /* Send alert if zone violation detected. */
            alert_msg.zone_id = 0;
            alert_msg.lat = loc_msg.lat;
            alert_msg.lon = loc_msg.lon;
            alert_msg.timestamp = loc_msg.timestamp;

            if (tasks_send_geofence_alert(&alert_msg, pdMS_TO_TICKS(100))) {
                log_warn("Geofence alert: zone %u, lat=%.4f, lon=%.4f",
                         alert_msg.zone_id, alert_msg.lat, alert_msg.lon);
            }
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}

/**
 * Fall Detection Task
 * Receives IMU samples, runs the fall detection algorithm, and raises
 * alerts when a fall is confirmed.
 */
static void vFallDetectTask(void *pvParameters)
{
    (void)pvParameters;

    log_info("Fall detection task started");

    /* IMU data from base sensor task. */
    imu_data_t imu_msg;
    fall_alert_t alert_msg;
    uint32_t tick_counter = 0;

    for (;;) {
        /* Receive IMU data from sensor task queue. */
        void *comm_q = tasks_get_comm_handle();
        if (comm_q) {
            /* Poll for IMU data — in production this would be a dedicated queue. */
            if (xQueueReceive((QueueHandle_t)comm_q, &imu_msg, pdMS_TO_TICKS(20)) == pdPASS) {
                /* Feed sample to fall detection algorithm. */
                fall_sample_t sample;
                sample.ax = imu_msg.ax;
                sample.ay = imu_msg.ay;
                sample.az = imu_msg.az;
                sample.gx = imu_msg.gx;
                sample.gy = imu_msg.gy;
                sample.gz = imu_msg.gz;
                sample.tick = xTaskGetTickCount();

                fall_event_t event = fall_detect_feed(&sample);

                if (event.alarm_ready) {
                    alert_msg.lat = 0.0;   /* Would get from GPS */
                    alert_msg.lon = 0.0;
                    alert_msg.confidence = event.confidence;
                    alert_msg.timestamp = (uint32_t)time(NULL);

                    if (tasks_send_fall_alert(&alert_msg, pdMS_TO_TICKS(200))) {
                        log_error("FALL ALERT: confidence=%.2f, consecutive=%u",
                                  event.confidence, event.consecutive);
                    }
                }

                tick_counter++;
            }
        }

        /* Periodic evaluation tick for time-based fall detection. */
        if (tick_counter >= (1000 / FALL_DETECT_TICK_MS)) {
            tick_counter = 0;
            fall_detect_run();
        }

        vTaskDelay(pdMS_TO_TICKS(FALL_DETECT_TICK_MS));
    }
}

/**
 * Battery Optimization Task
 * Monitors battery level and adjusts GPS/PPG sampling rates dynamically.
 * Broadcasts new rates to other tasks via message queue.
 */
static void vBatteryOptTask(void *pvParameters)
{
    (void)pvParameters;

    log_info("Battery optimization task started");

    batt_opt_msg_t opt_msg;

    for (;;) {
        vTaskDelay(pdMS_TO_TICKS(1000));  /* Tick every second. */

        /* Run the optimizer tick. */
        bool changed = battery_optimizer_tick();

        if (changed) {
            optimizer_config_t cfg;
            if (battery_optimizer_get_config(&cfg)) {
                opt_msg.gps_interval_s = cfg.gps_interval_s;
                opt_msg.ppg_interval_s = cfg.ppg_interval_s;
                opt_msg.tier = cfg.tier;

                if (tasks_broadcast_battery_opt(&opt_msg, pdMS_TO_TICKS(100))) {
                    log_info("BATTERY OPT: tier=%u, GPS=%us, PPG=%us",
                             opt_msg.tier, opt_msg.gps_interval_s, opt_msg.ppg_interval_s);
                }
            }
        }
    }
}

/**
 * BLE Pairing Task
 * Manages BLE advertising state machine for initial device provisioning.
 * Handles PIN verification and credential reception from family APP.
 */
static void vBLEPairTask(void *pvParameters)
{
    (void)pvParameters;

    log_info("BLE pairing task started");

    /* Start advertising immediately if not yet provisioned. */
    if (!ble_pair_is_provisioned()) {
        ble_pair_start_advertising();
        log_info("BLE advertising started, PIN: %06lu",
                 (unsigned long)s_state.pin_code);
    }

    uint32_t adv_tick = 0;

    for (;;) {
        vTaskDelay(pdMS_TO_TICKS(1000));
        adv_tick++;

        /* Periodic BLE tick for timer management. */
        ble_pair_tick();

        /* If provisioned, stop advertising. */
        if (ble_pair_is_provisioned()) {
            ble_pair_stop_advertising();
            log_info("Device provisioned — BLE advertising stopped");
            /* Transition to normal operation. */
            vTaskDelete(NULL);
            return;
        }

        /* Restart advertising if it was stopped unexpectedly. */
        if (adv_tick > 3600) {  /* 1 hour */
            adv_tick = 0;
            if (s_state.state == BLE_STATE_IDLE) {
                ble_pair_start_advertising();
            }
        }
    }
}

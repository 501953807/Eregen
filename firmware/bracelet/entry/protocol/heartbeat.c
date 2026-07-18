/*
 * Eregen (颐贞) - Heartbeat Module Implementation
 * Publishes heartbeat messages every 5 minutes via Cat1 AT interface.
 * Uses FreeRTOS timer for periodic triggering on embedded target.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>

#ifndef TEST_MODE
/* Embedded mode headers */
#include "heartbeat.h"
#include "../common/log.h"
#include "../common/crc16.h"
#include "cat1_at.h"
#include "battery_adc.h"
#include "FreeRTOS.h"
#include "timers.h"

static TimerHandle_t s_heartbeat_timer = NULL;
#endif

static volatile int s_running = 0;

/* Device ID (set by main.c) */
extern char s_device_id[];

#ifndef TEST_MODE
/* Heartbeat interval: 5 minutes */
#define HEARTBEAT_INTERVAL_MS    (5 * 60 * 1000)

static void heartbeat_timer_callback(TimerHandle_t xTimer)
{
    (void)xTimer;

    battery_status_t batt;
    batt.voltage_mv = 0;
    batt.percent = 0;
    batt = battery_get_status();

    char json_buf[128];
    int len = snprintf(json_buf, sizeof(json_buf),
        "{\"type\":\"heartbeat\",\"dev_id\":\"%s\",\"bat\":%u,\"ts\":%lu}",
        s_device_id, (unsigned)batt.percent,
        (unsigned long)s_gps_timestamp);

    if (len <= 0 || len >= (int)sizeof(json_buf)) {
        log_error("Heartbeat JSON build failed");
        return;
    }

    uint16_t crc = crc16_calc((const uint8_t *)json_buf, (uint16_t)len);
    json_buf[len]     = (char)((crc >> 8) & 0xFF);
    json_buf[len + 1] = (char)(crc & 0xFF);
    json_buf[len + 2] = '\0';

    cat1_mqtt_publish("eregen/device/bracelet/BR-CLOUD/up",
        (const uint8_t *)json_buf, (uint16_t)(len + 2));
    log_info("HEARTBEAT PUBLISHED: %s, bat=%u%%", s_device_id, (unsigned)batt.percent);
}

void heartbeat_start(void)
{
    if (s_running) {
        return;
    }

    s_heartbeat_timer = xTimerCreate(
        "hb_timer",
        pdMS_TO_TICKS(HEARTBEAT_INTERVAL_MS),
        pdTRUE,
        NULL,
        heartbeat_timer_callback
    );

    if (s_heartbeat_timer) {
        xTimerStart(s_heartbeat_timer, 0);
        s_running = 1;
        log_info("Heartbeat started (interval=%lu ms)",
                 (unsigned long)HEARTBEAT_INTERVAL_MS);
    } else {
        log_error("Failed to create heartbeat timer");
    }
}

void heartbeat_stop(void)
{
    if (!s_running) {
        return;
    }

    if (s_heartbeat_timer) {
        xTimerStop(s_heartbeat_timer, 0);
        s_running = 0;
        log_info("Heartbeat stopped");
    }
}
#else
/* Host test mode: simple flag-based implementation */

void heartbeat_start(void)
{
    if (s_running) {
        return;
    }
    s_running = 1;
}

void heartbeat_stop(void)
{
    s_running = 0;
}
#endif /* TEST_MODE */

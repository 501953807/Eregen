/*
 * Eregen (颐贞) - Schedule Engine Implementation
 * Manages medication reminder scheduling with time comparison.
 * Supports daily recurring rules sorted by time.
 * Works with med_rule_parser for rule data and nvs_store for persistence.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "schedule_engine.h"

#include <string.h>
#include <stdio.h>

#ifdef TEST_MODE
#include <stdlib.h>
#else
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "nvs_flash.h"
#endif

/* Log tag */
#ifdef TEST_MODE
static const char *TAG = "sched";
#define ESP_LOGI(...) printf(__VA_ARGS__)
#define ESP_LOGW(...) fprintf(stderr, __VA_ARGS__)
#define ESP_OK            0
#define ESP_ERR_NO_MEM   (-5)
#define ESP_FAIL          (-1)
#else
static const char *TAG = "sched";
#endif

/* ---- Internal state ---- */

static uint8_t    s_last_triggered_idx = 0xFF;
static bool       s_today_acknowledged = false;
static uint32_t   s_current_day = 0;

/**
 * Convert hour:minute to minutes since midnight.
 */
static inline uint16_t time_to_minutes(uint8_t hour, uint8_t minute)
{
    return (uint16_t)(hour * 60 + minute);
}

/**
 * Get current time in minutes since midnight.
 * In TEST_MODE, derives from a mock epoch set by tests.
 */
static uint16_t get_current_minutes(void)
{
#ifdef TEST_MODE
    extern uint32_t s_schedule_mock_epoch;
    uint32_t hour = (s_schedule_mock_epoch / 3600) % 24;
    uint32_t minute = (s_schedule_mock_epoch / 60) % 60;
    return (uint16_t)(hour * 60 + minute);
#else
    struct timeval tv;
    gettimeofday(&tv, NULL);
    time_t epoch = tv.tv_sec;
    struct tm tm_info;
    localtime_r(&epoch, &tm_info);
    return (uint16_t)(tm_info.tm_hour * 60 + tm_info.tm_min);
#endif
}

/**
 * Get current day number from epoch seconds.
 */
static uint32_t get_current_day(void)
{
#ifdef TEST_MODE
    extern uint32_t s_schedule_mock_epoch;
    return (uint32_t)(s_schedule_mock_epoch / 86400);
#else
    time_t epoch = time(NULL);
    return (uint32_t)(epoch / 86400);
#endif
}

void schedule_engine_init(void)
{
    s_last_triggered_idx = 0xFF;
    s_today_acknowledged = false;
    s_current_day = get_current_day();

#ifndef TEST_MODE
    /* Try loading rules from NVS storage via nvs_load_rules */
    med_rule_t loaded[MED_RULE_PARSER_MAX_RULES];
    uint8_t count = 0;
    if (nvs_load_rules(loaded, &count) && count > 0) {
        med_rule_load_raw(loaded, count);
        ESP_LOGI(TAG, "Loaded %d rules from NVS", count);
    }
#endif
}

void schedule_engine_reload(void)
{
    /* Persist current internal rules to NVS */
    uint8_t count = med_rule_count();
    if (count > 0) {
        med_rule_t rules_buf[MED_RULE_PARSER_MAX_RULES];
        for (uint8_t i = 0; i < count; i++) {
            const med_rule_t *r = med_rule_get(i);
            if (r) memcpy(&rules_buf[i], r, sizeof(med_rule_t));
        }
        nvs_save_rules(rules_buf, count);
    }
}

uint32_t schedule_next_trigger(void)
{
    /* Check if day has rolled over — reset acknowledgment state */
    uint32_t day = get_current_day();
    if (day != s_current_day) {
        s_current_day = day;
        s_today_acknowledged = false;
        s_last_triggered_idx = 0xFF;
    }

    if (s_today_acknowledged)
        return UINT32_MAX;

    uint16_t current_min = get_current_minutes();
    uint16_t closest = 0xFFFF;

    uint8_t count = med_rule_count();
    for (uint8_t i = 0; i < count; i++) {
        const med_rule_t *rule = med_rule_get(i);
        if (!rule || !rule->enabled)
            continue;

        uint16_t rule_min = time_to_minutes(rule->hour, rule->minute);

        if (rule_min >= current_min && rule_min < closest) {
            closest = rule_min;
        }
    }

    if (closest == 0xFFFF)
        return UINT32_MAX;

    uint16_t diff_min = closest - current_min;
    return (uint32_t)diff_min * 60;
}

bool schedule_check_triggered(void)
{
    /* Check if day has rolled over */
    uint32_t day = get_current_day();
    if (day != s_current_day) {
        s_current_day = day;
        s_today_acknowledged = false;
        s_last_triggered_idx = 0xFF;
    }

    if (s_today_acknowledged)
        return false;

    uint8_t count = med_rule_count();
    if (count == 0)
        return false;

    uint16_t current_min = get_current_minutes();
    uint16_t closest = 0xFFFF;
    int closest_idx = -1;

    for (uint8_t i = 0; i < count; i++) {
        const med_rule_t *rule = med_rule_get(i);
        if (!rule || !rule->enabled)
            continue;

        uint16_t rule_min = time_to_minutes(rule->hour, rule->minute);

        if (rule_min >= current_min && rule_min < closest) {
            closest = rule_min;
            closest_idx = (int)i;
        }
    }

    if (closest_idx < 0)
        return false;

    if (current_min >= closest) {
        s_last_triggered_idx = (uint8_t)closest_idx;
        return true;
    }

    return false;
}

void schedule_acknowledge(void)
{
    s_today_acknowledged = true;
    ESP_LOGI(TAG, "Reminder acknowledged, index=%d", (int)s_last_triggered_idx);
}

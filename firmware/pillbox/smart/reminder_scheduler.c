/*
 * Eregen (颐贞) - Medication Reminder Scheduler Implementation
 * Smart pillbox tier — NVS-persisted scheduling with time comparison
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "reminder_scheduler.h"

#include <string.h>
#include <stdio.h>

#ifdef TEST_MODE
#include <stdlib.h>
#else
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "nvs_flash.h"
#include "esp_log.h"
#endif

/* NVS key for rules */
#define NVS_KEY_RULES         "med_rules"
#define NVS_KEY_COUNT         "rule_count"

/* RTC time source (from ESP32-C3 internal timer or external) */
static uint32_t s_current_epoch = 0;

/* Internal rule storage */
static reminder_rule_t s_rules[SCHEDULER_MAX_RULES];
static uint8_t s_rule_count = 0;

/* Whether reminders are currently paused */
static bool s_paused = false;

#ifdef TEST_MODE
static const char *TAG = "sched";
#define ESP_LOGI(...) printf(__VA_ARGS__)
#define ESP_LOGW(...) fprintf(stderr, __VA_ARGS__)
#define ESP_LOGE(...) fprintf(stderr, __VA_ARGS__)
#define ESP_OK            0
#define ESP_ERR_NO_MEM   (-5)
#define ESP_ERR_NOT_FOUND (-6)
#define ESP_ERR_INVALID_ARG (-11)
#define ESP_ERR_INVALID_STATE (-12)
#define ESP_FAIL          (-1)
#else
static const char *TAG = "sched";
#endif

/**
 * Convert HH:MM to minutes since midnight.
 */
static uint16_t time_to_minutes(const reminder_time_t *t)
{
    return (uint16_t)(t->hour * 60 + t->minute);
}

/**
 * Get current time in minutes since midnight from RTC.
 * In production this reads ESP32-C3 RTC clock or NTP-synced time.
 * For testing, s_current_epoch is used as a mock source.
 */
static uint16_t get_current_minutes(void)
{
#ifdef TEST_MODE
    /* In test mode, derive minutes from epoch for reproducibility */
    uint32_t hour = (s_current_epoch / 3600) % 24;
    uint32_t minute = (s_current_epoch / 60) % 60;
    return (uint16_t)(hour * 60 + minute);
#else
    /* Real mode: read RTC time via system clock */
    struct timeval tv;
    gettimeofday(&tv, NULL);
    time_t epoch = tv.tv_sec;
    struct tm tm_info;
    localtime_r(&epoch, &tm_info);
    return (uint16_t)(tm_info.tm_hour * 60 + tm_info.tm_min);
#endif
}

esp_err_t scheduler_init(void)
{
    memset(s_rules, 0, sizeof(s_rules));
    s_rule_count = 0;
    s_paused = false;

#ifndef TEST_MODE
    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READONLY, &handle);
    if (ret != ESP_OK) {
        ESP_LOGW(TAG, "NVS open failed, starting with empty schedule");
        return ESP_OK; /* Not fatal */
    }

    /* Read rule count */
    uint8_t count = 0;
    ret = nvs_get_u8(handle, NVS_KEY_COUNT, &count);
    if (ret == ESP_OK && count > 0 && count <= SCHEDULER_MAX_RULES) {
        /* Read rules blob */
        size_t needed_size = count * sizeof(reminder_rule_t);
        uint8_t *buf = malloc(needed_size);
        if (buf != NULL) {
            ret = nvs_get_blob(handle, NVS_KEY_RULES, buf, &needed_size);
            if (ret == ESP_OK) {
                memcpy(s_rules, buf, needed_size);
                s_rule_count = count;
                ESP_LOGI(TAG, "Loaded %d rules from NVS", count);
            }
            free(buf);
        }
    }

    nvs_close(handle);
#endif

    return ESP_OK;
}

esp_err_t scheduler_add_rule(const reminder_rule_t *rule)
{
    if (rule == NULL || s_rule_count >= SCHEDULER_MAX_RULES)
        return (s_rule_count >= SCHEDULER_MAX_RULES) ? ESP_ERR_NO_MEM : ESP_ERR_INVALID_ARG;

    memcpy(&s_rules[s_rule_count], rule, sizeof(reminder_rule_t));
    s_rules[s_rule_count].enabled = true;
    s_rule_count++;

#ifndef TEST_MODE
    /* Persist to NVS */
    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READWRITE, &handle);
    if (ret == ESP_OK) {
        nvs_set_u8(handle, NVS_KEY_COUNT, s_rule_count);
        nvs_set_blob(handle, NVS_KEY_RULES, s_rules,
                     s_rule_count * sizeof(reminder_rule_t));
        nvs_commit(handle);
        nvs_close(handle);
    }
#endif

    ESP_LOGI(TAG, "Added rule at %02d:%02d, type=%d, compartment=%d",
             rule->time.hour, rule->time.minute,
             rule->med_type, rule->compartment_index);
    return ESP_OK;
}

esp_err_t scheduler_remove_rule(uint8_t index)
{
    if (index >= s_rule_count)
        return ESP_ERR_NOT_FOUND;

    /* Shift remaining rules down */
    memmove(&s_rules[index], &s_rules[index + 1],
            (s_rule_count - index - 1) * sizeof(reminder_rule_t));
    s_rule_count--;

    /* Mark removed slot */
    memset(&s_rules[s_rule_count], 0, sizeof(reminder_rule_t));

#ifndef TEST_MODE
    /* Persist */
    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READWRITE, &handle);
    if (ret == ESP_OK) {
        nvs_set_u8(handle, NVS_KEY_COUNT, s_rule_count);
        nvs_set_blob(handle, NVS_KEY_RULES, s_rules,
                     s_rule_count * sizeof(reminder_rule_t));
        nvs_commit(handle);
        nvs_close(handle);
    }
#endif

    ESP_LOGI(TAG, "Removed rule at index %d, %d remaining", index, s_rule_count);
    return ESP_OK;
}

bool scheduler_check_pending(reminder_rule_t *out_rule)
{
    if (s_paused || s_rule_count == 0)
        return false;

    uint16_t current_min = get_current_minutes();
    if (current_min == 0)
        return false;

    /* Find the earliest enabled rule whose time >= current */
    uint16_t closest = 0xFFFF;
    int closest_idx = -1;

    for (uint8_t i = 0; i < s_rule_count; i++) {
        if (!s_rules[i].enabled)
            continue;

        uint16_t rule_min = time_to_minutes(&s_rules[i].time);

        if (rule_min >= current_min && rule_min < closest) {
            closest = rule_min;
            closest_idx = i;
        }
    }

    if (closest_idx < 0)
        return false;

    if (out_rule != NULL) {
        memcpy(out_rule, &s_rules[closest_idx], sizeof(reminder_rule_t));
    }

    return true;
}

esp_err_t scheduler_replace_rules(const reminder_rule_t *rules, uint8_t count)
{
    if (rules == NULL || count > SCHEDULER_MAX_RULES)
        return ESP_ERR_INVALID_ARG;

    /* Clear existing */
    memset(s_rules, 0, sizeof(s_rules));
    s_rule_count = 0;

    /* Add new rules */
    for (uint8_t i = 0; i < count; i++) {
        esp_err_t ret = scheduler_add_rule(&rules[i]);
        if (ret != ESP_OK)
            return ret;
    }

    ESP_LOGI(TAG, "Replaced %d rules", count);
    return ESP_OK;
}

uint8_t scheduler_get_rule_count(void)
{
    return s_rule_count;
}

esp_err_t scheduler_get_rule(uint8_t index, reminder_rule_t *out_rule)
{
    if (index >= s_rule_count || out_rule == NULL)
        return ESP_ERR_NOT_FOUND;

    memcpy(out_rule, &s_rules[index], sizeof(reminder_rule_t));
    return ESP_OK;
}

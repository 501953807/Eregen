/*
 * Eregen (颐贞) - APP Linkage Command Parser Implementation
 * Smart pillbox tier — Parse MQTT downlink commands from cloud/family app
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "app_link.h"

#include <string.h>
#include <stdio.h>

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "esp_log.h"
#include "cJSON.h"

#include "voice_reminder.h"
#include "reminder_scheduler.h"
#include "volume_control.h"
#include "ota_handler.h"

/* MQTT topic for device commands */
#define MQTT_CMD_TOPIC      "eregen/device/PX-+/cmd"

/* Message type strings */
#define MSG_TYPE_MED_RULE   "med_rule"
#define MSG_TYPE_TTS        "tts"
#define MSG_TYPE_CONFIG     "config"
#define MSG_TYPE_OTA        "ota"
#define MSG_TYPE_PAUSE      "pause_reminder"
#define MSG_TYPE_RESUME     "resume_reminder"

/* Log tag */
static const char *TAG = "applink";

/**
 * Parse a JSON "med_rule" message and apply rules to scheduler.
 */
static esp_err_t parse_med_rule(const cJSON *root)
{
    const cJSON *rules_arr = cJSON_GetObjectItem(root, "rules");
    if (!cJSON_IsArray(rules_arr)) {
        ESP_LOGE(TAG, "Missing or invalid 'rules' array");
        return ESP_ERR_INVALID_ARG;
    }

    int count = cJSON_GetArraySize(rules_arr);
    if (count > SCHEDULER_MAX_RULES || count <= 0) {
        ESP_LOGE(TAG, "Invalid rule count: %d", count);
        return ESP_ERR_INVALID_ARG;
    }

    reminder_rule_t rules[SCHEDULER_MAX_RULES];
    memset(rules, 0, sizeof(rules));

    for (int i = 0; i < count; i++) {
        const cJSON *rule = cJSON_GetArrayItem(rules_arr, i);
        if (!cJSON_IsObject(rule))
            continue;

        /* Parse time string "HH:MM" */
        const cJSON *time_obj = cJSON_GetObjectItem(rule, "time");
        if (!cJSON_IsString(time_obj))
            continue;

        int hour = 0, minute = 0;
        if (sscanf(time_obj->valuestring, "%d:%d", &hour, &minute) != 2) {
            ESP_LOGW(TAG, "Invalid time format: %s", time_obj->valuestring);
            continue;
        }

        if (hour < 0 || hour > 23 || minute < 0 || minute > 59) {
            ESP_LOGW(TAG, "Time out of range: %d:%d", hour, minute);
            continue;
        }

        rules[i].time.hour = (uint8_t)hour;
        rules[i].time.minute = (uint8_t)minute;

        /* Parse dose count */
        const cJSON *dose = cJSON_GetObjectItem(rule, "dose");
        rules[i].dose_count = cJSON_IsNumber(dose) ? (uint8_t)dose->valueint : 1;

        /* Parse medicine type */
        const cJSON *type = cJSON_GetObjectItem(rule, "type");
        if (cJSON_IsString(type)) {
            if (strcmp(type->valuestring, "capsule") == 0)
                rules[i].med_type = MED_TYPE_CAPSULE;
            else if (strcmp(type->valuestring, "tablet") == 0)
                rules[i].med_type = MED_TYPE_TABLET;
            else if (strcmp(type->valuestring, "syrup") == 0)
                rules[i].med_type = MED_TYPE_SYRUP;
            else if (strcmp(type->valuestring, "injection") == 0)
                rules[i].med_type = MED_TYPE_INJECTION;
        }

        /* Parse compartment index */
        const cJSON *comp = cJSON_GetObjectItem(rule, "compartment");
        rules[i].compartment_index = cJSON_IsNumber(comp) ?
                                     (uint8_t)comp->valueint : i;
    }

    return scheduler_replace_rules(rules, (uint8_t)count);
}

/**
 * Parse a "config" message and apply settings.
 */
static esp_err_t parse_config(const cJSON *root)
{
    const cJSON *settings = cJSON_GetObjectItem(root, "settings");
    if (!cJSON_IsObject(settings))
        return ESP_ERR_INVALID_ARG;

    /* Parse volume */
    const cJSON *vol = cJSON_GetObjectItem(settings, "volume");
    if (cJSON_IsNumber(vol)) {
        uint8_t v = (uint8_t)vol->valueint;
        if (v <= 100) {
            volume_set(v);
            tts_set_volume(v);
        }
    }

    return ESP_OK;
}

esp_err_t applink_init(void)
{
    ESP_LOGI(TAG, "APP linkage initialized");
    return ESP_OK;
}

esp_err_t applink_parse_mqtt_message(const char *topic,
                                     const uint8_t *payload,
                                     size_t payload_len)
{
    if (topic == NULL || payload == NULL || payload_len == 0)
        return ESP_ERR_INVALID_ARG;

    /* Parse JSON payload */
    cJSON *root = cJSON_ParseWithLength((const char *)payload, payload_len);
    if (root == NULL) {
        ESP_LOGE(TAG, "Failed to parse JSON payload");
        return ESP_FAIL;
    }

    /* Extract message type */
    const cJSON *type = cJSON_GetObjectItem(root, "type");
    if (!cJSON_IsString(type)) {
        ESP_LOGE(TAG, "Missing message type field");
        cJSON_Delete(root);
        return ESP_FAIL;
    }

    esp_err_t ret = ESP_FAIL;

    if (strcmp(type->valuestring, MSG_TYPE_MED_RULE) == 0) {
        ret = parse_med_rule(root);
    } else if (strcmp(type->valuestring, MSG_TYPE_TTS) == 0) {
        const cJSON *text = cJSON_GetObjectItem(root, "text");
        if (cJSON_IsString(text)) {
            tts_speak(text->valuestring);
            ret = ESP_OK;
        }
    } else if (strcmp(type->valuestring, MSG_TYPE_CONFIG) == 0) {
        ret = parse_config(root);
    } else if (strcmp(type->valuestring, MSG_TYPE_OTA) == 0) {
        ret = ota_handle_command(topic, (const uint8_t *)payload, (uint16_t)payload_len);
    } else if (strcmp(type->valuestring, MSG_TYPE_PAUSE) == 0) {
        ret = applink_handle_pause_reminder();
    } else if (strcmp(type->valuestring, MSG_TYPE_RESUME) == 0) {
        ret = applink_handle_resume_reminder();
    } else {
        ESP_LOGW(TAG, "Unknown message type: %s", type->valuestring);
    }

    cJSON_Delete(root);
    return ret;
}

esp_err_t applink_handle_set_rules(const void *rules, uint8_t count)
{
    return scheduler_replace_rules((const reminder_rule_t *)rules, count);
}

esp_err_t applink_handle_pause_reminder(void)
{
    ESP_LOGI(TAG, "Reminders paused by remote command");
    /* In a full implementation this would set a flag in the scheduler */
    return ESP_OK;
}

esp_err_t applink_handle_resume_reminder(void)
{
    ESP_LOGI(TAG, "Reminders resumed by remote command");
    return ESP_OK;
}

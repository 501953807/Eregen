/*
 * Eregen (颐贞) - Volume Control Module Implementation
 * Smart pillbox tier — TTS volume via buttons + NVS persistence
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "volume_control.h"

#include <stdio.h>

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "nvs_flash.h"
#include "esp_log.h"

#include "voice_reminder.h"

/* NVS key for volume */
#define NVS_KEY_VOLUME      "tts_volume"

/* Volume step percentage */
#define VOLUME_STEP         10

/* Log tag */
static const char *TAG = "volume";

/* Current volume */
static uint8_t s_volume = VOLUME_DEFAULT;

esp_err_t volume_init(void)
{
    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READONLY, &handle);
    if (ret == ESP_OK) {
        uint8_t vol = VOLUME_DEFAULT;
        ret = nvs_get_u8(handle, NVS_KEY_VOLUME, &vol);
        if (ret == ESP_OK) {
            s_volume = vol;
            ESP_LOGI(TAG, "Loaded volume=%d%% from NVS", s_volume);
        }
        nvs_close(handle);
    }

    return ESP_OK;
}

uint8_t volume_get(void)
{
    return s_volume;
}

esp_err_t volume_set(uint8_t percent)
{
    if (percent > 100)
        return ESP_ERR_INVALID_ARG;

    s_volume = percent;

    /* Persist to NVS */
    nvs_handle_t handle;
    esp_err_t ret = nvs_open("pillbox", NVS_READWRITE, &handle);
    if (ret == ESP_OK) {
        nvs_set_u8(handle, NVS_KEY_VOLUME, s_volume);
        nvs_commit(handle);
        nvs_close(handle);
    }

    /* Sync with TTS module */
    tts_set_volume(s_volume);

    ESP_LOGI(TAG, "Volume set to %d%%", s_volume);
    return ESP_OK;
}

uint8_t volume_increase(void)
{
    uint8_t new_vol = s_volume + VOLUME_STEP;
    if (new_vol > 100)
        new_vol = 100;
    volume_set(new_vol);
    return new_vol;
}

uint8_t volume_decrease(void)
{
    if (s_volume < VOLUME_STEP) {
        volume_set(0);
        return 0;
    }
    uint8_t new_vol = s_volume - VOLUME_STEP;
    volume_set(new_vol);
    return new_vol;
}

uint8_t volume_handle_button(bool increase)
{
    if (increase) {
        return volume_increase();
    } else {
        return volume_decrease();
    }
}

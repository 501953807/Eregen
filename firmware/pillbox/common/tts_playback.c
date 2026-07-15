/*
 * Eregen (颐贞) - Chinese TTS Playback Implementation
 * SYN5300 module via UART interface.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "tts_playback.h"

#include "driver/uart.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/queue.h"
#include "freertos/semphr.h"

#include "esp_log.h"
#include <string.h>

static const char *TAG = "tts";

/* Internal state */
static tts_done_callback_t s_cb           = NULL;
static uint8_t             s_volume        = TTS_VOLUME_DEFAULT;
static bool                s_playing       = false;
static SemaphoreHandle_t   s_play_mutex    = NULL;

/* Text queue */
static char    s_text_queue[TTS_QUEUE_MAX][TTS_TEXT_MAX_LEN];
static uint8_t s_queue_head               = 0;
static uint8_t s_queue_tail               = 0;
static uint8_t s_queue_count              = 0;

/**
 * Initialize the TTS subsystem.
 */
esp_err_t tts_init(tts_done_callback_t cb)
{
    s_cb = cb;
    s_volume = TTS_VOLUME_DEFAULT;
    s_playing = false;
    s_queue_count = 0;
    s_queue_head = 0;
    s_queue_tail = 0;
    memset(s_text_queue, 0, sizeof(s_text_queue));

    s_play_mutex = xSemaphoreCreateMutex();
    if (!s_play_mutex) {
        ESP_LOGE(TAG, "Failed to create play mutex");
        return ESP_ERR_NO_MEM;
    }

    /* Configure UART for SYN5300 */
    uart_config_t uart_conf = {
        .baud_rate = TTS_UART_BAUD,
        .data_bits = UART_DATA_8_BITS,
        .parity    = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
        .flow_ctrl = UART_HW_FLOWCTRL_DISABLE,
    };

    esp_err_t ret = uart_param_config(TTS_UART_NUM, &uart_conf);
    if (ret != ESP_OK) return ret;

    ret = uart_set_pin(TTS_UART_NUM, TTS_UART_TX_PIN,
                       TTS_UART_RX_PIN, UART_PIN_NO_CHANGE,
                       UART_PIN_NO_CHANGE);
    if (ret != ESP_OK) return ret;

    ret = uart_driver_install(TTS_UART_NUM, 256, 256, 0, NULL, 0);
    if (ret != ESP_OK) return ret;

    ESP_LOGI(TAG, "TTS initialized on UART%d, volume=%d%%",
             TTS_UART_NUM, s_volume);
    return ESP_OK;
}

/**
 * Queue a text for TTS playback.
 */
esp_err_t tts_play(const char *text)
{
    if (!text || !s_play_mutex) return ESP_ERR_INVALID_STATE;

    BaseType_t xHigherPriorityTaskWoken = pdFALSE;

    if (xSemaphoreTakeFromISR(s_play_mutex, &xHigherPriorityTaskWoken) == pdTRUE) {
        if (s_queue_count >= TTS_QUEUE_MAX) {
            xSemaphoreGiveFromISR(s_play_mutex, &xHigherPriorityTaskWoken);
            return ESP_ERR_NO_MEM;
        }

        size_t len = strlen(text);
        if (len >= TTS_TEXT_MAX_LEN) {
            len = TTS_TEXT_MAX_LEN - 1;
        }
        memcpy(s_text_queue[s_queue_tail], text, len + 1);
        s_queue_tail = (s_queue_tail + 1) % TTS_QUEUE_MAX;
        s_queue_count++;

        xSemaphoreGiveFromISR(s_play_mutex, &xHigherPriorityTaskWoken);

        if (xHigherPriorityTaskWoken) {
            portYIELD_FROM_ISR(xHigherPriorityTaskWoken);
        }
    }

    return ESP_OK;
}

/**
 * Stop current TTS playback and clear queue.
 */
esp_err_t tts_stop(void)
{
    if (!s_play_mutex) return ESP_ERR_INVALID_STATE;

    if (xSemaphoreTake(s_play_mutex, pdMS_TO_TICKS(100)) == pdTRUE) {
        s_queue_count = 0;
        s_queue_head = 0;
        s_queue_tail = 0;
        s_playing = false;
        xSemaphoreGive(s_play_mutex);
    }

    /* Send stop command to SYN5300 via UART */
    const char stop_cmd[] = "\xAA\x01\x00\xFF";
    uart_write_bytes(TTS_UART_NUM, stop_cmd, sizeof(stop_cmd) - 1);

    return ESP_OK;
}

/**
 * Check if TTS is currently playing.
 */
bool tts_is_playing(void)
{
    if (!s_play_mutex) return false;

    if (xSemaphoreTake(s_play_mutex, pdMS_TO_TICKS(100)) == pdTRUE) {
        bool playing = s_playing;
        xSemaphoreGive(s_play_mutex);
        return playing;
    }
    return false;
}

/**
 * Set TTS volume.
 */
esp_err_t tts_set_volume(uint8_t percent)
{
    if (percent > TTS_VOLUME_MAX) {
        return ESP_ERR_INVALID_ARG;
    }
    s_volume = percent;

    /* Send volume command to SYN5300 */
    /* Protocol: 0xAA 0x04 0x01 [volume_byte] 0xFF */
    uint8_t vol_byte = (uint8_t)((percent * 0xFF) / TTS_VOLUME_MAX);
    uint8_t cmd[5] = { 0xAA, 0x04, 0x01, vol_byte, 0xFF };
    uart_write_bytes(TTS_UART_NUM, cmd, sizeof(cmd));

    return ESP_OK;
}

/**
 * Internal: drain the text queue by sending each text over UART.
 * Call from background task loop.
 */
void tts_drain_queue(void)
{
    if (!s_play_mutex) return;

    if (xSemaphoreTake(s_play_mutex, pdMS_TO_TICKS(10)) == pdTRUE) {
        if (s_queue_count > 0 && !s_playing) {
            /* Get next text from queue head */
            const char *text = s_text_queue[s_queue_head];
            size_t len = strlen(text);

            s_queue_head = (s_queue_head + 1) % TTS_QUEUE_MAX;
            s_queue_count--;
            s_playing = true;

            xSemaphoreGive(s_play_mutex);

            /* Send text to SYN5300 via UART */
            /* SYN5300 expects: 0xAA [length_hi] [length_lo] [text...] 0xFF */
            uint8_t header[4] = {
                0xAA,
                (uint8_t)(len >> 8),
                (uint8_t)(len & 0xFF),
                0xFF
            };
            uart_write_bytes(TTS_UART_NUM, header, 3);
            uart_write_bytes(TTS_UART_NUM, text, len);
            uart_write_bytes(TTS_UART_NUM, "\xFF", 1);
        } else {
            xSemaphoreGive(s_play_mutex);
        }
    }
}

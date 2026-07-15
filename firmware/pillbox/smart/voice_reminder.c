/*
 * Eregen (颐贞) - Voice Reminder Module Implementation (TTS via SYN5300)
 * Smart pillbox tier — Chinese text-to-speech medication reminders
 *
 * SYN5300 protocol: UART 115200 8N1
 *   Command frame: [0xAA][CMD][LEN][DATA...][0x55]
 *   CMD_SPEAK = 0x01, CMD_VOLUME = 0x02, CMD_STOP = 0x03
 *   Volume data: 0x00 (mute) to 0x64 (100%), default 0x50 (80%)
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "voice_reminder.h"

#include <string.h>
#include <stdio.h>

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/queue.h"
#include "freertos/ringbuf.h"

#include "driver/uart.h"
#include "esp_log.h"

/* UART configuration for SYN5300 */
#define TTS_UART_NUM          UART_NUM_1
#define TTS_TX_PIN            GPIO_NUM_2
#define TTS_RX_PIN            GPIO_NUM_3
#define TTS_BAUD_RATE         115200
#define TTS_STACK_SIZE        (2048)
#define TTS_QUEUE_LEN         8
#define TTS_FRAME_BUF_SIZE    64

/* SYN5300 command codes */
#define SYN_CMD_SPEAK         0x01
#define SYN_CMD_VOLUME        0x02
#define SYN_CMD_STOP          0x03

/* SYN5300 frame delimiters */
#define SYN_FRAME_START       0xAA
#define SYN_FRAME_END         0x55

/* Log tag */
static const char *TAG = "tts";

/* Internal state */
static int s_uart_num = -1;
static uint8_t s_volume = 80;           /* Default 80% */
static bool s_playing = false;
static QueueHandle_t s_text_queue = NULL;

/* TTS playback task handle */
static TaskHandle_t s_tts_task = NULL;

/**
 * Build and send a SYN5300 command frame over UART.
 */
static esp_err_t tts_send_frame(uint8_t cmd, const uint8_t *data, uint8_t len)
{
    if (s_uart_num < 0)
        return ESP_ERR_INVALID_STATE;

    uint8_t buf[TTS_FRAME_BUF_SIZE];
    uint8_t offset = 0;

    buf[offset++] = SYN_FRAME_START;
    buf[offset++] = cmd;
    buf[offset++] = len;

    if (data != NULL && len > 0) {
        memcpy(buf + offset, data, len);
        offset += len;
    }

    buf[offset++] = SYN_FRAME_END;

    size_t written;
    esp_err_t ret = uart_write_bytes(s_uart_num, buf, offset, portMAX_DELAY);
    if (ret >= 0) {
        return ESP_OK;
    }
    return ret;
}

/**
 * TTS playback task — dequeues text strings and sends speak commands.
 */
static void tts_playback_task(void *pvParameter)
{
    (void)pvParameter;
    char text_buf[128];
    size_t bytes_read;

    ESP_LOGI(TAG, "TTS playback task started");

    for (;;) {
        bytes_read = (size_t)xQueueReceive(s_text_queue, text_buf, portMAX_DELAY);
        if (bytes_read == 0)
            continue;

        /* Trim trailing whitespace */
        size_t text_len = strlen(text_buf);
        while (text_len > 0 && (text_buf[text_len - 1] == ' ' ||
               text_buf[text_len - 1] == '\n' || text_buf[text_len - 1] == '\r')) {
            text_buf[--text_len] = '\0';
        }

        if (text_len == 0)
            continue;

        ESP_LOGI(TAG, "Speaking: %s", text_buf);
        s_playing = true;

        /* Send speak command with text as UTF-8 data */
        if (text_len <= TTS_FRAME_BUF_SIZE - 4) {
            tts_send_frame(SYN_CMD_SPEAK, (const uint8_t *)text_buf, (uint8_t)text_len);
        } else {
            /* Truncate overly long text */
            text_buf[TTS_FRAME_BUF_SIZE - 5] = '\0';
            tts_send_frame(SYN_CMD_SPEAK, (const uint8_t *)text_buf,
                           (uint8_t)(TTS_FRAME_BUF_SIZE - 5));
        }

        /* Simulate playback duration: ~150ms per character */
        uint32_t play_ms = (uint32_t)text_len * 150;
        vTaskDelay(pdMS_TO_TICKS(play_ms));

        s_playing = false;
    }
}

esp_err_t tts_init(int uart_num)
{
    if (uart_num < 0 || uart_num >= UART_NUM_MAX)
        return ESP_ERR_INVALID_ARG;

    s_uart_num = uart_num;

    uart_config_t uart_config = {
        .baud_rate = TTS_BAUD_RATE,
        .data_bits = UART_DATA_8_BITS,
        .parity    = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
        .flow_ctrl = UART_HW_FLOWCTRL_DISABLE,
    };

    esp_err_t ret = uart_param_config(uart_num, &uart_config);
    if (ret != ESP_OK)
        return ret;

    ret = uart_set_pin(uart_num, TTS_TX_PIN, TTS_RX_PIN, UART_PIN_NO_CHANGE,
                       UART_PIN_NO_CHANGE);
    if (ret != ESP_OK)
        return ret;

    ret = uart_driver_install(uart_num, TTS_FRAME_BUF_SIZE * 4, 0, 0, NULL, 0);
    if (ret != ESP_OK)
        return ret;

    /* Create text queue */
    s_text_queue = xQueueCreate(TTS_QUEUE_LEN, sizeof(text_buf));
    if (s_text_queue == NULL) {
        uart_driver_delete(uart_num);
        return ESP_ERR_NO_MEM;
    }

    /* Create playback task */
    ret = xTaskCreate(tts_playback_task, "tts_play", TTS_STACK_SIZE,
                      NULL, tskIDLE_PRIORITY + 1, &s_tts_task);
    if (ret != pdPASS) {
        vQueueDelete(s_text_queue);
        uart_driver_delete(uart_num);
        return ESP_FAIL;
    }

    /* Apply default volume */
    uint8_t vol_byte = (uint8_t)((s_volume * 0x64) / 100);
    tts_send_frame(SYN_CMD_VOLUME, &vol_byte, 1);

    ESP_LOGI(TAG, "TTS initialized on UART%d, volume=%d%%", uart_num, s_volume);
    return ESP_OK;
}

esp_err_t tts_speak(const char *text)
{
    if (text == NULL || s_text_queue == NULL)
        return ESP_ERR_INVALID_ARG;

    /* Copy text into queue buffer */
    char copy[128];
    strncpy(copy, text, sizeof(copy) - 1);
    copy[sizeof(copy) - 1] = '\0';

    if (xQueueSend(s_text_queue, copy, pdMS_TO_TICKS(100)) != pdTRUE)
        return ESP_ERR_NO_MEM;

    return ESP_OK;
}

void tts_stop(void)
{
    if (s_uart_num < 0)
        return;

    tts_send_frame(SYN_CMD_STOP, NULL, 0);
    s_playing = false;

    /* Clear any pending texts in queue */
    if (s_text_queue != NULL) {
        uint32_t pending;
        pending = uxQueueMessagesWaiting(s_text_queue);
        if (pending > 0) {
            void *item;
            for (uint32_t i = 0; i < pending; i++) {
                xQueueReceive(s_text_queue, &item, 0);
            }
        }
    }

    ESP_LOGI(TAG, "TTS stopped");
}

bool tts_is_playing(void)
{
    return s_playing;
}

esp_err_t tts_set_volume(uint8_t percent)
{
    if (percent > 100)
        return ESP_ERR_INVALID_ARG;

    s_volume = percent;

    /* Map 0-100 to 0x00-0x64 */
    uint8_t vol_byte = (uint8_t)((percent * 0x64) / 100);
    return tts_send_frame(SYN_CMD_VOLUME, &vol_byte, 1);
}

uint8_t tts_get_volume(void)
{
    return s_volume;
}

/*
 * Eregen (颐贞) - Chinese TTS Playback Implementation
 * SYN5300 module communication layer.
 *
 * In ESP-IDF: sends protocol frames over UART.
 * In TEST_MODE: maintains an in-memory queue for host compilation.
 *
 * 2026 Eregen (颐贞). All rights reserved.
 */

#include "tts_playback.h"

#ifdef TEST_MODE
#include <string.h>
#include <stdio.h>
#include <unistd.h>
#else
#include "driver/uart.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "freertos/queue.h"
#include "esp_log.h"
#endif

/* Max pending text messages */
#define TTS_QUEUE_MAX     8

/* Max text length per message */
#define TTS_TEXT_MAX_LEN  128

/* SYN5300 UART configuration (ESP-IDF only) */
#ifndef TEST_MODE
#define TTS_UART_NUM      UART_NUM_1
#define TTS_UART_TX_PIN   GPIO_NUM_20
#define TTS_UART_BAUD     9600
#endif

/* Internal state */
static bool s_playing = false;

#ifdef TEST_MODE
/* Host-side text queue */
static char    s_text_queue[TTS_QUEUE_MAX][TTS_TEXT_MAX_LEN];
static uint8_t s_queue_head = 0;
static uint8_t s_queue_tail = 0;
static uint8_t s_queue_count = 0;

/* Forward declaration */
static void tts_drain_queue(void);
#else
/* ESP-IDF side: FreeRTOS queue */
static QueueHandle_t s_q_handle = NULL;
#endif

/**
 * Build and send a SYN5300 text frame.
 * Protocol: 0xAA [len_hi] [len_lo] [text...] 0xFF
 */
static void tts_send_frame(const char *text, size_t len)
{
#ifdef TEST_MODE
    /* Mock: print what would be sent */
    printf("[TTS-MOCK] Speaking: %s\n", text);
#else
    uint8_t header[3] = {
        0xAA,
        (uint8_t)(len >> 8),
        (uint8_t)(len & 0xFF)
    };
    uart_write_bytes(TTS_UART_NUM, header, 3);
    uart_write_bytes(TTS_UART_NUM, text, len);
    uart_write_bytes(TTS_UART_NUM, "\xFF", 1);
#endif
}

/**
 * Initialize the TTS subsystem on first call.
 */
static void tts_init_once(void)
{
#ifdef TEST_MODE
    if (s_queue_count == 0 && s_queue_head == 0 && s_queue_tail == 0) {
        memset(s_text_queue, 0, sizeof(s_text_queue));
    }
#else
    if (s_q_handle != NULL) return;  /* Already initialized */

    /* Configure UART for SYN5300 */
    uart_config_t uart_conf = {
        .baud_rate = TTS_UART_BAUD,
        .data_bits = UART_DATA_8_BITS,
        .parity    = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
        .flow_ctrl = UART_HW_FLOWCTRL_DISABLE,
    };

    uart_param_config(TTS_UART_NUM, &uart_conf);
    uart_set_pin(TTS_UART_NUM, TTS_UART_TX_PIN,
                 TTS_UART_RX_PIN, UART_PIN_NO_CHANGE,
                 UART_PIN_NO_CHANGE);
    uart_driver_install(TTS_UART_NUM, 256, 256, 0, NULL, 0);

    s_q_handle = xQueueCreate(TTS_QUEUE_MAX, TTS_TEXT_MAX_LEN);
#endif
}

/**
 * Queue a text string for TTS playback and begin speaking if idle.
 */
void tts_speak(const char *text)
{
    if (!text) return;

    tts_init_once();

    size_t len = strlen(text);
    if (len >= TTS_TEXT_MAX_LEN) {
        len = TTS_TEXT_MAX_LEN - 1;
    }

#ifdef TEST_MODE
    if (s_queue_count >= TTS_QUEUE_MAX) {
        /* Queue full -- drop oldest */
        if (s_queue_count > 0) {
            s_queue_head = (s_queue_head + 1) % TTS_QUEUE_MAX;
            s_queue_count--;
        }
    }
    memcpy(s_text_queue[s_queue_tail], text, len + 1);
    s_queue_tail = (s_queue_tail + 1) % TTS_QUEUE_MAX;
    s_queue_count++;

    /* If not currently playing, drain immediately */
    if (!s_playing) {
        tts_drain_queue();
    }
#else
    if (s_q_handle) {
        xQueueSend(s_q_handle, text, pdMS_TO_TICKS(100));
    }
#endif
}

/**
 * Drain the text queue by sending each queued text.
 * Called when TTS is idle.
 */
#ifdef TEST_MODE
static void tts_drain_queue(void)
{
    while (s_queue_count > 0 && !s_playing) {
        const char *text = s_text_queue[s_queue_head];
        size_t len = strlen(text);
        if (len == 0) break;

        s_playing = true;
        tts_send_frame(text, len);

        /* Simulate playback duration: ~100ms per character */
        usleep((useconds_t)(len * 100));

        s_playing = false;
        s_queue_head = (s_queue_head + 1) % TTS_QUEUE_MAX;
        s_queue_count--;
    }
}
#endif

/**
 * Check if TTS is currently playing.
 */
bool tts_is_playing(void)
{
    return s_playing;
}

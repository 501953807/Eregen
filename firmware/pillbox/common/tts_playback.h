/*
 * Eregen (颐贞) - Chinese TTS Playback Wrapper
 * SYN5300 module via UART interface.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef TTS_PLAYBACK_H
#define TTS_PLAYBACK_H

#include "esp_err.h"
#include <stddef.h>
#include <stdint.h>

/* UART configuration for SYN5300 */
#define TTS_UART_NUM          UART_NUM_1
#define TTS_UART_TX_PIN       GPIO_NUM_20
#define TTS_UART_RX_PIN       GPIO_NUM_21
#define TTS_UART_BAUD         9600

/* TTS text queue: max pending texts */
#define TTS_QUEUE_MAX         8

/* Max text length per message */
#define TTS_TEXT_MAX_LEN      128

/* Volume range */
#define TTS_VOLUME_MIN        0
#define TTS_VOLUME_MAX        100
#define TTS_VOLUME_DEFAULT    80

/* Callback type for TTS completion notification */
typedef void (*tts_done_callback_t)(void);

/**
 * Initialize the TTS subsystem.
 * Configures UART and starts the playback task.
 *
 * @param cb  Completion callback (called when queued text finishes playing), or NULL
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t tts_init(tts_done_callback_t cb);

/**
 * Queue a Chinese text string for TTS playback.
 * Texts are played sequentially; this function returns immediately.
 *
 * @param text  Null-terminated UTF-8 text string
 * @return ESP_OK on success, ESP_ERR_NO_MEM if queue full
 */
esp_err_t tts_play(const char *text);

/**
 * Stop current TTS playback immediately.
 * Clears the text queue.
 *
 * @return ESP_OK on success
 */
esp_err_t tts_stop(void);

/**
 * Check if TTS is currently playing.
 *
 * @return true if playing, false otherwise
 */
bool tts_is_playing(void);

/**
 * Set TTS volume (0-100%).
 *
 * @param percent Volume percentage (0=mute, 100=max)
 * @return ESP_OK on success, ESP_ERR_INVALID_ARG if out of range
 */
esp_err_t tts_set_volume(uint8_t percent);

#endif /* TTS_PLAYBACK_H */

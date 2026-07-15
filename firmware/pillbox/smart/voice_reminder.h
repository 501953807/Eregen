/*
 * Eregen (颐贞) - Voice Reminder Module (TTS via SYN5300)
 * Smart pillbox tier — Chinese text-to-speech medication reminders
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef VOICE_REMINDER_H
#define VOICE_REMINDER_H

#include "esp_err.h"
#include <stdint.h>
#include <stdbool.h>

/**
 * Initialize TTS module via UART.
 *
 * @param uart_num UART peripheral number (e.g., UART_NUM_1)
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t tts_init(int uart_num);

/**
 * Send Chinese text to TTS module for asynchronous playback.
 * Text is queued; playback happens in order.
 *
 * @param text Chinese text string, e.g. "爷爷，该吃降压药了"
 * @return ESP_OK on success, error code if queue is full
 */
esp_err_t tts_speak(const char *text);

/**
 * Stop current TTS playback immediately.
 */
void tts_stop(void);

/**
 * Check whether TTS is currently playing.
 *
 * @return true if speaking, false otherwise
 */
bool tts_is_playing(void);

/**
 * Set TTS volume (0-100%).
 *
 * @param percent Volume percentage (0=mute, 100=max, default=80)
 * @return ESP_OK on success
 */
esp_err_t tts_set_volume(uint8_t percent);

/**
 * Get current TTS volume.
 *
 * @return Current volume percentage (0-100)
 */
uint8_t tts_get_volume(void);

#endif /* VOICE_REMINDER_H */

/*
 * Eregen (颐贞) - Chinese TTS Playback Interface
 * SYN5300 module via UART interface.
 *
 * Compatible with ESP-IDF and standalone host compilation (TEST_MODE).
 *
 * 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef TTS_PLAYBACK_H
#define TTS_PLAYBACK_H

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

/**
 * Queue a Chinese text string for TTS playback.
 * Blocks briefly if the internal queue is full; returns immediately otherwise.
 *
 * @param text Null-terminated UTF-8 text string
 */
void tts_speak(const char *text);

/**
 * Check if TTS is currently playing.
 *
 * @return true if playing, false otherwise
 */
bool tts_is_playing(void);

#endif /* TTS_PLAYBACK_H */

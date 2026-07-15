/*
 * Eregen (颐贞) - SOS Button Detection Header
 * Physical button with debounce, long press detection, anti-false-trigger
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef SOS_BUTTON_H
#define SOS_BUTTON_H

#include <stdint.h>
#include <stdbool.h>

/* Debounce time in milliseconds */
#define SOS_DEBOUNCE_MS      50U

/* Long press threshold in milliseconds */
#define SOS_LONG_PRESS_MS    3000U

/* Anti-false-trigger: consecutive readings required */
#define SOS_CONSECUTIVE_REQ  3U

/* Check interval for button state machine */
#define SOS_CHECK_INTERVAL_MS 10U

/*
 * Initialize the SOS button GPIO and state machine.
 */
void sos_init(void);

/*
 * Call periodically at SOS_CHECK_INTERVAL_MS intervals.
 * Updates internal state machine (debounce, long press tracking).
 */
void sos_task(void);

/*
 * Check if the SOS button has just been pressed (momentary press).
 * @return true if a valid press was detected since last call.
 */
bool sos_is_pressed(void);

/*
 * Check if the SOS long press condition is met.
 * @return true if button has been held for >= SOS_LONG_PRESS_MS ms.
 */
bool sos_is_long_press(void);

/*
 * Get the current hold time of the SOS button in milliseconds.
 * Returns 0 if button is not being held.
 * @return Hold time in ms, or 0 if released.
 */
uint32_t sos_get_hold_time_ms(void);

/*
 * Reset the "just pressed" flag.
 * Must be called by the application after handling a press event.
 */
void sos_reset_pressed_flag(void);

/*
 * Reset the long press flag.
 * Must be called by the application after handling a long press event.
 */
void sos_reset_long_press_flag(void);

/* Test-mode helper: set mock GPIO state */
#ifdef TEST_MODE
void sos_set_mock_state(bool state);
#endif

#endif /* SOS_BUTTON_H */

/*
 * Eregen (颐贞) - SOS Button Detection Implementation
 * Physical button with debounce, long press detection, anti-false-trigger
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "sos_button.h"

#ifdef TEST_MODE
#include <string.h>
/* In test mode, mock GPIO state is set by the test harness */
static bool s_mock_button_state = false;
bool sos_gpio_read_raw(void)
{
    return s_mock_button_state;
}
void sos_set_mock_state(bool state)
{
    s_mock_button_state = state;
}
#else
#include "gd32e230_gpio.h"
#include "FreeRTOS.h"
#include "task.h"
static bool sos_gpio_read_raw(void)
{
    return (gpio_input_bit_get(SOS_BUTTON_GPIO_PORT, SOS_BUTTON_GPIO_PIN) == RESET);
}
#endif

/* Internal state */
typedef struct {
    uint16_t stable_count;     /* Consecutive same-state readings */
    uint32_t hold_start_tick;  /* Tick count when button became stable-pressed */
    uint32_t hold_time_ms;     /* Current hold duration in ms (persists after release) */
    bool     just_pressed;     /* Flag: valid press detected since last clear */
    bool     just_long_press;  /* Flag: valid long press detected since last clear */
    bool     current_state;    /* Current debounced button state */
} sos_state_t;

static sos_state_t s_sos;

/*
 * Initialize the SOS button.
 */
void sos_init(void)
{
    memset(&s_sos, 0, sizeof(s_sos));
    s_sos.current_state = false;
    s_sos.just_pressed = false;
    s_sos.just_long_press = false;
    s_sos.stable_count = 0;
    s_sos.hold_start_tick = 0;
    s_sos.hold_time_ms = 0;
}

/*
 * Call periodically at SOS_CHECK_INTERVAL_MS intervals.
 * Updates internal state machine.
 */
void sos_task(void)
{
    bool raw = sos_gpio_read_raw();

    if (raw) {
        /* Button appears pressed */
        if (s_sos.current_state == false) {
            /* Count consecutive pressed readings for debounce + anti-false-trigger */
            s_sos.stable_count++;

            /* Need MORE than CONSECUTIVE_REQ to confirm (i.e., CONSECUTIVE_REQ+1) */
            if (s_sos.stable_count > SOS_CONSECUTIVE_REQ) {
                s_sos.current_state = true;
#ifdef TEST_MODE
                s_sos.hold_start_tick = s_sos.stable_count;
#else
                s_sos.hold_start_tick = xTaskGetTickCount();
#endif
                s_sos.just_pressed = true;
            } else {
                /* Still debouncing -- don't track hold time yet */
                return;
            }
        }
        /* current_state == true: continue counting and tracking hold time */
        s_sos.stable_count++;
#ifdef TEST_MODE
        s_sos.hold_time_ms = (uint32_t)(s_sos.stable_count - SOS_CONSECUTIVE_REQ) *
                             SOS_CHECK_INTERVAL_MS;
#else
        uint32_t elapsed = (xTaskGetTickCount() - s_sos.hold_start_tick) *
                           pdTICKS_TO_MS(1);
        s_sos.hold_time_ms = elapsed;
#endif

        /* Check for long press */
        if (s_sos.hold_time_ms >= SOS_LONG_PRESS_MS) {
            s_sos.just_long_press = true;
        }
    } else {
        /* Button released */
        if (s_sos.current_state == true) {
            /* Was pressed, now released */
            s_sos.current_state = false;
            s_sos.just_pressed = false;
        }
        s_sos.stable_count = 0;
        s_sos.hold_start_tick = 0;
        /* Do NOT clear hold_time_ms -- let application read the last hold duration */
    }
}

/*
 * Check if the SOS button has just been pressed.
 */
bool sos_is_pressed(void)
{
    return s_sos.just_pressed;
}

/*
 * Check if the SOS long press condition is met.
 */
bool sos_is_long_press(void)
{
    return s_sos.just_long_press;
}

/*
 * Get the current hold time of the SOS button in milliseconds.
 * Returns 0 if button was never pressed or hold_time was cleared.
 * Note: hold_time persists after release until explicitly cleared.
 */
uint32_t sos_get_hold_time_ms(void)
{
    return s_sos.hold_time_ms;
}

/*
 * Reset the "just pressed" flag.
 * Must be called by the application after handling a press event.
 */
void sos_reset_pressed_flag(void)
{
    s_sos.just_pressed = false;
}

/*
 * Reset the "just long pressed" flag.
 * Must be called by the application after handling a long press event.
 */
void sos_reset_long_press_flag(void)
{
    s_sos.just_long_press = false;
}

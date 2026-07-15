/*
 * Eregen (颐贞) - Button Input Implementation
 * Polling-based button detection with debounce and press classification
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "button_input.h"

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "esp_log.h"
#include "driver/gpio.h"

static const char *TAG = "button_input";

/* Button GPIO pins (from main.c skeleton) */
#define PIN_BUTTON_ENTER    GPIO_NUM_0
#define PIN_BUTTON_RIGHT    GPIO_NUM_9

/* Internal state */
typedef struct {
    gpio_num_t pin;
    bool last_level;
    uint32_t press_start_tick;
    bool is_pressed;
    uint32_t last_debounce_tick;
    uint32_t release_tick;
    uint8_t press_count;       /* For double-press detection */
} button_state_t;

static button_state_t s_buttons[BUTTON_COUNT];
static button_event_t s_last_event = BUTTON_NONE;
static bool s_event_consumed = true;

static void button_scan(void);
static void button_classify(button_id_t id);

/**
 * Initialize button GPIO pins with pull-up resistors.
 */
esp_err_t buttons_init(void)
{
    gpio_config_t btn_cfg = {
        .pin_bit_mask = (1ULL << PIN_BUTTON_ENTER) | (1ULL << PIN_BUTTON_RIGHT),
        .mode = GPIO_MODE_INPUT,
        .pull_up_en = GPIO_PULLUP_ENABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type = GPIO_INTR_DISABLE,
    };
    esp_err_t ret = gpio_config(&btn_cfg);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "GPIO config failed: %s", esp_err_to_name(ret));
        return ret;
    }

    /* Initialize button states */
    for (int i = 0; i < BUTTON_COUNT; i++) {
        s_buttons[i].pin = (i == BUTTON_ENTER) ? PIN_BUTTON_ENTER : PIN_BUTTON_RIGHT;
        s_buttons[i].last_level = true;
        s_buttons[i].press_start_tick = 0;
        s_buttons[i].is_pressed = false;
        s_buttons[i].last_debounce_tick = 0;
        s_buttons[i].release_tick = 0;
        s_buttons[i].press_count = 0;
    }

    s_last_event = BUTTON_NONE;
    s_event_consumed = true;

    ESP_LOGI(TAG, "Buttons initialized: Enter=GPIO%d, Right=GPIO%d",
             PIN_BUTTON_ENTER, PIN_BUTTON_RIGHT);

    return ESP_OK;
}

/**
 * Poll all buttons and return the latest button event.
 * Must be called periodically from the main loop.
 */
button_event_t buttons_get_event(void)
{
    button_scan();
    return s_last_event;
}

/**
 * Clear the last consumed button event.
 */
void buttons_clear_event(void)
{
    s_last_event = BUTTON_NONE;
    s_event_consumed = true;
}

/* ---- Internal helpers ---- */

/**
 * Scan all buttons and detect press events.
 */
static void button_scan(void)
{
    uint32_t now = xTaskGetTickCount();

    for (int i = 0; i < BUTTON_COUNT; i++) {
        bool level = gpio_get_level(s_buttons[i].pin);

        /* Active low: button pressed = GPIO low */
        bool pressed = !level;

        /* Debounce: ignore changes within debounce window */
        if ((now - s_buttons[i].last_debounce_tick) < pdMS_TO_TICKS(BUTTON_DEBOUNCE_MS)) {
            continue;
        }

        if (pressed != s_buttons[i].last_level) {
            s_buttons[i].last_debounce_tick = now;

            if (pressed) {
                /* Button just pressed */
                s_buttons[i].press_start_tick = now;
                s_buttons[i].is_pressed = true;
            } else {
                /* Button just released — classify the press */
                s_buttons[i].is_pressed = false;
                button_classify(i);
            }
        }
    }
}

/**
 * Classify the type of press based on timing.
 */
static void button_classify(button_id_t id)
{
    uint32_t press_duration = xTaskGetTickCount() - s_buttons[id].press_start_tick;
    button_event_t event = BUTTON_NONE;

    /* Check for double press */
    uint32_t time_since_release = xTaskGetTickCount() - s_buttons[id].release_tick;
    if (time_since_release < pdMS_TO_TICKS(BUTTON_DOUBLE_GAP_MS) &&
        s_buttons[id].press_count > 0) {
        event = BUTTON_DOUBLE_PRESS;
        s_buttons[id].press_count = 0;
    } else {
        s_buttons[id].press_count++;
        s_buttons[id].release_tick = xTaskGetTickCount();

        if (press_duration >= pdMS_TO_TICKS(BUTTON_LONG_PRESS_MS)) {
            event = BUTTON_LONG_PRESS;
        } else {
            event = BUTTON_SHORT_PRESS;
        }
    }

    /* Only update if we have a new event and previous was already consumed */
    if (event != BUTTON_NONE && s_event_consumed) {
        s_last_event = event;
        s_event_consumed = false;
    }
}

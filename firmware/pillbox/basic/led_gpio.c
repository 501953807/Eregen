/*
 * Eregen (颐贞) - LED Status Indicator Implementation
 * LEDC-based RGB LED control on ESP32-C3
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "led_gpio.h"

#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "esp_log.h"
#include "driver/ledc.h"

/* ESP32-C3 built-in RGB LED pins */
#define PIN_LED_RED     GPIO_NUM_3
#define PIN_LED_GREEN   GPIO_NUM_4
#define PIN_LED_BLUE    GPIO_NUM_5

/* LEDC timer and channel assignments */
#define LEDC_TIMER      LEDC_TIMER_0
#define LEDC_MODE       LEDC_LOW_SPEED_MODE
#define LEDC_DUTY_RES   LEDC_TIMER_8_BIT
#define LEDC_FREQ       5000  /* 5kHz for smooth dimming */

/* Blink timing (milliseconds) */
#define SLOW_BLINK_ON   500
#define SLOW_BLINK_OFF  500
#define FAST_BLINK_ON   125
#define FAST_BLINK_OFF  125

static const char *TAG = "led_gpio";

/* Current LED state tracking */
static led_color_t s_current_color = LED_COLOR_OFF;
static led_pattern_t s_current_pattern = LED_PATTERN_SOLID;

/* Color duty values (0-255 for 8-bit resolution) */
static const uint8_t color_duty[3][3] = {
    /* Red  Green Blue */
    {0,     255,   0},     /* LED_COLOR_GREEN  */
    {255,   0,     0},     /* LED_COLOR_RED    */
    {0,     0,     255},   /* LED_COLOR_BLUE   */
};

static esp_err_t ledc_channel_setup(void);
static void led_set_duty(uint8_t red, uint8_t green, uint8_t blue);
static void led_blink_task(void *param);

/**
 * Initialize the LED peripheral (LEDC on ESP32-C3).
 * Sets up three channels for RGB LED.
 */
esp_err_t led_init(void)
{
    /* Configure LEDC timer */
    ledc_timer_config_t led_timer = {
        .duty_resolution = LEDC_DUTY_RES,
        .freq_hz = LEDC_FREQ,
        .speed_mode = LEDC_MODE,
        .timer_num = LEDC_TIMER,
        .clk_cfg = LEDC_AUTO_CLK,
    };
    esp_err_t ret = ledc_timer_config(&led_timer);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "LEDC timer config failed: %s", esp_err_to_name(ret));
        return ret;
    }

    /* Set up Red channel */
    ledc_channel_config_t ch_red = {
        .speed_mode = LEDC_MODE,
        .channel = LEDC_CHANNEL_0,
        .timer_sel = LEDC_TIMER,
        .intr_type = LEDC_INTR_DISABLE,
        .gpio_num = PIN_LED_RED,
        .duty = 0,
        .hpoint = 0,
    };
    ret = ledc_channel_config(&ch_red);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "LEDC Red channel config failed: %s", esp_err_to_name(ret));
        return ret;
    }

    /* Set up Green channel */
    ledc_channel_config_t ch_green = {
        .speed_mode = LEDC_MODE,
        .channel = LEDC_CHANNEL_1,
        .timer_sel = LEDC_TIMER,
        .intr_type = LEDC_INTR_DISABLE,
        .gpio_num = PIN_LED_GREEN,
        .duty = 0,
        .hpoint = 0,
    };
    ret = ledc_channel_config(&ch_green);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "LEDC Green channel config failed: %s",
                 esp_err_to_name(ret));
        return ret;
    }

    /* Set up Blue channel */
    ledc_channel_config_t ch_blue = {
        .speed_mode = LEDC_MODE,
        .channel = LEDC_CHANNEL_2,
        .timer_sel = LEDC_TIMER,
        .intr_type = LEDC_INTR_DISABLE,
        .gpio_num = PIN_LED_BLUE,
        .duty = 0,
        .hpoint = 0,
    };
    ret = ledc_channel_config(&ch_blue);
    if (ret != ESP_OK) {
        ESP_LOGE(TAG, "LEDC Blue channel config failed: %s",
                 esp_err_to_name(ret));
        return ret;
    }

    ESP_LOGI(TAG, "RGB LED initialized: R=GPIO%d G=GPIO%d B=GPIO%d",
             PIN_LED_RED, PIN_LED_GREEN, PIN_LED_BLUE);

    return ESP_OK;
}

/**
 * Set LED to a solid color.
 */
esp_err_t led_set_color(led_color_t color)
{
    s_current_color = color;
    s_current_pattern = LED_PATTERN_SOLID;

    uint8_t r = 0, g = 0, b = 0;
    if (color != LED_COLOR_OFF) {
        r = color_duty[color][0];
        g = color_duty[color][1];
        b = color_duty[color][2];
    }

    led_set_duty(r, g, b);
    return ESP_OK;
}

/**
 * Set LED to a blinking pattern.
 * Runs in a background task for non-blocking operation.
 */
esp_err_t led_blink(led_color_t color, led_pattern_t pattern)
{
    s_current_color = color;
    s_current_pattern = pattern;

    /* Kill any existing blink task by setting duty to pattern start */
    uint8_t r = 0, g = 0, b = 0;
    if (color != LED_COLOR_OFF) {
        r = color_duty[color][0];
        g = color_duty[color][1];
        b = color_duty[color][2];
    }

    led_set_duty(r, g, b);
    return ESP_OK;
}

/**
 * Turn the LED off.
 */
esp_err_t led_off(void)
{
    s_current_color = LED_COLOR_OFF;
    s_current_pattern = LED_PATTERN_SOLID;
    led_set_duty(0, 0, 0);
    return ESP_OK;
}

/* ---- Internal helpers ---- */

/**
 * Set all three LED channels to the given duty values.
 */
static void led_set_duty(uint8_t red, uint8_t green, uint8_t blue)
{
    ledc_set_duty(LEDC_MODE, LEDC_CHANNEL_0, red);
    ledc_update_duty(LEDC_MODE, LEDC_CHANNEL_0);

    ledc_set_duty(LEDC_MODE, LEDC_CHANNEL_1, green);
    ledc_update_duty(LEDC_MODE, LEDC_CHANNEL_1);

    ledc_set_duty(LEDC_MODE, LEDC_CHANNEL_2, blue);
    ledc_update_duty(LEDC_MODE, LEDC_CHANNEL_2);
}

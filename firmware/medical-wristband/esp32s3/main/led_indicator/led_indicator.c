/*
 * Eregen Medical Wristband - LED Indicator Module
 * Color-coded alert LED based on patient condition
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "led_indicator.h"
#include <string.h>
#include "esp_log.h"
#include "driver/ledc.h"
#include "driver/gpio.h"

static const char *TAG = "led_indicator";

// LED GPIO pins (RGB)
#define LED_RED_PIN     GPIO_NUM_12
#define LED_GREEN_PIN   GPIO_NUM_13
#define LED_BLUE_PIN    GPIO_NUM_14

// LEDC settings
#define LEDC_TIMER      LEDC_TIMER_0
#define LEDC_MODE       LEDC_LOW_SPEED_MODE
#define LEDC_CHANNEL_RED    LEDC_CHANNEL_0
#define LEDC_CHANNEL_GREEN  LEDC_CHANNEL_1
#define LEDC_CHANNEL_BLUE   LEDC_CHANNEL_2
#define LEDC_FREQ       (3000)
#define LEDC_DUTY_RES   LEDC_TIMER_13_BIT
#define LEDC_DUTY       (8191)

static ledc_handle_t s_ledc_handle = NULL;

esp_err_t led_indicator_init(void) {
    ledc_timer_config_t timer_conf = {
        .speed_mode = LEDC_MODE,
        .timer_num = LEDC_TIMER,
        .duty_resolution = LEDC_DUTY_RES,
        .freq_hz = LEDC_FREQ,
        .clk_cfg = LEDC_AUTO_CLK,
    };
    ESP_ERROR_CHECK(ledc_timer_config(&timer_conf));

    // Configure RGB channels
    ledc_channel_config_t channel_conf[3] = {
        {LEDC_CHANNEL_RED,   LEDC_MODE, LEDC_TIMER, LED_RED_PIN,   0, LEDC_INTR_DISABLE},
        {LEDC_CHANNEL_GREEN, LEDC_MODE, LEDC_TIMER, LED_GREEN_PIN, 0, LEDC_INTR_DISABLE},
        {LEDC_CHANNEL_BLUE,  LEDC_MODE, LEDC_TIMER, LED_BLUE_PIN,  0, LEDC_INTR_DISABLE},
    };

    for (int i = 0; i < 3; i++) {
        ESP_ERROR_CHECK(ledc_channel_config(&channel_conf[i]));
    }

    ESP_LOGI(TAG, "LED indicator initialized");
    return ESP_OK;
}

esp_err_t led_indicator_set_color(led_color_t color) {
    uint32_t duty_r = 0, duty_g = 0, duty_b = 0;

    switch (color) {
        case LED_GREEN:  duty_g = LEDC_DUTY; break;
        case LED_RED:    duty_r = LEDC_DUTY; break;
        case LED_YELLOW: duty_r = LEDC_DUTY; duty_g = LEDC_DUTY; break;
        case LED_BLUE:   duty_b = LEDC_DUTY; break;
        case LED_WHITE:  duty_r = LEDC_DUTY; duty_g = LEDC_DUTY; duty_b = LEDC_DUTY; break;
        case LED_OFF:    duty_r = 0; duty_g = 0; duty_b = 0; break;
    }

    ledc_set_duty(LEDC_MODE, LEDC_CHANNEL_RED,   duty_r);
    ledc_set_duty(LEDC_MODE, LEDC_CHANNEL_GREEN, duty_g);
    ledc_set_duty(LEDC_MODE, LEDC_CHANNEL_BLUE,  duty_b);
    ledc_update_duty(LEDC_MODE, LEDC_CHANNEL_RED);
    ledc_update_duty(LEDC_MODE, LEDC_CHANNEL_GREEN);
    ledc_update_duty(LEDC_MODE, LEDC_CHANNEL_BLUE);

    return ESP_OK;
}

esp_err_t led_indicator_blink(led_color_t color, int times, int interval_ms) {
    for (int i = 0; i < times; i++) {
        led_indicator_set_color(color);
        vTaskDelay(pdMS_TO_TICKS(interval_ms));
        led_indicator_set_color(LED_OFF);
        vTaskDelay(pdMS_TO_TICKS(interval_ms));
    }
    return ESP_OK;
}

void led_indicator_set_alert(const char *alert_type) {
    if (strstr(alert_type, "SOS") || strstr(alert_type, "emergency")) {
        // Rapid red blink for emergency
        led_indicator_blink(LED_RED, 10, 100);
    } else if (strstr(alert_type, "fall")) {
        // Yellow blink for fall detection
        led_indicator_blink(LED_YELLOW, 5, 200);
    } else if (strstr(alert_type, "medication")) {
        // Blue blink for medication alert
        led_indicator_blink(LED_BLUE, 3, 300);
    }
}

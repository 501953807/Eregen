/*
 * Eregen (颐贞) - Universal Motor Control Implementation
 * Controls 28BYJ-48 stepper motor via ULN2003 driver.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "motor_control.h"

#include "driver/gpio.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"

#include "esp_log.h"

static const char *TAG = "motor";

/* Step sequence for 28BYJ-48 (4-phase full-step) */
static const uint8_t seq[8][4] = {
    {1, 0, 0, 0},  /* Phase A */
    {1, 1, 0, 0},  /* Phase A+B */
    {0, 1, 0, 0},  /* Phase B */
    {0, 1, 1, 0},  /* Phase B+C */
    {0, 0, 1, 0},  /* Phase C */
    {0, 0, 1, 1},  /* Phase C+D */
    {0, 0, 0, 1},  /* Phase D */
    {1, 0, 0, 1},  /* Phase D+A */
};

/* Internal state */
static motor_dir_t s_direction     = MOTOR_DIR_CW;
static uint32_t    s_rpm            = 10;
static int32_t     s_position       = 0;
static bool        s_running        = false;

static void motor_set_pin(uint8_t idx, uint8_t level);

/**
 * Initialize motor control GPIO pins.
 */
esp_err_t motor_init(void)
{
    gpio_config_t io_conf = {
        .pin_bit_mask = (1ULL << MOTOR_PIN_IN1) |
                        (1ULL << MOTOR_PIN_IN2) |
                        (1ULL << MOTOR_PIN_IN3) |
                        (1ULL << MOTOR_PIN_IN4),
        .mode         = GPIO_MODE_OUTPUT,
        .pull_up_en   = GPIO_PULLUP_DISABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type    = GPIO_INTR_DISABLE,
    };

    esp_err_t ret = gpio_config(&io_conf);
    if (ret == ESP_OK) {
        motor_stop();
        ESP_LOGI(TAG, "Motor initialized: IN1=%d IN2=%d IN3=%d IN4=%d",
                 MOTOR_PIN_IN1, MOTOR_PIN_IN2,
                 MOTOR_PIN_IN3, MOTOR_PIN_IN4);
    }
    return ret;
}

/**
 * Set motor speed in RPM.
 */
esp_err_t motor_set_speed(uint32_t rpm)
{
    if (rpm < MOTOR_MIN_RPM || rpm > MOTOR_MAX_RPM) {
        ESP_LOGE(TAG, "RPM out of range: %u (min=%d, max=%d)",
                 rpm, MOTOR_MIN_RPM, MOTOR_MAX_RPM);
        return ESP_ERR_INVALID_ARG;
    }
    s_rpm = rpm;
    return ESP_OK;
}

/**
 * Set motor rotation direction.
 */
esp_err_t motor_set_direction(motor_dir_t dir)
{
    s_direction = dir;
    return ESP_OK;
}

/**
 * Move the motor a specified number of steps.
 * Blocks until movement completes.
 */
esp_err_t motor_step(int32_t steps)
{
    if (steps == 0) return ESP_OK;

    bool forward = (s_direction == MOTOR_DIR_CW);
    int32_t remaining = steps > 0 ? steps : -steps;
    uint8_t idx = 0;

    /* Calculate delay per half-step in microseconds */
    /* RPM = revolutions/min => half-steps/min = RPM * 8
       => half-step period = 60e6 / (RPM * 8) us */
    uint32_t delay_us = 60000000UL / ((uint32_t)s_rpm * 8);
    if (delay_us < 50) delay_us = 50;   /* Clamp minimum */

    s_running = true;

    for (int32_t i = 0; i < (int32_t)remaining; i++) {
        if (!s_running) return ESP_ERR_INVALID_STATE;

        uint8_t phase[4];
        if (forward) {
            phase[0] = seq[idx][0];
            phase[1] = seq[idx][1];
            phase[2] = seq[idx][2];
            phase[3] = seq[idx][3];
        } else {
            /* Reverse sequence */
            phase[0] = seq[idx][0];
            phase[1] = seq[idx][2];
            phase[2] = seq[idx][1];
            phase[3] = seq[idx][3];
        }

        motor_set_pin(0, phase[0]);
        motor_set_pin(1, phase[1]);
        motor_set_pin(2, phase[2]);
        motor_set_pin(3, phase[3]);

        idx++;
        if (idx >= 8) idx = 0;

        vTaskDelay(us_to_ticks(delay_us));

        if (forward) {
            s_position++;
        } else {
            s_position--;
        }
    }

    motor_stop();
    return ESP_OK;
}

/**
 * Stop the motor immediately.
 */
esp_err_t motor_stop(void)
{
    s_running = false;
    motor_set_pin(0, 0);
    motor_set_pin(1, 0);
    motor_set_pin(2, 0);
    motor_set_pin(3, 0);
    return ESP_OK;
}

/**
 * Get the current software-tracked position.
 */
int32_t motor_get_position(void)
{
    return s_position;
}

/* ---- Internal helpers ---- */

static void motor_set_pin(uint8_t idx, uint8_t level)
{
    gpio_num_t pins[] = { MOTOR_PIN_IN1, MOTOR_PIN_IN2,
                          MOTOR_PIN_IN3, MOTOR_PIN_IN4 };
    gpio_set_level(pins[idx], level);
}

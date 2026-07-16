/*
 * Eregen (颐贞) - Pillbox Motor Control Implementation
 * Controls 28BYJ-48 stepper motor via ULN2003 driver.
 *
 * Compatible with ESP-IDF and standalone host compilation (TEST_MODE).
 *
 * 2026 Eregen (颐贞). All rights reserved.
 */

#include "motor_control.h"

#ifdef TEST_MODE
#include <stdio.h>
#include <unistd.h>
#else
#include "esp_log.h"
#include "driver/gpio.h"
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#endif

/* Step sequence for 28BYJ-48 (8-step microstepping for smoother motion) */
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
static bool s_ready        = true;
static int32_t s_position  = 0;
static uint8_t s_rpm       = MOTOR_DEFAULT_RPM;

#ifdef TEST_MODE
/* Mock: track whether step() was called with valid range */
bool s_step_called = false;
int32_t s_step_steps = 0;

/* Mock: force motor into stuck/busy state for testing error paths */
bool s_motor_stuck = false;
#endif

/**
 * Set a single GPIO output pin level.
 * In TEST_MODE this is a no-op; in ESP-IDF it drives real hardware.
 */
static void set_pin(uint8_t pin_idx, uint8_t level)
{
#ifdef TEST_MODE
    (void)pin_idx;
    (void)level;
#else
    static const gpio_num_t pins[] = {
        GPIO_NUM_0, GPIO_NUM_1, GPIO_NUM_2, GPIO_NUM_3
    };
    gpio_set_level(pins[pin_idx], level);
#endif
}

/**
 * Initialize motor control GPIO pins.
 */
void motor_control_init(void)
{
#ifndef TEST_MODE
    gpio_config_t io_conf = {
        .pin_bit_mask = (1ULL << GPIO_NUM_0) |
                        (1ULL << GPIO_NUM_1) |
                        (1ULL << GPIO_NUM_2) |
                        (1ULL << GPIO_NUM_3),
        .mode         = GPIO_MODE_OUTPUT,
        .pull_up_en   = GPIO_PULLUP_DISABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type    = GPIO_INTR_DISABLE,
    };
    gpio_config(&io_conf);
#endif

    s_ready      = true;
    s_position   = 0;
    s_rpm        = MOTOR_DEFAULT_RPM;

    /* Ensure all pins are off */
    set_pin(0, 0);
    set_pin(1, 0);
    set_pin(2, 0);
    set_pin(3, 0);

#ifdef TEST_MODE
    s_step_called = false;
    s_step_steps  = 0;
    s_motor_stuck = false;
#endif

#ifndef TEST_MODE
    ESP_LOGI("motor", "Motor control initialized");
#endif
}

/**
 * Move the motor a specified number of full steps.
 * Uses 8-step microstepping sequence (each logical step = 2 physical pulses).
 * Blocks until movement completes.
 *
 * @param steps Number of logical steps to move (0-255)
 * @return true if success, false if motor busy or invalid input
 */
bool motor_control_step(uint8_t steps)
{
    if (!s_ready) {
#ifndef TEST_MODE
        ESP_LOGW("motor", "Motor not ready, ignoring %u steps", steps);
#endif
        return false;
    }

    if (steps == 0) {
        return true;
    }

    /* Safety limit check */
    if ((uint32_t)steps > MOTOR_MAX_STEPS) {
#ifndef TEST_MODE
        ESP_LOGE("motor", "Steps %u exceeds max %d", steps, MOTOR_MAX_STEPS);
#endif
        return false;
    }

    s_ready     = false;
    s_step_called = true;
    s_step_steps  = steps;

    /* Calculate delay per half-step in microseconds.
     * RPM = revolutions/min => half-steps/min = RPM * 8
     * => half-step period = 60e6 / (RPM * 8) us */
    uint32_t delay_us = 60000000UL / ((uint32_t)s_rpm * 8);
    if (delay_us < 50) delay_us = 50;  /* Clamp minimum */

    for (uint8_t i = 0; i < steps; i++) {
        for (uint8_t phase = 0; phase < 8; phase++) {
            uint8_t p[4];
            p[0] = seq[phase][0];
            p[1] = seq[phase][1];
            p[2] = seq[phase][2];
            p[3] = seq[phase][3];

            set_pin(0, p[0]);
            set_pin(1, p[1]);
            set_pin(2, p[2]);
            set_pin(3, p[3]);

#ifdef TEST_MODE
            /* Simulate delay without blocking real OS */
            usleep((useconds_t)delay_us);
#else
            vTaskDelay(us_to_ticks(delay_us));
#endif
        }

        s_position++;
    }

    /* Turn off all coils after movement */
    set_pin(0, 0);
    set_pin(1, 0);
    set_pin(2, 0);
    set_pin(3, 0);

    s_ready = true;
    return true;
}

/**
 * Reset software position counter to zero (home).
 * Does not physically move the motor.
 */
void motor_control_home(void)
{
    s_position = 0;
    s_ready = true;

#ifndef TEST_MODE
    ESP_LOGI("motor", "Motor homed, position reset to 0");
#endif
}

/**
 * Check if the motor is idle (not currently moving).
 */
bool motor_control_is_ready(void)
{
    /* In TEST_MODE with stuck flag, always report busy */
#ifdef TEST_MODE
    if (s_motor_stuck) return false;
#endif
    return s_ready;
}

/* ---- Test-mode mock hooks ---- */

#ifdef TEST_MODE

void motor_control_mock_set_stuck(bool stuck)
{
    s_motor_stuck = stuck;
    if (stuck) {
        s_ready = false;
    } else {
        s_ready = true;
    }
}

#endif /* TEST_MODE */

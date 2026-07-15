/*
 * Eregen (颐贞) - Universal Motor Control Interface
 * Controls 28BYJ-48 stepper motor via ULN2003 driver using ESP32-C3 LEDC/PWM.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef MOTOR_CONTROL_H
#define MOTOR_CONTROL_H

#include "esp_err.h"
#include <stdint.h>

/* GPIO pin assignments for 28BYJ-48 via ULN2003 on ESP32-C3 */
#define MOTOR_PIN_IN1       GPIO_NUM_0
#define MOTOR_PIN_IN2       GPIO_NUM_1
#define MOTOR_PIN_IN3       GPIO_NUM_2
#define MOTOR_PIN_IN4       GPIO_NUM_3

/* 28BYJ-48 specs: 64 internal steps x 64:1 gear ratio = 4096 steps/rev */
#define STEPS_PER_REV       4096

/* Speed range: 1-100 RPM */
#define MOTOR_MIN_RPM       1
#define MOTOR_MAX_RPM       100

/* Direction enum */
typedef enum {
    MOTOR_DIR_CW,   /* Clockwise */
    MOTOR_DIR_CCW   /* Counter-clockwise */
} motor_dir_t;

/**
 * Initialize the motor control subsystem.
 * Configures GPIO pins for ULN2003 driver.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t motor_init(void);

/**
 * Set motor speed in RPM.
 *
 * @param rpm Speed in rotations per minute (1-100)
 * @return ESP_OK on success, ESP_ERR_INVALID_ARG if out of range
 */
esp_err_t motor_set_speed(uint32_t rpm);

/**
 * Set motor rotation direction.
 *
 * @param dir CW or CCW
 * @return ESP_OK on success
 */
esp_err_t motor_set_direction(motor_dir_t dir);

/**
 * Move the motor a specified number of steps.
 * Blocks until movement completes.
 *
 * @param steps Number of steps to move (positive = CW, negative = CCW)
 * @return ESP_OK on success, error code if interrupted
 */
esp_err_t motor_step(int32_t steps);

/**
 * Stop the motor immediately.
 *
 * @return ESP_OK on success
 */
esp_err_t motor_stop(void);

/**
 * Get the current software-tracked position in steps from origin.
 *
 * @return Current position in steps (signed, may be negative)
 */
int32_t motor_get_position(void);

#endif /* MOTOR_CONTROL_H */

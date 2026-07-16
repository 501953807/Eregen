/*
 * Eregen (颐贞) - Pillbox Motor Control Interface
 * Controls 28BYJ-48 stepper motor via ULN2003 driver using ESP32-C3 LEDC/PWM.
 *
 * Compatible with ESP-IDF and standalone host compilation (TEST_MODE).
 *
 * 28BYJ-48 specs: 64 internal steps x 64:1 gear ratio = 4096 steps/rev.
 *
 * 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef MOTOR_CONTROL_H
#define MOTOR_CONTROL_H

#include <stdint.h>
#include <stdbool.h>

/* Default speed: 10 RPM */
#define MOTOR_DEFAULT_RPM     10

/* Maximum steps allowed per single call (safety limit) */
#define MOTOR_MAX_STEPS       16384

/**
 * Initialize the motor control subsystem.
 * Configures GPIO pins (or mock equivalents under TEST_MODE).
 */
void motor_control_init(void);

/**
 * Move the motor a specified number of steps.
 * Blocks until movement completes.
 *
 * @param steps Number of steps to move (positive = CW, negative = CCW)
 * @return true if success, false if steps exceed safety limit or motor stuck
 */
bool motor_control_step(uint8_t steps);

/**
 * Reset software position counter to zero (home).
 * Does not physically move the motor.
 */
void motor_control_home(void);

/**
 * Check if the motor is idle (not currently moving).
 *
 * @return true if motor is ready for a new command
 */
bool motor_control_is_ready(void);

/* ---- Test-mode hooks (only available when compiled with TEST_MODE) ---- */

#ifdef TEST_MODE

/**
 * Simulate a stuck/busy motor for testing the error path.
 * When true, motor_control_is_ready() always returns false
 * and motor_control_step() always returns false.
 */
void motor_control_mock_set_stuck(bool stuck);

#endif /* TEST_MODE */

#endif /* MOTOR_CONTROL_H */

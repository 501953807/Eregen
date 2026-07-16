/*
 * Eregen (颐贞) - Auto Dispensing Module
 * Auto pillbox tier — complete dispensing flow:
 *   1. Rotate motor to target compartment
 *   2. Confirm position with photoelectric sensor
 *   3. Open dispensing door
 *   4. Wait for user to take medication (timeout configurable)
 *   5. Report success or miss
 *
 * Compile (host):  gcc -DTEST_MODE -I. dispensing.c opto_sensor.c -o dispensing_test
 * Compile (ESP32): idf_component_register with motor_control.h, opto_sensor.h
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef DISPENSING_H
#define DISPENSING_H

#include <stdint.h>
#include <stdbool.h>
#include "state_machine.h"

/* Number of compartments in the rotary tray */
#define DISPENSING_MAX_COMPARTMENTS  8

/* Default timeout for user to take medication (milliseconds) */
#define DISPENSE_DEFAULT_TIMEOUT_MS  (5 * 60 * 1000)  /* 5 minutes */

/* Steps per compartment on the rotary tray */
#define STEPS_PER_COMPARTMENT        512

/* Dispensing result codes */
typedef enum {
    DISPENSE_OK = 0,       /* Medication taken successfully */
    DISPENSE_TIMEOUT,      /* User did not take medication within timeout */
    DISPENSE_JAM,          /* Motor jam detected (position mismatch) */
    DISPENSE_EMPTY         /* Compartment was already empty */
} dispense_result_t;

/**
 * Dispense medication from a specific compartment.
 *
 * Complete dispensing flow:
 *   1. Transition state machine to DISPENSING
 *   2. Rotate motor to target compartment
 *   3. Verify position via opto_sensor
 *   4. Open dispensing door
 *   5. Wait for user to remove medication (opto_sensor polling)
 *   6. On removal → DISPENSE_OK, on timeout → retry once → DISPENSE_TIMEOUT
 *
 * @param compartment  Compartment index (0 to DISPENSING_MAX_COMPARTMENTS-1)
 * @param dose_count   Number of pills to dispense (1-4)
 * @return dispense_result_t indicating outcome
 */
dispense_result_t dispense_medication(uint8_t compartment, uint8_t dose_count);

/**
 * Cancel an ongoing dispensing operation.
 * Closes the dispensing door and returns to IDLE state.
 *
 * @return true if cancellation was successful
 */
bool dispense_cancel(void);

/**
 * Get the configured timeout value in milliseconds.
 *
 * @return Timeout in ms (default 300000 = 5 minutes)
 */
uint32_t dispense_get_timeout(void);

/**
 * Set the dispensing timeout value.
 *
 * @param ms Timeout in milliseconds
 */
void dispense_set_timeout(uint32_t ms);

#ifdef TEST_MODE
/**
 * Mock: override the delay function used during dispensing.
 * Allows tests to control timing without actually sleeping.
 */
void dispensing_set_mock_delay(void (*fn)(uint32_t));

/**
 * Mock: set a callback invoked on each poll iteration during
 * wait_for_removal(). Allows tests to inject state changes
 * (e.g., simulate medication being taken after N polls).
 */
void dispensing_set_mock_poll_hook(void (*fn)(void));
#endif

#endif /* DISPENSING_H */

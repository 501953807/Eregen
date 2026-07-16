/*
 * Eregen (颐贞) - Photoelectric Sensor Driver
 * Auto pillbox tier — IR beam-break detection for medication removal.
 *
 * Hardware: ITR2800E photoelectric sensor on ESP32-C3 GPIO.
 * Active-low interrupt: medication present = low beam, removed = high interrupt.
 *
 * Compile (host):  gcc -DTEST_MODE -I. opto_sensor.c -o opto_sensor_test
 * Compile (ESP32): idf_component_register with driver/gpio.h
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef OPTO_SENSOR_H
#define OPTO_SENSOR_H

#include <stdint.h>
#include <stdbool.h>

/* Debounce time in milliseconds to avoid false interrupts */
#define OPTO_DEBOUNCE_MS   50

/**
 * Initialize the photoelectric sensor.
 * In ESP-IDF: configures GPIO pin as input with pull-up and installs ISR.
 * In TEST_MODE: resets internal mock state.
 *
 * @return true if initialized successfully, false on failure
 */
bool opto_sensor_init(void);

/**
 * Read the current sensor state.
 *
 * @return true if medication has been REMOVED (beam broken),
 *         false if medication is PRESENT (beam intact).
 */
bool opto_sensor_read(void);

/**
 * Get the last known state (for debugging / logging).
 *
 * @return true = beam broken (medication removed), false = beam intact
 */
bool opto_sensor_get_last_state(void);

/**
 * Reset the internal state to simulate "medication present" (beam intact).
 * Call after dispensing to re-arm the sensor.
 */
void opto_sensor_reset(void);

#ifdef TEST_MODE
/**
 * Mock: set the simulated sensor state directly.
 * Used by unit tests to simulate medication being placed/removed.
 *
 * @param broken true = beam broken (medication removed), false = beam intact
 */
void opto_sensor_set_mock_state(bool broken);

/**
 * Mock: get the current simulated sensor state.
 */
bool opto_sensor_get_mock_state(void);
#endif

#endif /* OPTO_SENSOR_H */

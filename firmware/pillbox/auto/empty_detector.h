/*
 * Eregen (颐贞) - Empty Compartment Detector
 * Auto pillbox tier — detects which compartments are empty (no medication).
 *
 * Iterates through all compartments using the photoelectric sensor.
 * A compartment is considered empty if the sensor does not detect
 * any medication present at that position.
 *
 * Compile (host):  gcc -DTEST_MODE -I. empty_detector.c opto_sensor.c -o empty_detector_test
 * Compile (ESP32): idf_component_register with opto_sensor.h, motor_control.h
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef EMPTY_DETECTOR_H
#define EMPTY_DETECTOR_H

#include <stdint.h>
#include <stdbool.h>

/* Maximum number of compartments supported */
#define EMPTY_DETECTOR_MAX_COMPARTMENTS  8

/* Steps per compartment on the rotary tray */
#define EMPTY_DETECTOR_STEPS_PER_COMPARTMENT  512

/**
 * Check all compartments and return a bitmap of empty ones.
 * Each bit corresponds to a compartment: bit 0 = compartment 0, etc.
 *
 * @return Bitmask where 1 means empty, 0 means has medication.
 *         E.g., 0x05 means compartments 0 and 2 are empty.
 */
uint8_t empty_check_all_compartments(void);

/**
 * Check a single compartment for medication presence.
 *
 * @param compartment Compartment index (0 to EMPTY_DETECTOR_MAX_COMPARTMENTS-1)
 * @return true if compartment is EMPTY (no medication), false if has medication
 */
bool empty_check_single(uint8_t compartment);

/**
 * Get the number of empty compartments out of the total.
 *
 * @return Count of empty compartments (0 to EMPTY_DETECTOR_MAX_COMPARTMENTS)
 */
uint8_t empty_check_count(void);

#ifdef TEST_MODE
/**
 * Mock: pre-set the empty status of a specific compartment.
 * Used by unit tests to simulate known empty/full states.
 *
 * @param compartment Compartment index
 * @param empty       true = simulate empty, false = simulate full
 */
void empty_detector_set_mock_status(uint8_t compartment, bool empty);

/**
 * Mock: clear all mock statuses (reset to "all full").
 */
void empty_detector_clear_mock(void);
#endif

#endif /* EMPTY_DETECTOR_H */

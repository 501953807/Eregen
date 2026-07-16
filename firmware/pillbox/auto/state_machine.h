/*
 * Eregen (颐贞) - Pillbox Auto Tier State Machine Header
 * Manages the lifecycle of automatic medication dispensing.
 *
 * Compatible with ESP-IDF and standalone host compilation (TEST_MODE).
 *
 * 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef STATE_MACHINE_H
#define STATE_MACHINE_H

#include <stdint.h>
#include <stdbool.h>

/* ---- States ---- */

typedef enum {
    STATE_BOOT = 0,
    STATE_CONNECT,
    STATE_IDLE,
    STATE_REMINDER,
    STATE_DISPENSING,
    STATE_DETECT,
    STATE_REPORT,
    STATE_ERROR
} pillbox_state_t;

/* ---- Error codes ---- */

typedef enum {
    ERR_NONE = 0,
    ERR_MOTOR_STUCK,
    ERR_MED_JAM,
    ERR_SENSOR_FAIL,
    ERR_EMPTY_COMPARTMENT
} pillbox_error_t;

/* ---- Context ---- */

typedef struct {
    pillbox_state_t current_state;
    pillbox_error_t last_error;
    uint8_t current_compartment;
    uint8_t current_dose;
    bool error_occurred;
} pillbox_context_t;

/**
 * Initialize the state machine and reset all context to boot defaults.
 */
void state_machine_init(void);

/**
 * Run one tick of the state machine.
 * Transitions between states based on internal conditions.
 *
 * @return The new state after this tick.
 */
pillbox_state_t state_machine_run(void);

/**
 * Force a state transition (used for error recovery or manual overrides).
 * Any state can transition to ERROR; ERROR can transition to IDLE.
 * Normal transitions follow the state machine diagram in the spec.
 *
 * @param new_state Target state to transition to
 * @return true if transition accepted, false if invalid
 */
bool state_machine_transition(pillbox_state_t new_state);

/**
 * Get the last recorded error code.
 *
 * @return Error code (ERR_NONE if no error)
 */
pillbox_error_t state_machine_get_last_error(void);

/**
 * Clear the error flag and reset context to idle.
 */
void state_machine_clear_error(void);

/* ---- Test-mode accessors (only available when compiled with TEST_MODE) ---- */

#ifdef TEST_MODE

/**
 * Get a pointer to the internal state machine context.
 * Only available when compiled with TEST_MODE for testing purposes.
 */
pillbox_context_t *state_machine_get_context(void);

/**
 * Mock control: set WiFi connected status.
 */
void state_machine_mock_set_wifi_connected(bool connected);

/**
 * Mock control: set MQTT connected status.
 */
void state_machine_mock_set_mqtt_connected(bool connected);

/**
 * Mock control: set reminder ready flag.
 */
void state_machine_mock_set_reminder_ready(bool ready);

/**
 * Mock control: set dispensing done flag.
 */
void state_machine_mock_set_dispensing_done(bool done);

/**
 * Mock control: set detection done flag.
 */
void state_machine_mock_set_detection_done(bool done);

/**
 * Mock control: set report done flag.
 */
void state_machine_mock_set_report_done(bool done);

/**
 * Mock control: inject an error on next tick.
 */
void state_machine_mock_inject_error(pillbox_error_t error_code);

/**
 * Test-mode helper: force the state machine to a specific state,
 * bypassing the normal transition table. Useful for tests that
 * directly exercise dispense functions without running the full SM.
 */
void state_machine_force_state(pillbox_state_t state);

#endif /* TEST_MODE */

#endif /* STATE_MACHINE_H */

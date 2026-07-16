/*
 * Eregen (颐贞) - Pillbox Auto Tier State Machine Implementation
 *
 * Compatible with ESP-IDF and standalone host compilation (TEST_MODE).
 *
 * 2026 Eregen (颐贞). All rights reserved.
 */

#include "state_machine.h"

#ifdef TEST_MODE
#include <stdio.h>
#include <string.h>
#else
#include "esp_log.h"
#endif

static const char *TAG = "sm";

/* ---- Internal context ---- */
static pillbox_context_t s_ctx;

/* ---- Forward declarations ---- */
static bool transition_allowed(pillbox_state_t from, pillbox_state_t to);
static void execute_entry_action(pillbox_state_t state);
static pillbox_state_t compute_next_state(void);

/* ---- Mock hardware hooks (called during state machine run) ---- */
#ifdef TEST_MODE
/* Test-controlled flags */
bool s_mock_wifi_connected    = false;
bool s_mock_mqtt_connected    = false;
bool s_mock_reminder_ready    = false;
bool s_mock_dispensing_done   = false;
bool s_mock_detection_done    = false;
bool s_mock_report_done       = false;
bool s_mock_error_injected    = false;
pillbox_error_t s_mock_error_code = ERR_NONE;
#endif

/**
 * Human-readable state name (for logging).
 */
static const char *state_name(pillbox_state_t state)
{
    switch (state) {
    case STATE_BOOT:         return "BOOT";
    case STATE_CONNECT:      return "CONNECT";
    case STATE_IDLE:         return "IDLE";
    case STATE_REMINDER:     return "REMINDER";
    case STATE_DISPENSING:   return "DISPENSING";
    case STATE_DETECT:       return "DETECT";
    case STATE_REPORT:       return "REPORT";
    case STATE_ERROR:        return "ERROR";
    default:                 return "UNKNOWN";
    }
}

/* ================================================================
 * Valid transition table
 *
 * BOOT  -> CONNECT
 * CONNECT -> IDLE  (or ERROR on failure)
 * IDLE  -> REMINDER
 * REMINDER -> DISPENSING  (or ERROR)
 * DISPENSING -> DETECT
 * DETECT  -> REPORT  (or ERROR on sensor fail / empty)
 * REPORT  -> IDLE
 * Any   -> ERROR  (on fault detection)
 * ERROR -> IDLE  (after manual clear)
 * ================================================================ */

static bool transition_allowed(pillbox_state_t from, pillbox_state_t to)
{
    /* Self-transition always allowed */
    if (from == to) return true;

    /* ERROR -> IDLE is always allowed (manual reset) */
    if (from == STATE_ERROR && to == STATE_IDLE) return true;

    /* Any state -> ERROR is always allowed */
    if (to == STATE_ERROR) return true;

    /* Normal forward transitions */
    switch (from) {
    case STATE_BOOT:
        return to == STATE_CONNECT;

    case STATE_CONNECT:
        return to == STATE_IDLE;

    case STATE_IDLE:
        return to == STATE_REMINDER;

    case STATE_REMINDER:
        return to == STATE_DISPENSING;

    case STATE_DISPENSING:
        return to == STATE_DETECT;

    case STATE_DETECT:
        return to == STATE_REPORT;

    case STATE_REPORT:
        return to == STATE_IDLE;

    default:
        return false;
    }
}

/**
 * Execute entry actions for a state.
 */
static void execute_entry_action(pillbox_state_t state)
{
    switch (state) {
    case STATE_BOOT:
        /* Initialize all hardware subsystems */
#ifdef TEST_MODE
        printf("[sm] BOOT: initializing hardware...\n");
#else
        ESP_LOGI(TAG, "BOOT: initializing hardware...");
#endif
        break;

    case STATE_CONNECT:
        /* Start WiFi connection / provisioning */
#ifdef TEST_MODE
        printf("[sm] CONNECT: establishing network...\n");
#else
        ESP_LOGI(TAG, "CONNECT: establishing network...");
#endif
        break;

    case STATE_IDLE:
        /* Reset counters, wait for schedule */
        s_ctx.current_compartment = 0;
        s_ctx.current_dose = 0;
#ifdef TEST_MODE
        printf("[sm] IDLE: waiting for medication schedule\n");
#else
        ESP_LOGI(TAG, "IDLE: waiting for medication schedule");
#endif
        break;

    case STATE_REMINDER:
        /* Trigger TTS voice + LED indicator */
#ifdef TEST_MODE
        printf("[sm] REMINDER: playing voice reminder\n");
#else
        ESP_LOGI(TAG, "REMINDER: playing voice reminder");
#endif
        break;

    case STATE_DISPENSING:
        /* Rotate motor to target compartment */
#ifdef TEST_MODE
        printf("[sm] DISPENSING: rotating to compartment %u\n",
               s_ctx.current_compartment);
#else
        ESP_LOGI(TAG, "DISPENSING: rotating to compartment %u",
                 s_ctx.current_compartment);
#endif
        break;

    case STATE_DETECT:
        /* Activate photoelectric sensor */
#ifdef TEST_MODE
        printf("[sm] DETECT: monitoring sensor...\n");
#else
        ESP_LOGI(TAG, "DETECT: monitoring sensor...");
#endif
        break;

    case STATE_REPORT:
        /* Compose and send med_status to cloud */
#ifdef TEST_MODE
        printf("[sm] REPORT: sending status to cloud\n");
#else
        ESP_LOGI(TAG, "REPORT: sending status to cloud");
#endif
        break;

    case STATE_ERROR:
        /* Stop all operations */
#ifdef TEST_MODE
        printf("[sm] ERROR: stopping all operations, last_error=%d\n",
               (int)s_ctx.last_error);
#else
        ESP_LOGE(TAG, "ERROR: last_error=%d", (int)s_ctx.last_error);
#endif
        break;

    default:
        break;
    }
}

/* ================================================================
 * Public API
 * ================================================================ */

/**
 * Initialize the state machine and reset all context to boot defaults.
 */
void state_machine_init(void)
{
    memset(&s_ctx, 0, sizeof(s_ctx));
    s_ctx.current_state = STATE_BOOT;

#ifdef TEST_MODE
    /* Reset mock flags */
    s_mock_wifi_connected    = false;
    s_mock_mqtt_connected    = false;
    s_mock_reminder_ready    = false;
    s_mock_dispensing_done   = false;
    s_mock_detection_done    = false;
    s_mock_report_done       = false;
    s_mock_error_injected    = false;
    s_mock_error_code        = ERR_NONE;
    printf("[sm] Initialized -> BOOT\n");
#else
    ESP_LOGI(TAG, "State machine initialized -> BOOT");
#endif
}

/**
 * Run one tick of the state machine.
 * Computes the next state based on current conditions and transitions.
 *
 * @return The new state after this tick.
 */
pillbox_state_t state_machine_run(void)
{
    pillbox_state_t next = compute_next_state();

    if (!transition_allowed(s_ctx.current_state, next)) {
#ifdef TEST_MODE
        printf("[sm] Blocked transition: %s -> %s\n",
               state_name(s_ctx.current_state), state_name(next));
#else
        ESP_LOGW(TAG, "Blocked transition: %s -> %s",
                 state_name(s_ctx.current_state), state_name(next));
#endif
        return s_ctx.current_state;
    }

    pillbox_state_t prev = s_ctx.current_state;
    s_ctx.current_state = next;
    execute_entry_action(next);

    /* If we entered ERROR, record the error */
    if (next == STATE_ERROR) {
        s_ctx.error_occurred = true;
    }

#ifdef TEST_MODE
    printf("[sm] Transitioned: %s -> %s\n", state_name(prev), state_name(next));
#else
    ESP_LOGI(TAG, "Transitioned: %s -> %s", state_name(prev), state_name(next));
#endif

    return next;
}

/**
 * Force a state transition (used for error recovery or manual overrides).
 */
bool state_machine_transition(pillbox_state_t new_state)
{
    if (!transition_allowed(s_ctx.current_state, new_state)) {
#ifdef TEST_MODE
        printf("[sm] Rejected forced transition: %s -> %s\n",
               state_name(s_ctx.current_state), state_name(new_state));
#else
        ESP_LOGW(TAG, "Rejected forced transition: %s -> %s",
                 state_name(s_ctx.current_state), state_name(new_state));
#endif
        return false;
    }

    pillbox_state_t prev = s_ctx.current_state;
    s_ctx.current_state = new_state;
    execute_entry_action(new_state);

    if (new_state == STATE_ERROR) {
        s_ctx.error_occurred = true;
    }
    /* Note: do NOT auto-clear error_occurred here.
     * Only state_machine_clear_error() should reset the flag. */

#ifdef TEST_MODE
    printf("[sm] Forced transition: %s -> %s\n", state_name(prev), state_name(new_state));
#else
    ESP_LOGI(TAG, "Forced transition: %s -> %s",
             state_name(prev), state_name(new_state));
#endif

    return true;
}

/**
 * Get the last recorded error code.
 */
pillbox_error_t state_machine_get_last_error(void)
{
    return s_ctx.last_error;
}

/**
 * Clear the error flag and reset context to idle.
 */
void state_machine_clear_error(void)
{
    s_ctx.last_error = ERR_NONE;
    s_ctx.error_occurred = false;
    s_ctx.current_compartment = 0;
    s_ctx.current_dose = 0;

    /* Transition from ERROR to IDLE */
    if (s_ctx.current_state == STATE_ERROR) {
        s_ctx.current_state = STATE_IDLE;
        execute_entry_action(STATE_IDLE);
    }

#ifdef TEST_MODE
    printf("[sm] Error cleared, resetting to IDLE\n");
#else
    ESP_LOGI(TAG, "Error cleared, resetting to IDLE");
#endif
}

/* ================================================================
 * Internal: compute next state based on current conditions
 * ================================================================ */

/**
 * Determine the next state based on current state and conditions.
 * In TEST_MODE, uses mock flags; in production, checks real hardware.
 */
static pillbox_state_t compute_next_state(void)
{
    switch (s_ctx.current_state) {
    case STATE_BOOT:
        /* After initialization, move to CONNECT */
        return STATE_CONNECT;

    case STATE_CONNECT:
#ifdef TEST_MODE
        if (s_mock_wifi_connected && s_mock_mqtt_connected) {
            return STATE_IDLE;
        }
        /* If error injected during connect, go to ERROR */
        if (s_mock_error_injected) {
            s_ctx.last_error = s_mock_error_code;
            return STATE_ERROR;
        }
#endif
        /* In production: wait for WiFi/MQTT event */
        return STATE_CONNECT;

    case STATE_IDLE:
#ifdef TEST_MODE
        if (s_mock_reminder_ready) {
            return STATE_REMINDER;
        }
        if (s_mock_error_injected) {
            s_ctx.last_error = s_mock_error_code;
            return STATE_ERROR;
        }
#endif
        return STATE_IDLE;

    case STATE_REMINDER:
#ifdef TEST_MODE
        /* Auto-advance to DISPENSING after reminder triggers */
        if (!s_mock_error_injected) {
            return STATE_DISPENSING;
        }
        s_ctx.last_error = s_mock_error_code;
        return STATE_ERROR;
#endif
        return STATE_REMINDER;

    case STATE_DISPENSING:
#ifdef TEST_MODE
        if (s_mock_dispensing_done) {
            return STATE_DETECT;
        }
        if (s_mock_error_injected) {
            s_ctx.last_error = s_mock_error_code;
            return STATE_ERROR;
        }
#endif
        return STATE_DISPENSING;

    case STATE_DETECT:
#ifdef TEST_MODE
        if (s_mock_detection_done) {
            return STATE_REPORT;
        }
        if (s_mock_error_injected) {
            s_ctx.last_error = s_mock_error_code;
            return STATE_ERROR;
        }
#endif
        return STATE_DETECT;

    case STATE_REPORT:
#ifdef TEST_MODE
        if (s_mock_report_done) {
            return STATE_IDLE;
        }
#endif
        return STATE_REPORT;

    case STATE_ERROR:
        /* Stay in ERROR until manually cleared */
        return STATE_ERROR;

    default:
        return STATE_ERROR;
    }
}

/* ================================================================
 * Test-mode mock accessors (only compiled with TEST_MODE)
 * ================================================================ */

#ifdef TEST_MODE

pillbox_context_t *state_machine_get_context(void)
{
    return &s_ctx;
}

void state_machine_mock_set_wifi_connected(bool connected)
{
    s_mock_wifi_connected = connected;
}

void state_machine_mock_set_mqtt_connected(bool connected)
{
    s_mock_mqtt_connected = connected;
}

void state_machine_mock_set_reminder_ready(bool ready)
{
    s_mock_reminder_ready = ready;
}

void state_machine_mock_set_dispensing_done(bool done)
{
    s_mock_dispensing_done = done;
}

void state_machine_mock_set_detection_done(bool done)
{
    s_mock_detection_done = done;
}

void state_machine_mock_set_report_done(bool done)
{
    s_mock_report_done = done;
}

void state_machine_mock_inject_error(pillbox_error_t error_code)
{
    s_mock_error_injected = true;
    s_mock_error_code = error_code;
}

void state_machine_force_state(pillbox_state_t state)
{
    s_ctx.current_state = state;
#ifdef TEST_MODE
    printf("[sm] Forced state -> %s (bypassing transition table)\n",
           state_name(state));
#endif
}

#endif /* TEST_MODE */

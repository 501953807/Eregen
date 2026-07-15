/*
 * Eregen (颐贞) - Medication Reminder Scheduler
 * Smart pillbox tier — NVS-persisted scheduling with MQTT updates
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef REMINDER_SCHEDULER_H
#define REMINDER_SCHEDULER_H

#ifdef TEST_MODE
#include <stdint.h>
#include <stdbool.h>

typedef int esp_err_t;
#define ESP_OK            0
#define ESP_ERR_NO_MEM   (-5)
#define ESP_ERR_NOT_FOUND (-6)
#define ESP_ERR_INVALID_ARG (-11)
#define ESP_FAIL          (-1)
#else
#include "esp_err.h"
#endif

/* Maximum rules per device */
#define SCHEDULER_MAX_RULES     10

/* Time format: HH:MM stored as minutes since midnight (0-1439) */
typedef struct {
    uint8_t hour;       /* 0-23 */
    uint8_t minute;     /* 0-59 */
} reminder_time_t;

/* Medicine types */
typedef enum {
    MED_TYPE_CAPSULE = 0,
    MED_TYPE_TABLET,
    MED_TYPE_SYRUP,
    MED_TYPE_INJECTION,
    MED_TYPE_COUNT
} medicine_type_t;

/* A single reminder rule */
typedef struct {
    reminder_time_t time;       /* When to remind (HH:MM) */
    uint8_t dose_count;         /* Number of pills this dose */
    medicine_type_t med_type;   /* Type of medicine */
    uint8_t compartment_index;  /* Which compartment to open */
    bool enabled;               /* Rule active flag */
} reminder_rule_t;

/**
 * Initialize scheduler. Loads rules from NVS.
 *
 * @return ESP_OK on success, error code otherwise
 */
esp_err_t scheduler_init(void);

/**
 * Add a medication reminder rule.
 *
 * @param rule Pointer to rule to add (copied internally)
 * @return ESP_OK on success, ESP_ERR_NO_MEM if at capacity
 */
esp_err_t scheduler_add_rule(const reminder_rule_t *rule);

/**
 * Remove a rule by its index (0-based within internal array).
 *
 * @param index Rule index to remove
 * @return ESP_OK on success, ESP_ERR_NOT_FOUND if invalid index
 */
esp_err_t scheduler_remove_rule(uint8_t index);

/**
 * Check whether any rule is pending right now.
 * Compares current RTC time against all enabled rules.
 * Must be called every ~60 seconds.
 *
 * @param[out] out_rule If non-NULL, filled with the pending rule
 * @return true if a reminder is due, false otherwise
 */
bool scheduler_check_pending(reminder_rule_t *out_rule);

/**
 * Replace all rules at once (used for MQTT bulk update).
 * Clears existing rules then adds new ones.
 *
 * @param rules Array of rules
 * @param count Number of rules in array
 * @return ESP_OK on success
 */
esp_err_t scheduler_replace_rules(const reminder_rule_t *rules, uint8_t count);

/**
 * Get current number of stored rules.
 *
 * @return Rule count
 */
uint8_t scheduler_get_rule_count(void);

/**
 * Get a rule by index (for inspection / testing).
 *
 * @param index Rule index
 * @param[out] out_rule Filled with rule data
 * @return ESP_OK on success
 */
esp_err_t scheduler_get_rule(uint8_t index, reminder_rule_t *out_rule);

#endif /* REMINDER_SCHEDULER_H */

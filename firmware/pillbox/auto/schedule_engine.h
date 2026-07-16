/*
 * Eregen (颐贞) - Schedule Engine
 * Auto pillbox tier — manages medication reminder scheduling,
 * next-trigger computation, and acknowledgment.
 *
 * Compile (host):  gcc -DTEST_MODE -I. med_rule_parser.c schedule_engine.c -o schedule_test
 * Compile (ESP32): idf_component_register with esp_err.h
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef SCHEDULE_ENGINE_H
#define SCHEDULE_ENGINE_H

#include "med_rule_parser.h"
#include "nvs_store.h"
#include <stdint.h>

/**
 * Initialize the schedule engine.
 * Loads rules from NVS storage if available, otherwise starts empty.
 */
void schedule_engine_init(void);

/**
 * Reload rules from persistent storage (NVS or file-based mock).
 * Call after med_rule_parse to persist then reload.
 */
void schedule_engine_reload(void);

/**
 * Compute seconds until the next scheduled reminder.
 * Scans all enabled rules, finds the earliest one whose time
 * is >= current system time.
 *
 * @return Seconds until next reminder. If no pending rule, returns UINT32_MAX.
 */
uint32_t schedule_next_trigger(void);

/**
 * Check whether a reminder is currently triggered (reminder time reached).
 * Must be called periodically (e.g., every 60 seconds).
 *
 * @return true if a reminder is due, false otherwise.
 */
bool schedule_check_triggered(void);

/**
 * Acknowledge the current pending reminder.
 * Marks the most recently triggered rule as acknowledged so it
 * won't fire again until the next day's cycle.
 */
void schedule_acknowledge(void);

#endif /* SCHEDULE_ENGINE_H */

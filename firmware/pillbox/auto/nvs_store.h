/*
 * Eregen (颐贞) - NVS Storage Module
 * Auto pillbox tier — non-volatile storage for medication rules.
 * On ESP32-C3: uses native NVS flash.
 * In TEST_MODE: uses file-based mock storage.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef NVS_STORE_H
#define NVS_STORE_H

#include "med_rule_parser.h"
#include <stdbool.h>
#include <stdint.h>

/**
 * Initialize NVS storage namespace.
 * On ESP32: opens/initializes the "pillbox" NVS partition.
 * In TEST_MODE: no-op.
 */
void nvs_init(void);

/**
 * Save medication rules to non-volatile storage.
 * Serializes the rule array and stores it.
 *
 * @param rules  Array of rules to save
 * @param count  Number of rules
 * @return true on success, false on failure
 */
bool nvs_save_rules(const med_rule_t *rules, uint8_t count);

/**
 * Load medication rules from non-volatile storage.
 *
 * @param rules  Output buffer for loaded rules
 * @param count  On input: capacity; on output: actual number loaded
 * @return true on success, false if no data or error
 */
bool nvs_load_rules(med_rule_t *rules, uint8_t *count);

#endif /* NVS_STORE_H */

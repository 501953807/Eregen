/*
 * Eregen (颐贞) - Medication Rule Parser
 * Auto pillbox tier — parse downlink "med_rule" messages without external JSON libs
 *
 * Compile (host):  gcc -DTEST_MODE -I. med_rule_parser.c -o med_rule_parser_test
 * Compile (ESP32): idf_component_register with esp_err.h
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef MED_RULE_PARSER_H
#define MED_RULE_PARSER_H

#include <stdint.h>
#include <stdbool.h>

/* Maximum number of medication rules per device */
#define MED_RULE_PARSER_MAX_RULES  8

/* Medicine type enum — matches reminder_scheduler.h for consistency */
typedef enum {
    MED_TYPE_CAPSULE  = 0,
    MED_TYPE_TABLET   = 1,
    MED_TYPE_LIQUID   = 2,
    MED_TYPE_INJECTION = 3,
    MED_TYPE_COUNT
} med_type_t;

/* A single medication rule parsed from JSON */
typedef struct {
    uint8_t hour;       /* 0-23 */
    uint8_t minute;     /* 0-59 */
    uint8_t dose;       /* number of pills */
    uint8_t type;       /* med_type_t: capsule/tablet/liquid/injection */
    bool enabled;       /* rule active flag */
    char name[16];      /* medication name (optional, max 15 chars + NUL) */
} med_rule_t;

/**
 * Parse a JSON string containing medication rules.
 * Accepts the MQTT downlink format:
 *   {"type":"med_rule","dev_id":"PX-XXXX","rules":[...]}
 *
 * Only the "rules" array is extracted; "type" and "dev_id" are ignored
 * after validation that the message is a med_rule type.
 *
 * @param json_str   Null-terminated JSON string
 * @param rules      Output array to store parsed rules
 * @param max_rules  Capacity of the rules array (max MED_RULE_PARSER_MAX_RULES)
 * @return Number of rules successfully parsed (0 if none), negative on error:
 *         -1 = NULL input, -2 = not a med_rule message, -3 = no rules array,
 *         -4 = malformed rule entry, -5 = rule count exceeds capacity
 */
int med_rule_parse(const char *json_str, med_rule_t *rules, int max_rules);

/**
 * Get a pointer to the internal copy of a parsed rule by index.
 * Rules are stored internally after the last successful med_rule_parse call.
 *
 * @param index Rule index (0-based)
 * @return Pointer to the rule, or NULL if invalid index
 */
const med_rule_t* med_rule_get(uint8_t index);

/**
 * Get the number of rules currently stored in the internal buffer.
 *
 * @return Rule count
 */
uint8_t med_rule_count(void);

/**
 * Clear all internally stored rules.
 */
void med_rule_clear(void);

/**
 * Load rules directly from a raw binary blob (used by NVS store).
 * Not intended for general use — bypasses JSON parsing.
 *
 * @param src    Source array of rules
 * @param count  Number of rules to load
 */
void med_rule_load_raw(const med_rule_t *src, uint8_t count);

#endif /* MED_RULE_PARSER_H */

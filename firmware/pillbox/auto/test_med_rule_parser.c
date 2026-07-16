/*
 * Eregen (颐贞) - Medication Rule Parser & Schedule Engine Tests
 * Host-mode tests: valid JSON, invalid JSON, empty rules, boundary values.
 *
 * Compile:  gcc -DTEST_MODE -I. med_rule_parser.c nvs_store.c schedule_engine.c -o test_med_rule
 * Run:      ./test_med_rule
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <stdint.h>

#include "med_rule_parser.h"
#include "nvs_store.h"
#include "schedule_engine.h"

/* ---- Mock RTC epoch for schedule engine ---- */
uint32_t s_schedule_mock_epoch = 0;

/* ---- Test helpers ---- */
static int tests_run = 0;
static int tests_passed = 0;
static int test_assert_failures = 0;

#define TEST(name) \
    tests_run++; \
    printf("  TEST: %s ... ", #name); \
    test_assert_failures = 0;

#define ASSERT(cond, msg) \
    do { \
        if (!(cond)) { \
            printf("FAIL [%s]\n", msg); \
            test_assert_failures++; \
        } \
    } while(0)

#define CHECK_PASS() \
    do { \
        if (test_assert_failures == 0) { \
            tests_passed++; \
            printf("PASSED\n"); \
        } else { \
            printf("FAILED (%d assertions)\n", test_assert_failures); \
        } \
        break; \
    } while(0)

/* ---- Test cases ---- */

/* Test 1: Parse valid JSON with capsule rule */
static void test_parse_valid_capsule(void)
{
    TEST(parse_valid_capsule)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\",\"dev_id\":\"PX-ABCD1234\","
        "\"rules\":[{\"time\":\"08:00\",\"dose\":1,\"type\":\"capsule\",\"name\":\"降压药\"}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == 1, "count == 1");
    ASSERT(rules[0].hour == 8, "hour == 8");
    ASSERT(rules[0].minute == 0, "minute == 0");
    ASSERT(rules[0].dose == 1, "dose == 1");
    ASSERT(rules[0].type == MED_TYPE_CAPSULE, "type == CAPSULE");
    ASSERT(rules[0].enabled == true, "enabled == true");
    CHECK_PASS();
}

/* Test 2: Parse valid JSON with multiple rules of different types */
static void test_parse_multiple_rules(void)
{
    TEST(parse_multiple_rules)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\",\"dev_id\":\"PX-XXXX\","
        "\"rules\":["
        "{\"time\":\"07:00\",\"dose\":1,\"type\":\"capsule\",\"name\":\"维生素\"},"
        "{\"time\":\"12:00\",\"dose\":2,\"type\":\"tablet\",\"name\":\"钙片\"},"
        "{\"time\":\"21:00\",\"dose\":1,\"type\":\"liquid\",\"name\":\"止咳糖浆\"}"
        "]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == 3, "count == 3");
    ASSERT(rules[0].hour == 7 && rules[0].dose == 1 && rules[0].type == MED_TYPE_CAPSULE, "rule 0 fields");
    ASSERT(strcmp(rules[0].name, "维生素") == 0, "rule 0 name");
    ASSERT(rules[1].hour == 12 && rules[1].dose == 2 && rules[1].type == MED_TYPE_TABLET, "rule 1 fields");
    ASSERT(strcmp(rules[1].name, "钙片") == 0, "rule 1 name");
    ASSERT(rules[2].hour == 21 && rules[2].dose == 1 && rules[2].type == MED_TYPE_LIQUID, "rule 2 fields");
    ASSERT(strcmp(rules[2].name, "止咳糖浆") == 0, "rule 2 name");
    CHECK_PASS();
}

/* Test 3: Parse injection type */
static void test_parse_injection_type(void)
{
    TEST(parse_injection_type)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":[{\"time\":\"14:30\",\"dose\":1,\"type\":\"injection\"}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == 1, "count == 1");
    ASSERT(rules[0].type == MED_TYPE_INJECTION, "type == INJECTION");
    ASSERT(rules[0].hour == 14, "hour == 14");
    ASSERT(rules[0].minute == 30, "minute == 30");
    CHECK_PASS();
}

/* Test 4: Invalid JSON — missing type field */
static void test_parse_missing_type(void)
{
    TEST(parse_missing_type)
    med_rule_clear();

    const char *json =
        "{\"dev_id\":\"PX-XXXX\","
        "\"rules\":[{\"time\":\"08:00\",\"dose\":1}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -2, "count == -2 (not med_rule)");
    CHECK_PASS();
}

/* Test 5: Invalid JSON — wrong message type */
static void test_parse_wrong_type(void)
{
    TEST(parse_wrong_type)
    med_rule_clear();

    const char *json =
        "{\"type\":\"heartbeat\",\"dev_id\":\"PX-XXXX\",\"bat\":85}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -2, "count == -2 (wrong type)");
    CHECK_PASS();
}

/* Test 6: Invalid JSON — missing rules array */
static void test_parse_missing_rules_array(void)
{
    TEST(parse_missing_rules_array)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\",\"dev_id\":\"PX-XXXX\"}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -3, "count == -3 (no rules array)");
    CHECK_PASS();
}

/* Test 7: Empty rules array */
static void test_parse_empty_rules(void)
{
    TEST(parse_empty_rules)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\",\"dev_id\":\"PX-XXXX\",\"rules\":[]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -3, "count == -3 (empty array)");
    CHECK_PASS();
}

/* Test 8: Boundary values — hour 0 and 23 */
static void test_parse_boundary_hours(void)
{
    TEST(parse_boundary_hours)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":["
        "{\"time\":\"00:00\",\"dose\":1,\"type\":\"capsule\"},"
        "{\"time\":\"23:59\",\"dose\":2,\"type\":\"tablet\"}"
        "]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == 2, "count == 2");
    ASSERT(rules[0].hour == 0 && rules[0].minute == 0, "boundary 00:00");
    ASSERT(rules[1].hour == 23 && rules[1].minute == 59, "boundary 23:59");
    CHECK_PASS();
}

/* Test 9: Missing optional fields — dose defaults to 1, name defaults empty */
static void test_parse_minimal_rule(void)
{
    TEST(parse_minimal_rule)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":[{\"time\":\"06:00\"}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == 1, "count == 1");
    ASSERT(rules[0].hour == 6, "hour == 6");
    ASSERT(rules[0].minute == 0, "minute == 0");
    ASSERT(rules[0].dose == 1, "dose default == 1");
    ASSERT(rules[0].type == MED_TYPE_CAPSULE, "type default == CAPSULE");
    ASSERT(rules[0].enabled == true, "enabled default == true");
    ASSERT(strcmp(rules[0].name, "") == 0, "name default == \"\"");
    CHECK_PASS();
}

/* Test 10: Malformed time format */
static void test_parse_malformed_time(void)
{
    TEST(parse_malformed_time)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":[{\"time\":\"abc\",\"dose\":1}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -3, "count == -3 (no valid rules)");
    CHECK_PASS();
}

/* Test 11: Time out of range — hour 25 */
static void test_parse_hour_out_of_range(void)
{
    TEST(parse_hour_out_of_range)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":[{\"time\":\"25:00\",\"dose\":1}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -3, "count == -3 (hour out of range)");
    CHECK_PASS();
}

/* Test 12: Minute out of range — minute 60 */
static void test_parse_minute_out_of_range(void)
{
    TEST(parse_minute_out_of_range)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":[{\"time\":\"08:60\",\"dose\":1}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -3, "count == -3 (minute out of range)");
    CHECK_PASS();
}

/* Test 13: NULL input */
static void test_parse_null_input(void)
{
    TEST(parse_null_input)
    med_rule_clear();

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    int count = med_rule_parse(NULL, rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -1, "NULL json -> -1");

    count = med_rule_parse("", rules, MED_RULE_PARSER_MAX_RULES);
    ASSERT(count == -2, "empty json -> -2");

    count = med_rule_parse(NULL, NULL, 0);
    ASSERT(count == -1, "NULL args -> -1");
    CHECK_PASS();
}

/* Test 14: med_rule_get and med_rule_count after parse */
static void test_rule_accessors(void)
{
    TEST(rule_accessors)
    med_rule_clear();

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":["
        "{\"time\":\"08:00\",\"dose\":1,\"type\":\"capsule\",\"name\":\"阿司匹林\"},"
        "{\"time\":\"20:00\",\"dose\":2,\"type\":\"tablet\",\"name\":\"维生素D\"}"
        "]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);

    ASSERT(med_rule_count() == 2, "count == 2");

    const med_rule_t *r0 = med_rule_get(0);
    ASSERT(r0 != NULL, "get(0) not NULL");
    ASSERT(r0->hour == 8, "get(0)->hour == 8");
    ASSERT(strcmp(r0->name, "阿司匹林") == 0, "get(0)->name");

    const med_rule_t *r1 = med_rule_get(1);
    ASSERT(r1 != NULL, "get(1) not NULL");
    ASSERT(r1->hour == 20, "get(1)->hour == 20");
    ASSERT(strcmp(r1->name, "维生素D") == 0, "get(1)->name");

    const med_rule_t *r_invalid = med_rule_get(5);
    ASSERT(r_invalid == NULL, "get(5) == NULL");
    CHECK_PASS();
}

/* Test 15: NVS save and load (mock file-based) */
static void test_nvs_save_load(void)
{
    TEST(nvs_save_load)
    med_rule_clear();

    /* Prepare rules */
    med_rule_t rules[] = {
        {8, 0, 1, MED_TYPE_CAPSULE, true, "阿莫西林"},
        {12, 30, 2, MED_TYPE_TABLET, true, "二甲双胍"},
    };
    uint8_t count = 2;

    /* Save via nvs_store */
    bool saved = nvs_save_rules(rules, count);
    ASSERT(saved == true, "save succeeded");

    /* Clear internal state */
    med_rule_clear();
    ASSERT(med_rule_count() == 0, "count cleared");

    /* Load back */
    med_rule_t loaded_rules[MED_RULE_PARSER_MAX_RULES];
    uint8_t loaded_count = 0;
    bool load_ok = nvs_load_rules(loaded_rules, &loaded_count);
    ASSERT(load_ok == true, "load succeeded");
    ASSERT(loaded_count == 2, "loaded_count == 2");

    ASSERT(loaded_rules[0].hour == 8, "loaded[0].hour");
    ASSERT(loaded_rules[0].dose == 1, "loaded[0].dose");
    ASSERT(strcmp(loaded_rules[0].name, "阿莫西林") == 0, "loaded[0].name");

    ASSERT(loaded_rules[1].hour == 12, "loaded[1].hour");
    ASSERT(loaded_rules[1].minute == 30, "loaded[1].minute");
    ASSERT(loaded_rules[1].dose == 2, "loaded[1].dose");
    ASSERT(strcmp(loaded_rules[1].name, "二甲双胍") == 0, "loaded[1].name");
    CHECK_PASS();
}

/* Test 16: Schedule engine — next trigger calculation */
static void test_schedule_next_trigger(void)
{
    TEST(schedule_next_trigger)
    med_rule_clear();

    /* Set mock time to 07:00 on day 0 */
    s_schedule_mock_epoch = 7 * 3600;

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":["
        "{\"time\":\"08:00\",\"dose\":1,\"type\":\"capsule\"},"
        "{\"time\":\"12:00\",\"dose\":2,\"type\":\"tablet\"}"
        "]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);

    uint32_t next = schedule_next_trigger();
    ASSERT(next == 3600, "next_trigger == 3600s (1h until 08:00)");
    CHECK_PASS();
}

/* Test 17: Schedule engine — check triggered */
static void test_schedule_check_triggered(void)
{
    TEST(schedule_check_triggered)
    med_rule_clear();

    s_schedule_mock_epoch = 8 * 3600; /* 08:00 */

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":[{\"time\":\"08:00\",\"dose\":1,\"type\":\"capsule\"}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);

    bool triggered = schedule_check_triggered();
    ASSERT(triggered == true, "triggered at rule time");
    CHECK_PASS();
}

/* Test 18: Schedule engine — acknowledge clears trigger */
static void test_schedule_acknowledge(void)
{
    TEST(schedule_acknowledge)
    med_rule_clear();

    s_schedule_mock_epoch = 8 * 3600;

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":[{\"time\":\"08:00\",\"dose\":1,\"type\":\"capsule\"}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);

    /* First check should trigger */
    bool triggered = schedule_check_triggered();
    ASSERT(triggered == true, "first check triggers");

    /* Acknowledge */
    schedule_acknowledge();

    /* Next trigger should be UINT32_MAX since today is done */
    uint32_t next = schedule_next_trigger();
    ASSERT(next == UINT32_MAX, "after ack, next == UINT32_MAX");
    CHECK_PASS();
}

/* Test 19: Schedule engine — no pending rules */
static void test_schedule_no_pending(void)
{
    TEST(schedule_no_pending)
    med_rule_clear();

    s_schedule_mock_epoch = 12 * 3600; /* noon, no rules */

    uint32_t next = schedule_next_trigger();
    ASSERT(next == UINT32_MAX, "no rules -> UINT32_MAX");

    bool triggered = schedule_check_triggered();
    ASSERT(triggered == false, "not triggered with no rules");
    CHECK_PASS();
}

/* Test 20: Schedule engine — day rollover resets acknowledgment */
static void test_schedule_day_rollover(void)
{
    TEST(schedule_day_rollover)
    med_rule_clear();

    s_schedule_mock_epoch = 8 * 3600;

    const char *json =
        "{\"type\":\"med_rule\","
        "\"rules\":[{\"time\":\"08:00\",\"dose\":1,\"type\":\"capsule\"}]}";

    med_rule_t rules[MED_RULE_PARSER_MAX_RULES];
    med_rule_parse(json, rules, MED_RULE_PARSER_MAX_RULES);

    schedule_check_triggered();
    schedule_acknowledge();

    /* Move to next day */
    s_schedule_mock_epoch = 24 * 3600 + 8 * 3600; /* day 1, 08:00 */

    bool triggered = schedule_check_triggered();
    ASSERT(triggered == true, "triggers again on new day");
    CHECK_PASS();
}

/* ---- Main ---- */
int main(void)
{
    printf("\n=== Eregen Auto Pillbox — Med Rule Parser & Schedule Engine Tests ===\n\n");

    /* Parser tests */
    test_parse_valid_capsule();
    test_parse_multiple_rules();
    test_parse_injection_type();
    test_parse_missing_type();
    test_parse_wrong_type();
    test_parse_missing_rules_array();
    test_parse_empty_rules();
    test_parse_boundary_hours();
    test_parse_minimal_rule();
    test_parse_malformed_time();
    test_parse_hour_out_of_range();
    test_parse_minute_out_of_range();
    test_parse_null_input();
    test_rule_accessors();

    /* NVS store tests */
    test_nvs_save_load();

    /* Schedule engine tests */
    test_schedule_next_trigger();
    test_schedule_check_triggered();
    test_schedule_acknowledge();
    test_schedule_no_pending();
    test_schedule_day_rollover();

    printf("\n=== Results: %d/%d tests passed ===\n", tests_passed, tests_run);

    return (tests_passed == tests_run) ? 0 : 1;
}

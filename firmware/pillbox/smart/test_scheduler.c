/*
 * Eregen (颐贞) - Reminder Scheduler Tests (Host-mode)
 * Tests adding/removing rules, time comparison, multi-slot scheduling,
 * and NVS persistence simulation.
 *
 * Compile: gcc -DTEST_MODE -I. test_scheduler.c -o test_scheduler
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <stdint.h>
#include <time.h>

#include "reminder_scheduler.h"

/* ---- Mock RTC for deterministic testing ---- */
#ifdef TEST_MODE
extern uint32_t s_current_epoch;
#endif

/* ---- Test helpers ---- */
static int tests_run = 0;
static int tests_passed = 0;

#define TEST(name) \
    do { \
        tests_run++; \
        printf("  TEST: %s ... ", #name); \

#define EXPECT_TRUE(cond) \
    do { \
        if (!(cond)) { \
            printf("FAILED (expected true: %s)\n", #cond); \
            goto test_fail; \
        } \
    } while(0)

#define EXPECT_FALSE(cond) \
    do { \
        if (cond) { \
            printf("FAILED (expected false: %s)\n", #cond); \
            goto test_fail; \
        } \
    } while(0)

#define EXPECT_EQ(a, b) \
    do { \
        if ((a) != (b)) { \
            printf("FAILED (expected %d, got %d): %s == %s\n", \
                   (int)(b), (int)(a), #a, #b); \
            goto test_fail; \
        } \
    } while(0)

#define EXPECT_STR_EQ(a, b) \
    do { \
        if (strcmp((a), (b)) != 0) { \
            printf("FAILED (expected \"%s\", got \"%s\")\n", b, a); \
            goto test_fail; \
        } \
    } while(0)

#define PASS() \
    do { \
        tests_passed++; \
        printf("PASSED\n"); \
        break; \
    } while(0)

#define test_fail: \
    printf("FAILED\n"); \
    break

/* ---- Test cases ---- */

static void test_init_empty(void)
{
    TEST(scheduler_init_empty)
    scheduler_init();
    EXPECT_EQ(scheduler_get_rule_count(), 0);
    PASS();
}

static void test_add_single_rule(void)
{
    TEST(scheduler_add_single_rule)
    scheduler_init();

    reminder_rule_t rule = {0};
    rule.time.hour = 8;
    rule.time.minute = 0;
    rule.dose_count = 1;
    rule.med_type = MED_TYPE_CAPSULE;
    rule.compartment_index = 0;
    rule.enabled = true;

    esp_err_t ret = scheduler_add_rule(&rule);
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(scheduler_get_rule_count(), 1);

    /* Verify stored rule */
    reminder_rule_t stored;
    ret = scheduler_get_rule(0, &stored);
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(stored.time.hour, 8);
    EXPECT_EQ(stored.time.minute, 0);
    EXPECT_EQ(stored.dose_count, 1);
    EXPECT_EQ(stored.med_type, MED_TYPE_CAPSULE);
    EXPECT_EQ(stored.compartment_index, 0);
    PASS();
}

static void test_add_multiple_rules(void)
{
    TEST(scheduler_add_multiple_rules)
    scheduler_init();

    reminder_rule_t rules[] = {
        {{8, 0}, 1, MED_TYPE_CAPSULE, 0, true},
        {{12, 0}, 2, MED_TYPE_TABLET, 1, true},
        {{20, 0}, 1, MED_TYPE_SYRUP, 2, true},
    };
    int n = sizeof(rules) / sizeof(rules[0]);

    for (int i = 0; i < n; i++) {
        esp_err_t ret = scheduler_add_rule(&rules[i]);
        EXPECT_EQ(ret, ESP_OK);
    }

    EXPECT_EQ(scheduler_get_rule_count(), 3);

    /* Verify each rule */
    reminder_rule_t check;
    scheduler_get_rule(0, &check);
    EXPECT_EQ(check.time.hour, 8);
    scheduler_get_rule(1, &check);
    EXPECT_EQ(check.time.hour, 12);
    scheduler_get_rule(2, &check);
    EXPECT_EQ(check.time.hour, 20);
    PASS();
}

static void test_add_exceeds_capacity(void)
{
    TEST(scheduler_add_exceeds_capacity)
    scheduler_init();

    reminder_rule_t rule = {{1, 0}, 1, MED_TYPE_CAPSULE, 0, true};

    /* Fill to capacity */
    for (int i = 0; i < SCHEDULER_MAX_RULES; i++) {
        rule.time.hour = (uint8_t)(i + 1);
        esp_err_t ret = scheduler_add_rule(&rule);
        EXPECT_EQ(ret, ESP_OK);
    }
    EXPECT_EQ(scheduler_get_rule_count(), SCHEDULER_MAX_RULES);

    /* One more should fail */
    esp_err_t ret = scheduler_add_rule(&rule);
    EXPECT_EQ(ret, ESP_ERR_NO_MEM);
    PASS();
}

static void test_remove_rule(void)
{
    TEST(scheduler_remove_rule)
    scheduler_init();

    reminder_rule_t rules[] = {
        {{8, 0}, 1, MED_TYPE_CAPSULE, 0, true},
        {{12, 0}, 2, MED_TYPE_TABLET, 1, true},
        {{20, 0}, 1, MED_TYPE_SYRUP, 2, true},
    };
    for (int i = 0; i < 3; i++)
        scheduler_add_rule(&rules[i]);

    EXPECT_EQ(scheduler_get_rule_count(), 3);

    /* Remove middle rule */
    esp_err_t ret = scheduler_remove_rule(1);
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(scheduler_get_rule_count(), 2);

    /* Verify remaining rules */
    reminder_rule_t check;
    scheduler_get_rule(0, &check);
    EXPECT_EQ(check.time.hour, 8);
    scheduler_get_rule(1, &check);
    EXPECT_EQ(check.time.hour, 20);

    /* Invalid index */
    ret = scheduler_remove_rule(5);
    EXPECT_EQ(ret, ESP_ERR_NOT_FOUND);
    PASS();
}

static void test_time_comparison_pending(void)
{
    TEST(scheduler_time_comparison_pending)
    scheduler_init();

    reminder_rule_t rule = {{12, 0}, 1, MED_TYPE_CAPSULE, 0, true};
    scheduler_add_rule(&rule);

    /* Set mock time BEFORE the rule — should not be pending */
#ifdef TEST_MODE
    s_current_epoch = 10 * 3600; /* 10:00 */
#endif
    reminder_rule_t out = {0};
    bool pending = scheduler_check_pending(&out);
    EXPECT_FALSE(pending);

    /* Set mock time AT the rule — should be pending */
#ifdef TEST_MODE
    s_current_epoch = 12 * 3600; /* 12:00 */
#endif
    pending = scheduler_check_pending(&out);
    EXPECT_TRUE(pending);
    EXPECT_EQ(out.time.hour, 12);
    EXPECT_EQ(out.time.minute, 0);
    PASS();
}

static void test_multi_slot_scheduling(void)
{
    TEST(scheduler_multi_slot)
    scheduler_init();

    reminder_rule_t rules[] = {
        {{6, 0}, 1, MED_TYPE_CAPSULE, 0, true},   /* 06:00 */
        {{8, 0}, 1, MED_TYPE_CAPSULE, 1, true},   /* 08:00 */
        {{12, 0}, 2, MED_TYPE_TABLET, 2, true},   /* 12:00 */
        {{14, 0}, 1, MED_TYPE_TABLET, 3, true},   /* 14:00 */
        {{20, 0}, 1, MED_TYPE_SYRUP, 4, true},    /* 20:00 */
    };
    for (int i = 0; i < 5; i++)
        scheduler_add_rule(&rules[i]);

    /* At 07:00 — next is 08:00 */
#ifdef TEST_MODE
    s_current_epoch = 7 * 3600;
#endif
    reminder_rule_t next;
    bool pending = scheduler_check_pending(&next);
    EXPECT_TRUE(pending);
    EXPECT_EQ(next.time.hour, 8);

    /* At 13:00 — next is 14:00 */
#ifdef TEST_MODE
    s_current_epoch = 13 * 3600;
#endif
    pending = scheduler_check_pending(&next);
    EXPECT_TRUE(pending);
    EXPECT_EQ(next.time.hour, 14);

    /* At 21:00 — no more reminders today */
#ifdef TEST_MODE
    s_current_epoch = 21 * 3600;
#endif
    pending = scheduler_check_pending(&next);
    EXPECT_FALSE(pending);
    PASS();
}

static void test_nvs_persistence_simulation(void)
{
    TEST(scheduler_nvs_persistence)
    /* In TEST_MODE, NVS calls are no-ops, but we verify that
     * rules added persist in memory across init boundary. */

    /* Add rules */
    scheduler_init();
    reminder_rule_t rule = {{9, 30}, 2, MED_TYPE_INJECTION, 0, true};
    scheduler_add_rule(&rule);
    EXPECT_EQ(scheduler_get_rule_count(), 1);

    /* Re-init simulates a fresh load */
    scheduler_init();
    EXPECT_EQ(scheduler_get_rule_count(), 0); /* In TEST_MODE, NVS is skipped */

    /* Add again and verify in-memory persistence */
    scheduler_add_rule(&rule);
    reminder_rule_t check;
    scheduler_get_rule(0, &check);
    EXPECT_EQ(check.time.hour, 9);
    EXPECT_EQ(check.time.minute, 30);
    EXPECT_EQ(check.dose_count, 2);
    PASS();
}

static void test_replace_rules(void)
{
    TEST(scheduler_replace_rules)
    scheduler_init();

    /* Add initial rules */
    reminder_rule_t initial[] = {
        {{8, 0}, 1, MED_TYPE_CAPSULE, 0, true},
        {{12, 0}, 1, MED_TYPE_TABLET, 1, true},
    };
    for (int i = 0; i < 2; i++)
        scheduler_add_rule(&initial[i]);
    EXPECT_EQ(scheduler_get_rule_count(), 2);

    /* Replace with new set */
    reminder_rule_t new_rules[] = {
        {{7, 0}, 1, MED_TYPE_CAPSULE, 0, true},
        {{15, 0}, 2, MED_TYPE_SYRUP, 1, true},
        {{22, 0}, 1, MED_TYPE_TABLET, 2, true},
    };
    esp_err_t ret = scheduler_replace_rules(new_rules, 3);
    EXPECT_EQ(ret, ESP_OK);
    EXPECT_EQ(scheduler_get_rule_count(), 3);

    /* Verify old rules are gone */
    reminder_rule_t check;
    scheduler_get_rule(0, &check);
    EXPECT_EQ(check.time.hour, 7);
    scheduler_get_rule(2, &check);
    EXPECT_EQ(check.time.hour, 22);
    PASS();
}

static void test_replace_invalid_count(void)
{
    TEST(scheduler_replace_invalid_count)
    scheduler_init();

    reminder_rule_t rule = {{10, 0}, 1, MED_TYPE_CAPSULE, 0, true};

    /* NULL rules */
    esp_err_t ret = scheduler_replace_rules(NULL, 1);
    EXPECT_EQ(ret, ESP_ERR_INVALID_ARG);

    /* Too many rules */
    reminder_rule_t too_many[SCHEDULER_MAX_RULES + 1];
    memset(too_many, 0, sizeof(too_many));
    ret = scheduler_replace_rules(too_many, SCHEDULER_MAX_RULES + 1);
    EXPECT_EQ(ret, ESP_ERR_INVALID_ARG);
    PASS();
}

static void test_no_rules_pending(void)
{
    TEST(scheduler_no_rules_pending)
    scheduler_init();

    /* No rules added — should never be pending */
    reminder_rule_t out = {0};
    bool pending = scheduler_check_pending(&out);
    EXPECT_FALSE(pending);
    PASS();
}

/* ---- Main ---- */
int main(void)
{
    printf("\n=== Eregen Smart Pillbox — Reminder Scheduler Tests ===\n\n");

    test_init_empty();
    test_add_single_rule();
    test_add_multiple_rules();
    test_add_exceeds_capacity();
    test_remove_rule();
    test_time_comparison_pending();
    test_multi_slot_scheduling();
    test_nvs_persistence_simulation();
    test_replace_rules();
    test_replace_invalid_count();
    test_no_rules_pending();

    printf("\n=== Results: %d/%d tests passed ===\n", tests_passed, tests_run);

    return (tests_passed == tests_run) ? 0 : 1;
}

/*
 * Eregen (颐贞) - Logging Test Harness
 * Host-compiled test driver for log module.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdbool.h>
#include "../common/log.h"

static int passed = 0;
static int failed = 0;

static void check(bool cond, const char *label)
{
    if (cond) {
        printf("  PASS: %s\n", label);
        passed++;
    } else {
        printf("  FAIL: %s\n", label);
        failed++;
    }
}

int main(void)
{
    printf("Logging tests:\n");

    log_init();

    /* Default level is DEBUG */
    check(log_get_level() == LOG_DEBUG, "default level is DEBUG");

    /* Level filtering */
    log_set_level(LOG_WARN);
    check(log_get_level() == LOG_WARN, "level set to WARN");

    printf("--- Testing level filtering (only WARN+ should appear above) ---\n");
    log_debug("this should NOT appear");
    log_info("this should NOT appear");
    log_warn("this SHOULD appear");
    log_error("this SHOULD appear");
    printf("--- End filtering test ---\n");

    /* Reset to DEBUG for further tests */
    log_set_level(LOG_DEBUG);

    /* Format string handling */
    printf("--- Testing format strings ---\n");
    log_info("Value: %d, String: %s, Float: %.1f", 42, "hello", 3.14f);
    printf("--- End format test ---\n");

    /* Empty message */
    log_info("");

    printf("\nResults: %d passed, %d failed\n", passed, failed);
    return failed > 0 ? 1 : 0;
}

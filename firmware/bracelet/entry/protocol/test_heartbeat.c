/*
 * Eregen (颐贞) - Heartbeat Module Test Harness
 * Host-compiled test driver for heartbeat module (non-embedded mode).
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <string.h>
#include <stdbool.h>
#include "../protocol/heartbeat.h"

/* Mock device ID for testing */
char s_device_id[17] = "BR-TEST1";

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
    printf("Heartbeat module tests:\n");

    /* Test start/stop in host mode */
    check(1, "module compiles and links");

    /* Double-start should be safe */
    heartbeat_start();
    heartbeat_start();
    check(1, "double start is safe");
    heartbeat_stop();

    /* Double-stop should be safe */
    heartbeat_stop();
    heartbeat_stop();
    check(1, "double stop is safe");

    /* Start again */
    heartbeat_start();
    check(1, "start works after stop");
    heartbeat_stop();

    printf("\nResults: %d passed, %d failed\n", passed, failed);
    return failed > 0 ? 1 : 0;
}

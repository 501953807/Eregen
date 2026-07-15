/*
 * Eregen (颐贞) - CRC16-CCITT Implementation
 * Polynomial: 0x1021, Initial value: 0xFFFF
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "crc16.h"

/* Test harness entry point when compiled as host test */
#ifdef TEST_MODE
int crc16_test(void);
int main(void) { return crc16_test(); }
#endif

uint16_t crc16_calc(const uint8_t *data, uint16_t len)
{
    if (!data || len == 0) {
        return 0;
    }

    uint16_t crc = 0xFFFF;

    for (uint16_t i = 0; i < len; i++) {
        crc ^= ((uint16_t)data[i] << 8);
        for (uint8_t j = 0; j < 8; j++) {
            if (crc & 0x8000) {
                crc = (crc << 1) ^ 0x1021;
            } else {
                crc = crc << 1;
            }
        }
    }

    return crc;
}

#ifdef TEST_MODE
#include <stdio.h>
#include <string.h>

static int passed = 0;
static int failed = 0;

static void check(const char *label, uint16_t actual, uint16_t expected)
{
    if (actual == expected) {
        printf("  PASS: %s\n", label);
        passed++;
    } else {
        printf("  FAIL: %s (got 0x%04X, expected 0x%04X)\n", label, actual, expected);
        failed++;
    }
}

int crc16_test(void)
{
    printf("CRC16-CCITT tests:\n");

    /* Empty input returns 0 */
    check("empty input", crc16_calc(NULL, 0), 0);

    /* Known test vector: "123456789" -> 0x29B1 */
    const uint8_t vec[] = "123456789";
    check("standard vector", crc16_calc(vec, sizeof(vec) - 1), 0x29B1);

    /* Single byte */
    check("single byte 0x00", crc16_calc((const uint8_t *)"\x00", 1), 0xD0CB);
    check("single byte 0xFF", crc16_calc((const uint8_t *)"\xFF", 1), 0x44C9);

    /* All zeros */
    uint8_t zeros[256];
    memset(zeros, 0, sizeof(zeros));
    check("256 zero bytes", crc16_calc(zeros, 256), 0xAEFC);

    printf("\nResults: %d passed, %d failed\n", passed, failed);
    return failed > 0 ? 1 : 0;
}
#endif

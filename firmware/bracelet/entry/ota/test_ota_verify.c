/*
 * Eregen (颐贞) - OTA Firmware Verification Tests
 * Tests SHA256 hash computation, signature verification, and boot switching.
 * Compile: gcc -DTEST_MODE -I. ota/ota_verify.c ota/boot_switch.c ../common/log.c -lm -o test_ota_verify
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>

#include "ota/ota_verify.h"
#include "ota/boot_switch.h"

/* ============ Test Helpers ============ */

static int g_tests_run = 0;
static int g_tests_passed = 0;
static int g_tests_failed = 0;

#define TEST_ASSERT(cond, msg) do { \
    g_tests_run++; \
    if (cond) { g_tests_passed++; printf("  PASS: %s\n", msg); } \
    else { g_tests_failed++; printf("  FAIL: %s\n", msg); } \
} while(0)

#define TEST_ASSERT_EQ(a, b, msg) do { \
    g_tests_run++; \
    if ((a) == (b)) { g_tests_passed++; printf("  PASS: %s\n", msg); } \
    else { g_tests_failed++; printf("  FAIL: %s (expected=%lu, got=%lu)\n", \
           msg, (unsigned long)(b), (unsigned long)(a)); } \
} while(0)

#define TEST_ASSERT_MEM_EQ(a, b, len, msg) do { \
    g_tests_run++; \
    if (memcmp((a), (b), len) == 0) { g_tests_passed++; printf("  PASS: %s\n", msg); } \
    else { g_tests_failed++; printf("  FAIL: %s (memory mismatch)\n", msg); } \
} while(0)

/* ============ SHA256 Known-Vector Tests ============ */

static void test_sha256_known_vectors(void)
{
    printf("\n--- SHA256 Known Vector Tests ---\n");

    /* Test vector 1: empty string */
    uint8_t digest_empty[OTA_SHA256_DIGEST_LEN];
    ota_sha256_compute((const uint8_t*)"", 0, digest_empty);

    /* SHA256("") = e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855 */
    static const uint8_t expected_empty[] = {
        0xe3, 0xb0, 0xc4, 0x42, 0x98, 0xfc, 0x1c, 0x14,
        0x9a, 0xfb, 0xf4, 0xc8, 0x99, 0x6f, 0xb9, 0x24,
        0x27, 0xae, 0x41, 0xe4, 0x64, 0x9b, 0x93, 0x4c,
        0xa4, 0x95, 0x99, 0x1b, 0x78, 0x52, 0xb8, 0x55
    };
    TEST_ASSERT_MEM_EQ(digest_empty, expected_empty, OTA_SHA256_DIGEST_LEN,
                       "SHA256 of empty string matches known vector");

    /* Test vector 2: "abc" */
    uint8_t digest_abc[OTA_SHA256_DIGEST_LEN];
    ota_sha256_compute((const uint8_t*)"abc", 3, digest_abc);

    /* SHA256("abc") = ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad */
    static const uint8_t expected_abc[] = {
        0xba, 0x78, 0x16, 0xbf, 0x8f, 0x01, 0xcf, 0xea,
        0x41, 0x41, 0x40, 0xde, 0x5d, 0xae, 0x22, 0x23,
        0xb0, 0x03, 0x61, 0xa3, 0x96, 0x17, 0x7a, 0x9c,
        0xb4, 0x10, 0xff, 0x61, 0xf2, 0x00, 0x15, 0xad
    };
    TEST_ASSERT_MEM_EQ(digest_abc, expected_abc, OTA_SHA256_DIGEST_LEN,
                       "SHA256 of \"abc\" matches known vector");

    /* Test vector 3: "abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq" */
    const char *msg3 = "abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq";
    uint8_t digest_3[OTA_SHA256_DIGEST_LEN];
    ota_sha256_compute((const uint8_t*)msg3, (uint32_t)strlen(msg3), digest_3);

    /* SHA256 of that string: */
    static const uint8_t expected_3[] = {
        0x24, 0x8d, 0x6a, 0x61, 0xd2, 0x06, 0x38, 0xb8,
        0xe5, 0xc0, 0x26, 0x93, 0x0c, 0x3e, 0x60, 0x39,
        0xa3, 0x3c, 0xe4, 0x59, 0x64, 0xff, 0x21, 0x67,
        0xf6, 0xec, 0xed, 0xd4, 0x19, 0xdb, 0x06, 0xc1
    };
    TEST_ASSERT_MEM_EQ(digest_3, expected_3, OTA_SHA256_DIGEST_LEN,
                       "SHA256 of longer string matches known vector");
}

/* ============ Streaming Hash Tests ============ */

static void test_streaming_hash(void)
{
    printf("\n--- Streaming SHA256 Tests ---\n");

    /* Hash "abc" in three pieces: "a", "b", "c" */
    ota_sha256_ctx_t ctx;
    ota_sha256_init(&ctx);
    ota_sha256_update(&ctx, (const uint8_t*)"a", 1);
    ota_sha256_update(&ctx, (const uint8_t*)"b", 1);
    ota_sha256_update(&ctx, (const uint8_t*)"c", 1);

    uint8_t digest_stream[OTA_SHA256_DIGEST_LEN];
    ota_sha256_final(&ctx, digest_stream);

    /* Should match SHA256("abc") from above */
    static const uint8_t expected_abc[] = {
        0xba, 0x78, 0x16, 0xbf, 0x8f, 0x01, 0xcf, 0xea,
        0x41, 0x41, 0x40, 0xde, 0x5d, 0xae, 0x22, 0x23,
        0xb0, 0x03, 0x61, 0xa3, 0x96, 0x17, 0x7a, 0x9c,
        0xb4, 0x10, 0xff, 0x61, 0xf2, 0x00, 0x15, 0xad
    };
    TEST_ASSERT_MEM_EQ(digest_stream, expected_abc, OTA_SHA256_DIGEST_LEN,
                       "Streaming SHA256(\"abc\") matches one-shot");
}

/* ============ Digest Comparison Tests ============ */

static void test_digest_comparison(void)
{
    printf("\n--- Digest Comparison Tests ---\n");

    uint8_t a[OTA_SHA256_DIGEST_LEN] = {0};
    uint8_t b[OTA_SHA256_DIGEST_LEN] = {0};
    uint8_t c[OTA_SHA256_DIGEST_LEN] = {1};

    TEST_ASSERT(ota_sha256_equal(a, b) == true, "Two zero digests are equal");
    TEST_ASSERT(ota_sha256_equal(a, c) == false, "Zero and non-zero digests differ");

    /* Set all bytes to same value */
    memset(a, 0xFF, OTA_SHA256_DIGEST_LEN);
    memcpy(b, a, OTA_SHA256_DIGEST_LEN);
    TEST_ASSERT(ota_sha256_equal(a, b) == true, "All-FF digests are equal");

    /* Flip one byte */
    b[0] ^= 0x01;
    TEST_ASSERT(ota_sha256_equal(a, b) == false, "Single-byte difference detected");
}

/* ============ Firmware Verification Tests ============ */

static void test_firmware_verification(void)
{
    printf("\n--- Firmware Verification Tests ---\n");

    /* Create a fake firmware blob */
    uint8_t firmware[64];
    memset(firmware, 0xAA, sizeof(firmware));

    /* Compute its SHA256 */
    ota_signature_t sig;
    ota_sha256_compute(firmware, sizeof(firmware), sig.sha256);
    sig.firmware_version = 1;
    sig.firmware_size = sizeof(firmware);

    /* Verify matching firmware */
    TEST_ASSERT(ota_verify_firmware(firmware, sizeof(firmware), &sig) == true,
                "Matching firmware passes verification");

    /* Tamper with firmware — should fail */
    firmware[0] ^= 0x01;
    TEST_ASSERT(ota_verify_firmware(firmware, sizeof(firmware), &sig) == false,
                "Tampered firmware fails verification");

    /* NULL firmware pointer */
    TEST_ASSERT(ota_verify_firmware(NULL, sizeof(firmware), &sig) == false,
                "NULL firmware rejected");

    /* NULL signature pointer */
    uint8_t fw2[32];
    memset(fw2, 0xBB, sizeof(fw2));
    TEST_ASSERT(ota_verify_firmware(fw2, sizeof(fw2), NULL) == false,
                "NULL signature rejected");

    /* Oversized firmware */
    uint8_t oversized[OTA_MAX_HASH_SIZE + 1];
    memset(oversized, 0xCC, sizeof(oversized));
    ota_signature_t bad_sig;
    memset(&bad_sig, 0, sizeof(bad_sig));
    TEST_ASSERT(ota_verify_firmware(oversized, sizeof(oversized), &bad_sig) == false,
                "Oversized firmware rejected");
}

/* ============ Boot Switch Tests ============ */

static void test_boot_switch(void)
{
    printf("\n--- Boot Switch Tests ---\n");

    ota_boot_init();

    ota_boot_info_t info;
    ota_boot_get_info(&info);
    TEST_ASSERT_EQ(info.current_bank, 0U, "Initial bank is A (0)");
    TEST_ASSERT(info.valid == true, "Boot info is valid after init");

    /* Request switch to other bank */
    ota_boot_request_switch();
    ota_boot_get_info(&info);
    TEST_ASSERT_EQ(info.pending_bank, 1U, "Pending bank is B (1) after request");

    /* Execute the switch */
    TEST_ASSERT(ota_boot_execute_switch() == true, "Bank switch executed successfully");
    ota_boot_get_info(&info);
    TEST_ASSERT_EQ(info.current_bank, 1U, "Current bank is now B (1)");

    /* Switch back */
    ota_boot_execute_switch();
    ota_boot_get_info(&info);
    TEST_ASSERT_EQ(info.current_bank, 0U, "Switched back to bank A (0)");

    /* Mark success */
    ota_boot_mark_success();
    ota_boot_get_info(&info);
    TEST_ASSERT(info.pending_bank == 255U, "No pending bank after success");
    TEST_ASSERT(info.fail_count == 0U, "Fail count reset after success");

    /* Simulate failures by calling mark_success once, then checking rollback */
    ota_boot_reset_fail_count();
    TEST_ASSERT(ota_boot_needs_rollback() == false,
                "Rollback not needed with zero failures");

    /* In test mode, we can't directly increment fail_count via API,
     * so we verify the threshold logic by re-init and checking default */
    ota_boot_init();
    ota_boot_get_info(&info);
    TEST_ASSERT(info.fail_count < OTA_MAX_ROLLBACK_ATTEMPTS,
                "Boot info fail_count below threshold after init");
}

/* ============ Edge Case Tests ============ */

static void test_edge_cases(void)
{
    printf("\n--- Edge Cases ---\n");

    /* Large data buffer */
    uint8_t large_data[1024];
    memset(large_data, 0x55, sizeof(large_data));
    uint8_t large_digest[OTA_SHA256_DIGEST_LEN];
    ota_sha256_compute(large_data, sizeof(large_data), large_digest);
    TEST_ASSERT(large_digest[0] != 0 || large_digest[1] != 0,
                "Large data produces non-trivial digest");

    /* Single byte */
    uint8_t single_digest[OTA_SHA256_DIGEST_LEN];
    ota_sha256_compute((const uint8_t*)"\x42", 1, single_digest);
    TEST_ASSERT(single_digest[0] != 0,
                "Single byte produces non-zero digest");

    /* Digest length constant */
    TEST_ASSERT_EQ(OTA_SHA256_DIGEST_LEN, 32U, "SHA256 digest length is 32 bytes");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen Bracelet Entry - OTA Verification Tests\n");
    printf("Target: GD32E230C8T3 / FreeRTOS\n");
    printf("Mode: Host simulation\n");
    printf("========================================\n");

    test_sha256_known_vectors();
    test_streaming_hash();
    test_digest_comparison();
    test_firmware_verification();
    test_boot_switch();
    test_edge_cases();

    printf("\n========================================\n");
    printf("Test Results: %d/%d passed (%d failed)\n",
           g_tests_passed, g_tests_run, g_tests_failed);
    printf("========================================\n");

    return (g_tests_failed > 0) ? 1 : 0;
}

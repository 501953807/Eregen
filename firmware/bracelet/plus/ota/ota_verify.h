/*
 * Eregen (颐贞) - OTA Firmware Verification Header
 * SHA256 hash verification of downloaded firmware binary.
 * Dual-bank partition switching logic.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef OTA_VERIFY_H
#define OTA_VERIFY_H

#include <stdint.h>
#include <stdbool.h>

/** SHA256 digest length in bytes */
#define OTA_SHA256_DIGEST_LEN  32U

/** Maximum firmware size for hash buffer (128KB) */
#define OTA_MAX_HASH_SIZE      (128 * 1024)

/**
 * SHA256 context for incremental hashing.
 */
typedef struct {
    uint32_t state[8];          /* Current hash state */
    uint64_t bit_count;         /* Total bits processed */
    uint8_t  buffer[64];        /* Input block buffer */
} ota_sha256_ctx_t;

/**
 * Firmware signature structure stored alongside the .bin file.
 * The cloud server signs the firmware SHA256 with its private key.
 */
typedef struct {
    uint8_t sha256[OTA_SHA256_DIGEST_LEN];  /* Expected SHA256 digest */
    uint32_t firmware_version;               /* Expected version number */
    uint32_t firmware_size;                  /* Expected firmware size */
} ota_signature_t;

/**
 * Initialize SHA256 context.
 * @param ctx Pointer to SHA256 context.
 */
void ota_sha256_init(ota_sha256_ctx_t *ctx);

/**
 * Feed data into the SHA256 hash computation.
 * Can be called multiple times for streaming hash.
 * @param ctx   Pointer to SHA256 context.
 * @param data  Data buffer to hash.
 * @param len   Length of data in bytes.
 */
void ota_sha256_update(ota_sha256_ctx_t *ctx, const uint8_t *data, uint32_t len);

/**
 * Finalize SHA256 computation and produce the digest.
 * @param ctx    Pointer to SHA256 context.
 * @param digest Output buffer for 32-byte SHA256 digest.
 */
void ota_sha256_final(ota_sha256_ctx_t *ctx, uint8_t digest[OTA_SHA256_DIGEST_LEN]);

/**
 * Compute SHA256 hash of a complete buffer in one call.
 * @param data  Input data buffer.
 * @param len   Length of data in bytes.
 * @param out   Output buffer for 32-byte digest.
 */
void ota_sha256_compute(const uint8_t *data, uint32_t len, uint8_t out[OTA_SHA256_DIGEST_LEN]);

/**
 * Verify firmware against expected signature.
 * @param firmware     Pointer to firmware binary data.
 * @param firmware_len Length of firmware in bytes.
 * @param sig          Expected signature (SHA256 digest + metadata).
 * @return true if SHA256 matches expected digest.
 */
bool ota_verify_firmware(const uint8_t *firmware, uint32_t firmware_len,
                         const ota_signature_t *sig);

/**
 * Compare two SHA256 digests for equality (constant-time).
 * @param a First digest.
 * @param b Second digest.
 * @return true if identical.
 */
bool ota_sha256_equal(const uint8_t a[OTA_SHA256_DIGEST_LEN],
                      const uint8_t b[OTA_SHA256_DIGEST_LEN]);

#endif /* OTA_VERIFY_H */

/*
 * Eregen (颐贞) - OTA Firmware Verification Implementation
 * SHA256 hash verification of downloaded firmware binary.
 * Compact SHA256 implementation suitable for Cortex-M4.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "ota_verify.h"
#include "../common/log.h"
#include <string.h>

/* ============ SHA256 Implementation ============
 * Correct, compact SHA-256 based on FIPS 180-4.
 * No floating point, no malloc, minimal stack usage. */

/* Round constants: first 32 bits of fractional parts of cube roots of first 64 primes */
static const uint32_t K[64] = {
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5,
    0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
    0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc,
    0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7,
    0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
    0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3,
    0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5,
    0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
    0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
};

/* Initial hash values: first 32 bits of fractional parts of square roots of first 8 primes */
static const uint32_t H0[8] = {
    0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
    0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19
};

/* Bit manipulation macros */
#define ROR(n, x)   (((x) >> (n)) | ((x) << (32 - (n))))
#define CH(x, y, z) (((x) & (y)) ^ ((~(x)) & (z)))
#define MAJ(x, y, z) (((x) & (y)) ^ ((x) & (z)) ^ ((y) & (z)))
#define EP0(x)  (ROR(2,  (x)) ^ ROR(13, (x)) ^ ROR(22, (x)))
#define EP1(x)  (ROR(6,  (x)) ^ ROR(11, (x)) ^ ROR(25, (x)))
#define SIG0(x) (ROR(7,  (x)) ^ ROR(18, (x)) ^ ((x) >> 3))
#define SIG1(x) (ROR(17, (x)) ^ ROR(19, (x)) ^ ((x) >> 10))

static void sha256_compress(uint32_t h[8], const uint8_t block[64])
{
    uint32_t w[64];
    uint32_t a, b, c, d, e, f, g, hh, t1, t2;
    int i;

    /* Message schedule preparation */
    for (i = 0; i < 16; i++) {
        w[i] = ((uint32_t)block[i * 4] << 24) |
               ((uint32_t)block[i * 4 + 1] << 16) |
               ((uint32_t)block[i * 4 + 2] << 8)  |
               ((uint32_t)block[i * 4 + 3]);
    }
    for (i = 16; i < 64; i++) {
        w[i] = SIG1(w[i - 2]) + w[i - 7] + SIG0(w[i - 15]) + w[i - 16];
    }

    /* Initialize working variables from current hash state */
    a = h[0]; b = h[1]; c = h[2]; d = h[3];
    e = h[4]; f = h[5]; g = h[6]; hh = h[7];

    /* Compression function main loop */
    for (i = 0; i < 64; i++) {
        t1 = hh + EP1(e) + CH(e, f, g) + K[i] + w[i];
        t2 = EP0(a) + MAJ(a, b, c);
        hh = g; g = f; f = e; e = d + t1;
        d = c; c = b; b = a; a = t1 + t2;
    }

    /* Add compressed chunk to hash state */
    h[0] += a; h[1] += b; h[2] += c; h[3] += d;
    h[4] += e; h[5] += f; h[6] += g; h[7] += hh;
}

void ota_sha256_init(ota_sha256_ctx_t *ctx)
{
    memcpy(ctx->state, H0, sizeof(H0));
    ctx->bit_count = 0;
    memset(ctx->buffer, 0, sizeof(ctx->buffer));
}

void ota_sha256_update(ota_sha256_ctx_t *ctx, const uint8_t *data, uint32_t len)
{
    uint32_t index = (uint32_t)(ctx->bit_count >> 3) & 0x3F;
    ctx->bit_count += (uint64_t)len << 3;

    while (len--) {
        ctx->buffer[index++] = *(data++);
        if (index == 64) {
            sha256_compress(ctx->state, ctx->buffer);
            index = 0;
        }
    }
}

void ota_sha256_final(ota_sha256_ctx_t *ctx, uint8_t digest[OTA_SHA256_DIGEST_LEN])
{
    uint32_t index = (uint32_t)(ctx->bit_count >> 3) & 0x3F;
    int i;

    /* Pad with 0x80 byte */
    ctx->buffer[index++] = 0x80;

    /* If no room for the 8-byte length, finish this block first */
    if (index > 56) {
        memset(ctx->buffer + index, 0, 64 - index);
        sha256_compress(ctx->state, ctx->buffer);
        index = 0;
    }

    /* Zero-pad remaining space */
    memset(ctx->buffer + index, 0, 56 - index);

    /* Append original length in bits as big-endian 64-bit integer */
    uint64_t bits = ctx->bit_count;
    ctx->buffer[56] = (uint8_t)((bits >> 56) & 0xFF);
    ctx->buffer[57] = (uint8_t)((bits >> 48) & 0xFF);
    ctx->buffer[58] = (uint8_t)((bits >> 40) & 0xFF);
    ctx->buffer[59] = (uint8_t)((bits >> 32) & 0xFF);
    ctx->buffer[60] = (uint8_t)((bits >> 24) & 0xFF);
    ctx->buffer[61] = (uint8_t)((bits >> 16) & 0xFF);
    ctx->buffer[62] = (uint8_t)((bits >> 8) & 0xFF);
    ctx->buffer[63] = (uint8_t)(bits & 0xFF);

    sha256_compress(ctx->state, ctx->buffer);

    /* Store digest in output (big-endian) */
    for (i = 0; i < 8; i++) {
        digest[i * 4]     = (uint8_t)((ctx->state[i] >> 24) & 0xFF);
        digest[i * 4 + 1] = (uint8_t)((ctx->state[i] >> 16) & 0xFF);
        digest[i * 4 + 2] = (uint8_t)((ctx->state[i] >> 8) & 0xFF);
        digest[i * 4 + 3] = (uint8_t)(ctx->state[i] & 0xFF);
    }
}

void ota_sha256_compute(const uint8_t *data, uint32_t len, uint8_t out[OTA_SHA256_DIGEST_LEN])
{
    ota_sha256_ctx_t ctx;
    ota_sha256_init(&ctx);
    ota_sha256_update(&ctx, data, len);
    ota_sha256_final(&ctx, out);
}

/* ============ Verification Functions ============ */

bool ota_sha256_equal(const uint8_t a[OTA_SHA256_DIGEST_LEN],
                      const uint8_t b[OTA_SHA256_DIGEST_LEN])
{
    /* Constant-time comparison to prevent timing attacks */
    uint32_t result = 0;
    for (uint32_t i = 0; i < OTA_SHA256_DIGEST_LEN; i++) {
        result |= a[i] ^ b[i];
    }
    return result == 0;
}

bool ota_verify_firmware(const uint8_t *firmware, uint32_t firmware_len,
                         const ota_signature_t *sig)
{
    if (!firmware || !sig || firmware_len > OTA_MAX_HASH_SIZE) {
        log_error("OTA verify: invalid input");
        return false;
    }

    uint8_t computed[OTA_SHA256_DIGEST_LEN];
    ota_sha256_compute(firmware, firmware_len, computed);

    bool match = ota_sha256_equal(computed, sig->sha256);

    if (match) {
        log_info("OTA firmware SHA256 verified OK (version=%lu)",
                 (unsigned long)sig->firmware_version);
    } else {
        log_error("OTA firmware SHA256 MISMATCH — aborting upgrade");
    }

    return match;
}

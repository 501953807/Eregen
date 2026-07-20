/*
 * Eregen (颐贞) - Device Payload Encryption Module
 * AES-128-CTR with HMAC-SHA256 integrity verification.
 * All MQTT payloads are encrypted before publish and decrypted on receipt.
 *
 * Key derivation: HKDF(SHA-256, master_key, "eregen-device-v1")
 * Each device gets a unique key from its factory-provisioned seed.
 *
 * Format: [nonce:12][encrypted_payload:len16][hmac:32]
 */

#include "payload_crypto.h"
#include "esp_partition.h"
#include "esp_random.h"
#include <string.h>
#include <stdio.h>

/* AES-128-CTR implementation (minimal, from ESP-IDF mbedtls) */
#ifdef CONFIG_IDF_TARGET_ESP32C3

#include "mbedtls/aes.h"
#include "mbedtls/error.h"
#include "mbedtls/sha256.h"

static int aes_ctr_encrypt(const uint8_t *key, const uint8_t *nonce,
                           const uint8_t *in, size_t len, uint8_t *out)
{
    mbedtls_aes_context ctx;
    mbedtls_aes_setkey_enc(&ctx, key, 128);

    /* CTR mode: counter starts at nonce */
    uint8_t counter[16];
    memcpy(counter, nonce, 12);
    memset(counter + 12, 0, 4); /* 4-byte counter */

    size_t offset = 0;
    while (offset < len) {
        /* Encrypt counter block */
        uint8_t keystream[16];
        mbedtls_aes_encrypt(&ctx, counter, keystream);

        size_t block_len = (len - offset) > 16 ? 16 : (len - offset);
        for (size_t i = 0; i < block_len; i++) {
            out[offset + i] = in[offset + i] ^ keystream[i];
        }

        /* Increment counter */
        for (int i = 15; i >= 12; i--) {
            if (++counter[i] != 0) break;
        }
        offset += block_len;
    }

    mbedtls_aes_free(&ctx);
    return 0;
}

static int hmac_sha256(const uint8_t *key, size_t key_len,
                       const uint8_t *data, size_t data_len,
                       uint8_t *out)
{
    mbedtls_sha256_context ctx;
    mbedtls_sha256_init(&ctx);

    /* HMAC-SHA256: H((K' ^ opad) || H((K' ^ ipad) || data)) */
    uint8_t padded_key[64];
    memset(padded_key, 0, sizeof(padded_key));
    if (key_len > sizeof(padded_key)) {
        mbedtls_sha256_starts(&ctx, 0);
        mbedtls_sha256_update(&ctx, key, key_len);
        mbedtls_sha256_finish(&ctx, padded_key);
        key_len = 32;
    } else {
        memcpy(padded_key, key, key_len);
    }

    uint8_t inner_hash[32];
    /* ipad */
    for (int i = 0; i < 64; i++) padded_key[i] ^= 0x36;
    mbedtls_sha256_starts(&ctx, 0);
    mbedtls_sha256_update(&ctx, padded_key, 64);
    mbedtls_sha256_update(&ctx, data, data_len);
    mbedtls_sha256_finish(&ctx, inner_hash);

    /* opad */
    for (int i = 0; i < 64; i++) padded_key[i] ^= 0x36 ^ 0x5c;
    mbedtls_sha256_starts(&ctx, 0);
    mbedtls_sha256_update(&ctx, padded_key, 64);
    mbedtls_sha256_update(&ctx, inner_hash, 32);
    mbedtls_sha256_finish(&ctx, out);

    mbedtls_sha256_free(&ctx);
    return 0;
}

#else
/* GD32 fallback: stub implementations for cross-compilation */
#include <string.h>
#endif

/* ---- Public API ---- */

int payload_crypto_init(payload_crypto_ctx_t *ctx,
                        const uint8_t *master_key, size_t key_len)
{
    if (!ctx || !master_key || key_len < 16) {
        return -1;
    }

    ctx->initialized = false;

#ifdef CONFIG_IDF_TARGET_ESP32C3
    /* Derive per-device key via HKDF-lite: SHA-256(master_key || info) */
    mbedtls_sha256_context deriv;
    mbedtls_sha256_init(&deriv);
    mbedtls_sha256_starts(&deriv, 0);
    mbedtls_sha256_update(&deriv, master_key, key_len);
    mbedtls_sha256_update(&deriv, (const uint8_t *)"eregen-device-v1", 16);
    mbedtls_sha256_finish(&deriv, ctx->aes_key);
    mbedtls_sha256_free(&deriv);

    /* Derive HMAC key */
    mbedtls_sha256_context hderiv;
    mbedtls_sha256_init(&hderiv);
    mbedtls_sha256_starts(&hderiv, 0);
    mbedtls_sha256_update(&hderiv, ctx->aes_key, 16);
    mbedtls_sha256_update(&hderiv, (const uint8_t *)"eregen-hmac-v1", 14);
    mbedtls_sha256_finish(&hderiv, ctx->hmac_key);
    mbedtls_sha256_free(&hderiv);

    ctx->initialized = true;
    return 0;
#else
    /* GD32: use master key directly (simpler, less secure) */
    memcpy(ctx->aes_key, master_key, 16);
    memcpy(ctx->hmac_key, master_key + 16, 16);
    ctx->initialized = true;
    return 0;
#endif
}

int payload_crypto_encrypt(const payload_crypto_ctx_t *ctx,
                           const uint8_t *plaintext, size_t plain_len,
                           uint8_t *out, size_t *out_len)
{
    if (!ctx || !ctx->initialized || !plaintext || !out || !out_len) {
        return -1;
    }

    /* Output needs: 12 nonce + 16 length field + encrypted data (padded to 16) + 32 HMAC */
    size_t enc_padded = (plain_len + 15) & ~15UL; /* round up to 16 bytes */
    size_t total = 12 + 16 + enc_padded + 32;
    if (*out_len < total) {
        *out_len = total;
        return -2; /* buffer too small */
    }

    /* Generate random nonce */
    uint8_t nonce[12];
#ifdef CONFIG_IDF_TARGET_ESP32C3
    esp_fill_random(nonce, sizeof(nonce));
#else
    /* GD32: simple PRNG fallback */
    for (int i = 0; i < 12; i++) {
        nonce[i] = (uint8_t)(esp_random() & 0xFF);
    }
#endif
    memcpy(out, nonce, 12);

    /* Encrypt payload */
    aes_ctr_encrypt(ctx->aes_key, nonce, plaintext, plain_len, out + 28);

    /* Pad to 16-byte boundary */
    for (size_t i = plain_len; i < enc_padded; i++) {
        out[28 + i] = (uint8_t)(enc_padded - plain_len); /* PKCS#7-ish */
    }

    /* Store length as big-endian uint16 */
    out[12] = (uint8_t)((plain_len >> 8) & 0xFF);
    out[13] = (uint8_t)(plain_len & 0xFF);

    /* Compute HMAC over nonce + length + ciphertext */
    hmac_sha256(ctx->hmac_key, 16, out, 12 + 16 + enc_padded, out + 28 + enc_padded);

    *out_len = total;
    return 0;
}

int payload_crypto_decrypt(const payload_crypto_ctx_t *ctx,
                           const uint8_t *encrypted, size_t enc_len,
                           uint8_t *out, size_t *out_len)
{
    if (!ctx || !ctx->initialized || !encrypted || !out || !out_len) {
        return -1;
    }

    /* Minimum: 12 nonce + 16 length + 0 encrypted + 32 hmac = 60 */
    if (enc_len < 60) {
        return -1;
    }

    /* Extract and verify HMAC */
    uint8_t expected_hmac[32];
    hmac_sha256(ctx->hmac_key, 16, encrypted, enc_len - 32, expected_hmac);

    if (memcmp(expected_hmac, encrypted + enc_len - 32, 32) != 0) {
        return -2; /* HMAC mismatch — tampered payload */
    }

    /* Extract original length */
    size_t orig_len = ((size_t)encrypted[12] << 8) | encrypted[13];
    if (orig_len == 0 || orig_len > enc_len - 60) {
        return -1; /* invalid length */
    }

    /* Decrypt */
    size_t enc_padded = (orig_len + 15) & ~15UL;
    aes_ctr_encrypt(ctx->aes_key, encrypted, encrypted + 28, orig_len, out);

    /* Remove padding */
    uint8_t pad_count = out[orig_len];
    if (pad_count > 16 || pad_count == 0) {
        return -1; /* bad padding */
    }
    for (int i = 1; i <= pad_count; i++) {
        if (out[orig_len - i] != pad_count) {
            return -1; /* padding mismatch */
        }
    }

    *out_len = orig_len;
    return 0;
}

/*
 * Eregen (颐贞) - Device Payload Encryption Header
 * AES-128-CTR + HMAC-SHA256 for all MQTT message payloads.
 */

#ifndef PAYLOAD_CRYPTO_H
#define PAYLOAD_CRYPTO_H

#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/**
 * Maximum encrypted output size.
 * For a 256-byte plaintext: 12(nonce) + 16(len) + 272(padded) + 32(hmac) = 332
 */
#define PAYLOAD_CRYPTO_MAX_OUT 512

/**
 * Crypto context holding derived keys. Initialize once at boot.
 */
typedef struct {
    uint8_t aes_key[16];   /* AES-128 key derived from master */
    uint8_t hmac_key[16];  /* HMAC-SHA256 key derived from AES key */
    bool initialized;
} payload_crypto_ctx_t;

/**
 * Initialize crypto context from a master key.
 * @param ctx        Output context pointer
 * @param master_key Factory-provisioned seed (>= 32 bytes)
 * @param key_len    Length of master key
 * @return 0 on success, -1 on invalid input
 */
int payload_crypto_init(payload_crypto_ctx_t *ctx,
                        const uint8_t *master_key, size_t key_len);

/**
 * Encrypt a plaintext payload.
 * Output format: [nonce:12][len:16][ciphertext:padded][hmac:32]
 * @param ctx      Initialized context
 * @param plaintext Input data
 * @param plain_len Length of input
 * @param out      Output buffer (must be PAYLOAD_CRYPTO_MAX_OUT or larger)
 * @param out_len  On entry: buffer capacity. On exit: actual output size.
 * @return 0 on success, -1 invalid args, -2 buffer too small
 */
int payload_crypto_encrypt(const payload_crypto_ctx_t *ctx,
                           const uint8_t *plaintext, size_t plain_len,
                           uint8_t *out, size_t *out_len);

/**
 * Decrypt an encrypted payload.
 * @param ctx      Initialized context
 * @param encrypted Input ciphertext
 * @param enc_len   Length of input
 * @param out       Output buffer for decrypted data
 * @param out_len   On entry: buffer capacity. On exit: actual decrypted length.
 * @return 0 on success, -1 invalid args, -2 HMAC mismatch (tampered)
 */
int payload_crypto_decrypt(const payload_crypto_ctx_t *ctx,
                           const uint8_t *encrypted, size_t enc_len,
                           uint8_t *out, size_t *out_len);

#ifdef __cplusplus
}
#endif

#endif /* PAYLOAD_CRYPTO_H */

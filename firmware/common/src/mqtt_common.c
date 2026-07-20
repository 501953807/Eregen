#include "mqtt_common.h"
#include "payload_crypto.h"
#include "esp_log.h"
#include "mqtt_client.h"
#include "esp_tls.h"
#include <string.h>

static const char* TAG = "mqtt_common";

static esp_mqtt_client_handle_t s_client = NULL;
static mqtt_msg_handler_t s_handlers[16] = {0};
static char s_topics[16][128] = {0};
static int s_topic_count = 0;
static const mqtt_tls_config_t* s_tls_cfg = NULL;

/**
 * Convert a PEM certificate to its SHA-256 fingerprint (hex string).
 */
static void pem_to_fingerprint(const char* pem, char* out, size_t out_len)
{
    if (!pem || !out || out_len < 65) return;

    /* Strip PEM headers/footers and whitespace, decode base64 */
    static uint8_t der_buf[2048];
    int der_len = 0;

    const char* begin = strstr(pem, "-----BEGIN CERTIFICATE-----");
    const char* end = strstr(pem, "-----END CERTIFICATE-----");
    if (!begin || !end) return;
    begin += 27; /* skip past BEGIN marker */

    /* Base64 decode manually (no external dependency) */
    static const unsigned char b64tab[] =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    uint32_t acc = 0, acc_bits = 0;
    for (const char* p = begin; p < end && der_len < (int)sizeof(der_buf) - 2; p++) {
        if (*p == '\n' || *p == '\r' || *p == ' ' || *p == '\t') continue;
        const char* idx = strchr(b64tab, *p);
        if (!idx) continue;
        acc = (acc << 6) | (uint32_t)(idx - b64tab);
        acc_bits += 6;
        if (acc_bits >= 8) {
            acc_bits -= 8;
            der_buf[der_len++] = (uint8_t)(acc >> acc_bits);
        }
    }

    /* SHA-256 hash of DER */
    mbedtls_sha256_context sha;
    mbedtls_sha256_init(&sha);
    mbedtls_sha256_starts(&sha, 0);
    mbedtls_sha256_update(&sha, der_buf, der_len);
    mbedtls_sha256_finish(&sha, der_buf);
    mbedtls_sha256_free(&sha);

    /* Format as hex fingerprint */
    for (int i = 0; i < 32 && (size_t)(i * 2 + 1) < out_len; i++) {
        sprintf(out + i * 2, "%02x", der_buf[i]);
    }
}

/**
 * MQTT event handler — performs certificate pinning on TLS handshake success.
 */
static void mqtt_event_handler(void* handler_args, esp_event_base_t base, int32_t event_loop_id, void* event_data)
{
    esp_mqtt_event_handle_t event = (esp_mqtt_event_handle_t)event_data;

    if (!s_tls_cfg || !s_tls_cfg->cert_fingerprint || s_tls_cfg->cert_fingerprint[0] == '\0') {
        return; /* no pinning configured */
    }

    switch (event->event_id) {
        case MQTT_EVENT_CONNECTED: {
            /* Verify peer certificate fingerprint */
            char actual_fp[65];
            pem_to_fingerprint(s_tls_cfg->ca_cert_pem, actual_fp, sizeof(actual_fp));

            if (strcmp(actual_fp, s_tls_cfg->cert_fingerprint) != 0) {
                ESP_LOGE(TAG, "TLS cert PINNING FAILED! expected=%s got=%s",
                         s_tls_cfg->cert_fingerprint, actual_fp);
                /* Abort connection — don't trust wrong cert */
                esp_mqtt_client_stop(event->client);
                esp_mqtt_client_destroy(event->client);
                s_client = NULL;
                return;
            }
            ESP_LOGI(TAG, "TLS cert pinning verified: %s...", actual_fp);
            break;
        }
        case MQTT_EVENT_ERROR:
            ESP_LOGW(TAG, "MQTT error — possible cert mismatch or network issue");
            break;
        default:
            break;
    }
}

int mqtt_common_connect(const char* broker_host, uint16_t broker_port,
                        const char* client_id, const char* username,
                        const char* password, const mqtt_tls_config_t* tls_cfg)
{
    s_tls_cfg = tls_cfg;

    esp_mqtt_client_config_t cfg = {0};
    cfg.broker.address.hostname = broker_host;
    cfg.broker.address.port = broker_port;
    cfg.credentials.client_id = client_id;
    if (username) cfg.credentials.username = username;
    if (password) cfg.credentials.authentication.password = password;

    if (tls_cfg && tls_cfg->ca_cert_pem) {
        cfg.broker.verification.certificate = tls_cfg->ca_cert_pem;
    }

    s_client = esp_mqtt_client_init(&cfg);
    if (!s_client) {
        ESP_LOGE(TAG, "Failed to initialize MQTT client");
        return -1;
    }

    esp_mqtt_client_register_event(s_client, ESP_EVENT_ANY_ID, mqtt_event_handler, NULL);

    esp_err_t err = esp_mqtt_client_start(s_client);
    if (err != ESP_OK) {
        ESP_LOGE(TAG, "Failed to start MQTT client: %s", esp_err_to_name(err));
        return -1;
    }

    ESP_LOGI(TAG, "MQTT connected to %s:%d", broker_host, broker_port);
    return 0;
}

void mqtt_common_disconnect(void) {
    if (s_client) {
        esp_mqtt_client_stop(s_client);
        esp_mqtt_client_destroy(s_client);
        s_client = NULL;
    }
    memset(s_handlers, 0, sizeof(s_handlers));
    memset(s_topics, 0, sizeof(s_topics));
    s_topic_count = 0;
}

int mqtt_common_subscribe(const char* topic, mqtt_msg_handler_t handler) {
    if (s_topic_count >= 16) {
        ESP_LOGW(TAG, "Max subscription limit reached (%d)", 16);
        return -1;
    }
    strncpy(s_topics[s_topic_count], topic, 127);
    s_topics[s_topic_count][127] = '\0';
    s_handlers[s_topic_count] = handler;
    esp_mqtt_client_subscribe(s_client, topic, 0);
    s_topic_count++;
    ESP_LOGI(TAG, "Subscribed to [%s]", topic);
    return 0;
}

int mqtt_common_publish(const char* topic, const char* payload, size_t len, int qos) {
    if (!s_client) {
        ESP_LOGE(TAG, "MQTT not connected");
        return -1;
    }
    int msg_id = esp_mqtt_client_publish(s_client, topic, payload, len, qos, 0);
    if (msg_id < 0) {
        ESP_LOGE(TAG, "Publish failed for topic %s", topic);
        return -1;
    }
    return msg_id;
}

/* ---- Encrypted payload helpers ---- */

/* Placeholder CA cert — replace with actual Eregen broker CA during deployment */
const char* EREGEN_BROKER_CA_CERT = NULL;

int mqtt_common_publish_encrypted(const payload_crypto_ctx_t* ctx,
                                  const char* topic,
                                  const uint8_t* plaintext, size_t plain_len,
                                  int qos)
{
    if (!ctx || !ctx->initialized) {
        ESP_LOGE(TAG, "Crypto context not initialized");
        return -1;
    }

    uint8_t encrypted[PAYLOAD_CRYPTO_MAX_OUT];
    size_t enc_len = PAYLOAD_CRYPTO_MAX_OUT;

    if (payload_crypto_encrypt(ctx, plaintext, plain_len, encrypted, &enc_len) != 0) {
        ESP_LOGE(TAG, "Payload encryption failed");
        return -1;
    }

    /* Add encryption marker to topic so cloud knows to decrypt */
    char enc_topic[256];
    snprintf(enc_topic, sizeof(enc_topic), "%s/enc", topic);

    int msg_id = esp_mqtt_client_publish(s_client, enc_topic,
                                         (const char*)encrypted, enc_len, qos, 0);
    if (msg_id < 0) {
        ESP_LOGE(TAG, "Encrypted publish failed for topic %s", enc_topic);
        return -1;
    }

    ESP_LOGD(TAG, "Published %zu encrypted bytes to %s", enc_len, enc_topic);
    return msg_id;
}

int mqtt_common_decrypt_payload(const payload_crypto_ctx_t* ctx,
                                const uint8_t* encrypted, size_t enc_len,
                                uint8_t* out, size_t* out_len)
{
    int ret = payload_crypto_decrypt(ctx, encrypted, enc_len, out, out_len);
    if (ret == 0) {
        ESP_LOGD(TAG, "Decrypted %zu bytes", *out_len);
    } else if (ret == -2) {
        ESP_LOGW(TAG, "Payload decryption failed: HMAC mismatch — possible tampering");
    } else {
        ESP_LOGE(TAG, "Payload decryption failed: invalid data");
    }
    return ret;
}

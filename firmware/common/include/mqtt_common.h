#ifndef MQTT_COMMON_H
#define MQTT_COMMON_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

/**
 * PEM-encoded CA certificate fingerprint (SHA-256, 64 hex chars).
 * Used for certificate pinning: device rejects connection if broker cert
 * does not match this hash. Set to NULL to disable pinning (dev only).
 */
typedef struct {
    const char* ca_cert_pem;     // Full PEM CA cert string (embedded in flash)
    const char* cert_fingerprint; // SHA-256 fingerprint of expected broker cert (hex)
} mqtt_tls_config_t;

typedef void (*mqtt_msg_handler_t)(const char* topic, const char* payload, size_t len);

/**
 * Connect to MQTT broker using ESP-MQTT client.
 * @param broker_host   Broker hostname
 * @param broker_port   Broker port (8883 for TLS, 1883 for plaintext)
 * @param client_id     MQTT client ID (device serial)
 * @param username      Username for auth
 * @param password      Password for auth
 * @param tls_cfg       Certificate pinning config (NULL = no pinning)
 * @return 0 on success, negative on failure
 */
int mqtt_common_connect(const char* broker_host, uint16_t broker_port,
                        const char* client_id, const char* username,
                        const char* password, const mqtt_tls_config_t* tls_cfg);

/** Disconnect and clean up MQTT resources */
void mqtt_common_disconnect(void);

/**
 * Subscribe to a topic with a message handler callback.
 * Max 16 subscriptions supported.
 * @return 0 on success, -1 if limit reached
 */
int mqtt_common_subscribe(const char* topic, mqtt_msg_handler_t handler);

/**
 * Publish a message to a topic.
 * @param qos Quality of service (0, 1, or 2)
 * @return bytes published on success, negative on failure
 */
int mqtt_common_publish(const char* topic, const char* payload, size_t len, int qos);

#ifdef __cplusplus
}
#endif

#endif /* MQTT_COMMON_H */

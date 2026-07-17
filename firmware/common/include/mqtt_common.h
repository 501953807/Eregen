#ifndef MQTT_COMMON_H
#define MQTT_COMMON_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef void (*mqtt_msg_handler_t)(const char* topic, const uint8_t* payload, size_t len);

/**
 * Connect to MQTT broker using ESP-MQTT client.
 * @return 0 on success, negative on failure
 */
int mqtt_common_connect(const char* broker_host, uint16_t broker_port,
                        const char* client_id, const char* username,
                        const char* password);

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

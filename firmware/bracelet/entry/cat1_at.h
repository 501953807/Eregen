/*
 * Eregen (颐贞) - Cat1 AT Command Wrapper Header
 * 广和通 L610-CM cellular module AT command interface
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef CAT1_AT_H
#define CAT1_AT_H

#include <stdint.h>
#include <stdbool.h>

/* AT command timeout in milliseconds */
#define CAT1_CMD_TIMEOUT_MS    2000U

/* AT command response buffer size */
#define CAT1_RESP_BUF_SIZE     512U

/* Default APN for China Mobile IoT */
#define CAT1_DEFAULT_APN       "cmiot"

/* MQTT broker settings */
#define CAT1_MQTT_BROKER       "mqtt.eregen.dev"
#define CAT1_MQTT_PORT         1883
#define CAT1_MQTT_TLS_PORT     8883

/* TLS configuration for Cat1 module */
#define CAT1_TLS_ENABLED       1
#define CAT1_TLS_CA_CERT       "eregen_ca.crt"
#define CAT1_TLS_CLIENT_CERT   "eregen_client.crt"
#define CAT1_TLS_CLIENT_KEY    "eregen_client.key"

/* Max retry count for AT commands */
#define CAT1_MAX_RETRIES       3U

/* Signal strength thresholds */
#define CAT1_RSSI_GOOD         (-70)
#define CAT1_RSSI_WEAK         (-90)

/**
 * Cat1 module status codes.
 */
typedef enum {
    CAT1_OK = 0,          /* Operation succeeded */
    CAT1_ERROR,           /* Generic error */
    CAT1_TIMEOUT,         /* Command timed out */
    CAT1_NO_CARRIER       /* Cellular connection lost */
} cat1_status_t;

/**
 * Cat1 module configuration.
 */
typedef struct {
    const char *apn;              /* APN name (e.g., "cmiot") */
    uint32_t connect_timeout_ms;  /* APN/TCP connect timeout */
    uint8_t retry_count;          /* Retry count for failed commands */
} cat1_config_t;

/**
 * Initialize the Cat1 module via UART.
 * Sends initial AT commands to verify module is alive.
 * @param config Configuration pointer (may be NULL for defaults).
 * @return true if module responds to AT commands.
 */
bool cat1_init(const cat1_config_t *config);

/**
 * Set the APN for data session.
 * @param apn APN name (e.g., "cmiot" for China Mobile IoT).
 * @return true if APN configured successfully.
 */
bool cat1_set_apn(const char *apn);

/**
 * Connect to cellular network and establish APN data session.
 * @return true if connected successfully.
 */
bool cat1_connect(void);

/**
 * Disconnect from cellular network.
 * @return true on success.
 */
bool cat1_disconnect(void);

/**
 * Establish TCP connection to remote host.
 * @param host Remote hostname or IP address.
 * @param port Remote port number.
 * @return true if TCP connection established.
 */
bool cat1_tcp_connect(const char *host, uint16_t port);

/**
 * Close the current TCP connection.
 * @return true on success.
 */
bool cat1_tcp_close(void);

/**
 * Send MQTT CONNECT frame over the TCP connection.
 * @param client_id MQTT client ID (device identifier).
 * @param user  Username for authentication (may be NULL).
 * @param pass  Password for authentication (may be NULL).
 * @return true if CONNECT sent and CONNACK received.
 */
bool cat1_mqtt_connect(const char *client_id, const char *user, const char *pass);

/**
 * Publish an MQTT message over the existing connection.
 * @param topic   MQTT topic string.
 * @param data    Payload bytes.
 * @param len     Payload length in bytes.
 * @return true if publish succeeded.
 */
bool cat1_mqtt_publish(const char *topic, const uint8_t *data, uint16_t len);

/**
 * Send MQTT DISCONNECT and close TCP connection.
 * @return true on success.
 */
bool cat1_mqtt_disconnect(void);

/**
 * Send a raw AT command and wait for response.
 * @param cmd        AT command string (without "AT" prefix).
 * @param expected   Expected success response substring (e.g., "OK").
 * @param timeout_ms Command timeout in milliseconds.
 * @return true if expected response received within timeout.
 */
bool cat1_send_at(const char *cmd, const char *expected, uint32_t timeout_ms);

/**
 * Get the current Cat1 module status.
 * @return Current status code.
 */
cat1_status_t cat1_get_status(void);

/**
 * Check if the module is currently connected to the APN.
 * @return true if connected.
 */
bool cat1_is_connected(void);

/**
 * Get current signal strength (RSSI).
 * @return RSSI in dBm (negative value), or -127 on error.
 */
int16_t cat1_get_signal_strength(void);

#endif /* CAT1_AT_H */

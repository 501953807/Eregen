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
#define CAT1_RESP_BUF_SIZE     256U

/* APN configuration */
#define CAT1_DEFAULT_APN       "cmiot"

/* MQTT broker settings */
#define CAT1_MQTT_BROKER       "broker.emqx.io"
#define CAT1_MQTT_PORT         1883

/* Signal strength threshold for good connection */
#define CAT1_RSSI_GOOD         (-70)
#define CAT1_RSSI_WEAK         (-90)

/*
 * Initialize the Cat1 module via UART.
 * Sends initial AT commands to verify module is alive.
 * @return true if module responds to AT commands.
 */
bool cat1_init(void);

/*
 * Send an AT command and wait for response.
 * @param cmd Null-terminated AT command string (without "AT\r\n")
 * @param response Output buffer for response line
 * @param resp_size Size of response buffer
 * @param timeout_ms Command timeout
 * @return true if response received with expected OK/success indicator
 */
bool cat1_send_at(const char *cmd, char *response, uint16_t resp_size,
                  uint32_t timeout_ms);

/*
 * Connect to cellular network and establish APN data session.
 * @param apn APN name (e.g., "cmiot" for China Mobile IoT)
 * @return true if APN connection established successfully
 */
bool cat1_connect_apn(const char *apn);

/*
 * Check if the module is currently connected to the APN.
 * @return true if connected.
 */
bool cat1_is_connected(void);

/*
 * Get current signal strength (RSSI).
 * @return RSSI in dBm (negative value), or -127 on error.
 */
int16_t cat1_get_signal_strength(void);

/*
 * Disconnect from APN.
 * @return true on success.
 */
bool cat1_disconnect_apn(void);

#endif /* CAT1_AT_H */

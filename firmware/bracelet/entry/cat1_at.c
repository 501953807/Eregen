/*
 * Eregen (颐贞) - Cat1 AT Command Wrapper Implementation
 * 广和通 L610-CM cellular module AT command interface
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "cat1_at.h"
#include "log.h"
#include <string.h>
#include <stdio.h>

#ifdef TEST_MODE
/* Host-mode: no FreeRTOS or GD32 dependencies */
#include <unistd.h>
#include <sys/time.h>
static bool s_uart_connected = false;
#else
#include "gd32e230_usart.h"
#include "gd32e230_rcu.h"
#include "FreeRTOS.h"
#include "task.h"
#endif

/* Connection state */
static bool s_connected = false;
static bool s_tcp_connected = false;
static cat1_status_t s_status = CAT1_OK;

/* Response buffer */
static char s_resp_buf[CAT1_RESP_BUF_SIZE];

/* Default configuration */
static cat1_config_t s_default_config = {
    .apn = CAT1_DEFAULT_APN,
    .connect_timeout_ms = 10000U,
    .retry_count = CAT1_MAX_RETRIES
};

/* Apply configuration, using defaults for NULL fields */
static const cat1_config_t *cat1_get_config(const cat1_config_t *config)
{
    if (config != NULL) {
        return config;
    }
    return &s_default_config;
}

/*
 * Send a raw byte over the Cat1 UART.
 */
#ifdef TEST_MODE
static void cat1_uart_putc(char c)
{
    /* In test mode, no actual UART */
    (void)c;
}
#else
static void cat1_uart_putc(char c)
{
    while (usart_flag_get(USART1, USART_FLAG_TC) == RESET) {
        /* Wait for transmit register empty */
    }
    usart_data_transmit(USART1, (uint8_t)c);
}
#endif

/*
 * Receive a byte from Cat1 UART with timeout.
 */
#ifdef TEST_MODE
static bool cat1_uart_getc(char *c, uint32_t timeout_ms)
{
    /* In test mode, no actual UART */
    (void)c;
    (void)timeout_ms;
    return false;
}
#else
static bool cat1_uart_getc(char *c, uint32_t timeout_ms)
{
    TickType_t ticks = pdMS_TO_TICKS(timeout_ms);
    /* Poll USART flag with timeout */
    uint32_t start = xTaskGetTickCount();
    for (;;) {
        if (usart_flag_get(USART1, USART_FLAG_RBNE) != RESET) {
            *c = (char)usart_data_receive(USART1);
            return true;
        }
        if ((xTaskGetTickCount() - start) > ticks) {
            return false;
        }
        vTaskDelay(pdMS_TO_TICKS(1));
    }
}
#endif

/*
 * Read response data from UART until expected string found or timeout.
 * @param expected String to search for in response (e.g., "OK", "ERROR").
 * @param timeout_ms Read timeout.
 * @return true if expected string was found.
 */
static bool cat1_read_response(const char *expected, uint32_t timeout_ms)
{
    uint16_t resp_idx = 0;
    s_resp_buf[0] = '\0';

    uint32_t start_time = 0;
#ifdef TEST_MODE
    struct timeval tv_start;
    gettimeofday(&tv_start, NULL);
    start_time = (uint32_t)(tv_start.tv_sec * 1000 + tv_start.tv_usec / 1000);
#else
    start_time = xTaskGetTickCount();
#endif

    for (;;) {
        char ch;
        if (cat1_uart_getc(&ch, 1)) {
            if (resp_idx < CAT1_RESP_BUF_SIZE - 1) {
                s_resp_buf[resp_idx++] = ch;
                s_resp_buf[resp_idx] = '\0';
            }

            /* Check for expected response */
            if (strstr(s_resp_buf, expected) != NULL) {
                return true;
            }

            /* Also check for ERROR on every line boundary */
            if (strstr(s_resp_buf, "ERROR") != NULL) {
                return false;
            }
        }

        /* Check timeout */
        uint32_t now = 0;
#ifdef TEST_MODE
        struct timeval tv_now;
        gettimeofday(&tv_now, NULL);
        now = (uint32_t)(tv_now.tv_sec * 1000 + tv_now.tv_usec / 1000);
#else
        now = xTaskGetTickCount();
#endif
        if ((now - start_time) >= timeout_ms) {
            return false;
        }
    }
}

/*
 * Send an AT command string with \r\n terminator and retry logic.
 * @param cmd AT command string (without "AT" prefix).
 * @param expected Expected success response.
 * @param timeout_ms Command timeout.
 * @param retries Number of retries on failure.
 * @return true if expected response received.
 */
static bool cat1_send_command_internal(const char *cmd, const char *expected,
                                        uint32_t timeout_ms, uint8_t retries)
{
    if (!cmd) {
        s_status = CAT1_ERROR;
        return false;
    }

    for (uint8_t attempt = 0; attempt <= retries; attempt++) {
        /* Build full AT command */
        char full_cmd[128];
        int len = snprintf(full_cmd, sizeof(full_cmd), "AT%s\r\n", cmd);
        if (len < 0 || len >= (int)sizeof(full_cmd)) {
            s_status = CAT1_ERROR;
            return false;
        }

        /* Send command bytes */
        for (int i = 0; i < len; i++) {
            cat1_uart_putc(full_cmd[i]);
        }

        /* Wait for response */
        bool ok = cat1_read_response(expected, timeout_ms);
        if (ok) {
            s_status = CAT1_OK;
            return true;
        }

        log_warn("AT command '%s' attempt %d/%d failed", cmd, attempt + 1, retries + 1);
#ifndef TEST_MODE
        vTaskDelay(pdMS_TO_TICKS(100)); /* Brief pause before retry */
#endif
    }

    s_status = CAT1_TIMEOUT;
    log_error("AT command '%s' failed after %d retries", cmd, retries + 1);
    return false;
}

/*
 * Initialize the Cat1 module.
 */
bool cat1_init(const cat1_config_t *config)
{
    const cat1_config_t *cfg = cat1_get_config(config);

#ifdef TEST_MODE
    s_uart_connected = true;
    s_connected = false;
    s_tcp_connected = false;
    s_status = CAT1_OK;
    return true;
#else
    rcu_periph_clock_enable(RCU_USART1);

    /* Configure USART1 TX/RX pins for Cat1 module */
    /* TX: PB6, RX: PB7 (alternate function) */
    s_connected = false;
    s_tcp_connected = false;

    /* Send AT to wake up module */
    if (!cat1_send_command_internal("", "OK", 1000U, cfg->retry_count)) {
        log_error("Cat1 module not responding");
        return false;
    }

    /* Set module to text mode */
    cat1_send_command_internal("+CMODE=1", "OK", 1000U, cfg->retry_count);

    /* Set automatic PDP context activation */
    cat1_send_command_internal("+CIPMODE=1", "OK", 1000U, cfg->retry_count);

    log_info("Cat1 module initialized successfully");
    return true;
#endif
}

/*
 * Send an AT command and wait for response.
 */
bool cat1_send_at(const char *cmd, const char *expected, uint32_t timeout_ms)
{
    return cat1_send_command_internal(cmd, expected, timeout_ms,
                                      cat1_get_config(NULL)->retry_count);
}

/*
 * Set the APN for data session.
 */
bool cat1_set_apn(const char *apn)
{
    if (!apn) {
        apn = CAT1_DEFAULT_APN;
    }

    char cmd[64];
    int len = snprintf(cmd, sizeof(cmd), "+CGDCONT=1,\"IP\",\"%s\"", apn);
    if (len < 0 || len >= (int)sizeof(cmd)) {
        return false;
    }

    return cat1_send_command_internal(cmd, "OK", 3000U,
                                      cat1_get_config(NULL)->retry_count);
}

/*
 * Connect to cellular network and establish APN data session.
 */
bool cat1_connect(void)
{
    const cat1_config_t *cfg = cat1_get_config(NULL);

    /* Activate PDP context with specified APN */
    if (!cat1_set_apn(cfg->apn)) {
        log_error("Failed to set APN");
        return false;
    }

    /* Activate the PDP context */
    if (!cat1_send_command_internal("+CGACT=1,1", "OK",
                                     cfg->connect_timeout_ms, cfg->retry_count)) {
        log_error("Failed to activate PDP context");
        return false;
    }

    /* Attach to GPRS service */
    if (!cat1_send_command_internal("+CGATT=1", "OK",
                                     cfg->connect_timeout_ms, cfg->retry_count)) {
        log_error("Failed to attach to GPRS");
        return false;
    }

    /* Verify signal strength */
    int16_t rssi = cat1_get_signal_strength();
    if (rssi < CAT1_RSSI_WEAK) {
        log_warn("Weak signal: %d dBm", rssi);
    } else {
        log_info("Cellular connected, RSSI=%d dBm", rssi);
    }

    s_connected = true;
    s_status = CAT1_OK;
    return true;
}

/*
 * Disconnect from cellular network.
 */
bool cat1_disconnect(void)
{
    cat1_send_command_internal("+CGACT=0,1", "OK", 3000U, 0);
    cat1_send_command_internal("+CGATT=0", "OK", 3000U, 0);
    s_connected = false;
    s_tcp_connected = false;
    s_status = CAT1_NO_CARRIER;
    return true;
}

/*
 * Establish TCP connection to remote host.
 */
bool cat1_tcp_connect(const char *host, uint16_t port)
{
    if (!host || !s_connected) {
        s_status = CAT1_ERROR;
        return false;
    }

    const cat1_config_t *cfg = cat1_get_config(NULL);

    /* Start single IP session */
    char cmd[128];
    int len = snprintf(cmd, sizeof(cmd), "+CIPSTART=\"TCP\",\"%s\",%d", host, port);
    if (len < 0 || len >= (int)sizeof(cmd)) {
        return false;
    }

    bool ok = cat1_send_command_internal(cmd, "CONNECT",
                                          cfg->connect_timeout_ms, cfg->retry_count);
    if (ok) {
        s_tcp_connected = true;
        log_info("TCP connected to %s:%d", host, port);
    } else {
        s_status = CAT1_NO_CARRIER;
    }
    return ok;
}

/*
 * Close the current TCP connection.
 */
bool cat1_tcp_close(void)
{
    if (!s_tcp_connected) {
        return true;
    }

    cat1_send_command_internal("+CIPCLOSE", "OK", 2000U, 0);
    s_tcp_connected = false;
    return true;
}

/*
 * Send MQTT CONNECT frame over the TCP connection.
 * Sends the binary CONNECT packet via AT+CIPSEND, then reads CONNACK.
 */
bool cat1_mqtt_connect(const char *client_id, const char *user, const char *pass)
{
    if (!client_id || !s_tcp_connected) {
        s_status = CAT1_ERROR;
        return false;
    }

    /* Build MQTT CONNECT packet (variable length based on credentials) */
    uint8_t pkt[128];
    uint16_t pkt_len = 0;

    /* Fixed header: CONNECT message, remaining length encoded */
    pkt[pkt_len++] = 0x10; /* CONNECT type */

    /* Variable header */
    /* Protocol name: "MQTT" */
    pkt[pkt_len++] = 0x00; pkt[pkt_len++] = 0x04;
    pkt[pkt_len++] = 'M'; pkt[pkt_len++] = 'Q';
    pkt[pkt_len++] = 'T'; pkt[pkt_len++] = 'T';
    pkt[pkt_len++] = 0x04; /* MQTT protocol level 4 */

    /* Connect flags */
    uint8_t flags = 0x02; /* Clean session */
    if (user && user[0] != '\0') {
        flags |= 0x04; /* Username flag */
    }
    if (pass && pass[0] != '\0') {
        flags |= 0x01; /* Password flag */
    }
    pkt[pkt_len++] = flags;

    /* Keep alive: 60 seconds */
    pkt[pkt_len++] = 0x3C; pkt[pkt_len++] = 0x00;

    /* Payload: client ID */
    uint16_t cid_len = (uint16_t)strlen(client_id);
    pkt[pkt_len++] = (uint8_t)(cid_len >> 8);
    pkt[pkt_len++] = (uint8_t)(cid_len & 0xFF);
    memcpy(pkt + pkt_len, client_id, cid_len);
    pkt_len += cid_len;

    /* Payload: username */
    if (user && user[0] != '\0') {
        uint16_t ulen = (uint16_t)strlen(user);
        pkt[pkt_len++] = (uint8_t)(ulen >> 8);
        pkt[pkt_len++] = (uint8_t)(ulen & 0xFF);
        memcpy(pkt + pkt_len, user, ulen);
        pkt_len += ulen;
    }

    /* Payload: password */
    if (pass && pass[0] != '\0') {
        uint16_t plen = (uint16_t)strlen(pass);
        pkt[pkt_len++] = (uint8_t)(plen >> 8);
        pkt[pkt_len++] = (uint8_t)(plen & 0xFF);
        memcpy(pkt + pkt_len, pass, plen);
        pkt_len += plen;
    }

    /* Encode remaining length (simple case: < 128) */
    uint8_t remaining_len = pkt_len - 1; /* subtract the message type byte */
    uint8_t encoded_len[4];
    int encoded_len_count = 0;
    do {
        uint8_t encoded_byte = remaining_len % 128;
        remaining_len /= 128;
        if (remaining_len > 0) {
            encoded_byte |= 0x80;
        }
        encoded_len[encoded_len_count++] = encoded_byte;
    } while (remaining_len > 0);

    /* Send via CIPSEND */
    const cat1_config_t *cfg = cat1_get_config(NULL);
    char send_cmd[64];
    int slen = snprintf(send_cmd, sizeof(send_cmd),
                        "+CIPSEND=%d", pkt_len + encoded_len_count);
    if (slen < 0 || slen >= (int)sizeof(send_cmd)) {
        return false;
    }

    if (!cat1_send_command_internal(send_cmd, ">", cfg->connect_timeout_ms, cfg->retry_count)) {
        return false;
    }

    /* Send fixed header (remaining length encoded + message type) */
    for (int i = 0; i < encoded_len_count; i++) {
        cat1_uart_putc(encoded_len[i]);
    }
    cat1_uart_putc(pkt[0]); /* CONNECT message type */

    /* Send variable header and payload */
    for (uint16_t i = 1; i < pkt_len; i++) {
        cat1_uart_putc(pkt[i]);
    }

    /* Wait for CONNACK (0x20) */
    bool ok = cat1_read_response("CONNACK", cfg->connect_timeout_ms);
    if (ok) {
        log_info("MQTT CONNECT successful");
    } else {
        log_error("MQTT CONNECT failed");
    }
    return ok;
}

/*
 * Publish an MQTT message over the existing connection.
 */
bool cat1_mqtt_publish(const char *topic, const uint8_t *data, uint16_t len)
{
    if (!topic || !data || !s_tcp_connected) {
        s_status = CAT1_ERROR;
        return false;
    }

    const cat1_config_t *cfg = cat1_get_config(NULL);

    /* Calculate PUBLEN: topic_len(2) + topic + data */
    uint16_t topic_len = (uint16_t)strlen(topic);
    uint16_t total_len = 2 + topic_len + len;

    /* Encode remaining length */
    uint8_t encoded_len[4];
    int encoded_len_count = 0;
    uint16_t remaining = total_len + 1; /* +1 for DUP/QoS/RETAIN byte */
    do {
        uint8_t byte = remaining % 128;
        remaining /= 128;
        if (remaining > 0) {
            byte |= 0x80;
        }
        encoded_len[encoded_len_count++] = byte;
    } while (remaining > 0);

    /* Total bytes to send */
    uint16_t send_total = encoded_len_count + 1 + total_len;

    /* Send CIPSEND command */
    char send_cmd[64];
    int slen = snprintf(send_cmd, sizeof(send_cmd), "+CIPSEND=%d", send_total);
    if (slen < 0 || slen >= (int)sizeof(send_cmd)) {
        return false;
    }

    if (!cat1_send_command_internal(send_cmd, ">", cfg->connect_timeout_ms, cfg->retry_count)) {
        return false;
    }

    /* SEND: remaining length encoded */
    for (int i = 0; i < encoded_len_count; i++) {
        cat1_uart_putc(encoded_len[i]);
    }

    /* PUBACK fixed header: DUP=0, QoS=1, Remaining length follows */
    cat1_uart_putc(0x30); /* PUBLISH, QoS 1 */

    /* Topic length (big-endian) */
    cat1_uart_putc((uint8_t)(topic_len >> 8));
    cat1_uart_putc((uint8_t)(topic_len & 0xFF));

    /* Topic string */
    for (uint16_t i = 0; i < topic_len; i++) {
        cat1_uart_putc(topic[i]);
    }

    /* Message ID for QoS 1 */
    static uint16_t s_msg_id = 1;
    cat1_uart_putc((uint8_t)(s_msg_id >> 8));
    cat1_uart_putc((uint8_t)(s_msg_id & 0xFF));
    s_msg_id++;

    /* Payload data */
    for (uint16_t i = 0; i < len; i++) {
        cat1_uart_putc(data[i]);
    }

    /* Wait for PUBACK */
    bool ok = cat1_read_response("PUBACK", cfg->connect_timeout_ms);
    return ok;
}

/*
 * Send MQTT DISCONNECT and close TCP connection.
 */
bool cat1_mqtt_disconnect(void)
{
    /* Send MQTT DISCONNECT (fixed header: 0xE0, remaining length: 0x00) */
    if (s_tcp_connected) {
        cat1_uart_putc(0xE0);
        cat1_uart_putc(0x00);
    }

    /* Close TCP connection */
    cat1_tcp_close();
    log_info("MQTT disconnected");
    return true;
}

/*
 * Get the current Cat1 module status.
 */
cat1_status_t cat1_get_status(void)
{
    return s_status;
}

/*
 * Check if the module is currently connected to the APN.
 */
bool cat1_is_connected(void)
{
    return s_connected;
}

/*
 * Get current signal strength (RSSI).
 */
int16_t cat1_get_signal_strength(void)
{
    if (!cat1_send_command_internal("+CSQ", "+CSQ:", 2000U, 0)) {
        return -127;
    }

    /* Parse response: +CSQ: <rssi>,<ber> */
    char *p = strstr(s_resp_buf, "+CSQ:");
    if (!p) {
        return -127;
    }

    int rssi = atoi(p + 5);
    if (rssi == 99) {
        return -127;  /* 99 means not detectable */
    }

    /* Convert 0-31 scale to dBm: RSSI = -113 + 2*rssi */
    return (int16_t)(-113 + 2 * rssi);
}

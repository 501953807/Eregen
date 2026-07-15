/*
 * Eregen (颐贞) - Cat1 AT Command Wrapper Implementation
 * 广和通 L610-CM cellular module AT command interface
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "cat1_at.h"
#include "gd32e230_usart.h"
#include "gd32e230_rcu.h"
#include "FreeRTOS.h"
#include "task.h"

/* Default AT command timeout */
#ifndef CAT1_CMD_TIMEOUT_MS
#define CAT1_CMD_TIMEOUT_MS    2000U
#endif

/* Response buffer */
static char s_resp_buf[CAT1_RESP_BUF_SIZE];

/* Connection state */
static bool s_connected = false;

/*
 * Send a raw byte over the Cat1 UART.
 * In production, this would use USART2 or USART3.
 */
static void cat1_uart_putc(char c)
{
    while (usart_flag_get(USART1, USART_FLAG_TC) == RESET) {
        /* Wait for transmit register empty */
    }
    usart_data_transmit(USART1, (uint8_t)c);
}

/*
 * Receive a byte from Cat1 UART with timeout.
 * Returns true if a byte was received within timeout.
 */
static bool cat1_uart_getc(char *c, uint32_t timeout_ms)
{
    TickType_t ticks = pdMS_TO_TICKS(timeout_ms);
    /* Placeholder: real implementation polls USART flag */
    (void)c;
    (void)ticks;
    return false;
}

/*
 * Send an AT command string with \r\n terminator.
 */
static bool cat1_send_command(const char *cmd, char *response,
                               uint16_t resp_size, uint32_t timeout_ms)
{
    if (!cmd || !response || resp_size == 0) {
        return false;
    }

    /* Send AT command */
    uint16_t cmd_len = (uint16_t)strlen(cmd);
    for (uint16_t i = 0; i < cmd_len; i++) {
        cat1_uart_putc(cmd[i]);
    }
    cat1_uart_putc('\r');
    cat1_uart_putc('\n');

    /* Wait for response */
    uint32_t start = xTaskGetTickCount();
    uint16_t resp_idx = 0;

    for (;;) {
        char ch;
        /* Poll for incoming data (simplified) */
        if (usart_flag_get(USART1, USART_FLAG_RBNE) != RESET) {
            ch = (char)usart_data_receive(USART1);
            if (resp_idx < resp_size - 1) {
                response[resp_idx++] = ch;
                response[resp_idx] = '\0';
            }

            /* Check for OK or ERROR response */
            if (strstr(response, "OK\r\n") != NULL) {
                return true;
            }
            if (strstr(response, "ERROR\r\n") != NULL) {
                return false;
            }
        }

        /* Check timeout */
        if ((xTaskGetTickCount() - start) > pdMS_TO_TICKS(timeout_ms)) {
            return false;
        }

        vTaskDelay(pdMS_TO_TICKS(1));
    }
}

/*
 * Initialize the Cat1 module.
 * Sends AT command and verifies module responds.
 */
bool cat1_init(void)
{
    rcu_periph_clock_enable(RCU_USART1);

    /* Configure USART1 TX/RX pins for Cat1 module */
    /* TX: PB6, RX: PB7 (alternate function) */
    s_connected = false;

    /* Send AT to wake up module */
    if (!cat1_send_command("AT", s_resp_buf, CAT1_RESP_BUF_SIZE, 1000U)) {
        return false;
    }

    /* Set module to text mode */
    cat1_send_command("AT+CMODE=1", s_resp_buf, CAT1_RESP_BUF_SIZE, 1000U);

    /* Set automatic PDP context activation */
    cat1_send_command("AT+CIPMODE=1", s_resp_buf, CAT1_RESP_BUF_SIZE, 1000U);

    return true;
}

/*
 * Send an AT command and wait for response.
 */
bool cat1_send_at(const char *cmd, char *response, uint16_t resp_size,
                  uint32_t timeout_ms)
{
    char full_cmd[64];
    uint16_t len = snprintf(full_cmd, sizeof(full_cmd), "AT%s", cmd);
    if (len >= sizeof(full_cmd)) {
        return false;
    }
    return cat1_send_command(full_cmd, response, resp_size, timeout_ms);
}

/*
 * Connect to cellular network and establish APN data session.
 */
bool cat1_connect_apn(const char *apn)
{
    if (!apn) {
        apn = CAT1_DEFAULT_APN;
    }

    /* Activate PDP context with specified APN */
    char cmd[64];
    snprintf(cmd, sizeof(cmd), "+CGDCONT=1,\"IP\",\"%s\"", apn);
    if (!cat1_send_command(cmd, s_resp_buf, CAT1_RESP_BUF_SIZE, 3000U)) {
        return false;
    }

    /* Activate the PDP context */
    if (!cat1_send_command("AT+CGACT=1,1", s_resp_buf, CAT1_RESP_BUF_SIZE, 5000U)) {
        return false;
    }

    /* Attach to GPRS service */
    if (!cat1_send_command("AT+CGATT=1", s_resp_buf, CAT1_RESP_BUF_SIZE, 5000U)) {
        return false;
    }

    s_connected = true;
    return true;
}

/*
 * Check if the module is connected to the APN.
 */
bool cat1_is_connected(void)
{
    /* Query PDP context state */
    if (!cat1_send_command("AT+CGACT?", s_resp_buf, CAT1_RESP_BUF_SIZE, 2000U)) {
        return false;
    }

    /* If previous connection succeeded, trust it */
    return s_connected;
}

/*
 * Get current signal strength (RSSI).
 * Returns RSSI in dBm (negative value), or -127 on error.
 */
int16_t cat1_get_signal_strength(void)
{
    if (!cat1_send_command("AT+CSQ", s_resp_buf, CAT1_RESP_BUF_SIZE, 2000U)) {
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

/*
 * Disconnect from APN.
 */
bool cat1_disconnect_apn(void)
{
    cat1_send_command("AT+CGACT=0,1", s_resp_buf, CAT1_RESP_BUF_SIZE, 3000U);
    cat1_send_command("AT+CGATT=0", s_resp_buf, CAT1_RESP_BUF_SIZE, 3000U);
    s_connected = false;
    return true;
}

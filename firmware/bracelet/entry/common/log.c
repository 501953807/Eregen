/*
 * Eregen (颐贞) - Logging System Implementation
 * Outputs formatted messages over UART with level prefix.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "log.h"
#include <stdarg.h>
#include <stdio.h>
#include <string.h>

#ifdef TEST_MODE
int log_test(void);
int main(void) { return log_test(); }
#endif

/* Level prefix strings */
static const char *log_prefixes[LOG_LEVEL_COUNT] = {
    "[D]",  /* DEBUG */
    "[I]",  /* INFO  */
    "[W]",  /* WARN  */
    "[E]"   /* ERROR */
};

/* Current minimum log level */
static log_level_t s_log_level = LOG_DEBUG;

/* UART transmit function pointer — set by board_init on embedded target */
static void (*s_uart_tx_func)(const uint8_t *data, uint16_t len) = NULL;

/*
 * Register the UART transmit callback.
 * Called by board_init() or equivalent on the embedded platform.
 * @param tx_func Function pointer that writes bytes to UART.
 */
void log_register_uart_tx(void (*tx_func)(const uint8_t *data, uint16_t len))
{
    s_uart_tx_func = tx_func;
}

void log_init(void)
{
    s_log_level = LOG_DEBUG;
}

void log_set_level(log_level_t level)
{
    if (level >= 0 && level < LOG_LEVEL_COUNT) {
        s_log_level = level;
    }
}

log_level_t log_get_level(void)
{
    return s_log_level;
}

static void log_write(log_level_t level, const char *fmt, va_list args)
{
    if (level < s_log_level) {
        return;
    }

    /* Format the message into a local buffer */
    char buf[256];
    va_list args_copy;
    va_copy(args_copy, args);
    int len = vsnprintf(buf, sizeof(buf), fmt, args_copy);
    va_end(args_copy);

    if (len <= 0) {
        return;
    }

    /* On embedded: prepend level prefix and send via UART */
#ifdef __EMBEDDED__
    if (s_uart_tx_func) {
        char out[264];
        uint16_t out_idx = 0;

        /* Write prefix */
        const char *prefix = log_prefixes[level];
        while (*prefix && out_idx < sizeof(out) - 2) {
            out[out_idx++] = *prefix++;
        }

        /* Write formatted message + newline */
        int msg_len = snprintf(out + out_idx, sizeof(out) - out_idx, "%s\n", buf);
        out_idx += (uint16_t)msg_len;

        s_uart_tx_func((const uint8_t *)out, out_idx);
    }
#else
    /* On host (TEST_MODE): print to stdout */
    printf("%s %s", log_prefixes[level], buf);
    if (buf[len - 1] != '\n') {
        printf("\n");
    }
#endif
}

void log_debug(const char *fmt, ...)
{
    va_list args;
    va_start(args, fmt);
    log_write(LOG_DEBUG, fmt, args);
    va_end(args);
}

void log_info(const char *fmt, ...)
{
    va_list args;
    va_start(args, fmt);
    log_write(LOG_INFO, fmt, args);
    va_end(args);
}

void log_warn(const char *fmt, ...)
{
    va_list args;
    va_start(args, fmt);
    log_write(LOG_WARN, fmt, args);
    va_end(args);
}

void log_error(const char *fmt, ...)
{
    va_list args;
    va_start(args, fmt);
    log_write(LOG_ERROR, fmt, args);
    va_end(args);
}

#ifdef TEST_MODE
#include <stdio.h>
#include <stdbool.h>

static int passed = 0;
static int failed = 0;

static void check(bool cond, const char *label)
{
    if (cond) {
        printf("  PASS: %s\n", label);
        passed++;
    } else {
        printf("  FAIL: %s\n", label);
        failed++;
    }
}

int log_test(void)
{
    printf("Logging tests:\n");

    log_init();

    /* Default level is DEBUG */
    check(log_get_level() == LOG_DEBUG, "default level is DEBUG");

    /* Level filtering */
    log_set_level(LOG_WARN);
    check(log_get_level() == LOG_WARN, "level set to WARN");

    printf("--- Testing level filtering (only WARN+ should appear above) ---\n");
    log_debug("this should NOT appear");
    log_info("this should NOT appear");
    log_warn("this SHOULD appear");
    log_error("this SHOULD appear");
    printf("--- End filtering test ---\n");

    /* Reset to DEBUG for further tests */
    log_set_level(LOG_DEBUG);

    /* Format string handling */
    printf("--- Testing format strings ---\n");
    log_info("Value: %d, String: %s, Float: %.1f", 42, "hello", 3.14f);
    printf("--- End format test ---\n");

    /* Empty message */
    log_info("");

    printf("\nResults: %d passed, %d failed\n", passed, failed);
    return failed > 0 ? 1 : 0;
}
#endif

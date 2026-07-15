/*
 * Eregen (颐贞) - Logging System Header
 * UART-based structured logging with configurable severity levels.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef LOG_H
#define LOG_H

#include <stdint.h>

/** Log severity levels. */
typedef enum {
    LOG_DEBUG = 0,
    LOG_INFO,
    LOG_WARN,
    LOG_ERROR,
    LOG_LEVEL_COUNT
} log_level_t;

/**
 * Initialize the logging subsystem.
 * Must be called before any log_* function.
 */
void log_init(void);

/**
 * Set the minimum log level. Messages below this level are suppressed.
 * @param level Minimum level to output.
 */
void log_set_level(log_level_t level);

/**
 * Get the current minimum log level.
 * @return Current log level.
 */
log_level_t log_get_level(void);

/**
 * Log a debug message.
 */
void log_debug(const char *fmt, ...);

/**
 * Log an informational message.
 */
void log_info(const char *fmt, ...);

/**
 * Log a warning message.
 */
void log_warn(const char *fmt, ...);

/**
 * Log an error message.
 */
void log_error(const char *fmt, ...);

#endif /* LOG_H */

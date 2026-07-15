/*
 * Eregen (颐贞) - APP Linkage Command Parser
 * Smart pillbox tier — Parse MQTT downlink commands from cloud/family app
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef APP_LINK_H
#define APP_LINK_H

#include "esp_err.h"
#include <stddef.h>

/**
 * Initialize APP linkage subsystem.
 * Sets up MQTT subscription callbacks for pillbox commands.
 *
 * @return ESP_OK on success
 */
esp_err_t applink_init(void);

/**
 * Parse an incoming MQTT message and dispatch to appropriate handler.
 * Handles message types: "med_rule", "tts", "config".
 *
 * @param topic MQTT topic string
 * @param payload Message payload JSON
 * @param payload_len Length of payload
 * @return ESP_OK if message was handled, ESP_FAIL if unknown type
 */
esp_err_t applink_parse_mqtt_message(const char *topic,
                                     const uint8_t *payload,
                                     size_t payload_len);

/**
 * Handle a "med_rule" downlink message.
 * Replaces current reminder rules with those from the message.
 *
 * @param rules Parsed rule array
 * @param count Number of rules
 * @return ESP_OK on success
 */
esp_err_t applink_handle_set_rules(const void *rules, uint8_t count);

/**
 * Handle a "pause_reminder" downlink command.
 * Temporarily disables all reminders until resume.
 *
 * @return ESP_OK on success
 */
esp_err_t applink_handle_pause_reminder(void);

/**
 * Handle a "resume_reminder" downlink command.
 * Re-enables reminders after pause.
 *
 * @return ESP_OK on success
 */
esp_err_t applink_handle_resume_reminder(void);

#endif /* APP_LINK_H */

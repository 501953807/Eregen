/*
 * Eregen (颐贞) - Heartbeat Module Header
 * Periodic heartbeat publisher via Cat1/MQTT interface.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef HEARTBEAT_H
#define HEARTBEAT_H

/**
 * Start periodic heartbeat publishing.
 * Must be called after board_init() and cat1_init().
 * Sends heartbeat every 5 minutes via MQTT.
 */
void heartbeat_start(void);

/**
 * Stop periodic heartbeat publishing.
 */
void heartbeat_stop(void);

#endif /* HEARTBEAT_H */

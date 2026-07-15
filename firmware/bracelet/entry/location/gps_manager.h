/*
 * Eregen (颐贞) - GPS Location Manager Header
 * Manages GPS query intervals and Cat1 connection modes.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef GPS_MANAGER_H
#define GPS_MANAGER_H

#include <stdint.h>
#include <stdbool.h>
#include "gps_nmea.h"

/**
 * Location modes for power/accuracy trade-off.
 */
typedef enum {
    LOC_NORMAL = 0,       /* Normal: GPS every 30s, Cat1 sleeps between queries */
    LOC_ALERT,            /* Alert: GPS every 1s, Cat1 always online */
    LOC_POWER_SAVE        /* Power save: GPS off, use last known position */
} location_mode_t;

/**
 * Location manager state.
 */
typedef struct {
    location_mode_t current_mode;
    location_mode_t target_mode;
    uint32_t last_fix_time;   /* Tick count of last GPS fix */
    uint32_t interval_ms;     /* Query interval: NORMAL=30000, ALERT=1000, POWER_SAVE=0 */
    bool cat1_online;         /* Whether Cat1 TCP/MQTT connection is active */
} gps_manager_t;

/**
 * Initialize the GPS location manager.
 * Sets default mode to LOC_NORMAL with 30s interval.
 */
void gps_manager_init(void);

/**
 * Set the location mode. Triggers mode transition logic.
 * @param mode Target location mode.
 */
void gps_manager_set_mode(location_mode_t mode);

/**
 * Get the current location mode.
 * @return Current location mode.
 */
location_mode_t gps_manager_get_mode(void);

/**
 * Get the latest GPS location.
 * In POWER_SAVE mode, returns the last known fix (may be stale).
 * @param fix Output buffer for the GPS fix (must not be NULL).
 * @return true if a valid fix is available, false otherwise.
 */
bool gps_manager_get_location(gps_fix_t *fix);

/**
 * Periodic tick function — call from RTOS task at ~100ms intervals.
 * Handles:
 *   - Mode transitions (NORMAL <-> ALERT <-> POWER_SAVE)
 *   - GPS query scheduling based on interval
 *   - Cat1 connection management
 */
void gps_manager_tick(void);

#endif /* GPS_MANAGER_H */

/*
 * Eregen (颐贞) - Geofence Manager Header (Plus Tier)
 * Manages configurable safe-zone geofences stored in NVS.
 * Supports up to 5 circular zones received via MQTT config updates.
 * Triggers alerts when the elder leaves any enabled zone.
 *
 * MIT License
 *
 * Copyright (c) 2026 Eregen (颐贞). All rights reserved.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

#ifndef GEOFENCE_MANAGER_H
#define GEOFENCE_MANAGER_H

#include <stdint.h>
#include <stdbool.h>

/* Maximum number of geofence zones supported. */
#define GEOFENCE_MGR_MAX_ZONES       5U

/* Minimum zone radius in meters (reject smaller zones). */
#define GEOFENCE_MGR_MIN_RADIUS_M    50.0f

/* Default zone radius when not specified by cloud. */
#define GEOFENCE_MGR_DEFAULT_RADIUS  200.0f

/* Alert cooldown: minimum seconds between duplicate alerts for same zone. */
#define GEOFENCE_MGR_ALERT_COOLDOWN_S  300U

/* Geofence zone type constants. */
#define GEOFENCE_ZONE_CIRCLE  1U

/**
 * Geofence zone definition — stored in NVS and updated via MQTT.
 */
typedef struct {
    uint8_t   id;             /* Zone index 0..MAX_ZONES-1 */
    uint8_t   type;           /* GEOFENCE_ZONE_CIRCLE */
    char      name[24];       /* Human-readable label, e.g. "Home" */
    float     center_lat;     /* Center latitude (degrees) */
    float     center_lon;     /* Center longitude (degrees) */
    float     radius_meters;  /* Zone radius in meters */
    bool      enabled;        /* Whether this zone is active */
} geofence_zone_t;

/**
 * Geofence manager result codes.
 */
typedef enum {
    GEOFENCE_OK            =  0,
    GEOFENCE_ERR_NULL_PTR  = -1,
    GEOFENCE_ERR_NO_ZONE   = -2,
    GEOFENCE_ERR_OUTSIDE   = -3,
    GEOFENCE_ERR_FULL      = -4,
    GEOFENCE_ERR_INVALID   = -5,
} geofence_result_t;

/**
 * Geofence state reported to the alert task.
 */
typedef struct {
    bool         inside;          /* True if inside at least one zone */
    uint8_t      exited_zone_id;  /* Zone ID that was just exited (0xFF if none) */
    uint8_t      entered_zone_id; /* Zone ID that was just entered (0xFF if none) */
    uint32_t     last_alert_tick; /* FreeRTOS tick count of last alert sent */
} geofence_state_t;

/*
 * Initialize the geofence manager.
 * Loads configured zones from NVS storage.
 * @param nvs_store Pointer to NVS store abstraction.
 * @return GEOFENCE_OK on success.
 */
int geofence_init(void *nvs_store);

/*
 * Load all zones from NVS into RAM.
 * @return Number of zones loaded, or negative on error.
 */
int geofence_load_from_nvs(void);

/*
 * Save all zones to NVS storage.
 * @return GEOFENCE_OK on success.
 */
int geofence_save_to_nvs(void);

/*
 * Add or update a geofence zone.
 * If a zone with the same ID exists it is updated; otherwise a new zone is added.
 * @param zone Pointer to zone definition to add/update.
 * @return GEOFENCE_OK on success, GEOFENCE_ERR_FULL if max zones reached.
 */
int geofence_add_zone(const geofence_zone_t *zone);

/*
 * Remove a geofence zone by ID.
 * @param zone_id Zone ID to remove.
 * @return GEOFENCE_OK on success, GEOFENCE_ERR_NO_ZONE if not found.
 */
int geofence_remove_zone(uint8_t zone_id);

/*
 * Get a zone by ID.
 * @param zone_id Zone ID to retrieve.
 * @param[out] out Output buffer for zone data (must not be NULL).
 * @return GEOFENCE_OK on success.
 */
int geofence_get_zone(uint8_t zone_id, geofence_zone_t *out);

/*
 * Check whether the current GPS position is inside any enabled geofence zone.
 * Uses haversine distance for circle containment.
 * @param lat Current latitude in degrees.
 * @param lon Current longitude in degrees.
 * @param[out] state Output geofence state (may be NULL if caller ignores state).
 * @return GEOFENCE_OK on success.
 */
int geofence_check_position(double lat, double lon, geofence_state_t *state);

/*
 * Calculate haversine distance between two GPS points.
 * @param lat1 Point 1 latitude (degrees).
 * @param lon1 Point 1 longitude (degrees).
 * @param lat2 Point 2 latitude (degrees).
 * @param lon2 Point 2 longitude (degrees).
 * @return Distance in meters.
 */
float geofence_haversine_distance(float lat1, float lon1, float lat2, float lon2);

/*
 * Get the number of configured zones.
 * @return Number of zones currently stored.
 */
uint8_t geofence_get_zone_count(void);

/*
 * Reset all zones and clear NVS storage.
 * Called during factory reset or initial provisioning.
 * @return GEOFENCE_OK on success.
 */
int geofence_reset_all(void);

/* ---- Test helpers ---- */

#ifdef TEST_MODE
/* Inject mock zone data without touching NVS. */
void geofence_set_mock_zones(const geofence_zone_t *zones, uint8_t count);

/* Clear mock zones. */
void geofence_clear_mock_zones(void);
#endif

#endif /* GEOFENCE_MANAGER_H */

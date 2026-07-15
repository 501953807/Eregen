/*
 * Eregen (颐贞) - Local Geofence Header
 * Circle and polygon geofence definitions with point-in-fence checks.
 * Uses haversine distance for circles, ray casting for polygons.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef GEOFENCE_H
#define GEOFENCE_H

#include <stdint.h>
#include <stdbool.h>

/** Maximum number of geofences supported. */
#define MAX_GEOFENCES 8U

/** Maximum points in a polygon geofence. */
#define GEOFENCE_MAX_POINTS 32U

/** Geofence type constants. */
#define GEOFENCE_TYPE_CIRCLE   1U
#define GEOFENCE_TYPE_POLYGON  2U

/**
 * Circular geofence definition.
 * Center is given as latitude/longitude; radius in meters.
 */
typedef struct {
    float center_lat;       /* Center latitude in degrees */
    float center_lon;       /* Center longitude in degrees */
    float radius_meters;    /* Radius in meters */
    bool enabled;           /* Whether this fence is active */
    const char *name;       /* Human-readable name (e.g., "home") */
} geofence_circle_t;

/**
 * Polygon geofence definition.
 * Points are stored as alternating lat/lon pairs.
 */
typedef struct {
    float points_lat[GEOFENCE_MAX_POINTS];  /* Latitude of each vertex */
    float points_lon[GEOFENCE_MAX_POINTS];  /* Longitude of each vertex */
    uint8_t point_count;                    /* Number of vertices (3+ for triangle) */
    bool enabled;                           /* Whether this fence is active */
    const char *name;                       /* Human-readable name */
} geofence_polygon_t;

/**
 * Union type for either geofence variant.
 * The `type` field indicates which member is valid (1=circle, 2=polygon).
 */
typedef union {
    geofence_circle_t circle;
    geofence_polygon_t polygon;
    uint8_t type; /* 1 = circle, 2 = polygon */
} geofence_entry_t;

/**
 * Check if a point is inside a circular geofence.
 * Uses haversine formula for distance calculation.
 * @param lat    Point latitude in degrees.
 * @param lon    Point longitude in degrees.
 * @param fence  Pointer to the circular geofence.
 * @return true if the point is within the fence radius.
 */
bool geofence_point_in_circle(float lat, float lon, const geofence_circle_t *fence);

/**
 * Check if a point is inside a polygonal geofence.
 * Uses the ray casting algorithm.
 * @param lat    Point latitude in degrees.
 * @param lon    Point longitude in degrees.
 * @param fence  Pointer to the polygonal geofence.
 * @return true if the point is inside the polygon.
 */
bool geofence_point_in_polygon(float lat, float lon, const geofence_polygon_t *fence);

/**
 * Check a point against an array of geofences.
 * @param lat    Point latitude in degrees.
 * @param lon    Point longitude in degrees.
 * @param fences Array of geofence entries to check.
 * @param count  Number of entries in the array.
 * @return 0 if point is inside all fences,
 *         >0 if point exited the fence at that index,
 *         -1 on error (invalid input or count > MAX_GEOFENCES).
 */
int geofence_check(float lat, float lon, const geofence_entry_t *fences, uint8_t count);

#endif /* GEOFENCE_H */

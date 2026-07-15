/*
 * Eregen (颐贞) - Local Geofence Implementation
 * Haversine-based circle containment and ray-casting polygon check.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "geofence.h"
#include <math.h>
#include <string.h>

#ifdef TEST_MODE
/* Host-mode: no embedded dependencies. Log stubs discard all output. */
static void log_debug(const char *fmt, ...) { (void)fmt; }
static void log_info(const char *fmt, ...) { (void)fmt; }
static void log_warn(const char *fmt, ...) { (void)fmt; }
static void log_error(const char *fmt, ...) { (void)fmt; }
#else
#include "log.h"
#endif

/* Earth radius in meters (WGS84 mean radius) */
#define EARTH_RADIUS_M 6371000.0f

/* Degrees to radians conversion factor */
#define DEG_TO_RAD 0.017453292519943295f

/*
 * Convert degrees to radians.
 */
static float deg_to_rad(float deg)
{
    return deg * DEG_TO_RAD;
}

/*
 * Haversine distance between two points on Earth's surface.
 * Returns distance in meters.
 */
static float haversine_distance(float lat1, float lon1, float lat2, float lon2)
{
    float dlat = deg_to_rad(lat2 - lat1);
    float dlon = deg_to_rad(lon2 - lon1);

    float a = sinf(dlat / 2.0f) * sinf(dlat / 2.0f) +
              cosf(deg_to_rad(lat1)) * cosf(deg_to_rad(lat2)) *
              sinf(dlon / 2.0f) * sinf(dlon / 2.0f);

    float c = 2.0f * atan2f(sqrtf(a), sqrtf(1.0f - a));

    return EARTH_RADIUS_M * c;
}

/*
 * Check if a point is inside a circular geofence.
 */
bool geofence_point_in_circle(float lat, float lon, const geofence_circle_t *fence)
{
    if (!fence || !fence->enabled) {
        return false;
    }

    float dist = haversine_distance(lat, lon,
                                     fence->center_lat, fence->center_lon);
    return dist <= fence->radius_meters;
}

/*
 * Ray casting algorithm for point-in-polygon test.
 * Casts a horizontal ray from the point to the right and counts
 * edge crossings. Odd count = inside, even count = outside.
 */
bool geofence_point_in_polygon(float lat, float lon, const geofence_polygon_t *fence)
{
    if (!fence || !fence->enabled || fence->point_count < 3) {
        return false;
    }

    bool inside = false;
    uint8_t j = fence->point_count - 1; /* Last vertex */

    for (uint8_t i = 0; i < fence->point_count; i++) {
        float xi = fence->points_lat[i];
        float yi = fence->points_lon[i];
        float xj = fence->points_lat[j];
        float yj = fence->points_lon[j];

        /* Check if the ray from point crosses this edge */
        if (((yi > lon) != (yj > lon)) &&
            (lat < (xj - xi) * (lon - yi) / (yj - yi) + xi)) {
            inside = !inside;
        }
        j = i;
    }

    return inside;
}

/*
 * Check a point against an array of geofences.
 * Returns:
 *   0   = point is inside ALL enabled fences
 *   n>0 = point exited fence at index n (first fence where it fails)
 *  -1   = error (invalid input)
 */
int geofence_check(float lat, float lon, const geofence_entry_t *fences, uint8_t count)
{
    if (!fences || count == 0 || count > MAX_GEOFENCES) {
        log_error("Invalid geofence input: ptr=%p count=%u", (void*)fences, count);
        return -1;
    }

    int first_exit = 0;

    for (uint8_t i = 0; i < count; i++) {
        bool in_fence = false;

        switch (fences[i].type) {
        case GEOFENCE_TYPE_CIRCLE:
            in_fence = geofence_point_in_circle(lat, lon, &fences[i].circle);
            break;
        case GEOFENCE_TYPE_POLYGON:
            in_fence = geofence_point_in_polygon(lat, lon, &fences[i].polygon);
            break;
        default:
            log_warn("Unknown geofence type %u at index %u", fences[i].type, i);
            continue;
        }

        if (!in_fence) {
            if (first_exit == 0) {
                first_exit = (int)i + 1; /* 1-based index */
            }
        }
    }

    if (first_exit > 0) {
        log_warn("Point (%.4f, %.4f) exited geofence at index %d",
                 lat, lon, first_exit);
    }

    return first_exit;
}

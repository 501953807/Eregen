/*
 * Eregen (颐贞) - Geofence Manager Implementation (Plus Tier)
 * NVS-backed zone management with haversine-based containment checks.
 * Zones are configured via MQTT messages from the cloud or family APP.
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

#include "geofence_manager.h"
#include <math.h>
#include <string.h>

#ifndef M_PI_F
#define M_PI_F 3.14159265358979323846f
#endif

/* Earth mean radius in meters (WGS84) */
#define EARTH_RADIUS_M  6371000.0f

/* NVS key for geofence zones blob */
#define GEOFENCE_NVS_KEY  "geofence_zones"

/* ---- Internal state ---- */

/** Pointer to NVS store abstraction — set during init. */
static void *s_nvs_store = NULL;

/** In-memory zone table (up to MAX_ZONES). */
static geofence_zone_t s_zones[GEOFENCE_MGR_MAX_ZONES];

/** Number of valid entries in s_zones. */
static uint8_t s_zone_count = 0;

/** Whether we are using mock data (test mode). */
static bool s_mock_mode = false;

/* ---- Haversine helper ---- */

static float deg_to_rad(float deg)
{
    return deg * (M_PI_F / 180.0f);
}

float geofence_haversine_distance(float lat1, float lon1, float lat2, float lon2)
{
    float dlat = deg_to_rad(lat2 - lat1);
    float dlon = deg_to_rad(lon2 - lon1);

    float a = sinf(dlat * 0.5f) * sinf(dlat * 0.5f) +
              cosf(deg_to_rad(lat1)) * cosf(deg_to_rad(lat2)) *
              sinf(dlon * 0.5f) * sinf(dlon * 0.5f);

    float c = 2.0f * atan2f(sqrtf(a), sqrtf(1.0f - a));
    return EARTH_RADIUS_M * c;
}

/* ---- NVS persistence helpers ---- */

/*
 * Serialize zones to a flat byte buffer for NVS storage.
 * Each zone takes sizeof(geofence_zone_t) bytes.
 * Returns the number of bytes written.
 */
static int serialize_zones(uint8_t *buf, uint16_t buf_len)
{
    if (!buf || buf_len < (GEOFENCE_MGR_MAX_ZONES * sizeof(geofence_zone_t))) {
        return -1;
    }

    /* First byte: count of zones */
    buf[0] = s_zone_count;

    /* Followed by zone structs */
    memcpy(buf + 1, s_zones, s_zone_count * sizeof(geofence_zone_t));

    return (int)(1 + s_zone_count * sizeof(geofence_zone_t));
}

/*
 * Deserialize zones from an NVS blob.
 * Returns number of zones loaded, or negative on error.
 */
static int deserialize_zones(const uint8_t *buf, uint16_t len)
{
    if (!buf || len < 1) {
        return -1;
    }

    uint8_t count = buf[0];
    if (count > GEOFENCE_MGR_MAX_ZONES) {
        return -1;
    }

    s_zone_count = count;
    memcpy(s_zones, buf + 1, count * sizeof(geofence_zone_t));

    return (int)count;
}

/* ---- Public API ---- */

int geofence_init(void *nvs_store)
{
    if (!nvs_store) {
        return GEOFENCE_ERR_NULL_PTR;
    }

    s_nvs_store = nvs_store;
    s_zone_count = 0;
    memset(s_zones, 0, sizeof(s_zones));
    s_mock_mode = false;

    return geofence_load_from_nvs();
}

int geofence_load_from_nvs(void)
{
    if (!s_nvs_store) {
        return GEOFENCE_ERR_NULL_PTR;
    }

    /*
     * NVS store abstraction interface (expected by caller):
     *   void *nvs_handle;
     *   int nvs_read(void *handle, const char *key, void *buf, uint16_t len);
     *   int nvs_write(void *handle, const char *key, const void *buf, uint16_t len);
     * We cast the opaque pointer and call through function pointers stored
     * at known offsets, or rely on the caller providing a struct.
     *
     * For portability we define a minimal expected layout here.
     * The actual implementation depends on the NVS driver used.
     */
    typedef struct {
        void   *handle;
        int    (*read_fn)(void *, const char *, void *, uint16_t);
        int    (*write_fn)(void *, const char *, const void *, uint16_t);
    } nvs_ops_t;

    nvs_ops_t *ops = (nvs_ops_t *)s_nvs_store;

    if (!ops->read_fn || !ops->write_fn) {
        return GEOFENCE_ERR_NULL_PTR;
    }

    uint8_t raw[(GEOFENCE_MGR_MAX_ZONES * sizeof(geofence_zone_t)) + 1];
    int ret = ops->read_fn(ops->handle, GEOFENCE_NVS_KEY, raw, sizeof(raw));

    if (ret == 0) {
        int count = deserialize_zones(raw, sizeof(raw));
        if (count >= 0) {
            return count;
        }
    }

    /* No valid data in NVS — start with empty zone list. */
    s_zone_count = 0;
    return 0;
}

int geofence_save_to_nvs(void)
{
    if (!s_nvs_store) {
        return GEOFENCE_ERR_NULL_PTR;
    }

    typedef struct {
        void   *handle;
        int    (*read_fn)(void *, const char *, void *, uint16_t);
        int    (*write_fn)(void *, const char *, const void *, uint16_t);
    } nvs_ops_t;

    nvs_ops_t *ops = (nvs_ops_t *)s_nvs_store;

    if (!ops->write_fn) {
        return GEOFENCE_ERR_NULL_PTR;
    }

    uint8_t raw[(GEOFENCE_MGR_MAX_ZONES * sizeof(geofence_zone_t)) + 1];
    int len = serialize_zones(raw, sizeof(raw));
    if (len <= 0) {
        return GEOFENCE_ERR_INVALID;
    }

    return ops->write_fn(ops->handle, GEOFENCE_NVS_KEY, raw, (uint16_t)len);
}

int geofence_add_zone(const geofence_zone_t *zone)
{
    if (!zone) {
        return GEOFENCE_ERR_NULL_PTR;
    }

    /* Validate zone parameters. */
    if (zone->type != GEOFENCE_ZONE_CIRCLE) {
        return GEOFENCE_ERR_INVALID;
    }
    if (zone->radius_meters < GEOFENCE_MGR_MIN_RADIUS_M) {
        return GEOFENCE_ERR_INVALID;
    }
    if (zone->id >= GEOFENCE_MGR_MAX_ZONES) {
        return GEOFENCE_ERR_INVALID;
    }

    /* Check if zone already exists — update in place. */
    for (uint8_t i = 0; i < s_zone_count; i++) {
        if (s_zones[i].id == zone->id) {
            strncpy(s_zones[i].name, zone->name, sizeof(s_zones[i].name) - 1);
            s_zones[i].name[sizeof(s_zones[i].name) - 1] = '\0';
            s_zones[i].center_lat = zone->center_lat;
            s_zones[i].center_lon = zone->center_lon;
            s_zones[i].radius_meters = zone->radius_meters;
            s_zones[i].enabled = zone->enabled;
            return GEOFENCE_OK;
        }
    }

    /* Add new zone if room available. */
    if (s_zone_count >= GEOFENCE_MGR_MAX_ZONES) {
        return GEOFENCE_ERR_FULL;
    }

    uint8_t idx = s_zone_count++;
    s_zones[idx].id = zone->id;
    s_zones[idx].type = zone->type;
    strncpy(s_zones[idx].name, zone->name, sizeof(s_zones[idx].name) - 1);
    s_zones[idx].name[sizeof(s_zones[idx].name) - 1] = '\0';
    s_zones[idx].center_lat = zone->center_lat;
    s_zones[idx].center_lon = zone->center_lon;
    s_zones[idx].radius_meters = zone->radius_meters;
    s_zones[idx].enabled = zone->enabled;

    return GEOFENCE_OK;
}

int geofence_remove_zone(uint8_t zone_id)
{
    for (uint8_t i = 0; i < s_zone_count; i++) {
        if (s_zones[i].id == zone_id) {
            /* Shift remaining zones down. */
            memmove(&s_zones[i], &s_zones[i + 1],
                    (s_zone_count - i - 1) * sizeof(geofence_zone_t));
            s_zone_count--;
            return GEOFENCE_OK;
        }
    }
    return GEOFENCE_ERR_NO_ZONE;
}

int geofence_get_zone(uint8_t zone_id, geofence_zone_t *out)
{
    if (!out) {
        return GEOFENCE_ERR_NULL_PTR;
    }

    for (uint8_t i = 0; i < s_zone_count; i++) {
        if (s_zones[i].id == zone_id) {
            *out = s_zones[i];
            return GEOFENCE_OK;
        }
    }
    return GEOFENCE_ERR_NO_ZONE;
}

int geofence_check_position(double lat, double lon, geofence_state_t *state)
{
    if (s_zone_count == 0) {
        /* No zones configured — always considered "inside". */
        if (state) {
            state->inside = true;
            state->exited_zone_id = 0xFF;
            state->entered_zone_id = 0xFF;
        }
        return GEOFENCE_OK;
    }

    bool was_inside = (state && state->inside);
    bool is_inside = false;
    uint8_t exited = 0xFF;
    uint8_t entered = 0xFF;

    for (uint8_t i = 0; i < s_zone_count; i++) {
        if (!s_zones[i].enabled) {
            continue;
        }

        float dist = geofence_haversine_distance(
            (float)lat, (float)lon,
            s_zones[i].center_lat, s_zones[i].center_lon
        );

        bool in_zone = (dist <= s_zones[i].radius_meters);

        if (in_zone) {
            is_inside = true;
        } else {
            /* Just exited this zone. */
            if (was_inside) {
                exited = s_zones[i].id;
            }
        }
    }

    /* Determine enter/exit transitions. */
    if (is_inside && !was_inside) {
        /* Entered some zone. */
        for (uint8_t i = 0; i < s_zone_count; i++) {
            if (s_zones[i].enabled) {
                float dist = geofence_haversine_distance(
                    (float)lat, (float)lon,
                    s_zones[i].center_lat, s_zones[i].center_lon
                );
                if (dist <= s_zones[i].radius_meters) {
                    entered = s_zones[i].id;
                    break;
                }
            }
        }
    }

    if (state) {
        state->inside = is_inside;
        state->exited_zone_id = exited;
        state->entered_zone_id = entered;
    }

    if (exited != 0xFF) {
        return GEOFENCE_ERR_OUTSIDE;
    }

    return GEOFENCE_OK;
}

uint8_t geofence_get_zone_count(void)
{
    return s_zone_count;
}

int geofence_reset_all(void)
{
    if (!s_nvs_store) {
        return GEOFENCE_ERR_NULL_PTR;
    }

    typedef struct {
        void   *handle;
        int    (*read_fn)(void *, const char *, void *, uint16_t);
        int    (*write_fn)(void *, const char *, const void *, uint16_t);
    } nvs_ops_t;

    nvs_ops_t *ops = (nvs_ops_t *)s_nvs_store;

    if (ops->write_fn) {
        /* Write zero count to clear NVS. */
        uint8_t clear_buf[1] = {0};
        ops->write_fn(ops->handle, GEOFENCE_NVS_KEY, clear_buf, 1);
    }

    s_zone_count = 0;
    memset(s_zones, 0, sizeof(s_zones));
    return GEOFENCE_OK;
}

/* ---- Test helpers ---- */

#ifdef TEST_MODE

static geofence_zone_t s_test_zones[GEOFENCE_MGR_MAX_ZONES];
static uint8_t s_test_zone_count = 0;

void geofence_set_mock_zones(const geofence_zone_t *zones, uint8_t count)
{
    s_mock_mode = true;
    s_test_zone_count = (count > GEOFENCE_MGR_MAX_ZONES) ? GEOFENCE_MGR_MAX_ZONES : count;
    memcpy(s_test_zones, zones, s_test_zone_count * sizeof(geofence_zone_t));
}

void geofence_clear_mock_zones(void)
{
    s_mock_mode = false;
    s_test_zone_count = 0;
}

#endif

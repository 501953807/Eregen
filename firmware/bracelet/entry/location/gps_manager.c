/*
 * Eregen (颐贞) - GPS Location Manager Implementation
 * Manages GPS query intervals and Cat1 connection modes.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "location/gps_manager.h"
#include "cat1_at.h"
#include "log.h"
#include <string.h>

/* Tick interval for this manager (milliseconds) */
#define MANAGER_TICK_MS      100U

/* Time to consider a fix stale in normal mode (seconds) */
#define FIX_STALE_THRESHOLD  60U

/* Last known position for power-save fallback */
static gps_fix_t s_last_known_fix;

/* Manager state */
static gps_manager_t s_mgr;

/* Internal: elapsed ticks since last GPS query */
static uint32_t s_query_ticks = 0;

/*
 * Apply mode-specific settings.
 */
static void apply_mode_settings(location_mode_t mode)
{
    switch (mode) {
    case LOC_NORMAL:
        s_mgr.interval_ms = 30000U;
        log_info("Location mode: NORMAL (30s interval)");
        break;
    case LOC_ALERT:
        s_mgr.interval_ms = 1000U;
        log_info("Location mode: ALERT (1s interval)");
        break;
    case LOC_POWER_SAVE:
        s_mgr.interval_ms = 0U;
        log_info("Location mode: POWER_SAVE (GPS off)");
        break;
    }
}

/*
 * Manage Cat1 connection based on mode.
 */
static void manage_cat1_connection(void)
{
    if (s_mgr.cat1_online) {
        /* Currently online — stay online only in ALERT mode */
        if (s_mgr.current_mode == LOC_NORMAL) {
            /* Transitioning to normal: keep Cat1 connected for periodic uploads */
            /* In production, would schedule periodic MQTT publish */
        } else if (s_mgr.current_mode == LOC_POWER_SAVE) {
            /* Power save: disconnect Cat1 to save power */
            if (cat1_is_connected()) {
                cat1_mqtt_disconnect();
                cat1_disconnect();
                s_mgr.cat1_online = false;
                log_info("Cat1 disconnected (power save)");
            }
        }
        /* ALERT mode: Cat1 stays online, no action needed */
    } else {
        /* Not online — connect if in ALERT mode */
        if (s_mgr.current_mode == LOC_ALERT) {
            if (!cat1_is_connected()) {
                if (cat1_connect()) {
                    cat1_tcp_connect(CAT1_MQTT_BROKER, CAT1_MQTT_PORT);
                    s_mgr.cat1_online = true;
                    log_info("Cat1 connected for ALERT mode");
                }
            }
        }
    }
}

/*
 * Perform a single GPS query.
 */
static void query_gps(void)
{
    gps_fix_t fix;
    if (gps_get_fix(&fix) && fix.valid) {
        s_last_known_fix = fix;
        s_mgr.last_fix_time = 0; /* Reset stale counter */
        log_debug("GPS fix acquired: lat=%.4f lon=%.4f",
                  fix.latitude, fix.longitude);
    } else {
        log_debug("No valid GPS fix available");
    }
}

/*
 * Initialize the GPS location manager.
 */
void gps_manager_init(void)
{
    memset(&s_mgr, 0, sizeof(s_mgr));
    memset(&s_last_known_fix, 0, sizeof(s_last_known_fix));

    s_mgr.current_mode = LOC_NORMAL;
    s_mgr.target_mode = LOC_NORMAL;
    s_mgr.interval_ms = 30000U;
    s_mgr.cat1_online = false;
    s_mgr.last_fix_time = 0;
    s_query_ticks = 0;

    log_info("GPS manager initialized (NORMAL mode)");
}

/*
 * Set the location mode. Triggers mode transition logic.
 */
void gps_manager_set_mode(location_mode_t mode)
{
    if (mode == s_mgr.current_mode) {
        return; /* No change needed */
    }

    s_mgr.target_mode = mode;
    log_info("Switching location mode: %d -> %d",
             s_mgr.current_mode, mode);

    /* Apply new settings immediately */
    apply_mode_settings(mode);

    /* Handle Cat1 connection */
    manage_cat1_connection();

    /* Update current mode after transition completes */
    s_mgr.current_mode = mode;

    /* Reset query timer for new interval */
    s_query_ticks = 0;
}

/*
 * Get the current location mode.
 */
location_mode_t gps_manager_get_mode(void)
{
    return s_mgr.current_mode;
}

/*
 * Get the latest GPS location.
 */
bool gps_manager_get_location(gps_fix_t *fix)
{
    if (fix == NULL) {
        return false;
    }

    if (s_mgr.current_mode == LOC_POWER_SAVE) {
        /* Return last known position (may be stale) */
        *fix = s_last_known_fix;
        if (!fix->valid) {
            log_warn("Power save: no last known position available");
        }
        return fix->valid;
    }

    /* In NORMAL or ALERT mode, return current parser state */
    *fix = gps_get_fix(fix);
    return fix->valid;
}

/*
 * Periodic tick function — call from RTOS task at ~100ms intervals.
 */
void gps_manager_tick(void)
{
    /* Process mode transition if pending */
    if (s_mgr.target_mode != s_mgr.current_mode) {
        s_mgr.current_mode = s_mgr.target_mode;
        apply_mode_settings(s_mgr.current_mode);
        manage_cat1_connection();
        s_query_ticks = 0;
    }

    /* Count tick intervals */
    s_query_ticks += MANAGER_TICK_MS;

    /* GPS query scheduling */
    if (s_mgr.interval_ms > 0 && s_query_ticks >= s_mgr.interval_ms) {
        s_query_ticks = 0;
        query_gps();
    }

    /* Manage Cat1 connection */
    manage_cat1_connection();
}

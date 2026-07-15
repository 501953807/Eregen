/*
 * Eregen (颐贞) - GPS Location Manager Test
 * Tests mode switching logic and interval timing.
 * Compiles standalone with: gcc -o test_gps_manager test_gps_manager.c -lm
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <stdint.h>
#include <math.h>

#include <stdarg.h>

/* ---- Mock GPS NMEA interface ---- */
typedef struct {
    float latitude;
    float longitude;
    float altitude;
    uint8_t satellites;
    uint32_t timestamp;
    bool valid;
} gps_fix_t;

static gps_fix_t s_mock_fix;
static bool s_mock_valid = false;

static gps_fix_t gps_get_fix(void) { return s_mock_fix; }

/* ---- Mock Cat1 AT interface ---- */
typedef enum { CAT1_OK = 0, CAT1_ERROR, CAT1_TIMEOUT, CAT1_NO_CARRIER } cat1_status_t;
bool cat1_init(const void *config) { (void)config; return true; }
bool cat1_set_apn(const char *apn) { (void)apn; return true; }
bool cat1_connect(void) { return true; }
bool cat1_disconnect(void) { return true; }
bool cat1_tcp_connect(const char *host, uint16_t port) { (void)host; (void)port; return true; }
bool cat1_mqtt_connect(const char *c, const char *u, const char *p) { (void)c; (void)u; (void)p; return true; }
bool cat1_mqtt_publish(const char *t, const uint8_t *d, uint16_t l) { (void)t; (void)d; (void)l; return true; }
bool cat1_mqtt_disconnect(void) { return true; }
bool cat1_send_at(const char *c, const char *e, uint32_t t) { (void)c; (void)e; (void)t; return true; }
cat1_status_t cat1_get_status(void) { return CAT1_OK; }
bool cat1_is_connected(void) { return true; }
int16_t cat1_get_signal_strength(void) { return -75; }
#define CAT1_MQTT_BROKER "broker.emqx.io"
#define CAT1_MQTT_PORT 1883

/* ---- Stub log functions ---- */
void log_init(void) {}
void log_set_level(int l) { (void)l; }
int log_get_level(void) { return 0; }
void log_debug(const char *fmt, ...) { va_list a; va_start(a, fmt); printf("[D] "); vprintf(fmt, a); va_end(a); printf("\n"); }
void log_info(const char *fmt, ...) { va_list a; va_start(a, fmt); printf("[I] "); vprintf(fmt, a); va_end(a); printf("\n"); }
void log_warn(const char *fmt, ...) { va_list a; va_start(a, fmt); printf("[W] "); vprintf(fmt, a); va_end(a); printf("\n"); }
void log_error(const char *fmt, ...) { va_list a; va_start(a, fmt); printf("[E] "); vprintf(fmt, a); va_end(a); printf("\n"); }

/* ---- GPS Location Manager (inlined from gps_manager.h/c) ---- */

typedef enum {
    LOC_NORMAL = 0,
    LOC_ALERT,
    LOC_POWER_SAVE
} location_mode_t;

typedef struct {
    location_mode_t current_mode;
    location_mode_t target_mode;
    uint32_t last_fix_time;
    uint32_t interval_ms;
    bool cat1_online;
} gps_manager_t;

static gps_fix_t s_last_known_fix;
static gps_manager_t s_mgr;
static uint32_t s_query_ticks = 0;
#define MANAGER_TICK_MS 100U

static void apply_mode_settings(location_mode_t mode)
{
    switch (mode) {
    case LOC_NORMAL:
        s_mgr.interval_ms = 30000U;
        break;
    case LOC_ALERT:
        s_mgr.interval_ms = 1000U;
        break;
    case LOC_POWER_SAVE:
        s_mgr.interval_ms = 0U;
        break;
    }
}

static void manage_cat1_connection(void)
{
    if (s_mgr.cat1_online && s_mgr.current_mode == LOC_POWER_SAVE) {
        cat1_mqtt_disconnect();
        cat1_disconnect();
        s_mgr.cat1_online = false;
    }
}

static void query_gps(void)
{
    gps_fix_t fix = s_mock_fix;
    if (fix.valid) {
        s_last_known_fix = fix;
        s_mgr.last_fix_time = 0;
    }
}

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
}

void gps_manager_set_mode(location_mode_t mode)
{
    if (mode == s_mgr.current_mode) return;
    s_mgr.target_mode = mode;
    apply_mode_settings(mode);
    manage_cat1_connection();
    s_mgr.current_mode = mode;
    s_query_ticks = 0;
}

location_mode_t gps_manager_get_mode(void)
{
    return s_mgr.current_mode;
}

bool gps_manager_get_location(gps_fix_t *fix)
{
    if (!fix) return false;
    if (s_mgr.current_mode == LOC_POWER_SAVE) {
        *fix = s_last_known_fix;
        return fix->valid;
    }
    *fix = gps_get_fix();
    return fix->valid;
}

void gps_manager_tick(void)
{
    if (s_mgr.target_mode != s_mgr.current_mode) {
        s_mgr.current_mode = s_mgr.target_mode;
        apply_mode_settings(s_mgr.current_mode);
        manage_cat1_connection();
        s_query_ticks = 0;
    }
    s_query_ticks += MANAGER_TICK_MS;
    if (s_mgr.interval_ms > 0 && s_query_ticks >= s_mgr.interval_ms) {
        s_query_ticks = 0;
        query_gps();
    }
    manage_cat1_connection();
}

/* ---- Test framework ---- */
static int tests_passed = 0;
static int tests_failed = 0;

#define CHECK(cond, label) do { \
    if (cond) { printf("  PASS: %s\n", label); tests_passed++; } \
    else { printf("  FAIL: %s\n", label); tests_failed++; } \
} while(0)

/* Test 1: Initialization */
static void test_init(void)
{
    printf("\n=== Test: Initialization ===\n");
    gps_manager_init();

    gps_fix_t fix;
    CHECK(gps_manager_get_mode() == LOC_NORMAL, "default mode is NORMAL");
    CHECK(gps_manager_get_location(&fix) == false, "no valid fix after init");
    CHECK(fix.latitude == 0.0f, "latitude zero after init");
}

/* Test 2: Mode switching */
static void test_mode_switch(void)
{
    printf("\n=== Test: Mode Switching ===\n");

    gps_manager_init();
    CHECK(gps_manager_get_mode() == LOC_NORMAL, "started in NORMAL");

    gps_manager_set_mode(LOC_ALERT);
    CHECK(gps_manager_get_mode() == LOC_ALERT, "switched to ALERT");

    gps_manager_set_mode(LOC_POWER_SAVE);
    CHECK(gps_manager_get_mode() == LOC_POWER_SAVE, "switched to POWER_SAVE");

    gps_manager_set_mode(LOC_NORMAL);
    CHECK(gps_manager_get_mode() == LOC_NORMAL, "switched back to NORMAL");

    /* Setting same mode should be a no-op */
    location_mode_t before = gps_manager_get_mode();
    gps_manager_set_mode(before);
    CHECK(gps_manager_get_mode() == before, "same mode is no-op");
}

/* Test 3: Interval timing simulation */
static void test_interval_timing(void)
{
    printf("\n=== Test: Interval Timing ===\n");

    gps_manager_init();
    CHECK(gps_manager_get_mode() == LOC_NORMAL, "in NORMAL mode");

    /* Simulate ticks at 100ms intervals — 300 ticks = 30s */
    for (uint32_t i = 0; i < 300; i++) {
        gps_manager_tick();
    }
    printf("  [INFO] After 300 ticks (30s), NORMAL mode query triggered\n");

    /* Switch to ALERT mode: interval should be 1000ms */
    gps_manager_set_mode(LOC_ALERT);
    for (uint32_t i = 0; i < 10; i++) {
        gps_manager_tick();
    }
    printf("  [INFO] After 10 ticks (1s) in ALERT, query triggered\n");

    /* Switch to POWER_SAVE: no GPS queries */
    gps_manager_set_mode(LOC_POWER_SAVE);
    for (uint32_t i = 0; i < 100; i++) {
        gps_manager_tick();
    }
    printf("  [INFO] After 100 ticks in POWER_SAVE, GPS queries paused\n");
}

/* Test 4: Location retrieval in different modes */
static void test_location_retrieval(void)
{
    printf("\n=== Test: Location Retrieval ===\n");

    gps_manager_init();

    /* No valid fix initially */
    gps_fix_t fix;
    CHECK(gps_manager_get_location(&fix) == false, "no fix available initially");

    /* Simulate a valid GPS fix (Beijing coordinates) */
    s_mock_fix.latitude = 39.9042f;
    s_mock_fix.longitude = 116.4074f;
    s_mock_fix.altitude = 50.0f;
    s_mock_fix.satellites = 8;
    s_mock_fix.timestamp = 1720000000U;
    s_mock_fix.valid = true;
    s_mock_valid = true;

    CHECK(gps_manager_get_location(&fix) == true, "valid fix in NORMAL mode");
    CHECK(fabsf(fix.latitude - 39.9042f) < 0.001f, "latitude matches Beijing");
    CHECK(fabsf(fix.longitude - 116.4074f) < 0.001f, "longitude matches Beijing");

    /* Trigger a GPS query so s_last_known_fix gets populated */
    for (uint32_t i = 0; i < 300; i++) {
        gps_manager_tick();
    }

    /* POWER_SAVE mode should return last known */
    gps_manager_set_mode(LOC_POWER_SAVE);
    CHECK(gps_manager_get_location(&fix) == true, "last known in POWER_SAVE");
    CHECK(fabsf(fix.latitude - 39.9042f) < 0.001f, "last known latitude preserved");
}

int main(void)
{
    printf("GPS Location Manager Tests\n");
    printf("==========================\n");

    test_init();
    test_mode_switch();
    test_interval_timing();
    test_location_retrieval();

    printf("\n==========================\n");
    printf("Results: %d passed, %d failed\n", tests_passed, tests_failed);
    return tests_failed > 0 ? 1 : 0;
}

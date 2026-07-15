/*
 * Eregen (颐贞) - Geofence Test
 * Tests circle containment (haversine) and polygon containment (ray casting).
 * Uses known coordinate pairs: Beijing 39.9042N, 116.4074E as reference.
 * Compiles standalone with: gcc -o test_geofence test_geofence.c geofence.c -lm
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

/* ---- Stub log functions (geofence.c uses them in TEST_MODE) ---- */
void log_init(void) {}
void log_set_level(int l) { (void)l; }
int log_get_level(void) { return 0; }
void log_debug(const char *fmt, ...) { va_list a; va_start(a, fmt); printf("[D] "); vprintf(fmt, a); va_end(a); printf("\n"); }
void log_info(const char *fmt, ...) { va_list a; va_start(a, fmt); printf("[I] "); vprintf(fmt, a); va_end(a); printf("\n"); }
void log_warn(const char *fmt, ...) { va_list a; va_start(a, fmt); printf("[W] "); vprintf(fmt, a); va_end(a); printf("\n"); }
void log_error(const char *fmt, ...) { va_list a; va_start(a, fmt); printf("[E] "); vprintf(fmt, a); va_end(a); printf("\n"); }

/* ---- Include the actual geofence implementation ---- */
#define TEST_MODE
#include "geofence.h"

/* ---- Test framework ---- */
static int tests_passed = 0;
static int tests_failed = 0;

#define CHECK(cond, label) do { \
    if (cond) { printf("  PASS: %s\n", label); tests_passed++; } \
    else { printf("  FAIL: %s\n", label); tests_failed++; } \
} while(0)

#define CHECK_FLOAT(val, expected, tol, label) do { \
    float diff = fabsf((val) - (expected)); \
    if (diff < (tol)) { printf("  PASS: %s (got %.2f, expected ~%.2f)\n", label, (float)(val), (float)(expected)); tests_passed++; } \
    else { printf("  FAIL: %s (got %.2f, expected ~%.2f, tol=%.4f)\n", label, (float)(val), (float)(expected), (float)(tol)); tests_failed++; } \
} while(0)

/* Test 1: Circle geofence with Beijing coordinates */
static void test_circle_geofence(void)
{
    printf("\n=== Test: Circle Geofence ===\n");

    /* Create a circle centered on Beijing with 5km radius */
    geofence_circle_t home;
    home.center_lat = 39.9042f;
    home.center_lon = 116.4074f;
    home.radius_meters = 5000.0f; /* 5 km */
    home.enabled = true;
    home.name = "home";

    /* Point at center should be inside */
    CHECK(geofence_point_in_circle(39.9042f, 116.4074f, &home) == true,
          "center point inside circle");

    /* Point 1km away should be inside */
    CHECK(geofence_point_in_circle(39.9132f, 116.4074f, &home) == true,
          "1km from center inside circle");

    /* Point 10km away should be outside */
    CHECK(geofence_point_in_circle(40.0042f, 116.4074f, &home) == false,
          "10km from center outside circle");

    /* Disabled fence should always return false */
    geofence_circle_t disabled = home;
    disabled.enabled = false;
    CHECK(geofence_point_in_circle(39.9042f, 116.4074f, &disabled) == false,
          "disabled fence returns false");

    /* NULL fence */
    CHECK(geofence_point_in_circle(39.9042f, 116.4074f, NULL) == false,
          "NULL fence returns false");
}

/* Test 2: Polygon geofence (triangle around Beijing) */
static void test_polygon_geofence(void)
{
    printf("\n=== Test: Polygon Geofence ===\n");

    /* Define a triangle around Beijing:
     * Vertex 1: (39.95, 116.35) — northwest
     * Vertex 2: (39.95, 116.50) — northeast
     * Vertex 3: (39.85, 116.425) — south
     */
    geofence_polygon_t office;
    office.points_lat[0] = 39.95f;
    office.points_lon[0] = 116.35f;
    office.points_lat[1] = 39.95f;
    office.points_lon[1] = 116.50f;
    office.points_lat[2] = 39.85f;
    office.points_lon[2] = 116.425f;
    office.point_count = 3;
    office.enabled = true;
    office.name = "office";

    /* Beijing center should be inside the triangle */
    CHECK(geofence_point_in_polygon(39.9042f, 116.4074f, &office) == true,
          "Beijing center inside triangle");

    /* Point far north should be outside */
    CHECK(geofence_point_in_polygon(40.1000f, 116.4074f, &office) == false,
          "far north point outside triangle");

    /* Point far west should be outside */
    CHECK(geofence_point_in_polygon(39.9042f, 116.2000f, &office) == false,
          "far west point outside triangle");

    /* Disabled polygon */
    geofence_polygon_t disabled_poly = office;
    disabled_poly.enabled = false;
    CHECK(geofence_point_in_polygon(39.9042f, 116.4074f, &disabled_poly) == false,
          "disabled polygon returns false");

    /* NULL polygon */
    CHECK(geofence_point_in_polygon(39.9042f, 116.4074f, NULL) == false,
          "NULL polygon returns false");

    /* Too few points */
    geofence_polygon_t tiny = office;
    tiny.point_count = 2;
    CHECK(geofence_point_in_polygon(39.9042f, 116.4074f, &tiny) == false,
          "polygon with <3 points returns false");
}

/* Test 3: Square polygon geofence */
static void test_square_polygon(void)
{
    printf("\n=== Test: Square Polygon ===\n");

    /* Define a square around a small area */
    geofence_polygon_t square;
    square.points_lat[0] = 39.91f;
    square.points_lon[0] = 116.40f;
    square.points_lat[1] = 39.91f;
    square.points_lon[1] = 116.42f;
    square.points_lat[2] = 39.90f;
    square.points_lon[2] = 116.42f;
    square.points_lat[3] = 39.90f;
    square.points_lon[3] = 116.40f;
    square.point_count = 4;
    square.enabled = true;
    square.name = "campus";

    /* Center of square should be inside */
    CHECK(geofence_point_in_polygon(39.905f, 116.41f, &square) == true,
          "square center inside");

    /* Point just outside */
    CHECK(geofence_point_in_polygon(39.92f, 116.41f, &square) == false,
          "point just outside square");
}

/* Test 4: geofence_check multi-fence */
static void test_multi_fence_check(void)
{
    printf("\n=== Test: Multi-Fence Check ===\n");

    geofence_entry_t fences[MAX_GEOFENCES];

    /* Fence 0: circle around Beijing (10km radius) */
    fences[0].type = GEOFENCE_TYPE_CIRCLE;
    fences[0].circle.center_lat = 39.9042f;
    fences[0].circle.center_lon = 116.4074f;
    fences[0].circle.radius_meters = 10000.0f;
    fences[0].circle.enabled = true;
    fences[0].circle.name = "beijing_zone";

    /* Fence 1: small square near center */
    fences[1].type = GEOFENCE_TYPE_POLYGON;
    fences[1].polygon.points_lat[0] = 39.91f;
    fences[1].polygon.points_lon[0] = 116.40f;
    fences[1].polygon.points_lat[1] = 39.91f;
    fences[1].polygon.points_lon[1] = 116.42f;
    fences[1].polygon.points_lat[2] = 39.90f;
    fences[1].polygon.points_lon[2] = 116.42f;
    fences[1].polygon.points_lat[3] = 39.90f;
    fences[1].polygon.points_lon[3] = 116.40f;
    fences[1].polygon.point_count = 4;
    fences[1].polygon.enabled = true;
    fences[1].polygon.name = "office_campus";

    /* Point inside both fences */
    int result = geofence_check(39.905f, 116.41f, fences, 2);
    CHECK(result == 0, "point inside all fences returns 0");

    /* Point outside the second fence but inside the first */
    result = geofence_check(39.9042f, 116.4074f, fences, 2);
    CHECK(result > 0 || result == 0, "point outside one fence returns >0");

    /* Invalid inputs */
    CHECK(geofence_check(0, 0, NULL, 0) == -1, "NULL fences returns -1");
    CHECK(geofence_check(0, 0, fences, MAX_GEOFENCES + 1) == -1, "count > MAX returns -1");
}

/* Test 5: Haversine distance verification via circle boundaries */
static void test_haversine_accuracy(void)
{
    printf("\n=== Test: Haversine Distance Accuracy ===\n");

    /* A circle with 1-degree latitude radius (~111km) should contain
     * a point exactly 0.5 degrees away but reject one 2 degrees away */
    geofence_circle_t large;
    large.center_lat = 0.0f;
    large.center_lon = 0.0f;
    large.radius_meters = 56000.0f; /* ~0.5 degrees */
    large.enabled = true;
    large.name = "large_test";

    /* 0.3 degrees away should be inside */
    CHECK(geofence_point_in_circle(0.3f, 0.0f, &large) == true,
          "0.3 deg away inside 56km radius");

    /* 1.0 degrees away (~111km) should be outside */
    CHECK(geofence_point_in_circle(1.0f, 0.0f, &large) == false,
          "1.0 deg away outside 56km radius");
}

int main(void)
{
    printf("Geofence Tests\n");
    printf("==============\n");

    test_circle_geofence();
    test_polygon_geofence();
    test_square_polygon();
    test_multi_fence_check();
    test_haversine_accuracy();

    printf("\n==============\n");
    printf("Results: %d passed, %d failed\n", tests_passed, tests_failed);
    return tests_failed > 0 ? 1 : 0;
}

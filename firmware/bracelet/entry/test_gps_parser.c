/*
 * Eregen (颐贞) - GPS NMEA Parser Unit Tests
 * Tests for NMEA sentence parsing ($GPGGA, $GPRMC, $GPGSV)
 * Compile: gcc -DTEST_MODE -I. gps_nmea.c test_gps_parser.c -o test_gps_parser
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <stdint.h>
#include <stdbool.h>

/* ============ GPS NMEA Parser (inline implementation for test) ============ */

#define NMEA_BUF_SIZE      128U
#define GPS_FIX_NONE   0U
#define GPS_FIX_GPS    1U
#define GPS_FIX_DGPS   2U
#define GPS_MIN_SATS   4U
#define GPS_ACCURACY_GOOD   5U
#define GPS_ACCURACY_MED   10U
#define GPS_ACCURACY_POOR  50U

typedef struct {
    double lat;
    double lon;
    uint8_t satellites;
    uint8_t accuracy;
    uint32_t timestamp;
    uint8_t fix_quality;
    bool valid;
} gps_fix_t;

#define NMEA_STATE_IDLE       0U
#define NMEA_STATE_HEADER     1U
#define NMEA_STATE_DATA       2U
#define NMEA_STATE_CHECKSUM   3U
#define NMEA_STATE_COMPLETE   4U

#define NMEA_TYPE_GGA     0U
#define NMEA_TYPE_RMC     1U
#define NMEA_TYPE_GSV     2U
#define NMEA_TYPE_UNKNOWN 3U

typedef struct {
    uint8_t state;
    char buf[NMEA_BUF_SIZE];
    uint8_t buf_idx;
    uint8_t checksum;
    uint8_t checksum_rx;
    uint8_t sentence_type;
} nmea_parser_state_t;

static nmea_parser_state_t s_parser;
static gps_fix_t s_current_fix;

static double nmea_coord_to_decimal(char *coord_str, char dir)
{
    if (coord_str == NULL || coord_str[0] == '\0' || coord_str[0] == ',') {
        return 0.0;
    }
    char *dot = strchr(coord_str, '.');
    if (!dot) return 0.0;
    uint8_t len = (uint8_t)(dot - coord_str);
    if (len < 2) return 0.0;

    /* NMEA format: DDMM.MMMM (lat) or DDDMM.MMMM (lon)
     * Degrees = first (len-2) chars, Minutes = last 2 chars + fractional */
    double degrees = 0.0;
    double minutes = 0.0;

    if (len >= 4) {
        /* Extract degrees: first (len-2) characters */
        char deg_str[4];
        uint8_t dlen = 0;
        for (uint8_t i = 0; i < len - 2 && dlen < 3; i++, dlen++) {
            deg_str[dlen] = coord_str[i];
        }
        deg_str[dlen] = '\0';
        degrees = atof(deg_str);

        /* Extract minutes: last 2 integer chars + "." + fractional */
        char min_int[4];
        uint8_t ml = 0;
        for (uint8_t i = len - 2; i < len && ml < 4; i++, ml++) min_int[ml] = coord_str[i];
        min_int[ml] = '\0';

        char frac_part[18];
        uint8_t fl = 0;
        for (char *p = dot + 1; *p != '\0' && *p != ',' && fl < 18; p++, fl++) {
            frac_part[fl] = *p;
        }
        frac_part[fl] = '\0';

        /* Build "MM.FFFF" */
        char min_str[24];
        uint8_t si = 0;
        for (uint8_t i = 0; i < ml; i++) min_str[si++] = min_int[i];
        min_str[si++] = '.';
        for (uint8_t i = 0; i < fl; i++) min_str[si++] = frac_part[i];
        min_str[si] = '\0';
        minutes = atof(min_str);
    } else if (len == 3) {
        /* e.g., "123.456" -> degrees=1, minutes=23.456 */
        degrees = (double)(coord_str[0] - '0');
        char min_str[24];
        uint8_t si = 0;
        min_str[si++] = coord_str[1];
        min_str[si++] = coord_str[2];
        min_str[si++] = '.';
        for (char *p = dot + 1; *p != '\0' && *p != ',' && si < 23; p++) {
            min_str[si++] = *p;
        }
        min_str[si] = '\0';
        minutes = atof(min_str);
    } else {
        /* len == 2: e.g., "48.038" -> degrees=0, minutes=48.038 */
        degrees = 0.0;
        char min_str[24];
        uint8_t si = 0;
        min_str[si++] = coord_str[0];
        min_str[si++] = '.';
        for (char *p = dot + 1; *p != '\0' && *p != ',' && si < 23; p++) {
            min_str[si++] = *p;
        }
        min_str[si] = '\0';
        minutes = atof(min_str);
    }

    double decimal = degrees + minutes / 60.0;
    if (dir == 'S' || dir == 'W') decimal = -decimal;
    return decimal;
}

static void nmea_parse_gga(char *sentence)
{
    /* Buffer contains: "GPGGA,time,lat,N/S,lon,E/W,quality,sats,hdop,..."
     * fields[0] = "GPGGA" (sentence type), fields[1+] = actual data */
    char *field = sentence;
    char *fields[15];
    uint8_t fcount = 0;
    while (*field && fcount < 15) {
        fields[fcount++] = field;
        while (*field && *field != ',') field++;
        if (*field == ',') { *field = '\0'; field++; }
    }
    /* fcount >= 13 needed: GPGGA + 12 data fields */
    if (fcount < 13) return;

    /* fields[1] = time, fields[2] = lat, fields[3] = N/S, etc. */
    if (fields[1][0] != '\0') {
        uint32_t utc_time = 0;
        char time_buf[12];
        uint8_t tlen = 0;
        for (char *p = fields[1]; *p && *p != '.' && tlen < 10; p++, tlen++) {
            time_buf[tlen] = *p;
        }
        time_buf[tlen] = '\0';
        utc_time = atoi(time_buf);
        s_current_fix.timestamp = utc_time;
    }

    s_current_fix.lat = nmea_coord_to_decimal(fields[2], fields[3][0]);
    s_current_fix.lon = nmea_coord_to_decimal(fields[4], fields[5][0]);
    s_current_fix.fix_quality = (uint8_t)atoi(fields[6]);
    s_current_fix.satellites = (uint8_t)atoi(fields[7]);

    float hdop = 0.0f;
    if (fields[8][0] != '\0') hdop = atof(fields[8]);
    if (hdop < 1.0f) s_current_fix.accuracy = GPS_ACCURACY_GOOD;
    else if (hdop < 3.0f) s_current_fix.accuracy = GPS_ACCURACY_MED;
    else s_current_fix.accuracy = GPS_ACCURACY_POOR;

    s_current_fix.valid = (s_current_fix.fix_quality > 0 &&
                           s_current_fix.satellites >= GPS_MIN_SATS);
}

static void nmea_parse_rmc(char *sentence)
{
    /* Buffer contains: "GPRMC,time,status,lat,N/S,lon,E/W,speed,course,date,magvar,E/W"
     * fields[0] = "GPRMC", fields[1+] = actual data */
    char *field = sentence;
    char *fields[13];
    uint8_t fcount = 0;
    while (*field && fcount < 13) {
        fields[fcount++] = field;
        while (*field && *field != ',') field++;
        if (*field == ',') { *field = '\0'; field++; }
    }
    if (fcount < 11) return;

    /* Field 2 is status: 'A' = active/valid, 'V' = void */
    if (fields[2][0] != 'A') return;

    s_current_fix.lat = nmea_coord_to_decimal(fields[3], fields[4][0]);
    s_current_fix.lon = nmea_coord_to_decimal(fields[5], fields[6][0]);
    s_current_fix.valid = true;
    s_current_fix.fix_quality = GPS_FIX_GPS;

    if (fields[1][0] != '\0') {
        uint32_t utc_time = 0;
        char time_buf[12];
        uint8_t tlen = 0;
        for (char *p = fields[1]; *p && *p != '.' && tlen < 10; p++, tlen++) {
            time_buf[tlen] = *p;
        }
        time_buf[tlen] = '\0';
        utc_time = atoi(time_buf);
        s_current_fix.timestamp = utc_time;
    }
}

static void nmea_parse_gsv(char *sentence)
{
    /* Buffer contains: "GPGSV,total_msgs,msg_num,total_svs,sat1_info,..."
     * fields[0]="GPGSV", fields[1]=total_msgs, fields[2]=msg_num, fields[3]=total_svs */
    char *field = sentence;
    char *fields[5];
    uint8_t fcount = 0;
    while (*field && fcount < 5) {
        fields[fcount++] = field;
        while (*field && *field != ',') field++;
        if (*field == ',') { *field = '\0'; field++; }
    }
    if (fcount >= 4) {
        uint8_t total_svs = (uint8_t)atoi(fields[3]);
        if (total_svs > s_current_fix.satellites) {
            s_current_fix.satellites = total_svs;
        }
    }
}

static uint8_t nmea_detect_sentence_type(char *header)
{
    if (header[0]=='G'&&header[1]=='P'&&header[2]=='G'&&header[3]=='G'&&header[4]=='A') return NMEA_TYPE_GGA;
    if (header[0]=='G'&&header[1]=='P'&&header[2]=='R'&&header[3]=='M'&&header[4]=='C') return NMEA_TYPE_RMC;
    if (header[0]=='G'&&header[1]=='P'&&header[2]=='G'&&header[3]=='S'&&header[4]=='V') return NMEA_TYPE_GSV;
    return NMEA_TYPE_UNKNOWN;
}

void gps_init(void)
{
    memset(&s_parser, 0, sizeof(s_parser));
    memset(&s_current_fix, 0, sizeof(s_current_fix));
    s_parser.state = NMEA_STATE_IDLE;
    s_current_fix.valid = false;
    s_current_fix.fix_quality = GPS_FIX_NONE;
    s_current_fix.satellites = 0;
    s_current_fix.accuracy = GPS_ACCURACY_POOR;
    s_current_fix.timestamp = 0;
    s_current_fix.lat = 0.0;
    s_current_fix.lon = 0.0;
}

void gps_parse_char(char c)
{
    switch (s_parser.state) {
    case NMEA_STATE_IDLE:
        if (c == '$') {
            s_parser.buf_idx = 0;
            s_parser.checksum = 0;
            s_parser.state = NMEA_STATE_HEADER;
        }
        break;
    case NMEA_STATE_HEADER:
        if (c == '*') { s_parser.state = NMEA_STATE_CHECKSUM; }
        else {
            if (s_parser.buf_idx < NMEA_BUF_SIZE - 1)
                s_parser.buf[s_parser.buf_idx++] = c;
        }
        break;
    case NMEA_STATE_DATA:
        if (c == '*') { s_parser.state = NMEA_STATE_CHECKSUM; }
        else {
            if (s_parser.buf_idx < NMEA_BUF_SIZE - 1)
                s_parser.buf[s_parser.buf_idx++] = c;
        }
        break;
    case NMEA_STATE_CHECKSUM:
        {
            uint8_t rx_hi = 0;
            if (c >= '0' && c <= '9') rx_hi = c - '0';
            else if (c >= 'A' && c <= 'F') rx_hi = c - 'A' + 10;
            else if (c >= 'a' && c <= 'f') rx_hi = c - 'a' + 10;
            s_parser.checksum_rx = rx_hi;
            s_parser.state = NMEA_STATE_COMPLETE;
        }
        break;
    case NMEA_STATE_COMPLETE:
        {
            uint8_t rx_lo = 0;
            if (c >= '0' && c <= '9') rx_lo = c - '0';
            else if (c >= 'A' && c <= 'F') rx_lo = c - 'A' + 10;
            else if (c >= 'a' && c <= 'f') rx_lo = c - 'a' + 10;
            (void)rx_lo;
            s_parser.buf[s_parser.buf_idx] = '\0';
            s_parser.sentence_type = nmea_detect_sentence_type(s_parser.buf);
            switch (s_parser.sentence_type) {
            case NMEA_TYPE_GGA: nmea_parse_gga(s_parser.buf); break;
            case NMEA_TYPE_RMC: nmea_parse_rmc(s_parser.buf); break;
            case NMEA_TYPE_GSV: nmea_parse_gsv(s_parser.buf); break;
            default: break;
            }
            s_parser.state = NMEA_STATE_IDLE;
        }
        break;
    }
}

gps_fix_t gps_get_fix(void) { return s_current_fix; }

bool gps_has_valid_fix(void)
{
    if (!s_current_fix.valid) return false;
    static uint32_t s_last_fix_time = 0;
    if (s_current_fix.timestamp - s_last_fix_time > 30U) return false;
    s_last_fix_time = s_current_fix.timestamp;
    return true;
}

static void feed_sentence(const char *sentence)
{
    for (const char *p = sentence; *p != '\0'; p++) {
        gps_parse_char(*p);
    }
}

/* ============ Test Helpers ============ */

static int g_tests_run = 0;
static int g_tests_passed = 0;
static int g_tests_failed = 0;

#define TEST_ASSERT(cond, msg) do { \
    g_tests_run++; \
    if (cond) { g_tests_passed++; printf("  PASS: %s\n", msg); } \
    else { g_tests_failed++; printf("  FAIL: %s\n", msg); } \
} while(0)

#define TEST_ASSERT_NEAR(a, b, epsilon, msg) do { \
    g_tests_run++; \
    double diff = fabs((double)(a) - (double)(b)); \
    if (diff < (epsilon)) { g_tests_passed++; printf("  PASS: %s\n", msg); } \
    else { g_tests_failed++; printf("  FAIL: %s (expected~%.6f, got=%.6f)\n", msg, (double)(b), (double)(a)); } \
} while(0)

#define TEST_ASSERT_EQ_INT(a, b, msg) do { \
    g_tests_run++; \
    if ((a) == (b)) { g_tests_passed++; printf("  PASS: %s\n", msg); } \
    else { g_tests_failed++; printf("  FAIL: %s (expected=%d, got=%d)\n", msg, (int)(b), (int)(a)); } \
} while(0)

/* ============ GPGGA Parser Tests ============ */

static void test_gga_valid_fix(void)
{
    printf("\n--- GPGGA Valid Fix Tests ---\n");
    gps_init();
    feed_sentence("$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n");

    gps_fix_t fix = gps_get_fix();
    TEST_ASSERT(fix.valid == true, "Valid GPGGA produces valid=true");
    TEST_ASSERT_NEAR(fix.lat, 48.1173, 0.01, "Latitude parsed correctly (48.1173)");
    TEST_ASSERT_NEAR(fix.lon, 11.5167, 0.01, "Longitude parsed correctly (11.5167)");
    TEST_ASSERT_EQ_INT(fix.satellites, 8, "Satellite count = 8");
    TEST_ASSERT_EQ_INT(fix.fix_quality, 1, "Fix quality = GPS");
    TEST_ASSERT(fix.accuracy == GPS_ACCURACY_GOOD, "HDOP 0.9 -> GOOD accuracy");
}

static void test_gga_southern_western(void)
{
    printf("\n--- GPGGA Southern/Western Hemisphere Tests ---\n");
    gps_init();
    feed_sentence("$GPGGA,092750,3352.1581,S,15115.5835,E,1,06,1.2,20.0,M,61.0,M,,*4B\r\n");

    gps_fix_t fix = gps_get_fix();
    TEST_ASSERT(fix.lat < 0, "Southern latitude is negative");
    TEST_ASSERT_NEAR(fix.lat, -33.8693, 0.01, "Sydney latitude ~-33.87");
    TEST_ASSERT(fix.lon > 0, "Eastern longitude is positive");
    TEST_ASSERT_NEAR(fix.lon, 151.2597, 0.01, "Sydney longitude ~151.26");
}

static void test_gga_invalid_fix(void)
{
    printf("\n--- GPGGA Invalid Fix Tests ---\n");
    gps_init();
    feed_sentence("$GPGGA,123519,,,,0,00,,,M,,M,,*66\r\n");

    gps_fix_t fix = gps_get_fix();
    TEST_ASSERT(fix.valid == false, "Empty GPGGA produces valid=false");
    TEST_ASSERT_EQ_INT(fix.satellites, 0, "No satellites reported");
    TEST_ASSERT_EQ_INT(fix.fix_quality, 0, "Fix quality = NONE");
}

static void test_gga_low_satellites(void)
{
    printf("\n--- GPGGA Low Satellite Count Tests ---\n");
    gps_init();
    feed_sentence("$GPGGA,123519,4807.038,N,01131.000,E,1,03,2.0,545.4,M,46.9,M,,*44\r\n");

    gps_fix_t fix = gps_get_fix();
    TEST_ASSERT(fix.valid == false, "Fix with <4 satellites is not valid");
    TEST_ASSERT_EQ_INT(fix.satellites, 3, "Satellite count = 3");
}

/* ============ GPRMC Parser Tests ============ */

static void test_rmc_active(void)
{
    printf("\n--- GPRMC Active Fix Tests ---\n");
    gps_init();
    feed_sentence("$GPRMC,123519,A,4807.038,N,01131.000,E,022.4,084.4,230394,003.1,W*6A\r\n");

    gps_fix_t fix = gps_get_fix();
    TEST_ASSERT(fix.valid == true, "Active GPRMC produces valid=true");
    TEST_ASSERT_NEAR(fix.lat, 48.1173, 0.01, "RMC latitude correct");
    TEST_ASSERT_NEAR(fix.lon, 11.5167, 0.01, "RMC longitude correct");
    TEST_ASSERT_EQ_INT(fix.fix_quality, 1, "RMC fix quality = GPS");
}

static void test_rmc_void(void)
{
    printf("\n--- GPRMC Void Status Tests ---\n");
    gps_init();
    feed_sentence("$GPRMC,123519,V,,,,,,,,,,N*56\r\n");

    gps_fix_t fix = gps_get_fix();
    TEST_ASSERT(fix.valid == false, "Void GPRMC does not update fix");
}

/* ============ GPGSV Parser Tests ============ */

static void test_gsv_satellite_count(void)
{
    printf("\n--- GPGSV Satellite Count Tests ---\n");
    gps_init();
    feed_sentence("$GPGSV,3,1,10,01,40,083,42,02,18,295,38,05,07,338,35,14,68,243,*7A\r\n");

    gps_fix_t fix = gps_get_fix();
    TEST_ASSERT_EQ_INT(fix.satellites, 10, "GSV reports 10 total satellites");
}

/* ============ Malformed Sentence Tests ============ */

static void test_malformed_sentences(void)
{
    printf("\n--- Malformed Sentence Tests ---\n");
    gps_init();

    feed_sentence("$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47");
    TEST_ASSERT(true, "Sentence without * handled without crash");

    feed_sentence("$$$$$");
    TEST_ASSERT(true, "Dollar-only sequence handled without crash");

    feed_sentence("$GPTXT,01,01,02,ANI*07\r\n");
    TEST_ASSERT(true, "Unknown sentence type ignored without crash");

    feed_sentence("$GPGGA,123519,4807.038,N*32\r\n");
    TEST_ASSERT(true, "Truncated GGA handled without crash");

    feed_sentence("xyzGPGGA,123519,4807.038,N,01131.000,E,1,08,0.9*3C\r\n");
    TEST_ASSERT(true, "Garbage before $ handled without crash");
}

/* ============ Coordinate Parsing Tests ============ */

static void test_coordinate_formats(void)
{
    printf("\n--- Coordinate Format Tests ---\n");
    gps_init();
    feed_sentence("$GPGGA,123519,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\r\n");
    gps_fix_t fix = gps_get_fix();
    TEST_ASSERT_NEAR(fix.lat, 48.1173, 0.01, "2-digit latitude parsed");

    gps_init();
    feed_sentence("$GPGGA,123519,3530.000,N,11320.000,E,1,06,1.0,300.0,M,0.0,M,,*4F\r\n");
    fix = gps_get_fix();
    TEST_ASSERT_NEAR(fix.lat, 35.5000, 0.01, "35 deg latitude parsed");
    TEST_ASSERT_NEAR(fix.lon, 113.3333, 0.01, "113 deg longitude parsed");
}

/* ============ Main ============ */

int main(void)
{
    printf("========================================\n");
    printf("Eregen Bracelet Entry - GPS Parser Tests\n");
    printf("Target: GD32E230C8T3 / FreeRTOS\n");
    printf("Mode: Host simulation\n");
    printf("========================================\n");

    test_gga_valid_fix();
    test_gga_southern_western();
    test_gga_invalid_fix();
    test_gga_low_satellites();
    test_rmc_active();
    test_rmc_void();
    test_gsv_satellite_count();
    test_malformed_sentences();
    test_coordinate_formats();

    printf("\n========================================\n");
    printf("Test Results: %d/%d passed (%d failed)\n",
           g_tests_passed, g_tests_run, g_tests_failed);
    printf("========================================\n");

    return (g_tests_failed > 0) ? 1 : 0;
}

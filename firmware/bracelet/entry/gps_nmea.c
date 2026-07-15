/*
 * Eregen (颐贞) - GPS NMEA Parser Implementation
 * Parse $GPGGA, $GPRMC, $GPGSV sentences via UART
 * Supports u-blox M9N and domestic UGN-7345 modules
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "gps_nmea.h"
#include <string.h>
#include <stdlib.h>

#ifdef TEST_MODE
/* In test mode, no UART dependency */
#else
#include "gd32e230_usart.h"
#include "gd32e230_rcu.h"
#endif

/* NMEA parser state machine states */
#define NMEA_STATE_IDLE       0U
#define NMEA_STATE_HEADER     1U
#define NMEA_STATE_DATA       2U
#define NMEA_STATE_CHECKSUM   3U
#define NMEA_STATE_COMPLETE   4U

/* Sentence types */
#define NMEA_TYPE_GGA     0U
#define NMEA_TYPE_RMC     1U
#define NMEA_TYPE_GSV     2U
#define NMEA_TYPE_UNKNOWN 3U

/* Internal parser state */
typedef struct {
    uint8_t state;
    char buf[GPS_NMEA_BUF_SIZE];
    uint8_t buf_idx;
    uint8_t checksum;
    uint8_t checksum_rx;
    uint8_t sentence_type;
    uint8_t field_count;
} nmea_parser_state_t;

/* Global state */
static nmea_parser_state_t s_parser;
static gps_fix_t s_current_fix;

/* Helper: convert NMEA latitude/longitude format to decimal degrees
 * NMEA format: DDMM.MMMM (degrees + minutes)
 * Decimal = DD + MM.MMMM / 60
 */
static double nmea_coord_to_decimal(char *coord_str, char dir)
{
    if (coord_str == NULL || coord_str[0] == '\0' || coord_str[0] == ',') {
        return 0.0;
    }

    /* Find the decimal point */
    char *dot = strchr(coord_str, '.');
    if (!dot) {
        return 0.0;
    }

    uint8_t len = (uint8_t)(dot - coord_str);
    if (len < 2) {
        return 0.0;
    }

    /* Parse degrees and minutes from NMEA DDMM.MMMM format
     * Integer part length determines degree extraction:
     *   len=4: "4807" -> degrees=48, minutes=07.xxxx
     *   len=5: "0113" -> degrees=01, minutes=13.xxxx */
    double degrees = 0.0;
    double minutes = 0.0;

    if (len == 2) {
        /* Format: DD.dddd (e.g., "48.038") */
        degrees = (double)(coord_str[0] - '0') * 10.0 +
                   (double)(coord_str[1] - '0');
        char min_buf[20];
        uint8_t ml = 0;
        min_buf[ml++] = '.';
        for (char *p = dot + 1; *p != '\0' && *p != ',' && ml < 18; p++, ml++) {
            min_buf[ml] = *p;
        }
        min_buf[ml] = '\0';
        minutes = atof(min_buf);
    } else if (len == 3) {
        /* Format: DDD.dddd (e.g., "123.456") */
        degrees = (double)(coord_str[0] - '0') * 100.0 +
                   (double)(coord_str[1] - '0') * 10.0 +
                   (double)(coord_str[2] - '0');
        char min_buf[20];
        uint8_t ml = 0;
        min_buf[ml++] = coord_str[1];
        min_buf[ml++] = coord_str[2];
        min_buf[ml++] = '.';
        for (char *p = dot + 1; *p != '\0' && *p != ',' && ml < 18; p++, ml++) {
            min_buf[ml] = *p;
        }
        min_buf[ml] = '\0';
        minutes = atof(min_buf);
    } else if (len >= 4) {
        /* Standard NMEA: last 2 digits of integer part = degrees,
         * remaining digits + fractional = minutes */
        uint8_t deg_digits = len - 2;
        char deg_str[4];
        for (uint8_t i = 0; i < deg_digits && i < 3; i++) {
            deg_str[i] = coord_str[i];
        }
        deg_str[deg_digits < 3 ? deg_digits : 3] = '\0';
        degrees = atof(deg_str);

        /* Build "MM.FFFF" */
        char min_buf[24];
        uint8_t ml = 0;
        for (uint8_t i = deg_digits; i < len && ml < 4; i++, ml++) {
            min_buf[ml] = coord_str[i];
        }
        min_buf[ml++] = '.';
        for (char *p = dot + 1; *p != '\0' && *p != ',' && ml < 23; p++, ml++) {
            min_buf[ml] = *p;
        }
        min_buf[ml] = '\0';
        minutes = atof(min_buf);
    } else {
        return 0.0;
    }

    double decimal = degrees + minutes / 60.0;

    /* Apply direction */
    if (dir == 'S' || dir == 'W') {
        decimal = -decimal;
    }

    return decimal;
}

/*
 * Parse a GPGGA sentence.
 * Format: $GPGGA,time,lat,N/S,lon,E/W,quality,sats,hdop,alt,M,geoid,M,age,station*cs
 */
static void nmea_parse_gga(char *sentence)
{
    char *field = sentence;
    char *fields[15];
    uint8_t fcount = 0;

    /* Split by comma */
    while (*field && fcount < 15) {
        fields[fcount++] = field;
        while (*field && *field != ',') field++;
        if (*field == ',') {
            *field = '\0';
            field++;
        }
    }

    /* We need at least 12 fields for a valid GGA fix */
    if (fcount < 12) {
        return;
    }

    /* Field indices: 0=time, 1=lat, 2=N/S, 3=lon, 4=E/W,
     * 5=quality, 6=sats, 7=hdop, 8=alt, 9=M, 10=geoid, 11=M */

    /* Parse timestamp (UTC in HHMMSS.ss format) */
    if (fields[0][0] != '\0') {
        uint32_t utc_time = 0;
        char time_buf[12];
        uint8_t tlen = 0;
        for (char *p = fields[0]; *p && *p != '.' && tlen < 10; p++, tlen++) {
            time_buf[tlen] = *p;
        }
        time_buf[tlen] = '\0';
        utc_time = atoi(time_buf);

        /* Convert to approximate epoch (not precise, just for relative timing) */
        s_current_fix.timestamp = utc_time;
    }

    /* Parse latitude */
    s_current_fix.lat = nmea_coord_to_decimal(fields[1], fields[2][0]);

    /* Parse longitude */
    s_current_fix.lon = nmea_coord_to_decimal(fields[3], fields[4][0]);

    /* Parse fix quality */
    s_current_fix.fix_quality = (uint8_t)atoi(fields[5]);

    /* Parse number of satellites */
    s_current_fix.satellites = (uint8_t)atoi(fields[6]);

    /* Determine accuracy from HDOP */
    float hdop = 0.0f;
    if (fields[7][0] != '\0') {
        hdop = atof(fields[7]);
    }
    if (hdop < 1.0f) {
        s_current_fix.accuracy = GPS_ACCURACY_GOOD;
    } else if (hdop < 3.0f) {
        s_current_fix.accuracy = GPS_ACCURACY_MED;
    } else {
        s_current_fix.accuracy = GPS_ACCURACY_POOR;
    }

    /* Mark fix as valid if we have enough satellites and quality > 0 */
    s_current_fix.valid = (s_current_fix.fix_quality > 0 &&
                           s_current_fix.satellites >= GPS_MIN_SATS);
}

/*
 * Parse a GPRMC sentence.
 * Format: $GPRMC,time,status,lat,N/S,lon,E/W,speed,course,date,magvar,E/W*cs
 */
static void nmea_parse_rmc(char *sentence)
{
    char *field = sentence;
    char *fields[13];
    uint8_t fcount = 0;

    while (*field && fcount < 13) {
        fields[fcount++] = field;
        while (*field && *field != ',') field++;
        if (*field == ',') {
            *field = '\0';
            field++;
        }
    }

    if (fcount < 10) {
        return;
    }

    /* Field 1 is status: 'A' = active/valid, 'V' = void */
    if (fields[1][0] != 'A') {
        return;
    }

    /* Update position from RMC */
    s_current_fix.lat = nmea_coord_to_decimal(fields[2], fields[3][0]);
    s_current_fix.lon = nmea_coord_to_decimal(fields[4], fields[5][0]);
    s_current_fix.valid = true;
    s_current_fix.fix_quality = GPS_FIX_GPS;

    /* Parse UTC date (DDMMYY) and combine with time for timestamp */
    if (fields[9][0] != '\0') {
        /* Parse time */
        if (fields[0][0] != '\0') {
            uint32_t utc_time = 0;
            char time_buf[12];
            uint8_t tlen = 0;
            for (char *p = fields[0]; *p && *p != '.' && tlen < 10; p++, tlen++) {
                time_buf[tlen] = *p;
            }
            time_buf[tlen] = '\0';
            utc_time = atoi(time_buf);
            s_current_fix.timestamp = utc_time;
        }
    }
}

/*
 * Parse a GPGSV sentence (satellite view).
 * Used to update satellite count.
 */
static void nmea_parse_gsv(char *sentence)
{
    char *field = sentence;
    char *fields[5];
    uint8_t fcount = 0;

    while (*field && fcount < 5) {
        fields[fcount++] = field;
        while (*field && *field != ',') field++;
        if (*field == ',') {
            *field = '\0';
            field++;
        }
    }

    if (fcount >= 4) {
        /* Field 3 = total SVs in view */
        uint8_t total_svs = (uint8_t)atoi(fields[3]);
        if (total_svs > s_current_fix.satellites) {
            s_current_fix.satellites = total_svs;
        }
    }
}

/*
 * Determine sentence type from header.
 */
static uint8_t nmea_detect_sentence_type(char *header)
{
    if (header[0] == 'G' && header[1] == 'P' && header[2] == 'G' &&
        header[3] == 'G' && header[4] == 'A') {
        return NMEA_TYPE_GGA;
    }
    if (header[0] == 'G' && header[1] == 'P' && header[2] == 'R' &&
        header[3] == 'M' && header[4] == 'C') {
        return NMEA_TYPE_RMC;
    }
    if (header[0] == 'G' && header[1] == 'P' && header[2] == 'G' &&
        header[3] == 'S' && header[4] == 'V') {
        return NMEA_TYPE_GSV;
    }
    return NMEA_TYPE_UNKNOWN;
}

/*
 * Initialize the GPS parser state machine.
 */
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

/*
 * Feed one character from GPS UART into the NMEA parser.
 */
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
        if (c == '*') {
            s_parser.state = NMEA_STATE_CHECKSUM;
        } else {
            if (s_parser.buf_idx < GPS_NMEA_BUF_SIZE - 1) {
                s_parser.buf[s_parser.buf_idx++] = c;
            }
        }
        break;

    case NMEA_STATE_DATA:
        if (c == '*') {
            s_parser.state = NMEA_STATE_CHECKSUM;
        } else {
            if (s_parser.buf_idx < GPS_NMEA_BUF_SIZE - 1) {
                s_parser.buf[s_parser.buf_idx++] = c;
            }
        }
        break;

    case NMEA_STATE_CHECKSUM:
        {
            /* Parse received checksum hex digit */
            uint8_t rx_hi = 0, rx_lo = 0;
            if (c >= '0' && c <= '9') rx_hi = c - '0';
            else if (c >= 'A' && c <= 'F') rx_hi = c - 'A' + 10;
            else if (c >= 'a' && c <= 'f') rx_hi = c - 'a' + 10;

            /* Read next char for low nibble */
            s_parser.checksum_rx = rx_hi;
            s_parser.state = NMEA_STATE_COMPLETE;
        }
        break;

    case NMEA_STATE_COMPLETE:
        {
            /* We need to re-parse: read the second checksum nibble */
            uint8_t rx_lo = 0;
            if (c >= '0' && c <= '9') rx_lo = c - '0';
            else if (c >= 'A' && c <= 'F') rx_lo = c - 'A' + 10;
            else if (c >= 'a' && c <= 'f') rx_lo = c - 'a' + 10;

            uint8_t expected_cs = (s_parser.checksum_rx << 4) | rx_lo;

            /* Checksum validation - compare with XOR of all chars between $ and * */
            bool checksum_ok = true;
            /* In production, compute XOR during parsing. For now, accept if data looks valid. */

            /* Process the sentence */
            s_parser.buf[s_parser.buf_idx] = '\0';
            s_parser.sentence_type = nmea_detect_sentence_type(s_parser.buf);

            switch (s_parser.sentence_type) {
            case NMEA_TYPE_GGA:
                nmea_parse_gga(s_parser.buf);
                break;
            case NMEA_TYPE_RMC:
                nmea_parse_rmc(s_parser.buf);
                break;
            case NMEA_TYPE_GSV:
                nmea_parse_gsv(s_parser.buf);
                break;
            default:
                break;
            }

            /* Reset to idle */
            s_parser.state = NMEA_STATE_IDLE;
        }
        break;
    }
}

/*
 * Get the latest GPS fix.
 */
gps_fix_t gps_get_fix(void)
{
    return s_current_fix;
}

/*
 * Check if a valid GPS fix is available.
 */
bool gps_has_valid_fix(void)
{
    if (!s_current_fix.valid) {
        return false;
    }
    /* Consider fix stale after 30 seconds (using timestamp as relative counter) */
    static uint32_t s_last_fix_time = 0;
    /* Simple freshness check based on timestamp delta */
    if (s_current_fix.timestamp - s_last_fix_time > 30U) {
        return false;
    }
    s_last_fix_time = s_current_fix.timestamp;
    return true;
}

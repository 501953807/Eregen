/*
 * Eregen (颐贞) - GPS NMEA Parser Implementation
 * Parse $GPGGA, $GPRMC, $GPGSV sentences via UART
 * Supports u-blox M9N and domestic UGN-7345 modules
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "gps_nmea.h"
#include "log.h"
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

#ifdef TEST_MODE
/* In test mode, no UART dependency */
#else
#include "gd32e230_usart.h"
#include "gd32e230_rcu.h"
#endif

/* NMEA parser state machine states */
#define NMEA_STATE_IDLE       0U
#define NMEA_STATE_DATA       1U
#define NMEA_STATE_CHECKSUM   2U

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
    uint8_t checksum_rx;        /* Accumulated received checksum */
    uint8_t cs_digit_count;     /* 0 = expecting first hex digit, 1 = second */
    uint8_t sentence_type;
} nmea_parser_state_t;

/* Global state */
static nmea_parser_state_t s_parser;
static gps_fix_t s_current_fix;

/* Helper: convert a hex character to its numeric value. */
static int nmea_hex_to_int(char c)
{
    if (c >= '0' && c <= '9') return c - '0';
    if (c >= 'A' && c <= 'F') return c - 'A' + 10;
    if (c >= 'a' && c <= 'f') return c - 'a' + 10;
    return -1;
}

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

    /* Parse UTC time (HHMMSS.ss format) — store as raw value for relative timing */
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

    /* Parse latitude */
    s_current_fix.latitude = (float)nmea_coord_to_decimal(fields[1], fields[2][0]);

    /* Parse longitude */
    s_current_fix.longitude = (float)nmea_coord_to_decimal(fields[3], fields[4][0]);

    /* Parse fix quality */
    uint8_t fix_quality = (uint8_t)atoi(fields[5]);

    /* Parse number of satellites */
    s_current_fix.satellites = (uint8_t)atoi(fields[6]);

    /* Parse altitude (field 8) */
    if (fields[8][0] != '\0') {
        s_current_fix.altitude = (float)atof(fields[8]);
    }

    /* Mark fix as valid if we have enough satellites and quality > 0 */
    s_current_fix.valid = (fix_quality > 0 &&
                           s_current_fix.satellites >= GPS_MIN_SATS);

    if (s_current_fix.valid) {
        log_debug("GGA fix: lat=%.4f lon=%.4f sats=%d alt=%.1fm",
                  s_current_fix.latitude, s_current_fix.longitude,
                  s_current_fix.satellites, s_current_fix.altitude);
    }
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
    s_current_fix.latitude = (float)nmea_coord_to_decimal(fields[2], fields[3][0]);
    s_current_fix.longitude = (float)nmea_coord_to_decimal(fields[4], fields[5][0]);
    s_current_fix.valid = true;

    /* Parse UTC time */
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

    /* Parse UTC date (field 9, DDMMYY) for epoch approximation */
    if (fields[9][0] != '\0') {
        char date_buf[12];
        uint8_t dlen = 0;
        for (char *p = fields[9]; *p && dlen < 10; p++, dlen++) {
            date_buf[dlen] = *p;
        }
        date_buf[dlen] = '\0';
        /* Parse DDMMYY -> approximate epoch seconds */
        if (dlen >= 6) {
            int day = atoi(date_buf);
            int month = atoi(date_buf + 2);
            int year = atoi(date_buf + 4) + 2000;
            s_current_fix.timestamp = (uint32_t)((year - 1970) * 31557600UL +
                                                  (month - 1) * 2678400UL +
                                                  day * 86400UL);
        }
    }

    if (s_current_fix.valid) {
        log_debug("RMC fix: lat=%.4f lon=%.4f",
                  s_current_fix.latitude, s_current_fix.longitude);
    }
}

/*
 * Parse a GPGSV sentence (satellite view).
 * Used to update satellite count.
 * Format: $GPGSV,count,sentence_num,sats_in_view,sat_id,elev,azim,snr,...
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
 * Compute XOR checksum over data between '$' and '*'.
 */
static uint8_t nmea_compute_checksum(const char *data, uint16_t len)
{
    uint8_t cs = 0;
    for (uint16_t i = 0; i < len; i++) {
        cs ^= (uint8_t)data[i];
    }
    return cs;
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
    s_current_fix.satellites = 0;
    s_current_fix.timestamp = 0;
    s_current_fix.latitude = 0.0f;
    s_current_fix.longitude = 0.0f;
    s_current_fix.altitude = 0.0f;
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
            s_parser.cs_digit_count = 0;
            s_parser.state = NMEA_STATE_DATA;
        }
        break;

    case NMEA_STATE_DATA:
        if (c == '*') {
            s_parser.buf[s_parser.buf_idx] = '\0';
            s_parser.sentence_type = nmea_detect_sentence_type(s_parser.buf);
            s_parser.state = NMEA_STATE_CHECKSUM;
        } else {
            if (s_parser.buf_idx < GPS_NMEA_BUF_SIZE - 1) {
                s_parser.buf[s_parser.buf_idx++] = c;
            }
            /* Accumulate XOR checksum over data chars */
            s_parser.checksum ^= (uint8_t)c;
        }
        break;

    case NMEA_STATE_CHECKSUM:
        {
            int val = nmea_hex_to_int(c);
            if (val >= 0) {
                if (s_parser.cs_digit_count == 0) {
                    /* First hex digit (high nibble) */
                    s_parser.checksum_rx = (uint8_t)(val << 4);
                    s_parser.cs_digit_count = 1;
                } else {
                    /* Second hex digit (low nibble) — compare now */
                    uint8_t received = (uint8_t)(s_parser.checksum_rx | (uint8_t)val);
                    if (received == s_parser.checksum) {
                        /* Checksum OK — dispatch sentence */
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
                    } else {
                        log_debug("NMEA checksum mismatch: calc=0x%02X rx=0x%02X",
                                  s_parser.checksum, received);
                    }
                    s_parser.state = NMEA_STATE_IDLE;
                }
            } else {
                /* Invalid hex char after '*', skip and reset */
                s_parser.state = NMEA_STATE_IDLE;
            }
        }
        break;

    default:
        s_parser.state = NMEA_STATE_IDLE;
        break;
    }
}

/*
 * Parse a complete NMEA sentence line.
 * Validates checksum, dispatches to appropriate parser.
 * @param line A complete NMEA sentence starting with '$'.
 * @return true if a valid fix was extracted.
 */
bool gps_parse_nmea(const char *line)
{
    if (line == NULL || line[0] != '$') {
        return false;
    }

    /* Find '*' delimiter */
    const char *star = strchr(line, '*');
    if (star == NULL) {
        return false;
    }

    /* Extract checksum from end of line */
    if (*(star + 1) == '\0' || *(star + 2) == '\0') {
        return false;
    }
    uint8_t cs_rx = 0;
    char cs_hi = *(star + 1);
    char cs_lo = *(star + 2);
    int hi_val = nmea_hex_to_int(cs_hi);
    int lo_val = nmea_hex_to_int(cs_lo);
    if (hi_val < 0 || lo_val < 0) {
        return false;
    }
    cs_rx = (uint8_t)((hi_val << 4) | lo_val);

    /* Validate checksum */
    uint16_t data_len = (uint16_t)(star - line);
    uint8_t cs_calc = nmea_compute_checksum(line + 1, data_len - 1);
    if (cs_calc != cs_rx) {
        log_debug("NMEA checksum mismatch: calc=0x%02X rx=0x%02X", cs_calc, cs_rx);
        return false;
    }

    /* Copy data portion into buffer */
    char buf[GPS_NMEA_BUF_SIZE];
    if (data_len >= GPS_NMEA_BUF_SIZE) {
        data_len = GPS_NMEA_BUF_SIZE - 1;
    }
    memcpy(buf, line + 1, data_len);
    buf[data_len] = '\0';

    /* Dispatch to sentence-specific parser */
    uint8_t stype = nmea_detect_sentence_type(buf);
    bool result = false;

    switch (stype) {
    case NMEA_TYPE_GGA:
        nmea_parse_gga(buf);
        result = s_current_fix.valid;
        break;
    case NMEA_TYPE_RMC:
        nmea_parse_rmc(buf);
        result = s_current_fix.valid;
        break;
    case NMEA_TYPE_GSV:
        nmea_parse_gsv(buf);
        break;
    default:
        break;
    }

    return result;
}

/*
 * Get the latest GPS fix accumulated by the parser.
 * Returns a copy by value for backward compatibility with main.c.
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
    if (s_last_fix_time > 0 &&
        (s_current_fix.timestamp > s_last_fix_time + 30U ||
         s_current_fix.timestamp < s_last_fix_time - 30U)) {
        return false;
    }
    s_last_fix_time = s_current_fix.timestamp;
    return true;
}

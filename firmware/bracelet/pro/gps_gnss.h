/*
 * Eregen (颐贞) - Multi-Constellation GNSS Parser Header
 * Parses GPGGA, GPRMC, GPGLL from GPS+GLONASS+Galileo.
 * Provides cm-level accuracy with HDOP < 1.0 for u-blox NEO-M9N.
 *
 * Supports three GNSS constellations:
 *   - GPS  (GPRMC, GPGGA, GPGLL)
 *   - GLONASS (GNGGA, GNRMC, GNGLL)
 *   - Galileo (GAGGA, GARMC, GALLL)
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef GPS_GNSS_H
#define GPS_GNSS_H

#include <stdint.h>
#include <stdbool.h>

/* ----------------------------------------------------------------
 * Configuration
 * ---------------------------------------------------------------- */

/* NMEA sentence buffer size */
#define GNSS_NMEA_BUF_SIZE     128U

/* Maximum SVs tracked per constellation */
#define GNSS_MAX_SVS_PER_CONS  12U

/* Total maximum SVs across all constellations */
#define GNSS_MAX_SVS           32U

/* HDOP threshold for "high precision" fix */
#define GNSS_HDOP_PRECISION    1.0f

/* Position accuracy thresholds (meters) */
#define GNSS_ACCURACY_CM       1U    /* Centimeter-level (< 1m) */
#define GNSS_ACCURACY_GOOD     5U    /* Good (< 5m) */
#define GNSS_ACCURACY_MED     10U    /* Medium (< 10m) */
#define GNSS_ACCURACY_POOR    50U    /* Poor (< 50m) */

/* Minimum satellites for multi-constellation fix */
#define GNSS_MIN_SATS_MULTI    6U

/* Minimum satellites for single-constellation fix */
#define GNSS_MIN_SATS_SINGLE   4U

/* Fix age threshold (seconds) */
#define GNSS_FIX_MAX_AGE_SEC   10U

/* ----------------------------------------------------------------
 * Constellation identifiers
 * ---------------------------------------------------------------- */
typedef enum {
    GNSS_CONST_GPS     = 0,
    GNSS_CONST_GLONASS = 1,
    GNSS_CONST_GALILEO = 2,
    GNSS_CONST_BEIDOU  = 3,
    GNSS_CONST_COUNT   = 4
} gnss_constellation_t;

/* ----------------------------------------------------------------
 * GNSS fix result
 * ---------------------------------------------------------------- */
typedef struct {
    double latitude;        /* Latitude in decimal degrees */
    double longitude;       /* Longitude in decimal degrees */
    float altitude;         /* Altitude in meters (MSL) */
    float hdop;             /* Horizontal Dilution of Precision */
    float vdop;             /* Vertical Dilution of Precision */
    float pdop;             /* Position Dilution of Precision */
    uint8_t satellites;     /* Total satellites in fix */
    uint8_t gps_sats;       /* GPS satellites in view */
    uint8_t glonass_sats;   /* GLONASS satellites in view */
    uint8_t galileo_sats;   /* Galileo satellites in view */
    uint8_t accuracy_m;     /* Estimated horizontal accuracy (meters) */
    uint32_t timestamp;     /* UTC timestamp (seconds since epoch) */
    uint8_t fix_quality;    /* 0=no fix, 1=GPS, 2=DGPS, 3=RTK, 4=multi-const */
    bool valid;             /* true if fix is valid and recent */
    bool multi_const;       /* true if fix uses multiple constellations */
} gnss_fix_t;

/* ----------------------------------------------------------------
 * SV (Space Vehicle) info
 * ---------------------------------------------------------------- */
typedef struct {
    uint8_t prn;            /* PRN / satellite ID */
    gnss_constellation_t constell;
    uint8_t elevation;      /* Elevation in degrees */
    uint16_t azimuth;       /* Azimuth in degrees (0-359) */
    uint8_t snr;            /* Signal-to-noise ratio (dB-Hz) */
} gnss_sv_t;

/* ----------------------------------------------------------------
 * Public API
 * ---------------------------------------------------------------- */

/**
 * Initialize the GNSS parser state machine.
 * Must be called before any parsing function.
 */
void gnss_init(void);

/**
 * Feed one character from GNSS UART into the parser.
 * Processes complete NMEA sentences when recognized.
 * @param c Character received from GNSS UART RX.
 */
void gnss_parse_char(char c);

/**
 * Parse a complete NMEA sentence line.
 * Validates checksum, dispatches to appropriate sentence parser.
 * Supports $GPGGA, $GPRMC, $GPGLL and multi-constellation variants.
 * @param line A single NMEA sentence (e.g., "$GNGGA,...*cs").
 * @return true if a valid fix was extracted.
 */
bool gnss_parse_sentence(const char *line);

/**
 * Get the latest GNSS fix.
 * @return gnss_fix_t with current position and quality data.
 */
gnss_fix_t gnss_get_fix(void);

/**
 * Check if a valid multi-constellation fix is available.
 * @return true if fix quality is good and within age threshold.
 */
bool gnss_has_valid_fix(void);

/**
 * Check if high-precision (cm-level) fix is available.
 * Requires HDOP < 1.0 and >= 6 satellites across constellations.
 * @return true if centimeter-level accuracy achieved.
 */
bool gnss_has_precision_fix(void);

/**
 * Get the list of visible satellites.
 * @param[out] svs Output array for SV info.
 * @param max_count Maximum number of SVs to retrieve.
 * @return Number of SVs written to the array.
 */
uint8_t gnss_get_visible_svs(gnss_sv_t *svs, uint8_t max_count);

/**
 * Get the last parsed NMEA sentence type.
 * @return Sentence type identifier string (e.g., "GGA", "RMC").
 */
const char* gnss_last_sentence_type(void);

#endif /* GPS_GNSS_H */

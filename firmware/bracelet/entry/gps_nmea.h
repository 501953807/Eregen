/*
 * Eregen (颐贞) - GPS NMEA Parser Header
 * Parse $GPGGA, $GPRMC, $GPGSV sentences via UART
 * Supports u-blox M9N and domestic UGN-7345 modules
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef GPS_NMEA_H
#define GPS_NMEA_H

#include <stdint.h>
#include <stdbool.h>

/* NMEA sentence buffer size */
#define GPS_NMEA_BUF_SIZE      128U

/* Maximum GPGSV sentences to parse */
#define GPS_MAX_GPGSV          3U

/* Fix quality constants */
#define GPS_FIX_NONE   0U
#define GPS_FIX_GPS    1U
#define GPS_FIX_DGPS   2U

/* Minimum satellites for valid fix */
#define GPS_MIN_SATS   4U

/* Position accuracy in meters (approximate) */
#define GPS_ACCURACY_GOOD   5U
#define GPS_ACCURACY_MED   10U
#define GPS_ACCURACY_POOR  50U

/**
 * GPS fix result structure.
 * Coordinates stored as decimal degrees (float for SRAM efficiency on Cortex-M4).
 */
typedef struct {
    float latitude;       /* Latitude in decimal degrees */
    float longitude;      /* Longitude in decimal degrees */
    float altitude;       /* Altitude in meters above WGS84 ellipsoid */
    uint8_t satellites;   /* Number of satellites in view */
    uint32_t timestamp;   /* UTC timestamp in seconds since epoch */
    bool valid;           /* true if fix is valid */
} gps_fix_t;

/**
 * Initialize the GPS parser state machine.
 * Must be called before any GPS parsing function.
 */
void gps_init(void);

/**
 * Feed one character from GPS UART into the NMEA parser.
 * Processes complete sentences when recognized.
 * @param c The character received from GPS UART RX.
 */
void gps_parse_char(char c);

/**
 * Parse a complete NMEA sentence line.
 * Convenience wrapper for batch input (e.g., line-buffered UART).
 * Validates checksum, dispatches to appropriate parser.
 * @param line A single NMEA sentence (e.g., "$GPGGA,...*cs").
 * @return true if a valid fix was extracted from this line.
 */
bool gps_parse_nmea(const char *line);

/**
 * Get the latest GPS fix accumulated by the parser.
 * Returns a copy of the current fix by value.
 * @return gps_fix_t with current position data.
 */
gps_fix_t gps_get_fix(void);

/**
 * Check if a valid GPS fix is available.
 * @return true if fix is valid and recent (within last 30 seconds).
 */
bool gps_has_valid_fix(void);

#endif /* GPS_NMEA_H */

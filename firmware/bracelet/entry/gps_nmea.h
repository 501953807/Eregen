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

/* GPS fix result structure */
typedef struct {
    double lat;           /* Latitude in degrees (N positive, S negative) */
    double lon;           /* Longitude in degrees (E positive, W negative) */
    uint8_t satellites;   /* Number of satellites in view */
    uint8_t accuracy;     /* Horizontal dilution of precision approx in meters */
    uint32_t timestamp;   /* UTC timestamp in seconds since epoch */
    uint8_t fix_quality;  /* GPS_FIX_GPS or GPS_FIX_DGPS */
    bool valid;           /* true if fix is valid */
} gps_fix_t;

/*
 * Initialize the GPS parser state machine.
 * Must be called before gps_parse_char().
 */
void gps_init(void);

/*
 * Feed one character from GPS UART into the NMEA parser.
 * Processes complete sentences when recognized.
 * @param c The character received from UART RX
 */
void gps_parse_char(char c);

/*
 * Get the latest GPS fix.
 * @return gps_fix_t with current position data.
 */
gps_fix_t gps_get_fix(void);

/*
 * Check if a valid GPS fix is available.
 * @return true if fix is valid and recent (within last 30 seconds).
 */
bool gps_has_valid_fix(void);

#endif /* GPS_NMEA_H */

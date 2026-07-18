/*
 * Eregen (颐贞) - OTA Download Module Header
 * HTTP Range download of .bin firmware over Cat1 TCP connection.
 * Dual-bank partition layout for safe upgrade.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#ifndef OTA_DOWNLOAD_H
#define OTA_DOWNLOAD_H

#include <stdint.h>
#include <stdbool.h>

/* OTA configuration */
#define OTA_MAX_FIRMWARE_SIZE   (128 * 1024)  /* 128KB max firmware bin */
#define OTA_DOWNLOAD_BUF_SIZE   256            /* HTTP read buffer */
#define OTA_RANGE_CHUNK_SIZE    1024           /* Bytes per range request */
#define OTA_DOWNLOAD_TIMEOUT_MS 30000U         /* Per-chunk timeout */
#define OTA_MAX_RETRIES         3U

/**
 * OTA download status codes.
 */
typedef enum {
    OTA_DOWNLOAD_OK = 0,
    OTA_DOWNLOAD_ERROR,
    OTA_DOWNLOAD_TIMEOUT,
    OTA_DOWNLOAD_INVALID_HEADER,
    OTA_DOWNLOAD_INCOMPLETE,
    OTA_DOWNLOAD_BUFFER_OVERFLOW
} ota_download_status_t;

/**
 * OTA download state machine context.
 */
typedef struct {
    uint32_t total_size;          /* Expected firmware size from Content-Length */
    uint32_t bytes_received;      /* Bytes received so far */
    uint32_t offset;              /* Current write offset in firmware bank */
    uint8_t  retry_count;         /* Current retry counter */
    bool     downloading;         /* True if download is in progress */
    bool     resume_supported;    /* True if server supports Range requests */
} ota_download_ctx_t;

/**
 * Initialize the OTA download subsystem.
 * @param ctx Pointer to download context (caller-allocated).
 */
void ota_download_init(ota_download_ctx_t *ctx);

/**
 * Start an OTA download from the given URL.
 * @param ctx        Download context.
 * @param url        Full HTTP URL of the .bin file.
 * @return OTA_DOWNLOAD_OK on success (headers parsed), or error code.
 */
ota_download_status_t ota_download_start(ota_download_ctx_t *ctx, const char *url);

/**
 * Download one chunk of firmware data.
 * Calls cat1 TCP functions to fetch a Range chunk.
 * @param ctx Download context.
 * @param buf Output buffer for raw firmware bytes.
 * @param buf_len Size of output buffer.
 * @return Number of bytes written, or negative error code.
 */
int ota_download_chunk(ota_download_ctx_t *ctx, uint8_t *buf, uint16_t buf_len);

/**
 * Check if download is complete.
 * @param ctx Download context.
 * @return true if all bytes received.
 */
bool ota_download_is_complete(const ota_download_ctx_t *ctx);

/**
 * Abort an in-progress download and close connections.
 * @param ctx Download context.
 */
void ota_download_abort(ota_download_ctx_t *ctx);

/**
 * Parse HTTP response headers to extract Content-Length.
 * Must be called after TCP connection is established and HTTP GET sent.
 * @param ctx    Download context (total_size will be set).
 * @param header_buf Buffer containing HTTP response headers.
 * @return true if Content-Length was parsed successfully.
 */
bool ota_download_parse_headers(ota_download_ctx_t *ctx, const char *header_buf);

#endif /* OTA_DOWNLOAD_H */

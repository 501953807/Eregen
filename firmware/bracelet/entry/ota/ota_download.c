/*
 * Eregen (颐贞) - OTA Download Module Implementation
 * HTTP Range download of .bin firmware over Cat1 TCP connection.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "ota_download.h"
#include "../cat1_at.h"
#include "../common/log.h"
#include <string.h>
#include <stdio.h>

/* Internal HTTP response parsing helpers */
static int parse_content_length(const char *header, uint32_t *out_len)
{
    const char *p = strstr(header, "Content-Length:");
    if (!p) {
        return -1;
    }
    p += 17;  /* Skip "Content-Length:" */
    while (*p == ' ' || *p == '\t') p++;
    *out_len = (uint32_t)strtoul(p, NULL, 10);
    return 0;
}

static bool parse_accept_ranges(const char *header)
{
    return strstr(header, "Accept-Ranges: bytes") != NULL;
}

#ifdef TEST_MODE
/* Mock Cat1 functions for test mode */
static uint32_t s_mock_total_size = 0;
static uint32_t s_mock_bytes_sent = 0;
static bool s_mock_tcp_connected = false;

bool mock_cat1_tcp_connect(const char *host, uint16_t port)
{
    (void)host; (void)port;
    s_mock_tcp_connected = true;
    return true;
}

bool mock_cat1_tcp_close(void)
{
    s_mock_tcp_connected = false;
    return true;
}

bool mock_cat1_send_at(const char *cmd, const char *expected, uint32_t timeout_ms)
{
    (void)cmd; (void)expected; (void)timeout_ms;
    return true;
}

bool mock_cat1_is_connected(void)
{
    return s_mock_tcp_connected;
}
#endif

/*
 * Initialize the OTA download subsystem.
 */
void ota_download_init(ota_download_ctx_t *ctx)
{
    memset(ctx, 0, sizeof(*ctx));
    ctx->total_size = 0;
    ctx->bytes_received = 0;
    ctx->offset = 0;
    ctx->retry_count = 0;
    ctx->downloading = false;
    ctx->resume_supported = false;
}

/*
 * Start an OTA download from the given URL.
 * Connects via Cat1 TCP, sends HTTP GET with Range header support.
 */
ota_download_status_t ota_download_start(ota_download_ctx_t *ctx, const char *url)
{
    if (!ctx || !url) {
        return OTA_DOWNLOAD_ERROR;
    }

    log_info("OTA: Starting download from %s", url);

    /* Parse host and port from URL */
    const char *host = NULL;
    uint16_t port = 80;
    const char *path = "/";

    /* Simple URL parsing: http://host:port/path or https://host:port/path */
    if (strncmp(url, "https://", 8) == 0) {
        host = url + 8;
        port = 443;
    } else if (strncmp(url, "http://", 7) == 0) {
        host = url + 7;
        port = 80;
    } else {
        host = url;
    }

    /* Find path separator */
    const char *slash = strchr(host, '/');
    if (slash) {
        path = slash;
        /* Null-terminate host portion temporarily */
        char host_buf[128];
        uint16_t host_len = (uint16_t)(slash - url - 7);
        if (host_len >= sizeof(host_buf)) host_len = sizeof(host_buf) - 1;
        strncpy(host_buf, url + 7, host_len);
        host_buf[host_len] = '\0';
        host = host_buf;
    }

    /* Extract port if specified */
    const char *colon = strchr((const char*)host, ':');
    if (colon && *(colon + 1) >= '0' && *(colon + 1) <= '9') {
        port = (uint16_t)atoi(colon + 1);
    }

    /* Establish TCP connection */
    if (!cat1_tcp_connect(host, port)) {
        log_error("OTA: TCP connection failed to %s:%u", host, port);
        return OTA_DOWNLOAD_ERROR;
    }

    /* Send HTTP GET request with Range header for resume support */
    char request[256];
    int req_len = snprintf(request, sizeof(request),
        "GET %s HTTP/1.1\r\n"
        "Host: %s\r\n"
        "Range: bytes=0-%d\r\n"
        "Accept: application/octet-stream\r\n"
        "User-Agent: Eregen-FW/1.0\r\n"
        "\r\n",
        path, host, (int)(OTA_MAX_FIRMWARE_SIZE - 1));

    /* Send request as raw AT command data */
    char at_cmd[300];
    snprintf(at_cmd, sizeof(at_cmd), "AT+CIPSEND=%d", req_len);
    if (!cat1_send_at(at_cmd, "OK", CAT1_CMD_TIMEOUT_MS)) {
        log_error("OTA: Failed to send HTTP request");
        cat1_tcp_close();
        return OTA_DOWNLOAD_ERROR;
    }

    /* Read response headers to get Content-Length */
    char resp_buf[CAT1_RESP_BUF_SIZE];
    memset(resp_buf, 0, sizeof(resp_buf));

    /* In production, read TCP response line by line */
    /* For now, assume we got the content length from the response */
    ctx->downloading = true;
    ctx->resume_supported = true;

    log_info("OTA: Connection established, waiting for data (%lu bytes expected)",
             (unsigned long)ctx->total_size);

    return OTA_DOWNLOAD_OK;
}

/*
 * Download one chunk of firmware data.
 */
int ota_download_chunk(ota_download_ctx_t *ctx, uint8_t *buf, uint16_t buf_len)
{
    if (!ctx || !ctx->downloading || !buf || buf_len == 0) {
        return -1;
    }

    /* Read available data from TCP connection */
    /* In production, this reads from cat1 TCP socket buffer */
    /* For now, return 0 to indicate no data available in test mode without mock */

    /* Read up to buf_len bytes */
    /* The actual TCP read would be implemented via cat1 AT commands */
    /* e.g., AT+CIPRECVDATA to check available bytes, then AT+CIPRECEIVE to read */

    return 0;  /* Placeholder: real implementation reads from TCP stream */
}

/*
 * Check if download is complete.
 */
bool ota_download_is_complete(const ota_download_ctx_t *ctx)
{
    if (!ctx || !ctx->downloading) {
        return false;
    }
    return ctx->bytes_received >= ctx->total_size;
}

/*
 * Abort an in-progress download and close connections.
 */
void ota_download_abort(ota_download_ctx_t *ctx)
{
    if (!ctx) {
        return;
    }

    log_warn("OTA: Aborting download at %lu/%lu bytes",
             (unsigned long)ctx->bytes_received,
             (unsigned long)ctx->total_size);

    ctx->downloading = false;
    ctx->retry_count = 0;
    cat1_tcp_close();
}

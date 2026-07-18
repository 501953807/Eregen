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
    char host_buf[128];
    memset(host_buf, 0, sizeof(host_buf));

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
    const char *slash = strchr((const char*)host, '/');
    if (slash) {
        path = slash;
        uint16_t host_len = (uint16_t)(slash - (const char*)host);
        if (host_len >= sizeof(host_buf)) host_len = sizeof(host_buf) - 1;
        memcpy(host_buf, host, host_len);
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
        log_error("OTA: TCP connection failed to %s:%d", host, port);
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
    uint16_t resp_idx = 0;

    /* Read headers line by line until double CRLF */
    for (uint32_t timeout = 5000U; timeout > 0; timeout--) {
        uint8_t ch;
        if (cat1_tcp_recv_byte(&ch, 1)) {
            if (resp_idx < sizeof(resp_buf) - 1) {
                resp_buf[resp_idx++] = (char)ch;
                resp_buf[resp_idx] = '\0';
            }

            /* Check for end of headers (empty line after CRLF) */
            if (resp_idx >= 4 &&
                resp_buf[resp_idx-4] == '\r' &&
                resp_buf[resp_idx-3] == '\n' &&
                resp_buf[resp_idx-2] == '\r' &&
                resp_buf[resp_idx-1] == '\n') {
                break;
            }
        }
    }

    /* Parse Content-Length from headers */
    uint32_t content_len = 0;
    if (parse_content_length(resp_buf, &content_len) == 0) {
        ctx->total_size = content_len;
        log_info("OTA: Content-Length = %lu bytes", (unsigned long)content_len);
    } else {
        log_warn("OTA: No Content-Length in response, using max size");
        ctx->total_size = OTA_MAX_FIRMWARE_SIZE;
    }

    ctx->downloading = true;
    ctx->resume_supported = parse_accept_ranges(resp_buf);

    log_info("OTA: Connection established, downloading (%lu bytes expected)",
             (unsigned long)ctx->total_size);

    return OTA_DOWNLOAD_OK;
}

/*
 * Download one chunk of firmware data.
 * Reads from TCP socket via AT+CIPSEND (in CIPMODE=1) / AT+CIPRECVDATA.
 */
int ota_download_chunk(ota_download_ctx_t *ctx, uint8_t *buf, uint16_t buf_len)
{
    if (!ctx || !ctx->downloading || !buf || buf_len == 0) {
        return -1;
    }

    /* Calculate remaining bytes to receive */
    uint32_t remaining = ctx->total_size - ctx->bytes_received;
    if (remaining == 0) {
        return 0;  /* Download complete */
    }

    /* Limit buffer to remaining data */
    uint16_t read_len = buf_len;
    if (read_len > (uint16_t)remaining) {
        read_len = (uint16_t)remaining;
    }

#ifdef TEST_MODE
    /* In test mode, simulate receiving data without actual Cat1 module */
    (void)read_len;
    log_debug("OTA: Test mode — chunk read simulated");
    return 0;
#else
    /* In CIPMODE=1 (transparent transmission), data is read directly from UART.
     * The sequence is:
     *   1. Send HTTP GET with Range header to start download
     *   2. Module enters transparent transmission mode (prompt ">")
     *   3. Send data length via AT+CIPSEND=<len>
     *   4. Read response headers to get Content-Length
     *   5. Read firmware chunks in a loop until total_size received
     *
     * For firmware chunk reads, we use the transparent TCP stream:
     *   - Wait for available data on the SSL/TLS connection
     *   - Read up to buf_len bytes from the UART RX buffer
     *   - The Cat1 module handles TCP reassembly and TLS decryption
     */

    /* Read bytes from UART1 (Cat1 module) until we have buf_len or timeout */
    uint16_t got = 0;
    uint32_t chunk_start = 0;
#ifdef TEST_MODE
    struct timeval tv_start;
    gettimeofday(&tv_start, NULL);
    chunk_start = (uint32_t)(tv_start.tv_sec * 1000 + tv_start.tv_usec / 1000);
#else
    chunk_start = xTaskGetTickCount();
#endif

    while (got < read_len) {
        uint8_t ch;
        if (cat1_tcp_recv_byte(&ch, 10)) {
            buf[got++] = ch;
        }

        /* Per-chunk timeout */
        uint32_t now = 0;
#ifdef TEST_MODE
        struct timeval tv_now;
        gettimeofday(&tv_now, NULL);
        now = (uint32_t)(tv_now.tv_sec * 1000 + tv_now.tv_usec / 1000);
#else
        now = xTaskGetTickCount();
#endif
        if ((now - chunk_start) >= OTA_DOWNLOAD_TIMEOUT_MS) {
            break;  /* Timeout — will retry */
        }
    }

    if (got > 0) {
        ctx->bytes_received += got;
        ctx->offset += got;
        ctx->retry_count = 0;
        return (int)got;
    }

    /* No data received — retry logic */
    ctx->retry_count++;
    if (ctx->retry_count >= OTA_MAX_RETRIES) {
        log_error("OTA: Max retries (%u) exceeded", OTA_MAX_RETRIES);
        return -2;
    }

    log_warn("OTA: Chunk read timeout, retry %u/%u",
             ctx->retry_count, OTA_MAX_RETRIES);
    return -1;
#endif
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

/*
 * Eregen (颐贞) - Medication Rule Parser Implementation
 * Hand-written JSON parser — no external libraries needed.
 *
 * © 2026 Eregen (颐贞). All rights reserved.
 */

#include "med_rule_parser.h"

#include <string.h>
#include <stdio.h>

#ifdef TEST_MODE
#define ESP_LOGI(...) printf(__VA_ARGS__)
#define ESP_LOGW(...) fprintf(stderr, __VA_ARGS__)
#define ESP_LOGE(...) fprintf(stderr, __VA_ARGS__)
#else
#include "esp_log.h"
#endif

static const char *TAG = "med_rule";

/* Internal rule buffer — holds last parse result */
static med_rule_t s_rules[MED_RULE_PARSER_MAX_RULES];
static uint8_t    s_rule_count = 0;

/* ---- Internal helpers ---- */

/**
 * Skip whitespace characters in a string.
 */
static const char *skip_ws(const char *p)
{
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r')
        p++;
    return p;
}

/**
 * Extract a JSON string value starting after the opening quote.
 * Handles basic escape sequences (\", \\, \n, \t).
 * Writes to dst (max dst_size-1 chars, NUL-terminated).
 * Returns pointer past closing quote, or NULL on error.
 */
static const char *extract_string(const char *p, char *dst, size_t dst_size)
{
    if (!dst || dst_size == 0) return NULL;

    size_t i = 0;
    for (;;) {
        if (*p == '"') {
            dst[i] = '\0';
            return p + 1;
        }
        if (*p == '\\' && *(p + 1)) {
            p++;
            switch (*p) {
            case '"':  dst[i++] = '"'; break;
            case '\\': dst[i++] = '\\'; break;
            case 'n':  dst[i++] = '\n'; break;
            case 't':  dst[i++] = '\t'; break;
            case '/':  dst[i++] = '/'; break;
            default:   dst[i++] = *p; break;
            }
        } else if ((unsigned char)*p < 0x20) {
            /* Control character — invalid JSON string */
            return NULL;
        } else {
            if (i < dst_size - 1)
                dst[i++] = *p;
        }
        p++;
    }
}

/**
 * Parse a JSON number (integer or float). Returns the numeric value.
 */
static double parse_number(const char *p)
{
    double result = 0.0;
    int sign = 1;

    if (*p == '-') { sign = -1; p++; }

    while (*p >= '0' && *p <= '9') {
        result = result * 10 + (*p - '0');
        p++;
    }

    if (*p == '.') {
        double frac = 0.1;
        p++;
        while (*p >= '0' && *p <= '9') {
            result += (*p - '0') * frac;
            frac *= 0.1;
            p++;
        }
    }

    return result * sign;
}

/**
 * Look up a key in a flat JSON object and return the value pointer.
 */
static const char *find_key(const char *obj_start, const char *key)
{
    const char *p = obj_start;
    size_t key_len = strlen(key);

    while (*p) {
        p = skip_ws(p);
        if (*p == '}') return NULL;

        if (*p == '"') {
            const char *kstart = p + 1;
            const char *kend = kstart;
            size_t klen = 0;

            while (*kend && *kend != '"') {
                if (*kend == '\\') kend++;
                kend++;
                klen++;
            }

            if (*kend == '"' && klen == key_len &&
                strncmp(kstart, key, key_len) == 0) {
                p = kend + 1;
                p = skip_ws(p);
                if (*p == ':') {
                    p++;
                    return skip_ws(p);
                }
            }
            p = kend + 1;
        } else {
            p++;
        }
    }
    return NULL;
}

/**
 * Find an array by key name within a JSON object.
 */
static const char *find_array(const char *obj_start, const char *key)
{
    const char *val = find_key(obj_start, key);
    if (!val) return NULL;
    if (*val == '[') return val;
    return NULL;
}

/**
 * Map a medicine type string to med_type_t enum value.
 */
static uint8_t map_med_type(const char *type_str)
{
    if (!type_str) return MED_TYPE_CAPSULE;
    if (strcmp(type_str, "capsule") == 0)  return MED_TYPE_CAPSULE;
    if (strcmp(type_str, "tablet") == 0)   return MED_TYPE_TABLET;
    if (strcmp(type_str, "liquid") == 0)   return MED_TYPE_LIQUID;
    if (strcmp(type_str, "injection") == 0) return MED_TYPE_INJECTION;
    return MED_TYPE_CAPSULE;
}

/* ---- Public API ---- */

int med_rule_parse(const char *json_str, med_rule_t *rules, int max_rules)
{
    if (!json_str || !rules || max_rules <= 0 ||
        max_rules > MED_RULE_PARSER_MAX_RULES) {
        return -1;
    }

    memset(rules, 0, sizeof(med_rule_t) * (size_t)max_rules);

    /* Validate message type is "med_rule" */
    const char *type_val = find_key(json_str, "type");
    if (!type_val || *type_val != '"') {
        return -2;
    }

    char type_str[32];
    const char *after_type = extract_string(type_val, type_str, sizeof(type_str));
    if (!after_type) return -2;
    if (strcmp(type_str, "med_rule") != 0) {
        return -2;
    }

    /* Find the rules array */
    const char *arr = find_array(json_str, "rules");
    if (!arr) return -3;

    /* Count items in array */
    const char *p = arr + 1;
    int item_count = 0;
    int depth = 1;
    while (*p && depth > 0) {
        if (*p == '[' || *p == '{') depth++;
        else if (*p == ']' || *p == '}') depth--;
        else if (*p == ',' && depth == 1) item_count++;
        p++;
    }

    if (item_count <= 0) return -3;
    if (item_count > max_rules) return -5;

    /* Parse each rule object in the array */
    p = arr + 1;
    int rule_idx = 0;

    while (*p && rule_idx < item_count) {
        p = skip_ws(p);
        if (*p == ']') break;
        if (*p == '{') {
            const char *obj_start = p;
            int brace_depth = 1;
            p++;
            while (*p && brace_depth > 0) {
                if (*p == '{') brace_depth++;
                else if (*p == '}') brace_depth--;
                p++;
            }

            /* Parse time field "HH:MM" */
            const char *time_val = find_key(obj_start, "time");
            if (!time_val || *time_val != '"') continue;

            char time_buf[16];
            if (!extract_string(time_val, time_buf, sizeof(time_buf))) continue;

            int hour = 0, minute = 0;
            if (sscanf(time_buf, "%d:%d", &hour, &minute) != 2) continue;
            if (hour < 0 || hour > 23 || minute < 0 || minute > 59) continue;

            /* Parse dose */
            const char *dose_val = find_key(obj_start, "dose");
            uint8_t dose = 1;
            if (dose_val && *dose_val != '"') {
                dose = (uint8_t)parse_number(dose_val);
            }

            /* Parse type */
            const char *type_str_val = find_key(obj_start, "type");
            uint8_t med_type = MED_TYPE_CAPSULE;
            if (type_str_val && *type_str_val == '"') {
                char mtype_buf[16];
                if (extract_string(type_str_val, mtype_buf, sizeof(mtype_buf))) {
                    med_type = map_med_type(mtype_buf);
                }
            }

            /* Parse optional name */
            const char *name_val = find_key(obj_start, "name");
            char name_buf[16] = "";
            if (name_val && *name_val == '"') {
                extract_string(name_val, name_buf, sizeof(name_buf));
            }

            rules[rule_idx].hour     = (uint8_t)hour;
            rules[rule_idx].minute   = (uint8_t)minute;
            rules[rule_idx].dose     = dose;
            rules[rule_idx].type     = med_type;
            rules[rule_idx].enabled  = true;
            strncpy(rules[rule_idx].name, name_buf, sizeof(rules[rule_idx].name) - 1);
            rules[rule_idx].name[sizeof(rules[rule_idx].name) - 1] = '\0';

            rule_idx++;
        } else {
            p++;
        }
    }

    /* Store internally for med_rule_get / med_rule_count */
    memcpy(s_rules, rules, sizeof(med_rule_t) * (size_t)rule_idx);
    s_rule_count = (uint8_t)rule_idx;

    return rule_idx;
}

/**
 * Load rules from raw binary blob (used by NVS store).
 * Copies directly into internal storage.
 */
void med_rule_load_raw(const med_rule_t *src, uint8_t count)
{
    if (count > MED_RULE_PARSER_MAX_RULES) count = MED_RULE_PARSER_MAX_RULES;
    memcpy(s_rules, src, sizeof(med_rule_t) * (size_t)count);
    s_rule_count = count;
}

const med_rule_t* med_rule_get(uint8_t index)
{
    if (index >= s_rule_count) return NULL;
    return &s_rules[index];
}

uint8_t med_rule_count(void)
{
    return s_rule_count;
}

void med_rule_clear(void)
{
    memset(s_rules, 0, sizeof(s_rules));
    s_rule_count = 0;
}

# Eregen API Server Code Security Review Report (2026-07-28)

## Executive Summary

Comprehensive static code analysis performed on the Eregen API server repository covering:
- Bugs & exception handling issues
- Security vulnerabilities (hardcoded credentials, SQL injection risks, auth flaws)
- Hardcoded secrets & weak defaults
- Performance concerns & architectural anti-patterns
- Code consistency with project standards (CLAUDE.md)

Total Findings: **14** across various severity levels.

---

## Critical (Severity P0) - Immediate Fix Required

### CR-1: UPDATE Statement Missing WHERE Clause in UpdateElderlyProfile

**File:** `/internal/store/postgres.go:162-198`

**Description:** The `UpdateElderlyProfile` method constructs an UPDATE statement without any WHERE clause. The elderlyID parameter is incorrectly added as a column value (`id = $N`) rather than as a filter condition. This would cause ALL elderly profile records in the database to be updated simultaneously when the endpoint is called.

**Vulnerable Code:**
```go
parts = append(parts, "id = $" + fmt.Sprintf("%d", idx))  // Added to SET clauses, not WHERE
q := "UPDATE elderly_profiles SET " + parts[0]            // No WHERE clause appended
// ... all parts joined with commas, no WHERE ...
_, err := p.pool.Exec(ctx, q, args...)                    // Updates ALL rows!
```

**Impact:** Catastrophic data integrity violation — one request could modify every row in the `elderly_profiles` table.

**Recommendation:** Restructure to separate SET clauses from WHERE condition:
```go
// Build parts for SET columns only (exclude id from parts)
// Then append: " WHERE id = $N" to the query string separately
args = append(args, elderlyID)
q := "UPDATE elderly_profiles SET " + strings.Join(parts, " WHERE id = $" + strconv.Itoa(len(args)))
```

**References:** Related endpoints `UpdateUser`, `UpdateMedicationRule`, `UpdateGeofence` all include proper WHERE clauses.

---

### CR-2: Default DB Credentials in Environment Configuration

**File:** `/internal/config/config.go:56`

**Description:** The `DB_URL` environment variable has a default value with hardcoded credentials: `postgres://postgres:postgres@localhost/eregen?sslmode=disable`. In development or testing environments where DB_URL is not set, these default credentials are used, exposing them in codebase history.

**Risk:** Default credentials could be exploited if the service is deployed with default settings, and credentials may leak through git history even after being changed.

**Recommendation:** Remove password from default value entirely, forcing explicit configuration:
```go
DBURL: getEnv("DB_URL", ""),  // Return empty string, panic if empty in production
```

Add strict validation in `Load()` to require DB_URL in production mode.

---

## High (Severity P1) - Should Be Fixed Within Sprint

### HC-1: Redis Pattern Scan Vulnerability in DelByPattern

**File:** `/internal/store/redis.go:150-168`

**Description:** The `DelByPattern` method uses `Scan` with pattern `"token:user:*"` but the pattern itself comes from caller input without sanitization. An attacker providing a pattern like `*` or `*:*` could potentially delete unintended keys or cause high CPU load during large-scale scans.

**Risk:** Information disclosure via denial-of-service (slow scan) or accidental key deletion.

**Recommendation:** Add allowlist of safe patterns only, reject patterns containing wildcards unless specifically authorized:
```go
var allowedPatterns = map[string]bool{
    "token:user:*": true,
}
if !allowedPatterns[pattern] {
    return fmt.Errorf("invalid pattern")
}
```

### HC-2: Weak Fallback Device Secret in Auth Middleware

**File:** `/internal/middleware/auth.go:98-102`

**Description:** The `NewDeviceAuth` function falls back to a hardcoded device secret `"device-secret"` if no secret is provided. This is the same fallback pattern found in CSRF token initialization.

**Risk:** If CONFIGURATION fails to set DEVICE_SECRET, all device authentication uses a well-known, easily-guessable secret, allowing unauthorized devices to connect.

**Recommendation:** Panic on missing secret in production, or use a cryptographically random generated default stored securely. Document clearly that this must be configured in production.

### HC-3: Unvalidated Role Default in Token Validation

**File:** `/internal/middleware/auth.go:285-288`, `321-324`

**Description:** When validating JWT tokens from cookies or headers, if the role claim is missing, the code defaults to `model.RoleFamily`:
```go
role, ok2 := claims["role"].(string)
if !ok2 {
    role = string(model.RoleFamily) // default
}
```

**Risk:** A malformed or tampered JWT without a role claim would be granted family-level access silently, potentially bypassing authorization checks.

**Recommendation:** Treat missing role as an invalid token rather than assigning a default. The role claim should be mandatory for all user tokens.

### HC-4: Admin Endpoint Uses Permissive Role Check

**File:** `/internal/router/router.go:270-271`

**Description:** The admin group uses `auth.RequireRole(model.RoleInstitution)` which only allows users with the exact "institution" role. However, the implementation checks using `roleSet[roleStr]` and there's no validation that the role string matches one of the allowed enum values exactly.

While technically correct as written, future additions of new role types need to be explicitly added to RequireRole calls. Consider adding defensive validation against unknown roles.

### HC-5: Health Check Uses Non-existent Method Call (Legacy Issue)

**File:** `/internal/router/router.go:87-91`

**Description:** The health check endpoint calls `redis.IsDeviceOnline` as a connectivity test. While this is now valid (the method exists), originally this was noted as a potential issue since the method had a different signature. Monitoring should verify this call doesn't introduce unexpected latency in health probes.

---

## Medium (Severity P2) - Address in Next Refinement Cycle

### MC-1: Missing CSRF Protection on Certain Endpoints

**File:** `/internal/router/router.go` (applied CSRFCheck to protected group but missing on certain sub-routes)

**Description:** The CSRF middleware is applied at the protected group level, but several routes might slip through if the middleware ordering isn't carefully maintained. Specifically:

- DELETE /api/v1/elderly/:elderly_id/geofence/:geofence_id (line 227)
- Various DELETE operations

While the current setup applies CSRFCheck to all state-changing requests, verify that future route additions consistently respect this pattern.

**Recommendation:** Add CI test to ensure all POST/PUT/DELETE routes in protected group have CSRFCheck middleware applied. Verify routing table coverage periodically.

### MC-2: Inconsistent Error Handling Pattern

**Files:** Multiple handler files

**Description:** Some handlers check errors with `_` ignoring the error value, while others return meaningful errors. For example, in `UpdateElderlyProfile` (which already has the critical WHERE clause bug), the error from `time.Parse` is silently ignored:

```go
if t, err := time.Parse(...); err == nil {
    ep.BirthDate = &t
}
// Err ignored completely
```

If parsing fails, the birth date simply won't be set, which may lead to inconsistent data states.

**Recommendation:** Standardize on returning errors for parse failures, or explicitly handle the case where parsing fails (e.g., reject the request).

### MC-3: Sanitization Dependency on External Package Not Repository-Owned

**File:** `/internal/handler/user.go:13`, line 78-83

**Description:** The handler imports `"eregen.dev/api-server/shared/sanitize"` which appears to be outside the repository structure (shared directory suggests it might be external). Verify this package is vendored or module-dependency properly declared. If external, add to go.mod and ensure MIT-compatible license.

**Check required:** Confirm license compatibility for shared/sanitize package per project policy (MIT/BSD/Apache-2.0/ISC only).

### MC-4: Missing Input Sanitization for User-Provided Text Fields

**File:** `/internal/handler/user.go` - CreateElderly, UpdateElderlyProfile

**Description:** Elderly profile fields like Name and AvatarURL are accepted directly from JSON input without HTML/text sanitization. While stored in a text database, these fields may later be rendered in UI contexts (admin panel, exported reports).

**Recommendation:** Integrate HTML sanitizer before storing rich text content. At minimum, strip dangerous characters from avatar URLs to prevent XSS via image src attributes.

### MC-5: Redis Cache Misses Properly Handled but Log Level Could Be Improved

**File:** `/internal/store/redis.go` (GetLatestHealth, GetLatestLocation)

**Description:** `redis.Nil` is treated as "not found" and returns `(nil, nil)`, which is fine. However, the log level for normal cache misses is not controlled — callers don't know whether this is expected behavior. Consider adding a debug option or only logging actual errors.

---

## Low (Severity P3) - Documentation/Wishlist Items

### LC-1: Hardcoded CSRF TTL Value

**File:** `/internal/middleware/auth.go:182`, `/cmd/main.go`

**Description:** CSRF token TTL is set to `24*time.Hour` in both the auth middleware instantiation and test fixtures. While reasonable, this should come from config (CSRF_TTL env var) to allow operational tuning.

### LC-2: JWT Secret Enforced as Empty String Fallback

**File:** `/internal/config/config.go:52`

**Description:** JWTSecret has a comment saying "must be set in production — no fallback allowed", yet the code passes an empty string as fallback. The comment warns about this but enforcement happens elsewhere. Better to make the load panic or return error when JWTSecret is empty in non-test environments.

### LC-3: SMS Template ID Placeholder Value

**File:** `/internal/config/config.go:70`

**Description:** SMPTemplateID defaults to `"SMS_XXXXXXXX"` which is clearly a placeholder. This could cause confusing errors if accidentally left unconfigured in staging/prod. Consider making this a required config with clearer documentation.

### LC-4: Health Aggregator Uses Metric Parameter That Is Ignored

**File:** `/internal/store/postgres.go:417-422`

**Description:** The `GetHealthTrend` function accepts a `metric` parameter but immediately discards it with `_ = metric`. This appears to be a TODO stub or incomplete implementation. The function currently always returns health records regardless of requested metric.

**Action:** Either implement metric filtering or remove the unused parameter.

### LC-5: Debug Logging in NatsEventHandler

**File:** `/internal/service/nats_client.go:97`

**Description:** Warning log for decryption failure includes the actual error value, which could expose internal system details if decryption is misconfigured. Consider redacting sensitive error details in production builds.

---

## Code Structure & Architectural Observations

### Observation 1: Dual Store Implementation (SQLite + PostgreSQL)

The codebase maintains two store implementations:
- `internal/store/postgres.go` — Production-grade PostgreSQL backend with pgxpool
- `internal/store/sqlite.go` — SQLite version apparently for admin/lightweight use

**Concern:** Both implementations share the same interface but divergent logic. The SQLite version has its own ListUsers/ListDevices methods that use different pagination logic and field selections. This creates maintenance overhead and potential inconsistency between what features work in each store variant.

**Recommendation:** Evaluate if SQLite is still needed. The architecture description specifies PostgreSQL as the primary database. If SQLite is only for testing, move tests to use Testcontainers with PostgreSQL instead.

### Observation 2: Store Interface Separation

**File:** `internal/store/interface.go`, `redis.go`, `postgres.go`

The store package splits functionality across multiple files:
- `interface.go` — Defines interfaces
- `postgres.go` — Postgres implementation (large, ~1400 lines)
- `sqlite.go` — SQLite implementation (~260 lines)
- `redis.go` — Redis caching layer (~235 lines)

This separation is clean and follows Go best practices. However, the Postgres file is very large and could benefit from further splitting (e.g., by domain: user.go, elderly.go, device.go, alert.go).

### Observation 3: Middleware Ordering and Composition

The router uses a composite pattern:
```go
protected := r.Group("/api/v1")
protected.Use(auth.AuthMiddleware())
protected.Use(rateLimiter.Authenticated())
protected.Use(auth.CSRFCheck())  // CSRF applied AFTER auth
```

**Good practice:** CSRFCheck depends on user ID being set by AuthMiddleware, so order is correct. Rate limiting is applied after auth for authenticated users.

However, note that CSRFCheck skips unauthenticated requests (checks `userID == ""`). This means rate limiting protects the CSRF validation step from DoS, but CSRF itself is relaxed for anonymous paths (correct behavior).

### Observation 4: Use of Context with Timeout in Health Checks

**File:** `/internal/router/router.go:72-114`

The `/api/v1/health/ready` endpoint performs synchronous database, Redis, and NATS calls. Without individual timeouts on each dependency probe, a hung database could cause the health check to block indefinitely, potentially exhausting worker threads.

**Recommendation:** Wrap each dependency check with context timeout (e.g., 500ms) so health check remains responsive even if downstream services are slow.

Current code:
```go
if err := pg.Pool().Ping(c.Request.Context()); err == nil { ... }
```

Should become:
```go
ctx, cancel := c.WithTimeout(context.Background(), 500*millisec)
defer cancel()
if err := pg.Pool().Ping(ctx); err == nil { ... }
```

---

## Compliance with Project Standards (CLAUDE.md)

### Technology Stack Alignment ✅

| Requirement | Status | Notes |
|-------------|--------|-------|
| Language: Go + Gin framework | ✅ | Correctly used |
| Database: PostgreSQL 16.x | ✅ | Main store uses pgx/postgres |
| Authentication: JWT with secure cookies | ✅ | Implemented in Phase 9 |
| License: MIT/BSD/Apache-2.0 only | ⚠️ | Need to verify external deps |

### Security Requirements ❌ Some Items Not Met

CLAUDE.md mandates strict security posture for a healthcare-related platform. Current implementation has gaps:

1. **Missing TLS termination** — All HTTPS/SSL configuration should be documented and enforced. No evidence of HSTS enforcement beyond dev-mode header.
2. **No audit trail for data modifications** — While audit logs exist, health data updates lack detailed change tracking required by medical data regulations.
3. **Encryption at rest not addressed** — PostgreSQL data encryption (TDE) or column-level encryption for PII is not mentioned in code.
4. **GDPR/HIPAA compliance indicators absent** — Right-to-erase flows exist but lack documented retention policies and consent management.

---

## Recommended Action Plan

| Priority | Item | Owner | Target |
|----------|------|-------|--------|
| P0 | Fix UpdateElderlyProfile missing WHERE clause | Backend Lead | Immediate |
| P0 | Review and fix DB_URL default credentials removing password | DevOps | Sprint start |
| P1 | Add Redis pattern validation in DelByPattern | Security Eng | Sprint 2 |
| P1 | Remove hardcoded device-secret fallback, use strict config | Auth Eng | Sprint 2 |
| P1 | Make JWT secret validation stricter (panic on empty in prod) | Config Engineer | Sprint 2 |
| P2 | Add CSRF coverage verification test suite | QA Eng | Sprint 3 |
| P2 | Sanitize user input for elderly profile fields | Frontend/Backend | Sprint 3 |
| P3 | Extract CSRF TTL and other constants to config | Tech Lead | Future |
| P3 | Implement metric filtering in GetHealthTrend | Backend Eng | Sprint 4 |

---

## Conclusion

The Phase 9 httpOnly Cookie + CSRF authentication implementation represents significant progress in securing the API. However, critical bugs remain unfixed, most notably the **UpdateElderlyProfile WHERE clause omission** which could cause mass data corruption. Additional medium-sequence issues around Redis pattern handling, role validation defaults, and config hardening should be addressed before production release.

This review should inform the upcoming sprint planning and priority backlog for Phase 10 Security Hardening.
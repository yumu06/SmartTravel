# Phase 1 Authentication and Database Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade the travel backend database foundation and replace the legacy 24-hour JWT with Redis-backed access/refresh sessions while preserving the existing login contract.

**Architecture:** Keep the Gin/GORM modular monolith. Add focused configuration, Redis session, and token-manager units; controllers depend on an authentication service, while the existing middleware continues to load the authenticated user from MySQL. Database tuning and schema indexes are added without renaming existing packages.

**Tech Stack:** Go 1.20, Gin 1.9.1, GORM 1.25.9, MySQL 8, Redis 7, `github.com/golang-jwt/jwt/v5`, `github.com/redis/go-redis/v9`.

**Spec:** `../../../Ongoing_trip_backend_upgrade_design.md` (Phase 1), plus the approved constraint that Tencent WebServiceAPI is configuration-only during this phase.

## Global Constraints

- Access token TTL is 15 minutes.
- Refresh token TTL is 7 days (168 hours).
- JWT signing secrets come from `JWT_ACCESS_SECRET` and `JWT_REFRESH_SECRET`; no JWT secret is hard-coded.
- Existing `POST /travel/login` clients continue to receive a top-level `token` containing the access token.
- Tencent WebServiceAPI is stored as configuration only and is not called in Phase 1.
- The project remains a modular monolith.

---

### Task 1: Configuration and database foundation

**Files:**
- Modify: `config/Init.go`
- Modify: `config/application.yml`
- Modify: `TravelDate/tavelDate.go`
- Modify: `main.go`
- Modify: `TravelModel/Tuser.go`
- Modify: `TravelModel/Tpost.go`
- Create: `migrations/001_phase1_indexes.sql`
- Test: `config/Init_test.go`
- Test: `TravelDate/tavelDate_test.go`

**Interfaces:**
- Produces: `config.InitConfig() error`, environment-overridable Viper keys, and `TravelDate.ConfigurePool(*sql.DB, PoolConfig)`.
- Consumes: existing application YAML and GORM MySQL initialization.

- [ ] **Step 1: Write failing configuration tests**

  Verify environment variables override secrets and that a missing config file returns an error instead of panicking.

- [ ] **Step 2: Run the focused tests and verify RED**

  Run: `go test ./config ./TravelDate -run 'Test(InitConfig|ConfigurePool)' -v`
  Expected: compile failure because the new APIs do not exist.

- [ ] **Step 3: Implement the minimal configuration and pool APIs**

  Use `viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))`, `viper.AutomaticEnv()`, return initialization errors, configure max-open 50, max-idle 10, lifetime 30 minutes, idle time 10 minutes, and GORM slow-query threshold 200 ms.

- [ ] **Step 4: Add index migration and model constraints**

  Add unique indexes for OpenID, article favorites, and foot favorites; add the documented post/user, comment/post, foot/user, and bidirectional chat indexes using idempotent migration SQL.

- [ ] **Step 5: Run focused and full tests**

  Run: `go test ./config ./TravelDate -v` then `go test ./...`.
  Expected: PASS.

### Task 2: Token manager

**Files:**
- Replace: `pkg/jwt/jwt.go`
- Create: `pkg/jwt/jwt_test.go`

**Interfaces:**
- Produces: `NewManager(Config) (*Manager, error)`, `IssuePair(userID uint64) (Pair, error)`, `ParseAccess(string) (*Claims, error)`, and `ParseRefresh(string) (*Claims, error)`.
- Claims: subject user ID, JWT ID, session ID, token type, issuer, issued-at, and expiry.

- [ ] **Step 1: Write failing token behavior tests**

  Cover access/refresh claim types, exact configured TTLs with tolerance, wrong token type rejection, wrong secret rejection, and missing-secret validation.

- [ ] **Step 2: Verify RED**

  Run: `go test ./pkg/jwt -v`.
  Expected: compile failure because `Manager` does not exist.

- [ ] **Step 3: Implement minimal JWT v5 manager**

  Accept only HMAC signing, validate issuer and token type, and generate independent JTI values sharing one SID.

- [ ] **Step 4: Verify GREEN**

  Run: `go test ./pkg/jwt -v` then `go test ./...`.
  Expected: PASS.

### Task 3: Redis refresh-session store

**Files:**
- Create: `cache/redis.go`
- Create: `auth/session.go`
- Create: `auth/redis_session.go`
- Create: `auth/redis_session_test.go`

**Interfaces:**
- Produces: `SessionStore` with `Create`, `Rotate`, `Delete`, `Blacklist`, and `IsBlacklisted`; `RedisSessionStore` uses keys `travel:auth:refresh:{sid}` and `travel:auth:blacklist:{jti}`.
- Consumes: `redis.UniversalClient` so tests can use an in-memory Redis server.

- [ ] **Step 1: Write failing session lifecycle tests**

  Cover create, successful one-time rotation, replay rejection with session deletion, logout deletion, blacklist TTL, and Redis errors.

- [ ] **Step 2: Verify RED**

  Run: `go test ./auth -run TestRedisSession -v`.
  Expected: compile failure because the session store does not exist.

- [ ] **Step 3: Implement the store with atomic Lua rotation**

  Store user ID and refresh JTI as JSON with a seven-day TTL; rotation compares the old JTI and atomically replaces it, while mismatch deletes the session and reports replay.

- [ ] **Step 4: Verify GREEN**

  Run: `go test ./auth -v` then `go test ./...`.
  Expected: PASS.

### Task 4: Authentication service and HTTP endpoints

**Files:**
- Create: `auth/service.go`
- Create: `auth/service_test.go`
- Modify: `controller/user.go`
- Modify: `router/router.go`
- Modify: `middleware/AuthMiddleqware.go`
- Modify: `vo/userRequest.go`
- Modify: `main.go`
- Test: `controller/auth_test.go`

**Interfaces:**
- Produces: `Service.StartSession`, `Service.Refresh`, `Service.Logout`; `POST /travel/auth/refresh`; authenticated `POST /travel/auth/logout`.
- Consumes: JWT manager and session store.
- Compatibility: login returns both legacy `token` and structured `data.access_token`, `data.refresh_token`, and expiry values.

- [ ] **Step 1: Write failing service and handler tests**

  Cover login pair creation, refresh rotation, replay invalidation, logout session deletion plus access blacklist, malformed request 400, invalid refresh 401, and successful response shape.

- [ ] **Step 2: Verify RED**

  Run: `go test ./auth ./controller -v`.
  Expected: compile failure because service and handlers do not exist.

- [ ] **Step 3: Implement the service and dependency wiring**

  Keep WeChat identity lookup unchanged; replace legacy token issuance after user lookup, register refresh/logout routes, and initialize Redis/auth dependencies in `main`.

- [ ] **Step 4: Update middleware validation**

  Parse access tokens only, reject blacklisted JTI values, then preserve the existing MySQL user lookup and `authInfo` context contract.

- [ ] **Step 5: Verify GREEN and compatibility**

  Run: `go test ./auth ./controller ./middleware ./router -v` then `go test ./...`.
  Expected: PASS with the old top-level login token still present.

### Task 5: Final quality verification

**Files:**
- Modify only files required by formatting or discovered defects.

**Interfaces:**
- Consumes all Phase 1 deliverables.
- Produces a buildable, race-checked backend.

- [ ] **Step 1: Format all changed Go files**

  Run: `gofmt -w` on changed Go files.

- [ ] **Step 2: Run static checks**

  Run: `go vet ./...`.
  Expected: exit 0.

- [ ] **Step 3: Run tests with race detection**

  Run: `go test -race ./...`.
  Expected: exit 0 with zero failures.

- [ ] **Step 4: Build the service**

  Run: `go build ./...`.
  Expected: exit 0.

- [ ] **Step 5: Review requirements**

  Confirm secrets are not hard-coded, WebServiceAPI has no caller, all new routes are registered, the legacy login token remains, and each migration matches an actual query pattern.

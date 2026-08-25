# Core Business Performance Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an isolated benchmark database and authenticated load-test workflow, then record reproducible baselines for authorization, user information, article detail, and recommendation endpoints.

**Architecture:** Run the existing application on port 1017 with `MYSQL_DATABASE=travel_benchmark` and benchmark-only JWT secrets while leaving port 1016 and `travel_database` untouched. A deterministic dataset builder seeds the isolated database; a token command creates a real Redis-backed session; the existing Go load runner gains Bearer-header and URL-pool support.

**Tech Stack:** Go 1.20+, Gin, GORM, MySQL 8, Redis 7, Go standard-library HTTP load generator.

**Spec:** `docs/superpowers/specs/2026-08-22-core-business-performance-design.md`

## Global Constraints

- Never drop, truncate, or seed `travel_database`.
- The only database the setup command may create or reset is exactly `travel_benchmark`.
- Keep the existing service on port 1016 running; benchmark service uses port 1017.
- Do not call WeChat or Tencent APIs during load tests.
- Do not persist AppSecret, database passwords, JWT secrets, Access Tokens, Refresh Tokens, or SessionKey values in source or reports.
- Baseline stages stop increasing concurrency when error rate reaches 1%.

---

### Task 1: Authenticated multi-URL load runner

**Files:**
- Modify: `benchmark/loadtest/loadtest.go`
- Modify: `benchmark/loadtest/loadtest_test.go`
- Modify: `cmd/loadtest/main.go`

**Interfaces:**
- Consumes: `loadtest.Config` and the current GET load-test behavior.
- Produces: `Config.URLs []string`, `Config.Headers map[string]string`; requests rotate through URLs and copy configured headers; CLI accepts `-urls-file` and reads `BENCHMARK_ACCESS_TOKEN` without printing it.

- [ ] **Step 1: Write failing URL rotation and Authorization tests**

```go
func TestRunRotatesURLsAndSendsConfiguredHeader(t *testing.T) {
    // Two httptest paths each increment a counter and reject a missing Bearer value.
    // Run four requests with concurrency one.
    // Assert two calls per path and zero summarized errors.
}
```

- [ ] **Step 2: Verify RED**

Run: `go test ./benchmark/loadtest -run TestRunRotatesURLsAndSendsConfiguredHeader -v`

Expected: compile failure because `Config.URLs` and `Config.Headers` do not exist.

- [ ] **Step 3: Implement URL pool and headers**

Use `cfg.URLs[index%len(cfg.URLs)]` when URLs are present and preserve `cfg.URL` as the one-URL compatibility path. Clone headers onto every request before `client.Do`.

- [ ] **Step 4: Add CLI file loading**

`-urls-file` contains one absolute URL per non-empty line. If `BENCHMARK_ACCESS_TOKEN` is non-empty, configure `Authorization: Bearer <token>` internally; never echo the header.

- [ ] **Step 5: Verify GREEN**

Run: `go test ./benchmark/loadtest -v`

Expected: all load-runner behavior tests pass.

### Task 2: Deterministic benchmark dataset

**Files:**
- Create: `benchmark/seed/dataset.go`
- Create: `benchmark/seed/dataset_test.go`
- Create: `benchmark/seed/mysql.go`
- Create: `cmd/benchmark-setup/main.go`

**Interfaces:**
- Produces: `seed.Scale`, `seed.Dataset`, `seed.Build(Scale)`, `seed.ResetAndSeed(ctx, adminDSN, databaseName, Dataset)`, and `benchmark/post_ids.txt` containing 100 article IDs.
- Consumes: existing `TravelModel` types and GORM MySQL driver.

- [ ] **Step 1: Write failing deterministic-count tests**

```go
func TestBuildCreatesRequestedCountsWithoutDuplicateRelations(t *testing.T) {
    data := Build(Scale{Users: 10, Posts: 100, Comments: 200, Likes: 300, Favorites: 100, Foots: 20})
    // Assert exact slice lengths and unique user/post pairs for likes and favorites.
}
```

- [ ] **Step 2: Verify RED**

Run: `go test ./benchmark/seed -v`

Expected: compile failure because `Build` and `Scale` do not exist.

- [ ] **Step 3: Implement the in-memory dataset builder**

Use user IDs `1..N`; generate UUIDs for post, comment, and foot records; distribute timestamps over seven days; generate 100 unique likes and 30 unique favorites per user until requested counts are reached.

- [ ] **Step 4: Implement exact-name database safety**

`ResetAndSeed` must reject any `databaseName` other than `travel_benchmark` before opening an administrative connection. It creates the database when missing, drops only benchmark tables in foreign-key-safe order, runs `AutoMigrate`, inserts batches of 500, and verifies every final row count.

- [ ] **Step 5: Add setup command**

Read MySQL user/password from `BENCHMARK_MYSQL_USER` and `BENCHMARK_MYSQL_PASSWORD`. Seed 1,000 users, 10,000 posts, 50,000 comments, 100,000 likes, 30,000 favorites, and 10,000 foots. Write only article UUIDs to `benchmark/post_ids.txt`.

- [ ] **Step 6: Verify unit tests and setup**

Run: `go test ./benchmark/seed -v`, then execute the setup command with local benchmark credentials.

Expected: exact counts are printed for `travel_benchmark`; existing `travel_database` remains present.

### Task 3: Benchmark token provisioning

**Files:**
- Create: `cmd/benchmark-token/main.go`
- Create: `cmd/benchmark-token/main_test.go`

**Interfaces:**
- Produces: a command that accepts benchmark user ID 1, signs with environment-only benchmark JWT secrets, creates the Redis refresh session, and prints only the Access Token to stdout.
- Consumes: `pkg/jwt.Manager`, `auth.Service`, `auth.RedisSessionStore`, and Redis settings from environment.

- [ ] **Step 1: Write failing configuration-validation test**

```go
func TestBuildServiceRejectsMissingSecrets(t *testing.T) {
    _, _, err := buildService(config{RedisAddr: "127.0.0.1:6379"})
    if err == nil { t.Fatal("expected missing-secret error") }
}
```

- [ ] **Step 2: Verify RED**

Run: `go test ./cmd/benchmark-token -v`

Expected: compile failure because `buildService` does not exist.

- [ ] **Step 3: Implement token command**

Read `BENCHMARK_JWT_ACCESS_SECRET`, `BENCHMARK_JWT_REFRESH_SECRET`, Redis address/password, and user ID. Use 15-minute Access and 168-hour Refresh TTLs with issuer `ongoing-trip-benchmark`.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./cmd/benchmark-token -v`.

Expected: missing configuration is rejected without contacting Redis.

### Task 4: Isolated benchmark service and correctness checks

**Files:**
- Modify only when a correctness failure exposes a production bug; every such change requires a failing regression test first.

**Interfaces:**
- Consumes: seeded `travel_benchmark`, benchmark JWT environment, Redis, and benchmark token command.
- Produces: a running server on port 1017 and a process-local `BENCHMARK_ACCESS_TOKEN` used by the load runner.

- [ ] **Step 1: Start the benchmark server**

Set `SERVER_PORT=1017`, `MYSQL_DATABASE=travel_benchmark`, benchmark JWT secrets, and existing Redis/MySQL credentials; run the current application without stopping port 1016.

- [ ] **Step 2: Create benchmark token**

Run the token command with the same JWT/Redis environment and store stdout in a PowerShell process variable, then set `BENCHMARK_ACCESS_TOKEN` for load-test commands.

- [ ] **Step 3: Verify one request per scenario**

Expected HTTP 200 for authorization, user info, one article detail ID from `benchmark/post_ids.txt`, and recommendation limit 20. Stop if any response is non-2xx.

### Task 5: Core endpoint baselines

**Files:**
- Create: `benchmark/2026-08-22-core-business-baseline.md`

**Interfaces:**
- Consumes: port 1017, Access Token, post ID pool, and load runner.
- Produces: measured tables for the four core endpoints with concurrency 10, 50, 100, and 200 where error rate remains below 1%.

- [ ] **Step 1: Run authorization stages**

Use 500 requests at concurrency 10 and 5,000 requests for higher stages.

- [ ] **Step 2: Run user-information stages**

Use the same request counts and stop rule.

- [ ] **Step 3: Run article-detail stages**

Create an absolute URL file from the 100 article IDs and rotate through it.

- [ ] **Step 4: Run recommendation stages**

Use `limit=20` and the same stop rule.

- [ ] **Step 5: Record evidence**

Record commands, successful QPS, errors, P50/P95/P99/max, slow SQL, and whether 1017 remained healthy. Do not copy tokens or secrets.

### Task 6: Verification and optimization handoff

**Files:**
- Modify: `benchmark/2026-08-22-core-business-baseline.md`

**Interfaces:**
- Consumes: all benchmark tooling and baseline evidence.
- Produces: a ranked bottleneck list and one follow-up optimization plan per independent bottleneck.

- [ ] **Step 1: Run project verification**

Run: `gofmt` on changed Go files, `go test -count=1 ./...`, `go vet ./...`, and `go build ./...`.

- [ ] **Step 2: Confirm isolation**

Verify 1016 still responds and `travel_database` was not modified by setup commands.

- [ ] **Step 3: Rank bottlenecks**

Rank only observed issues such as auth MySQL lookup, article read/write amplification, recommendation sort cost, logging, or connection saturation. Do not implement optimizations without a baseline comparison target.

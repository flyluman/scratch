# Scratch — Go Hexagonal CRUD Skeleton

Production-grade **Go hexagonal CRUD backend** with soft delete, cursor pagination, audit trail, Valkey caching, JWT auth, OpenTelemetry tracing, and auto-generated OpenAPI docs.

**Entity:** `Item` (ID, Name, Description, CreatedAt, UpdatedAt)

---

## Quick Start

```bash
# 1. Rename module to your own
make rename MODULE=github.com/yourname/yourproject
go mod tidy

# 2. Start infrastructure
docker compose -f infra/local/docker-compose.yml up -d

# 3. Run
go run ./cmd/server

# 4. Verify
curl http://localhost:8080/livez
curl -X POST http://localhost:8080/v1/items -d '{"name":"test","description":"hello"}'
```

---

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                    cmd/server/                       │
│                  main.go (entry + DI)                │
├──────────────────────────────────────────────────────┤
│                    internal/                         │
│                                                      │
│  ┌──────────┐    ┌──────────────┐    ┌────────────┐ │
│  │  domain  │───▶│ application  │◀───│  adapters  │ │
│  │ (entity) │    │  (use cases) │    │ (in/out)   │ │
│  └──────────┘    └──────────────┘    └────────────┘ │
│                       │                    ▲         │
│                       │ via ports         │         │
│                       ▼                    │         │
│                  ┌──────────────────┐      │         │
│                  │     ports        │──────┘         │
│                  │  (interfaces)    │                │
│                  └──────────────────┘                │
│                       │                    ▲         │
│               ┌───────┴───────┐    ┌───────┴───────┐ │
│               │   provider/   │    │  platform/    │ │
│               │  id (UUID)    │    │ config │logger│ │
│               │  clock (Time) │    │ telemetry     │ │
│               │  repo (Fctry) │    │ featureenv    │ │
│               └───────────────┘    │ migrations    │ │
│                                    │ bus (events)  │ │
│                                    └───────────────┘ │
└──────────────────────────────────────────────────────┘
```

**Layers:**

| Layer | Package | Responsibility |
|---|---|---|
| **Domain** | `internal/domain` | Entities, validation, sentinel errors. Zero dependencies. |
| **Application** | `internal/application` | Use cases grouped in `ItemModule`. Publish events via `ports.EventBus` instead of calling adapters directly. |
| **Ports** | `internal/ports` | Interfaces: `ItemRepository`, `IDGenerator`, `Clock`, `Cache`, `Logger`, `AuditLogger`, `TokenValidator`, `FeatureFlags`, `EventBus`, `BaseRepository[T,ID]`. |
| **Adapters (Inbound)** | `internal/adapters/inbound/http` | HTTP handlers, middleware, JSON envelope, swagger, error mapping. Sub-handler per entity (`item/`, `user/`). |
| **Adapters (Outbound)** | `internal/adapters/outbound` | Port implementations: `memory/`, `postgres/`, `valkey/`, `audit/`, `auth/`, `feature/` (Valkey). |
| **Provider** | `internal/provider` | Zero-dependency port impls: `id/` (UUID), `clock/` (time), `repo/` (factory switching memory ↔ postgres). |
| **Platform** | `internal/platform` | Infrastructure: `config/`, `logger/` (zap adapter), `telemetry/`, `featureenv/`, `migrations/`, `bus/` (in-memory event bus). |

---

## Directory Structure

```
├── cmd/
│   └── server/               # Binary entry point
│       ├── main.go           # main(), InitializeApp(), App struct, App.Start()
│       └── swagger_docs.go   # OpenAPI annotations
├── migrations/               # SQL migration files only
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── internal/
│   ├── domain/               # Item entity + Validate()
│   ├── application/          # ItemModule (5 use cases), publish events to EventBus
│   ├── ports/                # Interfaces: repository, cache, logger, auth, audit, feature, event, id
│   ├── adapters/
│   │   ├── inbound/http/     # Main handler + RouteRegistrar interface
│   │   │   └── item/         # Item sub-handler (routes + CRUD + tests)
│   │   └── outbound/
│   │       ├── memory/       # In-memory repo (dev/test)
│   │       ├── postgres/     # Postgres repo + pool
│   │       ├── valkey/       # Valkey client + item cache + cached repo
│   │       ├── audit/        # Postgres audit logger + NoOp + EventHandler subscriber
│   │       ├── auth/         # JWT validator (JWKS)
│   │       └── feature/      # Feature flags (Valkey — client-side caching)
│   ├── provider/
│   │   ├── id/               # UUID generator (no external deps)
│   │   ├── clock/            # Real clock (no external deps)
│   │   └── repo/             # Repo factory (memory ↔ postgres switch)
│   └── platform/
│       ├── config/           # Env-based Config struct
│       ├── logger/           # Zap adapter implementing ports.Logger
│       ├── telemetry/        # OpenTelemetry tracing (OTLP gRPC)
│       ├── featureenv/       # Feature flags fallback (env vars)
│       ├── bus/              # In-memory event bus
│       └── migrations/       # Migration runner + pool helper
├── infra/
│   ├── docker/Dockerfile     # Multi-stage build
│   ├── local/docker-compose.yml  # Postgres 18 + Valkey 8 + app
│   └── k8s/base/             # Deployment, Service, ConfigMap
├── .env                      # Local env vars (gitignored)
├── .env.example              # Env var template
├── docs/                     # Auto-generated by swaggo
├── .dockerignore
├── .golangci.yml
├── .githooks/pre-commit      # make fmt lint vet
├── Makefile
└── README.md
```

---

## Configuration (Environment Variables)

All config via env vars with `SCRATCH_` prefix. See `internal/platform/config/config.go`.

| Variable | Default | Description |
|---|---|---|
| `SCRATCH_APP_ENV` | `dev` | Environment name |
| `SCRATCH_HTTP_ADDR` | `:8080` | Listen address |
| `SCRATCH_LOG_LEVEL` | `info` | Log level |
| **Postgres** | | |
| `SCRATCH_POSTGRES_DSN` | `postgres://postgres:postgres@localhost:5432/scratch?sslmode=disable` | Postgres connection string |
| `SCRATCH_RUN_MIGRATIONS` | `true` | Auto-run migrations on startup |
| **Valkey** | | |
| `SCRATCH_VALKEY_ADDR` | `localhost:6379` | Valkey address |
| `SCRATCH_VALKEY_USERNAME` | `` | Valkey username |
| `SCRATCH_VALKEY_PASSWORD` | `` | Valkey password |
| `SCRATCH_VALKEY_DB` | `0` | Valkey DB number |
| **Auth** | | |
| `SCRATCH_AUTH_ENABLED` | `false` | Enable JWT auth |
| `SCRATCH_AUTH_JWKS_ENDPOINT` | `` | JWKS endpoint URL |
| `SCRATCH_AUTH_JWKS_REFRESH_MIN` | `10` | JWKS cache refresh interval (min) |
| **Rate Limiting** | | |
| `SCRATCH_RATE_LIMIT_PER_SECOND` | `0` | Rate limit per IP (0 = disabled) |
| `SCRATCH_RATE_LIMIT_TTL_SEC` | `300` | Rate limiter cleanup interval |
| **Telemetry** | | |
| `SCRATCH_TELEMETRY_ENABLED` | `false` | Enable OpenTelemetry |
| `SCRATCH_TELEMETRY_OTLP_ENDPOINT` | `` | OTLP gRPC endpoint |
| **TLS** | | |
| `SCRATCH_TLS_ENABLED` | `false` | Enable TLS |
| `SCRATCH_TLS_CERT_FILE` | `` | TLS certificate path |
| `SCRATCH_TLS_KEY_FILE` | `` | TLS key path |
| **CORS** | | |
| `SCRATCH_CORS_ORIGINS` | `*` | Comma-separated origins |
| **Feature Flags** | | |
| `SCRATCH_FEATURE_CACHE_TTL_SEC` | `30` | Feature flag cache TTL (client-side) |
| **Feature Flag Overrides** (fallback when Valkey unavailable) | | |
| `SCRATCH_FEATURE_SOFT_DELETE` | `true` | Soft delete enabled |
| `SCRATCH_FEATURE_AUDIT` | `false` | Audit logging enabled |
| `SCRATCH_FEATURE_AUTH` | `false` | Auth enabled |
| `SCRATCH_FEATURE_TELEMETRY` | `false` | Telemetry enabled |
| `SCRATCH_FEATURE_VALKEY_CACHING` | `true` | Valkey caching enabled |

---

## API

### Health

```
GET /livez           → 204 No Content (liveness)
GET /readyz          → 200 or 503 (readiness — pings Postgres + Valkey)
```

### Scalar UI

```
GET /scalar       → Scalar UI (auto-generated)
```

### Items (CRUD)

All under `/v1`. Protected by JWT auth when `SCRATCH_AUTH_ENABLED=true`.

```
POST   /v1/items              Create item
GET    /v1/items              List items (cursor pagination: ?cursor=&limit=20)
GET    /v1/items/:id          Get item
PUT    /v1/items/:id          Update item
DELETE /v1/items/:id          Delete item (soft or hard, controlled by feature flag)
```

### Response Envelope

```json
// Success
{"success": true, "data": { ... }}

// Error
{"success": false, "error": {"code": "NOT_FOUND", "message": "resource not found"}}
```

### Pagination

```json
// Response
{"success": true, "data": {
  "items": [...],
  "next_cursor": "abc123:2026-06-15T10:00:00Z"
}}

// Next request
GET /v1/items?cursor=abc123:2026-06-15T10:00:00Z&limit=20
```

Cursor format: `{last_id}:{last_created_at_rfc3339}`. Composite comparison `(created_at, id)` for stable ordering.

---

## Feature Flags (Valkey-backed)

Feature flags are stored in Valkey hash `feature:flags:{env}`. Each app instance reads via `DoCache` with client-side caching (RESP3 tracking). When flags change in Valkey, all instances are invalidated automatically.

**Admin — toggle a flag:**
```bash
redis-cli HSET feature:flags:dev soft_delete false
# → Client receives invalidation → next read hits Valkey → fresh value
```

**Fallback:** When Valkey unavailable, reads from env vars (`SCRATCH_FEATURE_*`).

**Auto-seed:** On first startup, creates the hash with defaults via `HSETNX` (no-race — doesn't overwrite existing values).

**Cache TTL:** Configurable via `SCRATCH_FEATURE_CACHE_TTL_SEC`. If Valkey disconnects, stale cache served until TTL expiry, then falls back to env vars.

---

## Audit Trail

All create/update/delete operations are logged to `audit_events` table.

```
audit_events
├── id            TEXT (UUID)
├── action        TEXT (CREATE / UPDATE / DELETE)
├── resource_type TEXT (item)
├── resource_id   TEXT
├── actor_id      TEXT (from JWT claims.Subject, empty if anonymous)
├── old_value     JSONB (previous state)
├── new_value     JSONB (new state)
└── created_at    TIMESTAMPTZ
Indexes: BRIN(created_at) for time-series, BTREE(resource_type, resource_id) for lookups
```

When Postgres is unavailable, audit falls back to `NoOpAuditLogger` (no-op).

---

## Adding a New Module

Each entity follows the same pattern as `Item`. Example: adding `User`.

### 1. Domain — `internal/domain/user.go`

```go
type User struct {
    ID        string
    Name      string
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (u *User) Validate() error { ... }
```

### 2. Ports — `internal/ports/repository.go`

Add interface methods or create a new interface:

```go
type UserRepository interface {
    Save(ctx, user) error
    Get(ctx, id) (User, error)
    List(ctx, cursor, limit) ([]User, string, error)
    Update(ctx, user) error
    Delete(ctx, id) error
    SoftDelete(ctx, id) error
}
```

### 3. Application — `internal/application/user_module.go`

```go
type UserModule struct {
    Create *CreateUserUseCase
    Get    *GetUserUseCase
    List   *ListUserUseCase
    Update *UpdateUserUseCase
    Delete *DeleteUserUseCase
}

func NewUserModule(repo UserRepository, idGen IDGenerator, clock Clock, bus EventBus, flags FeatureFlags) *UserModule { ... }
```

Create individual use case files: `create_user.go`, `get_user.go`, `list_users.go`, `update_user.go`, `delete_user.go`.

### 4. Inbound Adapter — `internal/adapters/inbound/http/user/`

Create 7 files:

```
internal/adapters/inbound/http/user/
├── handler.go      # Handler struct + RegisterRoutes + NewHandler(*application.UserModule)
├── create.go       # createUser handler + CreateUserRequest
├── get.go          # getUser handler
├── list.go         # listUsers handler
├── update.go       # updateUser handler + UpdateUserRequest
├── delete.go       # deleteUser handler
├── response.go     # UserResponse + toUserResponse()
└── handler_test.go # Tests
```

`Handler` implements `httpapi.RouteRegistrar` — register routes on the `/v1` group.

### 5. Outbound Adapters — Repo implementations

Add `NewUserRepository` methods to each adapter:

| File | Change |
|---|---|
| `internal/adapters/outbound/memory/memory.go` | Add `UserRepository` struct + CRUD methods |
| `internal/adapters/outbound/postgres/user_repo.go` | Add `UserRepository` struct + Postgres queries |
| `internal/provider/repo/factory.go` | Add `NewUserRepository()` method |

### 6. Migration — `migrations/002_users.up.sql`

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 7. Wiring — `cmd/server/builder.go`

Add a method to `AppBuilder`:

```go
func (b *AppBuilder) WithUser() *AppBuilder {
    userMod := application.NewUserModule(
        b.repos.NewUserRepository(), b.idGen, b.clk, b.bus, b.flags,
    )
    b.handlers = append(b.handlers, user.NewHandler(userMod))
    return b
}
```

Then chain it in `InitializeApp()`:

```go
return NewAppBuilder(cfg, log, tel, pool, flags, auditLogger, tokenValidator).
    WithItem().
    WithUser().
    Build()
```

No changes to `main.go` or `handler.go` needed.

---

## Testing

```bash
make test                  # Unit tests (short mode, 9 handler tests)
make test-integration      # Integration tests (requires Docker)
```

Integration tests use `testcontainers-go` to spin up Postgres + Valkey containers.

Test layers:

| Layer | Tool | Test files |
|---|---|---|
| Handler | `httptest` + in-memory repo | `internal/adapters/inbound/http/handler_test.go` |
| Postgres repo | `testcontainers` | `internal/adapters/outbound/postgres/item_repo_test.go` |
| Valkey cached repo | `testcontainers` | `internal/adapters/outbound/valkey/cached_repo_test.go` |

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/labstack/echo/v4` | HTTP framework |
| `github.com/jackc/pgx/v5` | Postgres driver |
| `github.com/valkey-io/valkey-go` | Valkey client (auto-pipelining, client-side caching) |
| `github.com/sony/gobreaker` | Circuit breaker for Valkey cache |
| `go.opentelemetry.io/otel` | OpenTelemetry SDK (traces) |
| `go.uber.org/zap` | Structured logging |
| `github.com/golang-jwt/jwt/v5` | JWT parsing |
| `github.com/lestrrat-go/jwx` | JWKS fetching |
| `github.com/google/uuid` | UUID generation |
| `github.com/swaggo/echo-swagger` | Auto-generated OpenAPI spec |
| `golang.org/x/time` | Rate limiting (via Echo middleware) |
| `github.com/testcontainers/testcontainers-go` | Integration test containers |

---

## Deployment

### Docker

```bash
make docker-build   # Builds multi-stage Docker image
make docker-up      # Starts Postgres + Valkey + app via docker compose
```

### Kubernetes

Manifests in `infra/k8s/base/`:
- `deployment.yaml` — 2 replicas, liveness/readiness probes, resource limits
- `service.yaml` — ClusterIP on port 80
- `configmap.yaml` — env var config

```bash
kubectl apply -f infra/k8s/base/
```

### Binary

```bash
make build
./bin/scratch
./bin/scratch -mode=migrate   # Run migrations only
```

---

## Makefile Targets

```bash
make all              # fmt → vet → lint → tidy → build → test
make build            # go build (with -ldflags="-s -w")
make run              # go run
make test             # Unit tests
make lint             # golangci-lint
make vet              # go vet
make fmt              # go fmt
make swagger          # Re-generate OpenAPI spec
make migrate          # Run migrations only
make generate         # swagger (alias)
make docker-build     # Docker image
make docker-up        # docker compose up
make docker-down      # docker compose down
make rename MODULE=github.com/you/yourproject  # Rename module across all files
make clean            # Remove build artifacts
```

---

## Migrations

Migrations run **automatically on every startup** (idempotent via `CREATE TABLE IF NOT EXISTS` + `schema_migrations` tracking). SQL files read from `migrations/` at runtime; Go code in `internal/platform/migrations/`.

Initial migration `001_init` creates:
- `items` table (with auto-update trigger on `updated_at`)
- `audit_events` table (time-series with BRIN index)
- `schema_migrations` table (version tracking)

To add a new migration, create `migrations/002_*.up.sql` and `migrations/002_*.down.sql`.

---

## Design Decisions

| Decision | Rationale |
|---|---|
| **Manual DI** (no Wire) | Fewer dependencies, easier to follow, no code generation step |
| **Valkey client-side caching** | Zero-polling invalidation via RESP3 tracking, near-zero read latency |
| **Echo built-in rate limiter** | Production-tested |
| **swaggo for OpenAPI** | Auto-generated from code annotations, no hand-written spec to drift |
| **BRIN index on audit** | 100x smaller than B-tree for append-only audit data, perfect for range scans |
| **Soft delete by default** | Recoverable deletes, controlled by feature flag |
| **Cursor pagination** | Stable under writes, no offset drift, works at any scale |
| **Time-series audit** | JSONB old/new values for schema flexibility, BRIN for query perf |
| **ports.Cache interface** | Cache implementation swappable (valkey ↔ in-memory ↔ nop) without touching business logic |
| **ports.Logger interface** | Logging framework swappable (zap ↔ zerolog ↔ slog) without touching business logic |
| **FeatureFlags interface** | Runtime-refreshable, enables A/B testing and gradual rollout without restart |
| **BaseRepository[T,ID] generic** | Single interface for basic CRUD across all entities; entity-specific methods stay concrete |
| **Event Bus** | Use cases publish events instead of calling adapters. Adding side-effects (webhook, search index) means adding a subscriber, not modifying use cases |
| **Functional Options** | `NewHandler(cfg, handlers, WithLogger(l)...)` — backward-compatible, optional params, clear intent |
| **AppBuilder** | `InitializeApp()` is a fluent chain. Adding an entity = adding `.WithXxx()` method, zero changes to main.go |

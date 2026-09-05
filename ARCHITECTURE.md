# Mi3ad (ميعاد) — Architecture & Design Notes

## 1. What is Mi3ad?

Mi3ad is a small social scheduling network built with **Go**, the **Echo** framework (v5), and **GORM**. Users offer recurring weekly time slots (mentorship sessions, coaching calls, catch-ups, office hours), and other users book them. When a slot is booked, cancelled, or coming up, connected users get notified live over a websocket connection instead of refreshing or polling.

The goal is a compact, well-structured backend — big enough to show real patterns (auth, relational data, real-time push), small enough to actually finish, maintain, and explain in an interview.

## 2. Why Go?

- Compiled and statically typed — a lot of bugs get caught before runtime, and the app ships as a single binary with no runtime to install on the server.
- Built-in concurrency (goroutines + channels) maps naturally onto a websocket-heavy app — every connected client runs on its own lightweight goroutine, without the overhead of OS threads.
- Small, simple language surface — easy for anyone reviewing the repo on GitHub to follow the code without ramp-up time.
- Strong standard library (`net/http`, `encoding/json`) keeps third-party dependencies to a minimum.
- Fast compile times and fast execution, which makes iterating during development quick.

## 3. Why Echo (v5)?

- Minimal, unopinionated HTTP framework — a thin layer over `net/http` that doesn't hide what's happening underneath.
- Fast, radix-tree-based routing with low overhead; consistently benchmarks near the top of Go web frameworks.
- Built-in middleware (request logging, recovery, CORS, JWT) covers most of what a small API needs out of the box.
- Clean context API makes request binding, validation, and JSON responses simple to write and simple to read.
- We're on **v5** (Echo's current major version, requires Go 1.25+). Notable v5 differences from the older v4 tutorials you'll find online:
  - Handlers take `*echo.Context` (a struct pointer), not `echo.Context` (an interface).
  - `middleware.RequestLogger()` replaces `middleware.Logger()`.
  - `middleware.CORS(...)` now takes allowed origins as explicit arguments — there's no permissive zero-arg default anymore.
  - `Echo.Logger` is a `*slog.Logger`, not something with a `.Fatal()` method — server startup errors are handled with a plain `if err != nil`.

## 4. Tech stack

| Concern                     | Choice                                                | Why                                                                                                                                        |
| --------------------------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Web framework               | Echo v5                                               | see above                                                                                                                                  |
| Database                    | PostgreSQL                                            | relational data (users ↔ appointments) benefits from real constraints and relations                                                        |
| DB access / ORM             | **GORM** (`gorm.io/gorm` + `gorm.io/driver/postgres`) | chosen over raw `database/sql` for a Prisma-like workflow: models double as schema, `AutoMigrate` applies changes without hand-written SQL |
| Validation                  | `go-playground/validator`                             | struct-tag validation (`validate:"required,email"`), matches request DTOs cleanly                                                          |
| Auth                        | JWT (`golang-jwt/jwt`)                                | stateless, plugs straight into Echo's JWT middleware                                                                                       |
| Real-time                   | `gorilla/websocket`                                   | most mature, most widely used Go websocket library                                                                                         |
| Config                      | `godotenv`                                            | load `.env` locally, real env vars in production                                                                                           |
| Containerization (optional) | Docker                                                | makes the repo trivially runnable for anyone who clones it                                                                                 |

## 5. Current structure

```
mi3ad/
├── cmd/
│   └── api/
│       └── main.go                  // wiring: db connect, AutoMigrate, echo instance, routes, ws hub
├── internal/
│   ├── database/
│   │   ├── database.go              // Connect() + AutoMigrate() — the one place that knows the full schema
│   │   └── models/
│   │       └── user.go              // GORM model: User{ID, Name, Email, Password, CreatedAt}
│   ├── users/
│   │   ├── dto/
│   │   │   └── user.go              // CreateUserRequest (validated), UserResponse
│   │   ├── handler.go                // business logic + GORM calls, no repository layer
│   │   └── router.go                  // registers routes on the echo group
│   ├── appointments/
│   │   └── dto/                      // placeholder — next module
│   ├── common/
│   │   ├── dto/
│   │   │   └── response.go           // pure data: envelopes, ResponseStatus, PagingMeta, ValidationErrorResponse
│   │   └── response/
│   │       └── response.go           // echo-aware helpers: OK, Created, NotFound, ValidationError...
│   ├── ws/
│   │   ├── hub.go                    // connection registry + Notify()
│   │   ├── client.go
│   │   └── router.go                  // GET /ws upgrade endpoint
│   └── config/
│       └── config.go
├── go.mod
├── go.sum
├── .env.example
└── .gitignore
```

Note: there's no `migrations/` folder. That existed briefly (plain SQL files + `golang-migrate`) before we switched to GORM's `AutoMigrate` — see §9.

## 6. Why the repository layer was removed

We're keeping DB calls **inside `handler.go`**, with no separate `repository`/`store` layer:

- The project is intentionally small — a handful of entities with fairly simple queries, not a large domain with complex business rules.
- A repository interface + implementation per entity is real boilerplate for very little payoff at this scale: another file to write, another wiring step in `main.go`, another place a bug can hide.
- Keeping the query next to the logic that uses it means less ceremony and faster iteration — read the function top to bottom and see everything that happens. This holds just as true now that the "query" is a GORM call (`h.db.Create(&user)`) instead of raw SQL.
- Trade-off worth knowing: testing a handler this way means testing against a real (or Dockerized) database rather than mocking a repository. For a project this size that's an acceptable trade.

## 7. DTOs vs. Models — three different jobs, three different places

This project splits "shape of data" into three deliberately separate concerns:

| Package                     | Job                                                                                                                    | Depends on Echo? |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------- |
| `internal/database/models/` | The actual DB schema — what GORM reads to run `AutoMigrate`                                                            | No               |
| `internal/{feature}/dto/`   | The public API request/response shape for that feature                                                                 | No               |
| `internal/common/dto/`      | Shared response envelopes (`DataResponse`, `PagingMeta`, error shapes)                                                 | No               |
| `internal/common/response/` | Ergonomic functions (`OK`, `Created`, `NotFound`...) that pick the right envelope _and_ the right HTTP status together | Yes              |

Why models and DTOs are separate structs (not one struct doing both jobs): a GORM model needs `gorm` tags for schema (`primaryKey`, `not null`, `uniqueIndex`), while a response DTO needs `json` tags and should never accidentally expose a column you didn't mean to return (like `Password`). Renaming a DB column shouldn't silently rename a field in your public API, and vice versa.

Why `dto` and `response` are two files instead of one: `dto` has zero framework imports, so it's testable and reusable outside an HTTP handler. `response` exists specifically to make status-code/envelope mismatches impossible — `response.Created()` can only ever produce a success-shaped body at 201, `response.NotFound()` can only ever produce an error-shaped body at 404.

## 8. Websocket (external/real-time service)

- Lives in its own top-level package, `internal/ws/`, not nested inside any single feature.
- `ws.Hub` exposes a small interface, e.g. `Notify(userID string, event any)`. Feature handlers (like `appointments/handler.go`) call it directly right after a successful DB write to push a live update.
- Wired together once in `main.go` — so `appointments` never has to import `ws` directly, and there's no risk of circular imports as more features are added.

## 9. Database & migrations: GORM AutoMigrate

We use `database.AutoMigrate(db)` (in `internal/database/database.go`) instead of hand-written SQL migration files:

- Closer to the Prisma workflow: models _are_ the schema, and starting the app keeps the database in sync automatically.
- Every model gets added to one list in `database.AutoMigrate`, so there's a single place that knows the full schema.
- Trade-off worth knowing: `AutoMigrate` is **additive-only** — it creates tables/columns/indexes it doesn't see, but it never drops or renames anything. Rename a struct field and GORM adds a new column alongside the old one rather than renaming it; the old column has to be cleaned up by hand. This is the main thing Prisma's migration engine does better — accept it as a known limitation of this simpler approach rather than a bug.

## 10. Summary

| Decision          | Choice                                                                                     |
| ----------------- | ------------------------------------------------------------------------------------------ |
| Language          | Go                                                                                         |
| Framework         | Echo v5                                                                                    |
| ORM               | GORM + `gorm.io/driver/postgres`                                                           |
| Schema management | `AutoMigrate`, no manual SQL migrations                                                    |
| Validation        | `go-playground/validator` on request DTOs                                                  |
| Layer split       | Router → Handler (business logic + GORM calls)                                             |
| Repository layer  | Removed — not needed at this scale                                                         |
| Structure         | Feature-based / modular under `internal/`, with a dedicated `database/` package for schema |
| Response shapes   | Split into `common/dto` (pure data) and `common/response` (echo-aware helpers)             |
| Real-time         | Isolated `internal/ws/` package, wired via interface                                       |

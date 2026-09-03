# Mi3ad (ميعاد) — Architecture & Design Notes

## 1. What is Mi3ad?

Mi3ad is a small social scheduling network built with **Go** and the **Echo** framework. Users offer recurring weekly time slots (mentorship sessions, coaching calls, catch-ups, office hours), and other users book them. When a slot is booked, cancelled, or coming up, connected users get notified live over a websocket connection instead of refreshing or polling.

The goal is a compact, well-structured backend — big enough to show real patterns (auth, relational data, real-time push), small enough to actually finish, maintain, and explain in an interview.

## 2. Why Go?

- Compiled and statically typed — a lot of bugs get caught before runtime, and the app ships as a single binary with no runtime to install on the server.
- Built-in concurrency (goroutines + channels) maps naturally onto a websocket-heavy app — every connected client runs on its own lightweight goroutine, without the overhead of OS threads.
- Small, simple language surface — easy for anyone reviewing the repo on GitHub to follow the code without ramp-up time.
- Strong standard library (`net/http`, `database/sql`, `encoding/json`) keeps third-party dependencies to a minimum.
- Fast compile times and fast execution, which makes iterating during development quick.

## 3. Why Echo?

- Minimal, unopinionated HTTP framework — a thin layer over `net/http` that doesn't hide what's happening underneath.
- Fast, radix-tree-based routing with low overhead; consistently benchmarks near the top of Go web frameworks.
- Built-in middleware (logging, recovery, CORS, JWT) covers most of what a small API needs out of the box.
- Clean `echo.Context` API makes request binding, validation, and JSON responses simple to write and simple to read.
- Mature enough that docs and examples are never a problem.

## 4. Tech stack

| Concern | Choice | Why |
|---|---|---|
| Web framework | Echo | see above |
| Database | PostgreSQL | relational data (users ↔ appointments) benefits from real constraints and relations |
| DB access | `database/sql` + `pgx` | direct control, no ORM overhead for a project this size |
| Auth | JWT (`golang-jwt/jwt`) | stateless, plugs straight into Echo's JWT middleware |
| Real-time | `gorilla/websocket` | most mature, most widely used Go websocket library |
| Config | `godotenv` | load `.env` locally, real env vars in production |
| Containerization (optional) | Docker | makes the repo trivially runnable for anyone who clones it |

## 5. Recommended structure: feature-based / modular

```
mi3ad/
├── cmd/
│   └── api/
│       └── main.go              // wiring: db, echo instance, routes, ws hub
├── internal/
│   ├── users/
│   │   ├── dto/
│   │   │   └── user.go          // CreateUserRequest, UserResponse...
│   │   ├── handler.go           // business logic + DB calls
│   │   └── router.go            // registers routes on the echo group
│   ├── appointments/
│   │   ├── dto/
│   │   │   └── appointment.go
│   │   ├── handler.go
│   │   └── router.go
│   ├── common/
│   │   └── dto/
│   │       └── response.go      // shared API response envelope, errors, pagination
│   ├── ws/
│   │   ├── hub.go               // connection registry + broadcast
│   │   ├── client.go
│   │   └── router.go            // GET /ws upgrade endpoint
│   ├── db/
│   │   └── db.go                // connection pool setup
│   └── config/
│       └── config.go
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

**Recommendation: use this structure.** Reasons:

- Organizing by feature (`users`, `appointments`) instead of by technical layer keeps everything about one concept in one place — add a feature, add a folder, done.
- Easy to navigate on GitHub: opening `internal/appointments/` shows the whole feature in three files.
- Scales cleanly if more features get added later (reviews, notifications) without existing folders getting crowded.
- Each module is self-contained enough to read, reason about, or extract on its own.

## 6. Why the repository layer was removed

We're keeping DB calls **inside `handler.go`**, with no separate `repository`/`store` layer:

- The project is intentionally small — a handful of entities with fairly simple queries, not a large domain with complex business rules.
- A repository interface + implementation per entity is real boilerplate for very little payoff at this scale: another file to write, another wiring step in `main.go`, another place a bug can hide.
- Keeping the query next to the logic that uses it means less ceremony and faster iteration — read the function top to bottom and see everything that happens.
- Trade-off worth knowing: testing a handler this way means testing against a real (or Dockerized) database rather than mocking a repository. For a project this size that's an acceptable trade — arguably more meaningful than a mocked test anyway.

## 7. DTOs

- **Feature-specific DTOs** (`CreateUserRequest`, `AppointmentResponse`, etc.) live inside that feature's own `dto/` folder.
- **Global/shared DTOs** — the standard API response envelope, error shape, pagination metadata — live in `internal/common/dto/response.go` and get reused across every handler.

## 8. Websocket (external/real-time service)

- Lives in its own top-level package, `internal/ws/`, not nested inside any single feature.
- `ws.Hub` exposes a small interface, e.g. `Notify(userID string, event any)`. Feature handlers (like `appointments/handler.go`) call it directly right after a successful DB write to push a live update.
- Wired together once in `main.go` — so `appointments` never has to import `ws` directly, and there's no risk of circular imports as more features are added.

## 9. Summary

| Decision | Choice |
|---|---|
| Language | Go |
| Framework | Echo |
| Layer split | Router → Handler (business logic + DB) |
| Repository layer | Removed — not needed at this scale |
| Structure | Feature-based / modular under `internal/` |
| Global DTOs | `internal/common/dto/` |
| Real-time | Isolated `internal/ws/` package, wired via interface |

# B0 Implementation Sequence

Status: proposal. Starts only after D1 and D2 are approved. The plan assumes the Go option. Each row is one commit and is pushed on its own after its verification passes.

| # | Commit | Verification |
|---|---|---|
| 1 | `chore(api): align Go version and commit go.sum` | `go mod verify`, `go build ./...`, `go vet ./...` |
| 2 | `ci(api): pin Go version and run build, vet, test` | CI green on the branch |
| 3 | `feat(api): typed config with required-secret validation` | Unit tests for missing and weak secrets |
| 4 | `feat(api): http server with graceful shutdown and /v1/health` | Test starts server, calls health, stops cleanly |
| 5 | `feat(api): error envelope, trace id and request logging middleware` | Handler tests for error shape and header echo |
| 6 | `feat(db): connection pool and readiness check` | Integration test against compose Postgres |
| 7 | `feat(db): migration runner and baseline migration` | Apply, rollback, re-apply on empty DB |
| 8 | `feat(worker): worker entry point with Redis connectivity check` | Starts, reports ready, exits on signal |
| 9 | `feat(api): pagination and idempotency-key helpers` | Unit tests |
| 10 | `docs(api): move OpenAPI to /v1 with shared components` | Spec lint passes |
| 11 | `test(api): contract test that routes match OpenAPI` | Fails when a route is added without a spec entry |
| 12 | `chore(compose): healthchecks and profiles for api, worker, ai-service` | `docker compose up` reports healthy |
| 13 | `docs(backend): local setup, migrations and testing guide` | A second person follows it from a clean checkout |
| 14 | `docs(adr): record approved decisions D1 to D13` | Review |

Not in B0: authentication replacement (B1 first commit), any domain table other than the migration bookkeeping table, any reversal code.

## Exit checks for B0

1. `go build ./...`, `go vet ./...`, `go test ./...` pass locally and in CI.
2. `docker compose up` starts Postgres, Redis, API and worker, and `/v1/health` and `/v1/ready` respond.
3. Migration up and down work on an empty database.
4. OpenAPI lints and the route-versus-spec test passes.
5. Risk register and decision list have no unanswered blocker.

## Proposed first implementation unit

Commit 1: add the missing `go.sum`, set one Go version in `go.mod`, and confirm `go build ./...` and `go vet ./...` succeed. It changes no behaviour, makes CI trustworthy, and every later commit depends on it. It needs D11 answered.

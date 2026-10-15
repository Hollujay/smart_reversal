# ADR 0002: Backend Stack and Modular Monolith

Status: Proposed. Awaiting approval of D1 and D2 in `docs/backend/decisions-requiring-approval.md`.

## Context

ADR 0001 is a placeholder. The repository already assumes a Go API (`apps/api`, Chi, `Dockerfile.api`, CI job, `implement.md`) and a Python FastAPI service for AI (`apps/ai-service`). The backend brief proposes Python and FastAPI for the whole backend and also asks that existing conventions be followed.

## Decision (proposed)

1. Keep Go with Chi for the core API, worker, migrator and simulator.
2. Keep Python and FastAPI for the AI service only. The backend talks to it through a typed adapter with timeouts and rule-based fallbacks.
3. Build a modular monolith: one codebase, separately runnable API and worker processes, one Postgres database.
4. Use PostgreSQL 16 with row-level security for tenant isolation, and `NUMERIC(18,4)` for money.
5. Use Redis 7 Streams for ingestion and jobs, with a transactional outbox so a committed change is never lost before publication.
6. Isolate reversal execution in one package that only the reversal service can import.
7. Version the API under `/v1`, defined OpenAPI first in `packages/api-contracts`.
8. Run all financial execution through a mock provider, off by default, with a kill switch.

## Alternatives considered

| Option | Reason not chosen |
|---|---|
| Python and FastAPI for everything | Discards existing Go scaffolding, Dockerfile and CI. Reasonable if the team decides speed of delivery outweighs that cost. |
| Microservices | No scaling or ownership reason at this stage. Adds failure modes in a system where partial failure causes financial risk. |
| Database-only queue | Simpler, but no consumer groups or easy replay. Outbox pattern keeps the useful part. |

## Consequences

- Dependency versions must be confirmed at B0 time (`go.sum` is missing today).
- Go version must be aligned across `go.mod`, CI and docs (D11).
- Existing in-memory packages are treated as prototypes. They are replaced, not extended, where they conflict with the safety rules (auth, audit).
- If D1 goes to Python, this ADR is superseded and the module layout is translated, not redesigned.

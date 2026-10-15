# Risk and Assumption Register (B0)

Status: draft. Review at the end of each phase.

## Risks

| ID | Risk | Likelihood | Impact | Mitigation | Owner |
|---|---|---|---|---|---|
| K1 | Stack disagreement between brief and repo causes rework | High | High | Decide D1 first | Backend lead |
| K2 | Source playbooks not available, requirements guessed | High | High | Obtain documents, reconcile matrix before B1 | Team |
| K3 | Existing dev login ships by accident | Medium | High | Remove in B1 first commit, test that unknown users get 401 | Backend |
| K4 | Tenant filter missed in one query | Medium | High | Row-level security plus isolation tests | Backend |
| K5 | Reversal executed twice under retry or concurrency | Medium | Critical | Idempotency key, unique remediation lock, concurrency test | Backend |
| K6 | Timeout treated as failure and money reversed wrongly | Medium | Critical | Five outcome classes, status-query-first, tests for S1 | Backend |
| K7 | AI output influences a financial decision | Low | Critical | `aiclient` not importable from `reversal`, CI check | Backend |
| K8 | Contract drift with frontend | High | Medium | OpenAPI first, contract tests | Backend and frontend |
| K9 | Simulated data mistaken for verified data in reports | Medium | High | `simulated` flag, separate report sections | Backend |
| K10 | Real customer data enters fixtures | Low | High | Synthetic generator only, review fixtures in PRs | All |
| K11 | CI red from day one (missing go.sum, version mismatch) | Certain | Low | Fix in first B0 commit | Backend |
| K12 | Hackathon time pressure tempts skipping safety tests | Medium | High | Safety tests are exit criteria, not optional | Backend lead |

## Assumptions

| ID | Assumption | Status |
|---|---|---|
| A1 | All money movement is simulated for the hackathon | Stated in brief |
| A2 | A single bank tenant is enough for the demo, with a second tenant used for isolation tests | Unverified |
| A3 | The frontend will consume REST JSON only | Unverified |
| A4 | The AI service can be absent and the system still passes all safety tests | Required by brief |
| A5 | Postgres 16 and Redis 7 are acceptable | Matches compose file |
| A6 | Business calendar covers Nigerian public holidays and weekends | Unverified, needs source |
| A7 | Waiting thresholds and approval limits come from the v2 scope document | Unverified, see D13 |
| A8 | No production integration with NIBSS or any bank is attempted in Track H | Stated in brief |

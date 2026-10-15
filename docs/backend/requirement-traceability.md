# Backend Requirement Traceability Matrix (B0)

Status: draft. Source: the backend master brief only. The four playbook documents are not in the repository, so the "Playbook ref" column is left for the team to fill in. Do not treat this matrix as reconciled until that is done.

Repo state key: `none` nothing exists, `stub` empty or in-memory placeholder, `partial` usable base needing rework.

| ID | Requirement | Phase | Repo state | Notes |
|---|---|---|---|---|
| R01 | Modular monolith, separate worker, isolated reversal boundary | B0 | none | Needs D1 and D2 |
| R02 | ADRs for major decisions | B0 | stub | ADR 0001 is one line |
| R03 | Docker Compose, DB connection, migrations | B0 | partial | Compose exists, migrations empty |
| R04 | OpenAPI under `/v1` | B0 | partial | 9 unprefixed paths |
| R05 | CI foundation | B0 | partial | Go job fails on missing go.sum |
| R06 | Risk and assumption register | B0 | none | See `risk-assumption-register.md` |
| R10 | Tenant-aware models, tenant isolation tests | B1 | none | No tenant concept anywhere |
| R11 | JWT verification, RBAC, ABAC, privileged actions | B1 | stub | Dev login must go |
| R12 | Service-to-service auth | B1 | none | |
| R13 | Masking and tokenisation | B1 | none | |
| R14 | Append-only hash-chained audit | B1 | stub | In memory |
| R15 | Versioned config, business calendar, deadlines | B1 | none | |
| R20 | Canonical event schema, validation, normalisation | B2 | partial | Channel set incomplete |
| R21 | Channel adapters (NIP, POS, ATM, card, mobile, internet, USSD) | B2 | none | |
| R22 | Redis Streams, dedup, retry, dead letter, replay | B2 | none | |
| R23 | Deterministic synthetic simulator | B2 | none | Fixtures exist in `tests/fixtures` |
| R30 | Detectors: duplicate debit, no beneficiary credit, failed with funds withheld, POS timeout, unknown status, reversal failure, settlement discrepancy | B3 | stub | `exceptions.go` classifier to review |
| R31 | Codes DUP001, FAIL001, UNK001, REV001, STL001, CUS001 plus codes in v2 scope | B3 | none | Extra codes unknown until source docs are read |
| R32 | Configurable thresholds, risk grade, evidence refs, rule versions | B3 | none | |
| R40 | Exception queue, cases, assignment, transitions, priority, root cause | B4 | stub | `investigations` empty |
| R41 | SLA deadlines, escalation, complaint link, audit | B4 | none | |
| R50 | Status-query-first workflow | B5 | none | |
| R51 | Eight-check eligibility gate | B5 | none | See safety rules below |
| R52 | Maker-checker with amount and policy limits | B5 | stub | Protocol doc exists, no code |
| R53 | Remediation lock, idempotency keys, duplicate prevention | B5 | none | |
| R54 | Mock reversal provider, post-reversal verification, retry | B5 | none | |
| R55 | Kill switch, execution disabled by default | B5 | none | |
| R60 | Settlement ingestion, NIBSS adapter contract, reconciliation, reports | B6 | stub | Sample CSV fixture exists |
| R70 | Incidents, alerts, notification orchestration, providers, templates | B7 | none | |
| R71 | Two-stage customer notification, no refund claim before verification | B7 | none | |
| R80 | Read-only reporting APIs, synthetic vs verified metrics separation | B8 | none | |
| R81 | Metrics, structured logs, trace IDs | B8 | partial | `observability/logger.go` |
| R90 | Hardening, rate limits, upload safety, scanning, restore docs | B9 | none | |
| R100 | ReverseIt customer APIs, ownership enforcement, parser interface with fallback | B10 | none | |
| R110 | Scheduled jobs, scenario seeding, reset, simulated time, replay | B11 | none | |
| R120 | AI interfaces: anomaly, classification, root cause, parser, complaint wording, summary, feedback | cross | partial | AI service has two routers |
| R121 | AI output validated, advisory, never authorises money movement, safe fallback | cross | none | |

## Financial safety rules carried into every phase

These are non-negotiable and each needs a named test before the phase that owns it can close.

- S1 Five outcome classes. A timeout never means failed.
- S2 Eight reversal checks. Any failure blocks automatic reversal.
- S3 Three separate states: request accepted, processing, verified outcome.
- S4 Maker and checker must differ. Admin does not bypass this.
- S5 Evidence is immutable.
- S6 Retries and duplicate submits cause no second financial effect.
- S7 Execution is off by default and has a kill switch.
- S8 AI output cannot bypass rules, policy or approvals.
- S9 All execution in this project is simulated and labelled as such.

## Open items

- Reconcile every row with the playbook documents and fill in section references.
- Obtain the full list of v2 classification codes beyond the six named.

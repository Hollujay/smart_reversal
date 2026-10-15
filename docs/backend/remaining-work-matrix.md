# Remaining backend work: coverage matrix B2 to B11

Assessment only. No code was written for this document. Baseline: tag `demo-baseline-2026-10-15` at `bcdd89c` (upstream `develop` after PRs #13 and #14; 357 top-level tests passing; demo seed, play and reset verified). Source of requirements: prompts B2 to B11 of the v2 Integrated Scope Prompt Pack, including their "SCOPE v2 ADDITIONS", checked against the Full Scope, Vol 2 and the Complete Playbook (which add no backend requirement of their own, see `requirement-traceability.md`).

Classes: COMPLETE, PARTIAL, NOT IMPLEMENTED, BLOCKED (needs an external decision or party), NOT APPLICABLE. COMPLETE means implemented and covered by a test I could name. Things that work in the demo but only for the demo path are PARTIAL.

Note: `requirement-traceability.md` still shows most rows as "none" or "stub". It is stale. Refreshing it is the first documentation commit of the plan.

## B2 Ingestion, normalisation, correlation, simulator

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 | REST `/v1/events` adapter | COMPLETE | `server/ingest.go`, `txn/ingest.go`; `server/ingest_test.go`, `txn` tests |
| 1 | File-drop adapter (CSV/XML/fixed-width), message-consumer adapter, versioned mapping configs | NOT IMPLEMENTED | Only the canonical JSON envelope exists. Interface specs exist in `packages/api-contracts/interfaces/` |
| 2 | Idempotency 200 / 409 / 422 to dead letter | PARTIAL | 200 and 409 and quarantine done (`TestRepeatedEventsAreIdempotent`, `TestSameReferenceWithDifferentContentIsAConflict`, `TestInvalidEventsAreQuarantinedNotStored`). No data-quality alert on 409. No replay tool for `quarantined_events` (stream DLQ replay exists in `stream/replay.go`) |
| 3 | Order independence, late events, deterministic projection | COMPLETE | `txn/project.go`; `TestProjectionDoesNotDependOnArrivalOrder`, `TestLateAndOutOfOrderEventsConverge`. Per-source clock-skew computation and exposure: NOT IMPLEMENTED (received time is stored, skew is not computed) |
| 4 | Correlation by external refs (RRN, STAN, session, ledger id, gateway ref) with confidence rules; unmatched-evidence queue and audited manual link | NOT IMPLEMENTED | Correlation today is `instruction_ref` only, plus `session_id` stored. Quarantine is not a manual-link queue |
| 5 | Lifecycle projection incl. UNKNOWN, CONFLICTING, REVERSED | COMPLETE | `TestProjectionStates`, `TestATimeoutNeverBecomesAFailure`. Missing: separate SENT state and finer REVERSAL_* states (PARTIAL detail) |
| 6 | Source freshness (fresh/stale/down) exposed via API and attached to evidence | NOT IMPLEMENTED | |
| 7 | Backpressure, lag metrics, spill to disk | NOT IMPLEMENTED | Redis Streams give buffering; no bounds, no lag metric |
| 8 | Simulator: CLI/API, 18 named scenarios with seeds and ground-truth label file, burst, clock skew, outages | PARTIAL | Only the 3-case `internal/demo` scenario. Mock services `cmd/mock` are 501 stubs (`internal/mocks`). Simulated status, ledger, provider exist as DB tables in `reversal/sources.go` |
| 9 | Property tests; shuffled replay yields identical state | PARTIAL | Order test exists for the projection; no shuffled whole-scenario replay, no property-based test |
| v2 | Source types NIP, POS, ATM, card, mobile, internet, USSD, gateway, merchant; field list (debit/credit/switch/settlement status, retry count) | PARTIAL | Channel allow-list and rail pattern exist (`txn/event.go`). Switch, settlement and retry-count fields are not in the envelope |
| v2 | 50-transaction demo with one unknown, deterministic seed, time acceleration, detection latency p50/p95 | NOT IMPLEMENTED | |

## B3 Rules engine and detectors

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 | Declarative versioned rules, shadow mode, maker-checker, rollback | NOT IMPLEMENTED | Detectors are Go code with a `RuleVersion` constant (`detect/detect.go`). Thresholds are configurable (config service) |
| 2 | Detectors | PARTIAL | Done: DUP (EX-DUP) , DNC/FAIL001 (confirmed failure), UNK001, stuck sweep, late success closure; tests in `detect` (11). Missing: EX-INC, RVF/RVP (REV001 is opened by reversal failure only), FEE, VND, ATM, LTE (late success after reversal), MIS, RTY, DLY, STL (B6). DUP uses reference/session/settlement and a window, no allow-list for legitimate repeat payments |
| 3 | Exception record contents | PARTIAL | Case carries code, severity, exposure, rule id/version, root cause, evidence, explanation. Missing: confidence/outcome field, recommended next step, lane |
| 4 | Degraded-source marker and suppress/relabel | NOT IMPLEMENTED | Depends on B2 #6 |
| 5 | Re-query ladder with self-healing closure | PARTIAL | Late success closes a case (`TestATimeoutIsUnknownNotFailureAndALateSuccessClosesTheCase`); the sweep finds stuck transactions. No scheduler that queries the status source at configured intervals |
| 6 | Exposure: gross flagged, submitted, reversed, net | PARTIAL | `exposure_minor` per case; the four separate numbers are not computed |
| 7 | Configurable explainable priority score | PARTIAL | Fixed `priorityFor(code, risk)` in `cases/cases.go`. No weights for amount, age, SLA, repeat complaints |
| 8 | Golden-case suite with precision and recall, adversarial cases | NOT IMPLEMENTED | Unit tests cover some non-duplicate cases (`TestThingsThatAreNotDuplicates`) |
| v2 | Risk grade High/Medium/Low | COMPLETE | `TestDuplicateDebitsAreFlaggedOnTheLaterOneAndGradedBySession`, `TestDifferentSessionsAreMediumRisk` |
| v2 | Never treat unknown as failed; status query first | COMPLETE | `TestATimeoutNeverBecomesAFailure`, `TestUnknownPendingConflictingAndSuccessfulOutcomesAreNeverReversed` |
| v2 | Duplicate posting vs duplicate transfer; ML hook (A5) | NOT IMPLEMENTED / BLOCKED | ML hook needs the AI owner's agreement |
| v2 | Codes DUP001, FAIL001, UNK001, REV001, STL001, CUS001 and proposed extensions | PARTIAL | Pattern allows any code; only the first three are produced by detectors, REV001 by reversal, CUS001 by complaints |

## B4 Cases, state machine, SLA, complaints

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 | State machine, server-side validation, reason codes, side states | COMPLETE for the generic machine | `cases/cases.go`; `TestStateMachineAllowsOnlyDefinedMoves`, `TestTransitionsAreValidatedVersionedAndRecorded`. Per-exception-type workflow definitions: NOT IMPLEMENTED (one machine for all types) |
| 2 | Ownership: rule-based assignment, manual assign, team queues, reassign, bulk with audit | NOT IMPLEMENTED | `owner_id` column only. `routing` package (delegations, on-call, escalation steps, from B1) is not wired to cases. `team` scope is blocked by D21 |
| 3 | SLA: deadline from approved clocks, customer resolution clock, pause for external dependency, breach events, escalation ladder | PARTIAL | Deadline from approved clocks is done (`TestDeadlinesComeOnlyFromApprovedClocks`). No pause rules, breach or approaching-breach events, escalation execution, customer clock separate from case SLA |
| 4 | Notes, attachments (virus-scanned, hashed), evidence links, linked cases, resolution codes | PARTIAL | Notes and evidence done (`server/cases_test.go`). No attachments or scanning, no linked cases, no additional-check requests |
| 5 | Complaint intake with matcher and outcomes | PARTIAL | Exact-reference join or new CUS001 or unmatched claim (`TestComplaintsJoinTheDetectedCaseOrOpenACustomerCase`). No fuzzy match with confidence, no "verified-successful fast closure pack", no repeat-contact counter, no no-match guidance |
| 5 | Customer-safe status vocabulary | COMPLETE | `TestCustomerStatusNeverPromisesMoneyBeforeVerification` |
| 6 | Search and saved views | PARTIAL | Queue filters and cursor paging done. Search by reference/amount/date and saved views not done |
| 7 | Domain events via outbox for every transition | NOT IMPLEMENTED | Cases, reversals and notifications do not write to the outbox. The outbox is used for transaction updates only. This blocks analytics and B7 alerts that react to transitions |
| v2 | Case fields incl. internal target separate from regulatory deadline | PARTIAL | `sla_due_at` only; `internal_target_at` missing |

## B5 Controlled reversal

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 | Eight-check gate, unknown means hold, stale evidence cannot pass, all failed checks returned | COMPLETE | `reversal/gate.go`; `TestEachOfTheEightChecksCanBlock`, `TestUnknownPending...`, `TestEvidenceThatChangesBetweenApprovalAndExecution...`; `server/reversals_test.go` |
| 2 | Remediation lock | COMPLETE for in-platform | Unique index plus `TestOnlyOneLiveReversalCanExistPerTransactionEvenUnderConcurrency`. Detecting remediation done outside the platform: NOT IMPLEMENTED (BLOCKED on data availability) |
| 3 | Policy decision AUTO / NEEDS_APPROVAL / MANUAL_ONLY / BLOCKED with reasons and version; time-of-day, rollout %, confidence only additional | PARTIAL | Verdicts PASS_AUTO, PASS_NEEDS_APPROVAL, BLOCKED with policy version. No MANUAL_ONLY, time-of-day or rollout rules |
| 4 | Approvals: maker-checker, limits, re-validation at execution, rejection returns case to investigation | PARTIAL | Done and tested (`TestApprovalRulesAreEnforced`, DB trigger). Approval expiry and higher-limit escalation chain NOT IMPLEMENTED |
| 4 | Unknown-status REVIEW items with dual-approver override and fresh evidence | NOT IMPLEMENTED | We block unknown outright (safer than the scope). Override needs a product decision |
| 5 | Execution: key persisted before call, timeout is UNKNOWN, retry with same key | COMPLETE against simulated provider | `TestATimeoutIsUnknownAndTheSameKeyFindsOutWhatHappenedWithoutPostingTwice`. Bounded retries with jitter and circuit breaker: NOT IMPLEMENTED |
| 6 | Verification from ledger | PARTIAL | `TestAProviderThatClaimsSuccessWithoutAPostingIsCaught`, delayed posting test. Missing: fee reconciliation, partial reversal, restricted or closed account, recovery sub-case, post-reversal watch window for late success |
| 7 | Manual reversal recording and linking | NOT IMPLEMENTED | |
| 8 | Notification only after verification, deduplicated | COMPLETE | `TestConfirmedFailureIsReversedVerifiedAndOnlyThenReportedToTheCustomer`; failure and retry of delivery belongs to B7 |
| 9 | Kill switch: freeze submissions, park in-flight, reconcile after unfreeze | PARTIAL | `TestTheKillSwitchStopsAutomationAtEveryScope`, `TestAFreezeAfterApproval...` . "Park and reconcile after unfreeze" not explicit. The kill switch is checked only by the reversal service; ingestion, detector, scheduler and notifications do not check it ("every consumer checks it") |
| 10 | Chaos tests | PARTIAL | Timeout, duplicate request, concurrency, freeze mid-flight, delayed posting covered. Missing: concurrent maker vs automation, late success after reversal |
| v2 | REV001 on failure; scripted mock status and reversal services | PARTIAL | `TestARejectedReversalFails...`. Scripting is in DB tables, not the HTTP mock services of B11 |

## B6 Reconciliation

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 to 3 | Importers, matching strategies, stored matching decisions | NOT IMPLEMENTED | |
| 3 | EX-STL / STL001 cases with quantified breaks | NOT IMPLEMENTED | Code is allowed by the case pattern only |
| 4 | Reconcile remediation outcomes against ledger | PARTIAL | `GET /v1/reconciliation/reversals` (`server/reports.go`, `TestReconciliationAndSummary...`) compares reversals to the simulated ledger. Not against settlement records |
| 5 | Reconciliation workspace API, manual match with approval | NOT IMPLEMENTED | |
| 6 | Extension points M22, M24 | NOT IMPLEMENTED | |
| v2 | NIBSS and settlement files with seeded breaks | NOT IMPLEMENTED | Depends on B2 simulator |

## B7 Incidents, alerts, notifications

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1, 2 | Incident grouping, spike detection | NOT IMPLEMENTED | |
| 3 | Alert service with routing, dedupe, suppression, acknowledge, escalate | NOT IMPLEMENTED | Ops message templates exist in `notify` but nothing routes them. Needs B4 #7 transition events first |
| 4 | Customer notification service: templates, dedupe key, delivery tracking, retry, fallback, contact-change guard | PARTIAL | Templates, dedupe, two-stage rule and "no money claim before VERIFIED" are done (`TestNotificationsAreTwoStageDedupedAndNeverPromiseBeforeVerification`). Delivery tracking, retries, provider adapters, opt-out, wrong-number guard NOT IMPLEMENTED. `notify` has no tests of its own |
| 5 | Test sink adapter | NOT IMPLEMENTED | Rows are recorded with status `recorded` and `simulated: true` |

## B8 Reporting, integration health, observability

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 | Metrics catalogue (definition, formula, sources) and the listed metrics | PARTIAL | `GET /v1/reports/summary` gives counts, verified value, automation rate, average resolution time. No catalogue endpoint, no p50/p95, detection latency, SLA compliance, exposure split, false-positive rate |
| 2 | Controlled exports | NOT IMPLEMENTED | |
| 3 | Integration health per adapter | NOT IMPLEMENTED | |
| 4 | OpenTelemetry traces, correlation id, dashboards as code | PARTIAL | Trace ids and structured logs exist (`internal/observability`). No OTel export, no metrics endpoint |
| 5 | Data lineage | PARTIAL | Cases keep evidence, rule id and version, append-only event history; no lineage query |
| v2 | Dashboards: ops, executive, customer timeline | PARTIAL | Customer timeline done; others via summary only |

## B9 Security, resilience, release readiness

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 | STRIDE threat model | NOT IMPLEMENTED | `docs/security/threat-model.md` is a 64-byte stub |
| 1 | Auth, RBAC matrix, tenant RLS, hash-chained audit, SoD, kill switch, token vault, service assertions | COMPLETE (B1) | Tests across `identity`, `authz`, `audit`, `datashield`, `servauth`, `killswitch`; `TestEveryNonPublicRouteRequiresAuthentication` |
| 1 | Secrets manager, key rotation plan, rate limiting beyond unmask, secure headers, container scanning | PARTIAL | A CI secret scan exists, and govulncheck was run by hand (not confirmed as a CI step). Rate limit for unmask only. No secure-header or login rate-limit review |
| 1 | Production MFA | BLOCKED | Dev TOTP only (D20); needs the bank's identity provider |
| 2 | Privacy: data inventory, retention, legal hold | NOT IMPLEMENTED | Counsel review needed (NDPA) |
| 3 | Load tests, DR restore test, reconciliation-after-outage, runbooks, game day | NOT IMPLEMENTED | `docs/runbooks/disaster-recovery.md` exists as a plan only |
| 4, 5 | Release checklist, pen-test pack | NOT IMPLEMENTED | Live execution sign-off is outside engineering |

## B10 ReverseIt backend

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 | Customer endpoints, authenticated customer role, own-data only | NOT IMPLEMENTED | No customer audience login. `POST /v1/complaints` is agent-entered. Needs a design decision on customer authentication (BLOCKED on decision) |
| 2, 3 | Intake of pasted text or image, consent, redaction; parser interface (A9) | NOT IMPLEMENTED / BLOCKED | A9 owned by the AI developer |
| 4 | Ranked fuzzy matching, unmatched claim, already reversed / successful / duplicate report | PARTIAL | Exact reference match and unmatched claim only |
| 5 | Dispute classification with deterministic clock mapping; CUS002 path | NOT IMPLEMENTED | Clock registry exists (`configsvc`) |
| 6 | Deadline engine: regulatory and internal target, states, source status | PARTIAL | Pieces exist (`bizcal`, clocks with status). No on-track/approaching/urgent/overdue state or API; clock values are unverified (D17 to D19) |
| 7 | Complaint generator | NOT IMPLEMENTED | |
| 8 | Case submission with ack | PARTIAL | Case creation yes; priority/ops alert/ack partial |
| 9 | Unified status tracker | COMPLETE for operators | `GET /v1/cases/{id}/customer-status` |
| 10 | Escalation ladder from configuration | NOT IMPLEMENTED | Contact details need Compliance review |
| 11 | Feedback loop, labeled examples, reopen on dispute | NOT IMPLEMENTED | |
| 12 | Fraud and abuse flags | NOT IMPLEMENTED | Unmatched claims already cannot move money (verified by the gate) |

## B11 Automation, mocks, config, demo controls

| # | Requirement | Class | Evidence / gap |
|---|---|---|---|
| 1 | Schedulers, idempotent, observable, pausable by kill switch | PARTIAL | Worker runs detector consumer, stuck sweep and reversal sweep (`worker/worker.go`). Sweeps are not kill-switch aware at job level, no job registry or metrics, no SLA, stale-source, notification retry or file jobs. Uses in-process tickers, not Celery/APScheduler (ADR 0002 chose Go) |
| 2 | HTTP mock services with scripted outcomes, OpenAPI, contract tests | PARTIAL | OpenAPI for six interfaces and spec tests exist (`internal/contract`); mocks answer 501. Simulated adapters are in-database |
| 3 | Config service, maker-checker, hot reload, effective version on decisions | COMPLETE | `configsvc` (17 tests); gate records `policy_version` |
| 4 | Demo controls: seed, scenarios, reset, time acceleration, burst, run log export | PARTIAL | seed, play, reset and simulated flag done (`internal/demo`, `TestTheDemonstrationScenarioRunsEndToEndAndCanBeRepeated`). No 50-transaction scenario, acceleration, burst, run-log export |
| 5 | Metrics endpoint | NOT IMPLEMENTED | |
| 6 | Buffering and degraded-source state when a mock is down | NOT IMPLEMENTED | |

## Overlaps and cleanup (so nothing is built twice)

- `internal/transactions`, `internal/exceptions` are unwired prototypes superseded by `txn`, `cases`, `detect`. `internal/integrations`, `investigations`, `reconciliation`, `reversals` are empty stubs. Delete or repurpose in one cleanup commit before B6 reuses the `reconciliation` name.
- `routing` (delegations, on-call, escalation steps) is built and unused. B4 ownership, B4 SLA escalation, B7 alert routing and B10 escalation all need it. Wire it once.
- The simulated status, ledger and provider live in database tables. B11 HTTP mocks, B2 simulator and B6 file generators should drive those same tables rather than add a second simulation.
- The reversal reconciliation endpoint should become one view of the B6 matching store, not a separate implementation.
- Notification persistence (`notifications` table) is the base for B7 delivery tracking; extend it.

## Dependencies, unfinished security controls, decisions

- Domain events from case and reversal transitions (B4 #7) gate B7 alerts, B8 metrics and B11 jobs.
- Kill switch must be checked by the detector, ingestion decision to accept, scheduler jobs and notification dispatch (B5 #9, B11 #1).
- Customer authentication design (B10 #1) is a decision, not code. Proposed options: separate customer audience with channel-issued tokens, or agent-proxied only for the hackathon scope.
- D17 to D19 and approval thresholds, regulatory clocks, calendars remain unverified; every new feature reads them from configuration and blocks when absent.
- MFA stays development TOTP until the bank names an identity provider.
- AI-owned interfaces (A5 score, A9 parser): not modified; adapters only, with the owner's agreement.
- `team` audit scope stays inaccessible until D21 is revisited with ownership in B4.

## Proposed order and ordering conflicts

Prompt order is B2, B3, B4, B5, B6, B7, B8, B9, B10, B11. Two conflicts to flag:

1. **B3 and B4 need B11's config and scheduling pieces and B2's simulator.** Config is already done. A seeded scenario simulator (B2 #8) is needed first so B3 precision and recall can be measured. I propose building the simulator with B2 and not waiting for B11.
2. **B4 domain events (outbox) are a prerequisite for B7 and B8 but sit in B4.** I propose doing them in B4 as the first item, which keeps the original order.
3. B5 gaps (watch window, partial reversal, circuit breaker) depend on provider behaviours in the simulator; scheduled after B2/B3.

## Safest next feature

**Domain events for case, reversal and notification transitions through the existing outbox (B4 #7), together with a refreshed traceability matrix.**

Why: it is additive, adds no money movement, does not change any API response, unblocks B7, B8 and B11, and is verifiable by consumers in tests. The demo is untouched.

What already exists: the transactional outbox and relay (`outbox`), Redis Streams consumers with idempotency and DLQ (`stream`), tenant-scoped events, and a clean place to write them (every transition already goes through `cases.Service.Transition`/`RecordEvent` and `reversal` `moveCase`, in the same database transaction as the audit entry).

Proposed commits, each with its tests:

1. `docs: refresh requirement traceability to match the implemented backend` (this matrix becomes the source; no code).
2. `feat(events): event type registry and enqueue helper for domain events` (names, versioned payload with ids only, no personal data; test that payloads carry no account numbers).
3. `feat(cases): emit case.opened and case.transitioned through the outbox in the same transaction` (tests: atomic with the transition, rolled back with it, not duplicated on idempotent reopen).
4. `feat(reversal): emit reversal.requested, approved, submitted, verified and failed events` (tests: exactly one event per step, none for blocked gate, one verified event even when Advance repeats).
5. `feat(notify): emit notification.recorded` (test that a deduped notification emits nothing).
6. `test(events): domain events survive Redis outage and are consumed once per group` (integration over the real relay and stream).
7. Phase gate: full suite, demo seed/play/reset, check that the five demonstration properties still hold.

After that, the order I recommend: cleanup of dead packages; B2 simulator and correlation; B3 remaining detectors with the golden-case suite; B4 ownership, SLA pause and breach, search; B5 gaps; then B6 to B11 as in the matrix.

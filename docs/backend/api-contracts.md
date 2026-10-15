# Shared API Contracts (B0 inventory)

Status: draft. This lists the contracts that the frontend and AI developers will depend on. Nothing here is final until D3 and D4 in `decisions-requiring-approval.md` are answered. Changes to a published contract must be written down and sent to the affected developer.

## 1. Conventions to fix in B0

| Topic | Proposal |
|---|---|
| Base path | `/v1` |
| Auth | `Authorization: Bearer <JWT>` for operators and customers. Separate service credential for service-to-service calls. |
| Tenant | Taken from the verified token, never from a request body or query parameter. |
| Errors | One envelope: `{ "error": { "code": "...", "message": "...", "traceId": "...", "details": [] } }` |
| Pagination | `limit` (default 50, max 200) and opaque `cursor`. Response has `items` and `nextCursor`. |
| Filtering | Explicit query parameters per resource, documented in OpenAPI. |
| Idempotency | `Idempotency-Key` header required on all state-changing financial endpoints. |
| Money | Decimal string plus ISO 4217 currency. Never a float. |
| Time | RFC 3339 UTC. |
| Simulated data | Every simulated response carries `"simulated": true`. |
| Trace | `X-Trace-Id` accepted and echoed. |

## 2. Frontend-facing resources

| Resource | Phase | Existing in OpenAPI |
|---|---|---|
| Session and identity (`/v1/auth/*`, `/v1/me`) | B1 | `/auth/login` only |
| Transactions (list, detail, evidence) | B2 | `/transactions` list |
| Exceptions (queue, detail, classification) | B3 | `/exceptions` list |
| Cases (create, assign, transition, notes, evidence, timeline) | B4 | `/investigations` create |
| Reversals (propose, eligibility result, approve, status) | B5 | `/reversals/propose`, `/reversals/{id}/authorize` |
| Reconciliation (runs, mismatches, reports) | B6 | `/reconciliation/runs` |
| Incidents and notifications | B7 | none |
| Reporting and dashboards | B8 | none |
| Audit log | B1 | `/audit-logs` |
| ReverseIt customer endpoints | B10 | none |
| Scenario and demo controls | B11 | none |
| Health (`/v1/health`, `/v1/ready`) | B0 | `/health` |

## 3. AI integration contracts

Direction is backend to AI service. The backend validates every response, records provenance, applies a timeout, and falls back to rules when the service is down or returns invalid output.

| Capability | Request | Response must include |
|---|---|---|
| Anomaly score | Canonical transaction plus context features | `score` 0 to 1, `modelVersion`, `generatedAt` |
| Classification suggestion | Exception candidate and evidence references | `suggestedCode`, `confidence`, `modelVersion` |
| Root-cause recommendation | Case summary and evidence references | Ranked causes with reasons, `modelVersion` |
| Debit-alert parsing | Customer-supplied text or image reference | Extracted fields each with `confidence`, `parserVersion` |
| Complaint wording | Case facts and deadline data | Draft text, `modelVersion` |
| Case summary | Case timeline | Summary text, `modelVersion` |
| Feedback label | Case outcome and reviewer label | Accepted acknowledgement |

Rules for all of them: output is advisory, low confidence is shown as such, and no AI field is an input to the eight-check reversal gate.

## 4. Known drift to resolve

- OpenAPI paths lack `/v1`.
- No shared error, pagination or auth scheme components.
- `apps/web/src/types/domain.ts` needs to be diffed against the OpenAPI schemas before the first contract freeze.
- The AI service has no request or response models yet, so the contract above is a proposal for the AI developer to accept or amend.

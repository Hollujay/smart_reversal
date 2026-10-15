# Decisions Requiring Human Approval

Status: open. B0 implementation does not start until D1 and D2 are answered. Each entry has a recommendation, but the decision belongs to the team.

| ID | Decision | Recommendation | Why it matters |
|---|---|---|---|
| D1 | Backend language: keep Go (current repo direction) or switch to Python and FastAPI (brief) | Keep Go | ADR placeholder, `implement.md`, Dockerfile, CI and package skeleton all assume Go. Switching discards that work. Python stays for the AI service. If the team prefers Python for hackathon speed, say so now, because it changes every B0 file. |
| D2 | Broker: Redis Streams (brief) vs a database-backed outbox only | Redis Streams plus a transactional outbox in Postgres | Streams give replay and consumer groups. The outbox prevents losing events between a commit and a publish. |
| D3 | Public API shape: OpenAPI first, with server stubs checked against it | OpenAPI first, contract tests in CI | Prevents drift with the frontend. |
| D4 | Who signs off contract freezes | One named owner each for frontend and AI | The brief forbids silent contract changes. |
| D5 | Token signing: keep HS256 shared secret or move to asymmetric keys (RS256 or EdDSA) | Asymmetric keys with a JWKS style key set | Services other than the API can verify without holding a signing secret. |
| D6 | Identity source for the hackathon: local user table with hashed passwords, or an external IdP | Local table, behind an interface | Removes the current accept-anything login without needing an IdP. |
| D7 | Multi-tenancy model: shared schema with row-level security, or schema per tenant | Shared schema with RLS | Simpler migrations. Isolation is tested rather than assumed. |
| D8 | Customer identity for ReverseIt: separate token audience and user table | Separate audience | Keeps customer and operator authorisation from mixing. |
| D9 | Admin role and maker-checker | Admin cannot approve their own request and cannot bypass the gate | Current `RequireRoles` lets ADMIN through every check. |
| D10 | Data masking approach: format-preserving mask in responses, tokenised storage of account numbers | Both | Needed before any real-looking fixtures are loaded. |
| D11 | Go version: 1.22 in CI, 1.23 in `go.mod`, 1.24 in CONTRIBUTING, 1.27 local | Pick one and align all four | Build reproducibility. |
| D12 | Financial execution flag default and who may enable it | Off everywhere, enabling requires two approvers and an audit record | Brief section 5. |
| D13 | Where the v2 extra classification codes and approval limits come from | Team supplies the playbook sections | They cannot be invented. |

## Needed from the team before B0 coding

1. Answers to D1 and D2.
2. The four playbook documents, or the relevant sections, added to a shared location.
3. Confirmation of the frontend and AI owners named in D4.

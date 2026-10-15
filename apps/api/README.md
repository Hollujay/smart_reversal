# SMARTREVERSAL Core API Service

Go-based core financial operations and ledger integration service.

## Structure
- `cmd/api`: Service entrypoint
- `internal/config`: Environment configurations
- `internal/observability`: Structured logging and metrics
- `internal/auth`: Maker-checker authorization and JWT validation
- `internal/transactions`: Ingestion and query engine
- `internal/reconciliation`: Multi-source reconciliation engine
- `internal/exceptions`: Financial discrepancy and exception tracking
- `internal/investigations`: Investigation workflows
- `internal/reversals`: Controlled dual-authorization reversal dispatch
- `internal/audit`: Cryptographic hash-chained audit trail
- `internal/integrations`: Gateway and external adapter clients
- `migrations`: Database migration scripts

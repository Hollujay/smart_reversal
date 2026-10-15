# Backend Local Setup

Every command here was run on Linux with Go 1.27.2 and Docker 29. Commands under "Docker Compose" were not run, see the note there.

## Requirements

- Go 1.27.2 (the version in `apps/api/go.mod`)
- Docker, to run PostgreSQL 16 and Redis 7
- `make`, `curl`

## 1. Start PostgreSQL and Redis

Using Docker directly:

```bash
docker run -d --name sr-pg -p 5432:5432 \
  -e POSTGRES_USER=smartreversal_app \
  -e POSTGRES_PASSWORD=development_secret_only \
  -e POSTGRES_DB=smartreversal_db postgres:16-alpine
docker run -d --name sr-redis -p 6379:6379 redis:7-alpine
```

If a port is already taken on your machine, publish a different host port (for example `-p 55433:5432`) and set `DB_PORT` or `REDIS_PORT` in step 2.

### Docker Compose

`compose.yaml` defines postgres, redis, a one-shot migrate step, api, worker and the three mock services. The `docker compose` plugin was not available where this was written, so the file was checked as YAML and its topology was reproduced with plain `docker run` on a private network: migrations applied, api and worker became ready, mocks reported health, and the processes ran as a non-root user. `docker compose up` itself has not been run.

```bash
docker compose up --build
```

## 2. Configure

The Go processes read the process environment, not `.env` files. Load `.env` into your shell:

```bash
cp .env.example .env        # edit ports if needed
set -a; . ./.env; set +a
```

Required: `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET` (at least 32 bytes). The processes refuse to start without them and list every problem at once. The placeholder values in `.env.example` are accepted only when `ENVIRONMENT=development` or `test`. `FINANCIAL_EXECUTION_ENABLED` defaults to `false` and is refused in production.

## 3. Migrate

```bash
make db-migrate-status
make db-migrate
make db-migrate-down        # reverts the latest migration
```

Migration files live in `apps/api/migrations` as `NNNN_name.up.sql` and `NNNN_name.down.sql`. Each runs in one transaction. Editing a file after it has been applied is an error, so add a new migration instead.

## 4. Run

```bash
make dev-api       # http://localhost:8080
make dev-worker    # http://localhost:8081
curl localhost:8080/v1/health
curl localhost:8080/v1/ready
```

`/v1/ready` reports `postgres` for the API and `postgres` plus `redis` for the worker. A failing dependency returns 503 and only says `failed`, never the underlying error.

## 5. Test

Unit tests:

```bash
make test-api
```

Integration tests need a disposable PostgreSQL database and a Redis. They **skip** when these variables are unset, so a green run without them proves less. Use a database that is not your development one:

```bash
docker exec sr-pg createdb -U smartreversal_app smartreversal_test
export TEST_DATABASE_URL='postgres://smartreversal_app:development_secret_only@localhost:5432/smartreversal_test?sslmode=disable'
export TEST_REDIS_ADDR=localhost:6379
(set -a; . ./.env; set +a; DB_NAME=smartreversal_test make db-migrate)
make test-api-integration
```

`make test-api-integration` runs packages one at a time (`-p 1`) because they share the database. CI fails if any test is skipped.

Other checks that CI runs:

```bash
make lint-api                                            # gofmt and go vet
cd apps/api && go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
docker run --rm -v "$PWD:/repo" zricethezav/gitleaks:v8.30.1 \
  git /repo --config /repo/.gitleaks.toml --redact --no-banner
```

## What is not here yet

Authentication (B1), tenancy (B1), ingestion and the simulator (B2), detection (B3), cases (B4), reversal (B5). The mock services answer health checks only and return `501` on their interface routes. The login route is not mounted: the earlier handler that issued tokens without credentials was disabled and is not served.

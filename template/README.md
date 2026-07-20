# Usage Analytics

Usage analytics platform: high-volume telemetry ingest, per-tenant
dashboards, and scheduled exports.

## Repo layout

| Path | What it is |
| --- | --- |
| `frontend/` | React/TypeScript dashboard (Vite, MUI, Recharts) |
| `backend/` | Go API service (chi, pgx, Postgres) |
| `bazel/` | Offline export-integrity tooling (Bazel) |
| `data/` | Sample telemetry exports used by load tests and docs |

## Local development

Bring up Postgres and the API:

```sh
docker compose up postgres backend
```

Run the dashboard against it:

```sh
cd frontend
npm install
npm run dev
```

The dashboard falls back to bundled sample data when the API is
unreachable, so `npm run dev` alone also works.

## Tests

```sh
# frontend
cd frontend
npm run lint && npm run typecheck && npm run test:unit
npm run test:e2e            # needs `npx playwright install`

# backend
cd backend
go test ./...
DATABASE_URL=postgres://metrics:metrics@localhost:5432/metrics go test -tags integration ./internal/store/
```

## Notes

- `backend/internal/events/` is generated — edit `backend/tools/genevents`
  and re-run `go run ./tools/genevents` instead.
- Session fingerprinting intentionally uses a slow KDF (SEC-114); the
  `METRICS_WORKLOAD` env var scales the verification sweep size in CI.

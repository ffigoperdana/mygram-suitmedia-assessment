# Baseline Audit Report

Date: 2026-07-20

## Application structure

- `main.go`, `config/`, `database/`, `controllers/`, `middlewares/`, `models/`, `router/`, and `services/`: Go/Gin API backed by PostgreSQL.
- `services/object_storage.go`: authenticated upload proxy for S3-compatible Garage storage; this remains configured through `S3_*` environment variables.
- `mygram-frontend/`: React 18 + Vite + TypeScript application, with Vitest unit tests and Playwright smoke tests.
- `Dockerfile`: multi-stage, non-root Go API image.
- `mygram-frontend/Dockerfile`: Node build stage and Nginx runtime image.
- `docker-compose.fullstack.yml`: local PostgreSQL, API, and frontend stack.
- `docker-compose.prod.yml`: production image-based PostgreSQL, API, and frontend stack.

## Commands run

```powershell
$env:GOCACHE=(Join-Path (Get-Location) '.gocache')
$env:GOPATH=(Join-Path (Get-Location) '.gopath')
$env:GOMODCACHE=(Join-Path (Get-Location) '.gomodcache')
go test -count=1 ./...

Set-Location mygram-frontend
npm ci --ignore-scripts --no-audit --no-fund
npm run quality
npm run test:e2e

Set-Location ..
docker compose -f docker-compose.fullstack.yml config
docker compose -f docker-compose.prod.yml config
docker compose -f docker-compose.fullstack.yml build
```

## Results

| Check | Result |
| --- | --- |
| Go unit tests | Passed: `finalproject` and `finalproject/helpers`; other Go packages have no tests. |
| Frontend typecheck | Passed. |
| Frontend lint | Passed with zero warnings. |
| Frontend unit tests | Passed: 5 files, 10 tests. |
| Frontend production build | Passed. |
| Playwright smoke tests | Passed after baseline test-fixture correction: 12 tests across desktop, mobile, and tablet. |
| Docker Compose configuration | Both Compose files rendered successfully using runtime-only placeholder values. |
| Docker builds | API and frontend images built successfully. |

No test is failing at the end of this baseline.

## Changes made

- Copied the source working tree into this assessment repository without copying its `.git` directory or history.
- Added credential and local-secret patterns to root Git/Docker ignore files and to the frontend Docker ignore file. `.env.example` remains explicitly allowed.
- Corrected the Playwright social-link fixture so the link created by the test is distinct from the seeded link; this eliminates its strict-mode duplicate-locator failure without changing product behavior.
- Added this baseline report.

## Deployment risks identified

- `docker-compose.fullstack.yml` has development defaults, including a default PostgreSQL password; deployment must use a private secret-managed `.env` file with strong values.
- Both Compose files retain the obsolete `version` key. Docker Compose v5 accepts it but warns that it is ignored.
- The frontend Docker image uses Node 20 while one development dependency (`start-server-and-test`) declares Node 22+ or 24+; the production image build currently succeeds, but future scripts may require a Node upgrade.
- There is no dedicated backend database integration-test suite. Existing Go tests, frontend unit tests, and mocked Playwright smoke tests do not validate a running PostgreSQL/API integration path.
- Production requires externally provisioned PostgreSQL, Garage S3-compatible storage, and all required runtime secrets. No GCP, Redis, or deployment action was performed in this stage.

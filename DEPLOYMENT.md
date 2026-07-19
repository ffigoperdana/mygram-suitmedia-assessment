# MyGram Deployment Plan

This plan targets a homelab Coolify deployment with Jenkins as the main deployment runner and GHCR as the container registry.

## Recommended Architecture

Use one Docker Compose stack in Coolify for the first production version:

- `frontend`: React static build served by Nginx.
- `api`: Go Gin backend.
- `postgres`: persistent PostgreSQL database.

This is the best fit for a small fullstack MyGram app on a homelab. Split services later only if you need independent scaling, separate databases, or more strict network boundaries.

## Files Created Or Updated

- `.env.example`: documented local and production variables.
- `.env`: local-only development values, ignored by git.
- `docker-compose.fullstack.yml`: local fullstack build from source.
- `docker-compose.prod.yml`: production stack using GHCR images.
- `coolify.yml`: metadata pointing at the production compose file.
- `mygram-frontend/Dockerfile`: builds and serves React with Nginx.
- `mygram-frontend/nginx.conf`: SPA routing support.

## Domains

Use three domains/subdomains:

- Frontend: `https://mygram.example.com`
- API: same-origin through `https://mygram.example.com/api/*` for the first production stack. A separate `https://api.mygram.example.com` domain is optional once the proxy/TLS route is verified.
- API docs: `https://docs.mygram.example.com`

In production, set:

```bash
CORS_ALLOWED_ORIGINS=https://mygram.example.com,https://docs.mygram.example.com
PUBLIC_API_BASE_URL=
PUBLIC_DOCS_BASE_URL=https://docs.mygram.example.com
PUBLIC_OPENAPI_ENABLED=true
SWAGGER_UI_MODE=public
```

Replace the examples with your real Coolify domains.

When `PUBLIC_API_BASE_URL` is empty, the frontend calls `/api/v1/*` on the same origin and the frontend Nginx container proxies API requests to the backend service at `api:8080`. This avoids browser TLS/CORS failures while the separate API subdomain is still being stabilized.

The docs domain should show human-readable user API docs at `/`. If `/swagger` is opened on the docs domain, show Swagger UI backed by `/openapi/public.json`. Do not expose admin endpoints in public Swagger; keep the full/internal OpenAPI available only to trusted developers or behind admin/internal access.

## Required Secrets

Set these in Jenkins and/or Coolify. Do not commit them.

```bash
DB_PASSWORD=<strong-postgres-password>
JWT_SECRET=<random-32-plus-character-secret>
BACKEND_IMAGE=ghcr.io/<owner>/<repo>-api:<sha-or-tag>
FRONTEND_IMAGE=ghcr.io/<owner>/<repo>-web:<sha-or-tag>
CORS_ALLOWED_ORIGINS=https://mygram.example.com
PUBLIC_OPENAPI_ENABLED=true
SWAGGER_UI_MODE=public
CAP_ENABLED=true
CAP_BASE_URL=https://cap.fgdev.tech
CAP_SITE_KEY=<cap-site-key>
CAP_SECRET_KEY=<cap-secret-key>
CAP_REQUIRED_ON_LOGIN=true
S3_ENDPOINT=https://s3.fgdev.tech
S3_REGION=garage
S3_BUCKET=fgdev-media
S3_ACCESS_KEY_ID=<garage-access-key>
S3_SECRET_ACCESS_KEY=<garage-secret-key>
S3_FORCE_PATH_STYLE=true
S3_PUBLIC_BASE_URL=https://mygram.example.com/media
S3_UPLOAD_MAX_MB=4
```

For Jenkins:

```bash
GHCR_USERNAME=<github-username-or-org>
GHCR_TOKEN=<github-token-with-package-write>
COOLIFY_BASE_URL=http://127.0.0.1:8000
COOLIFY_RESOURCE_UUID=<mygram-coolify-application-uuid>
```

Store the Coolify API token as a Jenkins secret text credential named `coolify-api-token`. Do not use a stale `coolify-app-uuid` credential unless it actually contains the MyGram application UUID shown in the Coolify application URL.

The Jenkins deploy stage runs the Coolify API call from a temporary `curlimages/curl` container using Docker `--network host`. That is why `COOLIFY_BASE_URL` should normally be `http://127.0.0.1:8000`: the request is made from the Docker host network, not from inside the Jenkins container network.

## Optional Local Fullstack Docker Test

Local Docker testing is optional. For daily local work, run PostgreSQL through Laragon, start the backend with Go, and start the frontend with Vite. Use GitHub Actions/Jenkins as the preferred Docker build verification path.

```bash
docker compose -f docker-compose.fullstack.yml --env-file .env up --build
```

Expected local URLs:

- Frontend: `http://localhost:3000`
- API: `http://localhost:8080`
- Health: `http://localhost:8080/health`
- Swagger: `http://localhost:8080/swagger/index.html`
- Public OpenAPI: `http://localhost:8080/openapi/public.json`
- Docs: `http://localhost:3000/docs` during frontend development, then `https://docs.mygram.example.com` in production.

## Jenkins Pipeline Shape

Use Jenkins for deployment because it already lives in your homelab and can be connected to Coolify cleanly.

Pipeline stages:

1. Checkout repository from GitHub webhook.
2. Run backend checks:

```bash
go mod download
go test ./...
go vet ./...
```

3. Run frontend checks:

```bash
cd mygram-frontend
npm ci
npm run lint
npm run typecheck
npm run build
```

4. Build and push images:

```bash
docker build -t ghcr.io/<owner>/<repo>-api:${GIT_COMMIT} .
docker build \
  --build-arg VITE_API_BASE_URL= \
  -t ghcr.io/<owner>/<repo>-web:${GIT_COMMIT} \
  ./mygram-frontend

docker push ghcr.io/<owner>/<repo>-api:${GIT_COMMIT}
docker push ghcr.io/<owner>/<repo>-web:${GIT_COMMIT}
```

5. Update Coolify environment variables:

```bash
BACKEND_IMAGE=ghcr.io/<owner>/<repo>-api:main
FRONTEND_IMAGE=ghcr.io/<owner>/<repo>-web:main
```

6. Trigger Coolify redeploy using the Coolify API:

```bash
curl --fail --show-error --location --request GET \
  "http://127.0.0.1:8000/api/v1/deploy?uuid=<mygram-coolify-application-uuid>&force=false" \
  --header "Authorization: Bearer <coolify-api-token>"
```

For production CI/CD, Jenkins pushes both immutable SHA tags and mutable `:main` tags from the `main` branch. Coolify should use the mutable `:main` image tags so the Jenkins deploy webhook can redeploy the newest images without manually editing Coolify variables.

## GitHub Actions Role

Keep GitHub Actions smaller than Jenkins:

- Run on pull requests.
- Run Go tests and frontend build.
- Build-check backend and frontend Dockerfiles with `push: false`.
- Validate `docker-compose.prod.yml` and optional `docker-compose.fullstack.yml` config with dummy CI env values.
- Optionally run Playwright smoke tests after backend/frontend are stable.
- Do not deploy from GitHub Actions if Jenkins is the chosen deployment owner.

This avoids two systems racing to deploy the same app.

## Coolify Setup Steps

1. Create a new Coolify project named `mygram-fullstack`.
2. Create an application from the GitHub repository.
3. Select Docker Compose deployment.
4. Use `docker-compose.prod.yml`.
5. Configure environment variables:

```bash
DB_NAME=finalproject
DB_USER=postgres
DB_PASSWORD=<strong-postgres-password>
JWT_SECRET=<random-32-plus-character-secret>
JWT_EXPIRATION_HOURS=24
CORS_ALLOWED_ORIGINS=https://mygram.example.com,https://docs.mygram.example.com
PUBLIC_OPENAPI_ENABLED=true
SWAGGER_UI_MODE=public
CAP_ENABLED=true
CAP_BASE_URL=https://cap.fgdev.tech
CAP_SITE_KEY=<cap-site-key>
CAP_SECRET_KEY=<cap-secret-key>
CAP_REQUIRED_ON_LOGIN=true
BOOTSTRAP_ADMIN_EMAIL=
BOOTSTRAP_ADMIN_USERNAME=
BOOTSTRAP_ADMIN_PASSWORD=
BOOTSTRAP_ADMIN_AGE=21
BOOTSTRAP_USER_EMAIL=
BOOTSTRAP_USER_USERNAME=
BOOTSTRAP_USER_PASSWORD=
BOOTSTRAP_USER_AGE=18
S3_ENDPOINT=https://s3.fgdev.tech
S3_REGION=garage
S3_BUCKET=fgdev-media
S3_ACCESS_KEY_ID=<garage-access-key>
S3_SECRET_ACCESS_KEY=<garage-secret-key>
S3_FORCE_PATH_STYLE=true
S3_PUBLIC_BASE_URL=https://mygram.example.com/media
S3_UPLOAD_MAX_MB=4
BACKEND_IMAGE=ghcr.io/<owner>/<repo>-api:main
FRONTEND_IMAGE=ghcr.io/<owner>/<repo>-web:main
```

6. Add domains:

- `frontend` service on port `80`: `https://mygram.example.com`
- `api` service on port `8080`: keep internal for same-origin `/api/*` proxy, or expose `https://api.mygram.example.com` only after the TLS/proxy route is confirmed.
- `docs` frontend route/service on port `80`: `https://docs.mygram.example.com`

7. Enable HTTPS through Coolify.
8. Enable persistent volume for Postgres.
9. Enable automated Postgres backups.
10. Deploy.

## Production Verification

Check these after each deployment:

```bash
curl https://mygram.example.com/health
curl https://mygram.example.com/api/v1/photos
curl https://api.mygram.example.com/health # optional, only if the separate API domain is enabled
```

Then verify in browser:

- Frontend loads over HTTPS.
- Docs domain loads over HTTPS.
- `/swagger` on the docs domain shows only public user API endpoints.
- `/openapi/public.json` does not include `/api/v1/admin/*` or legacy course-project routes.
- Register works.
- Login returns and stores a token.
- Upload image and create photo works.
- Comments work.
- Social links work.
- Admin dashboard is visible only to admin users.
- PWA install prompt appears once after eligibility and does not repeat on every route change or revisit.
- Logout works.
- Refreshing a protected page keeps auth state if token is valid.

## Rollback

Keep image tags by git SHA. To rollback:

1. Find the previous good backend and frontend image SHAs in GHCR.
2. Change `BACKEND_IMAGE` and `FRONTEND_IMAGE` in Coolify to those tags.
3. Redeploy the Compose stack.
4. Verify health and core workflows.

## Notes For Backend Completion

Deployment should wait until these backend tasks are done:

- `database.StartDB()` is called from `main.go`.
- DB config comes from env variables.
- JWT secret comes from env variables.
- JWT expires after 24 hours.
- CORS is enabled.
- `/health/ready` does not panic when DB is unavailable.
- API tests pass against a test database.

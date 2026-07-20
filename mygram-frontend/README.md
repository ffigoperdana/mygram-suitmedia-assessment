# MyGram Frontend

React + Vite + TypeScript frontend for the MyGram backend API.

## Stack

- React Router for pages
- TanStack Query for server state
- Zustand for persisted auth state
- Tailwind CSS with shadcn/ui-compatible primitives
- Axios API client pointed at `VITE_API_BASE_URL`

## Implemented Areas

- Register, login, logout, persisted session, `/api/v1/me`, and banned-session handling
- Feed with search/filter, photo detail, comments, social links, and editable profile pages
- RBAC-aware navigation and admin-only route guard
- Admin dashboard with stats, user filtering, edit, ban, unban, and delete actions
- Human-readable API docs plus `/swagger` route for the backend public Swagger UI
- PWA manifest, service worker, offline shell, and one-time install prompt behavior

## Local Development

```bash
cd mygram-frontend
npm ci
npm run dev
```

Create `.env` from `.env.example` and set:

```bash
VITE_API_BASE_URL=http://localhost:8080
VITE_USE_SAME_ORIGIN_API=false
```

For production, set `VITE_USE_SAME_ORIGIN_API=true` when the frontend Nginx container proxies same-origin `/api/*` requests to the backend service. Use a full `https://api...` URL only after that API subdomain's TLS/proxy route is verified.

## Quality Gates

```bash
npm run typecheck
npm run lint
npm run test
npm run build
npm run test:e2e
npm run lighthouse:mobile
```

Or run the combined local gate:

```bash
npm run quality:full
```

`go.mod` in this folder is intentional. It acts only as a Go module boundary so backend `go test ./...` does not scan JavaScript dependencies under `node_modules`.

`npm run lighthouse:mobile` builds against `vite preview` on port `4173` and asserts mobile Lighthouse scores of 90+ for `/docs` and `/login`.

Playwright covers Chrome desktop, tablet, and mobile viewports. Real iOS Safari and Firefox checks still need the relevant device/browser available.

See `../TASK.md` for remaining Phase D work such as docs-domain deployment and real-device installed-PWA verification.

# MyGram Fullstack Task Handoff

## Current Repository Facts

- Backend is a Go Gin API in the repository root with module name `finalproject`.
- Existing entities are users, photos, comments, and social media links.
- Current route style is course-project style, for example `/photos/getall`, `/comments/create/:photoId`.
- `main.go` calls `database.StartDB()` and fails clearly when DB startup fails.
- `database/db.go` reads PostgreSQL settings from environment variables and runs one `AutoMigrate`.
- JWT uses `github.com/golang-jwt/jwt/v5`, env-driven secret, and 24 hour default expiration.
- Users now include `role` (`user`/`admin`) and `status` (`active`/`banned`) fields for RBAC and moderation.
- Optional Cap captcha config is available through `CAP_ENABLED`, `CAP_BASE_URL`, `CAP_SITE_KEY`, `CAP_SECRET_KEY`, and `CAP_REQUIRED_ON_LOGIN`.
- Backend exposes `/openapi/public.json`, a filtered public OpenAPI spec that excludes admin and legacy routes.
- Backend supports authenticated photo image uploads to S3-compatible object storage through `/api/v1/uploads/photos`.
- Swagger UI is controlled by `SWAGGER_UI_MODE`: `internal`, `public`, or `disabled`.
- Router configures CORS from `CORS_ALLOWED_ORIGINS`.
- Backend tests now cover auth, users, photos, comments, social media, CORS, and health paths.
- The old `mygram-frontend/` attempt was empty/incomplete. A fresh React scaffold now exists there.
- Production deploy target is Coolify on a homelab, with GHCR as image registry and Jenkins as the main deployment pipeline.

## Phase A - Baseline And Cleanup

- [x] A1. Run `git status --short` and identify unrelated local changes before editing.
- [x] A2. Run `go test ./...` and record the starting result. Current backend result passes after using `C:\Program Files\Go\bin\go.exe` and local Go caches.
- [x] A3. Run `go vet ./...` and record current failures, if any. Current backend result passes.
- [x] A4. Decide whether to keep module name `finalproject` or rename to a MyGram-specific module. Kept `finalproject` to avoid a noisy module-wide rename in the same backend-hardening pass.
- [x] A5. Clean README mojibake/encoding and remove claims that are not true yet, especially CORS and production readiness.
- [x] A6. Keep `.env` ignored and maintain `.env.example` as the documented template.

Acceptance: baseline failures are known, docs no longer misrepresent the current app, and no unrelated generated files are committed.

## Phase B - Backend Production Readiness

- [x] B1. Create a typed backend config layer that reads `PORT`, `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`, `JWT_SECRET`, `JWT_EXPIRATION_HOURS`, `CORS_ALLOWED_ORIGINS`, and `GIN_MODE` from environment variables.
- [x] B2. Optionally load `.env` only for local development. Do not require `.env` in Docker or Coolify.
- [x] B3. Fix database startup: call `database.StartDB()` from `main.go`, make database initialization fail clearly, and remove duplicate `AutoMigrate`.
- [x] B4. Make readiness checks handle nil/uninitialized database safely instead of panicking.
- [x] B5. Replace `github.com/dgrijalva/jwt-go` with `github.com/golang-jwt/jwt/v5`.
- [x] B6. Add JWT expiration with a 24 hour default and verify `exp`, signing method, and token validity.
- [x] B7. Replace `jwt.MapClaims` casts across middleware/controllers with a typed claims helper or safe extraction function.
- [x] B8. Add CORS middleware using env-configured origins. Include local frontend origins and production domains.
- [x] B9. Normalize request binding: always check `ShouldBindJSON`/`ShouldBind` errors and return structured `400` responses.
- [x] B10. Normalize response bodies and status codes across user, photo, comment, and social media controllers.
- [x] B11. Ensure no password hash is ever returned from API responses.
- [x] B12. Revisit "empty list" behavior. Prefer `200 []` over `404` for empty feeds/comments/social links because it is easier for the frontend.
- [x] B13. Add consistent ownership authorization for update/delete and return `403 Forbidden` instead of `401 Unauthorized` when the user is authenticated but not owner.
- [x] B14. Add optional REST aliases under `/api/v1`, for example `/api/v1/photos`, while keeping old routes until the frontend migration is complete.
- [x] B15. Regenerate Swagger/OpenAPI after route, auth, and response changes.
- [x] B16. Add RBAC foundation with `user`/`admin` roles and admin-only middleware.
- [x] B17. Add user status moderation with `active`/`banned`, login blocking, and token blocking for banned users.
- [x] B18. Add optional Cap captcha verification for registration and login.
- [x] B19. Add admin dashboard API endpoints for stats, user list/detail, update, ban/unban, and delete.
- [x] B20. Add public user-only OpenAPI output at `/openapi/public.json`, prune admin/legacy routes from that output, and make `/swagger/*any` configurable as internal, public, or disabled.

Acceptance: backend starts with env config, connects to Postgres, JWTs expire after 24h, CORS works from the React dev server, and `go test ./...` passes.

## Phase C - Backend Test Coverage

- [x] C1. Create test helpers for a real isolated PostgreSQL test database.
- [x] C2. Add user register/login tests for success, duplicate email, invalid email, low age, short password, and invalid login.
- [x] C3. Add photo API tests for create, list, get, update owner, reject non-owner update, delete owner, and validation failures.
- [x] C4. Add comment API tests for create on existing photo, reject missing photo, list by photo, update owner, reject non-owner update, and delete.
- [x] C5. Add social media API tests for create/list/update/delete plus URL validation.
- [x] C6. Add auth middleware tests for missing bearer token, malformed token, expired token, and valid token.
- [x] C7. Add readiness/liveness tests that cover DB connected and DB unavailable paths.
- [ ] C8. Add coverage output to CI. Target meaningful controller/middleware coverage, not only line percentage.
- [x] C9. Add API tests for `/api/v1/me`, RBAC admin access, admin stats/users, ban/unban, and banned-user blocking.

Acceptance: tests can run locally and in CI without touching the development database.

## Phase D - Frontend Implementation

Recommended stack: React + Vite + TypeScript, Tailwind, shadcn/ui-compatible primitives, TanStack Query for server state, Zustand for small persisted auth state. This is more production-appropriate than putting all API data in Redux. Redux is not needed unless the app later grows complex client-only workflows.

Phase D must treat the backend as already role-aware:

- Public user features use only `/api/v1/auth/*`, `/api/v1/me`, `/api/v1/photos`, `/api/v1/comments`, `/api/v1/photos/:photoId/comments`, and `/api/v1/social-media`.
- Admin-only features use `/api/v1/admin/*` and must be hidden from normal users in navigation and blocked by route guards.
- Cap captcha is frontend-visible only as a widget/token flow. Never expose `CAP_SECRET_KEY` in frontend code.
- API documentation for external users should be a user-facing docs site, not Swagger as the main experience.

- [x] D1. Run `cd mygram-frontend && npm install` and commit the generated lockfile.
- [x] D2. Run `npm run typecheck`, `npm run lint`, and `npm run build`; fix scaffold issues before adding features.
- [x] D3. Initialize or finish shadcn/ui components needed by the app: button, input, textarea, label, card, dialog, dropdown menu, alert, form, avatar, skeleton, toast.
- [x] D4. Implement auth pages with validation, API errors, persisted token, logout, and expired-token handling.
- [x] D4a. Integrate Cap captcha widget using frontend-safe env values only, for example `VITE_CAP_BASE_URL` and `VITE_CAP_SITE_KEY`, on register and, if enabled, login. The frontend submits `captcha_token`; it must never contain or log the Cap secret key.
- [x] D4b. Handle banned-user responses cleanly: if login or `/api/v1/me` returns `403` with banned account messaging, clear local auth and show a calm account-status screen.
- [x] D4c. Store auth in a production-minded way. Prefer an access-token storage approach that matches the current backend, but isolate it behind an auth store so it can move to HttpOnly cookies later without rewriting the app.
- [x] D5. Implement feed page with photo list, create photo, edit/delete owner controls, loading states, empty states, and responsive layout.
- [x] D6. Implement photo detail page with comments list, create/edit/delete comment, owner controls, and image fallback handling.
- [x] D7. Implement social links page with create/edit/delete and URL validation.
- [x] D8. Implement profile page using token/user data. If backend adds `/me`, use that instead of decoding token client-side.
- [x] D8a. Show admin dashboard entry only when `/api/v1/me` returns `role: "admin"`.
- [x] D8b. Add frontend route guards: anonymous users go to login, normal users cannot open admin routes, admins can open admin routes, and all guards survive browser refresh.
- [x] D8c. Implement admin dashboard overview using `/api/v1/admin/stats`: total users, active users, banned users, admin users, users seen in last 24h, total photos, total comments, total social media, and recent users.
- [x] D8d. Implement admin user table using `/api/v1/admin/users`: pagination, search by username/email, filters for role/status, empty states, loading skeleton, and error recovery.
- [x] D8e. Implement admin user detail/edit actions: update username/email/age/role/status, ban with reason, unban, and delete. Add confirmation dialogs for destructive actions and prevent UI attempts to ban/delete the current admin account.
- [x] D8f. Keep admin UI restrained and operational, not marketing-style: dense table, clear filters, predictable actions, and no decorative hero layout.
- [x] D9. Add an API compatibility layer for old routes and switch frontend calls to `/api/v1` routes. Do not use legacy course-project routes in new frontend code unless explicitly needed for backwards compatibility.
- [x] D10. Add frontend tests with Vitest + React Testing Library for auth, route guards, admin-only navigation, banned-user handling, API error rendering, and Cap captcha token submission shape.
- [x] D11. Add Playwright smoke tests for register/login, create photo, add comment, social link CRUD, admin dashboard access as admin, admin route rejection as normal user, and banned-user blocking.
- [x] D12. Polish accessibility: labels, keyboard navigation, focus rings, alt text, route titles, landmarks, reduced-motion handling, and non-overlapping responsive layouts.
- [x] D13. Build a human-readable API documentation experience for external users, separate from Swagger. It should explain authentication, headers, request/response examples, validation errors, and examples for photos, comments, and social media posting.
- [ ] D13a. Deploy the documentation experience on `https://docs.<your-domain>` as the primary API docs landing page. It should not require login just to read user API docs.
- [x] D13b. Include copyable examples for cURL, JavaScript `fetch`, and a short frontend integration example. Examples must use public user endpoints only and placeholder tokens.
- [x] D13c. Add an optional "Try request" console only if it is safe: it must require the user's own JWT, must never persist JWT beyond the current tab unless the user explicitly opts in, and must clearly target the configured API base URL.
- [x] D13d. Add a `/swagger` path on the docs domain that opens Swagger UI using `/openapi/public.json`. Public Swagger must show only user-facing endpoints, not `/api/v1/admin/*`.
- [x] D13e. Backend now serves a filtered public OpenAPI spec at `/openapi/public.json`. Frontend/docs work should consume this spec instead of the full internal Swagger.
- [x] D14. Add PWA support: web manifest, app icons, maskable icon, service worker, offline fallback shell, cache strategy for static assets, and safe handling for API calls while offline.
- [x] D14a. Implement an install prompt that appears only when install criteria are met and only once per user/browser after dismissal. Store dismissal state in local storage with a sensible cool-down or permanent "do not show again" flag.
- [x] D14b. Ensure PWA install behavior works on Android Chrome and degrades gracefully on iOS Safari, including iOS-friendly instructions only when needed. Do not show repeated popups during route changes or repeated visits.
- [ ] D14c. Manual device verification pending: install the PWA on Android Chrome, iOS Safari, and desktop Chromium where possible. Verify correct name/icon, standalone display mode, deep links route back into the SPA, auth state is respected after launch, and offline shell is not mistaken for a logged-in state.
- [x] D15. Add final frontend quality gates: mobile Lighthouse score 90+ for Performance, Accessibility, Best Practices, and SEO on key public pages; keep authenticated dashboards fast and responsive even if Lighthouse is run manually.
- [x] D15a. Test cross-platform layouts on mobile, tablet, and desktop breakpoints, including iOS Safari, Android Chrome, Chromium desktop, and Firefox desktop where possible.
- [x] D15b. Add a final "no slop" pass: remove placeholder text, console logs, unused components, dead routes, layout jumps, clipped text, repeated toasts/popups, and inconsistent loading/error states.
- [x] D15c. Add bundle/performance checks: lazy-load admin/docs-heavy routes, keep image sizes controlled, avoid shipping admin-only UI to anonymous landing flows when practical, and verify no secret env values are included in built assets.

Optional product improvements:

- [x] D16. Add simple search/filter for photos.
- [x] D17. Add profile editing if backend supports it.
- [x] D18. Add image upload to object storage using the backend as the trusted S3/Garage proxy. Frontend create/edit flows can upload an image file, receive the object URL, and still support manual image URLs.

Acceptance: frontend builds, can register/login against local backend, supports the core MyGram workflows, respects RBAC, includes admin dashboard flows, provides user-facing API docs, installs cleanly as a PWA, and passes the final quality gates.

## Phase E - Docker And CI Build Verification

- [x] E1. Configure GitHub Actions and Jenkins to build-check the Go API Dockerfile without requiring local Docker builds. First remote green run is still pending.
- [x] E2. Configure GitHub Actions and Jenkins to build-check `mygram-frontend/Dockerfile` with `npm ci`, frontend-safe build args, and the lockfile. First remote green run is still pending.
- [x] E3. Keep `docker-compose.fullstack.yml` as an optional local fullstack smoke path, but prefer local CLI commands for day-to-day development.
- [x] E4. Use `docker-compose.prod.yml` for Coolify production images.
- [x] E5. Remove stale or conflicting Compose services, especially optional Redis or old Nginx references, unless they are actually used.
- [x] E6. Add health checks for `postgres`, `api`, and `frontend` in Dockerfiles/Compose and add CI/Jenkins compose config checks. First remote runtime health check is still pending.
- [ ] E7. Run the first GitHub Actions and Jenkins Phase E checks remotely and record the green build URLs before Phase F deploy automation.

Acceptance: backend and frontend images build in CI/Jenkins, production compose is valid for Coolify, and local development remains runnable without Docker via `go run main.go` plus `npm run dev`.

## Phase F - CI/CD

Recommended split:

- GitHub Actions: PR quality gate and lightweight smoke/e2e.
- Jenkins: main branch build, GHCR push, and Coolify deployment trigger from homelab.

- [ ] F1. Remove or disable stale duplicate workflows after creating the final workflow.
- [ ] F2. Create one GitHub Actions workflow for PRs: Go test/vet, frontend lint/typecheck/build, and optional Playwright smoke.
- [ ] F3. Rewrite `Jenkinsfile` for fullstack: checkout, test backend, test frontend, build backend image, build frontend image, push both to GHCR, trigger Coolify redeploy.
- [ ] F4. Store GHCR credentials, the `coolify-api-token` Jenkins credential, the MyGram Coolify resource UUID/base URL, database password, JWT secret, and production domains in Jenkins/Coolify secrets.
- [ ] F5. Tag images with git SHA and optionally `latest` on main.
- [ ] F6. Add rollback instructions: redeploy previous GHCR SHA image tags in Coolify.
- [ ] F7. Add CI security gates: dependency audit, secret scan, and container image vulnerability scan before deploy.

Acceptance: pushing to main triggers Jenkins, pushes both images to GHCR, and Coolify redeploys successfully.

## Phase G - Coolify Production

- [ ] G1. Create or update the Coolify project for `mygram-fullstack`.
- [ ] G2. Use `docker-compose.prod.yml` as the deployment compose file.
- [ ] G3. Configure frontend domain, for example `https://mygram.example.com`.
- [ ] G4. Configure API access. Prefer same-origin `https://mygram.example.com/api/*` through the frontend Nginx proxy for the first production stack; add `https://api.mygram.example.com` only after its TLS/proxy route is verified.
- [ ] G5. Configure docs domain, for example `https://docs.mygram.example.com`. Route `/` to the human-readable API docs and `/swagger` to the public user-only Swagger UI.
- [ ] G6. Set `CORS_ALLOWED_ORIGINS` to include the frontend domain and docs domain if the docs site includes a safe "Try request" console.
- [ ] G7. Set Cap captcha production env values: `CAP_ENABLED=true`, `CAP_BASE_URL=https://cap.fgdev.tech`, `CAP_SITE_KEY`, `CAP_SECRET_KEY`, and `CAP_REQUIRED_ON_LOGIN` according to the chosen login friction.
- [ ] G7a. Set production docs behavior: `PUBLIC_OPENAPI_ENABLED=true` and `SWAGGER_UI_MODE=public` if the API container serves docs-domain `/swagger`, or `SWAGGER_UI_MODE=disabled` if Swagger UI is served only by the frontend docs app.
- [ ] G8. Set persistent volume for Postgres.
- [ ] G9. Verify `/health`, `/health/live`, `/health/ready`, public docs, Swagger user-only view, login, feed, comments, social links, admin dashboard, and PWA install behavior in production.
- [ ] G10. Configure automated Postgres backups in Coolify.
- [ ] G11. Add production rate limiting for auth, captcha, upload, and public docs/API routes through the reverse proxy or a backend middleware.
- [ ] G12. Revisit auth storage after deployment hardening. If the backend moves from bearer tokens to HttpOnly Secure SameSite cookies, update the auth store without changing app workflows.

Acceptance: production frontend loads over HTTPS, calls production API over HTTPS, docs domain works, public Swagger hides admin endpoints, PWA install works, and the stack survives container restart with database data intact.

## Final Definition Of Done

- Backend reads all secrets/config from env.
- JWT uses `golang-jwt/jwt/v5` with 24h expiration.
- CORS allows local and production frontend origins.
- RBAC, Cap captcha config, and admin user-management APIs are implemented and documented.
- Photo image upload uses backend-only S3/Garage credentials; no object storage secret is exposed to frontend code.
- Comprehensive backend API tests pass.
- React frontend is implemented and production-built.
- User-facing API docs are available on `docs.<domain>`, with public Swagger at `/swagger` showing only user endpoints.
- PWA install works on Android and degrades cleanly on iOS without repeated prompts.
- Final frontend quality pass reaches mobile Lighthouse 90+ on key public pages and has no obvious placeholder/sloppy UI states.
- Docker images build in CI/Jenkins; local Docker compose remains optional.
- Coolify production compose works.
- Jenkins deploy pipeline pushes GHCR images and triggers redeploy.
- GitHub Actions protects PR quality.
- README, Swagger, `TASK.md`, and `DEPLOYMENT.md` match the real project.

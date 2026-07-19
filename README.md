# MyGram - Social Media Backend API

MyGram is a Go/Gin social media backend for users, photos, comments, and social media links. This repository is being prepared for a fullstack app with:

- Backend: Go, Gin, GORM, PostgreSQL
- Frontend: React, Vite, TypeScript, Tailwind
- Deployment: Docker Compose on Coolify, with GHCR images and Jenkins pipeline support

See [TASK.md](TASK.md) for the phased implementation handoff and [DEPLOYMENT.md](DEPLOYMENT.md) for the Coolify/Jenkins deployment plan.

## Current Backend Features

- User registration and login
- Password hashing with bcrypt
- JWT authentication with 24 hour token expiration
- RBAC with `user` and `admin` roles
- Optional Cap captcha verification for registration/login
- Admin dashboard API for stats, user listing, user updates, ban/unban, and delete
- Photo CRUD with ownership authorization
- Authenticated image upload to S3-compatible object storage for photo media
- Same-origin media proxy at `/media/uploads/photos/*` for uploaded images
- Comment CRUD with ownership authorization
- Social media link CRUD with ownership authorization
- Health, liveness, and readiness endpoints
- Public OpenAPI spec at `/openapi/public.json`
- Swagger UI at `/swagger/index.html`, configurable as `internal`, `public`, or `disabled`
- CORS middleware for frontend integration
- Env-driven database, JWT, CORS, and port configuration

## API Endpoints

Legacy endpoints are still available:

```text
POST   /users/register
POST   /users/login

POST   /photos/create
GET    /photos/getall
GET    /photos/get/:photoId
PUT    /photos/update/:photoId
DELETE /photos/delete/:photoId

POST   /comments/create/:photoId
GET    /comments/getall
GET    /comments/getall/:photoId
GET    /comments/get/:commentId
PUT    /comments/update/:commentId
DELETE /comments/delete/:commentId

POST   /socialmedia/create
GET    /socialmedia/getall
GET    /socialmedia/get/:socialMediaId
PUT    /socialmedia/update/:socialMediaId
DELETE /socialmedia/delete/:socialMediaId
```

Cleaner aliases are also available under `/api/v1`:

```text
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/me
PATCH  /api/v1/me

POST   /api/v1/photos
GET    /api/v1/photos
GET    /api/v1/photos/:photoId
PUT    /api/v1/photos/:photoId
DELETE /api/v1/photos/:photoId

POST   /api/v1/uploads/photos
GET    /media/uploads/photos/*

GET    /api/v1/comments
GET    /api/v1/comments/:commentId
PUT    /api/v1/comments/:commentId
DELETE /api/v1/comments/:commentId

GET    /api/v1/photos/:photoId/comments
POST   /api/v1/photos/:photoId/comments

POST   /api/v1/social-media
GET    /api/v1/social-media
GET    /api/v1/social-media/:socialMediaId
PUT    /api/v1/social-media/:socialMediaId
DELETE /api/v1/social-media/:socialMediaId

GET    /api/v1/admin/stats
GET    /api/v1/admin/users
GET    /api/v1/admin/users/:userId
PATCH  /api/v1/admin/users/:userId
DELETE /api/v1/admin/users/:userId
POST   /api/v1/admin/users/:userId/ban
POST   /api/v1/admin/users/:userId/unban
```

Protected routes require:

```text
Authorization: Bearer <jwt>
```

## Environment

Copy `.env.example` to `.env` for local development.

Required backend variables:

```bash
PORT=8080
GIN_MODE=debug
JWT_SECRET=replace-with-a-random-32-plus-character-secret
JWT_EXPIRATION_HOURS=24
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:5173
PUBLIC_OPENAPI_ENABLED=true
SWAGGER_UI_MODE=internal

CAP_ENABLED=false
CAP_BASE_URL=https://cap.fgdev.tech
CAP_SITE_KEY=replace-with-cap-site-key
CAP_SECRET_KEY=replace-with-cap-secret-key
CAP_REQUIRED_ON_LOGIN=true

S3_ENDPOINT=https://s3.fgdev.tech
S3_REGION=garage
S3_BUCKET=fgdev-media
S3_ACCESS_KEY_ID=replace-with-garage-access-key
S3_SECRET_ACCESS_KEY=replace-with-garage-secret-key
S3_FORCE_PATH_STYLE=true
# Prefer this in production so images load through the MyGram domain.
S3_PUBLIC_BASE_URL=https://mygram.example.com/media
S3_UPLOAD_MAX_MB=4

DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=admin
DB_NAME=finalproject
DB_PORT=5432
DB_SSLMODE=disable
```

## Local Development Without Docker

```bash
go mod download
go test ./...
go run main.go
```

The backend starts on `http://localhost:8080` by default.

In another terminal:

```bash
cd mygram-frontend
npm ci
npm run dev
```

The frontend dev server starts on `http://localhost:3000` by default. Use Laragon/PostgreSQL locally and keep your local `.env` out of git.

## Optional Local Fullstack Docker

Local Docker is optional. The preferred Docker verification path is GitHub Actions and Jenkins before Coolify deployment.

```bash
docker compose -f docker-compose.fullstack.yml --env-file .env up --build
```

The active production compose file is `docker-compose.prod.yml`. Older Redis/test compose files were removed so there is only one optional local fullstack compose and one Coolify production compose.

Expected local URLs:

- Frontend: `http://localhost:3000`
- API: `http://localhost:8080`
- Public OpenAPI: `http://localhost:8080/openapi/public.json`
- Swagger: `http://localhost:8080/swagger/index.html`

`SWAGGER_UI_MODE=internal` shows the full developer Swagger UI. `SWAGGER_UI_MODE=public` shows only public user-facing endpoints. `SWAGGER_UI_MODE=disabled` removes the Swagger UI route while keeping `/openapi/public.json` if `PUBLIC_OPENAPI_ENABLED=true`.

## Test Database

The API tests expect a PostgreSQL database named `finalproject_test` using the env values from `main_test.go`. If the database is unavailable, DB-backed tests skip instead of touching the development database.

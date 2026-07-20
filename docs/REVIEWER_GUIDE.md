# MyGram Reviewer Guide

This guide is safe to keep in the public repository. It documents access and expected flows but intentionally contains no passwords, tokens, database credentials, or cloud credentials.

## Production URLs

- Frontend: <https://mygram-web-734925981385.asia-southeast2.run.app>
- API health: <https://mygram-api-734925981385.asia-southeast2.run.app/health>
- API readiness: <https://mygram-api-734925981385.asia-southeast2.run.app/health/ready>
- Public API guide: <https://mygram-web-734925981385.asia-southeast2.run.app/docs>
- Public OpenAPI: <https://mygram-api-734925981385.asia-southeast2.run.app/openapi/public.json>

The public OpenAPI document excludes `/api/v1/admin/*`. Administrative routes still require a valid JWT for a user with the `admin` role.

This is an assessment deployment rather than a hardened customer-production environment. HTTPS, security headers, RBAC, password hashing, Secret Manager, and Cloud SQL's authenticated socket are enabled. Redis-backed distributed rate limiting is enabled only after the Memorystore bootstrap and a successful CD run; it deliberately fails open during Redis outages, so public registration should use disposable review accounts and the service should not receive real personal or confidential data.

## Regular user flow

1. Open the frontend and select **Create an account**.
2. Register with a unique username and email, an age from 9 to 100, and a password of at least six characters.
3. Sign in with the registered email and password.
4. Review the feed and create a photo post. Uploaded media remains stored in Garage through its S3-compatible API.
5. Open a photo to add or manage comments.
6. Use **Social Links** to create, update, and remove social profiles.
7. Use **Profile** to update the current user's profile.
8. Review the public API documentation and sign out.

New public registrations always receive the `user` role. A user cannot assign the `admin` role through registration or profile updates.

## Administrator flow

1. Sign in with a separately shared administrator account.
2. Open **Admin Dashboard**. The navigation item and route are available only to an authenticated admin.
3. Review application statistics and the paginated user list.
4. Search or filter users by role and status.
5. Review a user, update allowed profile/role/status fields, or ban/unban the account.
6. Use destructive actions only on disposable review accounts. The API prevents an administrator from banning, deleting, or removing the admin role from their own account.

## Bootstrap account status

The Cloud Run CD workflow supplies these public reviewer identities:

- Regular user: `reviewer.user@example.com` (`reviewer-user`)
- Administrator: `reviewer.admin@example.com` (`reviewer-admin`)

Their passwords are not in GitHub or this guide. `scripts/gcp/bootstrap-reviewers.sh` stores them in `mygram-bootstrap-user-password` and `mygram-bootstrap-admin-password` in Secret Manager, and Cloud Run references the secrets without GitHub Actions reading their values.

On API startup, a missing reviewer row is created and its password is hashed before PostgreSQL storage. An existing matching account is reconciled to the latest secret password, restored to active status, and—only for the administrator—assigned the `admin` role. This makes password rotation deterministic while keeping the application database as the RBAC source of truth.

## Secure reviewer provisioning

Use the repository script from an authenticated Cloud Shell. It prompts twice with hidden input, requires at least 12 characters, creates or versions both secrets, grants the runtime identity access, sets the GitHub identity variables, and finally sets `REVIEWER_SEED_ENABLED=true`:

```bash
GITHUB_REPOSITORY=ffigoperdana/mygram-suitmedia-assessment \
  ./scripts/gcp/bootstrap-reviewers.sh
```

Then manually dispatch `CD - Cloud Run` or push the committed deployment configuration. Verify both logins after the CD job is green. Share the two passwords through a private channel or password manager, and delete or ban the accounts after the review if continued public access is unnecessary.

## Credential handoff template

Copy this template into a private message and replace the placeholders. Do not put the completed version in Git:

```text
MyGram technical assessment

Application: https://mygram-web-734925981385.asia-southeast2.run.app
Public API documentation: https://mygram-web-734925981385.asia-southeast2.run.app/docs

Regular reviewer account
Email: reviewer.user@example.com
Password: <share privately>

Administrator reviewer account
Email: reviewer.admin@example.com
Password: <share privately>

Suggested flow:
1. Sign in as the regular user and test feed, photo upload, comments, social links, profile, and logout.
2. Sign in as the administrator and open Admin Dashboard to review RBAC and user moderation.

The credentials are temporary and provided only for assessment review. Please do not reuse them on other systems or include them in public feedback, screenshots, or issue reports.
```

# MyGram Reviewer Guide

This guide is safe to keep in the public repository. It documents access and expected flows but intentionally contains no passwords, tokens, database credentials, or cloud credentials.

## Production URLs

- Frontend: <https://mygram-web-734925981385.asia-southeast2.run.app>
- API health: <https://mygram-api-734925981385.asia-southeast2.run.app/health>
- API readiness: <https://mygram-api-734925981385.asia-southeast2.run.app/health/ready>
- Public API guide: <https://mygram-web-734925981385.asia-southeast2.run.app/docs>
- Public OpenAPI: <https://mygram-api-734925981385.asia-southeast2.run.app/openapi/public.json>

The public OpenAPI document excludes `/api/v1/admin/*`. Administrative routes still require a valid JWT for a user with the `admin` role.

This is an assessment deployment rather than a hardened customer-production environment. HTTPS, security headers, RBAC, password hashing, Secret Manager, and Cloud SQL's authenticated socket are enabled. CAPTCHA and distributed rate limiting are currently disabled, so public registration should use disposable review accounts and the service should not receive real personal or confidential data.

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

The application supports optional `BOOTSTRAP_ADMIN_*` and `BOOTSTRAP_USER_*` environment variables. The Cloud Run CD workflow deliberately does not set them, so a normal deployment does not create hard-coded or publicly documented accounts.

When all bootstrap identity fields are blank, startup skips account creation. When a complete bootstrap identity is supplied, startup creates the account if it does not exist. Passwords are hashed by the model hook before storage. For an existing matching account, bootstrap can restore active status and ensure the admin role, but it does not replace that account's password.

## Secure one-time administrator provisioning

Use a dedicated Secret Manager secret and never commit the password. The following is an operator procedure, not a command for reviewers to run.

1. Create `mygram-bootstrap-admin-password` in Secret Manager, add a strong unique password as its current version, and grant `mygram-runtime@mygram-suitmedia-figo-2026.iam.gserviceaccount.com` Secret Accessor on that secret.
2. Choose an email and username used only for assessment review.
3. Temporarily update `mygram-api`:

   ```bash
   gcloud run services update mygram-api \
     --project=mygram-suitmedia-figo-2026 \
     --region=asia-southeast2 \
     --update-env-vars=BOOTSTRAP_ADMIN_EMAIL=reviewer-admin@example.com,BOOTSTRAP_ADMIN_USERNAME=reviewer-admin,BOOTSTRAP_ADMIN_AGE=21 \
     --update-secrets=BOOTSTRAP_ADMIN_PASSWORD=mygram-bootstrap-admin-password:latest
   ```

4. Wait for the revision to become ready and verify one successful login.
5. Remove the temporary bootstrap inputs. The database account remains:

   ```bash
   gcloud run services update mygram-api \
     --project=mygram-suitmedia-figo-2026 \
     --region=asia-southeast2 \
     --remove-env-vars=BOOTSTRAP_ADMIN_EMAIL,BOOTSTRAP_ADMIN_USERNAME,BOOTSTRAP_ADMIN_AGE \
     --remove-secrets=BOOTSTRAP_ADMIN_PASSWORD
   ```

6. Share the email and password through a private channel or password manager, never through the repository, a public issue, screenshots, or build logs.
7. Remove the runtime service account's access to the bootstrap secret and delete the secret when it is no longer required.
8. Delete or ban the reviewer accounts after the assessment if continued public access is unnecessary.

The next CD deployment also reconciles the API environment and secrets to the workflow's declared production configuration, which does not contain bootstrap credentials.

## Credential handoff template

Copy this template into a private message and replace the placeholders. Do not put the completed version in Git:

```text
MyGram technical assessment

Application: https://mygram-web-734925981385.asia-southeast2.run.app
Public API documentation: https://mygram-web-734925981385.asia-southeast2.run.app/docs

Regular reviewer account
Email: <share privately>
Password: <share privately>

Administrator reviewer account
Email: <share privately>
Password: <share privately>

Suggested flow:
1. Sign in as the regular user and test feed, photo upload, comments, social links, profile, and logout.
2. Sign in as the administrator and open Admin Dashboard to review RBAC and user moderation.

The credentials are temporary and provided only for assessment review. Please do not reuse them on other systems or include them in public feedback, screenshots, or issue reports.
```

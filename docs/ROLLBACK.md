# Cloud Run Rollback

Production images are tagged with the full Git commit SHA in Artifact Registry. Keep the backend and frontend on the same known-good SHA unless a specific mixed-version rollback has been tested.

## Roll back by image SHA

Authenticate with an operator identity that has Cloud Run deployment access, then set the previous successful SHA:

```bash
export PROJECT_ID="mygram-suitmedia-figo-2026"
export REGION="asia-southeast2"
export REGISTRY="asia-southeast2-docker.pkg.dev/mygram-suitmedia-figo-2026/mygram-containers"
export PREVIOUS_SHA="replace-with-previous-successful-full-commit-sha"

gcloud run services update mygram-api \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --image="$REGISTRY/mygram-api:$PREVIOUS_SHA"

gcloud run services update mygram-web \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --image="$REGISTRY/mygram-web:$PREVIOUS_SHA"
```

Run the same smoke tests used by CD after rollback:

```bash
API_URL="$(gcloud run services describe mygram-api --project="$PROJECT_ID" --region="$REGION" --format='value(status.url)')"
FRONTEND_URL="$(gcloud run services describe mygram-web --project="$PROJECT_ID" --region="$REGION" --format='value(status.url)')"

curl --fail-with-body "$FRONTEND_URL/"
curl --fail-with-body "$API_URL/health"
curl --fail-with-body "$API_URL/health/live"
curl --fail-with-body "$API_URL/health/ready"
```

## Roll back by existing Cloud Run revision

For the fastest application-only rollback, list revisions and route all traffic to the previous healthy revision:

```bash
gcloud run revisions list --service=mygram-api --project="$PROJECT_ID" --region="$REGION"
gcloud run services update-traffic mygram-api \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --to-revisions="PREVIOUS_API_REVISION=100"

gcloud run revisions list --service=mygram-web --project="$PROJECT_ID" --region="$REGION"
gcloud run services update-traffic mygram-web \
  --project="$PROJECT_ID" \
  --region="$REGION" \
  --to-revisions="PREVIOUS_WEB_REVISION=100"
```

Cloud Run configuration changes create revisions, so the API CORS update creates a revision after the initial API image deployment. Choose the revision whose image and environment configuration are both known-good.

Database schema compatibility must be checked before rolling back. The API currently runs GORM `AutoMigrate` during startup; an older image may not be compatible with a schema changed by a newer release. This workflow does not delete or roll back database data.

#!/usr/bin/env bash
set -Eeuo pipefail

PROJECT_ID="${GCP_PROJECT_ID:-mygram-suitmedia-figo-2026}"
RUNTIME_SERVICE_ACCOUNT="${RUNTIME_SERVICE_ACCOUNT:-mygram-runtime@${PROJECT_ID}.iam.gserviceaccount.com}"
ADMIN_EMAIL="${REVIEWER_ADMIN_EMAIL:-reviewer.admin@example.com}"
ADMIN_USERNAME="${REVIEWER_ADMIN_USERNAME:-reviewer-admin}"
USER_EMAIL="${REVIEWER_USER_EMAIL:-reviewer.user@example.com}"
USER_USERNAME="${REVIEWER_USER_USERNAME:-reviewer-user}"
ADMIN_SECRET="mygram-bootstrap-admin-password"
USER_SECRET="mygram-bootstrap-user-password"

for command_name in gcloud gh; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "Required command not found: ${command_name}" >&2
    exit 1
  fi
done

retry() {
  local max_attempts="$1"
  shift
  local attempt=1
  local delay_seconds=2

  until "$@"; do
    if (( attempt >= max_attempts )); then
      echo "Command failed after ${attempt} attempts: $*" >&2
      return 1
    fi
    echo "Command failed (attempt ${attempt}/${max_attempts}); retrying in ${delay_seconds}s..." >&2
    sleep "$delay_seconds"
    attempt=$((attempt + 1))
    delay_seconds=$((delay_seconds * 2))
  done
}

read_password() {
  local label="$1"
  local value=""
  local confirmation=""
  read -r -s -p "${label} password (minimum 12 characters): " value
  printf '\n' >&2
  read -r -s -p "Confirm ${label} password: " confirmation
  printf '\n' >&2
  if [[ "$value" != "$confirmation" ]]; then
    echo "${label} passwords do not match." >&2
    return 1
  fi
  if (( ${#value} < 12 )); then
    echo "${label} password must contain at least 12 characters." >&2
    return 1
  fi
  printf '%s' "$value"
}

gcloud auth list --filter=status:ACTIVE --format='value(account)' | grep -q . || {
  echo "Authenticate gcloud before running this script." >&2
  exit 1
}
gh auth status >/dev/null

REPOSITORY="${GITHUB_REPOSITORY:-$(gh repo view --json nameWithOwner --jq .nameWithOwner)}"
ADMIN_PASSWORD="${REVIEWER_ADMIN_PASSWORD:-$(read_password "Admin reviewer")}"
USER_PASSWORD="${REVIEWER_USER_PASSWORD:-$(read_password "Regular reviewer")}"

if (( ${#ADMIN_PASSWORD} < 12 || ${#USER_PASSWORD} < 12 )); then
  echo "Reviewer passwords must contain at least 12 characters." >&2
  exit 1
fi

gcloud config set project "$PROJECT_ID" >/dev/null
gcloud services enable secretmanager.googleapis.com --project="$PROJECT_ID"

put_secret() {
  local secret_name="$1"
  local secret_value="$2"
  if ! gcloud secrets describe "$secret_name" --project="$PROJECT_ID" >/dev/null 2>&1; then
    gcloud secrets create "$secret_name" \
      --project="$PROJECT_ID" \
      --replication-policy=automatic
  fi
  printf '%s' "$secret_value" | gcloud secrets versions add "$secret_name" \
    --project="$PROJECT_ID" \
    --data-file=- >/dev/null
  gcloud secrets add-iam-policy-binding "$secret_name" \
    --project="$PROJECT_ID" \
    --member="serviceAccount:${RUNTIME_SERVICE_ACCOUNT}" \
    --role=roles/secretmanager.secretAccessor \
    --quiet >/dev/null
}

put_secret "$ADMIN_SECRET" "$ADMIN_PASSWORD"
put_secret "$USER_SECRET" "$USER_PASSWORD"
unset ADMIN_PASSWORD USER_PASSWORD

retry 6 gh variable set REVIEWER_ADMIN_EMAIL --repo "$REPOSITORY" --body "$ADMIN_EMAIL"
retry 6 gh variable set REVIEWER_ADMIN_USERNAME --repo "$REPOSITORY" --body "$ADMIN_USERNAME"
retry 6 gh variable set REVIEWER_USER_EMAIL --repo "$REPOSITORY" --body "$USER_EMAIL"
retry 6 gh variable set REVIEWER_USER_USERNAME --repo "$REPOSITORY" --body "$USER_USERNAME"
retry 6 gh variable set REVIEWER_SEED_ENABLED --repo "$REPOSITORY" --body "true"

echo "Reviewer identities and Secret Manager passwords are ready."
echo "Admin login email: ${ADMIN_EMAIL}"
echo "Regular login email: ${USER_EMAIL}"
echo "Run the Cloud Run CD workflow to create or reconcile both accounts."

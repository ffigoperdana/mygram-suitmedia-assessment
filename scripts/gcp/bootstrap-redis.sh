#!/usr/bin/env bash
set -Eeuo pipefail

PROJECT_ID="${GCP_PROJECT_ID:-mygram-suitmedia-figo-2026}"
REGION="${GCP_REGION:-asia-southeast2}"
REDIS_INSTANCE="${REDIS_INSTANCE:-mygram-redis}"
VPC_NETWORK="${GCP_VPC_NETWORK:-mygram-vpc}"
VPC_SUBNET="${GCP_VPC_SUBNET:-mygram-cloud-run}"
SUBNET_RANGE="${GCP_VPC_SUBNET_RANGE:-10.20.0.0/24}"
DEPLOYER_SERVICE_ACCOUNT="${DEPLOYMENT_SERVICE_ACCOUNT:-github-mygram-deployer@mygram-suitmedia-figo-2026.iam.gserviceaccount.com}"

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

gcloud auth list --filter=status:ACTIVE --format='value(account)' | grep -q . || {
  echo "Authenticate gcloud before running this script." >&2
  exit 1
}
gh auth status >/dev/null

REPOSITORY="${GITHUB_REPOSITORY:-$(gh repo view --json nameWithOwner --jq .nameWithOwner)}"
PROJECT_NUMBER="$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')"
CLOUD_RUN_SERVICE_AGENT="service-${PROJECT_NUMBER}@serverless-robot-prod.iam.gserviceaccount.com"

gcloud config set project "$PROJECT_ID" >/dev/null
gcloud services enable redis.googleapis.com compute.googleapis.com --project="$PROJECT_ID"

if ! gcloud compute networks describe "$VPC_NETWORK" --project="$PROJECT_ID" >/dev/null 2>&1; then
  gcloud compute networks create "$VPC_NETWORK" \
    --project="$PROJECT_ID" \
    --subnet-mode=custom
fi

if ! gcloud compute networks subnets describe "$VPC_SUBNET" \
  --project="$PROJECT_ID" --region="$REGION" >/dev/null 2>&1; then
  gcloud compute networks subnets create "$VPC_SUBNET" \
    --project="$PROJECT_ID" \
    --region="$REGION" \
    --network="$VPC_NETWORK" \
    --range="$SUBNET_RANGE"
else
  EXISTING_SUBNET_NETWORK="$(gcloud compute networks subnets describe "$VPC_SUBNET" \
    --project="$PROJECT_ID" --region="$REGION" --format='value(network)')"
  if [[ "${EXISTING_SUBNET_NETWORK##*/}" != "$VPC_NETWORK" ]]; then
    echo "Subnet ${VPC_SUBNET} exists on a different VPC network." >&2
    exit 1
  fi
fi

for principal in "$DEPLOYER_SERVICE_ACCOUNT" "$CLOUD_RUN_SERVICE_AGENT"; do
  gcloud projects add-iam-policy-binding "$PROJECT_ID" \
    --member="serviceAccount:${principal}" \
    --role=roles/compute.networkUser \
    --condition=None \
    --quiet >/dev/null
done

if ! gcloud redis instances describe "$REDIS_INSTANCE" \
  --project="$PROJECT_ID" --region="$REGION" >/dev/null 2>&1; then
  gcloud redis instances create "$REDIS_INSTANCE" \
    --project="$PROJECT_ID" \
    --region="$REGION" \
    --tier=basic \
    --size=1 \
    --network="$VPC_NETWORK" \
    --connect-mode=direct-peering
else
  read -r EXISTING_TIER EXISTING_SIZE EXISTING_NETWORK < <(
    gcloud redis instances describe "$REDIS_INSTANCE" \
      --project="$PROJECT_ID" \
      --region="$REGION" \
      --format='value(tier,memorySizeGb,authorizedNetwork)'
  )
  if [[ "$EXISTING_TIER" != "BASIC" || "$EXISTING_SIZE" != "1" || \
    "${EXISTING_NETWORK##*/}" != "$VPC_NETWORK" ]]; then
    echo "Existing ${REDIS_INSTANCE} does not match Basic Tier 1 GB on ${VPC_NETWORK}." >&2
    exit 1
  fi
fi

REDIS_HOST="$(gcloud redis instances describe "$REDIS_INSTANCE" \
  --project="$PROJECT_ID" --region="$REGION" --format='value(host)')"
REDIS_PORT="$(gcloud redis instances describe "$REDIS_INSTANCE" \
  --project="$PROJECT_ID" --region="$REGION" --format='value(port)')"

if [[ -z "$REDIS_HOST" || -z "$REDIS_PORT" ]]; then
  echo "Memorystore did not return a host and port." >&2
  exit 1
fi

retry 6 gh variable set REDIS_ADDR --repo "$REPOSITORY" --body "${REDIS_HOST}:${REDIS_PORT}"
retry 6 gh variable set REDIS_HOST --repo "$REPOSITORY" --body "$REDIS_HOST"
retry 6 gh variable set REDIS_PORT --repo "$REPOSITORY" --body "$REDIS_PORT"
retry 6 gh variable set GCP_VPC_NETWORK --repo "$REPOSITORY" --body "$VPC_NETWORK"
retry 6 gh variable set GCP_VPC_SUBNET --repo "$REPOSITORY" --body "$VPC_SUBNET"

echo "Memorystore ${REDIS_INSTANCE} is ready at ${REDIS_HOST}:${REDIS_PORT}."
echo "GitHub Actions variables were updated for ${REPOSITORY}."

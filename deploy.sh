#!/bin/bash
set -e

PROJECT_ID="resqlink-58742"
REGION="asia-south1"

echo "=== RESQLINK Cloud Run Deployment ==="
echo "Project: $PROJECT_ID"
echo "Region: $REGION"
echo ""

# Enable required APIs
echo "[1/6] Enabling APIs..."
gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  --project=$PROJECT_ID

# Create Artifact Registry repo (if not exists)
echo "[2/6] Creating Artifact Registry..."
gcloud artifacts repositories create resqlink-docker \
  --repository-format=docker \
  --location=$REGION \
  --project=$PROJECT_ID 2>/dev/null || echo "  (already exists)"

# Build & deploy API
echo "[3/6] Building API Docker image via Cloud Build..."
gcloud builds submit \
  --tag ${REGION}-docker.pkg.dev/${PROJECT_ID}/resqlink-docker/resqlink-api:latest \
  --dockerfile=Dockerfile.api \
  --project=$PROJECT_ID

echo "[4/6] Deploying API to Cloud Run..."
gcloud run deploy resqlink-api \
  --image ${REGION}-docker.pkg.dev/${PROJECT_ID}/resqlink-docker/resqlink-api:latest \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars GCP_PROJECT_ID=$PROJECT_ID,GCP_LOCATION=$REGION \
  --memory 512Mi \
  --min-instances 0 \
  --max-instances 3 \
  --project=$PROJECT_ID

# Build & deploy Worker
echo "[5/6] Building Worker Docker image via Cloud Build..."
gcloud builds submit \
  --tag ${REGION}-docker.pkg.dev/${PROJECT_ID}/resqlink-docker/resqlink-worker:latest \
  --dockerfile=Dockerfile.worker \
  --project=$PROJECT_ID

echo "[6/6] Deploying Worker to Cloud Run..."
gcloud run deploy resqlink-worker \
  --image ${REGION}-docker.pkg.dev/${PROJECT_ID}/resqlink-docker/resqlink-worker:latest \
  --platform managed \
  --region $REGION \
  --no-allow-unauthenticated \
  --set-env-vars GCP_PROJECT_ID=$PROJECT_ID,GCP_LOCATION=$REGION \
  --memory 1Gi \
  --min-instances 0 \
  --max-instances 3 \
  --project=$PROJECT_ID

echo ""
echo "=== DEPLOYMENT COMPLETE ==="
API_URL=$(gcloud run services describe resqlink-api --region $REGION --project=$PROJECT_ID --format='value(status.url)')
echo "API URL: $API_URL"
echo ""
echo "Test with: curl $API_URL/health"

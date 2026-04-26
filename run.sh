#!/bin/bash
export GOOGLE_APPLICATION_CREDENTIALS="$(pwd)/credentials/service-account.json"
export GCP_PROJECT_ID="samaj-58742"
export GCP_LOCATION="asia-south1"

echo "Starting API Server..."
go run ./cmd/api &
API_PID=$!
echo "API Server PID: $API_PID"

echo "Starting AI Worker..."
go run ./cmd/worker &
WORKER_PID=$!
echo "AI Worker PID: $WORKER_PID"

echo "Servers are running in background. Logs are below:"
wait $API_PID $WORKER_PID

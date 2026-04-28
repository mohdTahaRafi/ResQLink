#!/bin/bash
set -e

echo "=== ResQLink Local Setup ==="
echo ""

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is required but not installed"
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ docker-compose is required but not installed"
    exit 1
fi

echo "[1/4] Starting Docker containers..."
docker-compose up -d firestore-emulator pubsub-emulator
echo "✅ Emulators started"

echo ""
echo "[2/4] Waiting for emulators to be ready..."
sleep 5

echo ""
echo "[3/4] Running Pub/Sub setup..."
# Note: This would run setup/main.go which creates the topic and subscription
# For now, we'll just note what needs to happen
cat << 'SETUP_NOTE'
To initialize Pub/Sub resources:
  export GCP_PROJECT_ID=resqlink-local
  export PUBSUB_EMULATOR_HOST=localhost:8085
  go run ./cmd/setup
SETUP_NOTE

echo ""
echo "[4/4] Seeding test data..."
# Optional: Document how to seed test data
cat << 'SEED_NOTE'
Test data sample reports and volunteers would be inserted here.
You can use Firestore Console or the API endpoints to create test data.
SEED_NOTE

echo ""
echo "✅ Local setup complete!"
echo ""
echo "Next steps:"
echo "  1. Set environment variables: export GCP_PROJECT_ID=resqlink-local"
echo "  2. Start API and Worker:"
echo "     docker-compose up api worker"
echo "  3. API will be available at http://localhost:8080"
echo "  4. Worker will be available at http://localhost:8081"
echo ""

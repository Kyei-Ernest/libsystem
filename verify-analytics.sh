#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🔍 Starting Analytics Verification..."

# 1. Register User
USER_ID=$((RANDOM % 10000))
EMAIL="analyticstest${USER_ID}@example.com"
PASSWORD="Password123!"

echo "1️⃣  Registering user $EMAIL..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8086/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"username\": \"analyticsuser${USER_ID}\",
    \"password\": \"$PASSWORD\",
    \"first_name\": \"Analytics\",
    \"last_name\": \"Tester\"
  }")

TOKEN=$(echo $REGISTER_RESPONSE | jq -r .data.token)
if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ Registration failed.${NC}"
    exit 1
fi

# 2. Create Collection
echo "2️⃣  Creating collection..."
COLLECTION_RESPONSE=$(curl -s -X POST http://localhost:8082/api/v1/collections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Analytics Collection",
    "description": "Collection for analytics test"
  }')
COLLECTION_ID=$(echo $COLLECTION_RESPONSE | jq -r .data.id)

# 3. Upload Document
echo "3️⃣  Uploading document..."
echo "Analytics content $(date +%s)" > analytics_doc.txt
UPLOAD_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/documents \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@analytics_doc.txt" \
  -F "title=Analytics Doc" \
  -F "description=Doc to track" \
  -F "collection_id=$COLLECTION_ID")
DOC_ID=$(echo $UPLOAD_RESPONSE | jq -r .data.id)
rm analytics_doc.txt

if [ "$DOC_ID" == "null" ]; then
    echo -e "${RED}❌ Upload failed.${NC}"
    exit 1
fi

# 4. Generate Events (View and Download)
echo "4️⃣  Generating events (View & Download)..."
curl -s -X POST "http://localhost:8081/api/v1/documents/${DOC_ID}/view" \
  -H "Authorization: Bearer $TOKEN" > /dev/null

curl -s -X POST "http://localhost:8081/api/v1/documents/${DOC_ID}/download" \
  -H "Authorization: Bearer $TOKEN" > /dev/null

echo "⏳ Waiting 15s for Kafa processing..."
sleep 15

# 5. Check Analytics Stats
echo "5️⃣  Checking Analytics Overview..."
# Note: Using localhost:8087 directly or via gateway if configured (gateway is 8085 usually, but nginx is 8080?)
# run-services.sh says: "Analytics Service: http://localhost:8087"
STATS_RESPONSE=$(curl -s -X GET http://localhost:8087/api/v1/analytics/overview)

VIEWS=$(echo $STATS_RESPONSE | jq -r .data.total_views)
DOWNLOADS=$(echo $STATS_RESPONSE | jq -r .data.total_downloads)

echo "Stats: Views=$VIEWS, Downloads=$DOWNLOADS"

if [ "$VIEWS" -ge 1 ] && [ "$DOWNLOADS" -ge 1 ]; then
    echo -e "${GREEN}✅ Analytics Service successfully tracked events!${NC}"
    exit 0
else
    echo -e "${RED}❌ Stats not updated. Expected at least 1 view and 1 download.${NC}"
    exit 1
fi

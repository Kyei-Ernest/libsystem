#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🔍 Starting Background Worker Verification..."

# 1. Register User
USER_ID=$((RANDOM % 10000))
EMAIL="workertest${USER_ID}@example.com"
PASSWORD="Password123!"

echo "1️⃣  Registering user $EMAIL..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8086/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"username\": \"workeruser${USER_ID}\",
    \"password\": \"$PASSWORD\",
    \"first_name\": \"Worker\",
    \"last_name\": \"Tester\"
  }")

TOKEN=$(echo $REGISTER_RESPONSE | jq -r .data.token)
if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ Registration failed. Response: $REGISTER_RESPONSE${NC}"
    exit 1
fi

# 2. Create Collection
echo "2️⃣  Creating collection..."
COLLECTION_RESPONSE=$(curl -s -X POST http://localhost:8082/api/v1/collections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Worker Test Collection",
    "description": "Collection for worker verification"
  }')

COLLECTION_ID=$(echo $COLLECTION_RESPONSE | jq -r .data.id)
if [ "$COLLECTION_ID" == "null" ] || [ -z "$COLLECTION_ID" ]; then
    echo -e "${RED}❌ Collection creation failed: $COLLECTION_RESPONSE${NC}"
    exit 1
fi

# 3. Upload Text File
echo "3️⃣  Uploading text file..."
UNIQUE_ID=$(date +%s)
SECRET_CONTENT="SecretWorkerContent${UNIQUE_ID}"
echo "This is a text file containing ${SECRET_CONTENT} that must be indexed." > worker_test.txt

UPLOAD_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/documents \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@worker_test.txt" \
  -F "title=Worker Test Doc" \
  -F "description=Testing text extraction" \
  -F "collection_id=$COLLECTION_ID")

DOC_ID=$(echo $UPLOAD_RESPONSE | jq -r .data.id)
if [ "$DOC_ID" == "null" ]; then
    echo -e "${RED}❌ Upload failed: $UPLOAD_RESPONSE${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Uploaded document $DOC_ID${NC}"

# 4. Wait for Processing
echo "4️⃣  Waiting for background processing (10 seconds)..."
sleep 10

# 5. Search for Content
echo "5️⃣  Searching for content '${SECRET_CONTENT}'..."
SEARCH_RESPONSE=$(curl -s -X GET "http://localhost:8084/api/v1/search?q=${SECRET_CONTENT}" \
  -H "Authorization: Bearer $TOKEN")

HITS=$(echo $SEARCH_RESPONSE | jq -r .data.total)

if [ "$HITS" != "null" ] && [ "$HITS" -gt 0 ]; then
    echo -e "${GREEN}✅ FOUND! The background worker successfully extracted and indexed the text content.${NC}"
    # Cleanup
    rm worker_test.txt
    exit 0
else
    echo -e "${RED}❌ NOT FOUND. Search response: $SEARCH_RESPONSE${NC}"
    echo "Check indexer-service logs for extraction errors."
    exit 1
fi

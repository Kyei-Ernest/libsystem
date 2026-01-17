#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "🔍 Starting Search Service Verification..."

# 1. Register User
echo "1️⃣  Registering new user..."
USER_ID=$((RANDOM % 10000))
EMAIL="searchtest${USER_ID}@example.com"
PASSWORD="Password123!"

REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8086/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"username\": \"searchuser${USER_ID}\",
    \"password\": \"$PASSWORD\",
    \"first_name\": \"Search\",
    \"last_name\": \"Tester\"
  }")

TOKEN=$(echo $REGISTER_RESPONSE | jq -r .data.token)

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ Registration failed. Response: $REGISTER_RESPONSE${NC}"
    exit 1
fi
echo -e "${GREEN}✅ User registered. Token obtained.${NC}"

# 1.5 Create Collection
echo "1️⃣.5️⃣  Creating collection..."
COLLECTION_RESPONSE=$(curl -s -X POST http://localhost:8082/api/v1/collections \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Search Test Collection",
    "description": "Collection for search verification"
  }')

COLLECTION_ID=$(echo $COLLECTION_RESPONSE | jq -r .data.id)

if [ "$COLLECTION_ID" == "null" ] || [ -z "$COLLECTION_ID" ]; then
    echo -e "${RED}❌ Collection creation failed. Response: $COLLECTION_RESPONSE${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Collection created. ID: $COLLECTION_ID${NC}"

# 2. Select/Create File
echo "2️⃣  Preparing test document..."
echo "This is a test document about ElasticSearch and Kafka integration. Unique ID: $(date +%s)" > search_test_doc.txt

# 3. Upload Document
echo "3️⃣  Uploading document..."
# COLLECTION_ID is already set above

UPLOAD_RESPONSE=$(curl -s -X POST http://localhost:8081/api/v1/documents \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@search_test_doc.txt" \
  -F "title=Integration Test Document" \
  -F "description=A document to test the full indexing pipeline" \
  -F "collection_id=$COLLECTION_ID")

DOC_ID=$(echo $UPLOAD_RESPONSE | jq -r .data.id)

if [[ "$UPLOAD_RESPONSE" == *"error"* ]] || [ "$DOC_ID" == "null" ]; then
    echo -e "${RED}❌ Upload failed. Response: $UPLOAD_RESPONSE${NC}"
    # If it failed, maybe collection ID issue. Check if we need to create a collection.
    # But let's see output first.
    exit 1
fi
echo -e "${GREEN}✅ Document uploaded. ID: $DOC_ID${NC}"

# 4. Wait for Indexing
echo "4️⃣  Waiting for indexing (5 seconds)..."
sleep 5

# 5. Search
echo "5️⃣  Searching for 'Integration'..."
SEARCH_RESPONSE=$(curl -s -X GET "http://localhost:8084/api/v1/search?q=Integration" \
  -H "Authorization: Bearer $TOKEN")

echo "Search Response: $SEARCH_RESPONSE"

if [[ "$SEARCH_RESPONSE" == *"Integration Test Document"* ]]; then
    echo -e "${GREEN}✅ FOUND! The document was successfully indexed and retrieved.${NC}"
else
    echo -e "${RED}❌ NOT FOUND. Indexing might still be in progress or failed.${NC}"
    echo "Check indexer-service logs."
    exit 1
fi

rm search_test_doc.txt
echo -e "${GREEN}🎉 Verification Successful!${NC}"

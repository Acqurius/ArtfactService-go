#!/bin/bash

# Test script for presigned URL generation and file download
# This script uploads a file, generates a presigned URL, and downloads the file using the token

set -e  # Exit on error

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_FILE="test_presigned.txt"
DOWNLOADED_FILE="downloaded_presigned.txt"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Presigned URL Test Script ===${NC}\n"

# Step 1: Create a test file
echo -e "${BLUE}Step 1: Creating test file...${NC}"
echo "This is a test file for presigned URL download - $(date)" > "$TEST_FILE"
echo -e "${GREEN}✓ Test file created: $TEST_FILE${NC}\n"

# Step 2: Upload the file
echo -e "${BLUE}Step 2: Uploading file...${NC}"
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/artifacts/innerop/upload" \
  -F "file=@$TEST_FILE")

echo "Upload response: $UPLOAD_RESPONSE"

# Extract UUID from response
UUID=$(echo "$UPLOAD_RESPONSE" | grep -o '"uuid":"[^"]*"' | cut -d'"' -f4)

if [ -z "$UUID" ]; then
  echo -e "${RED}✗ Failed to extract UUID from upload response${NC}"
  exit 1
fi

echo -e "${GREEN}✓ File uploaded successfully${NC}"
echo -e "  UUID: $UUID\n"

# Step 3: Generate presigned URL
echo -e "${BLUE}Step 3: Generating presigned URL...${NC}"

# Calculate expiration time (24 hours from now)
EXPIRATION=$(date -u -d "+24 hours" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+24H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null)

PRESIGNED_REQUEST=$(cat <<EOF
{
  "artifact_uuid": "$UUID",
  "valid_to": "$EXPIRATION",
  "max_downloads": 5
}
EOF
)

echo "Request payload:"
echo "$PRESIGNED_REQUEST"

PRESIGNED_RESPONSE=$(curl -s -X POST "$BASE_URL/genPresignedURL" \
  -H "Content-Type: application/json" \
  -d "$PRESIGNED_REQUEST")

echo -e "\nPresigned URL response: $PRESIGNED_RESPONSE"

# Extract token from response
TOKEN=$(echo "$PRESIGNED_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
PRESIGNED_URL=$(echo "$PRESIGNED_RESPONSE" | grep -o '"presigned_url":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}✗ Failed to extract token from presigned URL response${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Presigned URL generated successfully${NC}"
echo -e "  Token: $TOKEN"
echo -e "  URL: $PRESIGNED_URL\n"

# Step 4: Download file using presigned URL
echo -e "${BLUE}Step 4: Downloading file using presigned URL...${NC}"

HTTP_CODE=$(curl -s -w "%{http_code}" -o "$DOWNLOADED_FILE" \
  -X GET "$BASE_URL/artifacts/$TOKEN")

if [ "$HTTP_CODE" -eq 200 ]; then
  echo -e "${GREEN}✓ File downloaded successfully (HTTP $HTTP_CODE)${NC}"
  echo -e "  Downloaded to: $DOWNLOADED_FILE\n"
else
  echo -e "${RED}✗ Download failed with HTTP code: $HTTP_CODE${NC}"
  exit 1
fi

# Step 5: Verify file content
echo -e "${BLUE}Step 5: Verifying file content...${NC}"
echo "Original file content:"
cat "$TEST_FILE"
echo -e "\nDownloaded file content:"
cat "$DOWNLOADED_FILE"

if diff -q "$TEST_FILE" "$DOWNLOADED_FILE" > /dev/null; then
  echo -e "\n${GREEN}✓ File content matches! Test PASSED${NC}"
else
  echo -e "\n${RED}✗ File content mismatch! Test FAILED${NC}"
  exit 1
fi

# Step 6: Test download limit (optional)
echo -e "\n${BLUE}Step 6: Testing download limit...${NC}"
echo "Downloading 4 more times (max_downloads=5, already downloaded once)..."

for i in {2..5}; do
  HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null \
    -X GET "$BASE_URL/artifacts/$TOKEN")
  echo "  Download $i: HTTP $HTTP_CODE"
done

echo -e "\nAttempting 6th download (should fail)..."
HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null \
  -X GET "$BASE_URL/artifacts/$TOKEN")

if [ "$HTTP_CODE" -eq 403 ]; then
  echo -e "${GREEN}✓ Download limit enforced correctly (HTTP $HTTP_CODE)${NC}"
else
  echo -e "${RED}✗ Expected HTTP 403, got HTTP $HTTP_CODE${NC}"
fi

# Cleanup
echo -e "\n${BLUE}Cleaning up...${NC}"
rm -f "$TEST_FILE" "$DOWNLOADED_FILE"
echo -e "${GREEN}✓ Test files removed${NC}"

echo -e "\n${GREEN}=== All tests completed successfully! ===${NC}"

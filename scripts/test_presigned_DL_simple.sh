#!/bin/bash

# Simple presigned URL test script
# Usage: ./scripts/test_presigned_simple.sh [artifact_uuid]

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
ARTIFACT_UUID="$1"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

if [ -z "$ARTIFACT_UUID" ]; then
  echo -e "${RED}Error: Please provide an artifact UUID${NC}"
  echo "Usage: $0 <artifact_uuid>"
  echo ""
  echo "Example:"
  echo "  $0 550e8400-e29b-41d4-a716-446655440000"
  exit 1
fi

echo -e "${BLUE}Testing presigned URL for artifact: $ARTIFACT_UUID${NC}\n"

# Generate presigned URL
echo -e "${BLUE}1. Generating presigned URL...${NC}"
EXPIRATION=$(date -u -d "+24 hours" +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u -v+24H +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null)

RESPONSE=$(curl -s -X POST "$BASE_URL/genPresignedURL" \
  -H "Content-Type: application/json" \
  -d "{
    \"artifact_uuid\": \"$ARTIFACT_UUID\",
    \"valid_to\": \"$EXPIRATION\",
    \"max_downloads\": 3
  }")

echo "Response: $RESPONSE"

TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
PRESIGNED_URL=$(echo "$RESPONSE" | grep -o '"presigned_url":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}✗ Failed to generate presigned URL${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Presigned URL generated${NC}"
echo -e "  Token: $TOKEN"
echo -e "  URL: $PRESIGNED_URL\n"

# Download file
echo -e "${BLUE}2. Downloading file using presigned URL...${NC}"
OUTPUT_FILE="downloaded_$(date +%s).bin"

HTTP_CODE=$(curl -s -w "%{http_code}" -o "$OUTPUT_FILE" \
  -X GET "$BASE_URL/artifacts/$TOKEN")

if [ "$HTTP_CODE" -eq 200 ]; then
  FILE_SIZE=$(ls -lh "$OUTPUT_FILE" | awk '{print $5}')
  echo -e "${GREEN}✓ Download successful (HTTP $HTTP_CODE)${NC}"
  echo -e "  File: $OUTPUT_FILE"
  echo -e "  Size: $FILE_SIZE\n"
  echo -e "${GREEN}Test completed successfully!${NC}"
else
  echo -e "${RED}✗ Download failed (HTTP $HTTP_CODE)${NC}"
  rm -f "$OUTPUT_FILE"
  exit 1
fi

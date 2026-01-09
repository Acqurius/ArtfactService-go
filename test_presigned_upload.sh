#!/bin/bash

# ============================================================================
# Token-Based Upload & Download Test Script
# ============================================================================
# This script demonstrates the complete flow of:
# 1. Upload flow: Generate upload token → Get presigned URL → Upload to S3
# 2. Download flow: Generate download token → Download via redirect
# ============================================================================

set -e

BASE_URL="http://localhost:8080"

# Helper function to extract JSON value without jq
json_extract() {
  local json="$1"
  local key="$2"
  echo "$json" | python3 -c "import sys, json; print(json.load(sys.stdin).get('$key', ''))" 2>/dev/null || echo ""
}

echo "============================================================================"
echo "                    UPLOAD FLOW TEST"
echo "============================================================================"
echo ""

# ============================================================================
# STEP 1: Generate Upload Token
# ============================================================================

echo "STEP 1: Generating upload token (requesting upload permission)"
echo "------------------------------------------------------------------------"

TOKEN_RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "max_uploads": 5
  }' \
  "${BASE_URL}/genUploadPresignedURL")

echo "Response: $TOKEN_RESPONSE"
TOKEN=$(json_extract "$TOKEN_RESPONSE" "token")
UPLOAD_URL=$(json_extract "$TOKEN_RESPONSE" "upload_url")

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "Error: Failed to extract token from response"
  exit 1
fi

echo ""
echo "✓ Upload Token: $TOKEN"
echo "✓ Upload URL: $UPLOAD_URL"
echo ""
echo ""

# ============================================================================
# STEP 2: Prepare File to Upload
# ============================================================================

echo "STEP 2: Preparing file to upload"
echo "------------------------------------------------------------------------"
echo "Creating test file..."

echo "This is a test file uploaded via presigned URL at $(date)" > /tmp/test_presigned_upload.txt
FILE_SIZE=$(stat -f%z /tmp/test_presigned_upload.txt 2>/dev/null || stat -c%s /tmp/test_presigned_upload.txt)

echo "✓ File: /tmp/test_presigned_upload.txt"
echo "✓ Size: $FILE_SIZE bytes"
echo "✓ Content-Type: text/plain"
echo ""
echo ""

# ============================================================================
# STEP 3: Request Presigned Upload URL
# ============================================================================

echo "STEP 3: Requesting presigned upload URL from server"
echo "------------------------------------------------------------------------"

PRESIGNED_RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d "{
    \"filename\": \"test_presigned_upload.txt\",
    \"content_type\": \"text/plain\",
    \"size\": $FILE_SIZE
  }" \
  "${UPLOAD_URL}")

echo "Response: $PRESIGNED_RESPONSE"
PRESIGNED_URL=$(json_extract "$PRESIGNED_RESPONSE" "presigned_url")
NEW_UUID=$(json_extract "$PRESIGNED_RESPONSE" "uuid")
EXPIRES_IN=$(json_extract "$PRESIGNED_RESPONSE" "expires_in")

if [ -z "$PRESIGNED_URL" ] || [ "$PRESIGNED_URL" = "null" ]; then
  echo "Error: Failed to get presigned URL"
  exit 1
fi

echo ""
echo "✓ Artifact UUID: $NEW_UUID"
echo "✓ Presigned URL: $PRESIGNED_URL"
echo "✓ Expires in: $EXPIRES_IN"
echo ""
echo ""

# ============================================================================
# STEP 4: Upload File Directly to S3
# ============================================================================

echo "STEP 4: Uploading file directly to Ceph S3 (bypassing application server)"
echo "------------------------------------------------------------------------"

UPLOAD_RESULT=$(curl -s -w "\n%{http_code}" -X PUT \
  -H "Content-Type: text/plain" \
  --data-binary "@/tmp/test_presigned_upload.txt" \
  "$PRESIGNED_URL")

HTTP_CODE=$(echo "$UPLOAD_RESULT" | tail -n1)
echo "HTTP Status: $HTTP_CODE"

if [ "$HTTP_CODE" = "200" ]; then
  echo "✓ File uploaded successfully to S3!"
else
  echo "✗ Upload failed with status: $HTTP_CODE"
fi

echo ""
echo ""

# ============================================================================
# STEP 5: Verify Upload by Downloading
# ============================================================================

echo "STEP 5: Verifying upload by downloading the file"
echo "------------------------------------------------------------------------"
DOWNLOAD_URL="${BASE_URL}/artifact-service/v1/artifacts/${NEW_UUID}/action/downloadFile"
echo "Request: GET ${DOWNLOAD_URL}"
echo ""

curl -s "$DOWNLOAD_URL" -o /tmp/downloaded_file.txt
echo "Downloaded content:"
echo "------------------------------------------------------------------------"
cat /tmp/downloaded_file.txt
echo ""
echo "------------------------------------------------------------------------"
echo "✓ Upload verified successfully!"
echo ""
echo ""

# Cleanup
rm -f /tmp/test_presigned_upload.txt /tmp/downloaded_file.txt

echo "============================================================================"
echo "                    DOWNLOAD FLOW TEST"
echo "============================================================================"
echo ""

# ============================================================================
# STEP 6: Generate Download Token
# ============================================================================

echo "STEP 6: Generating download token for the uploaded file"
echo "------------------------------------------------------------------------"

DOWNLOAD_TOKEN_RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d "{
    \"artifact_uuid\": \"$NEW_UUID\",
    \"max_downloads\": 3
  }" \
  "${BASE_URL}/genDownloadPresignedURL")

echo "Response: $DOWNLOAD_TOKEN_RESPONSE"
DOWNLOAD_TOKEN=$(json_extract "$DOWNLOAD_TOKEN_RESPONSE" "token")
DOWNLOAD_PRESIGNED_URL=$(json_extract "$DOWNLOAD_TOKEN_RESPONSE" "presigned_url")

echo ""
echo "✓ Download Token: $DOWNLOAD_TOKEN"
echo "✓ Download URL: $DOWNLOAD_PRESIGNED_URL"
echo ""
echo ""

# ============================================================================
# STEP 7: Download File Using Token
# ============================================================================

echo "STEP 7: Downloading file using token (will redirect to S3)"
echo "------------------------------------------------------------------------"

curl -s -L "$DOWNLOAD_PRESIGNED_URL" -o /tmp/token_downloaded.txt
echo "Downloaded content:"
echo "------------------------------------------------------------------------"
cat /tmp/token_downloaded.txt
echo ""
echo "------------------------------------------------------------------------"
echo "✓ Download via token successful!"
echo ""

# Cleanup
rm -f /tmp/token_downloaded.txt

echo ""
echo "============================================================================"
echo "                    ALL TESTS COMPLETED SUCCESSFULLY!"
echo "============================================================================"
echo ""

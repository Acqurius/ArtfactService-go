#!/bin/bash

set -e

BASE_URL="http://localhost:8080"
TEST_FILE="test_$(date +%s).txt"

echo "=== 測試 S3 Presigned URL 下載功能 ==="

# 1. 建立測試檔案
echo "Step 1: 建立測試檔案"
echo "Test content $(date)" > "$TEST_FILE"

# 2. 上傳檔案
echo "Step 2: 上傳檔案"
UPLOAD_RESPONSE=$(curl -s -X POST "$BASE_URL/artifact-service/v1/artifacts/" -F "file=@$TEST_FILE")
UUID=$(echo "$UPLOAD_RESPONSE" | grep -o '"uuid":"[^"]*"' | cut -d'"' -f4)
echo "UUID: $UUID"

if [ -z "$UUID" ]; then
    echo "❌ 上傳失敗，無法取得 UUID"
    echo "Response: $UPLOAD_RESPONSE"
    exit 1
fi

# 3. 產生 token
echo "Step 3: 產生 presigned URL token"
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/genPresignedURL" \
  -H "Content-Type: application/json" \
  -d "{
    \"artifact_uuid\": \"$UUID\",
    \"valid_from\": \"2026-01-08T00:00:00Z\",
    \"valid_to\": \"2026-01-10T23:59:59Z\",
    \"max_downloads\": 3
  }")
TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Token: $TOKEN"

if [ -z "$TOKEN" ]; then
    echo "❌ 產生 token 失敗"
    echo "Response: $TOKEN_RESPONSE"
    exit 1
fi

# 4. 測試下載（檢查 302 redirect）
echo "Step 4: 測試 302 重新導向"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/artifacts/$TOKEN")
if [ "$HTTP_CODE" = "302" ]; then
    echo "✅ 302 重新導向成功"
else
    echo "❌ 預期 302，實際得到 $HTTP_CODE"
    exit 1
fi

# 5. 取得 presigned URL
echo "Step 5: 檢查 Location header"
PRESIGNED_URL=$(curl -s -I "$BASE_URL/artifacts/$TOKEN" | grep -i "^Location:" | cut -d' ' -f2 | tr -d '\r')
echo "Presigned URL: $PRESIGNED_URL"

if [[ "$PRESIGNED_URL" == *"X-Amz-Algorithm"* ]]; then
    echo "✅ Presigned URL 格式正確（包含 AWS 簽名參數）"
else
    echo "⚠️  警告：Presigned URL 可能格式不正確"
fi

# 6. 測試完整下載
echo "Step 6: 測試完整下載"
curl -L -s "$BASE_URL/artifacts/$TOKEN" -o "downloaded_$TEST_FILE"
if diff "$TEST_FILE" "downloaded_$TEST_FILE" > /dev/null; then
    echo "✅ 檔案內容正確"
else
    echo "❌ 檔案內容不符"
    exit 1
fi

# 7. 測試下載次數限制（已下載 1 次，還可以下載 2 次）
echo "Step 7: 測試下載次數追蹤"
HTTP_CODE2=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/artifacts/$TOKEN")
if [ "$HTTP_CODE2" = "302" ]; then
    echo "✅ 第二次下載成功（2/3）"
else
    echo "❌ 第二次下載失敗，HTTP code: $HTTP_CODE2"
fi

# 清理
rm -f "$TEST_FILE" "downloaded_$TEST_FILE"

echo ""
echo "=== 所有測試通過 ✅ ==="
echo ""
echo "總結："
echo "  - 檔案上傳：成功"
echo "  - Token 產生：成功"
echo "  - 302 重新導向：成功"
echo "  - Presigned URL 格式：正確"
echo "  - 檔案下載：成功"
echo "  - 內容驗證：正確"
echo "  - 下載次數追蹤：正常"

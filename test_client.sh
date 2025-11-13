#!/bin/bash

# Test script to verify dynamic header forwarding

echo "=== Testing Dynamic Header Forwarding ==="
echo ""

# Configuration
SERVER_URL="http://localhost:8000"
PROXY_URL="http://localhost:9090"
TEST_TOKEN="test-token-12345"

echo "Step 1: Testing SSE endpoint with custom headers"
echo "----------------------------------------------"

# Test the SSE message endpoint with custom headers
curl -X POST "${SERVER_URL}/message" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TEST_TOKEN}" \
  -H "X-Ctix-Base-Url: ${PROXY_URL}" \
  -H "X-Co-Base-Url: ${PROXY_URL}" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }' \
  -v

echo ""
echo ""
echo "Step 2: Check the proxy logs above to verify headers were forwarded"
echo "Expected headers in proxy logs:"
echo "  - Authorization: Bearer ${TEST_TOKEN}"
echo "  - X-Ctix-Base-Url: ${PROXY_URL}"
echo "  - X-Co-Base-Url: ${PROXY_URL}"
echo ""

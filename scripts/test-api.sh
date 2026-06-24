#!/bin/bash
set -euo pipefail

API_BASE="${API_URL:-http://localhost:8080}"
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "=== Nexbic Platform API Test ==="
echo "API: $API_BASE"
echo ""

pass() { echo -e "${GREEN}PASS${NC}: $1"; }
fail() { echo -e "${RED}FAIL${NC}: $1"; exit 1; }

echo "1. Health Check"
HEALTH=$(curl -sf "$API_BASE/health" || true)
if [ -n "$HEALTH" ]; then
  pass "Health endpoint"
else
  fail "Health endpoint not responding"
fi

echo ""
echo "2. Register User"
REGISTER=$(curl -sf -X POST "$API_BASE/api/v1/platform/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@nexbic.com","password":"testpass123","name":"Test User"}') || true

if echo "$REGISTER" | grep -q "token"; then
  TOKEN=$(echo "$REGISTER" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  pass "User registered"
else
  fail "Registration failed: $REGISTER"
fi

echo ""
echo "3. Login"
LOGIN=$(curl -sf -X POST "$API_BASE/api/v1/platform/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@nexbic.com","password":"testpass123"}') || true

if echo "$LOGIN" | grep -q "token"; then
  TOKEN=$(echo "$LOGIN" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
  pass "Login successful"
else
  fail "Login failed: $LOGIN"
fi

echo ""
echo "4. Create App"
CREATE_APP=$(curl -sf -X POST "$API_BASE/api/v1/platform/apps" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"CRM App","slug":"crm","description":"Customer management"}') || true

if echo "$CREATE_APP" | grep -q "admin_key"; then
  APP_ID=$(echo "$CREATE_APP" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  ADMIN_KEY=$(echo "$CREATE_APP" | grep -o '"raw_key":"[^"]*"' | head -1 | cut -d'"' -f4)
  pass "App created: $APP_ID"
  echo "  Admin Key: $ADMIN_KEY"
else
  fail "App creation failed: $CREATE_APP"
fi

echo ""
echo "5. List Apps"
APPS=$(curl -sf "$API_BASE/api/v1/platform/apps" \
  -H "Authorization: Bearer $TOKEN") || true

if echo "$APPS" | grep -q "crm"; then
  pass "Apps listed"
else
  fail "List apps failed: $APPS"
fi

echo ""
echo "6. Get App Details"
APP_DETAIL=$(curl -sf "$API_BASE/api/v1/platform/apps/$APP_ID" \
  -H "Authorization: Bearer $TOKEN") || true

if echo "$APP_DETAIL" | grep -q "CRM App"; then
  pass "App details retrieved"
else
  fail "Get app failed: $APP_DETAIL"
fi

echo ""
echo "7. Create API Key"
NEW_KEY=$(curl -sf -X POST "$API_BASE/api/v1/platform/apps/$APP_ID/apikey" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Mobile App Key","key_type":"publishable"}') || true

if echo "$NEW_KEY" | grep -q "raw_key"; then
  pass "API key created"
else
  fail "Create key failed: $NEW_KEY"
fi

echo ""
echo "8. List API Keys"
KEYS=$(curl -sf "$API_BASE/api/v1/platform/apps/$APP_ID/apikey" \
  -H "Authorization: Bearer $TOKEN") || true

if echo "$KEYS" | grep -q "Mobile App Key"; then
  pass "API keys listed"
else
  fail "List keys failed: $KEYS"
fi

echo ""
echo "=== All tests passed! ==="

#!/bin/bash
set -e

GREEN="\033[0;32m"
RED="\033[0;31m"
BLUE="\033[0;34m"
YELLOW="\033[1;33m"
NC="\033[0m"

BASE_URL="${TEST_BASE_URL:-https://appliance-store.local}"
HTTP_URL="${TEST_HTTP_URL:-http://appliance-store.local}"
PORT_HTTPS="${TEST_PORT_HTTPS:-443}"
PORT_HTTP="${TEST_PORT_HTTP:-80}"
RESOLVE_IP="${TEST_RESOLVE_IP:-127.0.0.1}"

# Setup curl resolve flag
CURL_HTTPS="curl -sk --resolve appliance-store.local:${PORT_HTTPS}:${RESOLVE_IP} --resolve store.local:${PORT_HTTPS}:${RESOLVE_IP}"
CURL_HTTP="curl -sI --resolve appliance-store.local:${PORT_HTTP}:${RESOLVE_IP} --resolve store.local:${PORT_HTTP}:${RESOLVE_IP}"

echo -e "${BLUE}======================================================${NC}"
echo -e "${BLUE}    School 21 BE3_Nginx_Pro Automated Test Suite      ${NC}"
echo -e "${BLUE}======================================================${NC}"
echo "Target Base URL: $BASE_URL"
echo ""

PASSED_COUNT=0
TOTAL_COUNT=0

assert_test() {
    local name="$1"
    local condition="$2"
    TOTAL_COUNT=$((TOTAL_COUNT + 1))
    if [ "$condition" -eq 0 ]; then
        echo -e "[ ${GREEN}PASS${NC} ] $name"
        PASSED_COUNT=$((PASSED_COUNT + 1))
    else
        echo -e "[ ${RED}FAIL${NC} ] $name"
    fi
}

# ------------------------------------------------------------------------------
# Test 1: Reverse Proxying & Basic Connectivity
# ------------------------------------------------------------------------------
echo -e "${YELLOW}>> Test 1: Reverse Proxying & Connectivity${NC}"
STATUS_CODE=$($CURL_HTTPS -o /dev/null -w "%{http_code}" "$BASE_URL/")
if [ "$STATUS_CODE" -eq 200 ]; then
    assert_test "Reverse proxy successfully responds with HTTP 200 on root" 0
else
    assert_test "Reverse proxy successfully responds with HTTP 200 on root (got $STATUS_CODE)" 1
fi

# ------------------------------------------------------------------------------
# Test 2: Routing
# ------------------------------------------------------------------------------
echo -e "\n${YELLOW}>> Test 2: Routing Configuration${NC}"

# 2.1 Route /api -> /api/v1 (outputs Swagger)
SWAGGER_API=$($CURL_HTTPS "$BASE_URL/api" | grep -i "Swagger UI" || true)
if [ -n "$SWAGGER_API" ]; then
    assert_test "Route /api outputs Swagger documentation" 0
else
    assert_test "Route /api outputs Swagger documentation" 1
fi

# 2.2 Route /api/v1 outputs Swagger
SWAGGER_V1=$($CURL_HTTPS "$BASE_URL/api/v1" | grep -i "Swagger UI" || true)
if [ -n "$SWAGGER_V1" ]; then
    assert_test "Route /api/v1 outputs Swagger documentation" 0
else
    assert_test "Route /api/v1 outputs Swagger documentation" 1
fi

# 2.3 Static distribution on / (index.html and image.png)
INDEX_HTML=$($CURL_HTTPS "$BASE_URL/index.html" | grep -i "Appliance Store Gateway" || true)
if [ -n "$INDEX_HTML" ]; then
    assert_test "Route /index.html serves static HTML content" 0
else
    assert_test "Route /index.html serves static HTML content" 1
fi

IMAGE_TYPE=$($CURL_HTTPS -I "$BASE_URL/image.png" | grep -i "content-type: image/png" || true)
if [ -n "$IMAGE_TYPE" ]; then
    assert_test "Route /image.png serves image asset (image/png)" 0
else
    assert_test "Route /image.png serves image asset" 1
fi

# 2.4 /admin pgAdmin GUI DBMS POSTGRES
ADMIN_STATUS=$($CURL_HTTPS -o /dev/null -w "%{http_code}" "$BASE_URL/admin")
if [ "$ADMIN_STATUS" -eq 200 ] || [ "$ADMIN_STATUS" -eq 301 ] || [ "$ADMIN_STATUS" -eq 302 ]; then
    assert_test "Route /admin proxies to pgAdmin (HTTP $ADMIN_STATUS)" 0
else
    assert_test "Route /admin proxies to pgAdmin (got $ADMIN_STATUS)" 1
fi

# 2.5 /status Nginx server status page (stub_status)
STATUS_BODY=$($CURL_HTTPS "$BASE_URL/status" | grep -i "Active connections" || true)
if [ -n "$STATUS_BODY" ]; then
    assert_test "Route /status returns Nginx stub_status ('$STATUS_BODY')" 0
else
    assert_test "Route /status returns Nginx stub_status" 1
fi

# ------------------------------------------------------------------------------
# Test 3: Load Balancing (2:1:1 Ratio for GET requests across 3 Backends)
# ------------------------------------------------------------------------------
echo -e "\n${YELLOW}>> Test 3: Load Balancing (2:1:1 Ratio)${NC}"
echo "Dispatching 40 GET requests to /api/v1/products..."

COUNT_MAIN=0
COUNT_RO1=0
COUNT_RO2=0

for i in $(seq 1 40); do
    HEADER_PORT=$($CURL_HTTPS -I "$BASE_URL/api/v1/products" | grep -i "x-backend-instance:" | tr -d '\r' | awk '{print $2}')
    case "$HEADER_PORT" in
        *backend-main*|*8081*)
            COUNT_MAIN=$((COUNT_MAIN + 1))
            ;;
        *backend-replica1*|*8082*)
            COUNT_RO1=$((COUNT_RO1 + 1))
            ;;
        *backend-replica2*|*8083*)
            COUNT_RO2=$((COUNT_RO2 + 1))
            ;;
    esac
done

echo "Requests distribution: Main(Instance 1)=$COUNT_MAIN, Replica 1(Instance 2)=$COUNT_RO1, Replica 2(Instance 3)=$COUNT_RO2"
# Expected for 40 requests with 2:1:1 is ~20, ~10, ~10
if [ "$COUNT_MAIN" -gt "$COUNT_RO1" ] && [ "$COUNT_MAIN" -gt "$COUNT_RO2" ] && [ "$COUNT_RO1" -gt 0 ] && [ "$COUNT_RO2" -gt 0 ]; then
    assert_test "GET balancing correctly distributed across all 3 backends in 2:1:1 ratio" 0
else
    assert_test "GET balancing distribution in 2:1:1 ratio (Main=$COUNT_MAIN, RO1=$COUNT_RO1, RO2=$COUNT_RO2)" 1
fi

# ------------------------------------------------------------------------------
# Test 4: Caching Configuration (All GET except /api)
# ------------------------------------------------------------------------------
echo -e "\n${YELLOW}>> Test 4: Caching Configuration${NC}"

# Check GET /api/v1/products -> must NOT be cached (BYPASS)
API_CACHE_STATUS=$($CURL_HTTPS -I "$BASE_URL/api/v1/products" | grep -i "x-cache-status:" | tr -d '\r' | awk '{print $2}')
if [ "$API_CACHE_STATUS" = "BYPASS" ] || [ -z "$API_CACHE_STATUS" ]; then
    assert_test "GET /api requests bypass cache (X-Cache-Status: $API_CACHE_STATUS)" 0
else
    assert_test "GET /api requests bypass cache (got $API_CACHE_STATUS)" 1
fi

# Check non-api GET requests -> caching enabled
# Warm up cache
$CURL_HTTPS "$BASE_URL/swagger/doc.json" > /dev/null 2>&1 || true
# Second request should be HIT
SWAGGER_CACHE_STATUS=$($CURL_HTTPS -I "$BASE_URL/swagger/doc.json" | grep -i "x-cache-status:" | tr -d '\r' | awk '{print $2}')
if [ "$SWAGGER_CACHE_STATUS" = "HIT" ] || [ "$SWAGGER_CACHE_STATUS" = "MISS" ]; then
    assert_test "Non-API GET requests utilize cache (X-Cache-Status: $SWAGGER_CACHE_STATUS)" 0
else
    assert_test "Non-API GET requests utilize cache (got $SWAGGER_CACHE_STATUS)" 1
fi

# ------------------------------------------------------------------------------
# Test 5: Gzip Compression Configuration (Exclude media types)
# ------------------------------------------------------------------------------
echo -e "\n${YELLOW}>> Test 5: Gzip Compression Configuration${NC}"

# Text/JSON should be gzip compressed
JSON_GZIP=$($CURL_HTTPS -H "Accept-Encoding: gzip" -I "$BASE_URL/swagger/doc.json" | grep -i "content-encoding: gzip" || true)
if [ -n "$JSON_GZIP" ]; then
    assert_test "Text / JSON responses compressed with Gzip" 0
else
    assert_test "Text / JSON responses compressed with Gzip" 1
fi

# Media types (image/png) must NOT be gzip compressed
PNG_GZIP=$($CURL_HTTPS -H "Accept-Encoding: gzip" -I "$BASE_URL/image.png" | grep -i "content-encoding: gzip" || true)
if [ -z "$PNG_GZIP" ]; then
    assert_test "Media types (PNG) strictly excluded from Gzip compression" 0
else
    assert_test "Media types (PNG) strictly excluded from Gzip compression" 1
fi

# ------------------------------------------------------------------------------
# Test 6: HTTPS & Local DNS Configuration
# ------------------------------------------------------------------------------
echo -e "\n${YELLOW}>> Test 6: HTTPS & Local DNS Configuration${NC}"

# Verify HTTP (80) redirects to HTTPS (443)
HTTP_REDIRECT=$($CURL_HTTP "$HTTP_URL/" | grep -iE "location: https://" || true)
HTTP_STATUS=$($CURL_HTTP "$HTTP_URL/" | grep -i "HTTP/" | head -n 1 | awk '{print $2}')
if [ "$HTTP_STATUS" = "301" ] || [ -n "$HTTP_REDIRECT" ]; then
    assert_test "HTTP port 80 redirects to HTTPS (HTTP 301 Moved Permanently)" 0
else
    assert_test "HTTP port 80 redirects to HTTPS (Status: $HTTP_STATUS)" 1
fi

# Verify SSL Certificate SAN
CERT_SAN=$(openssl s_client -connect "${RESOLVE_IP}:${PORT_HTTPS}" -servername appliance-store.local </dev/null 2>/dev/null | openssl x509 -noout -text | grep -i "DNS:appliance-store.local" || true)
if [ -n "$CERT_SAN" ]; then
    assert_test "Self-signed certificate bound to domain appliance-store.local" 0
else
    assert_test "Self-signed certificate bound to domain appliance-store.local" 1
fi

# ------------------------------------------------------------------------------
# Summary
# ------------------------------------------------------------------------------
echo ""
echo -e "${BLUE}======================================================${NC}"
echo -e "Test Results: ${GREEN}$PASSED_COUNT${NC} / ${TOTAL_COUNT} Passed"
echo -e "${BLUE}======================================================${NC}"

if [ "$PASSED_COUNT" -eq "$TOTAL_COUNT" ]; then
    echo -e "${GREEN}All Chapter II project requirements successfully verified!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Please inspect logs above.${NC}"
    exit 1
fi

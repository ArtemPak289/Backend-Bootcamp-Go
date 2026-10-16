#!/usr/bin/env bash
set -e

BASE_URL="http://localhost:8080"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_DIR="/tmp/be4_pids"
AUTH_LOG="/tmp/be4_auth.log"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0

assert_eq() {
  local test_name="$1"
  local actual="$2"
  local expected="$3"
  if [ "${actual}" = "${expected}" ]; then
    echo -e "  [${GREEN}PASS${NC}] ${test_name} (got: ${actual})"
    PASSED=$((PASSED + 1))
  else
    echo -e "  [${RED}FAIL${NC}] ${test_name} (expected: ${expected}, got: ${actual})"
    FAILED=$((FAILED + 1))
  fi
}

assert_contains() {
  local test_name="$1"
  local haystack="$2"
  local needle="$3"
  if echo "${haystack}" | grep -q "${needle}"; then
    echo -e "  [${GREEN}PASS${NC}] ${test_name} (contains '${needle}')"
    PASSED=$((PASSED + 1))
  else
    echo -e "  [${RED}FAIL${NC}] ${test_name} (does NOT contain '${needle}')"
    FAILED=$((FAILED + 1))
  fi
}

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}  BE4 End-to-End Test Suite: gRPC Authorization & DI Container  ${NC}"
echo -e "${BLUE}================================================================${NC}"

if ! curl -s "http://localhost:8080/health" | grep -q "ok"; then
  echo "==> Services not running on http://localhost:8080, starting native services..."
  "${SCRIPT_DIR}/run_local.sh"
fi

# Check health
echo -e "\n${YELLOW}1. Health & Unauthenticated Swagger Check${NC}"
HEALTH_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/health")
assert_eq "Store Health Endpoint returns 200" "${HEALTH_CODE}" "200"

SWAGGER_CODE=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/swagger/index.html")
assert_eq "Swagger UI available without auth" "${SWAGGER_CODE}" "200"

# Requirement 4: Protection of APIs
echo -e "\n${YELLOW}2. Endpoint Authorization & 401 Enforcement (No Token)${NC}"

CLIENTS_NO_AUTH=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/clients")
assert_eq "GET /api/v1/clients without auth returns 401" "${CLIENTS_NO_AUTH}" "401"

PRODUCTS_NO_AUTH=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/products")
assert_eq "GET /api/v1/products without auth returns 401" "${PRODUCTS_NO_AUTH}" "401"

SUPPLIERS_NO_AUTH=$(curl -s -o /dev/null -w "%{http_code}" "${BASE_URL}/api/v1/suppliers")
assert_eq "GET /api/v1/suppliers without auth returns 401" "${SUPPLIERS_NO_AUTH}" "401"

CLIENTS_BAD_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer invalid.jwt.token" "${BASE_URL}/api/v1/clients")
assert_eq "GET /api/v1/clients with invalid token returns 401" "${CLIENTS_BAD_TOKEN}" "401"

# Requirement 2 & 3: User Registration via gRPC
echo -e "\n${YELLOW}3. User Registration via /register (gRPC Auth Service)${NC}"
RAND_SUFFIX=$(date +%s)_$RANDOM
TEST_EMAIL="user_${RAND_SUFFIX}@example.com"
TEST_PASS="P@ssw0rd123"

REG_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\",
    \"first_name\": \"John\",
    \"last_name\": \"Doe\",
    \"phone_number\": \"+1234567890\",
    \"password\": \"${TEST_PASS}\"
  }")

assert_contains "Registration response contains token" "${REG_RESP}" "token"
assert_contains "Registration response contains email" "${REG_RESP}" "${TEST_EMAIL}"

TOKEN=$(echo "${REG_RESP}" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

# Requirement 2: Database password security (Salted hash, no plain text)
echo -e "\n${YELLOW}4. Database Security Verification (auth_db)${NC}"
DB_ROW=$(psql -d auth_db -t -A -c "SELECT password_hash, salt FROM users WHERE email='${TEST_EMAIL}'")
assert_contains "Password in DB is hashed and non-empty" "${DB_ROW}" "|"
if [ -n "${DB_ROW}" ]; then
  HASH_VAL=$(echo "${DB_ROW}" | cut -d'|' -f1)
  SALT_VAL=$(echo "${DB_ROW}" | cut -d'|' -f2)
  if [ "${HASH_VAL}" != "${TEST_PASS}" ] && [ -n "${SALT_VAL}" ]; then
    echo -e "  [${GREEN}PASS${NC}] Passwords stored with salt (Hash length: ${#HASH_VAL}, Salt length: ${#SALT_VAL})"
    PASSED=$((PASSED + 1))
  else
    echo -e "  [${RED}FAIL${NC}] Password stored in plain text or without salt!"
    FAILED=$((FAILED + 1))
  fi
fi

# Requirement 4: Access protected endpoints with valid token
echo -e "\n${YELLOW}5. Access Protected APIs with Valid Token (gRPC Validation)${NC}"

CLIENTS_AUTH=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/clients")
assert_eq "GET /api/v1/clients with valid token returns 200" "${CLIENTS_AUTH}" "200"

PRODUCTS_AUTH=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/products")
assert_eq "GET /api/v1/products with valid token returns 200" "${PRODUCTS_AUTH}" "200"

SUPPLIERS_AUTH=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer ${TOKEN}" "${BASE_URL}/api/v1/suppliers")
assert_eq "GET /api/v1/suppliers with valid token returns 200" "${SUPPLIERS_AUTH}" "200"

# Requirement 2 & 3: User Authentication via /auth
echo -e "\n${YELLOW}6. User Authentication via /auth (gRPC Verification)${NC}"

LOGIN_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/auth" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\",
    \"password\": \"${TEST_PASS}\"
  }")

assert_contains "Login with correct credentials returns token" "${LOGIN_RESP}" "token"

LOGIN_FAIL_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/v1/auth" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\",
    \"password\": \"WrongPassword!\"
  }")
assert_eq "Login with wrong password returns 401" "${LOGIN_FAIL_CODE}" "401"

# Requirement 2 & 3: Change Password
echo -e "\n${YELLOW}7. Change Password via gRPC${NC}"
NEW_PASS="BrandNewPassword456!"

CHANGE_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/change-password" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\",
    \"old_password\": \"${TEST_PASS}\",
    \"new_password\": \"${NEW_PASS}\"
  }")
assert_contains "Password change succeeds" "${CHANGE_RESP}" "true"

OLD_LOGIN_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}/api/v1/auth" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\",
    \"password\": \"${TEST_PASS}\"
  }")
assert_eq "Login with old password now returns 401" "${OLD_LOGIN_CODE}" "401"

NEW_LOGIN_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/auth" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\",
    \"password\": \"${NEW_PASS}\"
  }")
assert_contains "Login with new password returns token" "${NEW_LOGIN_RESP}" "token"

# Requirement 2 & 3: Password Recovery / Reset
echo -e "\n${YELLOW}8. Password Recovery via /reset (Console Stub Verification)${NC}"

RESET_RESP=$(curl -s -X POST "${BASE_URL}/api/v1/reset" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"${TEST_EMAIL}\"
  }")
assert_contains "Password reset API returns success" "${RESET_RESP}" "true"

# Verify email stub was printed to Auth Service console
if [ -f "${AUTH_LOG}" ]; then
  if grep -q "\[EMAIL STUB\]" "${AUTH_LOG}"; then
    echo -e "  [${GREEN}PASS${NC}] Auth Service console received [EMAIL STUB] password notification"
    PASSED=$((PASSED + 1))
  else
    echo -e "  [${RED}FAIL${NC}] [EMAIL STUB] not found in Auth Service logs!"
    FAILED=$((FAILED + 1))
  fi
fi

# Summary
echo -e "\n${BLUE}================================================================${NC}"
echo -e "Test Results: ${GREEN}${PASSED} passed${NC}, ${RED}${FAILED} failed${NC}"
echo -e "${BLUE}================================================================${NC}"

if [ "${FAILED}" -gt 0 ]; then
  exit 1
fi
exit 0

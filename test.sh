#!/bin/bash

# Config Manager End-to-End Test Script
# Tests all major functionality of the config manager service

set -e  # Exit on error
set +e  # Disable exit on error temporarily for API calls that might fail

BASE_URL="http://localhost:8089"
CONTENT_TYPE="Content-Type: application/json"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to print test results
print_test() {
    local test_name=$1
    local status=$2
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓${NC} $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        ((TESTS_FAILED++))
    fi
}

# Helper function to make API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -z "$data" ]; then
        curl -s -X "$method" "$BASE_URL$endpoint" -H "$CONTENT_TYPE"
    else
        curl -s -X "$method" "$BASE_URL$endpoint" -H "$CONTENT_TYPE" -d "$data"
    fi
}

# Helper to extract JSON field
extract_json() {
    echo "$1" | grep -o "\"$2\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | sed 's/.*"\([^"]*\)"/\1/'
}

# Helper to extract JSON numeric field
extract_json_number() {
    echo "$1" | grep -o "\"$2\"[[:space:]]*:[[:space:]]*[0-9]*" | grep -o "[0-9]*$"
}

echo "=========================================="
echo "Config Manager E2E Test Suite"
echo "=========================================="
echo ""

# Wait for service to be ready
echo "⏳ Waiting for service to be ready..."
for i in {1..30}; do
    if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Service is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}✗${NC} Service failed to start"
        exit 1
    fi
    sleep 1
done
echo ""

# Test 1: Health Check
echo "=== Test 1: Health Check ==="
RESPONSE=$(curl -s "$BASE_URL/health")
if echo "$RESPONSE" | grep -q "ok"; then
    print_test "Health check endpoint" "PASS"
else
    print_test "Health check endpoint" "FAIL"
fi
echo ""

# Test 2: Create Service
echo "=== Test 2: Service Management ==="
SERVICE_RESPONSE=$(api_call POST "/api/v1/services" '{"name":"test-service"}')
SERVICE_ID=$(extract_json "$SERVICE_RESPONSE" "id")

if [ -n "$SERVICE_ID" ]; then
    print_test "Create service" "PASS"
    echo "  Service ID: $SERVICE_ID"
else
    print_test "Create service" "FAIL"
    echo "  Response: $SERVICE_RESPONSE"
fi

# Test 3: List Services
SERVICES_LIST=$(api_call GET "/api/v1/services")
if echo "$SERVICES_LIST" | grep -q "test-service"; then
    print_test "List services" "PASS"
else
    print_test "List services" "FAIL"
fi

# Test 4: Get Service by ID
SERVICE_DETAIL=$(api_call GET "/api/v1/services/$SERVICE_ID")
if echo "$SERVICE_DETAIL" | grep -q "$SERVICE_ID"; then
    print_test "Get service by ID" "PASS"
else
    print_test "Get service by ID" "FAIL"
fi
echo ""

# Test 5: Create Environment
echo "=== Test 3: Environment Management ==="
ENV_RESPONSE=$(api_call POST "/api/v1/services/$SERVICE_ID/environments" '{"name":"production"}')
ENV_ID=$(extract_json "$ENV_RESPONSE" "id")

if [ -n "$ENV_ID" ]; then
    print_test "Create environment" "PASS"
    echo "  Environment ID: $ENV_ID"
else
    print_test "Create environment" "FAIL"
    echo "  Response: $ENV_RESPONSE"
fi

# Test 6: List Environments
ENV_LIST=$(api_call GET "/api/v1/services/$SERVICE_ID/environments")
if echo "$ENV_LIST" | grep -q "production"; then
    print_test "List environments" "PASS"
else
    print_test "List environments" "FAIL"
fi
echo ""

# Test 7: Config Operations - String
echo "=== Test 4: Config Operations - String ==="
CONFIG_RESPONSE=$(api_call POST "/api/v1/configs/$ENV_ID/app.name" '{"value":"MyApp","value_type":"string","performed_by":"test-script"}')
if echo "$CONFIG_RESPONSE" | grep -q "MyApp"; then
    print_test "Create string config" "PASS"
else
    print_test "Create string config" "FAIL"
fi

# Test 8: Get Config
GET_CONFIG=$(api_call GET "/api/v1/configs/$ENV_ID/app.name")
if echo "$GET_CONFIG" | grep -q "MyApp"; then
    print_test "Get string config" "PASS"
else
    print_test "Get string config" "FAIL"
fi
echo ""

# Test 9: Config Operations - Integer
echo "=== Test 5: Config Operations - Integer ==="
INT_CONFIG=$(api_call POST "/api/v1/configs/$ENV_ID/database.max_connections" '{"value":100,"value_type":"int","performed_by":"test-script"}')
if echo "$INT_CONFIG" | grep -q "100"; then
    print_test "Create integer config" "PASS"
else
    print_test "Create integer config" "FAIL"
fi

GET_INT=$(api_call GET "/api/v1/configs/$ENV_ID/database.max_connections")
if echo "$GET_INT" | grep -q "100"; then
    print_test "Get integer config" "PASS"
else
    print_test "Get integer config" "FAIL"
fi
echo ""

# Test 10: Config Operations - Boolean
echo "=== Test 6: Config Operations - Boolean ==="
BOOL_CONFIG=$(api_call POST "/api/v1/configs/$ENV_ID/app.debug_mode" '{"value":true,"value_type":"bool","performed_by":"test-script"}')
if echo "$BOOL_CONFIG" | grep -q "true"; then
    print_test "Create boolean config" "PASS"
else
    print_test "Create boolean config" "FAIL"
fi
echo ""

# Test 11: Config Operations - JSON
echo "=== Test 7: Config Operations - JSON ==="
JSON_CONFIG=$(api_call POST "/api/v1/configs/$ENV_ID/database.config" '{"value":{"host":"localhost","port":5432},"value_type":"json","performed_by":"test-script"}')
if echo "$JSON_CONFIG" | grep -q "localhost"; then
    print_test "Create JSON config" "PASS"
else
    print_test "Create JSON config" "FAIL"
fi
echo ""

# Test 12: List All Configs
echo "=== Test 8: List All Configs ==="
ALL_CONFIGS=$(api_call GET "/api/v1/configs/$ENV_ID")
if echo "$ALL_CONFIGS" | grep -q "app.name" && echo "$ALL_CONFIGS" | grep -q "database.max_connections"; then
    print_test "List all configs" "PASS"
    # Count configs
    CONFIG_COUNT=$(echo "$ALL_CONFIGS" | grep -o '"key"' | wc -l)
    echo "  Total configs: $CONFIG_COUNT"
else
    print_test "List all configs" "FAIL"
fi
echo ""

# Test 13: Update Config (Versioning)
echo "=== Test 9: Config Versioning ==="
UPDATE_CONFIG=$(api_call POST "/api/v1/configs/$ENV_ID/database.max_connections" '{"value":150,"performed_by":"test-script"}')
if echo "$UPDATE_CONFIG" | grep -q "150"; then
    print_test "Update config (create version 2)" "PASS"
else
    print_test "Update config (create version 2)" "FAIL"
fi

# Test 14: Get Version History
VERSION_HISTORY=$(api_call GET "/api/v1/configs/$ENV_ID/database.max_connections/versions")
if echo "$VERSION_HISTORY" | grep -q "version"; then
    print_test "Get version history" "PASS"
    VERSION_COUNT=$(echo "$VERSION_HISTORY" | grep -o '"version"' | wc -l)
    echo "  Total versions: $VERSION_COUNT"
else
    print_test "Get version history" "FAIL"
fi
echo ""

# Test 15: Rollback Config
echo "=== Test 10: Config Rollback ==="
ROLLBACK=$(api_call POST "/api/v1/configs/$ENV_ID/database.max_connections/rollback" '{"version":1,"performed_by":"test-script"}')
if echo "$ROLLBACK" | grep -q "100"; then
    print_test "Rollback to version 1" "PASS"
else
    print_test "Rollback to version 1" "FAIL"
fi

# Verify rollback
VERIFY_ROLLBACK=$(api_call GET "/api/v1/configs/$ENV_ID/database.max_connections")
if echo "$VERIFY_ROLLBACK" | grep -q "100"; then
    print_test "Verify rollback value" "PASS"
else
    print_test "Verify rollback value" "FAIL"
fi
echo ""

# Test 16: Audit Logs
echo "=== Test 11: Audit Logs ==="
AUDIT_LOGS=$(api_call GET "/api/v1/services/$SERVICE_ID/audit?limit=10&offset=0")
if echo "$AUDIT_LOGS" | grep -q "audit_logs"; then
    print_test "Get audit logs" "PASS"
    # Count audit entries
    AUDIT_COUNT=$(echo "$AUDIT_LOGS" | grep -o '"action"' | wc -l)
    echo "  Total audit entries: $AUDIT_COUNT"
else
    print_test "Get audit logs" "FAIL"
fi

# Check for specific actions
if echo "$AUDIT_LOGS" | grep -q "create" && echo "$AUDIT_LOGS" | grep -q "update" && echo "$AUDIT_LOGS" | grep -q "rollback"; then
    print_test "Verify audit actions (create, update, rollback)" "PASS"
else
    print_test "Verify audit actions" "FAIL"
fi
echo ""

# Test 17: Update Service
echo "=== Test 12: Update Operations ==="
UPDATE_SERVICE=$(api_call PUT "/api/v1/services/$SERVICE_ID" '{"name":"test-service-updated"}')
if echo "$UPDATE_SERVICE" | grep -q "test-service-updated"; then
    print_test "Update service name" "PASS"
else
    print_test "Update service name" "FAIL"
fi

# Test 18: Update Environment
UPDATE_ENV=$(api_call PUT "/api/v1/environments/$ENV_ID" '{"name":"production-updated"}')
if echo "$UPDATE_ENV" | grep -q "production-updated"; then
    print_test "Update environment name" "PASS"
else
    print_test "Update environment name" "FAIL"
fi
echo ""

# Test 19: Delete Config
echo "=== Test 13: Delete Operations ==="
DELETE_CONFIG=$(api_call DELETE "/api/v1/configs/$ENV_ID/app.debug_mode")
if [ "$(echo "$DELETE_CONFIG" | wc -c)" -le 2 ]; then
    print_test "Delete config" "PASS"
else
    print_test "Delete config" "FAIL"
fi

# Verify deletion
VERIFY_DELETE=$(api_call GET "/api/v1/configs/$ENV_ID")
if ! echo "$VERIFY_DELETE" | grep -q "app.debug_mode"; then
    print_test "Verify config deletion" "PASS"
else
    print_test "Verify config deletion" "FAIL"
fi
echo ""

# Test 20: Error Handling
echo "=== Test 14: Error Handling ==="

# Get non-existent config
ERROR_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/api/v1/configs/$ENV_ID/non.existent.key")
if [ "$ERROR_RESPONSE" = "404" ]; then
    print_test "404 on non-existent config" "PASS"
else
    print_test "404 on non-existent config" "FAIL"
fi

# Invalid service ID
ERROR_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "$BASE_URL/api/v1/services/invalid-uuid")
if [ "$ERROR_RESPONSE" = "404" ] || [ "$ERROR_RESPONSE" = "500" ]; then
    print_test "Error on invalid service ID" "PASS"
else
    print_test "Error on invalid service ID" "FAIL"
fi
echo ""

# Cleanup (optional)
echo "=== Test 15: Cleanup ==="
DELETE_ENV=$(api_call DELETE "/api/v1/environments/$ENV_ID")
if [ "$(echo "$DELETE_ENV" | wc -c)" -le 2 ]; then
    print_test "Delete environment" "PASS"
else
    print_test "Delete environment" "FAIL"
fi

DELETE_SERVICE=$(api_call DELETE "/api/v1/services/$SERVICE_ID")
if [ "$(echo "$DELETE_SERVICE" | wc -c)" -le 2 ]; then
    print_test "Delete service" "PASS"
else
    print_test "Delete service" "FAIL"
fi
echo ""

# Summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Tests Passed:  ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed:  ${RED}$TESTS_FAILED${NC}"
echo "Total Tests:   $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi

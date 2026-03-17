# Config Manager Test Guide

## End-to-End Test Script

The `test.sh` script provides comprehensive end-to-end testing for the Config Manager service. It tests all major functionality including CRUD operations, versioning, rollback, and audit logging.

### Prerequisites

- Docker and Docker Compose installed
- Service running on port 8089 (via `docker compose up -d`)
- `curl` command available

### Running the Tests

```bash
# Start the services
docker compose up -d

# Wait for services to be healthy (about 5-10 seconds)
docker compose ps

# Run the test script
./test.sh
```

### Test Coverage

The test script covers:

#### 1. **Health Check** (1 test)
- Service availability

#### 2. **Service Management** (3 tests)
- Create service
- List services
- Get service by ID

#### 3. **Environment Management** (2 tests)
- Create environment
- List environments for a service

#### 4. **Config Operations - String** (2 tests)
- Create string config
- Get string config

#### 5. **Config Operations - Integer** (2 tests)
- Create integer config
- Get integer config with type conversion

#### 6. **Config Operations - Boolean** (1 test)
- Create boolean config

#### 7. **Config Operations - JSON** (1 test)
- Create JSON object config

#### 8. **List All Configs** (1 test)
- List all configs in an environment

#### 9. **Config Versioning** (2 tests)
- Update config (creates new version)
- Get version history

#### 10. **Config Rollback** (2 tests)
- Rollback to previous version
- Verify rollback worked correctly

#### 11. **Audit Logs** (2 tests)
- Get audit logs for a service
- Verify audit actions (create, update, rollback)

#### 12. **Update Operations** (2 tests)
- Update service name
- Update environment name

#### 13. **Delete Operations** (2 tests)
- Delete config
- Verify config deletion

#### 14. **Error Handling** (2 tests)
- 404 on non-existent config
- Error on invalid service ID

#### 15. **Cleanup** (2 tests)
- Delete environment
- Delete service

**Total: 27 tests**

### Expected Output

When all tests pass, you should see:

```
==========================================
Test Summary
==========================================
Tests Passed:  27
Tests Failed:  0
Total Tests:   27

✓ All tests passed!
```

### Test Features

- **Colored output**: Green checkmarks for passed tests, red X for failed tests
- **Detailed logging**: Shows IDs and counts for created resources
- **Automatic cleanup**: Removes test data after completion
- **Service health check**: Waits up to 30 seconds for service to be ready
- **Error handling**: Tests both success and error scenarios

### Debugging Failed Tests

If tests fail, check:

1. **Service logs**:
   ```bash
   docker compose logs config-service
   ```

2. **Database connectivity**:
   ```bash
   docker compose ps
   # Both config-service and postgres should be "Up" and "healthy"
   ```

3. **Manual API test**:
   ```bash
   curl http://localhost:8089/health
   # Should return: {"status":"ok"}
   ```

4. **Run test in debug mode**:
   ```bash
   bash -x ./test.sh 2>&1 | less
   ```

### SDK Integration Test

There's also an SDK integration test in `test/sdk_test.go` that demonstrates:

- Type-safe getters (GetString, GetInt, GetBool, GetFloat, GetJSON)
- Default values
- Setting configs
- Listing all configs
- Background polling
- Error handling

To run the SDK test:

```bash
# First, create a service and environment via API
# Update the envID in test/sdk_test.go
# Then run:
go run test/sdk_test.go
```

### Continuous Integration

This test script can be integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions
- name: Start services
  run: docker compose up -d
  
- name: Wait for services
  run: sleep 10
  
- name: Run tests
  run: ./test.sh
  
- name: Cleanup
  run: docker compose down -v
  if: always()
```

### Troubleshooting

**Issue**: "Service failed to start"
- **Solution**: Check if port 8089 is available: `lsof -i :8089`

**Issue**: "Connection refused"
- **Solution**: Verify Docker containers are running: `docker compose ps`

**Issue**: Tests timeout
- **Solution**: Increase wait time in test script or check service performance

**Issue**: Database errors
- **Solution**: Reset database: `docker compose down -v && docker compose up -d`

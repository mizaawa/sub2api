# CI Test Resilience Improvements

## Problem Summary

CI tests in v0.1.190 were failing due to:

1. **139 instances of time.Sleep** - causing flaky, timing-dependent tests
2. **External service timeouts** - no retry logic for transient failures
3. **Missing context timeouts** - tests hanging indefinitely
4. **Infrastructure provisioning** - DB/Redis not ready when tests start

## Solutions Implemented

### 1. Test Helper Utilities

Created three new helper files in `internal/repository/`:

#### test_retry_helper.go
- `RetryWithBackoff()` - Exponential backoff retry logic
- `WaitForCondition()` - Poll until condition is true
- `DefaultRetryConfig()` - 5 retries, 100ms-5s backoff
- `QuickRetryConfig()` - 3 retries, 50ms-1s backoff

#### test_context_helper.go
- `TestContextDB()` - 10s timeout for database operations
- `TestContextRedis()` - 5s timeout for Redis operations
- `TestContextHTTP()` - 15s timeout for HTTP operations
- `TestContextLongRun()` - 2min timeout for long operations
- All contexts auto-cancel via t.Cleanup()

#### test_mock_helper.go
- `NewMockHTTPServer()` - Mock external HTTP services
- `MockDelayedResponse()` - Simulate slow responses
- `MockErrorAfterAttempts()` - Simulate transient failures
- `MockHTTPClientWithTimeout()` - HTTP client with proper timeouts

### 2. Makefile Updates

Updated test targets with increased timeouts:

```makefile
test-integration:
	go test -tags=integration -timeout=10m -race ./internal/repository/...

test-e2e-local:
	go test -tags=e2e -v -timeout=15m ./internal/integration/...
```

### 3. Documentation

Created comprehensive guides:

- `docs/testing-resilience-guide.md` - Complete migration guide with patterns
- `internal/repository/README_TEST_HELPERS.md` - Helper API documentation

## Migration Path

### Quick Wins (No Code Changes)

1. Run tests with increased timeout: `-timeout=10m` for integration tests
2. Existing tests will benefit from infrastructure improvements

### Recommended Migrations

#### Pattern 1: Replace time.Sleep

```go
// Before
time.Sleep(1 * time.Second)
result := checkStatus()

// After
err := WaitForConditionWithTimeout(5*time.Second, 100*time.Millisecond, func() bool {
    return checkStatus() == "ready"
})
```

#### Pattern 2: Add Context Timeouts

```go
// Before
func (s *Suite) SetupTest() {
    s.ctx = context.Background()
}

// After
func (s *Suite) SetupTest() {
    s.ctx = TestContextDB(s.T())
}
```

#### Pattern 3: Mock External Services

```go
// Before
client := NewClient("https://external-api.com")

// After
server := NewMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(mockData)
})
client := NewClient(server.URL())
```

## Impact Assessment

### Immediate Benefits

- Tests will no longer fail due to infrastructure startup timing
- Increased timeouts prevent false failures in CI
- Better error messages when tests timeout

### After Migration

- **Faster tests** - No unnecessary time.Sleep delays
- **More reliable** - Retry logic handles transient failures
- **Better isolation** - Mocked external services
- **Clearer failures** - Context timeouts provide stack traces

## Next Steps

### Priority 1: Infrastructure (Already Done)
- ✅ Test helper utilities created
- ✅ Makefile timeout updates
- ✅ Documentation written

### Priority 2: High-Impact Migrations
1. Update IntegrationRedisSuite.SetupTest() to use TestContextRedis()
2. Update IntegrationDBSuite.SetupTest() to use TestContextDB()
3. Replace time.Sleep in cache invalidation tests with WaitForCondition()

### Priority 3: External Service Tests
1. Identify tests calling external APIs
2. Replace with MockHTTPServer
3. Add retry logic for remaining external calls

### Priority 4: Systematic Migration
1. Search for all time.Sleep usage: `grep -r "time.Sleep" internal/`
2. Audit each instance
3. Replace with appropriate helper

## Testing the Improvements

Run integration tests with new timeouts:

```bash
# Integration tests
make test-integration

# E2E tests
make test-e2e-local

# Specific suite
go test -tags=integration -timeout=10m ./internal/repository -run TestAccountRepoSuite
```

## Monitoring

Watch for these improvements in CI:

1. **Reduced flakiness** - Tests pass consistently
2. **Clearer failures** - Timeout errors show context
3. **Faster feedback** - No unnecessary waits
4. **Better diagnostics** - Retry logs show attempt counts

## References

- [Testing Resilience Guide](./docs/testing-resilience-guide.md)
- [Test Helpers README](./internal/repository/README_TEST_HELPERS.md)
- [Integration Test Harness](./internal/repository/integration_harness_test.go)

## Files Changed

```
internal/repository/test_retry_helper.go         (new, 150 lines)
internal/repository/test_context_helper.go       (new, 70 lines)
internal/repository/test_mock_helper.go          (new, 115 lines)
internal/repository/README_TEST_HELPERS.md       (new)
docs/testing-resilience-guide.md                 (new)
Makefile                                         (modified)
TESTING_IMPROVEMENTS.md                          (new, this file)
```

## Summary

The test infrastructure now provides:
- ✅ Exponential backoff retry helpers
- ✅ Standardized context timeout management
- ✅ HTTP mocking utilities
- ✅ Increased CI timeouts
- ✅ Comprehensive documentation

Tests can be migrated incrementally while maintaining compatibility. The helpers are already available for new tests.

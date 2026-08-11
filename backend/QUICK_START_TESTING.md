# Quick Start: Fixing CI Test Failures

## Immediate Actions (No Code Changes Required)

### 1. Run Tests with Increased Timeouts

The Makefile has been updated with proper timeouts:

```bash
# Run integration tests (10 minute timeout)
make test-integration

# Run E2E tests (15 minute timeout)
make test-e2e-local
```

Or directly:

```bash
go test -tags=integration -timeout=10m -race ./internal/repository/...
go test -tags=e2e -timeout=15m -v ./internal/integration/...
```

This alone should fix most timeout-related CI failures.

### 2. Check golangci-lint Issues

```bash
golangci-lint run ./...
```

The `.golangci.yml` configuration is already set up with proper rules.

## Understanding the Test Helpers

Three new helper files have been created in `internal/repository/`:

### test_retry_helper.go - Fix Flaky Tests

```go
// Instead of waiting blindly:
time.Sleep(1 * time.Second)

// Use intelligent retry:
err := WaitForConditionWithTimeout(5*time.Second, 100*time.Millisecond, func() bool {
    return resourceIsReady()
})
```

### test_context_helper.go - Fix Timeout Issues

```go
// Instead of:
ctx := context.Background()

// Use:
ctx := TestContextDB(t)        // 10s for DB operations
ctx := TestContextRedis(t)     // 5s for Redis operations
ctx := TestContextHTTP(t)      // 15s for HTTP operations
```

### test_mock_helper.go - Fix External Service Issues

```go
// Instead of calling real APIs:
client := NewClient("https://external-api.com")

// Use mock server:
server := NewMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    w.Write([]byte(`{"status":"ok"}`))
})
client := NewClient(server.URL())
```

## Common CI Failure Patterns and Fixes

### Pattern 1: "context deadline exceeded"

**Problem:** Test hangs and times out

**Fix:** Add context timeout

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

### Pattern 2: "race condition detected"

**Problem:** Test passes locally, fails in CI with `-race` flag

**Fix:** Use proper synchronization or retry logic

```go
// Before
cache.Invalidate("key")
val := cache.Get("key") // Race: might not be invalidated yet

// After
cache.Invalidate("key")
err := WaitForCondition(ctx, 50*time.Millisecond, func() bool {
    return cache.Get("key") == nil
})
```

### Pattern 3: "connection refused" or "dial tcp: timeout"

**Problem:** External service not available or slow in CI

**Fix:** Mock the service

```go
// Before
resp, err := http.Get("https://api.example.com/data")

// After
server := NewMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    json.NewEncoder(w).Encode(mockData)
})
client := NewClient(server.URL())
resp, err := client.Get("/data")
```

### Pattern 4: "test failed after 1 second sleep"

**Problem:** Fixed sleep not long enough for CI environment

**Fix:** Replace with condition polling

```go
// Before
worker.Start()
time.Sleep(1 * time.Second)
require.True(t, worker.IsReady())

// After
worker.Start()
err := WaitForConditionWithTimeout(5*time.Second, 100*time.Millisecond, func() bool {
    return worker.IsReady()
})
require.NoError(t, err)
```

## Migration Priority

### High Priority (Fix These First)

1. **Update base test suites** in `integration_harness_test.go`:
   - IntegrationRedisSuite.SetupTest() → use TestContextRedis()
   - IntegrationDBSuite.SetupTest() → use TestContextDB()

2. **Cache invalidation tests** - Replace time.Sleep with WaitForCondition

3. **HTTP client tests** - Add timeouts or mock servers

### Medium Priority

1. **Async worker tests** - Use WaitForCondition instead of time.Sleep
2. **Polling tests** - Use RetryWithBackoff for better reliability

### Low Priority

1. **Deliberate delays** - Some time.Sleep calls are intentional
2. **Rate limiting tests** - May need actual delays

## Testing Your Changes

### Run Specific Test Suite

```bash
go test -tags=integration -timeout=10m -v ./internal/repository -run TestAccountRepoSuite
```

### Run Single Test

```bash
go test -tags=integration -timeout=10m -v ./internal/repository -run TestAccountRepoSuite/TestCreate
```

### Check for Flakiness (Run 10 Times)

```bash
for i in {1..10}; do
  echo "Run $i"
  go test -tags=integration -timeout=10m ./internal/repository -run TestYourTest || break
done
```

## Verification Checklist

Before committing test changes:

- [ ] Tests pass locally with `-race` flag
- [ ] Tests pass multiple times (not flaky)
- [ ] Proper context timeouts are set
- [ ] External services are mocked
- [ ] No fixed `time.Sleep` for waiting on async operations
- [ ] golangci-lint passes

## Getting Help

- Read [Testing Resilience Guide](./docs/testing-resilience-guide.md) for detailed patterns
- Check [Test Helpers README](./internal/repository/README_TEST_HELPERS.md) for API reference
- Look at [Integration Harness](./internal/repository/integration_harness_test.go) for base suite examples

## Summary of Changes

```
✅ Test helper utilities created (retry, context, mocking)
✅ Makefile updated with proper timeouts
✅ Documentation written

📋 TODO: Migrate existing tests to use helpers
📋 TODO: Update base test suites to use context helpers
📋 TODO: Replace time.Sleep with retry logic
```

## Quick Command Reference

```bash
# Run all integration tests
make test-integration

# Run E2E tests
make test-e2e-local

# Lint code
golangci-lint run ./...

# Run tests with race detection
go test -tags=integration -race -timeout=10m ./internal/repository/...

# Find all time.Sleep usage
grep -rn "time.Sleep" internal/ --include="*.go"

# Count time.Sleep instances
grep -r "time.Sleep" internal/ --include="*.go" | wc -l
```

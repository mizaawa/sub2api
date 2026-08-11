# Testing Resilience Guide

## Overview

This guide documents the test infrastructure improvements to address CI flakiness, timeouts, and external service failures.

## Problem Statement

Version v0.1.190 introduced test failures caused by:

1. **139 instances of time.Sleep** causing flaky tests
2. **External service timeouts** without proper retry logic
3. **Missing context timeout handling** in integration tests
4. **Fixed 1-second sleeps** in E2E tests between requests

## Solution Components

### 1. Exponential Backoff Retry (test_retry_helper.go)

Replace fixed time.Sleep with intelligent retry logic:

```go
// Replace this pattern:
time.Sleep(1 * time.Second)
result := fetchData()

// With exponential backoff:
err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
    return fetchData()
})
```

Available configurations:
- DefaultRetryConfig() - 5 retries, 100ms to 5s backoff (integration tests)
- QuickRetryConfig() - 3 retries, 50ms to 1s backoff (fast tests)

Key functions:
- RetryWithBackoff() - Execute with exponential backoff
- WaitForCondition() - Poll until condition is true
- WaitForConditionWithTimeout() - Convenience wrapper with timeout

### 2. Context Timeout Management (test_context_helper.go)

Standardized context timeouts for different operation types:

```go
// Database operations
ctx := TestContextDB(t) // 10s timeout

// Redis operations
ctx := TestContextRedis(t) // 5s timeout

// HTTP operations
ctx := TestContextHTTP(t) // 15s timeout

// Long-running operations
ctx := TestContextLongRun(t) // 2min timeout

// Custom timeout
ctx := TestContext(t, 45*time.Second)
```

All contexts automatically cancel on test cleanup via t.Cleanup().

### 3. HTTP Mock Helpers (test_mock_helper.go)

Mock external services to eliminate network flakiness.

## Migration Patterns

### Pattern 1: Replace time.Sleep with Retry

Before:
```go
func TestSomething(t *testing.T) {
    createResource()
    time.Sleep(1 * time.Second)
    result := checkResource()
    require.True(t, result.Ready)
}
```

After:
```go
func TestSomething(t *testing.T) {
    ctx := TestContextDefault(t)
    createResource()
    
    err := WaitForCondition(ctx, 100*time.Millisecond, func() bool {
        result := checkResource()
        return result.Ready
    })
    require.NoError(t, err)
}
```

### Pattern 2: Add Context Timeouts

Before:
```go
func (s *RepoSuite) TestQuery() {
    s.ctx = context.Background() // No timeout
    result, err := s.repo.Query(s.ctx, params)
}
```

After:
```go
func (s *RepoSuite) SetupTest() {
    s.ctx = TestContextDB(s.T())
    // rest of setup
}
```

### Pattern 3: Mock External Services

Before:
```go
func TestAPICall(t *testing.T) {
    // Calls real external API - flaky in CI
    result, err := client.Call("https://api.example.com/data")
}
```

After:
```go
func TestAPICall(t *testing.T) {
    server := NewMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(mockData)
    })
    
    client := NewClient(server.URL())
    result, err := client.Call("/data")
}
```

## Integration Test Best Practices

1. Always Use Context Timeouts
2. Use Retry for Eventually-Consistent Operations
3. Mock External Dependencies
4. Set Appropriate Timeouts:
   - DB queries: 10s (TestContextDB)
   - Redis operations: 5s (TestContextRedis)
   - HTTP calls: 15s (TestContextHTTP)
   - Background jobs: 2min (TestContextLongRun)

## CI Configuration Updates

Recommended test flags:

```makefile
test-integration:
	go test -tags=integration -timeout=10m -race ./internal/repository/...

test-e2e:
	go test -tags=e2e -timeout=15m -v ./internal/integration/...
```

## Summary

- Replace time.Sleep with RetryWithBackoff or WaitForCondition
- Always use context timeouts via TestContext helpers
- Mock external services with MockHTTPServer
- Test timeout scenarios explicitly
- Increase test timeout flags in CI (10-15min for integration/E2E)

These changes improve test reliability, reduce flakiness, and provide better error messages when tests fail.

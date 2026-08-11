# Integration Test Helpers

This directory contains test helper utilities for integration tests.

## Available Helpers

### 1. test_retry_helper.go - Exponential Backoff and Retry

Replace fixed `time.Sleep` calls with intelligent retry logic:

```go
// Retry with default config (5 retries, 100ms-5s backoff)
err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
    return someOperation()
})

// Quick retry for fast operations (3 retries, 50ms-1s backoff)
err := RetryWithBackoff(ctx, QuickRetryConfig(), func() error {
    return quickOperation()
})

// Wait for a condition to become true
err := WaitForConditionWithTimeout(5*time.Second, 100*time.Millisecond, func() bool {
    return resourceIsReady()
})
```

### 2. test_context_helper.go - Context Timeout Management

Standardized context timeouts with automatic cleanup:

```go
// Database operations (10s timeout)
ctx := TestContextDB(t)

// Redis operations (5s timeout)
ctx := TestContextRedis(t)

// HTTP operations (15s timeout)
ctx := TestContextHTTP(t)

// Long-running operations (2min timeout)
ctx := TestContextLongRun(t)

// Custom timeout
ctx := TestContext(t, 30*time.Second)
```

All contexts are automatically cancelled via `t.Cleanup()`.

### 3. test_mock_helper.go - HTTP Mocking

Mock external HTTP services:

```go
// Basic mock server
server := NewMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(200)
    w.Write([]byte(`{"status":"ok"}`))
})

// Delayed response
handler := MockDelayedResponse(200*time.Millisecond, 200, "success")

// Simulate transient failures (fails twice, then succeeds)
handler := MockErrorAfterAttempts(2, 500, 200, "success")

// Timeout scenario
handler := MockTimeoutResponse()

// HTTP client with timeout
client := MockHTTPClientWithTimeout(5 * time.Second)
```

## Usage Examples

### Example 1: Test Eventually-Consistent Cache

Before:
```go
func (s *CacheSuite) TestInvalidation() {
    s.cache.Invalidate("key")
    time.Sleep(1 * time.Second) // Hope it's done
    val := s.cache.Get("key")
    s.Nil(val)
}
```

After:
```go
func (s *CacheSuite) TestInvalidation() {
    ctx := TestContextRedis(s.T())
    s.cache.Invalidate("key")
    
    err := WaitForCondition(ctx, 50*time.Millisecond, func() bool {
        return s.cache.Get("key") == nil
    })
    s.NoError(err)
}
```

### Example 2: Test Async Processing

Before:
```go
func TestWorker(t *testing.T) {
    queue.Enqueue(job)
    time.Sleep(2 * time.Second)
    require.True(t, job.IsComplete())
}
```

After:
```go
func TestWorker(t *testing.T) {
    ctx := TestContextDefault(t)
    queue.Enqueue(job)
    
    err := WaitForCondition(ctx, 100*time.Millisecond, func() bool {
        return job.IsComplete()
    })
    require.NoError(t, err)
}
```

### Example 3: Test with External Service Mock

Before:
```go
func TestService(t *testing.T) {
    // Calls real API, flaky in CI
    client := NewClient("https://api.example.com")
    result, err := client.Fetch()
}
```

After:
```go
func TestService(t *testing.T) {
    server := NewMockHTTPServer(t, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(mockData)
    })
    
    client := NewClient(server.URL())
    result, err := client.Fetch()
}
```

### Example 4: Test with Retry Logic

Before:
```go
func TestFlaky(t *testing.T) {
    result, err := flakyOperation()
    require.NoError(t, err) // Fails randomly
}
```

After:
```go
func TestFlaky(t *testing.T) {
    ctx := TestContextDefault(t)
    var result *Result
    
    err := RetryWithBackoff(ctx, DefaultRetryConfig(), func() error {
        var err error
        result, err = flakyOperation()
        return err
    })
    require.NoError(t, err)
}
```

## Best Practices

1. **Always use context timeouts** - Replace `context.Background()` with `TestContext*()` helpers
2. **Replace time.Sleep** - Use `RetryWithBackoff` or `WaitForCondition` instead
3. **Mock external services** - Use `NewMockHTTPServer` instead of calling real APIs
4. **Choose appropriate timeouts**:
   - DB queries: `TestContextDB(t)` - 10s
   - Redis ops: `TestContextRedis(t)` - 5s
   - HTTP calls: `TestContextHTTP(t)` - 15s
   - Long jobs: `TestContextLongRun(t)` - 2min
5. **Test resilience** - Use `MockErrorAfterAttempts` to test retry logic

## See Also

- [Testing Resilience Guide](../../docs/testing-resilience-guide.md) - Complete migration guide
- [integration_harness_test.go](./integration_harness_test.go) - Base test suites

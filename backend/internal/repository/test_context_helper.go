//go:build integration

package repository

import (
	"context"
	"testing"
	"time"
)

const (
	// Default timeouts for different test operation types
	defaultTestTimeout      = 30 * time.Second
	defaultDBTimeout        = 10 * time.Second
	defaultRedisTimeout     = 5 * time.Second
	defaultHTTPTimeout      = 15 * time.Second
	defaultLongRunTimeout   = 2 * time.Minute
)

// TestContext creates a context with timeout for integration tests
func TestContext(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	
	return ctx
}

// TestContextDefault creates a context with default timeout (30s)
func TestContextDefault(t *testing.T) context.Context {
	return TestContext(t, defaultTestTimeout)
}

// TestContextDB creates a context optimized for database operations
func TestContextDB(t *testing.T) context.Context {
	return TestContext(t, defaultDBTimeout)
}

// TestContextRedis creates a context optimized for Redis operations
func TestContextRedis(t *testing.T) context.Context {
	return TestContext(t, defaultRedisTimeout)
}

// TestContextHTTP creates a context optimized for HTTP operations
func TestContextHTTP(t *testing.T) context.Context {
	return TestContext(t, defaultHTTPTimeout)
}

// TestContextLongRun creates a context for long-running operations
func TestContextLongRun(t *testing.T) context.Context {
	return TestContext(t, defaultLongRunTimeout)
}

// WithTestDeadline wraps an existing context with test cleanup
func WithTestDeadline(t *testing.T, ctx context.Context, timeout time.Duration) context.Context {
	t.Helper()
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	t.Cleanup(cancel)
	
	return ctx
}

// ContextWithValue creates a context with value and test cleanup
func ContextWithValue(t *testing.T, parent context.Context, key, val interface{}) context.Context {
	t.Helper()
	
	if parent == nil {
		parent = context.Background()
	}
	
	return context.WithValue(parent, key, val)
}

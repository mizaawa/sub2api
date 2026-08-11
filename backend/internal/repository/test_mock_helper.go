//go:build integration

package repository

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// MockHTTPServer creates a test HTTP server with configurable responses
type MockHTTPServer struct {
	server  *httptest.Server
	handler http.HandlerFunc
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer(t *testing.T, handler http.HandlerFunc) *MockHTTPServer {
	t.Helper()
	
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	
	return &MockHTTPServer{
		server:  server,
		handler: handler,
	}
}

// URL returns the base URL of the mock server
func (m *MockHTTPServer) URL() string {
	return m.server.URL
}

// Close closes the mock server
func (m *MockHTTPServer) Close() {
	m.server.Close()
}

// MockHTTPClientWithTimeout creates an HTTP client with timeout for testing
func MockHTTPClientWithTimeout(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			ResponseHeaderTimeout: timeout / 2,
			IdleConnTimeout:       timeout,
			DisableKeepAlives:     true,
		},
	}
}

// MockDelayedResponse simulates a delayed HTTP response
func MockDelayedResponse(delay time.Duration, statusCode int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			http.Error(w, "request cancelled", http.StatusRequestTimeout)
			return
		case <-time.After(delay):
		}
		
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}
}

// MockTimeoutResponse simulates a response that never completes
func MockTimeoutResponse() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}
}

// MockErrorAfterAttempts returns success after N failed attempts
func MockErrorAfterAttempts(attempts int, errorCode int, successCode int, body string) http.HandlerFunc {
	count := 0
	return func(w http.ResponseWriter, r *http.Request) {
		count++
		if count <= attempts {
			w.WriteHeader(errorCode)
			_, _ = w.Write([]byte("error"))
			return
		}
		
		w.WriteHeader(successCode)
		_, _ = w.Write([]byte(body))
	}
}

// AssertContextNotExpired checks that the context hasn't timed out
func AssertContextNotExpired(t *testing.T, ctx context.Context) {
	t.Helper()
	
	select {
	case <-ctx.Done():
		t.Fatalf("context expired: %v", ctx.Err())
	default:
		// Context is still valid
	}
}

// AssertWithinDuration checks that a function completes within the given duration
func AssertWithinDuration(t *testing.T, duration time.Duration, fn func()) {
	t.Helper()
	
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()
	
	select {
	case <-done:
		// Success
	case <-time.After(duration):
		t.Fatalf("operation did not complete within %v", duration)
	}
}

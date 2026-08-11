//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// RetryConfig holds configuration for exponential backoff retries
type RetryConfig struct {
	MaxRetries      int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	BackoffMultiple float64
	Timeout         time.Duration
}

// DefaultRetryConfig returns sensible defaults for test retries
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      5,
		InitialBackoff:  100 * time.Millisecond,
		MaxBackoff:      5 * time.Second,
		BackoffMultiple: 2.0,
		Timeout:         30 * time.Second,
	}
}

// QuickRetryConfig returns faster retry configuration for unit-like tests
func QuickRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		InitialBackoff:  50 * time.Millisecond,
		MaxBackoff:      1 * time.Second,
		BackoffMultiple: 2.0,
		Timeout:         10 * time.Second,
	}
}

// RetryWithBackoff executes fn with exponential backoff until success or max retries
func RetryWithBackoff(ctx context.Context, cfg RetryConfig, fn func() error) error {
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("retry context cancelled after %d attempts: %w (last error: %v)", attempt, ctx.Err(), lastErr)
			case <-time.After(backoff):
			}

			backoff = time.Duration(float64(backoff) * cfg.BackoffMultiple)
			if backoff > cfg.MaxBackoff {
				backoff = cfg.MaxBackoff
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if errors.Is(lastErr, context.Canceled) || errors.Is(lastErr, context.DeadlineExceeded) {
			return lastErr
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", cfg.MaxRetries+1, lastErr)
}

// WaitForCondition polls condition until it returns true or context times out
func WaitForCondition(ctx context.Context, interval time.Duration, condition func() bool) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if condition() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// WaitForConditionWithTimeout is a convenience wrapper for WaitForCondition with timeout
func WaitForConditionWithTimeout(timeout, interval time.Duration, condition func() bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WaitForCondition(ctx, interval, condition)
}

// ExponentialBackoffSequence generates a sequence of backoff durations
func ExponentialBackoffSequence(initial, max time.Duration, multiplier float64, count int) []time.Duration {
	sequence := make([]time.Duration, count)
	current := initial

	for i := 0; i < count; i++ {
		sequence[i] = current
		current = time.Duration(float64(current) * multiplier)
		if current > max {
			current = max
		}
	}

	return sequence
}

// JitterBackoff adds random jitter to prevent thundering herd
func JitterBackoff(base time.Duration, jitterFactor float64) time.Duration {
	if jitterFactor <= 0 {
		return base
	}

	jitter := time.Duration(float64(base) * jitterFactor * (0.5 + 0.5*float64(time.Now().UnixNano()%1000)/1000.0))
	return base + jitter
}

// CalculateBackoff computes exponential backoff with jitter
func CalculateBackoff(attempt int, initial, max time.Duration, multiplier, jitter float64) time.Duration {
	backoff := float64(initial) * math.Pow(multiplier, float64(attempt))
	if backoff > float64(max) {
		backoff = float64(max)
	}

	if jitter > 0 {
		backoff += float64(initial) * jitter * (0.5 + 0.5*float64(time.Now().UnixNano()%1000)/1000.0)
	}

	return time.Duration(backoff)
}

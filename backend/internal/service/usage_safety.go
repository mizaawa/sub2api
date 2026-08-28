package service

import (
	"errors"
	"fmt"
	"math"

	"github.com/tidwall/gjson"
)

// Upstream usage is untrusted billing input. Keep this above currently
// supported context windows while rejecting malformed or injected values.
const maxReportedUsageTokens = 100_000_000

// ErrInvalidUsageTokens is returned when usage reaches a billing boundary in
// a form that is not safe to account for.  Upstream adapters normally reject
// these values earlier, but keeping the invariant at the billing boundary
// also protects internal callers and historical compatibility paths.
var ErrInvalidUsageTokens = errors.New("invalid usage token counts")

// maxBillableRequestCount keeps count-based billing from turning a malformed
// integer into an effectively unbounded charge.  It intentionally shares the
// same conservative ceiling as token usage reported by upstreams.
const maxBillableRequestCount = maxReportedUsageTokens

// validateUsageTokens validates the in-memory representation used by all
// token pricing paths.  Every bucket is bounded before arithmetic or float
// conversion; the detailed cache buckets must also fit in one bounded cache
// creation total.
func validateUsageTokens(tokens UsageTokens) error {
	fields := []struct {
		name  string
		value int
	}{
		{"input_tokens", tokens.InputTokens},
		{"image_input_tokens", tokens.ImageInputTokens},
		{"output_tokens", tokens.OutputTokens},
		{"cache_creation_tokens", tokens.CacheCreationTokens},
		{"cache_read_tokens", tokens.CacheReadTokens},
		{"cache_creation_5m_tokens", tokens.CacheCreation5mTokens},
		{"cache_creation_1h_tokens", tokens.CacheCreation1hTokens},
		{"image_output_tokens", tokens.ImageOutputTokens},
	}
	for _, field := range fields {
		if field.value < 0 || field.value > maxReportedUsageTokens {
			return fmt.Errorf("%w: %s=%d", ErrInvalidUsageTokens, field.name, field.value)
		}
	}
	if _, ok := addBoundedReportedUsageInts(tokens.CacheCreation5mTokens, tokens.CacheCreation1hTokens); !ok {
		return fmt.Errorf("%w: cache_creation_breakdown", ErrInvalidUsageTokens)
	}
	return nil
}

func validUsageTokens(tokens UsageTokens) bool {
	return validateUsageTokens(tokens) == nil
}

// sumNonNegativeInts performs overflow-checked addition for context totals.
// It deliberately does not apply the per-bucket usage ceiling: a legitimate
// request can have input, cache, and output buckets whose combined total is
// greater than one bucket's limit.
func sumNonNegativeInts(values ...int) (int, bool) {
	total := 0
	for _, value := range values {
		if value < 0 || total > math.MaxInt-value {
			return 0, false
		}
		total += value
	}
	return total, true
}

func validBillableRequestCount(count int) bool {
	return count >= 0 && count <= maxBillableRequestCount
}

func validReportedUsageTokenCount(value int64) bool {
	return value >= 0 && value <= maxReportedUsageTokens
}

func sanitizeOpenAIUsage(usage *OpenAIUsage) bool {
	if usage == nil {
		return false
	}
	if !validReportedUsageTokenCount(int64(usage.InputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.ImageInputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.OutputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.CacheCreationInputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.CacheReadInputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.ImageOutputTokens)) {
		*usage = OpenAIUsage{}
		return false
	}
	return true
}

func sanitizeClaudeUsage(usage *ClaudeUsage) bool {
	if usage == nil {
		return false
	}
	if !validReportedUsageTokenCount(int64(usage.InputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.OutputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.CacheCreationInputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.CacheReadInputTokens)) ||
		!validReportedUsageTokenCount(int64(usage.CacheCreation5mTokens)) ||
		!validReportedUsageTokenCount(int64(usage.CacheCreation1hTokens)) ||
		usage.CacheCreation5mTokens > maxReportedUsageTokens-usage.CacheCreation1hTokens ||
		!validReportedUsageTokenCount(int64(usage.ImageOutputTokens)) {
		*usage = ClaudeUsage{}
		return false
	}
	return true
}

// boundedReportedUsageGJSONInt parses one upstream JSON token field and
// applies the same upper bound used by the in-memory usage sanitizers. The
// parser accepts integral decimal/exponent notation (for provider variants
// that serialize numbers that way) but rejects strings, fractions, negatives,
// and values that exceed the billing guard.
func boundedReportedUsageGJSONInt(value gjson.Result) (int, bool) {
	if !value.Exists() || value.Type == gjson.Null {
		return 0, false
	}
	parsed, ok := boundedJSONNonNegativeInt(value)
	if !ok || parsed < 0 || parsed > maxReportedUsageTokens {
		return 0, false
	}
	return parsed, true
}

// validateOpenAIUsageJSON validates every token-shaped field that any OpenAI
// usage parser currently reads. Unknown metadata is intentionally ignored, but
// a present known field must be a bounded non-negative integer. This is used
// for both the top-level usage object and nested hosted-tool usage objects.
func validateOpenAIUsageJSON(value gjson.Result) bool {
	if !value.Exists() || !value.IsObject() {
		return false
	}
	for _, path := range []string{
		"input_tokens",
		"prompt_tokens",
		"output_tokens",
		"completion_tokens",
		"total_tokens",
		"input_tokens_details.cached_tokens",
		"prompt_tokens_details.cached_tokens",
		"input_tokens_details.cache_write_tokens",
		"prompt_tokens_details.cache_write_tokens",
		"input_tokens_details.cache_creation_tokens",
		"prompt_tokens_details.cache_creation_tokens",
		"output_tokens_details.image_tokens",
		"completion_tokens_details.image_tokens",
		"input_tokens_details.image_tokens",
		"prompt_tokens_details.image_tokens",
		"cache_read_input_tokens",
		"cache_read_tokens",
		"cached_tokens",
		"cache_write_tokens",
		"cache_creation_input_tokens",
		"cache_write_input_tokens",
		"cache_creation_tokens",
	} {
		field := value.Get(path)
		if !field.Exists() || field.Type == gjson.Null {
			continue
		}
		if _, ok := boundedReportedUsageGJSONInt(field); !ok {
			return false
		}
	}
	return true
}

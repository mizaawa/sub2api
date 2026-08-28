package service

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/tidwall/gjson"
)

// parseBoundedReportedUsageInt accepts the numeric representations emitted by
// the provider adapters while enforcing one invariant at the boundary: usage
// values are finite, integral, non-negative, and bounded before any arithmetic
// or conversion to the platform int type occurs.
//
// JSON decoded into map[string]any normally arrives as float64, while a few
// adapters retain json.Number or string values. Supporting all of those forms
// keeps the parser compatible without allowing truncation (for example 1.9 ->
// 1) or float overflow to manufacture a billable value.
func parseBoundedReportedUsageInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return boundedReportedUsageNativeInt(v)
	case int8:
		return boundedReportedUsageNativeInt(int(v))
	case int16:
		return boundedReportedUsageNativeInt(int(v))
	case int32:
		return boundedReportedUsageNativeInt(int(v))
	case int64:
		if v < 0 || v > int64(maxReportedUsageTokens) {
			return 0, false
		}
		return int(v), true
	case uint:
		if uint64(v) > uint64(maxReportedUsageTokens) {
			return 0, false
		}
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		if uint64(v) > uint64(maxReportedUsageTokens) {
			return 0, false
		}
		return int(v), true
	case uint64:
		if v > uint64(maxReportedUsageTokens) {
			return 0, false
		}
		return int(v), true
	case float32:
		return parseBoundedReportedUsageFloat(float64(v))
	case float64:
		return parseBoundedReportedUsageFloat(v)
	case json.Number:
		return parseBoundedReportedUsageText(string(v))
	case string:
		return parseBoundedReportedUsageText(strings.TrimSpace(v))
	default:
		return 0, false
	}
}

func boundedReportedUsageNativeInt(value int) (int, bool) {
	if value < 0 || value > maxReportedUsageTokens {
		return 0, false
	}
	return value, true
}

func parseBoundedReportedUsageFloat(value float64) (int, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 ||
		value > float64(maxReportedUsageTokens) || math.Trunc(value) != value {
		return 0, false
	}
	return int(value), true
}

func parseBoundedReportedUsageText(raw string) (int, bool) {
	if raw == "" || !gjson.Valid(raw) {
		return 0, false
	}
	parsed := gjson.Parse(raw)
	if parsed.Type != gjson.Number {
		return 0, false
	}
	return boundedReportedUsageGJSONInt(parsed)
}

func readOptionalBoundedUsageGJSONInt(value gjson.Result) (int, bool) {
	if !value.Exists() || value.Type == gjson.Null {
		return 0, true
	}
	return boundedReportedUsageGJSONInt(value)
}

func validateClaudeUsageJSON(value gjson.Result) bool {
	if !value.Exists() || !value.IsObject() {
		return false
	}
	for _, path := range []string{
		"input_tokens",
		"output_tokens",
		"cache_creation_input_tokens",
		"cache_read_input_tokens",
		"cached_tokens",
		"image_output_tokens",
		"cache_creation.ephemeral_5m_input_tokens",
		"cache_creation.ephemeral_1h_input_tokens",
	} {
		if _, ok := readOptionalBoundedUsageGJSONInt(value.Get(path)); !ok {
			return false
		}
	}
	if cacheCreation := value.Get("cache_creation"); cacheCreation.Exists() &&
		cacheCreation.Type != gjson.Null && !cacheCreation.IsObject() {
		return false
	}
	return true
}

func addBoundedReportedUsageInts(left, right int) (int, bool) {
	if left < 0 || right < 0 || left > maxReportedUsageTokens ||
		right > maxReportedUsageTokens || left > maxReportedUsageTokens-right {
		return 0, false
	}
	return left + right, true
}

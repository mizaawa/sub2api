package openai_ws_v2

import (
	"math"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// boundedReportedUsageGJSONInt is the WS relay's local copy of the gateway
// usage boundary.  This package is intentionally independent from the parent
// service package, so the parser must enforce the same invariant here before
// any conversion or accumulation occurs.
func boundedReportedUsageGJSONInt(value gjson.Result) (int, bool) {
	if !value.Exists() || value.Type != gjson.Number {
		return 0, false
	}
	raw := strings.TrimSpace(value.Raw)
	if raw == "" || len(raw) > 64 || strings.HasPrefix(raw, "-") {
		return 0, false
	}
	// Token counts are bounded at 100M, well below float64's exact integer
	// range. ParseFloat is used only after rejecting non-finite values and then
	// the lexical form is checked so tiny fractional exponents cannot underflow
	// to an accepted zero.
	if strings.ContainsAny(raw, ".eE") {
		if !isIntegralJSONNumber(raw) {
			return 0, false
		}
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) ||
		parsed < 0 || parsed > float64(maxReportedUsageTokens) || math.Trunc(parsed) != parsed {
		return 0, false
	}
	return int(parsed), true
}

// isIntegralJSONNumber validates the decimal/exponent spelling without
// allowing a fractional value to be rounded or underflowed by ParseFloat.
func isIntegralJSONNumber(raw string) bool {
	mantissaEnd := len(raw)
	for i := 0; i < len(raw); i++ {
		if raw[i] == 'e' || raw[i] == 'E' {
			mantissaEnd = i
			break
		}
	}
	mantissa := raw[:mantissaEnd]
	if mantissa == "" {
		return false
	}
	dot := strings.IndexByte(mantissa, '.')
	fractionDigits := 0
	digitCount := 0
	zero := true
	for i := 0; i < len(mantissa); i++ {
		c := mantissa[i]
		if c == '.' {
			if dot != i {
				return false
			}
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
		digitCount++
		zero = zero && c == '0'
		if dot >= 0 && i > dot {
			fractionDigits++
		}
	}
	if digitCount == 0 {
		return false
	}
	exponent := 0
	if mantissaEnd < len(raw) {
		expRaw := raw[mantissaEnd+1:]
		if expRaw == "" {
			return false
		}
		sign := 1
		if expRaw[0] == '+' || expRaw[0] == '-' {
			if expRaw[0] == '-' {
				sign = -1
			}
			expRaw = expRaw[1:]
		}
		if expRaw == "" || len(expRaw) > 3 {
			return false
		}
		for i := 0; i < len(expRaw); i++ {
			if expRaw[i] < '0' || expRaw[i] > '9' {
				return false
			}
			exponent = exponent*10 + int(expRaw[i]-'0')
		}
		if exponent > 100 {
			// Zero remains a valid zero regardless of exponent magnitude, but
			// reject a non-zero value before ParseFloat can overflow.
			if !zero {
				return false
			}
			return true
		}
		exponent *= sign
	}

	// A decimal is integral when all digits shifted past the decimal point
	// are zero.  Compute that condition using the raw digits and exponent.
	shift := exponent - fractionDigits
	if shift >= 0 {
		return true
	}
	remove := -shift
	for i := len(mantissa) - 1; i >= 0 && remove > 0; i-- {
		if mantissa[i] == '.' {
			continue
		}
		if mantissa[i] != '0' {
			return false
		}
		remove--
	}
	return remove == 0 || zero
}

func validateOpenAIUsageJSON(value gjson.Result) bool {
	if !value.Exists() || !value.IsObject() {
		return false
	}
	for _, path := range []string{
		"input_tokens", "prompt_tokens", "output_tokens", "completion_tokens", "total_tokens",
		"input_tokens_details.cached_tokens", "prompt_tokens_details.cached_tokens",
		"input_tokens_details.cache_write_tokens", "prompt_tokens_details.cache_write_tokens",
		"input_tokens_details.cache_creation_tokens", "prompt_tokens_details.cache_creation_tokens",
		"output_tokens_details.image_tokens", "completion_tokens_details.image_tokens",
		"input_tokens_details.image_tokens", "prompt_tokens_details.image_tokens",
		"cache_read_input_tokens", "cache_read_tokens", "cached_tokens", "cache_write_tokens",
		"cache_creation_input_tokens", "cache_write_input_tokens", "cache_creation_tokens",
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

package service

import (
	"math"
	"testing"

	"github.com/tidwall/gjson"
)

func TestParseBoundedReportedUsageIntRejectsFractionAndNonFinite(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "fractional float", value: 1.5},
		{name: "negative float", value: -1.0},
		{name: "nan", value: math.NaN()},
		{name: "positive infinity", value: math.Inf(1)},
		{name: "oversized integer", value: maxReportedUsageTokens + 1},
		{name: "fractional string", value: "1.5"},
		{name: "huge exponent string", value: "1e1000000000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, ok := parseBoundedReportedUsageInt(tt.value); ok || got != 0 {
				t.Fatalf("parseBoundedReportedUsageInt(%#v) = (%d, %v), want (0, false)", tt.value, got, ok)
			}
		})
	}

	for _, value := range []any{0, 1.0, "42", int64(99)} {
		if got, ok := parseBoundedReportedUsageInt(value); !ok || got < 0 {
			t.Fatalf("parseBoundedReportedUsageInt(%#v) = (%d, %v), want valid", value, got, ok)
		}
	}
}

func TestExtractSSEUsagePatchRejectsMalformedPresentField(t *testing.T) {
	service := &GatewayService{}
	malformed := map[string]any{
		"type":  "message_delta",
		"usage": map[string]any{"output_tokens": 1.5},
	}
	if patch := service.extractSSEUsagePatch(malformed); patch != nil {
		t.Fatalf("malformed usage produced patch: %+v", patch)
	}

	usage := &ClaudeUsage{InputTokens: 120, OutputTokens: 8}
	service.parseSSEUsage(`{"type":"message_delta","usage":{"output_tokens":1.5}}`, usage)
	if *usage != (ClaudeUsage{InputTokens: 120, OutputTokens: 8}) {
		t.Fatalf("malformed patch changed prior usage: %+v", *usage)
	}
}

func TestParseClaudeUsagePassthroughRejectsMalformedUsage(t *testing.T) {
	usage := &ClaudeUsage{InputTokens: 120, OutputTokens: 8}
	(&GatewayService{}).parseSSEUsagePassthrough(
		`{"type":"message_delta","usage":{"output_tokens":1e1000000000}}`,
		usage,
	)
	if *usage != (ClaudeUsage{InputTokens: 120, OutputTokens: 8}) {
		t.Fatalf("malformed passthrough usage changed prior usage: %+v", *usage)
	}

	parsed := parseClaudeUsageFromResponseBody([]byte(`{"usage":{"input_tokens":1.5,"output_tokens":2}}`))
	if parsed == nil || *parsed != (ClaudeUsage{}) {
		t.Fatalf("malformed non-stream usage was accepted: %+v", parsed)
	}
}

func TestAntigravitySSEUsageRejectsFractionalAndOverflowValues(t *testing.T) {
	usage := &ClaudeUsage{InputTokens: 120, OutputTokens: 8}
	(&AntigravityGatewayService{}).extractSSEUsage(
		`data: {"type":"message_delta","usage":{"output_tokens":1.5}}`,
		usage,
	)
	if *usage != (ClaudeUsage{InputTokens: 120, OutputTokens: 8}) {
		t.Fatalf("fractional Antigravity usage changed prior usage: %+v", *usage)
	}

	usage = &ClaudeUsage{}
	(&AntigravityGatewayService{}).extractSSEUsage(
		`data: {"type":"message_delta","usage":{"output_tokens":100000001}}`,
		usage,
	)
	if *usage != (ClaudeUsage{}) {
		t.Fatalf("oversized Antigravity usage was accepted: %+v", *usage)
	}
}

func TestExtractGeminiUsageRejectsInvalidArithmeticInputs(t *testing.T) {
	tests := []string{
		`{"usageMetadata":{"promptTokenCount":100000001,"candidatesTokenCount":1}}`,
		`{"usageMetadata":{"promptTokenCount":10,"cachedContentTokenCount":11,"candidatesTokenCount":1}}`,
		`{"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":1.5}}`,
		`{"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":1,"thoughtsTokenCount":100000000}}`,
		`{"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":1,"candidatesTokensDetails":[{"modality":"IMAGE","tokenCount":1e1000000000}]}}`,
	}
	for _, raw := range tests {
		got := extractGeminiUsage([]byte(raw))
		if got == nil || *got != (ClaudeUsage{}) {
			t.Fatalf("extractGeminiUsage(%s) = %+v, want zero usage", raw, got)
		}
	}

	valid := extractGeminiUsage([]byte(`{"usageMetadata":{"promptTokenCount":100,"cachedContentTokenCount":20,"candidatesTokenCount":30,"thoughtsTokenCount":5,"candidatesTokensDetails":[{"modality":"TEXT","tokenCount":30},{"modality":"IMAGE","tokenCount":7}]}}`))
	if valid == nil || valid.InputTokens != 80 || valid.OutputTokens != 35 || valid.CacheReadInputTokens != 20 || valid.ImageOutputTokens != 7 {
		t.Fatalf("valid Gemini usage parsed incorrectly: %+v", valid)
	}
}

func TestAddBoundedReportedUsageInts(t *testing.T) {
	if got, ok := addBoundedReportedUsageInts(maxReportedUsageTokens-1, 1); !ok || got != maxReportedUsageTokens {
		t.Fatalf("valid bounded addition = (%d, %v)", got, ok)
	}
	if got, ok := addBoundedReportedUsageInts(maxReportedUsageTokens, 1); ok || got != 0 {
		t.Fatalf("overflowing bounded addition = (%d, %v)", got, ok)
	}

	usage := &ClaudeUsage{InputTokens: 10, CacheReadInputTokens: maxReportedUsageTokens}
	if got, ok := addBoundedReportedUsageInts(usage.CacheReadInputTokens, usage.InputTokens); ok || got != 0 {
		t.Fatalf("force-cache style overflow = (%d, %v)", got, ok)
	}

	if node := gjson.Parse(`1e1000000000`); node.Type != gjson.Number {
		t.Fatalf("test fixture was not parsed as a number: %v", node.Type)
	}
}

func TestClaudeCacheRewriteRejectsMalformedOrOverflowingBreakdown(t *testing.T) {
	usage := ClaudeUsage{
		CacheCreation5mTokens: maxReportedUsageTokens,
		CacheCreation1hTokens: 1,
	}
	original := usage
	if applyCacheTTLOverride(&usage, "1h") {
		t.Fatal("overflowing cache breakdown should not be rewritten")
	}
	if usage != original {
		t.Fatalf("overflowing cache breakdown changed: before=%+v after=%+v", original, usage)
	}

	obj := map[string]any{
		"cache_creation": map[string]any{
			"ephemeral_5m_input_tokens": 1.5,
			"ephemeral_1h_input_tokens": 2.0,
		},
	}
	if rewriteCacheCreationJSON(obj, "1h") {
		t.Fatal("fractional cache breakdown should not be rewritten")
	}
	if got := obj["cache_creation"].(map[string]any)["ephemeral_5m_input_tokens"]; got != 1.5 {
		t.Fatalf("malformed cache breakdown changed: %#v", obj)
	}
}

func TestReconcileCachedTokensRejectsMalformedValues(t *testing.T) {
	usage := map[string]any{"cached_tokens": 1.5}
	if reconcileCachedTokens(usage) {
		t.Fatal("fractional cached_tokens should not be reconciled")
	}
	if _, exists := usage["cache_read_input_tokens"]; exists {
		t.Fatalf("malformed cached_tokens produced a cache_read field: %#v", usage)
	}

	usage = map[string]any{
		"cache_read_input_tokens": 0,
		"cached_tokens":           maxReportedUsageTokens + 1,
	}
	if reconcileCachedTokens(usage) {
		t.Fatal("oversized cached_tokens should not be reconciled")
	}
}

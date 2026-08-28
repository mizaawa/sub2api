package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/tidwall/gjson"
)

func TestSanitizeOpenAIUsageRejectsInjectedValues(t *testing.T) {
	usage := OpenAIUsage{InputTokens: maxReportedUsageTokens + 1, OutputTokens: 1}
	if sanitizeOpenAIUsage(&usage) {
		t.Fatal("expected oversized usage to be rejected")
	}
	if usage != (OpenAIUsage{}) {
		t.Fatalf("rejected usage must be zeroed, got %+v", usage)
	}

	usage = OpenAIUsage{InputTokens: 10, OutputTokens: 2, CacheReadInputTokens: 3}
	if !sanitizeOpenAIUsage(&usage) {
		t.Fatal("expected normal usage to be accepted")
	}
	if _, ok := extractOpenAIUsageFromJSONBytes([]byte(`{"usage":{"input_tokens":100000001,"output_tokens":1}}`)); ok {
		t.Fatal("expected oversized JSON usage to be rejected before billing")
	}
}

func TestSanitizeClaudeUsageRejectsNegativeValues(t *testing.T) {
	usage := ClaudeUsage{InputTokens: -1}
	if sanitizeClaudeUsage(&usage) {
		t.Fatal("expected negative usage to be rejected")
	}
	if usage != (ClaudeUsage{}) {
		t.Fatalf("rejected usage must be zeroed, got %+v", usage)
	}
}

func TestParseSSEUsagePassthroughRejectsInjectedValueWithoutLosingPriorUsage(t *testing.T) {
	usage := &ClaudeUsage{InputTokens: 120, OutputTokens: 8}
	(&GatewayService{}).parseSSEUsagePassthrough(`{"type":"message_delta","usage":{"output_tokens":100000001}}`, usage)
	if usage.InputTokens != 120 || usage.OutputTokens != 8 {
		t.Fatalf("malformed SSE usage changed prior valid usage: %+v", usage)
	}
}

func TestExtractSSEUsageRejectsInjectedValueWithoutLosingPriorUsage(t *testing.T) {
	usage := &ClaudeUsage{InputTokens: 120}
	(&AntigravityGatewayService{}).extractSSEUsage(
		`data: {"type":"message_delta","usage":{"output_tokens":-1}}`,
		usage,
	)
	if usage.InputTokens != 120 || usage.OutputTokens != 0 {
		t.Fatalf("malformed Antigravity usage changed prior valid usage: %+v", usage)
	}
}

func TestMergeAnthropicUsageRejectsInjectedValue(t *testing.T) {
	usage := &ClaudeUsage{InputTokens: 120}
	mergeAnthropicUsage(usage, apicompat.AnthropicUsage{InputTokens: 100000001, OutputTokens: 1})
	if usage.InputTokens != 120 || usage.OutputTokens != 0 {
		t.Fatalf("malformed native Anthropic usage was accepted: %+v", usage)
	}
}

func TestAddOpenAIUsageIgnoresInjectedValueWithoutLosingPriorUsage(t *testing.T) {
	usage := OpenAIUsage{InputTokens: 120, OutputTokens: 8}
	addOpenAIUsage(&usage, OpenAIUsage{InputTokens: maxReportedUsageTokens + 1, OutputTokens: 1})
	if usage.InputTokens != 120 || usage.OutputTokens != 8 {
		t.Fatalf("malformed bridge usage changed prior valid usage: %+v", usage)
	}
}

func TestAddOpenAIUsagePreservesPriorUsageOnAccumulationOverflow(t *testing.T) {
	usage := OpenAIUsage{InputTokens: maxReportedUsageTokens - 2, OutputTokens: 8}
	addOpenAIUsage(&usage, OpenAIUsage{InputTokens: 10, OutputTokens: 1})
	if usage.InputTokens != maxReportedUsageTokens-2 || usage.OutputTokens != 8 {
		t.Fatalf("overflowing usage snapshot must be ignored atomically: %+v", usage)
	}
}

func TestExtractOpenAIUsageRejectsMalformedHostedImageUsage(t *testing.T) {
	base := `"usage":{"input_tokens":100,"output_tokens":5}`
	tests := []struct {
		name  string
		image string
	}{
		{name: "oversized output", image: `{"output_tokens_details":{"image_tokens":100000001}}`},
		{name: "negative input", image: `{"input_tokens_details":{"image_tokens":-1}}`},
		{name: "fractional output", image: `{"output_tokens_details":{"image_tokens":1.5}}`},
		{name: "hostile exponent", image: `{"output_tokens_details":{"image_tokens":1e1000000000}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := []byte(`{` + base + `,"tool_usage":{"image_gen":` + tt.image + `}}`)
			usage, ok := extractOpenAIUsageFromJSONBytes(body)
			if !ok {
				t.Fatal("base usage should remain usable when nested tool usage is malformed")
			}
			if usage.InputTokens != 100 || usage.OutputTokens != 5 || usage.ImageInputTokens != 0 || usage.ImageOutputTokens != 0 {
				t.Fatalf("malformed nested usage changed billable usage: %+v", usage)
			}
		})
	}
}

func TestMergeHostedImageGenToolUsageNilAndBounded(t *testing.T) {
	mergeHostedImageGenToolUsage(gjson.Parse(`{"output_tokens_details":{"image_tokens":5}}`), nil)
	usage := OpenAIUsage{InputTokens: 10, OutputTokens: 2}
	mergeHostedImageGenToolUsage(gjson.Parse(`{"output_tokens_details":{"image_tokens":5},"input_tokens_details":{"image_tokens":3}}`), &usage)
	if usage.ImageOutputTokens != 5 || usage.ImageInputTokens != 3 {
		t.Fatalf("valid nested image usage was not merged: %+v", usage)
	}
}

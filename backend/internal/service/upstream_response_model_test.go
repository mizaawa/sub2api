//go:build unit

package service

import (
	"context"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type downstreamModelSettingRepoStub struct {
	value string
}

func (r *downstreamModelSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (r *downstreamModelSettingRepoStub) GetValue(context.Context, string) (string, error) {
	return r.value, nil
}

func (r *downstreamModelSettingRepoStub) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (r *downstreamModelSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (r *downstreamModelSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (r *downstreamModelSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (r *downstreamModelSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestDownstreamModelConsistencyBypassSettingDefaultsClosed(t *testing.T) {
	repo := &downstreamModelSettingRepoStub{}
	svc := &SettingService{settingRepo: repo}

	require.False(t, svc.IsDownstreamModelConsistencyBypassEnabled(context.Background()))
	repo.value = "true"
	require.True(t, svc.IsDownstreamModelConsistencyBypassEnabled(context.Background()))
	require.False(t, (*SettingService)(nil).IsDownstreamModelConsistencyBypassEnabled(context.Background()))
}

func TestRewriteDeclaredResponseModelsCoversSupportedProtocols(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		model   string
		rewrite func([]byte, string) []byte
		want    string
	}{
		{
			name:    "openai root and nested",
			input:   `{"model":"gpt-upstream","response":{"model":"gpt-versioned"}}`,
			model:   "gpt-public",
			rewrite: rewriteOpenAIResponseModel,
			want:    `{"model":"gpt-public","response":{"model":"gpt-public"}}`,
		},
		{
			name:    "anthropic root and message",
			input:   `{"model":"claude-upstream","message":{"model":"claude-versioned"}}`,
			model:   "claude-public",
			rewrite: rewriteAnthropicResponseModel,
			want:    `{"model":"claude-public","message":{"model":"claude-public"}}`,
		},
		{
			name:    "gemini wrappers",
			input:   `{"modelVersion":"gemini-upstream","response":{"modelVersion":"gemini-inner","response":{"modelVersion":"gemini-deep"}}}`,
			model:   "gemini-public",
			rewrite: rewriteGeminiResponseModel,
			want:    `{"modelVersion":"gemini-public","response":{"modelVersion":"gemini-public","response":{"modelVersion":"gemini-public"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.JSONEq(t, tt.want, string(tt.rewrite([]byte(tt.input), tt.model)))
		})
	}
}

func TestRewriteDeclaredResponseModelDoesNotInsertMissingFields(t *testing.T) {
	input := []byte(`{"type":"response.output_text.delta","delta":"hello"}`)
	require.Equal(t, input, rewriteOpenAIResponseModel(input, "gpt-public"))
}

func TestForcedModelRewriteLeavesRawAuditObservationIntact(t *testing.T) {
	raw := []byte(`{"type":"response.completed","response":{"model":"gpt-upstream-versioned"}}`)
	observer := &upstreamResponseModelObserver{}
	observer.ObserveOpenAI(raw, "response.completed")

	rewritten := rewriteOpenAIResponseModel(raw, "gpt-public")

	require.Equal(t, "gpt-upstream-versioned", observer.Model())
	require.Equal(t, "gpt-public", gjson.GetBytes(rewritten, "response.model").String())
}

func TestModelReplacementKeepsLegacyExactMatchUnlessForced(t *testing.T) {
	openAI := &OpenAIGatewayService{}
	body := []byte(`{"model":"gpt-upstream-versioned"}`)
	require.Equal(t, body, openAI.replaceModelInResponseBody(body, "gpt-upstream", "gpt-public"))
	require.Equal(t, "gpt-public", gjson.GetBytes(openAI.replaceModelInResponseBody(body, "gpt-upstream", "gpt-public", true), "model").String())

	line := `data: {"response":{"model":"gpt-upstream-versioned"}}`
	require.Equal(t, line, openAI.replaceModelInSSELine(line, "gpt-upstream", "gpt-public"))
	require.Equal(t, "gpt-public", gjson.Get(strings.TrimPrefix(openAI.replaceModelInSSELine(line, "gpt-upstream", "gpt-public", true), "data: "), "response.model").String())

	ws := []byte(`{"type":"response.completed","response":{"model":"gpt-upstream-versioned"}}`)
	require.Equal(t, ws, replaceOpenAIWSMessageModel(ws, "gpt-upstream", "gpt-public"))
	require.Equal(t, "gpt-public", gjson.GetBytes(replaceOpenAIWSMessageModel(ws, "gpt-upstream", "gpt-public", true), "response.model").String())

	anthropic := &GatewayService{}
	anthropicBody := []byte(`{"message":{"model":"claude-upstream-versioned"}}`)
	require.Equal(t, anthropicBody, anthropic.replaceModelInResponseBody(anthropicBody, "claude-upstream", "claude-public"))
	require.Equal(t, "claude-public", gjson.GetBytes(anthropic.replaceModelInResponseBody(anthropicBody, "claude-upstream", "claude-public", true), "message.model").String())
}

func TestUpstreamResponseModelObserverTerminalWinsAndRecordsConflict(t *testing.T) {
	observer := &upstreamResponseModelObserver{}

	observer.ObserveOpenAI([]byte(`{"type":"response.created","response":{"model":"gpt-5.5"}}`), "response.created")
	observer.ObserveOpenAI([]byte(`{"type":"response.completed","response":{"model":"gpt-5.4"}}`), "response.completed")

	require.Equal(t, "gpt-5.4", observer.Model())
	require.True(t, observer.Conflict())
}

func TestUpstreamResponseModelObserverSupportsAnthropicAndGeminiShapes(t *testing.T) {
	t.Run("anthropic", func(t *testing.T) {
		observer := &upstreamResponseModelObserver{}
		observer.ObserveAnthropic([]byte(`{"type":"message_start","message":{"model":"claude-sonnet-4-20250514"}}`))
		require.Equal(t, "claude-sonnet-4-20250514", observer.Model())
	})

	t.Run("gemini outer and nested", func(t *testing.T) {
		observer := &upstreamResponseModelObserver{}
		observer.ObserveGemini([]byte(`{"response":{"modelVersion":"gemini-2.5-pro"}}`))
		observer.ObserveGemini([]byte(`{"modelVersion":"gemini-2.5-pro-latest"}`))
		require.Equal(t, "gemini-2.5-pro-latest", observer.Model())
		require.True(t, observer.Conflict())
	})
}

func TestUpstreamResponseModelObservationAttemptReset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	first := beginUpstreamResponseModelObservation(c)
	first.Observe("failed-attempt-model", false)
	second := beginUpstreamResponseModelObservation(c)
	second.Observe("successful-attempt-model", false)

	require.Equal(t, "successful-attempt-model", observedUpstreamResponseModel(c))
	require.False(t, observedUpstreamResponseModelConflict(c))
}

func TestUpstreamModelMismatchThreeStateAndCaseInsensitiveComparison(t *testing.T) {
	require.Nil(t, upstreamModelMismatch("gpt-5.5", ""))

	matched := upstreamModelMismatch("gpt-5.5", "GPT-5.5")
	require.NotNil(t, matched)
	require.False(t, *matched)

	mismatched := upstreamModelMismatch("gpt-5.5", "gpt-5.4")
	require.NotNil(t, mismatched)
	require.True(t, *mismatched)
}

func TestObserveOpenAISSEBodyIgnoresMalformedPayload(t *testing.T) {
	observer := &upstreamResponseModelObserver{}
	observeOpenAISSEBody(observer, "data: not-json\n\ndata: {\"type\":\"response.completed\",\"response\":{\"model\":\"gpt-5.4\"}}\n\n")

	require.Equal(t, "gpt-5.4", observer.Model())
	require.False(t, observer.Conflict())
}

func TestObserveAntigravityGeminiSSELineReadsWrapperModelWithoutUnwrap(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
	}{
		{
			name:    "top-level sibling",
			payload: `{"modelVersion":"gemini-3-pro","response":{"candidates":[]}}`,
			want:    "gemini-3-pro",
		},
		{
			name:    "single wrapper",
			payload: `{"response":{"modelVersion":"gemini-3-pro","candidates":[]}}`,
			want:    "gemini-3-pro",
		},
		{
			name:    "nested response after one wrapper",
			payload: `{"response":{"response":{"modelVersion":"gemini-3-pro","candidates":[]}}}`,
			want:    "gemini-3-pro",
		},
		{
			name:    "outer declaration takes precedence",
			payload: `{"modelVersion":"gemini-outer","response":{"modelVersion":"gemini-inner","candidates":[]}}`,
			want:    "gemini-outer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(nil)
			beginUpstreamResponseModelObservation(c)

			svc := &AntigravityGatewayService{}
			svc.observeAntigravityGeminiSSELine(c, "data: "+tt.payload)

			require.Equal(t, tt.want, observedUpstreamResponseModel(c))
			require.False(t, observedUpstreamResponseModelConflict(c))
		})
	}
}

func TestUpstreamResponseModelObserverRejectsMalformedJSONWithModelField(t *testing.T) {
	observer := &upstreamResponseModelObserver{}
	observer.ObserveOpenAI([]byte(`{"response":{"model":"gpt-5.4"}`), "response.completed")

	require.Empty(t, observer.Model())
}

func TestUpstreamResponseModelObserverBoundsUntrustedModelName(t *testing.T) {
	observer := &upstreamResponseModelObserver{}
	observer.Observe("  "+strings.Repeat("模", upstreamResponseModelMaxLength+1)+"  ", false)

	require.Len(t, []rune(observer.Model()), upstreamResponseModelMaxLength)
}

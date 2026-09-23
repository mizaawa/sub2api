//go:build unit

package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseSeedanceRequestPreservesMultimodalMetadata(t *testing.T) {
	info, err := ParseSeedanceRequest([]byte(`{
		"model":"doubao-seedance-1-0-pro",
		"content":[{"type":"text","text":"make it cinematic"},{"type":"image_url","image_url":{"url":"https://img.example/input.png"}}],
		"duration":-1,"generate_audio":true,"future_field":{"keep":true}
	}`))
	require.NoError(t, err)
	require.Equal(t, "doubao-seedance-1-0-pro", info.Model)
	require.Equal(t, "make it cinematic", info.Prompt)
	require.Equal(t, []string{"https://img.example/input.png"}, info.InputImageURLs)
}

func TestForwardSeedanceCustomCreateRewritesOnlyModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"client-seedance","content":[{"type":"text","text":"hello"},{"type":"image_url","image_url":{"url":"https://img.example/input.png"}}],"duration":-1,"generate_audio":true,"future_field":{"keep":true}}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v3/contents/generations/tasks", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	account := customVideoTestAccount()
	account.Credentials["base_url"] = "https://ark.example.test"
	account.Credentials["model_mapping"] = map[string]any{"client-seedance": "doubao-seedance-1-0-pro"}
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"id":"task-1"}`)),
	}}
	svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardSeedance(context.Background(), c, account, SeedanceEndpointCreate, "", body)
	require.NoError(t, err)
	require.Equal(t, "https://ark.example.test/api/v3/contents/generations/tasks", upstream.lastReq.URL.String())
	require.Equal(t, http.MethodPost, upstream.lastReq.Method)
	require.Equal(t, "Bearer custom-secret", upstream.lastReq.Header.Get("Authorization"))
	require.JSONEq(t, `{"model":"doubao-seedance-1-0-pro","content":[{"type":"text","text":"hello"},{"type":"image_url","image_url":{"url":"https://img.example/input.png"}}],"duration":-1,"generate_audio":true,"future_field":{"keep":true}}`, string(upstream.lastBody))
	require.Equal(t, "seedance:task-1", result.ResponseID)
	require.Empty(t, recorder.Body.Bytes(), "Custom create is committed after task binding")
	require.True(t, svc.WriteBufferedGrokMediaResponse(c, result))
	require.JSONEq(t, `{"id":"task-1"}`, recorder.Body.String())
}

func TestForwardSeedanceStatusAndDeleteStripInternalTaskPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name   string
		method string
		endp   GrokMediaEndpoint
	}{
		{name: "status", method: http.MethodGet, endp: SeedanceEndpointStatus},
		{name: "delete", method: http.MethodDelete, endp: SeedanceEndpointDelete},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(tt.method, "/api/v3/contents/generations/tasks/task-1", nil)
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":"task-1","status":"succeeded","usage":{"completion_tokens":42}}`)),
			}}
			svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}
			result, err := svc.ForwardSeedance(context.Background(), c, customVideoTestAccount(), tt.endp, "seedance:task-1", nil)
			require.NoError(t, err)
			require.Equal(t, tt.method, upstream.lastReq.Method)
			require.Equal(t, "https://custom.example.test/proxy/v1/api/v3/contents/generations/tasks/task-1", upstream.lastReq.URL.String())
			require.Equal(t, "seedance:task-1", result.ResponseID)
			if tt.endp == SeedanceEndpointStatus {
				require.Equal(t, 42, result.Usage.OutputTokens)
			}
		})
	}
}

func TestForwardSeedanceStatusOnlyReportsCompletionTokensWhenSucceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, status := range []string{"queued", "running", "failed", "cancelled", "expired", "succeeded"} {
		t.Run(status, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v3/contents/generations/tasks/task-1", nil)
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":"task-1","status":"` + status + `","usage":{"completion_tokens":42}}`)),
			}}
			svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}
			result, err := svc.ForwardSeedance(context.Background(), c, customVideoTestAccount(), SeedanceEndpointStatus, "seedance:task-1", nil)
			require.NoError(t, err)
			if status == "succeeded" {
				require.Equal(t, 42, result.Usage.OutputTokens)
			} else {
				require.Zero(t, result.Usage.OutputTokens)
			}
		})
	}
}

func TestForwardSeedancePreservesUpstreamErrorsWithoutRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"video","content":[{"type":"text","text":"waves"}]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v3/contents/generations/tasks", bytes.NewReader(body))
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"error":{"code":"QuotaExceeded","message":"quota exhausted"}}`)),
	}}
	svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}
	_, err := svc.ForwardSeedance(context.Background(), c, customVideoTestAccount(), SeedanceEndpointCreate, "", body)
	require.Error(t, err)
	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	require.Contains(t, recorder.Body.String(), "QuotaExceeded")
	require.Len(t, upstream.requests, 1)
}

func TestSeedanceCapabilityAllowsConfiguredOpenAIAndAllCustomAPIKeys(t *testing.T) {
	openAI := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":             "sk-ark",
			"base_url":            "https://ark.example.test",
			"openai_capabilities": []string{"seedance"},
		},
	}
	require.True(t, openAI.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilitySeedance))
	delete(openAI.Credentials, "openai_capabilities")
	require.False(t, openAI.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilitySeedance))
	require.True(t, customVideoTestAccount().SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilitySeedance))
}

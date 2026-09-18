//go:build unit

package service

import (
	"context"
	"testing"
)

func TestCustomMonitorSupportsOpenAICompatibleAPIModes(t *testing.T) {
	if err := validateProvider(MonitorProviderCustom); err != nil {
		t.Fatalf("custom provider should be supported: %v", err)
	}
	for _, mode := range []string{MonitorAPIModeChatCompletions, MonitorAPIModeResponses} {
		if err := validateAPIMode(MonitorProviderCustom, mode); err != nil {
			t.Fatalf("custom %s mode should be valid: %v", mode, err)
		}
	}

	if err := validateTemplateCreateParams(ChannelMonitorRequestTemplateCreateParams{
		Name:     "custom responses",
		Provider: MonitorProviderCustom,
		APIMode:  MonitorAPIModeResponses,
	}); err != nil {
		t.Fatalf("custom responses template should be valid: %v", err)
	}
}

func TestRunCheckForModel_CustomChatCompletions(t *testing.T) {
	h := &openAICaptureHandler{}
	endpoint := setupFakeOpenAI(t, h)

	res := runCheckForModel(context.Background(), MonitorProviderCustom, endpoint, "custom-key", "custom-model", nil)

	if res.Status != MonitorStatusOperational {
		t.Fatalf("custom chat request should pass challenge, got status=%s message=%q", res.Status, res.Message)
	}
	if h.lastPath != providerOpenAIPath {
		t.Fatalf("expected chat completions path %q, got %q", providerOpenAIPath, h.lastPath)
	}
	if h.lastBody["model"] != "custom-model" {
		t.Fatalf("expected custom model in request body, got %v", h.lastBody["model"])
	}
	if h.lastHeaders.Get("Authorization") != "Bearer custom-key" {
		t.Fatalf("expected bearer auth header, got %q", h.lastHeaders.Get("Authorization"))
	}
}

func TestRunCheckForModel_CustomResponses(t *testing.T) {
	h := &openAICaptureHandler{responsesLeadingReasoning: true}
	endpoint := setupFakeOpenAI(t, h)

	res := runCheckForModel(context.Background(), MonitorProviderCustom, endpoint, "custom-key", "custom-model", &CheckOptions{
		APIMode: MonitorAPIModeResponses,
	})

	if res.Status != MonitorStatusOperational {
		t.Fatalf("custom responses request should pass challenge, got status=%s message=%q", res.Status, res.Message)
	}
	if h.lastPath != providerOpenAIResponsesPath {
		t.Fatalf("expected responses path %q, got %q", providerOpenAIResponsesPath, h.lastPath)
	}
	if _, ok := h.lastBody["instructions"]; !ok {
		t.Fatal("custom responses body should contain instructions")
	}
	if _, ok := h.lastBody["input"]; !ok {
		t.Fatal("custom responses body should contain input")
	}
}

func TestApplyMonitorUpdate_SwitchToCustomPreservesResponsesMode(t *testing.T) {
	custom := MonitorProviderCustom
	existing := &ChannelMonitor{
		Provider:        MonitorProviderOpenAI,
		APIMode:         MonitorAPIModeResponses,
		PrimaryModel:    "compatible-model",
		IntervalSeconds: 60,
	}

	if err := applyMonitorUpdate(existing, ChannelMonitorUpdateParams{Provider: &custom}); err != nil {
		t.Fatalf("switching an OpenAI responses monitor to custom failed: %v", err)
	}
	if existing.APIMode != MonitorAPIModeResponses {
		t.Fatalf("expected responses mode to be preserved, got %q", existing.APIMode)
	}
}

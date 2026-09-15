package service

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseOpenAIWSEventEnvelope(t *testing.T) {
	eventType, responseID, response := parseOpenAIWSEventEnvelope([]byte(`{"type":"response.completed","response":{"id":"resp_1","model":"gpt-5.1"}}`))
	require.Equal(t, "response.completed", eventType)
	require.Equal(t, "resp_1", responseID)
	require.True(t, response.Exists())
	require.Equal(t, `{"id":"resp_1","model":"gpt-5.1"}`, response.Raw)

	eventType, responseID, response = parseOpenAIWSEventEnvelope([]byte(`{"type":"response.delta","id":"evt_1"}`))
	require.Equal(t, "response.delta", eventType)
	require.Empty(t, responseID, "an event id must not become a response continuation id")
	require.False(t, response.Exists())

	eventType, responseID, response = parseOpenAIWSEventEnvelope([]byte(`{"type":"response.delta","id":"resp_2"}`))
	require.Equal(t, "response.delta", eventType)
	require.Equal(t, "resp_2", responseID)
	require.False(t, response.Exists())

	eventType, responseID, response = parseOpenAIWSEventEnvelope([]byte(`{"type":"response.output_text.delta","response_id":"resp_3","delta":"hello"}`))
	require.Equal(t, "response.output_text.delta", eventType)
	require.Equal(t, "resp_3", responseID)
	require.False(t, response.Exists())

	eventType, responseID, response = parseOpenAIWSEventEnvelope([]byte(`{"type":"response.output_text.delta","response_id":"evt_3","delta":"hello"}`))
	require.Equal(t, "response.output_text.delta", eventType)
	require.Empty(t, responseID, "a non-resp response_id must not become a response continuation id")
	require.False(t, response.Exists())

	eventType, responseID, response = parseOpenAIWSEventEnvelope([]byte(`{"type":"response.output_text.delta","response":{"id":"evt_nested"},"response_id":"resp_4"}`))
	require.Equal(t, "response.output_text.delta", eventType)
	require.Equal(t, "resp_4", responseID, "an invalid response.id must not mask a valid response_id")
	require.True(t, response.Exists())

	eventType, responseID, response = parseOpenAIWSEventEnvelope([]byte(`{"type":"response.output_text.delta","response":{"id":"resp_5"},"response_id":"evt_5","id":"evt_5_event"}`))
	require.Equal(t, "response.output_text.delta", eventType)
	require.Equal(t, "resp_5", responseID)
	require.True(t, response.Exists())

	eventType, responseID, response = parseOpenAIWSEventEnvelope([]byte(`{"type":"response.output_text.delta","response":{"id":"RESP_upper"}}`))
	require.Equal(t, "response.output_text.delta", eventType)
	require.Empty(t, responseID, "response IDs must use the canonical lowercase resp_ prefix")
	require.True(t, response.Exists())

	eventType, responseID, response = parseOpenAIWSEventEnvelope([]byte(`{"id":"evt_2"}`))
	require.Empty(t, eventType)
	require.Empty(t, responseID, "an untyped event id must not become a response continuation id")
	require.False(t, response.Exists())
}

func TestParseOpenAIWSResponseUsageFromCompletedEvent(t *testing.T) {
	usage := &OpenAIUsage{}
	parseOpenAIWSResponseUsageFromCompletedEvent(
		[]byte(`{"type":"response.completed","response":{"usage":{"input_tokens":11,"output_tokens":7,"input_tokens_details":{"cached_tokens":3}}}}`),
		usage,
	)
	require.Equal(t, 11, usage.InputTokens)
	require.Equal(t, 7, usage.OutputTokens)
	require.Equal(t, 3, usage.CacheReadInputTokens)

	parseOpenAIWSResponseUsageFromCompletedEvent(
		[]byte(`{"type":"response.completed","response":{"usage":{"prompt_tokens":19,"completion_tokens":5,"prompt_tokens_details":{"cached_tokens":4}}}}`),
		usage,
	)
	require.Equal(t, 19, usage.InputTokens)
	require.Equal(t, 5, usage.OutputTokens)
	require.Equal(t, 4, usage.CacheReadInputTokens)
}

func TestOpenAIWSEventShouldParseUsageTerminalEvents(t *testing.T) {
	t.Parallel()

	for _, eventType := range []string{
		"response.completed",
		"response.done",
		"response.failed",
		"response.incomplete",
		"response.cancelled",
		"response.canceled",
	} {
		require.True(t, openAIWSEventShouldParseUsage(eventType), eventType)
		require.True(t, openAIWSEventShouldParseUsage("  "+eventType+"  "), eventType)
	}
	require.False(t, openAIWSEventShouldParseUsage("response.output_text.delta"))
	require.False(t, openAIWSEventShouldParseUsage(""))
}

func TestOpenAIWSErrorEventHelpers_ConsistentWithWrapper(t *testing.T) {
	message := []byte(`{"type":"error","error":{"type":"invalid_request_error","code":"invalid_request","message":"invalid input"}}`)
	codeRaw, errTypeRaw, errMsgRaw := parseOpenAIWSErrorEventFields(message)

	wrappedReason, wrappedRecoverable := classifyOpenAIWSErrorEvent(message)
	rawReason, rawRecoverable := classifyOpenAIWSErrorEventFromRaw(codeRaw, errTypeRaw, errMsgRaw)
	require.Equal(t, wrappedReason, rawReason)
	require.Equal(t, wrappedRecoverable, rawRecoverable)

	wrappedStatus := openAIWSErrorHTTPStatus(message)
	rawStatus := openAIWSErrorHTTPStatusFromRaw(codeRaw, errTypeRaw)
	require.Equal(t, wrappedStatus, rawStatus)
	require.Equal(t, http.StatusBadRequest, rawStatus)

	wrappedCode, wrappedType, wrappedMsg := summarizeOpenAIWSErrorEventFields(message)
	rawCode, rawType, rawMsg := summarizeOpenAIWSErrorEventFieldsFromRaw(codeRaw, errTypeRaw, errMsgRaw)
	require.Equal(t, wrappedCode, rawCode)
	require.Equal(t, wrappedType, rawType)
	require.Equal(t, wrappedMsg, rawMsg)
}

func TestOpenAIWSMessageLikelyContainsToolCalls(t *testing.T) {
	require.False(t, openAIWSMessageLikelyContainsToolCalls([]byte(`{"type":"response.output_text.delta","delta":"hello"}`)))
	require.True(t, openAIWSMessageLikelyContainsToolCalls([]byte(`{"type":"response.output_item.added","item":{"tool_calls":[{"id":"tc1"}]}}`)))
	require.True(t, openAIWSMessageLikelyContainsToolCalls([]byte(`{"type":"response.output_item.added","item":{"type":"function_call"}}`)))
}

func TestReplaceOpenAIWSMessageModel_OptimizedStillCorrect(t *testing.T) {
	noModel := []byte(`{"type":"response.output_text.delta","delta":"hello"}`)
	require.Equal(t, string(noModel), string(replaceOpenAIWSMessageModel(noModel, "gpt-5.1", "custom-model")))

	rootOnly := []byte(`{"type":"response.created","model":"gpt-5.1"}`)
	require.Equal(t, `{"type":"response.created","model":"custom-model"}`, string(replaceOpenAIWSMessageModel(rootOnly, "gpt-5.1", "custom-model")))

	responseOnly := []byte(`{"type":"response.completed","response":{"model":"gpt-5.1"}}`)
	require.Equal(t, `{"type":"response.completed","response":{"model":"custom-model"}}`, string(replaceOpenAIWSMessageModel(responseOnly, "gpt-5.1", "custom-model")))

	both := []byte(`{"model":"gpt-5.1","response":{"model":"gpt-5.1"}}`)
	require.Equal(t, `{"model":"custom-model","response":{"model":"custom-model"}}`, string(replaceOpenAIWSMessageModel(both, "gpt-5.1", "custom-model")))
}

func TestRestoreOpenAIWSPassthroughResponseModel_RespectsAuditBypass(t *testing.T) {
	payload := []byte(`{"type":"response.completed","response":{"model":"mapped-model"}}`)

	require.Equal(t, string(payload), string(restoreOpenAIWSPassthroughResponseModel(payload, "requested-model", "mapped-model", false)))
	require.Equal(t, `{"type":"response.completed","response":{"model":"requested-model"}}`, string(restoreOpenAIWSPassthroughResponseModel(payload, "requested-model", "mapped-model", true)))
}

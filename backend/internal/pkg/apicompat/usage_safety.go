package apicompat

// Usage values originate in an upstream response and are therefore treated
// as untrusted billing input. Keep this limit aligned with the gateway
// service's boundary guard while keeping the protocol package dependency-free.
const maxUsageTokens = 100_000_000

func validUsageToken(value int) bool {
	return value >= 0 && value <= maxUsageTokens
}

func addUsageTokens(left, right int) (int, bool) {
	if !validUsageToken(left) || !validUsageToken(right) || left > maxUsageTokens-right {
		return 0, false
	}
	return left + right, true
}

// subtractUsageTokensClamped preserves the protocol's long-standing
// conversion semantics for inconsistent cache breakdowns: the uncached
// portion bottoms out at zero when cache tokens exceed the reported input.
// Every operand is still validated before arithmetic so malformed or
// unreasonably large upstream values are never allowed through.
func subtractUsageTokensClamped(value int, sub ...int) (int, bool) {
	if !validUsageToken(value) {
		return 0, false
	}
	result := value
	for _, part := range sub {
		if !validUsageToken(part) {
			return 0, false
		}
		if part >= result {
			result = 0
			continue
		}
		result -= part
	}
	return result, true
}

func validAnthropicUsage(usage AnthropicUsage) bool {
	return validUsageToken(usage.InputTokens) &&
		validUsageToken(usage.OutputTokens) &&
		validUsageToken(usage.CacheCreationInputTokens) &&
		validUsageToken(usage.CacheReadInputTokens)
}

func validResponsesInputDetails(details *ResponsesInputTokensDetails) bool {
	if details == nil {
		return true
	}
	return validUsageToken(details.CachedTokens) &&
		validUsageToken(details.AudioTokens) &&
		validUsageToken(details.CacheCreationTokens) &&
		validUsageToken(details.CacheWriteTokens)
}

func validResponsesOutputDetails(details *ResponsesOutputTokensDetails) bool {
	if details == nil {
		return true
	}
	return validUsageToken(details.ReasoningTokens) &&
		validUsageToken(details.AudioTokens) &&
		validUsageToken(details.AcceptedPredictionTokens) &&
		validUsageToken(details.RejectedPredictionTokens)
}

func validResponsesUsage(usage *ResponsesUsage) bool {
	if usage == nil {
		return true
	}
	return validUsageToken(usage.InputTokens) &&
		validUsageToken(usage.OutputTokens) &&
		validUsageToken(usage.TotalTokens) &&
		validUsageToken(usage.CacheCreationInputTokens) &&
		validResponsesInputDetails(usage.InputTokensDetails) &&
		validResponsesOutputDetails(usage.OutputTokensDetails)
}

func validChatTokenDetails(details *ChatTokenDetails) bool {
	if details == nil {
		return true
	}
	return validUsageToken(details.CachedTokens) &&
		validUsageToken(details.AudioTokens) &&
		validUsageToken(details.CacheCreationTokens) &&
		validUsageToken(details.CacheWriteTokens) &&
		validUsageToken(details.ReasoningTokens) &&
		validUsageToken(details.AcceptedPredictionTokens) &&
		validUsageToken(details.RejectedPredictionTokens)
}

func validChatUsage(usage *ChatUsage) bool {
	if usage == nil {
		return true
	}
	return validUsageToken(usage.PromptTokens) &&
		validUsageToken(usage.CompletionTokens) &&
		validUsageToken(usage.TotalTokens) &&
		validChatTokenDetails(usage.PromptTokensDetails) &&
		validChatTokenDetails(usage.CompletionTokensDetails)
}

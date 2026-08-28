package antigravity

// Usage metadata is supplied by an upstream provider and later feeds billing
// calculations in the gateway. Keep the adapter's own arithmetic bounded so a
// malformed value cannot wrap before the service-level sanitizer sees it.
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

func subtractUsageTokens(value, sub int) (int, bool) {
	if !validUsageToken(value) || !validUsageToken(sub) || sub > value {
		return 0, false
	}
	return value - sub, true
}

func validGeminiUsageMetadata(metadata *GeminiUsageMetadata) bool {
	if metadata == nil {
		return true
	}
	if !validUsageToken(metadata.PromptTokenCount) ||
		!validUsageToken(metadata.CandidatesTokenCount) ||
		!validUsageToken(metadata.CachedContentTokenCount) ||
		!validUsageToken(metadata.TotalTokenCount) ||
		!validUsageToken(metadata.ThoughtsTokenCount) {
		return false
	}
	for _, detail := range metadata.CandidatesTokensDetails {
		if !validUsageToken(detail.TokenCount) {
			return false
		}
	}
	for _, detail := range metadata.PromptTokensDetails {
		if !validUsageToken(detail.TokenCount) {
			return false
		}
	}
	return true
}

func sanitizeClaudeUsage(usage *ClaudeUsage) bool {
	if usage == nil {
		return false
	}
	if !validUsageToken(usage.InputTokens) ||
		!validUsageToken(usage.OutputTokens) ||
		!validUsageToken(usage.CacheCreationInputTokens) ||
		!validUsageToken(usage.CacheReadInputTokens) ||
		!validUsageToken(usage.ImageOutputTokens) {
		*usage = ClaudeUsage{}
		return false
	}
	return true
}

func usageFromGeminiMetadata(metadata *GeminiUsageMetadata) (ClaudeUsage, bool) {
	if metadata == nil || !validGeminiUsageMetadata(metadata) {
		return ClaudeUsage{}, metadata == nil
	}
	inputTokens, ok := subtractUsageTokens(metadata.PromptTokenCount, metadata.CachedContentTokenCount)
	if !ok {
		return ClaudeUsage{}, false
	}
	outputTokens, ok := addUsageTokens(metadata.CandidatesTokenCount, metadata.ThoughtsTokenCount)
	if !ok {
		return ClaudeUsage{}, false
	}
	imageTokens := metadata.ImageOutputTokens()
	if !validUsageToken(imageTokens) {
		return ClaudeUsage{}, false
	}
	return ClaudeUsage{
		InputTokens:          inputTokens,
		OutputTokens:         outputTokens,
		CacheReadInputTokens: metadata.CachedContentTokenCount,
		ImageOutputTokens:    imageTokens,
	}, true
}

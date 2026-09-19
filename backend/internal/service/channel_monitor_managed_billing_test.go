//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGatewayManagedMonitorUsageSkipsUserBillingButKeepsAccountQuota(t *testing.T) {
	usageRepo := &openAIRecordUsageBestEffortLogRepoStub{}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newGatewayRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, userRepo, subRepo)
	groupID := int64(71)

	err := svc.RecordUsage(context.Background(), &RecordUsageInput{
		Result: &ForwardResult{
			RequestID: "managed_gateway_usage",
			Usage: ClaudeUsage{
				InputTokens:  1000,
				OutputTokens: 500,
			},
			Model:    "claude-sonnet-4",
			Duration: time.Second,
		},
		APIKey: &APIKey{
			ID:          501,
			Purpose:     APIKeyPurposeChannelMonitor,
			GroupID:     &groupID,
			Group:       &Group{ID: groupID, RateMultiplier: 1},
			Quota:       100,
			RateLimit5h: 100,
			RateLimit1d: 100,
			RateLimit7d: 100,
		},
		User: &User{ID: 601},
		Account: &Account{
			ID:       701,
			Type:     AccountTypeAPIKey,
			Platform: PlatformAnthropic,
			Extra:    map[string]any{"quota_limit": 100.0},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.Greater(t, usageRepo.lastLog.TotalCost, 0.0)
	require.Zero(t, usageRepo.lastLog.ActualCost)
	require.Zero(t, userRepo.deductCalls)
	require.NotNil(t, billingRepo.lastCmd)
	require.Zero(t, billingRepo.lastCmd.BalanceCost)
	require.Zero(t, billingRepo.lastCmd.SubscriptionCost)
	require.Zero(t, billingRepo.lastCmd.APIKeyQuotaCost)
	require.Zero(t, billingRepo.lastCmd.APIKeyRateLimitCost)
	require.Greater(t, billingRepo.lastCmd.AccountQuotaCost, 0.0)
}

func TestOpenAIManagedMonitorUsageSkipsUserBillingButKeepsAccountQuota(t *testing.T) {
	usageRepo := &openAIRecordUsageBestEffortLogRepoStub{}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, userRepo, subRepo, nil)
	groupID := int64(72)

	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID: "managed_openai_usage",
			Usage: OpenAIUsage{
				InputTokens:  1000,
				OutputTokens: 500,
			},
			Model:    "gpt-5.1",
			Duration: time.Second,
		},
		APIKey: &APIKey{
			ID:          502,
			Purpose:     APIKeyPurposeChannelMonitor,
			GroupID:     &groupID,
			Group:       &Group{ID: groupID, RateMultiplier: 1},
			Quota:       100,
			RateLimit5h: 100,
			RateLimit1d: 100,
			RateLimit7d: 100,
		},
		User: &User{ID: 602},
		Account: &Account{
			ID:       702,
			Type:     AccountTypeAPIKey,
			Platform: PlatformOpenAI,
			Extra:    map[string]any{"quota_limit": 100.0},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.Greater(t, usageRepo.lastLog.TotalCost, 0.0)
	require.Zero(t, usageRepo.lastLog.ActualCost)
	require.Zero(t, userRepo.deductCalls)
	require.NotNil(t, billingRepo.lastCmd)
	require.Zero(t, billingRepo.lastCmd.BalanceCost)
	require.Zero(t, billingRepo.lastCmd.SubscriptionCost)
	require.Zero(t, billingRepo.lastCmd.APIKeyQuotaCost)
	require.Zero(t, billingRepo.lastCmd.APIKeyRateLimitCost)
	require.Greater(t, billingRepo.lastCmd.AccountQuotaCost, 0.0)
}

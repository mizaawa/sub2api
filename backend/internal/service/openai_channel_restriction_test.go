//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestOpenAISelectAccountForModelWithExclusions_ChannelMappedRestrictionRejectsEarly(t *testing.T) {
	t.Parallel()

	channelSvc := newTestChannelService(makeStandardRepo(Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceChannelMapped,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{"gpt-4o"}},
		},
		ModelMapping: map[string]map[string]string{
			PlatformOpenAI: {"gpt-4.1": "o3-mini"},
		},
	}, map[int64]string{10: PlatformOpenAI}))

	svc := &OpenAIGatewayService{
		accountRepo: stubOpenAIAccountRepo{accounts: []Account{
			{ID: 1, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true},
		}},
		channelService: channelSvc,
	}

	groupID := int64(10)
	_, err := svc.SelectAccountForModelWithExclusions(context.Background(), &groupID, "", "gpt-4.1", nil)
	require.ErrorIs(t, err, ErrNoAvailableAccounts)
	require.Contains(t, err.Error(), "channel pricing restriction")
}

func TestOpenAISelectAccountForModelWithExclusions_UpstreamRestrictionSkipsDisallowedAccount(t *testing.T) {
	t.Parallel()

	channelSvc := newTestChannelService(makeStandardRepo(Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceUpstream,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{"o3-mini"}},
		},
	}, map[int64]string{10: PlatformOpenAI}))

	svc := &OpenAIGatewayService{
		accountRepo: stubOpenAIAccountRepo{accounts: []Account{
			{
				ID:          1,
				Platform:    PlatformOpenAI,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    10,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gpt-4.1": "gpt-4o"},
				},
			},
			{
				ID:          2,
				Platform:    PlatformOpenAI,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    20,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gpt-4.1": "o3-mini"},
				},
			},
		}},
		channelService: channelSvc,
	}

	groupID := int64(10)
	account, err := svc.SelectAccountForModelWithExclusions(context.Background(), &groupID, "", "gpt-4.1", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, int64(2), account.ID)
}

func TestOpenAISelectAccountForModelWithExclusions_StickyRestrictedUpstreamFallsBack(t *testing.T) {
	t.Parallel()

	channelSvc := newTestChannelService(makeStandardRepo(Channel{
		ID:                 1,
		Status:             StatusActive,
		GroupIDs:           []int64{10},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceUpstream,
		ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{"o3-mini"}},
		},
	}, map[int64]string{10: PlatformOpenAI}))

	cache := &stubGatewayCache{
		sessionBindings: map[string]int64{"openai:sticky-session": 1},
	}
	svc := &OpenAIGatewayService{
		accountRepo: stubOpenAIAccountRepo{accounts: []Account{
			{
				ID:          1,
				Platform:    PlatformOpenAI,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    10,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gpt-4.1": "gpt-4o"},
				},
			},
			{
				ID:          2,
				Platform:    PlatformOpenAI,
				Status:      StatusActive,
				Schedulable: true,
				Priority:    20,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gpt-4.1": "o3-mini"},
				},
			},
		}},
		channelService: channelSvc,
		cache:          cache,
	}

	groupID := int64(10)
	account, err := svc.SelectAccountForModelWithExclusions(context.Background(), &groupID, "sticky-session", "gpt-4.1", nil)
	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, int64(2), account.ID)
	require.Equal(t, 1, cache.deletedSessions["openai:sticky-session"])
	require.Equal(t, int64(2), cache.sessionBindings["openai:sticky-session"])
}

func TestOpenAICustomSelectionDoesNotRejectUnpricedMonitorAlias(t *testing.T) {
	t.Parallel()

	groupID := int64(11)
	channel := Channel{
		ID:                 11,
		Status:             StatusActive,
		GroupIDs:           []int64{groupID},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceUpstream,
		// The monitor sends a public alias that is intentionally not listed
		// here. Custom upstreams own model validity and may resolve the alias
		// through the account's model_mapping.
		ModelPricing: []ChannelModelPricing{{Platform: PlatformCustom, Models: []string{"some-other-model"}}},
	}
	channelService := &ChannelService{}
	channelService.cache.Store(populateChannelCache(
		[]Channel{channel},
		map[int64]string{groupID: PlatformCustom},
	))

	account := Account{
		ID:          12,
		Platform:    PlatformCustom,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "custom-monitor-key",
			"base_url": "https://custom.example.test/v1",
			"model_mapping": map[string]any{
				"monitor-alias": "provider-model",
			},
		},
	}
	svc := &OpenAIGatewayService{
		accountRepo:    stubOpenAIAccountRepo{accounts: []Account{account}},
		channelService: channelService,
	}

	selection, err := svc.selectAccountWithLoadAwareness(
		context.Background(), &groupID, PlatformCustom, "", "monitor-alias", nil,
		false, "", false,
	)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, account.ID, selection.Account.ID)
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
}

func TestOpenAICustomAdvancedSelectionDoesNotRejectUnpricedMonitorAlias(t *testing.T) {
	groupID := int64(12)
	channel := Channel{
		ID:                 12,
		Status:             StatusActive,
		GroupIDs:           []int64{groupID},
		RestrictModels:     true,
		BillingModelSource: BillingModelSourceRequested,
		ModelPricing:       []ChannelModelPricing{{Platform: PlatformCustom, Models: []string{"provider-model"}}},
	}
	channelService := &ChannelService{}
	channelService.cache.Store(populateChannelCache(
		[]Channel{channel},
		map[int64]string{groupID: PlatformCustom},
	))

	account := Account{
		ID:          13,
		Platform:    PlatformCustom,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":       "custom-monitor-key",
			"base_url":      "https://custom.example.test/v1",
			"model_mapping": map[string]any{"monitor-alias": "provider-model"},
		},
	}
	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: []Account{account}},
		channelService:     channelService,
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
		cfg:                &config.Config{},
	}

	selection, _, err := svc.SelectAccountWithSchedulerForCapability(
		context.Background(), &groupID, "", "monitor-session", "monitor-alias", nil,
		OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityChatCompletions,
		false, false, true, PlatformCustom,
	)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, account.ID, selection.Account.ID)
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
}

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type countingCustomMappingAccountRepo struct {
	schedulerTestOpenAIAccountRepo
	getByIDsCalls int
}

func (r *countingCustomMappingAccountRepo) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	r.getByIDsCalls++
	return r.schedulerTestOpenAIAccountRepo.GetByIDs(ctx, ids)
}

func customMappingSchedulerAccount(id int64, mapping map[string]any) *Account {
	credentials := map[string]any{
		"api_key":  "custom-key",
		"base_url": "https://custom.example.test/v1",
	}
	if mapping != nil {
		credentials["model_mapping"] = mapping
	}
	return &Account{
		ID:          id,
		Platform:    PlatformCustom,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: credentials,
	}
}

func TestCustomMappingSchedulerPrefersMappedAccountBeforeTransparentFallback(t *testing.T) {
	mapped := customMappingSchedulerAccount(1, map[string]any{"public-a": "provider-b"})
	transparent := customMappingSchedulerAccount(2, nil)

	scheduler := &defaultOpenAIAccountScheduler{service: &OpenAIGatewayService{}}
	plan := scheduler.buildOpenAIAccountLoadPlan(context.Background(), OpenAIAccountScheduleRequest{
		Platform:       PlatformCustom,
		RequestedModel: "public-a",
	}, []*Account{transparent, mapped}, map[int64]*AccountLoadInfo{
		transparent.ID: {AccountID: transparent.ID, LoadRate: 0},
		mapped.ID:      {AccountID: mapped.ID, LoadRate: 100},
	})

	require.Len(t, plan.selectionOrder, 2)
	require.Equal(t, mapped.ID, plan.selectionOrder[0].account.ID)
	require.True(t, plan.selectionOrder[0].mappingPreferred)
	require.False(t, plan.selectionOrder[1].mappingPreferred)
}

func TestCustomMappingSchedulerKeepsTransparentFallbackWhenNoMappingMatches(t *testing.T) {
	first := customMappingSchedulerAccount(1, nil)
	second := customMappingSchedulerAccount(2, map[string]any{"other": "provider"})

	scheduler := &defaultOpenAIAccountScheduler{service: &OpenAIGatewayService{}}
	plan := scheduler.buildOpenAIAccountLoadPlan(context.Background(), OpenAIAccountScheduleRequest{
		Platform:       PlatformCustom,
		RequestedModel: "public-a",
	}, []*Account{first, second}, map[int64]*AccountLoadInfo{
		first.ID:  {AccountID: first.ID, LoadRate: 0},
		second.ID: {AccountID: second.ID, LoadRate: 100},
	})

	require.Len(t, plan.selectionOrder, 2)
	require.False(t, plan.selectionOrder[0].mappingPreferred)
	require.False(t, plan.selectionOrder[1].mappingPreferred)
}

func TestCustomMappingPreferenceDoesNotAffectOpenAIPlatform(t *testing.T) {
	mapped := customMappingSchedulerAccount(1, map[string]any{"public-a": "provider-b"})
	transparent := customMappingSchedulerAccount(2, nil)
	mapped.Platform = PlatformOpenAI
	transparent.Platform = PlatformOpenAI

	scheduler := &defaultOpenAIAccountScheduler{service: &OpenAIGatewayService{}}
	plan := scheduler.buildOpenAIAccountLoadPlan(context.Background(), OpenAIAccountScheduleRequest{
		Platform:       PlatformOpenAI,
		RequestedModel: "public-a",
	}, []*Account{transparent, mapped}, map[int64]*AccountLoadInfo{
		transparent.ID: {AccountID: transparent.ID, LoadRate: 0},
		mapped.ID:      {AccountID: mapped.ID, LoadRate: 100},
	})

	require.Len(t, plan.selectionOrder, 2)
	for _, candidate := range plan.selectionOrder {
		require.False(t, candidate.mappingPreferred)
	}
}

func TestCustomSnapshotRefreshCopiesRoutingCredentials(t *testing.T) {
	fresh := customMappingSchedulerAccount(11, map[string]any{"public-a": "provider-b"})
	fresh.Credentials["api_key"] = "fresh-key"
	fresh.Credentials["base_url"] = "https://fresh.example.test/v1"
	stale := *fresh
	stale.Credentials = map[string]any{"api_key": "old-key"}

	svc := &OpenAIGatewayService{
		accountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{*fresh}},
	}
	accounts := []Account{stale}
	svc.refreshCustomSchedulingMappings(context.Background(), accounts)

	require.Equal(t, "fresh-key", accounts[0].GetOpenAIApiKey())
	require.Equal(t, "https://fresh.example.test/v1", accounts[0].GetOpenAIBaseURL())
	require.Equal(t, "provider-b", accounts[0].GetMappedModel("public-a"))
}

func TestCustomSnapshotRefreshUsesOneBatchLookup(t *testing.T) {
	first := customMappingSchedulerAccount(101, map[string]any{"public-a": "provider-a"})
	second := customMappingSchedulerAccount(102, map[string]any{"public-a": "provider-b"})
	repo := &countingCustomMappingAccountRepo{
		schedulerTestOpenAIAccountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{*first, *second}},
	}
	svc := &OpenAIGatewayService{accountRepo: repo}
	accounts := []Account{*first, *second}

	svc.refreshCustomSchedulingMappings(context.Background(), accounts)

	require.Equal(t, 1, repo.getByIDsCalls)
	require.Equal(t, "provider-a", accounts[0].GetMappedModel("public-a"))
	require.Equal(t, "provider-b", accounts[1].GetMappedModel("public-a"))
}

func TestCustomSnapshotRefreshClearsRemovedMapping(t *testing.T) {
	fresh := customMappingSchedulerAccount(12, nil)
	stale := *fresh
	stale.Credentials = map[string]any{
		"api_key":       "old-key",
		"base_url":      "https://old.example.test/v1",
		"model_mapping": map[string]any{"public-a": "old-provider"},
	}

	svc := &OpenAIGatewayService{
		accountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{*fresh}},
	}
	accounts := []Account{stale}
	svc.refreshCustomSchedulingMappings(context.Background(), accounts)

	require.Equal(t, "public-a", accounts[0].GetMappedModel("public-a"))
	_, exists := accounts[0].Credentials["model_mapping"]
	require.False(t, exists)
}

func TestCustomHydrationPreservesSelectedRoutingCredentials(t *testing.T) {
	const accountID int64 = 13
	stale := customMappingSchedulerAccount(accountID, map[string]any{"public-a": "old-provider"})
	stale.Credentials["api_key"] = "old-key"
	stale.Credentials["base_url"] = "https://old.example.test/v1"
	selected := customMappingSchedulerAccount(accountID, map[string]any{"public-a": "provider-b"})
	selected.Credentials["api_key"] = "fresh-key"
	selected.Credentials["base_url"] = "https://fresh.example.test/v1"

	snapshot := &openAISnapshotCacheStub{accountsByID: map[int64]*Account{accountID: stale}}
	svc := &OpenAIGatewayService{schedulerSnapshot: &SchedulerSnapshotService{cache: snapshot}}
	hydrated, err := svc.hydrateSelectedAccount(context.Background(), selected)

	require.NoError(t, err)
	require.Equal(t, "provider-b", hydrated.GetMappedModel("public-a"))
	require.Equal(t, "fresh-key", hydrated.GetOpenAIApiKey())
	require.Equal(t, "https://fresh.example.test/v1", hydrated.GetOpenAIBaseURL())
}

func TestAdvancedCustomSchedulerSelectsAccountMappedToChannelTarget(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	defer resetOpenAIAdvancedSchedulerSettingCacheForTest()

	groupID := int64(14)
	transparent := customMappingSchedulerAccount(141, nil)
	transparent.Priority = 0
	transparent.GroupIDs = []int64{groupID}
	mapped := customMappingSchedulerAccount(142, map[string]any{"provider-b": "provider-c"})
	mapped.Priority = 10
	mapped.GroupIDs = []int64{groupID}

	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{transparent, mapped},
		accountsByID: map[int64]*Account{
			transparent.ID: transparent,
			mapped.ID:      mapped,
		},
	}
	cfg := newSchedulerTestOpenAIWSV2Config()
	cfg.Gateway.OpenAIWS.LBTopK = 1
	cfg.Gateway.OpenAIWS.SchedulerScoreWeights.Priority = 1
	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: []Account{*transparent, *mapped}},
		cfg:                cfg,
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
		schedulerSnapshot:  &SchedulerSnapshotService{cache: snapshotCache},
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
	}

	selection, _, err := svc.SelectAccountWithSchedulerForCapability(
		context.Background(), &groupID, "", "", "provider-b", nil,
		OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityChatCompletions,
		false, false, false, PlatformCustom,
	)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, mapped.ID, selection.Account.ID)
	require.Equal(t, "provider-c", selection.Account.GetMappedModel("provider-b"))
}

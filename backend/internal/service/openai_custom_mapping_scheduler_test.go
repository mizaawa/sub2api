package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func customMappingSchedulerTestAccount(id int64, mapping map[string]any) *Account {
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

func TestNormalizeOpenAICompatiblePlatformTreatsCompositeAsCustom(t *testing.T) {
	require.Equal(t, PlatformCustom, normalizeOpenAICompatiblePlatform(PlatformComposite))
	require.Equal(t, PlatformCustom, normalizeOpenAICompatiblePlatform(PlatformCustom))
	require.Equal(t, PlatformGrok, normalizeOpenAICompatiblePlatform(PlatformGrok))
	require.Equal(t, PlatformOpenAI, normalizeOpenAICompatiblePlatform(PlatformOpenAI))
}

func TestPrioritizeCustomMappedAccountsPrefersAliasMatch(t *testing.T) {
	mapped := customMappingSchedulerTestAccount(1, map[string]any{"alias-a": "provider-b"})
	transparent := customMappingSchedulerTestAccount(2, nil)
	// The transparent account is deliberately first to model a higher-priority
	// or fresher LRU candidate produced by an earlier sorting phase.
	accounts := []*Account{transparent, mapped}

	prioritizeCustomMappedAccounts(PlatformCustom, "alias-a", accounts)

	require.Equal(t, mapped.ID, accounts[0].ID)
	require.Equal(t, transparent.ID, accounts[1].ID)
}

func TestPrioritizeCustomMappedAccountsKeepsTransparentFallbackWhenNoMatch(t *testing.T) {
	first := customMappingSchedulerTestAccount(1, nil)
	second := customMappingSchedulerTestAccount(2, map[string]any{"other": "provider"})
	accounts := []*Account{first, second}

	prioritizeCustomMappedAccounts(PlatformCustom, "alias-a", accounts)

	require.Equal(t, first.ID, accounts[0].ID)
	require.Equal(t, second.ID, accounts[1].ID)
}

func TestAdvancedCustomMappingSelectionExhaustsMappedTierBeforeTransparent(t *testing.T) {
	mapped := customMappingSchedulerTestAccount(1, map[string]any{"alias-a": "provider-b"})
	transparent := customMappingSchedulerTestAccount(2, nil)
	service := &OpenAIGatewayService{}
	scheduler := &defaultOpenAIAccountScheduler{service: service}

	plan := scheduler.buildOpenAIAccountLoadPlan(context.Background(), OpenAIAccountScheduleRequest{
		Platform:       PlatformCustom,
		RequestedModel: "alias-a",
	}, []*Account{transparent, mapped}, map[int64]*AccountLoadInfo{
		transparent.ID: {AccountID: transparent.ID, LoadRate: 0},
		mapped.ID:      {AccountID: mapped.ID, LoadRate: 100},
	})

	require.Len(t, plan.selectionOrder, 2)
	require.Equal(t, mapped.ID, plan.selectionOrder[0].account.ID,
		"a mapped Custom account must be probed before an otherwise healthier transparent account")
	require.True(t, plan.selectionOrder[0].mappingPreferred)
	require.False(t, plan.selectionOrder[1].mappingPreferred)
}

func TestAdvancedCustomMappingSelectionLeavesOtherPlatformsScoreDriven(t *testing.T) {
	mapped := customMappingSchedulerTestAccount(1, map[string]any{"alias-a": "provider-b"})
	transparent := customMappingSchedulerTestAccount(2, nil)
	// This test exercises the OpenAI scheduler boundary, so the fixture
	// accounts must be OpenAI accounts. A pair of Custom accounts would be
	// rejected by the platform filter before score ordering is evaluated.
	mapped.Platform = PlatformOpenAI
	transparent.Platform = PlatformOpenAI
	scheduler := &defaultOpenAIAccountScheduler{service: &OpenAIGatewayService{}}

	plan := scheduler.buildOpenAIAccountLoadPlan(context.Background(), OpenAIAccountScheduleRequest{
		Platform:       PlatformOpenAI,
		RequestedModel: "alias-a",
	}, []*Account{transparent, mapped}, map[int64]*AccountLoadInfo{
		transparent.ID: {AccountID: transparent.ID, LoadRate: 0},
		mapped.ID:      {AccountID: mapped.ID, LoadRate: 100},
	})

	require.NotEmpty(t, plan.selectionOrder)
	for _, candidate := range plan.selectionOrder {
		require.False(t, candidate.mappingPreferred)
	}
}

func TestHydrateSelectedCustomAccountPreservesFreshModelMapping(t *testing.T) {
	const accountID int64 = 91
	snapshotCache := &openAISnapshotCacheStub{accountsByID: map[int64]*Account{
		accountID: customMappingSchedulerTestAccount(accountID, map[string]any{
			"public-a": "stale-provider-model",
		}),
	}}
	svc := &OpenAIGatewayService{
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
	}
	fresh := customMappingSchedulerTestAccount(accountID, map[string]any{
		"public-a": "provider-b",
	})

	hydrated, err := svc.hydrateSelectedAccount(context.Background(), fresh)
	require.NoError(t, err)
	require.NotNil(t, hydrated)
	require.Equal(t, "provider-b", hydrated.GetMappedModel("public-a"),
		"a stale scheduler snapshot must not overwrite the mapping that just passed selection")
}

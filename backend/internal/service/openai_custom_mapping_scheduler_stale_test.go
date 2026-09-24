package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// The scheduler snapshot intentionally lags the repository here. This is the
// state observed briefly after an account-level model_mapping edit, before the
// scheduler outbox worker publishes the new credentials.
func TestAdvancedCustomSchedulerRefreshesStaleSnapshotModelMapping(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	defer resetOpenAIAdvancedSchedulerSettingCacheForTest()

	groupID := int64(99101)
	transparent := customMappingSchedulerTestAccount(99111, nil)
	transparent.Priority = 0
	transparent.GroupIDs = []int64{groupID}
	mapped := customMappingSchedulerTestAccount(99112, map[string]any{"public-a": "provider-b"})
	mapped.Priority = 10
	mapped.GroupIDs = []int64{groupID}

	// Redis has the same candidates, but the mapped account's credentials are
	// stale and no longer contain the newly configured alias.
	staleTransparent := *transparent
	staleMapped := *mapped
	staleMapped.Credentials = map[string]any{
		"api_key":  "custom-key",
		"base_url": "https://custom.example.test/v1",
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{&staleTransparent, &staleMapped},
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
		context.Background(), &groupID, "", "", "public-a", nil,
		OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityChatCompletions,
		false, false, false, PlatformCustom,
	)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, mapped.ID, selection.Account.ID,
		"a stale snapshot must not make a transparent Custom account win over an account with a current alias mapping")
	require.Equal(t, "provider-b", selection.Account.GetMappedModel("public-a"))
}

func TestAdvancedCustomSchedulerRefreshesStaleStickySnapshotModelMapping(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	defer resetOpenAIAdvancedSchedulerSettingCacheForTest()

	groupID := int64(99104)
	transparent := customMappingSchedulerTestAccount(99141, nil)
	transparent.Priority = 0
	transparent.GroupIDs = []int64{groupID}
	mapped := customMappingSchedulerTestAccount(99142, map[string]any{"public-a": "provider-b"})
	mapped.Priority = 10
	mapped.GroupIDs = []int64{groupID}

	// Both the candidate list and the per-ID snapshot lookup are stale. This
	// is the path used by sticky/candidate revalidation in production; refreshing
	// only ListSchedulableAccounts is insufficient because the second lookup can
	// otherwise erase the mapping before forwarding.
	staleTransparent := *transparent
	staleMapped := *mapped
	staleMapped.Credentials = map[string]any{
		"api_key":  "custom-key",
		"base_url": "https://custom.example.test/v1",
	}
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{&staleTransparent, &staleMapped},
		accountsByID: map[int64]*Account{
			transparent.ID: &staleTransparent,
			mapped.ID:      &staleMapped,
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
		context.Background(), &groupID, "", "", "public-a", nil,
		OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityChatCompletions,
		false, false, false, PlatformCustom,
	)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, mapped.ID, selection.Account.ID)
	require.Equal(t, "provider-b", selection.Account.GetMappedModel("public-a"))
}

func TestAdvancedCustomSchedulerClearsRemovedSnapshotModelMapping(t *testing.T) {
	resetOpenAIAdvancedSchedulerSettingCacheForTest()
	defer resetOpenAIAdvancedSchedulerSettingCacheForTest()

	groupID := int64(99102)
	transparent := customMappingSchedulerTestAccount(99121, nil)
	transparent.Priority = 0
	transparent.GroupIDs = []int64{groupID}
	removed := customMappingSchedulerTestAccount(99122, map[string]any{"public-a": "old-provider"})
	removed.Priority = 10
	removed.GroupIDs = []int64{groupID}
	// The database row no longer has model_mapping, while Redis still does.
	freshRemoved := *removed
	freshRemoved.Credentials = map[string]any{
		"api_key":  "custom-key",
		"base_url": "https://custom.example.test/v1",
	}
	staleTransparent := *transparent
	staleRemoved := *removed
	snapshotCache := &openAISnapshotCacheStub{
		snapshotAccounts: []*Account{&staleTransparent, &staleRemoved},
		accountsByID: map[int64]*Account{
			transparent.ID: transparent,
			removed.ID:     &freshRemoved,
		},
	}

	svc := &OpenAIGatewayService{
		accountRepo:        schedulerTestOpenAIAccountRepo{accounts: []Account{*transparent, freshRemoved}},
		cfg:                newSchedulerTestOpenAIWSV2Config(),
		rateLimitService:   newOpenAIAdvancedSchedulerRateLimitService("true"),
		schedulerSnapshot:  &SchedulerSnapshotService{cache: snapshotCache},
		concurrencyService: NewConcurrencyService(schedulerTestConcurrencyCache{}),
	}
	svc.cfg.Gateway.OpenAIWS.LBTopK = 1
	svc.cfg.Gateway.OpenAIWS.SchedulerScoreWeights.Priority = 1

	selection, _, err := svc.SelectAccountWithSchedulerForCapability(
		context.Background(), &groupID, "", "", "public-a", nil,
		OpenAIUpstreamTransportAny, OpenAIEndpointCapabilityChatCompletions,
		false, false, false, PlatformCustom,
	)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, transparent.ID, selection.Account.ID,
		"a mapping removed in the database must not remain preferred from the stale snapshot")
}

func TestAdvancedCustomSchedulerRefreshesLegacyCustomRoutingCredentials(t *testing.T) {
	accountID := int64(99131)
	fresh := customMappingSchedulerTestAccount(accountID, map[string]any{
		"public-a": "provider-b",
	})
	// Snapshots written before Custom base_url was retained contain only the
	// API key. They must still become usable after the best-effort DB overlay.
	stale := *fresh
	stale.Credentials = map[string]any{"api_key": "custom-key"}

	svc := &OpenAIGatewayService{
		accountRepo: schedulerTestOpenAIAccountRepo{accounts: []Account{*fresh}},
	}
	accounts := []Account{stale}
	svc.refreshCustomSchedulingMappings(context.Background(), accounts)

	require.Equal(t, "https://custom.example.test/v1", accounts[0].GetOpenAIBaseURL())
	require.Equal(t, "provider-b", accounts[0].GetMappedModel("public-a"))
}

func TestAdvancedCustomHydrationClearsRemovedCustomMapping(t *testing.T) {
	// Selection can pass with the current database row while the final Redis
	// hydration still returns an older mapping. The removed alias must not be
	// resurrected into the account used for forwarding.
	groupID := int64(99103)
	selected := customMappingSchedulerTestAccount(99131, map[string]any{})
	selected.GroupIDs = []int64{groupID}
	stale := customMappingSchedulerTestAccount(99131, map[string]any{
		"public-a": "old-provider",
	})
	stale.GroupIDs = []int64{groupID}

	snapshotCache := &openAISnapshotCacheStub{
		accountsByID: map[int64]*Account{selected.ID: stale},
	}
	svc := &OpenAIGatewayService{
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
	}

	hydrated, err := svc.hydrateSelectedAccount(context.Background(), selected)
	require.NoError(t, err)
	require.NotNil(t, hydrated)
	require.Equal(t, "public-a", hydrated.GetMappedModel("public-a"))
}

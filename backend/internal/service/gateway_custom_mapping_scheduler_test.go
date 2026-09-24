package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func legacyCustomMappingTestAccount(id int64, priority int, mapping map[string]any) *Account {
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
		Priority:    priority,
		Status:      StatusActive,
		Schedulable: true,
		Credentials: credentials,
	}
}

func TestLegacyFilterByMinPriorityPrefersMappedCustomTier(t *testing.T) {
	transparent := accountWithLoad{
		account:  legacyCustomMappingTestAccount(1, 1, nil),
		loadInfo: &AccountLoadInfo{LoadRate: 0},
	}
	mapped := accountWithLoad{
		account:          legacyCustomMappingTestAccount(2, 9, map[string]any{"public-a": "provider-a"}),
		loadInfo:         &AccountLoadInfo{LoadRate: 90},
		mappingPreferred: true,
	}

	got := filterByMinPriority([]accountWithLoad{transparent, mapped})
	require.Len(t, got, 1)
	require.Equal(t, mapped.account.ID, got[0].account.ID)
}

func TestLegacyFilterByMinPriorityKeepsTransparentFallbackAfterMappedExhaustion(t *testing.T) {
	transparent := accountWithLoad{
		account:  legacyCustomMappingTestAccount(1, 1, nil),
		loadInfo: &AccountLoadInfo{LoadRate: 0},
	}
	mapped := accountWithLoad{
		account:          legacyCustomMappingTestAccount(2, 9, map[string]any{"public-a": "provider-a"}),
		loadInfo:         &AccountLoadInfo{LoadRate: 90},
		mappingPreferred: true,
	}

	first := filterByMinPriority([]accountWithLoad{transparent, mapped})
	require.Len(t, first, 1)

	second := filterByMinPriority([]accountWithLoad{transparent})
	require.Len(t, second, 1)
	require.Equal(t, transparent.account.ID, second[0].account.ID)
}

func TestCustomAccountMappingPreferenceSupportsCompositeRoute(t *testing.T) {
	account := legacyCustomMappingTestAccount(1, 1, map[string]any{"public-a": "provider-a"})
	require.True(t, customAccountMappingPreference(PlatformCustom, account, "public-a"))
	require.True(t, customAccountMappingPreference(PlatformComposite, account, "public-a"))
	require.False(t, customAccountMappingPreference(PlatformCustom, account, "other"))
}

func TestLegacyCustomSchedulerRefreshesStaleSnapshotModelMapping(t *testing.T) {
	groupID := int64(99201)
	transparent := legacyCustomMappingTestAccount(99211, 0, nil)
	transparent.GroupIDs = []int64{groupID}
	mapped := legacyCustomMappingTestAccount(99212, 10, map[string]any{"public-a": "provider-b"})
	mapped.GroupIDs = []int64{groupID}

	// The snapshot was published before the account edit and therefore has no
	// mapping on the otherwise eligible account.  The repository is current.
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
	repo := schedulerTestOpenAIAccountRepo{accounts: []Account{*transparent, *mapped}}
	svc := &GatewayService{
		accountRepo:       repo,
		cfg:               &config.Config{},
		schedulerSnapshot: &SchedulerSnapshotService{cache: snapshotCache},
	}

	selected, err := svc.selectAccountForModelWithPlatform(
		context.Background(), &groupID, "", "public-a", nil, PlatformCustom,
	)
	require.NoError(t, err)
	require.NotNil(t, selected)
	require.Equal(t, mapped.ID, selected.ID,
		"legacy Custom scheduling must refresh a stale snapshot before ranking mapped candidates")
	require.Equal(t, "provider-b", selected.GetMappedModel("public-a"))

	// Selection hydration reads the same stale Redis account.  The mapping that
	// passed selection must survive that final snapshot read as well.
	hydrated, err := svc.hydrateSelectedAccount(context.Background(), selected)
	require.NoError(t, err)
	require.Equal(t, "provider-b", hydrated.GetMappedModel("public-a"))
}

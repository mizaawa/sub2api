package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// A profit-control terminal refresh must not replace a freshly selected
// Custom account's routing credentials with a reduced/stale scheduler copy.
// The refreshed copy still supplies the latest scheduling metadata (including
// the rate multiplier), while model_mapping is part of the forwarding
// contract and must survive the replacement.
func TestProfitControlVetoLatestPreservesCustomModelMapping(t *testing.T) {
	selected := customMappingSchedulerTestAccount(99201, map[string]any{
		"deepseek-v4-pro-0813": "deepseek-v4",
	})
	rateMultiplier := 0.5
	selected.RateMultiplier = &rateMultiplier
	selected.UpdatedAt = time.Unix(200, 0)
	refreshed := *selected
	refreshed.Credentials = map[string]any{
		"api_key":  "custom-key",
		"base_url": "https://custom.example.test/v1",
	}
	// Equal timestamps exercise the replacement branch used by the terminal
	// refresh when Redis has not observed a newer account write yet.
	refreshed.UpdatedAt = selected.UpdatedAt

	snapshot := &SchedulerSnapshotService{cache: &openAISnapshotCacheStub{
		accountsByID: map[int64]*Account{selected.ID: &refreshed},
	}}
	gate := &openAIProfitControlGate{groupID: 1, platform: PlatformComposite, threshold: 1}
	ctx := context.WithValue(context.Background(), openAIProfitControlGateCtxKey{}, gate)

	got, vetoed, reason := profitControlVetoLatest(ctx, selected, snapshot)
	require.False(t, vetoed, reason)
	require.NotNil(t, got)
	require.Equal(t, "deepseek-v4", got.GetMappedModel("deepseek-v4-pro-0813"))
	require.Equal(t, "https://custom.example.test/v1", got.GetOpenAIBaseURL())
}

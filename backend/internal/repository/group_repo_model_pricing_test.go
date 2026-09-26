package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupRepository_ModelPricingPersistence(t *testing.T) {
	apiKeys, client := newAPIKeyRepoSQLite(t)
	repo := newGroupRepositoryWithSQL(client, apiKeys.sql)
	ctx := context.Background()
	input, output, zero, perRequest := 0.000002, 0.000006, 0.0, 0.15
	limit := 200000
	pricing := []service.ChannelModelPricing{{
		Platform:       service.PlatformGrok,
		Models:         []string{"grok4.7", "grok-4.7"},
		BillingMode:    service.BillingModeToken,
		InputPrice:     &input,
		OutputPrice:    &output,
		CacheReadPrice: &zero,
		Intervals: []service.PricingInterval{{
			MinTokens: 0, MaxTokens: &limit, InputPrice: &input, OutputPrice: &output,
		}},
	}, {
		Platform: service.PlatformGrok, Models: []string{"grok-image"},
		BillingMode: service.BillingModePerRequest, PerRequestPrice: &perRequest,
	}}
	group := &service.Group{
		Name: "grok-custom-pricing", Platform: service.PlatformGrok,
		Status: service.StatusActive, RateMultiplier: 1,
		SubscriptionType: service.SubscriptionTypeStandard,
		ModelPricing:     pricing,
	}
	require.NoError(t, repo.Create(ctx, group))

	stored, err := repo.GetByIDLite(ctx, group.ID)
	require.NoError(t, err)
	require.Equal(t, pricing, stored.ModelPricing)
	groups, _, err := repo.List(ctx, pagination.PaginationParams{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Equal(t, pricing, groups[0].ModelPricing)

	user := mustCreateAPIKeyRepoUser(t, ctx, client, "group-pricing@test.com")
	key := &service.APIKey{
		UserID: user.ID, Key: "sk-group-model-pricing", Name: "Group pricing",
		GroupID: &group.ID, Status: service.StatusActive,
	}
	require.NoError(t, apiKeys.Create(ctx, key))
	auth, err := apiKeys.GetByKeyForAuth(ctx, key.Key)
	require.NoError(t, err)
	require.NotNil(t, auth.Group)
	require.Equal(t, pricing, auth.Group.ModelPricing)

	updatedPrice := 0.000003
	stored.ModelPricing = []service.ChannelModelPricing{{
		Platform: service.PlatformGrok, Models: []string{"grok4.7"},
		BillingMode: service.BillingModeToken, InputPrice: &updatedPrice, OutputPrice: &zero,
	}}
	require.NoError(t, repo.Update(ctx, stored))
	auth, err = apiKeys.GetByKeyForAuth(ctx, key.Key)
	require.NoError(t, err)
	require.Equal(t, stored.ModelPricing, auth.Group.ModelPricing)

	stored.ModelPricing = []service.ChannelModelPricing{}
	require.NoError(t, repo.Update(ctx, stored))
	cleared, err := repo.GetByIDLite(ctx, group.ID)
	require.NoError(t, err)
	require.NotNil(t, cleared.ModelPricing)
	require.Empty(t, cleared.ModelPricing)

	stored.ModelPricing = nil
	require.NoError(t, repo.Update(ctx, stored))
	cleared, err = repo.GetByIDLite(ctx, group.ID)
	require.NoError(t, err)
	require.Nil(t, cleared.ModelPricing)
}

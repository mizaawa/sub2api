package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupModelPricingResolver_AllPlatformsWithoutBuiltInPrice(t *testing.T) {
	for _, platform := range []string{PlatformGrok, PlatformOpenAI, PlatformAnthropic, PlatformGemini, PlatformAntigravity, PlatformCustom, PlatformComposite} {
		t.Run(platform, func(t *testing.T) {
			pricingPlatform := platform
			if platform == PlatformComposite {
				pricingPlatform = PlatformCustom
			}
			inputPrice, outputPrice := 2e-6, 6e-6
			group := &Group{ID: 42, Platform: platform, ModelPricing: []ChannelModelPricing{{
				Platform: pricingPlatform, Models: []string{"grok-4.7"},
				InputPrice: &inputPrice, OutputPrice: &outputPrice,
			}}}
			billing := &BillingService{fallbackPrices: map[string]*ModelPricing{}}
			resolver := NewModelPricingResolver(nil, billing)
			resolved := resolver.Resolve(t.Context(), PricingInput{Model: "grok-4.7", GroupID: &group.ID, Group: group})
			require.Equal(t, PricingSourceGroup, resolved.Source)
			cost, err := billing.CalculateCostUnified(CostInput{
				Ctx: t.Context(), Model: "grok-4.7", GroupID: &group.ID, Group: group,
				Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 500}, RateMultiplier: 1.5, Resolver: resolver,
			})
			require.NoError(t, err)
			require.InDelta(t, 0.005, cost.TotalCost, 1e-12)
			require.InDelta(t, 0.0075, cost.ActualCost, 1e-12)
		})
	}
}

func TestGroupModelPricingResolver_PrecedenceAndFallback(t *testing.T) {
	groupPrice, channelPrice, defaultPrice := 2e-6, 4e-6, 8e-6
	group := &Group{ID: 42, Platform: PlatformGrok, ModelPricing: []ChannelModelPricing{{
		Platform: PlatformGrok, Models: []string{"grok-4.7"}, InputPrice: &groupPrice,
	}}}
	channels := &ChannelService{}
	channels.cache.Store(populateChannelCache([]Channel{{
		ID: 1, Status: StatusActive, GroupIDs: []int64{group.ID},
		ModelPricing: []ChannelModelPricing{{Platform: PlatformGrok, Models: []string{"grok-4.7", "channel-only"}, InputPrice: &channelPrice}},
	}}, map[int64]string{group.ID: group.Platform}))
	billing := &BillingService{pricingService: &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"grok-4.7": {InputCostPerToken: defaultPrice}, "default-only": {InputCostPerToken: defaultPrice},
	}}}
	resolver := NewModelPricingResolver(channels, billing)
	for _, tc := range []struct {
		model, source string
		price         float64
	}{
		{"grok-4.7", PricingSourceGroup, groupPrice},
		{"channel-only", PricingSourceChannel, channelPrice},
		{"default-only", PricingSourceLiteLLM, defaultPrice},
	} {
		t.Run(tc.model, func(t *testing.T) {
			resolved := resolver.Resolve(t.Context(), PricingInput{Model: tc.model, GroupID: &group.ID, Group: group})
			require.Equal(t, tc.source, resolved.Source)
			require.InDelta(t, tc.price, resolved.BasePricing.InputPricePerToken, 1e-12)
		})
	}
	otherGroupID := int64(43)
	resolved := resolver.Resolve(t.Context(), PricingInput{Model: "grok-4.7", GroupID: &otherGroupID, Group: group})
	require.Equal(t, PricingSourceLiteLLM, resolved.Source)
	require.InDelta(t, defaultPrice, resolved.BasePricing.InputPricePerToken, 1e-12)
}

func TestGroupModelPricing_Matching(t *testing.T) {
	wildcard, exact, wrongPlatform := 1.0, 2.0, 3.0
	group := &Group{Platform: PlatformGrok, ModelPricing: []ChannelModelPricing{
		{Platform: PlatformGrok, Models: []string{" grok-* "}, PerRequestPrice: &wildcard},
		{Platform: PlatformOpenAI, Models: []string{"grok-4.7"}, PerRequestPrice: &wrongPlatform},
		{Platform: PlatformGrok, Models: []string{" GROK-4.7 "}, PerRequestPrice: &exact},
	}}
	require.Equal(t, exact, *group.GetModelPricing(" grok-4.7 ").PerRequestPrice)
	require.Equal(t, wildcard, *group.GetModelPricing("GROK-NEW").PerRequestPrice)
	require.Nil(t, group.GetModelPricing("unconfigured"))
	matched := group.GetModelPricing("grok-4.7")
	matched.Models[0] = "changed"
	require.Equal(t, " GROK-4.7 ", group.ModelPricing[2].Models[0])
	group.Platform = PlatformAnthropic
	group.ModelPricing = []ChannelModelPricing{{Platform: PlatformAnthropic, Models: []string{"claude-sonnet-4.5"}}}
	require.NotNil(t, group.GetModelPricing("claude-sonnet-4-5"))
}

func TestGroupModelPricingResolver_PerRequestAndExplicitFree(t *testing.T) {
	for _, mode := range []BillingMode{BillingModeToken, BillingModePerRequest, BillingModeImage, BillingModeVideo} {
		t.Run(string(mode), func(t *testing.T) {
			price, zero := 0.25, 0.0
			group := &Group{ID: 42, Platform: PlatformGrok, ModelPricing: []ChannelModelPricing{{
				Platform: PlatformGrok, Models: []string{"grok-4.7"}, BillingMode: mode,
				InputPrice: &zero, OutputPrice: &zero, PerRequestPrice: &price,
			}}}
			billing := &BillingService{fallbackPrices: map[string]*ModelPricing{}}
			resolver := NewModelPricingResolver(nil, billing)
			cost, err := billing.CalculateCostUnified(CostInput{
				Ctx: t.Context(), Model: "grok-4.7", GroupID: &group.ID, Group: group,
				Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 500}, RequestCount: 2, RateMultiplier: 1, Resolver: resolver,
			})
			require.NoError(t, err)
			if mode == BillingModeToken {
				require.Zero(t, cost.ActualCost)
			} else {
				require.Equal(t, 0.5, cost.ActualCost)
			}
		})
	}
}

func TestAPIKeyAuthSnapshot_PreservesGroupModelPricing(t *testing.T) {
	price := 2e-6
	svc := &APIKeyService{}
	group := &Group{ID: 42, Platform: PlatformGrok, ModelPricing: []ChannelModelPricing{{
		Platform: PlatformGrok, Models: []string{"grok-4.7"}, InputPrice: &price,
	}}}
	key := &APIKey{ID: 1, GroupID: &group.ID, Group: group, User: &User{ID: 2}}
	snapshot := svc.snapshotFromAPIKey(context.Background(), key)
	data, err := json.Marshal(snapshot)
	require.NoError(t, err)
	var cached APIKeyAuthSnapshot
	require.NoError(t, json.Unmarshal(data, &cached))
	restored := svc.snapshotToAPIKey("test", &cached)
	require.Equal(t, group.ModelPricing, restored.Group.ModelPricing)
	require.NotNil(t, restored.Group.GetModelPricing("grok-4.7"))
	_, ok, err := svc.applyAuthCacheEntry("legacy", &APIKeyAuthCacheEntry{Snapshot: &APIKeyAuthSnapshot{Version: 20}})
	require.NoError(t, err)
	require.False(t, ok, "old caches without group pricing must be reloaded")
}

//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminGroupModelPricingCreateAllPlatforms(t *testing.T) {
	for _, platform := range []string{PlatformGrok, PlatformAnthropic, PlatformOpenAI, PlatformGemini, PlatformAntigravity, PlatformCustom, PlatformComposite} {
		t.Run(platform, func(t *testing.T) {
			inputPrice, outputPrice := 2e-6, 6e-6
			input := &CreateGroupInput{
				Name: "priced group", Platform: platform, RateMultiplier: 1,
				ModelPricing: []ChannelModelPricing{{
					Platform: PlatformAnthropic, Models: []string{" grok-4.7 "},
					InputPrice: &inputPrice, OutputPrice: &outputPrice,
				}},
			}
			repo := &groupRepoStubForAdmin{}
			svc := &adminServiceImpl{groupRepo: repo}
			group, err := svc.CreateGroup(context.Background(), input)
			require.NoError(t, err)
			require.Len(t, repo.created.ModelPricing, 1)
			pricing := group.ModelPricing[0]
			expectedPlatform := platform
			if platform == PlatformCustom || platform == PlatformComposite {
				expectedPlatform = PlatformCustom
			}
			require.Equal(t, expectedPlatform, pricing.Platform)
			require.Equal(t, []string{"grok-4.7"}, pricing.Models)
			require.Equal(t, BillingModeToken, pricing.BillingMode)
			require.Equal(t, inputPrice, *pricing.InputPrice)
			require.Equal(t, outputPrice, *pricing.OutputPrice)
			require.Equal(t, " grok-4.7 ", input.ModelPricing[0].Models[0])
		})
	}
}

func TestAdminGroupModelPricingUpdateReplaceOmitClear(t *testing.T) {
	inputPrice := 2e-6
	group := &Group{ID: 9, Platform: PlatformGrok, RateMultiplier: 1}
	repo := &groupRepoStubForAdmin{getByID: group}
	svc := &adminServiceImpl{groupRepo: repo}
	pricing := []ChannelModelPricing{{Models: []string{"grok-4.7"}, InputPrice: &inputPrice}}

	updated, err := svc.UpdateGroup(context.Background(), group.ID, &UpdateGroupInput{ModelPricing: &pricing})
	require.NoError(t, err)
	require.Len(t, updated.ModelPricing, 1)
	require.Equal(t, inputPrice, *repo.updated.ModelPricing[0].InputPrice)

	updated, err = svc.UpdateGroup(context.Background(), group.ID, &UpdateGroupInput{Name: "renamed"})
	require.NoError(t, err)
	require.Len(t, updated.ModelPricing, 1)
	require.Equal(t, inputPrice, *updated.ModelPricing[0].InputPrice)

	empty := []ChannelModelPricing{}
	updated, err = svc.UpdateGroup(context.Background(), group.ID, &UpdateGroupInput{ModelPricing: &empty})
	require.NoError(t, err)
	require.NotNil(t, updated.ModelPricing)
	require.Empty(t, updated.ModelPricing)
}

func TestAdminGroupModelPricingRejectsInvalidConfiguration(t *testing.T) {
	zero, negative := 0.0, -1.0
	for _, tc := range []struct {
		name    string
		pricing []ChannelModelPricing
	}{
		{"no model", []ChannelModelPricing{{InputPrice: &zero}}},
		{"blank model", []ChannelModelPricing{{Models: []string{"  "}, InputPrice: &zero}}},
		{"invalid wildcard", []ChannelModelPricing{{Models: []string{"grok*4.7"}, InputPrice: &zero}}},
		{"missing price", []ChannelModelPricing{{Models: []string{"grok-4.7"}}}},
		{"request price in token mode", []ChannelModelPricing{{Models: []string{"grok-4.7"}, PerRequestPrice: &zero}}},
		{"negative price", []ChannelModelPricing{{Models: []string{"grok-4.7"}, InputPrice: &negative}}},
		{"invalid mode", []ChannelModelPricing{{Models: []string{"grok-4.7"}, BillingMode: "other", InputPrice: &zero}}},
		{"duplicate model", []ChannelModelPricing{{Models: []string{"grok-4.7", " GROK-4.7 "}, InputPrice: &zero}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &groupRepoStubForAdmin{getByID: &Group{ID: 9, Platform: PlatformGrok, RateMultiplier: 1}}
			svc := &adminServiceImpl{groupRepo: repo}
			_, err := svc.CreateGroup(context.Background(), &CreateGroupInput{Name: "invalid", Platform: PlatformGrok, RateMultiplier: 1, ModelPricing: tc.pricing})
			require.Error(t, err)
			require.Nil(t, repo.created)
			_, err = svc.UpdateGroup(context.Background(), 9, &UpdateGroupInput{ModelPricing: &tc.pricing})
			require.Error(t, err)
			require.Nil(t, repo.updated)
		})
	}
}

func TestAdminGroupModelPricingPreservesExplicitZero(t *testing.T) {
	zero := 0.0
	pricing, err := normalizeGroupModelPricing(PlatformGrok, []ChannelModelPricing{{Models: []string{"grok-4.7"}, InputPrice: &zero, OutputPrice: &zero}})
	require.NoError(t, err)
	require.NotNil(t, pricing[0].InputPrice)
	require.NotNil(t, pricing[0].OutputPrice)
	require.Zero(t, *pricing[0].InputPrice)
	require.Zero(t, *pricing[0].OutputPrice)
}

func TestDuplicateGroupModelPricingIsIndependent(t *testing.T) {
	inputPrice, intervalPrice, maxTokens := 2e-6, 3e-6, 1000
	source := &Group{ModelPricing: []ChannelModelPricing{{
		Models: []string{"grok-4.7"}, InputPrice: &inputPrice,
		Intervals: []PricingInterval{{MaxTokens: &maxTokens, InputPrice: &intervalPrice}},
	}}}
	duplicate := cloneGroupForDuplicate(source, "operation")
	require.Equal(t, source.ModelPricing, duplicate.ModelPricing)
	duplicate.ModelPricing[0].Models[0] = "changed"
	*duplicate.ModelPricing[0].InputPrice = 10
	*duplicate.ModelPricing[0].Intervals[0].InputPrice = 20
	*duplicate.ModelPricing[0].Intervals[0].MaxTokens = 50
	require.Equal(t, "grok-4.7", source.ModelPricing[0].Models[0])
	require.Equal(t, 2e-6, *source.ModelPricing[0].InputPrice)
	require.Equal(t, 3e-6, *source.ModelPricing[0].Intervals[0].InputPrice)
	require.Equal(t, 1000, *source.ModelPricing[0].Intervals[0].MaxTokens)
}

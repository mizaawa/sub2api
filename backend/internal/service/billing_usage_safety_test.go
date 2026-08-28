//go:build unit

package service

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUsageTokensRejectsNegativeMaxIntAndCacheBreakdownOverflow(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*UsageTokens)
	}{
		{name: "negative input", mutate: func(tokens *UsageTokens) { tokens.InputTokens = -1 }},
		{name: "max int output", mutate: func(tokens *UsageTokens) { tokens.OutputTokens = math.MaxInt }},
		{name: "cache breakdown sum overflow", mutate: func(tokens *UsageTokens) {
			tokens.CacheCreation5mTokens = maxReportedUsageTokens
			tokens.CacheCreation1hTokens = 1
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := UsageTokens{}
			tt.mutate(&tokens)
			err := validateUsageTokens(tokens)
			require.Error(t, err)
			require.True(t, errors.Is(err, ErrInvalidUsageTokens), err)
		})
	}

	require.False(t, validUsageTokens(UsageTokens{InputTokens: math.MaxInt}))
	require.False(t, validUsageTokens(UsageTokens{InputTokens: math.MaxInt, OutputTokens: 1}))
	valid := UsageTokens{InputTokens: maxReportedUsageTokens, OutputTokens: maxReportedUsageTokens}
	require.NoError(t, validateUsageTokens(valid))
}

func TestBillingEntrypointsRejectInvalidUsageTokens(t *testing.T) {
	svc := newTestBillingService()
	invalid := UsageTokens{InputTokens: math.MaxInt}

	_, err := svc.CalculateCost("claude-sonnet-4", invalid, 1)
	require.ErrorIs(t, err, ErrInvalidUsageTokens)

	_, err = svc.CalculateCostWithLongContext("claude-sonnet-4", invalid, 1, 200000, 2)
	require.ErrorIs(t, err, ErrInvalidUsageTokens)

	_, err = svc.CalculateCostUnified(CostInput{
		Model:          "claude-sonnet-4",
		Tokens:         invalid,
		RateMultiplier: 1,
	})
	require.ErrorIs(t, err, ErrInvalidUsageTokens)
}

func TestCalculatePerRequestStatsCostNormalizesNonPositiveAndRejectsHugeCount(t *testing.T) {
	price := 0.05
	pricing := &ChannelModelPricing{PerRequestPrice: &price}

	for _, count := range []int{0, -1} {
		cost := calculatePerRequestStatsCost(pricing, count)
		require.NotNil(t, cost)
		require.InDelta(t, price, *cost, 1e-12)
	}
	require.Nil(t, calculatePerRequestStatsCost(pricing, math.MaxInt))
}

func TestAccountStatsCostRejectsInvalidUsageTokens(t *testing.T) {
	price := 0.001
	pricing := &ChannelModelPricing{
		BillingMode: BillingModeToken,
		InputPrice:  &price,
	}
	invalid := UsageTokens{CacheCreation5mTokens: maxReportedUsageTokens, CacheCreation1hTokens: 1}
	require.Nil(t, calculateStatsCost(pricing, invalid, 1))
	require.Nil(t, calculateTokenStatsCost(pricing, UsageTokens{OutputTokens: -1}))
}

func TestMediaBillingRejectsUnboundedCounts(t *testing.T) {
	svc := newTestBillingService()

	image := svc.CalculateImageCost("gemini-3-pro-image", "2K", math.MaxInt, nil, 1)
	require.Zero(t, image.TotalCost)
	require.Zero(t, image.ActualCost)

	video := svc.CalculateVideoCost("grok-imagine-video", "720p", math.MaxInt, 8, nil, 1)
	require.Zero(t, video.TotalCost)
	require.Zero(t, video.ActualCost)

	search := svc.CalculateWebSearchCost(math.MaxInt, nil, 1)
	require.Zero(t, search.TotalCost)
	require.Zero(t, search.ActualCost)
}

func TestUsageLogTotalTokensSaturatesAndRejectsNegativeRows(t *testing.T) {
	require.Zero(t, (*UsageLog)(nil).TotalTokens())
	require.Equal(t, math.MaxInt, (&UsageLog{InputTokens: math.MaxInt, OutputTokens: 1}).TotalTokens())
	require.Zero(t, (&UsageLog{InputTokens: 1, OutputTokens: -1}).TotalTokens())
}

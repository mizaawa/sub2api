package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildUserViewFromSummaryIncludesBoundGroupMultiplier(t *testing.T) {
	rate := 0.1
	groupID := int64(42)
	monitor := &ChannelMonitor{
		ID:                  7,
		Name:                "OpenAI",
		Provider:            MonitorProviderOpenAI,
		GroupID:             &groupID,
		GroupName:           "public",
		GroupRateMultiplier: rate,
		PrimaryModel:        "gpt-4o-mini",
	}

	view := buildUserViewFromSummary(monitor, MonitorStatusSummary{}, nil, nil)

	require.NotNil(t, view.GroupRateMultiplier)
	require.Equal(t, rate, *view.GroupRateMultiplier)
}

func TestBuildUserViewFromSummaryLeavesUnboundGroupMultiplierNil(t *testing.T) {
	monitor := &ChannelMonitor{ID: 7, GroupName: "", PrimaryModel: "gpt-4o-mini"}

	view := buildUserViewFromSummary(monitor, MonitorStatusSummary{}, nil, nil)

	require.Nil(t, view.GroupRateMultiplier)
}

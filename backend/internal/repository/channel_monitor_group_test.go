//go:build unit

package repository

import (
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/stretchr/testify/require"
)

func TestEntToServiceMonitorHydratesLiveGroupMetadata(t *testing.T) {
	groupID := int64(42)
	row := &dbent.ChannelMonitor{
		GroupID:   &groupID,
		GroupName: "legacy-group-name",
		Edges: dbent.ChannelMonitorEdges{
			Group: &dbent.Group{
				Name:           "OpenAI Standard",
				RateMultiplier: 0.1,
				Platform:       "openai",
			},
		},
	}

	monitor := entToServiceMonitor(row)

	require.NotNil(t, monitor.GroupID)
	require.Equal(t, groupID, *monitor.GroupID)
	require.NotSame(t, row.GroupID, monitor.GroupID)
	require.Equal(t, "OpenAI Standard", monitor.GroupName)
	require.Equal(t, 0.1, monitor.GroupRateMultiplier)
	require.Equal(t, "openai", monitor.GroupPlatform)
}

func TestEntToServiceMonitorPreservesLegacyMetadataWhenGroupIsUnavailable(t *testing.T) {
	groupID := int64(42)
	row := &dbent.ChannelMonitor{
		GroupID:   &groupID,
		GroupName: "legacy-group-name",
	}

	monitor := entToServiceMonitor(row)

	require.NotNil(t, monitor.GroupID)
	require.Equal(t, groupID, *monitor.GroupID)
	require.Equal(t, "legacy-group-name", monitor.GroupName)
	require.Zero(t, monitor.GroupRateMultiplier)
	require.Empty(t, monitor.GroupPlatform)
}

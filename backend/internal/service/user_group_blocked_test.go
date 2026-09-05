package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserCanBindGroupRejectsBlockedGroupBeforeVisibilityChecks(t *testing.T) {
	user := &User{BlockedGroups: []int64{7}}

	// A blocked public group is denied even though public groups normally do
	// not require an entry in AllowedGroups.
	require.False(t, user.CanBindGroup(7, false))
	// The same deny list also wins over an exclusive-group grant.
	user.AllowedGroups = []int64{7}
	require.False(t, user.CanBindGroup(7, true))
	// Other public groups retain the historical behavior.
	require.True(t, user.CanBindGroup(8, false))
}

func TestNormalizeBlockedGroupIDsDeduplicatesAndSorts(t *testing.T) {
	require.Equal(t, []int64{2, 5, 9}, normalizeBlockedGroupIDs([]int64{9, -1, 5, 2, 9, 0}))
	require.Equal(t, []int64{}, normalizeBlockedGroupIDs(nil))
}

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserCanBindGroupRejectsBlockedGroupBeforeVisibilityChecks(t *testing.T) {
	user := &User{BlockedGroups: []int64{7}}
	require.True(t, user.IsPublicGroupBlocked(7, false, SubscriptionTypeStandard))
	require.False(t, user.IsPublicGroupBlocked(7, true, SubscriptionTypeStandard))
	require.False(t, user.IsPublicGroupBlocked(7, false, SubscriptionTypeSubscription))

	// A blocked public group is denied even though public groups normally do
	// not require an entry in AllowedGroups.
	require.False(t, user.CanBindGroup(7, false))
	// The deny list is scoped to public standard groups; an exclusive group
	// that happens to reuse the ID remains governed by its allow list.
	user.AllowedGroups = []int64{7}
	require.True(t, user.CanBindGroup(7, true))
	require.True(t, user.CanBindGroupWithSubscriptionType(7, false, SubscriptionTypeSubscription))
	// Other public groups retain the historical behavior.
	require.True(t, user.CanBindGroup(8, false))
}

func TestNormalizeBlockedGroupIDsDeduplicatesAndSorts(t *testing.T) {
	require.Equal(t, []int64{2, 5, 9}, normalizeBlockedGroupIDs([]int64{9, -1, 5, 2, 9, 0}))
	require.Equal(t, []int64{}, normalizeBlockedGroupIDs(nil))
}

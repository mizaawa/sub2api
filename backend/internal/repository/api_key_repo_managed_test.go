package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyRepositoryHidesManagedKeysFromManagementQueries(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "managed-key-visibility@test.com")
	group, err := client.Group.Create().
		SetName("managed-key-group").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(1).
		Save(ctx)
	require.NoError(t, err)

	ordinary := &service.APIKey{
		UserID:  user.ID,
		Key:     "sk-ordinary-visible",
		Name:    "ordinary-visible",
		GroupID: &group.ID,
		Status:  service.StatusActive,
	}
	managed := &service.APIKey{
		UserID:  user.ID,
		Key:     "sk-managed-hidden",
		Name:    "managed-hidden",
		Purpose: service.APIKeyPurposeChannelMonitor,
		GroupID: &group.ID,
		Status:  service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, ordinary))
	require.NoError(t, repo.Create(ctx, managed))

	byID, err := repo.GetByID(ctx, managed.ID)
	require.NoError(t, err)
	require.Equal(t, service.APIKeyPurposeChannelMonitor, byID.Purpose)
	byKey, err := repo.GetByKey(ctx, managed.Key)
	require.NoError(t, err)
	require.Equal(t, service.APIKeyPurposeChannelMonitor, byKey.Purpose)
	forAuth, err := repo.GetByKeyForAuth(ctx, managed.Key)
	require.NoError(t, err)
	require.Equal(t, service.APIKeyPurposeChannelMonitor, forAuth.Purpose)

	page := pagination.PaginationParams{Page: 1, PageSize: 10}
	listed, result, err := repo.ListByUserID(ctx, user.ID, page, service.APIKeyListFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(1), result.Total)
	require.Equal(t, []int64{ordinary.ID}, apiKeyIDs(listed))

	all, err := repo.ListAllByUserID(ctx, user.ID, service.APIKeyListFilters{})
	require.NoError(t, err)
	require.Equal(t, []int64{ordinary.ID}, apiKeyIDs(all))

	owned, err := repo.VerifyOwnership(ctx, user.ID, []int64{ordinary.ID, managed.ID})
	require.NoError(t, err)
	require.Equal(t, []int64{ordinary.ID}, owned)

	count, err := repo.CountByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	found, err := repo.SearchAPIKeys(ctx, user.ID, "hidden", 10)
	require.NoError(t, err)
	require.Empty(t, found)

	groupKeys, groupPage, err := repo.ListByGroupID(ctx, group.ID, page)
	require.NoError(t, err)
	require.Equal(t, int64(1), groupPage.Total)
	require.Equal(t, []int64{ordinary.ID}, apiKeyIDs(groupKeys))
	groupCount, err := repo.CountByGroupID(ctx, group.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), groupCount)

	_, _, err = repo.GetKeyAndOwnerID(ctx, managed.ID)
	require.ErrorIs(t, err, service.ErrAPIKeyNotFound)
	exists, err := repo.ExistsByKey(ctx, managed.Key)
	require.NoError(t, err)
	require.True(t, exists)

	userKeys, err := repo.ListKeysByUserID(ctx, user.ID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{ordinary.Key, managed.Key}, userKeys)
	groupKeyValues, err := repo.ListKeysByGroupID(ctx, group.ID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{ordinary.Key, managed.Key}, groupKeyValues)

	replacement, err := client.Group.Create().
		SetName("managed-key-replacement-group").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(1).
		Save(ctx)
	require.NoError(t, err)
	updated, err := repo.UpdateGroupIDByUserAndGroup(ctx, user.ID, group.ID, replacement.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), updated)
	ordinaryAfterMove, err := repo.GetByID(ctx, ordinary.ID)
	require.NoError(t, err)
	require.Equal(t, replacement.ID, *ordinaryAfterMove.GroupID)
	managedAfterMove, err := repo.GetByID(ctx, managed.ID)
	require.NoError(t, err)
	require.Equal(t, group.ID, *managedAfterMove.GroupID)

	cleared, err := repo.ClearGroupIDByGroupID(ctx, group.ID)
	require.NoError(t, err)
	require.Zero(t, cleared)
	managedAfterClear, err := repo.GetByID(ctx, managed.ID)
	require.NoError(t, err)
	require.Equal(t, group.ID, *managedAfterClear.GroupID)
}

func apiKeyIDs(keys []service.APIKey) []int64 {
	ids := make([]int64, 0, len(keys))
	for i := range keys {
		ids = append(ids, keys[i].ID)
	}
	return ids
}

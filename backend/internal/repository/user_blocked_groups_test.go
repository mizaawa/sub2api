package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryBlockedGroupsRoundTripAndPublicFilterSQLite(t *testing.T) {
	repo, client := newUserEntRepo(t)
	ctx := context.Background()
	_, err := repo.sql.ExecContext(ctx, `
		CREATE TABLE user_blocked_groups (
			user_id INTEGER NOT NULL,
			group_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, group_id)
		)`)
	require.NoError(t, err)

	public, err := client.Group.Create().
		SetName("blocked-public").
		SetPlatform(service.PlatformOpenAI).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetIsExclusive(false).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	exclusive, err := client.Group.Create().
		SetName("blocked-exclusive").
		SetPlatform(service.PlatformOpenAI).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetIsExclusive(true).
		SetStatus(service.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	user := &service.User{
		Email:        "blocked-groups@example.com",
		PasswordHash: "test-password-hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, user))
	user.BlockedGroups = []int64{exclusive.ID, public.ID, public.ID}
	require.NoError(t, repo.Update(ctx, user, service.UserUpdateFields{BlockedGroups: true}))

	got, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, []int64{public.ID}, got.BlockedGroups)
}

//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// These small decorators keep the test focused on the two repository calls
// used by GetUserAllowedGroupIDSet without having to implement every method in
// the broad repository interfaces.
type plazaVisibilityUserRepo struct {
	UserRepository
	user *User
	err  error
}

func (r *plazaVisibilityUserRepo) GetByID(context.Context, int64) (*User, error) {
	return r.user, r.err
}

type plazaVisibilitySubscriptionRepo struct {
	UserSubscriptionRepository
	subscriptions []UserSubscription
	err           error
	calls         int
}

func (r *plazaVisibilitySubscriptionRepo) ListActiveByUserID(_ context.Context, userID int64) ([]UserSubscription, error) {
	r.calls++
	active := make([]UserSubscription, 0, len(r.subscriptions))
	for i := range r.subscriptions {
		sub := r.subscriptions[i]
		if sub.UserID == userID && sub.IsActive() {
			active = append(active, sub)
		}
	}
	return active, r.err
}

func TestGetUserAllowedGroupIDSetIncludesActiveSubscriptionGroups(t *testing.T) {
	now := time.Now()
	subRepo := &plazaVisibilitySubscriptionRepo{
		subscriptions: []UserSubscription{
			{UserID: 7, GroupID: 42, Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)},
			{UserID: 7, GroupID: 43, Status: SubscriptionStatusActive, ExpiresAt: now.Add(-time.Hour)},
			{UserID: 7, GroupID: 44, Status: SubscriptionStatusExpired, ExpiresAt: now.Add(time.Hour)},
			{UserID: 8, GroupID: 45, Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)},
		},
	}
	svc := &APIKeyService{
		userRepo:    &plazaVisibilityUserRepo{user: &User{ID: 7, AllowedGroups: []int64{9}}},
		userSubRepo: subRepo,
	}

	got, err := svc.GetUserAllowedGroupIDSet(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, map[int64]struct{}{9: {}, 42: {}}, got)
	require.Equal(t, 1, subRepo.calls)
}

func TestGetUserPlazaAllowedGroupIDSetRejectsSubscriptionAfterGroupTypeChange(t *testing.T) {
	now := time.Now()
	subRepo := &plazaVisibilitySubscriptionRepo{
		subscriptions: []UserSubscription{
			{UserID: 7, GroupID: 42, Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)},
			{UserID: 7, GroupID: 43, Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)},
		},
	}
	svc := &APIKeyService{
		userRepo:    &plazaVisibilityUserRepo{user: &User{ID: 7, AllowedGroups: []int64{9}}},
		userSubRepo: subRepo,
	}

	// Only group 42 is currently subscription-type. Group 43 represents a
	// formerly subscription group converted to standard+exclusive; its old
	// active subscription must not make it visible in the plaza.
	got, err := svc.GetUserPlazaAllowedGroupIDSet(
		context.Background(),
		7,
		map[int64]struct{}{42: {}},
	)
	require.NoError(t, err)
	require.Equal(t, map[int64]struct{}{9: {}, 42: {}}, got)
}

func TestGetUserPlazaAllowedGroupIDSetRequiresSubscriptionForCurrentSubscriptionGroup(t *testing.T) {
	now := time.Now()
	subRepo := &plazaVisibilitySubscriptionRepo{
		subscriptions: []UserSubscription{
			{UserID: 7, GroupID: 43, Status: SubscriptionStatusActive, ExpiresAt: now.Add(time.Hour)},
		},
	}
	svc := &APIKeyService{
		userRepo: &plazaVisibilityUserRepo{
			// 42 is a stale explicit grant; subscription groups must still use
			// the subscription entitlement as the source of truth.
			user: &User{ID: 7, AllowedGroups: []int64{42, 43, 44}},
		},
		userSubRepo: subRepo,
	}

	got, err := svc.GetUserPlazaAllowedGroupIDSet(
		context.Background(),
		7,
		map[int64]struct{}{42: {}, 43: {}},
	)
	require.NoError(t, err)
	require.Equal(t, map[int64]struct{}{43: {}, 44: {}}, got)
}

func TestGetUserAllowedGroupIDSetPropagatesAuthorizationDependencyErrors(t *testing.T) {
	failure := errors.New("repository unavailable")

	t.Run("user repository", func(t *testing.T) {
		subRepo := &plazaVisibilitySubscriptionRepo{}
		svc := &APIKeyService{
			userRepo:    &plazaVisibilityUserRepo{err: failure},
			userSubRepo: subRepo,
		}

		got, err := svc.GetUserAllowedGroupIDSet(context.Background(), 7)
		require.ErrorIs(t, err, failure)
		require.Nil(t, got)
		require.Zero(t, subRepo.calls)
	})

	t.Run("subscription repository", func(t *testing.T) {
		svc := &APIKeyService{
			userRepo:    &plazaVisibilityUserRepo{user: &User{ID: 7, AllowedGroups: []int64{9}}},
			userSubRepo: &plazaVisibilitySubscriptionRepo{err: failure},
		}

		got, err := svc.GetUserAllowedGroupIDSet(context.Background(), 7)
		require.ErrorIs(t, err, failure)
		require.Nil(t, got, "dependency failures must not degrade to a partial/anonymous view")
	})
}

func TestGetUserAllowedGroupIDSetKeepsExplicitGrantsWhenSubscriptionRepositoryMissing(t *testing.T) {
	svc := &APIKeyService{
		userRepo: &plazaVisibilityUserRepo{user: &User{ID: 7, AllowedGroups: []int64{9}}},
	}

	got, err := svc.GetUserAllowedGroupIDSet(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, map[int64]struct{}{9: {}}, got)
}

//go:build unit

package service

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

type channelMonitorAPIKeyRepoStub struct {
	APIKeyRepository
	created *APIKey
	deleted []int64
}

func (r *channelMonitorAPIKeyRepoStub) Create(_ context.Context, key *APIKey) error {
	key.ID = 41
	cloned := *key
	cloned.GroupID = cloneInt64Pointer(key.GroupID)
	r.created = &cloned
	return nil
}

func (r *channelMonitorAPIKeyRepoStub) GetByKey(_ context.Context, key string) (*APIKey, error) {
	if r.created == nil || r.created.Key != key {
		return nil, ErrAPIKeyNotFound
	}
	cloned := *r.created
	cloned.GroupID = cloneInt64Pointer(r.created.GroupID)
	return &cloned, nil
}

func (r *channelMonitorAPIKeyRepoStub) DeleteWithAudit(_ context.Context, id int64) error {
	r.deleted = append(r.deleted, id)
	return nil
}

type channelMonitorKeyUserRepoStub struct {
	UserRepository
	user *User
}

func (r *channelMonitorKeyUserRepoStub) GetByID(_ context.Context, id int64) (*User, error) {
	if r.user == nil || r.user.ID != id {
		return nil, ErrUserNotFound
	}
	cloned := *r.user
	return &cloned, nil
}

type channelMonitorKeyGroupRepoStub struct {
	GroupRepository
	group *Group
}

func (r *channelMonitorKeyGroupRepoStub) GetByIDLite(_ context.Context, id int64) (*Group, error) {
	if r.group == nil || r.group.ID != id {
		return nil, ErrGroupNotFound
	}
	cloned := *r.group
	return &cloned, nil
}

func TestAPIKeyServiceCreatesAndReclaimsChannelMonitorKey(t *testing.T) {
	repo := &channelMonitorAPIKeyRepoStub{}
	userRepo := &channelMonitorKeyUserRepoStub{user: &User{ID: 9, Status: StatusActive}}
	groupRepo := &channelMonitorKeyGroupRepoStub{group: &Group{ID: 7, Status: StatusActive}}
	svc := NewAPIKeyService(repo, userRepo, groupRepo, nil, nil, nil, nil)
	longUnsafeName := strings.Repeat("<", maxChannelMonitorNameRunes)

	key, err := svc.CreateChannelMonitorKey(context.Background(), 9, 7, longUnsafeName)

	require.NoError(t, err)
	require.NotNil(t, key)
	require.True(t, strings.HasPrefix(key.Key, "sk-"))
	require.Equal(t, 67, len(key.Key))
	require.Equal(t, int64(7), *key.GroupID)
	require.Equal(t, StatusAPIKeyActive, key.Status)
	require.Equal(t, APIKeyPurposeChannelMonitor, key.Purpose)
	require.True(t, strings.HasPrefix(key.Name, channelMonitorAPIKeyNamePrefix))
	require.NotContains(t, key.Name, "<")
	require.LessOrEqual(t, utf8.RuneCountInString(key.Name), maxChannelMonitorNameRunes)
	require.Equal(t, key.Name, repo.created.Name)

	// Cleanup relies on the immutable purpose, not the mutable display name.
	repo.created.Name = "renamed by ordinary key UI"
	err = svc.DeleteChannelMonitorKey(context.Background(), key.Key, 9)

	require.NoError(t, err)
	require.Equal(t, []int64{41}, repo.deleted)
}

func TestAPIKeyServiceRefusesToDeleteOrdinaryKeyWithSpoofedMonitorName(t *testing.T) {
	repo := &channelMonitorAPIKeyRepoStub{created: &APIKey{
		ID:     41,
		UserID: 9,
		Key:    "ordinary-user-key",
		Name:   channelMonitorAPIKeyNamePrefix + "spoofed",
	}}
	svc := NewAPIKeyService(repo, &channelMonitorKeyUserRepoStub{}, &channelMonitorKeyGroupRepoStub{}, nil, nil, nil, nil)

	err := svc.DeleteChannelMonitorKey(context.Background(), "ordinary-user-key", 9)

	require.ErrorIs(t, err, ErrInsufficientPerms)
	require.Empty(t, repo.deleted)
}

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type channelMonitorSortRepoStub struct {
	ChannelMonitorRepository
	updates []ChannelMonitorSortOrderUpdate
	err     error
}

func (r *channelMonitorSortRepoStub) UpdateSortOrders(_ context.Context, updates []ChannelMonitorSortOrderUpdate) error {
	r.updates = append([]ChannelMonitorSortOrderUpdate(nil), updates...)
	return r.err
}

func TestChannelMonitorServiceUpdateSortOrdersDelegatesToRepository(t *testing.T) {
	repo := &channelMonitorSortRepoStub{}
	svc := NewChannelMonitorService(repo, nil)
	want := []ChannelMonitorSortOrderUpdate{{ID: 9, SortOrder: 20}, {ID: 3, SortOrder: 10}}

	err := svc.UpdateSortOrders(context.Background(), want)

	require.NoError(t, err)
	require.Equal(t, want, repo.updates)
}

func TestChannelMonitorServiceUpdateSortOrdersPreservesRepositoryError(t *testing.T) {
	repo := &channelMonitorSortRepoStub{err: ErrChannelMonitorNotFound}
	svc := NewChannelMonitorService(repo, nil)

	err := svc.UpdateSortOrders(context.Background(), []ChannelMonitorSortOrderUpdate{{ID: 999, SortOrder: 1}})

	require.Error(t, err)
	require.ErrorIs(t, err, ErrChannelMonitorNotFound)
	require.True(t, errors.Is(err, ErrChannelMonitorNotFound))
}

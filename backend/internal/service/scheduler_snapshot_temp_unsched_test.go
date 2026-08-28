package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSchedulerSnapshotService_SkipsSnapshotCacheWhenTempUnschedDisabled(t *testing.T) {
	defer SetDisableTempUnschedulableRuntime(false)
	SetDisableTempUnschedulableRuntime(false)

	cache := &schedulerSnapshotCacheStub{
		snapshotHit: []*Account{
			{
				ID:          101,
				Platform:    PlatformOpenAI,
				Status:      StatusActive,
				Schedulable: true,
			},
		},
	}
	svc := &SchedulerSnapshotService{cache: cache}

	accounts, useMixed, err := svc.ListSchedulableAccounts(context.Background(), nil, PlatformOpenAI, false)
	require.NoError(t, err)
	require.False(t, useMixed)
	require.Len(t, accounts, 1)
	require.Equal(t, int64(101), accounts[0].ID)
	require.Zero(t, cache.captureCalls)

	SetDisableTempUnschedulableRuntime(true)

	accounts, useMixed, err = svc.ListSchedulableAccounts(context.Background(), nil, PlatformOpenAI, false)
	require.ErrorIs(t, err, ErrSchedulerCacheNotReady)
	require.False(t, useMixed)
	require.Nil(t, accounts)
	require.Equal(t, 1, cache.captureCalls)
}

func TestSchedulerSnapshotServiceFiltersStaleTransientAccountsFromCache(t *testing.T) {
	t.Cleanup(func() { SetDisableTempUnschedulableRuntime(false) })
	future := time.Now().Add(time.Hour)
	cache := &schedulerSnapshotCacheStub{
		snapshotHit: []*Account{
			{
				ID:                     201,
				Platform:               PlatformOpenAI,
				Status:                 StatusActive,
				Schedulable:            true,
				TempUnschedulableUntil: &future,
			},
			{
				ID:          202,
				Platform:    PlatformOpenAI,
				Status:      StatusActive,
				Schedulable: true,
			},
		},
	}
	svc := &SchedulerSnapshotService{cache: cache}

	// This models a snapshot published while the switch was enabled, then
	// read after an administrator turns the normal transient policy back on.
	SetDisableTempUnschedulableRuntime(false)
	accounts, useMixed, err := svc.ListSchedulableAccounts(context.Background(), nil, PlatformOpenAI, false)
	require.NoError(t, err)
	require.False(t, useMixed)
	require.Len(t, accounts, 1)
	require.Equal(t, int64(202), accounts[0].ID)
	require.Zero(t, cache.captureCalls)
}

type schedulerSnapshotCacheStub struct {
	snapshotHit    []*Account
	captureCalls   int
	lastSnapshot   []Account
	lastSnapshotOK bool
}

func (c *schedulerSnapshotCacheStub) GetSnapshot(context.Context, SchedulerBucket) ([]*Account, bool, error) {
	if c == nil || len(c.snapshotHit) == 0 {
		return nil, false, nil
	}
	return c.snapshotHit, true, nil
}

func (c *schedulerSnapshotCacheStub) CaptureBucketWriteToken(context.Context, SchedulerBucket) (SchedulerBucketWriteToken, error) {
	c.captureCalls++
	return SchedulerBucketWriteToken{Epoch: 1}, nil
}

func (c *schedulerSnapshotCacheStub) SetSnapshot(_ context.Context, _ SchedulerBucket, _ SchedulerBucketWriteToken, accounts []Account) error {
	c.lastSnapshot = append([]Account(nil), accounts...)
	c.lastSnapshotOK = true
	return nil
}

func (c *schedulerSnapshotCacheStub) RetireBucket(context.Context, SchedulerBucket) error {
	return nil
}

func (c *schedulerSnapshotCacheStub) ReopenBucket(context.Context, SchedulerBucket) (SchedulerBucketWriteToken, error) {
	return SchedulerBucketWriteToken{Epoch: 1}, nil
}

func (c *schedulerSnapshotCacheStub) TryAcquireGroupLifecycleLease(context.Context, int64, time.Duration) (SchedulerGroupLifecycleLease, bool, error) {
	return SchedulerGroupLifecycleLease{}, false, nil
}

func (c *schedulerSnapshotCacheStub) ReleaseGroupLifecycleLease(context.Context, SchedulerGroupLifecycleLease) error {
	return nil
}

func (c *schedulerSnapshotCacheStub) GetAccount(context.Context, int64) (*Account, error) {
	return nil, nil
}

func (c *schedulerSnapshotCacheStub) SetAccount(context.Context, *Account) error {
	return nil
}

func (c *schedulerSnapshotCacheStub) DeleteAccount(context.Context, int64) error {
	return nil
}

func (c *schedulerSnapshotCacheStub) UpdateLastUsed(context.Context, map[int64]time.Time) error {
	return nil
}

func (c *schedulerSnapshotCacheStub) TryLockBucket(context.Context, SchedulerBucket, time.Duration) (bool, error) {
	return false, nil
}

func (c *schedulerSnapshotCacheStub) UnlockBucket(context.Context, SchedulerBucket) error {
	return nil
}

func (c *schedulerSnapshotCacheStub) ListBuckets(context.Context) ([]SchedulerBucket, error) {
	return nil, nil
}

func (c *schedulerSnapshotCacheStub) GetOutboxWatermark(context.Context) (int64, error) {
	return 0, nil
}

func (c *schedulerSnapshotCacheStub) SetOutboxWatermark(context.Context, int64) error {
	return nil
}

var _ SchedulerCache = (*schedulerSnapshotCacheStub)(nil)

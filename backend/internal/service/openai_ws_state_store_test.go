package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAIWSStateStore_BindGetDeleteResponseAccount(t *testing.T) {
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	ctx := context.Background()
	groupID := int64(7)

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_abc", 101, time.Minute))

	accountID, err := store.GetResponseAccount(ctx, groupID, "resp_abc")
	require.NoError(t, err)
	require.Equal(t, int64(101), accountID)

	require.NoError(t, store.DeleteResponseAccount(ctx, groupID, "resp_abc"))
	accountID, err = store.GetResponseAccount(ctx, groupID, "resp_abc")
	require.NoError(t, err)
	require.Zero(t, accountID)
}

func TestOpenAIWSStateStore_HTTPResponseOwnerPersistsAcrossStoreInstances(t *testing.T) {
	cache := &stubGatewayCache{}
	ctx := context.Background()
	groupID := int64(8)
	writer := NewOpenAIWSStateStore(cache)

	require.NoError(t, writer.BindHTTPResponseOwner(ctx, groupID, "resp_owned", 201, 301, time.Minute))
	userID, apiKeyID, found, err := writer.GetHTTPResponseOwner(ctx, groupID, "resp_owned")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(201), userID)
	require.Equal(t, int64(301), apiKeyID)

	reader := NewOpenAIWSStateStore(cache)
	userID, apiKeyID, found, err = reader.GetHTTPResponseOwner(ctx, groupID, "resp_owned")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(201), userID)
	require.Zero(t, apiKeyID, "cross-instance ownership is intentionally keyed by user only")
}

type openAIHTTPResponseBindingTestValue struct {
	accountID int64
	userID    int64
}

type openAIHTTPResponseBindingTestCache struct {
	bindings       map[string]openAIHTTPResponseBindingTestValue
	legacyBindings map[string]int64
	setErr         error
	atomicSetCalls int
	atomicGetCalls int
	atomicDelCalls int
	legacySetCalls int
	legacyDelCalls []string
	setHasDeadline bool
	setDeadline    time.Duration
	atomicDelErr   error
	legacyDelErrs  map[string]error
}

func openAIHTTPResponseBindingTestKey(groupID int64, key string) string {
	return fmt.Sprintf("%d:%s", groupID, key)
}

func (c *openAIHTTPResponseBindingTestCache) GetSessionAccountID(_ context.Context, groupID int64, key string) (int64, error) {
	if value, ok := c.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, key)]; ok {
		return value, nil
	}
	return 0, ErrStickySessionNotFound
}

func (c *openAIHTTPResponseBindingTestCache) SetSessionAccountID(_ context.Context, groupID int64, key string, value int64, _ time.Duration) error {
	c.legacySetCalls++
	if c.legacyBindings == nil {
		c.legacyBindings = make(map[string]int64)
	}
	c.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, key)] = value
	return nil
}

func (c *openAIHTTPResponseBindingTestCache) RefreshSessionTTL(context.Context, int64, string, time.Duration) error {
	return nil
}

func (c *openAIHTTPResponseBindingTestCache) DeleteSessionAccountID(_ context.Context, groupID int64, key string) error {
	c.legacyDelCalls = append(c.legacyDelCalls, key)
	if err := c.legacyDelErrs[key]; err != nil {
		return err
	}
	delete(c.legacyBindings, openAIHTTPResponseBindingTestKey(groupID, key))
	return nil
}

func (c *openAIHTTPResponseBindingTestCache) SetOpenAIHTTPResponseBinding(
	ctx context.Context,
	groupID int64,
	key string,
	accountID, userID int64,
	_ time.Duration,
) error {
	c.atomicSetCalls++
	if deadline, ok := ctx.Deadline(); ok {
		c.setHasDeadline = true
		c.setDeadline = time.Until(deadline)
	}
	if c.setErr != nil {
		return c.setErr
	}
	if c.bindings == nil {
		c.bindings = make(map[string]openAIHTTPResponseBindingTestValue)
	}
	c.bindings[openAIHTTPResponseBindingTestKey(groupID, key)] = openAIHTTPResponseBindingTestValue{
		accountID: accountID,
		userID:    userID,
	}
	return nil
}

func (c *openAIHTTPResponseBindingTestCache) SetOpenAIHTTPResponseBindingWithLegacy(
	ctx context.Context,
	groupID int64,
	key, legacyAccountKey, legacyOwnerKey string,
	accountID, userID int64,
	ttl time.Duration,
) error {
	// Model the production transaction as one logical operation while retaining
	// the individual legacy maps for assertions about rolling-release mirrors.
	if err := c.setOpenAIHTTPResponseBinding(ctx, groupID, key, accountID, userID, ttl); err != nil {
		return err
	}
	if c.legacyBindings == nil {
		c.legacyBindings = make(map[string]int64)
	}
	c.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, legacyAccountKey)] = accountID
	c.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, legacyOwnerKey)] = userID
	return nil
}

func (c *openAIHTTPResponseBindingTestCache) setOpenAIHTTPResponseBinding(
	ctx context.Context,
	groupID int64,
	key string,
	accountID, userID int64,
	_ time.Duration,
) error {
	c.atomicSetCalls++
	if deadline, ok := ctx.Deadline(); ok {
		c.setHasDeadline = true
		c.setDeadline = time.Until(deadline)
	}
	if c.setErr != nil {
		return c.setErr
	}
	if c.bindings == nil {
		c.bindings = make(map[string]openAIHTTPResponseBindingTestValue)
	}
	c.bindings[openAIHTTPResponseBindingTestKey(groupID, key)] = openAIHTTPResponseBindingTestValue{
		accountID: accountID,
		userID:    userID,
	}
	return nil
}

func (c *openAIHTTPResponseBindingTestCache) GetOpenAIHTTPResponseBinding(
	_ context.Context,
	groupID int64,
	key string,
) (int64, int64, error) {
	c.atomicGetCalls++
	value, ok := c.bindings[openAIHTTPResponseBindingTestKey(groupID, key)]
	if !ok {
		return 0, 0, ErrStickySessionNotFound
	}
	return value.accountID, value.userID, nil
}

func (c *openAIHTTPResponseBindingTestCache) DeleteOpenAIHTTPResponseBinding(_ context.Context, groupID int64, key string) error {
	c.atomicDelCalls++
	if c.atomicDelErr != nil {
		return c.atomicDelErr
	}
	delete(c.bindings, openAIHTTPResponseBindingTestKey(groupID, key))
	return nil
}

func TestOpenAIWSStateStore_BindHTTPResponseUsesOneAtomicCacheWrite(t *testing.T) {
	cache := &openAIHTTPResponseBindingTestCache{}
	store := NewOpenAIWSStateStore(cache)
	ctx := context.Background()
	groupID := int64(18)

	require.NoError(t, store.BindHTTPResponse(ctx, groupID, "resp_atomic", 601, 701, 801, time.Minute))
	require.Equal(t, 1, cache.atomicSetCalls)
	require.Zero(t, cache.legacySetCalls, "atomic-capable caches must not receive legacy account/owner writes")
	require.True(t, cache.setHasDeadline)
	require.Greater(t, cache.setDeadline, 500*time.Millisecond)
	require.LessOrEqual(t, cache.setDeadline, openAIHTTPResponseBindRedisTimeout)

	accountID, err := store.GetResponseAccount(ctx, groupID, "resp_atomic")
	require.NoError(t, err)
	require.Equal(t, int64(601), accountID)
	userID, apiKeyID, found, err := store.GetHTTPResponseOwner(ctx, groupID, "resp_atomic")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(701), userID)
	require.Equal(t, int64(801), apiKeyID)
}

func TestOpenAIWSStateStore_HTTPResponseBindingPersistsAcrossStoreInstances(t *testing.T) {
	cache := &openAIHTTPResponseBindingTestCache{}
	ctx := context.Background()
	groupID := int64(19)
	writer := NewOpenAIWSStateStore(cache)
	require.NoError(t, writer.BindHTTPResponse(ctx, groupID, "resp_atomic_shared", 602, 702, 802, time.Minute))

	reader := NewOpenAIWSStateStore(cache)
	accountID, err := reader.GetResponseAccount(ctx, groupID, "resp_atomic_shared")
	require.NoError(t, err)
	require.Equal(t, int64(602), accountID)
	userID, apiKeyID, found, err := reader.GetHTTPResponseOwner(ctx, groupID, "resp_atomic_shared")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(702), userID)
	require.Zero(t, apiKeyID, "API key identity is intentionally process-local")
	require.Equal(t, 1, cache.atomicGetCalls, "account lookup must hydrate owner state from the same cache read")
}

func TestOpenAIWSStateStore_HTTPResponseOwnerLookupHydratesAccount(t *testing.T) {
	cache := &openAIHTTPResponseBindingTestCache{}
	ctx := context.Background()
	groupID := int64(22)
	writer := NewOpenAIWSStateStore(cache)
	require.NoError(t, writer.BindHTTPResponse(ctx, groupID, "resp_owner_first", 605, 705, 805, time.Minute))

	reader := NewOpenAIWSStateStore(cache)
	userID, apiKeyID, found, err := reader.GetHTTPResponseOwner(ctx, groupID, "resp_owner_first")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(705), userID)
	require.Zero(t, apiKeyID)
	accountID, err := reader.GetResponseAccount(ctx, groupID, "resp_owner_first")
	require.NoError(t, err)
	require.Equal(t, int64(605), accountID)
	require.Equal(t, 1, cache.atomicGetCalls, "owner lookup must hydrate account state from the same cache read")
}

func TestOpenAIWSStateStore_BindHTTPResponseFailureKeepsCoherentLocalState(t *testing.T) {
	cache := &openAIHTTPResponseBindingTestCache{setErr: errors.New("redis unavailable")}
	raw := NewOpenAIWSStateStore(cache)
	store, ok := raw.(*defaultOpenAIWSStateStore)
	require.True(t, ok)

	err := store.BindHTTPResponse(context.Background(), 20, "resp_atomic_failed", 603, 703, 803, time.Minute)
	require.Error(t, err)
	require.Equal(t, 1, cache.atomicSetCalls)
	require.Empty(t, cache.bindings)

	mapKey := openAIWSResponseAccountMapKey(20, "resp_atomic_failed")
	store.responseToAccountMu.RLock()
	_, hasAccount := store.responseToAccount[mapKey]
	store.responseToAccountMu.RUnlock()
	store.responseOwnerMu.RLock()
	_, hasOwner := store.responseOwners[mapKey]
	store.responseOwnerMu.RUnlock()
	require.True(t, hasAccount)
	require.True(t, hasOwner)

	accountID, getErr := store.GetResponseAccount(context.Background(), 20, "resp_atomic_failed")
	require.NoError(t, getErr)
	require.Equal(t, int64(603), accountID)
	userID, apiKeyID, found, getErr := store.GetHTTPResponseOwner(context.Background(), 20, "resp_atomic_failed")
	require.NoError(t, getErr)
	require.True(t, found)
	require.Equal(t, int64(703), userID)
	require.Equal(t, int64(803), apiKeyID)
}

type legacyHTTPResponseRollbackProbe struct {
	legacyBindings    map[string]int64
	setCalls          int
	secondSetContext  context.Context
	secondSetDeadline time.Time
	rollbackContext   context.Context
	rollbackDeadline  time.Time
	secondSetError    error
}

func (c *legacyHTTPResponseRollbackProbe) GetSessionAccountID(_ context.Context, groupID int64, key string) (int64, error) {
	if value, ok := c.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, key)]; ok {
		return value, nil
	}
	return 0, ErrStickySessionNotFound
}

func (c *legacyHTTPResponseRollbackProbe) SetSessionAccountID(ctx context.Context, groupID int64, key string, value int64, _ time.Duration) error {
	c.setCalls++
	if c.setCalls == 2 {
		c.secondSetContext = ctx
		if deadline, ok := ctx.Deadline(); ok {
			c.secondSetDeadline = deadline
		}
		if c.secondSetError != nil {
			return c.secondSetError
		}
	}
	if c.legacyBindings == nil {
		c.legacyBindings = make(map[string]int64)
	}
	c.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, key)] = value
	return nil
}

func (*legacyHTTPResponseRollbackProbe) RefreshSessionTTL(context.Context, int64, string, time.Duration) error {
	return nil
}

func (c *legacyHTTPResponseRollbackProbe) DeleteSessionAccountID(ctx context.Context, groupID int64, key string) error {
	c.rollbackContext = ctx
	if deadline, ok := ctx.Deadline(); ok {
		c.rollbackDeadline = deadline
	}
	delete(c.legacyBindings, openAIHTTPResponseBindingTestKey(groupID, key))
	return nil
}

func TestOpenAIWSStateStore_LegacyBindingRollbackUsesFreshTimeout(t *testing.T) {
	probe := &legacyHTTPResponseRollbackProbe{secondSetError: errors.New("owner write failed")}
	store := NewOpenAIWSStateStore(probe)
	err := store.BindHTTPResponse(context.Background(), 25, "resp_legacy_rollback", 607, 707, 807, time.Minute)
	require.ErrorIs(t, err, probe.secondSetError)
	require.Equal(t, 2, probe.setCalls)
	require.False(t, probe.secondSetDeadline.IsZero())
	require.False(t, probe.rollbackDeadline.IsZero())
	require.NotSame(t, probe.secondSetContext, probe.rollbackContext, "rollback must use a fresh context instead of the write context")
	require.Empty(t, probe.legacyBindings, "failed owner write must not leave an account-only legacy binding")
}

func TestOpenAIWSStateStore_CombinedCacheReadsLegacyResponseBindings(t *testing.T) {
	cache := &openAIHTTPResponseBindingTestCache{legacyBindings: make(map[string]int64)}
	groupID := int64(21)
	responseID := "resp_legacy_shared"
	cache.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, openAIWSResponseAccountCacheKey(responseID))] = 604
	cache.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, responseID))] = 704

	store := NewOpenAIWSStateStore(cache)
	accountID, err := store.GetResponseAccount(context.Background(), groupID, responseID)
	require.NoError(t, err)
	require.Equal(t, int64(604), accountID)
	userID, apiKeyID, found, err := store.GetHTTPResponseOwner(context.Background(), groupID, responseID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(704), userID)
	require.Zero(t, apiKeyID)
	require.Equal(t, 2, cache.atomicGetCalls, "new binding key should be checked before the legacy fallback")
}

func TestOpenAIWSStateStore_DeleteResponseAccountRemovesCombinedAndLegacyKeys(t *testing.T) {
	cache := &openAIHTTPResponseBindingTestCache{
		bindings:       make(map[string]openAIHTTPResponseBindingTestValue),
		legacyBindings: make(map[string]int64),
	}
	ctx := context.Background()
	groupID := int64(23)
	responseID := "resp_delete_all"
	combinedKey := openAIHTTPResponseBindingCacheKey(responseID)
	accountKey := openAIWSResponseAccountCacheKey(responseID)
	ownerKey := openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, responseID)
	cache.bindings[openAIHTTPResponseBindingTestKey(groupID, combinedKey)] = openAIHTTPResponseBindingTestValue{accountID: 606, userID: 706}
	cache.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, accountKey)] = 606
	cache.legacyBindings[openAIHTTPResponseBindingTestKey(groupID, ownerKey)] = 706

	store := NewOpenAIWSStateStore(cache)
	require.NoError(t, store.DeleteResponseAccount(ctx, groupID, responseID))
	require.Empty(t, cache.bindings)
	require.Empty(t, cache.legacyBindings)
	require.Equal(t, 1, cache.atomicDelCalls)
	require.ElementsMatch(t, []string{accountKey, ownerKey}, cache.legacyDelCalls)
}

func TestOpenAIWSStateStore_DeleteResponseAccountAggregatesCacheErrors(t *testing.T) {
	combinedErr := errors.New("combined delete failed")
	accountErr := errors.New("account delete failed")
	ownerErr := errors.New("owner delete failed")
	responseID := "resp_delete_errors"
	accountKey := openAIWSResponseAccountCacheKey(responseID)
	ownerKey := openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, responseID)
	cache := &openAIHTTPResponseBindingTestCache{
		atomicDelErr: combinedErr,
		legacyDelErrs: map[string]error{
			accountKey: accountErr,
			ownerKey:   ownerErr,
		},
	}

	err := NewOpenAIWSStateStore(cache).DeleteResponseAccount(context.Background(), 24, responseID)
	require.ErrorIs(t, err, combinedErr)
	require.ErrorIs(t, err, accountErr)
	require.ErrorIs(t, err, ownerErr)
	require.Equal(t, 1, cache.atomicDelCalls)
	require.ElementsMatch(t, []string{accountKey, ownerKey}, cache.legacyDelCalls)
}

func TestOpenAIWSStateStore_ResponseConnTTL(t *testing.T) {
	store := NewOpenAIWSStateStore(nil)
	store.BindResponseConn("resp_conn", "conn_1", 30*time.Millisecond)

	connID, ok := store.GetResponseConn("resp_conn")
	require.True(t, ok)
	require.Equal(t, "conn_1", connID)

	time.Sleep(60 * time.Millisecond)
	_, ok = store.GetResponseConn("resp_conn")
	require.False(t, ok)
}

func TestOpenAIWSStateStore_SessionTurnStateTTL(t *testing.T) {
	store := NewOpenAIWSStateStore(nil)
	store.BindSessionTurnState(9, "session_hash_1", "turn_state_1", 30*time.Millisecond)

	state, ok := store.GetSessionTurnState(9, "session_hash_1")
	require.True(t, ok)
	require.Equal(t, "turn_state_1", state)

	// group 隔离
	_, ok = store.GetSessionTurnState(10, "session_hash_1")
	require.False(t, ok)

	time.Sleep(60 * time.Millisecond)
	_, ok = store.GetSessionTurnState(9, "session_hash_1")
	require.False(t, ok)
}

func TestOpenAIWSStateStore_SessionConnTTL(t *testing.T) {
	store := NewOpenAIWSStateStore(nil)
	store.BindSessionConn(9, "session_hash_conn_1", "conn_1", 30*time.Millisecond)

	connID, ok := store.GetSessionConn(9, "session_hash_conn_1")
	require.True(t, ok)
	require.Equal(t, "conn_1", connID)

	// group 隔离
	_, ok = store.GetSessionConn(10, "session_hash_conn_1")
	require.False(t, ok)

	time.Sleep(60 * time.Millisecond)
	_, ok = store.GetSessionConn(9, "session_hash_conn_1")
	require.False(t, ok)
}

func TestOpenAIWSStateStore_GetResponseAccount_NoStaleAfterCacheMiss(t *testing.T) {
	cache := &stubGatewayCache{sessionBindings: map[string]int64{}}
	store := NewOpenAIWSStateStore(cache)
	ctx := context.Background()
	groupID := int64(17)
	responseID := "resp_cache_stale"
	cacheKey := openAIWSResponseAccountCacheKey(responseID)

	cache.sessionBindings[cacheKey] = 501
	accountID, err := store.GetResponseAccount(ctx, groupID, responseID)
	require.NoError(t, err)
	require.Equal(t, int64(501), accountID)

	delete(cache.sessionBindings, cacheKey)
	accountID, err = store.GetResponseAccount(ctx, groupID, responseID)
	require.NoError(t, err)
	require.Zero(t, accountID, "上游缓存失效后不应继续命中本地陈旧映射")
}

func TestOpenAIWSStateStore_MaybeCleanupRemovesExpiredIncrementally(t *testing.T) {
	raw := NewOpenAIWSStateStore(nil)
	store, ok := raw.(*defaultOpenAIWSStateStore)
	require.True(t, ok)

	expiredAt := time.Now().Add(-time.Minute)
	total := 2048
	store.responseToConnMu.Lock()
	for i := 0; i < total; i++ {
		store.responseToConn[fmt.Sprintf("resp_%d", i)] = openAIWSConnBinding{
			connID:    "conn_incremental",
			expiresAt: expiredAt,
		}
	}
	store.responseToConnMu.Unlock()

	store.lastCleanupUnixNano.Store(time.Now().Add(-2 * openAIWSStateStoreCleanupInterval).UnixNano())
	store.maybeCleanup()

	store.responseToConnMu.RLock()
	remainingAfterFirst := len(store.responseToConn)
	store.responseToConnMu.RUnlock()
	require.Less(t, remainingAfterFirst, total, "单轮 cleanup 应至少有进展")
	require.Greater(t, remainingAfterFirst, 0, "增量清理不要求单轮清空全部键")

	for i := 0; i < 8; i++ {
		store.lastCleanupUnixNano.Store(time.Now().Add(-2 * openAIWSStateStoreCleanupInterval).UnixNano())
		store.maybeCleanup()
	}

	store.responseToConnMu.RLock()
	remaining := len(store.responseToConn)
	store.responseToConnMu.RUnlock()
	require.Zero(t, remaining, "多轮 cleanup 后应逐步清空全部过期键")
}

func TestEnsureBindingCapacity_EvictsOneWhenMapIsFull(t *testing.T) {
	bindings := map[string]int{
		"a": 1,
		"b": 2,
	}

	ensureBindingCapacity(bindings, "c", 2)
	bindings["c"] = 3

	require.Len(t, bindings, 2)
	require.Equal(t, 3, bindings["c"])
}

func TestEnsureBindingCapacity_DoesNotEvictWhenUpdatingExistingKey(t *testing.T) {
	bindings := map[string]int{
		"a": 1,
		"b": 2,
	}

	ensureBindingCapacity(bindings, "a", 2)
	bindings["a"] = 9

	require.Len(t, bindings, 2)
	require.Equal(t, 9, bindings["a"])
}

type openAIWSStateStoreTimeoutProbeCache struct {
	setHasDeadline    bool
	getHasDeadline    bool
	deleteHasDeadline bool
	setDeadlineDelta  time.Duration
	getDeadlineDelta  time.Duration
	delDeadlineDelta  time.Duration
}

func (c *openAIWSStateStoreTimeoutProbeCache) GetSessionAccountID(ctx context.Context, _ int64, _ string) (int64, error) {
	if deadline, ok := ctx.Deadline(); ok {
		c.getHasDeadline = true
		c.getDeadlineDelta = time.Until(deadline)
	}
	return 123, nil
}

func (c *openAIWSStateStoreTimeoutProbeCache) SetSessionAccountID(ctx context.Context, _ int64, _ string, _ int64, _ time.Duration) error {
	if deadline, ok := ctx.Deadline(); ok {
		c.setHasDeadline = true
		c.setDeadlineDelta = time.Until(deadline)
	}
	return errors.New("set failed")
}

func (c *openAIWSStateStoreTimeoutProbeCache) RefreshSessionTTL(context.Context, int64, string, time.Duration) error {
	return nil
}

func (c *openAIWSStateStoreTimeoutProbeCache) DeleteSessionAccountID(ctx context.Context, _ int64, _ string) error {
	if deadline, ok := ctx.Deadline(); ok {
		c.deleteHasDeadline = true
		c.delDeadlineDelta = time.Until(deadline)
	}
	return nil
}

func TestOpenAIWSStateStore_RedisOpsUseShortTimeout(t *testing.T) {
	probe := &openAIWSStateStoreTimeoutProbeCache{}
	store := NewOpenAIWSStateStore(probe)
	ctx := context.Background()
	groupID := int64(5)

	err := store.BindResponseAccount(ctx, groupID, "resp_timeout_probe", 11, time.Minute)
	require.Error(t, err)

	accountID, getErr := store.GetResponseAccount(ctx, groupID, "resp_timeout_probe")
	require.NoError(t, getErr)
	require.Equal(t, int64(11), accountID, "本地缓存命中应优先返回已绑定账号")

	require.NoError(t, store.DeleteResponseAccount(ctx, groupID, "resp_timeout_probe"))

	require.True(t, probe.setHasDeadline, "SetSessionAccountID 应携带独立超时上下文")
	require.True(t, probe.deleteHasDeadline, "DeleteSessionAccountID 应携带独立超时上下文")
	require.False(t, probe.getHasDeadline, "GetSessionAccountID 本用例应由本地缓存命中，不触发 Redis 读取")
	require.Greater(t, probe.setDeadlineDelta, 2*time.Second)
	require.LessOrEqual(t, probe.setDeadlineDelta, 3*time.Second)
	require.Greater(t, probe.delDeadlineDelta, 2*time.Second)
	require.LessOrEqual(t, probe.delDeadlineDelta, 3*time.Second)

	probe2 := &openAIWSStateStoreTimeoutProbeCache{}
	store2 := NewOpenAIWSStateStore(probe2)
	accountID2, err2 := store2.GetResponseAccount(ctx, groupID, "resp_cache_only")
	require.NoError(t, err2)
	require.Equal(t, int64(123), accountID2)
	require.True(t, probe2.getHasDeadline, "GetSessionAccountID 在缓存未命中时应携带独立超时上下文")
	require.Greater(t, probe2.getDeadlineDelta, 2*time.Second)
	require.LessOrEqual(t, probe2.getDeadlineDelta, 3*time.Second)
}

func TestWithOpenAIWSStateStoreRedisTimeout_WithParentContext(t *testing.T) {
	ctx, cancel := withOpenAIWSStateStoreRedisTimeout(context.Background())
	defer cancel()
	require.NotNil(t, ctx)
	_, ok := ctx.Deadline()
	require.True(t, ok, "应附加短超时")
}

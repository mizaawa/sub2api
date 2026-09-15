package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ttlAwareOpenAIHTTPResponseBindingTestCache models a cache implementation
// that can report the remaining lifetime of a persisted combined binding.
// The embedded test cache keeps the existing value and call-count behavior.
type ttlAwareOpenAIHTTPResponseBindingTestCache struct {
	*openAIHTTPResponseBindingTestCache
	ttl time.Duration
}

func (c *ttlAwareOpenAIHTTPResponseBindingTestCache) GetOpenAIHTTPResponseBindingWithTTL(
	ctx context.Context,
	groupID int64,
	key string,
) (int64, int64, time.Duration, error) {
	accountID, userID, err := c.GetOpenAIHTTPResponseBinding(ctx, groupID, key)
	return accountID, userID, c.ttl, err
}

func TestOpenAIWSStateStore_CombinedBindingUsesPersistedRemainingTTL(t *testing.T) {
	groupID := int64(31)
	responseID := "resp_remaining_ttl"
	cache := &ttlAwareOpenAIHTTPResponseBindingTestCache{
		openAIHTTPResponseBindingTestCache: &openAIHTTPResponseBindingTestCache{
			bindings: map[string]openAIHTTPResponseBindingTestValue{
				openAIHTTPResponseBindingTestKey(groupID, openAIHTTPResponseBindingCacheKey(responseID)): {accountID: 901, userID: 902},
			},
		},
		ttl: 5 * time.Second,
	}

	raw := NewOpenAIWSStateStore(cache)
	accountID, err := raw.GetResponseAccount(context.Background(), groupID, responseID)
	require.NoError(t, err)
	require.Equal(t, int64(901), accountID)

	store, ok := raw.(*defaultOpenAIWSStateStore)
	require.True(t, ok)
	mapKey := openAIWSResponseAccountMapKey(groupID, responseID)
	store.responseToAccountMu.RLock()
	accountBinding, accountFound := store.responseToAccount[mapKey]
	store.responseToAccountMu.RUnlock()
	store.responseOwnerMu.RLock()
	ownerBinding, ownerFound := store.responseOwners[mapKey]
	store.responseOwnerMu.RUnlock()

	require.True(t, accountFound)
	require.True(t, ownerFound)
	accountRemaining := time.Until(accountBinding.expiresAt)
	ownerRemaining := time.Until(ownerBinding.expiresAt)
	require.Greater(t, accountRemaining, time.Duration(0))
	require.Greater(t, ownerRemaining, time.Duration(0))
	require.LessOrEqual(t, accountRemaining, cache.ttl)
	require.LessOrEqual(t, ownerRemaining, cache.ttl)
}

func TestOpenAIWSStateStore_LegacyOwnerFallbackDoesNotCacheUnknownTTL(t *testing.T) {
	groupID := int64(32)
	responseID := "resp_legacy_ttl"
	cache := &openAIHTTPResponseBindingTestCache{
		legacyBindings: map[string]int64{
			openAIHTTPResponseBindingTestKey(
				groupID,
				openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, responseID),
			): 903,
		},
	}

	store := NewOpenAIWSStateStore(cache)
	userID, apiKeyID, found, err := store.GetHTTPResponseOwner(context.Background(), groupID, responseID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, int64(903), userID)
	require.Zero(t, apiKeyID)

	delete(cache.legacyBindings, openAIHTTPResponseBindingTestKey(
		groupID,
		openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, responseID),
	))
	userID, apiKeyID, found, err = store.GetHTTPResponseOwner(context.Background(), groupID, responseID)
	require.ErrorIs(t, err, ErrStickySessionNotFound)
	require.Zero(t, userID)
	require.Zero(t, apiKeyID)
	require.False(t, found)
}

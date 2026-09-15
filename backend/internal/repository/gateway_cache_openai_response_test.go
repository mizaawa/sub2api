package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestGatewayCacheOpenAIHTTPResponseBindingUsesSingleRedisSet(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	require.NoError(t, client.Ping(ctx).Err())

	cache, ok := NewGatewayCache(client).(service.OpenAIHTTPResponseBindingCache)
	require.True(t, ok)
	groupID := int64(41)
	bindingKey := "openai:http-response-binding:test-hash"
	ttl := 2 * time.Minute

	commandsBefore := redisServer.CommandCount()
	require.NoError(t, cache.SetOpenAIHTTPResponseBinding(ctx, groupID, bindingKey, 901, 902, ttl))
	require.Equal(t, commandsBefore+1, redisServer.CommandCount(), "binding must be persisted by one Redis command")

	redisKey := buildSessionKey(groupID, bindingKey)
	require.Equal(t, []string{redisKey}, redisServer.Keys())
	stored, err := redisServer.Get(redisKey)
	require.NoError(t, err)
	require.JSONEq(t, `{"account_id":901,"user_id":902}`, stored)
	require.Equal(t, ttl, redisServer.TTL(redisKey))

	otherCache, ok := NewGatewayCache(client).(service.OpenAIHTTPResponseBindingCache)
	require.True(t, ok)
	accountID, userID, err := otherCache.GetOpenAIHTTPResponseBinding(ctx, groupID, bindingKey)
	require.NoError(t, err)
	require.Equal(t, int64(901), accountID)
	require.Equal(t, int64(902), userID)

	require.NoError(t, otherCache.DeleteOpenAIHTTPResponseBinding(ctx, groupID, bindingKey))
	_, _, err = cache.GetOpenAIHTTPResponseBinding(ctx, groupID, bindingKey)
	require.ErrorIs(t, err, service.ErrStickySessionNotFound)
}

func TestGatewayCacheOpenAIHTTPResponseBindingMissingAndCorrupt(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	cache, ok := NewGatewayCache(client).(service.OpenAIHTTPResponseBindingCache)
	require.True(t, ok)

	_, _, err := cache.GetOpenAIHTTPResponseBinding(ctx, 42, "missing")
	require.ErrorIs(t, err, service.ErrStickySessionNotFound)

	key := buildSessionKey(42, "corrupt")
	require.NoError(t, client.Set(ctx, key, `{"account_id":903}`, time.Minute).Err())
	_, _, err = cache.GetOpenAIHTTPResponseBinding(ctx, 42, "corrupt")
	require.Error(t, err)
	require.NotErrorIs(t, err, service.ErrStickySessionNotFound)
}

func TestGatewayCacheOpenAIHTTPResponseBindingWithLegacyMirrorsIsAtomic(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	cache, ok := NewGatewayCache(client).(service.OpenAIHTTPResponseBindingLegacyCache)
	require.True(t, ok)

	groupID := int64(43)
	combinedKey := "openai:http-response-binding:combined-hash"
	legacyAccountKey := "openai:response:legacy-account-hash"
	legacyOwnerKey := "openai:http-response-owner:user:legacy-owner-hash"
	ttl := 2 * time.Minute
	require.NoError(t, cache.SetOpenAIHTTPResponseBindingWithLegacy(
		ctx,
		groupID,
		combinedKey,
		legacyAccountKey,
		legacyOwnerKey,
		904,
		905,
		ttl,
	))

	combinedRedisKey := buildSessionKey(groupID, combinedKey)
	legacyAccountRedisKey := buildSessionKey(groupID, legacyAccountKey)
	legacyOwnerRedisKey := buildSessionKey(groupID, legacyOwnerKey)
	require.ElementsMatch(t, []string{
		combinedRedisKey,
		legacyAccountRedisKey,
		legacyOwnerRedisKey,
	}, redisServer.Keys())
	combined, err := redisServer.Get(combinedRedisKey)
	require.NoError(t, err)
	require.JSONEq(t, `{"account_id":904,"user_id":905}`, combined)
	account, err := redisServer.Get(legacyAccountRedisKey)
	require.NoError(t, err)
	require.Equal(t, "904", account)
	owner, err := redisServer.Get(legacyOwnerRedisKey)
	require.NoError(t, err)
	require.Equal(t, "905", owner)
	require.Equal(t, ttl, redisServer.TTL(combinedRedisKey))
	require.Equal(t, ttl, redisServer.TTL(legacyAccountRedisKey))
	require.Equal(t, ttl, redisServer.TTL(legacyOwnerRedisKey))
}

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

func TestGatewayCacheOpenAIHTTPResponseBindingWithTTLReportsRemainingLifetime(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	baseCache, ok := NewGatewayCache(client).(service.OpenAIHTTPResponseBindingCache)
	require.True(t, ok)
	ttlCache, ok := NewGatewayCache(client).(service.OpenAIHTTPResponseBindingTTLCache)
	require.True(t, ok)

	groupID := int64(52)
	bindingKey := "openai:http-response-binding:ttl-hash"
	ttl := 2 * time.Minute
	require.NoError(t, baseCache.SetOpenAIHTTPResponseBinding(ctx, groupID, bindingKey, 911, 912, ttl))

	accountID, userID, remaining, err := ttlCache.GetOpenAIHTTPResponseBindingWithTTL(ctx, groupID, bindingKey)
	require.NoError(t, err)
	require.Equal(t, int64(911), accountID)
	require.Equal(t, int64(912), userID)
	require.Greater(t, remaining, time.Duration(0))
	require.LessOrEqual(t, remaining, ttl)

	redisServer.FastForward(ttl + time.Millisecond)
	_, _, _, err = ttlCache.GetOpenAIHTTPResponseBindingWithTTL(ctx, groupID, bindingKey)
	require.ErrorIs(t, err, service.ErrStickySessionNotFound)
}

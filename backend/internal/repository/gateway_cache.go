package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const stickySessionPrefix = "sticky_session:"
const liveCallPrefix = "live:call:"

type gatewayCache struct {
	rdb *redis.Client
}

type openAIHTTPResponseBindingValue struct {
	AccountID int64 `json:"account_id"`
	UserID    int64 `json:"user_id"`
}

func NewGatewayCache(rdb *redis.Client) service.GatewayCache {
	return &gatewayCache{rdb: rdb}
}

// buildSessionKey 构建 session key，包含 groupID 实现分组隔离
// 格式: sticky_session:{groupID}:{sessionHash}
func buildSessionKey(groupID int64, sessionHash string) string {
	return fmt.Sprintf("%s%d:%s", stickySessionPrefix, groupID, sessionHash)
}

func (c *gatewayCache) GetSessionAccountID(ctx context.Context, groupID int64, sessionHash string) (int64, error) {
	key := buildSessionKey(groupID, sessionHash)
	accountID, err := c.rdb.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, service.ErrStickySessionNotFound
		}
		return 0, err
	}
	return accountID, nil
}

func (c *gatewayCache) SetSessionAccountID(ctx context.Context, groupID int64, sessionHash string, accountID int64, ttl time.Duration) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Set(ctx, key, accountID, ttl).Err()
}

// SetOpenAIHTTPResponseBinding stores both halves of an HTTP Responses binding
// in one Redis string. SET applies the value and TTL atomically, so a caller
// can never observe only the account or only the downstream owner.
func (c *gatewayCache) SetOpenAIHTTPResponseBinding(
	ctx context.Context,
	groupID int64,
	bindingKey string,
	accountID, userID int64,
	ttl time.Duration,
) error {
	if accountID <= 0 || userID <= 0 {
		return fmt.Errorf("invalid OpenAI HTTP response binding")
	}
	value, err := json.Marshal(openAIHTTPResponseBindingValue{AccountID: accountID, UserID: userID})
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, buildSessionKey(groupID, bindingKey), value, ttl).Err()
}

// SetOpenAIHTTPResponseBindingWithLegacy writes the new combined binding and
// the legacy account/owner mirrors in one Redis transaction. The mirrors keep
// response continuations readable while older gateway instances are still in
// the rolling-deployment window; the combined value remains the authoritative
// source for newer instances.
func (c *gatewayCache) SetOpenAIHTTPResponseBindingWithLegacy(
	ctx context.Context,
	groupID int64,
	bindingKey, legacyAccountKey, legacyOwnerKey string,
	accountID, userID int64,
	ttl time.Duration,
) error {
	if accountID <= 0 || userID <= 0 {
		return fmt.Errorf("invalid OpenAI HTTP response binding")
	}
	value, err := json.Marshal(openAIHTTPResponseBindingValue{AccountID: accountID, UserID: userID})
	if err != nil {
		return err
	}
	pipe := c.rdb.TxPipeline()
	pipe.Set(ctx, buildSessionKey(groupID, bindingKey), value, ttl)
	pipe.Set(ctx, buildSessionKey(groupID, legacyAccountKey), accountID, ttl)
	pipe.Set(ctx, buildSessionKey(groupID, legacyOwnerKey), userID, ttl)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *gatewayCache) GetOpenAIHTTPResponseBinding(
	ctx context.Context,
	groupID int64,
	bindingKey string,
) (accountID, userID int64, err error) {
	value, err := c.rdb.Get(ctx, buildSessionKey(groupID, bindingKey)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, 0, service.ErrStickySessionNotFound
		}
		return 0, 0, err
	}
	return decodeOpenAIHTTPResponseBinding(value)
}

func decodeOpenAIHTTPResponseBinding(value []byte) (accountID, userID int64, err error) {
	var binding openAIHTTPResponseBindingValue
	if err := json.Unmarshal(value, &binding); err != nil {
		return 0, 0, fmt.Errorf("decode OpenAI HTTP response binding: %w", err)
	}
	if binding.AccountID <= 0 || binding.UserID <= 0 {
		return 0, 0, fmt.Errorf("invalid OpenAI HTTP response binding")
	}
	return binding.AccountID, binding.UserID, nil
}

// GetOpenAIHTTPResponseBindingWithTTL reads a combined binding and reports
// the remaining Redis lifetime.  The TTL is measured after the value read and
// reduced by the local round-trip (plus one millisecond of clock-resolution
// slack), so a caller's process-local copy cannot intentionally outlive the
// Redis key merely because the read took time.
func (c *gatewayCache) GetOpenAIHTTPResponseBindingWithTTL(
	ctx context.Context,
	groupID int64,
	bindingKey string,
) (accountID, userID int64, ttl time.Duration, err error) {
	started := time.Now()
	key := buildSessionKey(groupID, bindingKey)
	value, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, 0, 0, service.ErrStickySessionNotFound
		}
		return 0, 0, 0, err
	}

	accountID, userID, err = decodeOpenAIHTTPResponseBinding(value)
	if err != nil {
		return 0, 0, 0, err
	}

	remaining, err := c.rdb.PTTL(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, 0, 0, service.ErrStickySessionNotFound
		}
		return 0, 0, 0, err
	}
	if remaining == -2 {
		// Redis uses -2 for a missing key (go-redis preserves this sentinel as
		// a duration value rather than scaling it by the PTTL precision). Treat
		// it as a miss instead of hydrating stale local state.
		return 0, 0, 0, service.ErrStickySessionNotFound
	}
	if remaining < 0 {
		// A positive TTL is expected for bindings written by this service. An
		// old or manually-created persistent key is still safe to read, but do
		// not create a process-local copy whose lifetime is guessed.
		return accountID, userID, 0, nil
	}

	// PTTL has millisecond resolution. Account for the elapsed GET/command
	// round-trip and one extra millisecond so local expiry is never later than
	// the Redis expiry due to rounding or the caller's bind latency.
	ttl = remaining - time.Since(started) - time.Millisecond
	if ttl < 0 {
		ttl = 0
	}
	return accountID, userID, ttl, nil
}

func (c *gatewayCache) DeleteOpenAIHTTPResponseBinding(ctx context.Context, groupID int64, bindingKey string) error {
	return c.rdb.Del(ctx, buildSessionKey(groupID, bindingKey)).Err()
}

func (c *gatewayCache) RefreshSessionTTL(ctx context.Context, groupID int64, sessionHash string, ttl time.Duration) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Expire(ctx, key, ttl).Err()
}

// DeleteSessionAccountID 删除粘性会话与账号的绑定关系。
// 当检测到绑定的账号不可用（如状态错误、禁用、不可调度等）时调用，
// 以便下次请求能够重新选择可用账号。
//
// DeleteSessionAccountID removes the sticky session binding for the given session.
// Called when the bound account becomes unavailable (e.g., error status, disabled,
// or unschedulable), allowing subsequent requests to select a new available account.
func (c *gatewayCache) DeleteSessionAccountID(ctx context.Context, groupID int64, sessionHash string) error {
	key := buildSessionKey(groupID, sessionHash)
	return c.rdb.Del(ctx, key).Err()
}

// Compile-time assertion: gatewayCache must implement CyberSessionBlockStore.
var _ service.CyberSessionBlockStore = (*gatewayCache)(nil)
var _ service.LiveCallStore = (*gatewayCache)(nil)
var _ service.OpenAIHTTPResponseBindingCache = (*gatewayCache)(nil)
var _ service.OpenAIHTTPResponseBindingTTLCache = (*gatewayCache)(nil)
var _ service.OpenAIHTTPResponseBindingLegacyCache = (*gatewayCache)(nil)

const cyberSessionBlockPrefix = "cyber_session_block:"

// SetCyberSessionBlocked 把被 cyber_policy 命中的会话写入屏蔽表（TTL 自动过期）。
// 存储值 "1" 作为存在标记（IsCyberSessionBlocked 只检查 key 是否存在，不读值）。
func (c *gatewayCache) SetCyberSessionBlocked(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Set(ctx, cyberSessionBlockPrefix+key, "1", ttl).Err()
}

// IsCyberSessionBlocked 查询会话是否在屏蔽表中。
func (c *gatewayCache) IsCyberSessionBlocked(ctx context.Context, key string) (bool, error) {
	n, err := c.rdb.Exists(ctx, cyberSessionBlockPrefix+key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

var claimLiveControllerScript = redis.NewScript(`
	local key = KEYS[1]
	local target = ARGV[1]
	local owner = ARGV[2]
	local current = redis.call('HGET', key, 'controller')
	if current == false or current == 'closed' then
		return 0
	end
	if target == 'observer' and current ~= 'pending' then
		return 0
	end
	if target == 'proxy' and current ~= 'pending' and current ~= 'observer' and
		(current ~= 'proxy' or redis.call('HGET', key, 'controller_owner') ~= owner) then
		return 0
	end
	redis.call('HSET', key, 'controller', target, 'controller_owner', owner)
	return 1
`)

var markLiveCallClosedScript = redis.NewScript(`
	local key = KEYS[1]
	if redis.call('EXISTS', key) == 0 then
		return 0
	end
	if redis.call('HGET', key, 'controller') == 'closed' then
		return 0
	end
	redis.call('HSET', key, 'controller', 'closed', 'controller_owner', '')
	redis.call('EXPIRE', key, ARGV[1])
	return 1
`)

var releaseLiveControllerScript = redis.NewScript(`
	local key = KEYS[1]
	if redis.call('HGET', key, 'controller') ~= 'proxy' or
		redis.call('HGET', key, 'controller_owner') ~= ARGV[1] then
		return 0
	end
	redis.call('HSET', key, 'controller', 'pending', 'controller_owner', '')
	return 1
`)

func liveCallKey(callHash string) string {
	return liveCallPrefix + callHash
}

func HashLiveCallID(callID string) string {
	sum := sha256.Sum256([]byte(callID))
	return hex.EncodeToString(sum[:])
}

func (c *gatewayCache) SaveLiveCall(ctx context.Context, record *service.LiveCallRecord, ttl time.Duration) error {
	if record == nil || record.CallHash == "" || record.CallID == "" {
		return fmt.Errorf("invalid live call record")
	}
	values := map[string]any{
		"call_id":          record.CallID,
		"account_id":       record.AccountID,
		"api_key_id":       record.APIKeyID,
		"user_id":          record.UserID,
		"group_id":         record.GroupID,
		"subscription_id":  record.SubscriptionID,
		"lease_id":         record.LeaseID,
		"model":            record.Model,
		"created_at":       record.CreatedAt.UnixMilli(),
		"expires_at":       record.ExpiresAt.UnixMilli(),
		"controller":       record.Controller,
		"controller_owner": record.ControllerOwner,
		"user_agent":       record.UserAgent,
		"ip_address":       record.IPAddress,
		"inbound_endpoint": record.InboundEndpoint,
		"attestation":      record.AttestationCiphertext,
	}
	key := liveCallKey(record.CallHash)
	pipe := c.rdb.TxPipeline()
	pipe.HSet(ctx, key, values)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *gatewayCache) GetLiveCall(ctx context.Context, callHash string) (*service.LiveCallRecord, error) {
	values, err := c.rdb.HGetAll(ctx, liveCallKey(callHash)).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, service.ErrLiveCallNotFound
	}
	parseInt := func(field string) int64 {
		value, _ := strconv.ParseInt(values[field], 10, 64)
		return value
	}
	createdAt := time.UnixMilli(parseInt("created_at"))
	expiresAt := time.UnixMilli(parseInt("expires_at"))
	return &service.LiveCallRecord{
		CallID:                values["call_id"],
		CallHash:              callHash,
		AccountID:             parseInt("account_id"),
		APIKeyID:              parseInt("api_key_id"),
		UserID:                parseInt("user_id"),
		GroupID:               parseInt("group_id"),
		SubscriptionID:        parseInt("subscription_id"),
		LeaseID:               values["lease_id"],
		Model:                 values["model"],
		CreatedAt:             createdAt,
		ExpiresAt:             expiresAt,
		Controller:            values["controller"],
		ControllerOwner:       values["controller_owner"],
		UserAgent:             values["user_agent"],
		IPAddress:             values["ip_address"],
		InboundEndpoint:       values["inbound_endpoint"],
		AttestationCiphertext: values["attestation"],
	}, nil
}

func (c *gatewayCache) ClaimLiveController(ctx context.Context, callHash, controller, owner string) (bool, error) {
	result, err := claimLiveControllerScript.Run(ctx, c.rdb, []string{liveCallKey(callHash)}, controller, owner).Int()
	return result == 1, err
}

func (c *gatewayCache) GetLiveController(ctx context.Context, callHash string) (string, error) {
	value, err := c.rdb.HGet(ctx, liveCallKey(callHash), "controller").Result()
	if err == redis.Nil {
		return "", service.ErrLiveCallNotFound
	}
	return value, err
}

func (c *gatewayCache) ReleaseLiveController(ctx context.Context, callHash, owner string) (bool, error) {
	result, err := releaseLiveControllerScript.Run(ctx, c.rdb, []string{liveCallKey(callHash)}, owner).Int()
	return result == 1, err
}

func (c *gatewayCache) MarkLiveCallClosed(ctx context.Context, callHash string, ttl time.Duration) (bool, error) {
	result, err := markLiveCallClosedScript.Run(ctx, c.rdb, []string{liveCallKey(callHash)}, int64(ttl.Seconds())).Int()
	return result == 1, err
}

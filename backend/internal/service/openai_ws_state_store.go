package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	openAIWSResponseAccountCachePrefix = "openai:response:"
	openAIHTTPResponseBindingPrefix    = "openai:http-response-binding:"
	openAIHTTPResponseOwnerUserPrefix  = "openai:http-response-owner:user:"
	openAIWSStateStoreCleanupInterval  = time.Minute
	openAIWSStateStoreCleanupMaxPerMap = 512
	openAIWSStateStoreMaxEntriesPerMap = 65536
	openAIWSStateStoreRedisTimeout     = 3 * time.Second
	openAIHTTPResponseBindRedisTimeout = time.Second
)

type openAIWSAccountBinding struct {
	accountID int64
	expiresAt time.Time
}

type openAIHTTPResponseOwnerBinding struct {
	userID    int64
	apiKeyID  int64
	expiresAt time.Time
}

type openAIWSConnBinding struct {
	connID    string
	expiresAt time.Time
}

type openAIWSTurnStateBinding struct {
	turnState string
	expiresAt time.Time
}

type openAIWSSessionConnBinding struct {
	connID    string
	expiresAt time.Time
}

// OpenAIHTTPResponseBindingCache is an optional GatewayCache capability that
// persists an HTTP Responses account/owner pair in one atomic cache write.
// The API key ID intentionally remains process-local: continuation ownership
// is portable across a user's keys, while account and user IDs must survive a
// process restart together.
type OpenAIHTTPResponseBindingCache interface {
	SetOpenAIHTTPResponseBinding(ctx context.Context, groupID int64, bindingKey string, accountID, userID int64, ttl time.Duration) error
	GetOpenAIHTTPResponseBinding(ctx context.Context, groupID int64, bindingKey string) (accountID, userID int64, err error)
	DeleteOpenAIHTTPResponseBinding(ctx context.Context, groupID int64, bindingKey string) error
}

// OpenAIHTTPResponseBindingTTLCache is an optional capability that returns the
// remaining lifetime of a persisted combined binding.  The state store uses
// this value to keep its process-local copy from surviving the Redis key.
// Implementations that cannot expose a remaining TTL can continue to satisfy
// OpenAIHTTPResponseBindingCache and use the compatibility fallback.
type OpenAIHTTPResponseBindingTTLCache interface {
	GetOpenAIHTTPResponseBindingWithTTL(
		ctx context.Context,
		groupID int64,
		bindingKey string,
	) (accountID, userID int64, ttl time.Duration, err error)
}

// OpenAIHTTPResponseBindingLegacyCache is an optional compatibility extension
// for rolling deployments. Implementations should write the combined binding
// and the two legacy mirrors in one atomic operation so an older instance can
// still resolve a response created by a newer instance.
type OpenAIHTTPResponseBindingLegacyCache interface {
	SetOpenAIHTTPResponseBindingWithLegacy(
		ctx context.Context,
		groupID int64,
		bindingKey, legacyAccountKey, legacyOwnerKey string,
		accountID, userID int64,
		ttl time.Duration,
	) error
}

// OpenAIWSStateStore 管理 WSv2 的粘连状态。
// - response_id -> account_id 用于续链路由
// - response_id -> conn_id 用于连接内上下文复用
//
// response_id -> account_id 优先走 GatewayCache（Redis），同时维护本地热缓存。
// response_id -> conn_id 仅在本进程内有效。
type OpenAIWSStateStore interface {
	BindResponseAccount(ctx context.Context, groupID int64, responseID string, accountID int64, ttl time.Duration) error
	GetResponseAccount(ctx context.Context, groupID int64, responseID string) (int64, error)
	DeleteResponseAccount(ctx context.Context, groupID int64, responseID string) error
	BindHTTPResponse(ctx context.Context, groupID int64, responseID string, accountID, userID, apiKeyID int64, ttl time.Duration) error
	BindHTTPResponseOwner(ctx context.Context, groupID int64, responseID string, userID, apiKeyID int64, ttl time.Duration) error
	GetHTTPResponseOwner(ctx context.Context, groupID int64, responseID string) (userID, apiKeyID int64, found bool, err error)

	BindResponseConn(responseID, connID string, ttl time.Duration)
	GetResponseConn(responseID string) (string, bool)
	DeleteResponseConn(responseID string)

	BindSessionTurnState(groupID int64, sessionHash, turnState string, ttl time.Duration)
	GetSessionTurnState(groupID int64, sessionHash string) (string, bool)
	DeleteSessionTurnState(groupID int64, sessionHash string)

	BindSessionConn(groupID int64, sessionHash, connID string, ttl time.Duration)
	GetSessionConn(groupID int64, sessionHash string) (string, bool)
	DeleteSessionConn(groupID int64, sessionHash string)
}

type defaultOpenAIWSStateStore struct {
	cache GatewayCache

	responseToAccountMu  sync.RWMutex
	responseToAccount    map[string]openAIWSAccountBinding
	responseOwnerMu      sync.RWMutex
	responseOwners       map[string]openAIHTTPResponseOwnerBinding
	responseToConnMu     sync.RWMutex
	responseToConn       map[string]openAIWSConnBinding
	sessionToTurnStateMu sync.RWMutex
	sessionToTurnState   map[string]openAIWSTurnStateBinding
	sessionToConnMu      sync.RWMutex
	sessionToConn        map[string]openAIWSSessionConnBinding

	lastCleanupUnixNano atomic.Int64
}

// NewOpenAIWSStateStore 创建默认 WS 状态存储。
func NewOpenAIWSStateStore(cache GatewayCache) OpenAIWSStateStore {
	store := &defaultOpenAIWSStateStore{
		cache:              cache,
		responseToAccount:  make(map[string]openAIWSAccountBinding, 256),
		responseOwners:     make(map[string]openAIHTTPResponseOwnerBinding, 256),
		responseToConn:     make(map[string]openAIWSConnBinding, 256),
		sessionToTurnState: make(map[string]openAIWSTurnStateBinding, 256),
		sessionToConn:      make(map[string]openAIWSSessionConnBinding, 256),
	}
	store.lastCleanupUnixNano.Store(time.Now().UnixNano())
	return store
}

// BindHTTPResponse persists the upstream account and downstream owner as one
// logical binding. Production caches implement the optional legacy extension
// and write the combined value plus compatibility mirrors atomically. Custom
// caches can implement the base interface and keep the combined single-key
// write. The legacy fallback keeps older GatewayCache implementations working
// and rolls back its first write if the second one fails. The coherent local
// pair is published before persistence, matching existing sticky-session
// availability when Redis is temporarily unavailable; persistence errors are
// still returned so callers can log and monitor them.
func (s *defaultOpenAIWSStateStore) BindHTTPResponse(
	ctx context.Context,
	groupID int64,
	responseID string,
	accountID, userID, apiKeyID int64,
	ttl time.Duration,
) error {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" || accountID <= 0 || userID <= 0 || apiKeyID <= 0 {
		return nil
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()
	s.bindLocalHTTPResponse(id, groupID, accountID, userID, apiKeyID, ttl)

	if s.cache != nil {
		cacheCtx, cancel := withOpenAIHTTPResponseBindRedisTimeout(ctx)
		defer cancel()

		if cache, ok := s.cache.(OpenAIHTTPResponseBindingLegacyCache); ok {
			if err := cache.SetOpenAIHTTPResponseBindingWithLegacy(
				cacheCtx,
				groupID,
				openAIHTTPResponseBindingCacheKey(id),
				openAIWSResponseAccountCacheKey(id),
				openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, id),
				accountID,
				userID,
				ttl,
			); err != nil {
				return err
			}
		} else if cache, ok := s.cache.(OpenAIHTTPResponseBindingCache); ok {
			if err := cache.SetOpenAIHTTPResponseBinding(
				cacheCtx,
				groupID,
				openAIHTTPResponseBindingCacheKey(id),
				accountID,
				userID,
				ttl,
			); err != nil {
				return err
			}
		} else {
			accountKey := openAIWSResponseAccountCacheKey(id)
			if err := s.cache.SetSessionAccountID(cacheCtx, groupID, accountKey, accountID, ttl); err != nil {
				return err
			}
			ownerKey := openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, id)
			if err := s.cache.SetSessionAccountID(cacheCtx, groupID, ownerKey, userID, ttl); err != nil {
				// The first write may have consumed the shared bind timeout while
				// the second write was in flight. Cleanup must get a fresh budget;
				// reusing cacheCtx would leave an account-only legacy binding behind.
				rollbackCtx, rollbackCancel := withOpenAIHTTPResponseBindRedisTimeout(context.Background())
				defer rollbackCancel()
				return errors.Join(err, s.rollbackLegacyHTTPResponseBinding(rollbackCtx, groupID, id))
			}
		}
	}

	return nil
}

func (s *defaultOpenAIWSStateStore) rollbackLegacyHTTPResponseBinding(
	ctx context.Context,
	groupID int64,
	responseID string,
) error {
	return errors.Join(
		s.cache.DeleteSessionAccountID(ctx, groupID, openAIWSResponseAccountCacheKey(responseID)),
		s.cache.DeleteSessionAccountID(ctx, groupID, openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, responseID)),
	)
}

func (s *defaultOpenAIWSStateStore) bindLocalHTTPResponse(
	responseID string,
	groupID, accountID, userID, apiKeyID int64,
	ttl time.Duration,
) {
	mapKey := openAIWSResponseAccountMapKey(groupID, responseID)
	expiresAt := time.Now().Add(ttl)

	// Always acquire these locks in this order. The pair is published together
	// before the remote write, so a temporary cache outage still leaves a
	// coherent single-process continuation; concurrent writers cannot interleave
	// the account and owner halves.
	s.responseToAccountMu.Lock()
	s.responseOwnerMu.Lock()
	ensureBindingCapacity(s.responseToAccount, mapKey, openAIWSStateStoreMaxEntriesPerMap)
	ensureBindingCapacity(s.responseOwners, mapKey, openAIWSStateStoreMaxEntriesPerMap)
	s.responseToAccount[mapKey] = openAIWSAccountBinding{accountID: accountID, expiresAt: expiresAt}
	s.responseOwners[mapKey] = openAIHTTPResponseOwnerBinding{
		userID: userID, apiKeyID: apiKeyID, expiresAt: expiresAt,
	}
	s.responseOwnerMu.Unlock()
	s.responseToAccountMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) BindHTTPResponseOwner(ctx context.Context, groupID int64, responseID string, userID, apiKeyID int64, ttl time.Duration) error {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" || userID <= 0 || apiKeyID <= 0 {
		return nil
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	if s.cache != nil {
		cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
		defer cancel()
		if err := s.cache.SetSessionAccountID(cacheCtx, groupID, openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, id), userID, ttl); err != nil {
			return err
		}
	}

	s.responseOwnerMu.Lock()
	ensureBindingCapacity(s.responseOwners, mapKey, openAIWSStateStoreMaxEntriesPerMap)
	s.responseOwners[mapKey] = openAIHTTPResponseOwnerBinding{
		userID: userID, apiKeyID: apiKeyID, expiresAt: time.Now().Add(ttl),
	}
	s.responseOwnerMu.Unlock()
	return nil
}

func (s *defaultOpenAIWSStateStore) GetHTTPResponseOwner(ctx context.Context, groupID int64, responseID string) (int64, int64, bool, error) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return 0, 0, false, nil
	}
	s.maybeCleanup()

	now := time.Now()
	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	s.responseOwnerMu.RLock()
	if binding, ok := s.responseOwners[mapKey]; ok && now.Before(binding.expiresAt) {
		s.responseOwnerMu.RUnlock()
		return binding.userID, binding.apiKeyID, true, nil
	}
	s.responseOwnerMu.RUnlock()

	if s.cache == nil {
		return 0, 0, false, nil
	}
	if accountID, userID, persistedTTL, found, err := s.getPersistedHTTPResponseBinding(ctx, groupID, id); err != nil {
		return 0, 0, false, err
	} else if found {
		if persistedTTL > 0 {
			s.bindLocalHTTPResponse(id, groupID, accountID, userID, 0, persistedTTL)
		}
		return userID, 0, true, nil
	}

	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	userID, err := s.cache.GetSessionAccountID(cacheCtx, groupID, openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, id))
	if err != nil || userID <= 0 {
		return 0, 0, false, err
	}
	// Legacy owner keys expose no remaining TTL. Do not hydrate a process-local
	// copy with a guessed lifetime: the Redis key may expire sooner than a
	// minute, and a stale local owner would incorrectly authorize a continuation.
	return userID, 0, true, nil
}

func (s *defaultOpenAIWSStateStore) BindResponseAccount(ctx context.Context, groupID int64, responseID string, accountID int64, ttl time.Duration) error {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" || accountID <= 0 {
		return nil
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	expiresAt := time.Now().Add(ttl)
	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	s.responseToAccountMu.Lock()
	ensureBindingCapacity(s.responseToAccount, mapKey, openAIWSStateStoreMaxEntriesPerMap)
	s.responseToAccount[mapKey] = openAIWSAccountBinding{accountID: accountID, expiresAt: expiresAt}
	s.responseToAccountMu.Unlock()

	if s.cache == nil {
		return nil
	}
	cacheKey := openAIWSResponseAccountCacheKey(id)
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	return s.cache.SetSessionAccountID(cacheCtx, groupID, cacheKey, accountID, ttl)
}

func cleanupExpiredHTTPResponseOwnerBindings(bindings map[string]openAIHTTPResponseOwnerBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func (s *defaultOpenAIWSStateStore) GetResponseAccount(ctx context.Context, groupID int64, responseID string) (int64, error) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return 0, nil
	}
	s.maybeCleanup()

	now := time.Now()
	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	s.responseToAccountMu.RLock()
	if binding, ok := s.responseToAccount[mapKey]; ok {
		if now.Before(binding.expiresAt) {
			accountID := binding.accountID
			s.responseToAccountMu.RUnlock()
			return accountID, nil
		}
	}
	s.responseToAccountMu.RUnlock()

	if s.cache == nil {
		return 0, nil
	}
	if accountID, userID, persistedTTL, found, err := s.getPersistedHTTPResponseBinding(ctx, groupID, id); err != nil {
		// Preserve the established response-account behavior: cache read failures
		// are treated as misses so transient Redis outages do not abort routing.
		return 0, nil
	} else if found {
		// A combined cache hit also carries the owner. Hydrate both local maps so
		// the authorization check and account resolution share this one read.
		if persistedTTL > 0 {
			s.bindLocalHTTPResponse(id, groupID, accountID, userID, 0, persistedTTL)
		}
		return accountID, nil
	}

	cacheKey := openAIWSResponseAccountCacheKey(id)
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	accountID, err := s.cache.GetSessionAccountID(cacheCtx, groupID, cacheKey)
	if err != nil || accountID <= 0 {
		// 缓存读取失败不阻断主流程，按未命中降级。
		return 0, nil
	}
	return accountID, nil
}

func (s *defaultOpenAIWSStateStore) DeleteResponseAccount(ctx context.Context, groupID int64, responseID string) error {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return nil
	}
	s.responseToAccountMu.Lock()
	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	delete(s.responseToAccount, mapKey)
	s.responseToAccountMu.Unlock()
	s.responseOwnerMu.Lock()
	delete(s.responseOwners, mapKey)
	s.responseOwnerMu.Unlock()

	if s.cache == nil {
		return nil
	}
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	var deleteErrs []error
	if cache, ok := s.cache.(OpenAIHTTPResponseBindingCache); ok {
		if err := cache.DeleteOpenAIHTTPResponseBinding(cacheCtx, groupID, openAIHTTPResponseBindingCacheKey(id)); err != nil {
			deleteErrs = append(deleteErrs, fmt.Errorf("delete combined HTTP response binding: %w", err))
		}
	}
	if err := s.cache.DeleteSessionAccountID(cacheCtx, groupID, openAIWSResponseAccountCacheKey(id)); err != nil {
		deleteErrs = append(deleteErrs, fmt.Errorf("delete legacy response account binding: %w", err))
	}
	if err := s.cache.DeleteSessionAccountID(cacheCtx, groupID, openAIHTTPResponseOwnerCacheKey(openAIHTTPResponseOwnerUserPrefix, id)); err != nil {
		deleteErrs = append(deleteErrs, fmt.Errorf("delete legacy HTTP response owner binding: %w", err))
	}
	return errors.Join(deleteErrs...)
}

func (s *defaultOpenAIWSStateStore) BindResponseConn(responseID, connID string, ttl time.Duration) {
	id := normalizeOpenAIWSResponseID(responseID)
	conn := strings.TrimSpace(connID)
	if id == "" || conn == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.responseToConnMu.Lock()
	ensureBindingCapacity(s.responseToConn, id, openAIWSStateStoreMaxEntriesPerMap)
	s.responseToConn[id] = openAIWSConnBinding{
		connID:    conn,
		expiresAt: time.Now().Add(ttl),
	}
	s.responseToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetResponseConn(responseID string) (string, bool) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return "", false
	}
	s.maybeCleanup()

	now := time.Now()
	s.responseToConnMu.RLock()
	binding, ok := s.responseToConn[id]
	s.responseToConnMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.connID) == "" {
		return "", false
	}
	return binding.connID, true
}

func (s *defaultOpenAIWSStateStore) DeleteResponseConn(responseID string) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return
	}
	s.responseToConnMu.Lock()
	delete(s.responseToConn, id)
	s.responseToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) BindSessionTurnState(groupID int64, sessionHash, turnState string, ttl time.Duration) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	state := strings.TrimSpace(turnState)
	if key == "" || state == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.sessionToTurnStateMu.Lock()
	ensureBindingCapacity(s.sessionToTurnState, key, openAIWSStateStoreMaxEntriesPerMap)
	s.sessionToTurnState[key] = openAIWSTurnStateBinding{
		turnState: state,
		expiresAt: time.Now().Add(ttl),
	}
	s.sessionToTurnStateMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetSessionTurnState(groupID int64, sessionHash string) (string, bool) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return "", false
	}
	s.maybeCleanup()

	now := time.Now()
	s.sessionToTurnStateMu.RLock()
	binding, ok := s.sessionToTurnState[key]
	s.sessionToTurnStateMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.turnState) == "" {
		return "", false
	}
	return binding.turnState, true
}

func (s *defaultOpenAIWSStateStore) DeleteSessionTurnState(groupID int64, sessionHash string) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return
	}
	s.sessionToTurnStateMu.Lock()
	delete(s.sessionToTurnState, key)
	s.sessionToTurnStateMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) BindSessionConn(groupID int64, sessionHash, connID string, ttl time.Duration) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	conn := strings.TrimSpace(connID)
	if key == "" || conn == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.sessionToConnMu.Lock()
	ensureBindingCapacity(s.sessionToConn, key, openAIWSStateStoreMaxEntriesPerMap)
	s.sessionToConn[key] = openAIWSSessionConnBinding{
		connID:    conn,
		expiresAt: time.Now().Add(ttl),
	}
	s.sessionToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetSessionConn(groupID int64, sessionHash string) (string, bool) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return "", false
	}
	s.maybeCleanup()

	now := time.Now()
	s.sessionToConnMu.RLock()
	binding, ok := s.sessionToConn[key]
	s.sessionToConnMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.connID) == "" {
		return "", false
	}
	return binding.connID, true
}

func (s *defaultOpenAIWSStateStore) DeleteSessionConn(groupID int64, sessionHash string) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return
	}
	s.sessionToConnMu.Lock()
	delete(s.sessionToConn, key)
	s.sessionToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) maybeCleanup() {
	if s == nil {
		return
	}
	now := time.Now()
	last := time.Unix(0, s.lastCleanupUnixNano.Load())
	if now.Sub(last) < openAIWSStateStoreCleanupInterval {
		return
	}
	if !s.lastCleanupUnixNano.CompareAndSwap(last.UnixNano(), now.UnixNano()) {
		return
	}

	// 增量限额清理，避免高规模下一次性全量扫描导致长时间阻塞。
	s.responseToAccountMu.Lock()
	cleanupExpiredAccountBindings(s.responseToAccount, now, openAIWSStateStoreCleanupMaxPerMap)
	s.responseToAccountMu.Unlock()

	s.responseOwnerMu.Lock()
	cleanupExpiredHTTPResponseOwnerBindings(s.responseOwners, now, openAIWSStateStoreCleanupMaxPerMap)
	s.responseOwnerMu.Unlock()

	s.responseToConnMu.Lock()
	cleanupExpiredConnBindings(s.responseToConn, now, openAIWSStateStoreCleanupMaxPerMap)
	s.responseToConnMu.Unlock()

	s.sessionToTurnStateMu.Lock()
	cleanupExpiredTurnStateBindings(s.sessionToTurnState, now, openAIWSStateStoreCleanupMaxPerMap)
	s.sessionToTurnStateMu.Unlock()

	s.sessionToConnMu.Lock()
	cleanupExpiredSessionConnBindings(s.sessionToConn, now, openAIWSStateStoreCleanupMaxPerMap)
	s.sessionToConnMu.Unlock()
}

func cleanupExpiredAccountBindings(bindings map[string]openAIWSAccountBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredConnBindings(bindings map[string]openAIWSConnBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredTurnStateBindings(bindings map[string]openAIWSTurnStateBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredSessionConnBindings(bindings map[string]openAIWSSessionConnBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func ensureBindingCapacity[T any](bindings map[string]T, incomingKey string, maxEntries int) {
	if len(bindings) < maxEntries || maxEntries <= 0 {
		return
	}
	if _, exists := bindings[incomingKey]; exists {
		return
	}
	// 固定上限保护：淘汰任意一项，优先保证内存有界。
	for key := range bindings {
		delete(bindings, key)
		return
	}
}

func normalizeOpenAIWSResponseID(responseID string) string {
	return strings.TrimSpace(responseID)
}

func openAIWSResponseAccountCacheKey(responseID string) string {
	sum := sha256.Sum256([]byte(responseID))
	return openAIWSResponseAccountCachePrefix + hex.EncodeToString(sum[:])
}

func openAIHTTPResponseBindingCacheKey(responseID string) string {
	sum := sha256.Sum256([]byte(responseID))
	return openAIHTTPResponseBindingPrefix + hex.EncodeToString(sum[:])
}

func openAIHTTPResponseOwnerCacheKey(prefix, responseID string) string {
	sum := sha256.Sum256([]byte(responseID))
	return prefix + hex.EncodeToString(sum[:])
}

// openAIWSResponseAccountMapKey 本地热缓存按分组隔离的 key，与 Redis 层保持一致，避免跨组命中。
func openAIWSResponseAccountMapKey(groupID int64, responseID string) string {
	return fmt.Sprintf("%d:%s", groupID, responseID)
}

func normalizeOpenAIWSTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return time.Hour
	}
	return ttl
}

func openAIWSSessionTurnStateKey(groupID int64, sessionHash string) string {
	hash := strings.TrimSpace(sessionHash)
	if hash == "" {
		return ""
	}
	return fmt.Sprintf("%d:%s", groupID, hash)
}

func withOpenAIWSStateStoreRedisTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, openAIWSStateStoreRedisTimeout)
}

func withOpenAIHTTPResponseBindRedisTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, openAIHTTPResponseBindRedisTimeout)
}

func (s *defaultOpenAIWSStateStore) getPersistedHTTPResponseBinding(
	ctx context.Context,
	groupID int64,
	responseID string,
) (accountID, userID int64, ttl time.Duration, found bool, err error) {
	cache, ok := s.cache.(OpenAIHTTPResponseBindingCache)
	if !ok {
		return 0, 0, 0, false, nil
	}
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	bindingKey := openAIHTTPResponseBindingCacheKey(responseID)
	if ttlCache, ok := s.cache.(OpenAIHTTPResponseBindingTTLCache); ok {
		accountID, userID, ttl, err = ttlCache.GetOpenAIHTTPResponseBindingWithTTL(cacheCtx, groupID, bindingKey)
	} else {
		accountID, userID, err = cache.GetOpenAIHTTPResponseBinding(cacheCtx, groupID, bindingKey)
		// Older cache implementations do not expose Redis' remaining TTL. Keep
		// their historical behavior while production caches use the TTL-aware
		// capability above.
		ttl = time.Minute
	}
	if errors.Is(err, ErrStickySessionNotFound) {
		return 0, 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, 0, false, err
	}
	if accountID <= 0 || userID <= 0 {
		return 0, 0, 0, false, fmt.Errorf("invalid persisted HTTP response binding")
	}
	return accountID, userID, ttl, true, nil
}

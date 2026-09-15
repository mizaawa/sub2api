package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_Hit(t *testing.T) {
	ctx := context.Background()
	groupID := int64(23)
	account := Account{
		ID:          2,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 2,
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_enabled": true,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()

	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_1", account.ID, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, "resp_prev_1", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, account.ID, selection.Account.ID)
	require.True(t, selection.Acquired)
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_QuotaAutoPausedMiss(t *testing.T) {
	ctx := context.Background()
	groupID := int64(23)
	account := Account{
		ID:          77,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 2,
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_enabled": true,
			"codex_5h_used_percent":                         96.0,
			"auto_pause_5h_threshold":                       0.95,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_quota", account.ID, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, "resp_prev_quota", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.Nil(t, selection, "超过 5h 配额阈值的账号不应继续命中 previous_response_id 粘连")

	// Auto-pause is transient, so the binding is preserved: the chain can resume on the
	// same account once the quota window resets.
	boundAccountID, getErr := store.GetResponseAccount(ctx, groupID, "resp_prev_quota")
	require.NoError(t, getErr)
	require.Equal(t, account.ID, boundAccountID)
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_RateLimitedMiss(t *testing.T) {
	ctx := context.Background()
	groupID := int64(23)
	rateLimitedUntil := time.Now().Add(30 * time.Minute)
	account := Account{
		ID:               12,
		Platform:         PlatformOpenAI,
		Type:             AccountTypeAPIKey,
		Status:           StatusActive,
		Schedulable:      true,
		Concurrency:      1,
		RateLimitResetAt: &rateLimitedUntil,
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_enabled": true,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_rl", account.ID, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, "resp_prev_rl", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.Nil(t, selection, "限额中的账号不应继续命中 previous_response_id 粘连")
	boundAccountID, getErr := store.GetResponseAccount(ctx, groupID, "resp_prev_rl")
	require.NoError(t, getErr)
	require.Equal(t, account.ID, boundAccountID, "限流是暂态，恢复后续链仍必须回到原账号")
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_DBRuntimeRecheckRateLimitedMiss(t *testing.T) {
	ctx := context.Background()
	groupID := int64(24)
	rateLimitedUntil := time.Now().Add(30 * time.Minute)
	staleAccount := &Account{
		ID:          13,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_enabled": true,
		},
	}
	dbAccount := Account{
		ID:               13,
		Platform:         PlatformOpenAI,
		Type:             AccountTypeAPIKey,
		Status:           StatusActive,
		Schedulable:      true,
		Concurrency:      1,
		RateLimitResetAt: &rateLimitedUntil,
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_enabled": true,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	snapshotCache := &openAISnapshotCacheStub{
		accountsByID: map[int64]*Account{dbAccount.ID: staleAccount},
	}
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{dbAccount}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
		schedulerSnapshot:  &SchedulerSnapshotService{cache: snapshotCache},
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_db_rl", dbAccount.ID, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, "resp_prev_db_rl", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.Nil(t, selection, "DB 中已限流的账号不应继续命中 previous_response_id 粘连")
	boundAccountID, getErr := store.GetResponseAccount(ctx, groupID, "resp_prev_db_rl")
	require.NoError(t, getErr)
	require.Equal(t, dbAccount.ID, boundAccountID, "DB 复检得到的限流是暂态，不应破坏原续链")
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_TransientAccountStatePreservesBinding(t *testing.T) {
	previousDisableState := IsDisableTempUnschedulableEnabled()
	SetDisableTempUnschedulableRuntime(false)
	t.Cleanup(func() { SetDisableTempUnschedulableRuntime(previousDisableState) })

	ctx := context.Background()
	groupID := int64(27)
	blockedUntil := time.Now().Add(30 * time.Minute)
	modelResetAt := blockedUntil.Format(time.RFC3339)
	tests := []struct {
		name   string
		mutate func(*Account)
	}{
		{
			name: "manual scheduling pause",
			mutate: func(account *Account) {
				account.Schedulable = false
			},
		},
		{
			name: "overload window",
			mutate: func(account *Account) {
				account.OverloadUntil = &blockedUntil
			},
		},
		{
			name: "temporary unschedulable window",
			mutate: func(account *Account) {
				account.TempUnschedulableUntil = &blockedUntil
			},
		},
		{
			name: "model rate limit",
			mutate: func(account *Account) {
				account.Extra[modelRateLimitsKey] = map[string]any{
					"gpt-5.1": map[string]any{"rate_limit_reset_at": modelResetAt},
				}
			},
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			account := Account{
				ID:          int64(80 + i),
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				Status:      StatusActive,
				Schedulable: true,
				Concurrency: 1,
				Extra: map[string]any{
					"openai_apikey_responses_websockets_v2_enabled": true,
				},
			}
			tc.mutate(&account)
			cache := &stubGatewayCache{}
			store := NewOpenAIWSStateStore(cache)
			svc := &OpenAIGatewayService{
				accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
				cache:              cache,
				cfg:                newOpenAIWSV2TestConfig(),
				concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
				openaiWSStateStore: store,
			}
			responseID := fmt.Sprintf("resp_prev_transient_%d", i)
			require.NoError(t, store.BindResponseAccount(ctx, groupID, responseID, account.ID, time.Hour))

			selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, responseID, "gpt-5.1", nil, false)
			require.NoError(t, err)
			require.Nil(t, selection)
			boundAccountID, getErr := store.GetResponseAccount(ctx, groupID, responseID)
			require.NoError(t, getErr)
			require.Equal(t, account.ID, boundAccountID)
		})
	}
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_RuntimeCooldownsPreserveBinding(t *testing.T) {
	previousDisableState := IsDisableTempUnschedulableEnabled()
	SetDisableTempUnschedulableRuntime(false)
	t.Cleanup(func() { SetDisableTempUnschedulableRuntime(previousDisableState) })

	for _, tc := range []struct {
		name  string
		block func(*OpenAIGatewayService, *Account)
	}{
		{
			name: "account runtime cooldown",
			block: func(svc *OpenAIGatewayService, account *Account) {
				svc.BlockAccountScheduling(account, time.Now().Add(time.Minute), "test")
			},
		},
		{
			name: "model runtime cooldown",
			block: func(svc *OpenAIGatewayService, account *Account) {
				svc.recordOpenAIAccountModelTransientFailure(account, "gpt-5.1", time.Now())
				svc.recordOpenAIAccountModelTransientFailure(account, "gpt-5.1", time.Now())
				require.True(t, svc.isOpenAIAccountModelRuntimeBlocked(account, "gpt-5.1"))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			groupID := int64(28)
			account := &Account{
				ID: 91, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Status: StatusActive, Schedulable: true, Concurrency: 1,
				Extra: map[string]any{"openai_apikey_responses_websockets_v2_enabled": true},
			}
			cache := &stubGatewayCache{}
			store := NewOpenAIWSStateStore(cache)
			svc := &OpenAIGatewayService{
				accountRepo:        stubOpenAIAccountRepo{accounts: []Account{*account}},
				cache:              cache,
				cfg:                newOpenAIWSV2TestConfig(),
				concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
				openaiWSStateStore: store,
				schedulerSnapshot: &SchedulerSnapshotService{
					cache: &openAISnapshotCacheStub{accountsByID: map[int64]*Account{account.ID: account}},
				},
			}
			tc.block(svc, account)
			responseID := "resp_prev_runtime_" + strings.ReplaceAll(tc.name, " ", "_")
			require.NoError(t, store.BindResponseAccount(ctx, groupID, responseID, account.ID, time.Hour))

			selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, responseID, "gpt-5.1", nil, false)
			require.NoError(t, err)
			require.Nil(t, selection)
			boundAccountID, getErr := store.GetResponseAccount(ctx, groupID, responseID)
			require.NoError(t, getErr)
			require.Equal(t, account.ID, boundAccountID)
		})
	}
}

type previousResponseLifecycleRepo struct {
	AccountRepository
	account *Account
	err     error
}

func (r previousResponseLifecycleRepo) GetByID(context.Context, int64) (*Account, error) {
	return r.account, r.err
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_DeletesOnlyPermanentIdentityFailures(t *testing.T) {
	active := func(platform string) *Account {
		return &Account{
			ID: 101, Platform: platform, Type: AccountTypeAPIKey,
			Status: StatusActive, Schedulable: true, Concurrency: 1,
			Extra: map[string]any{"openai_apikey_responses_websockets_v2_enabled": true},
		}
	}
	inactive := active(PlatformOpenAI)
	inactive.Status = StatusDisabled
	inactiveOAuthForceHTTP := active(PlatformOpenAI)
	inactiveOAuthForceHTTP.Type = AccountTypeOAuth
	inactiveOAuthForceHTTP.Status = StatusDisabled
	inactiveOAuthForceHTTP.Extra["openai_ws_force_http"] = true
	stale := active(PlatformOpenAI)

	for _, tc := range []struct {
		name            string
		repo            AccountRepository
		snapshotAccount *Account
		wantBinding     bool
	}{
		{name: "lookup error", repo: previousResponseLifecycleRepo{err: context.DeadlineExceeded}, wantBinding: true},
		{name: "DB recheck lookup error", repo: previousResponseLifecycleRepo{err: context.DeadlineExceeded}, snapshotAccount: stale, wantBinding: true},
		{name: "account missing", repo: previousResponseLifecycleRepo{}, wantBinding: false},
		{name: "DB recheck account missing", repo: previousResponseLifecycleRepo{}, snapshotAccount: stale, wantBinding: false},
		{name: "account inactive", repo: previousResponseLifecycleRepo{account: inactive}, wantBinding: false},
		{name: "OAuth force HTTP account inactive", repo: previousResponseLifecycleRepo{account: inactiveOAuthForceHTTP}, wantBinding: false},
		{name: "platform changed", repo: previousResponseLifecycleRepo{account: active(PlatformGrok)}, wantBinding: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			groupID := int64(29)
			cache := &stubGatewayCache{}
			store := NewOpenAIWSStateStore(cache)
			svc := &OpenAIGatewayService{
				accountRepo:        tc.repo,
				cache:              cache,
				cfg:                newOpenAIWSV2TestConfig(),
				concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
				openaiWSStateStore: store,
			}
			if tc.snapshotAccount != nil {
				svc.schedulerSnapshot = &SchedulerSnapshotService{
					cache: &openAISnapshotCacheStub{accountsByID: map[int64]*Account{101: tc.snapshotAccount}},
				}
			}
			responseID := "resp_prev_lifecycle_" + strings.ReplaceAll(tc.name, " ", "_")
			require.NoError(t, store.BindResponseAccount(ctx, groupID, responseID, 101, time.Hour))

			selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, responseID, "gpt-5.1", nil, false)
			require.NoError(t, err)
			require.Nil(t, selection)
			boundAccountID, getErr := store.GetResponseAccount(ctx, groupID, responseID)
			require.NoError(t, getErr)
			if tc.wantBinding {
				require.Equal(t, int64(101), boundAccountID)
			} else {
				require.Zero(t, boundAccountID)
			}
		})
	}
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_Excluded(t *testing.T) {
	ctx := context.Background()
	groupID := int64(23)
	account := Account{
		ID:          8,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_enabled": true,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_2", account.ID, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, "resp_prev_2", "gpt-5.1", map[int64]struct{}{account.ID: {}}, false)
	require.NoError(t, err)
	require.Nil(t, selection)
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_APIKeyForceHTTPHit(t *testing.T) {
	ctx := context.Background()
	groupID := int64(23)
	account := Account{
		ID:          11,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Extra: map[string]any{
			"openai_ws_force_http":            true,
			"responses_websockets_v2_enabled": true,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_force_http", account.ID, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, "resp_prev_force_http", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.NotNil(t, selection, "API-key force_http 仍支持官方 Responses HTTP continuation")
	require.NotNil(t, selection.Account)
	require.Equal(t, account.ID, selection.Account.ID)
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_OAuthForceHTTPIgnored(t *testing.T) {
	ctx := context.Background()
	groupID := int64(23)
	account := Account{
		ID:          12,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Extra: map[string]any{
			"openai_ws_force_http":            true,
			"responses_websockets_v2_enabled": true,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_oauth_force_http", account.ID, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, "resp_prev_oauth_force_http", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.Nil(t, selection, "OAuth force_http 无法在 HTTP 上保留 previous_response_id 粘连")
}

func TestOpenAIGatewayService_SelectStrictPreviousResponseID_OAuthHTTPBridgeHit(t *testing.T) {
	ctx := context.Background()
	groupID := int64(24)
	account := Account{
		ID:          13,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Extra: map[string]any{
			"openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModeHTTPBridge,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	cfg.Gateway.OpenAIWS.ModeRouterV2Enabled = true
	cfg.Gateway.OpenAIWS.IngressModeDefault = OpenAIWSIngressModeCtxPool
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_http_bridge", account.ID, time.Hour))

	selection, decision, err := svc.selectStrictAccountByPreviousResponseID(
		ctx,
		&groupID,
		"resp_prev_http_bridge",
		PlatformOpenAI,
		"session-http-bridge",
		"gpt-5.1",
		nil,
		OpenAIUpstreamTransportResponsesWebsocketV2Ingress,
		OpenAIEndpointCapabilityResponses,
		"",
		false,
	)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, account.ID, selection.Account.ID)
	require.True(t, decision.StickyPreviousHit)
	if selection.ReleaseFunc != nil {
		selection.ReleaseFunc()
	}
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_BusyKeepsSticky(t *testing.T) {
	ctx := context.Background()
	groupID := int64(23)
	accounts := []Account{
		{
			ID:          21,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Concurrency: 1,
			Priority:    0,
			Extra: map[string]any{
				"openai_apikey_responses_websockets_v2_enabled": true,
			},
		},
		{
			ID:          22,
			Platform:    PlatformOpenAI,
			Type:        AccountTypeAPIKey,
			Status:      StatusActive,
			Schedulable: true,
			Concurrency: 1,
			Priority:    9,
			Extra: map[string]any{
				"openai_apikey_responses_websockets_v2_enabled": true,
			},
		},
	}

	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	cfg.Gateway.Scheduling.StickySessionMaxWaiting = 2
	cfg.Gateway.Scheduling.StickySessionWaitTimeout = 30 * time.Second

	concurrencyCache := stubConcurrencyCache{
		acquireResults: map[int64]bool{
			21: false, // previous_response 命中的账号繁忙
			22: true,  // 次优账号可用（若回退会命中）
		},
		waitCounts: map[int64]int{
			21: 999,
		},
	}

	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: accounts},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(concurrencyCache),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_busy", 21, time.Hour))

	selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, "resp_prev_busy", "gpt-5.1", nil, false)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.NotNil(t, selection.Account)
	require.Equal(t, int64(21), selection.Account.ID, "busy previous_response sticky account should remain selected")
	require.False(t, selection.Acquired)
	require.NotNil(t, selection.WaitPlan)
	require.Equal(t, int64(21), selection.WaitPlan.AccountID)
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_CapabilityMismatchKeepsSticky(t *testing.T) {
	ctx := context.Background()
	groupID := int64(25)
	account := Account{
		ID:          31,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"openai_capabilities": []any{"chat_completions"},
		},
		Extra: map[string]any{
			"openai_apikey_responses_websockets_v2_enabled": true,
		},
	}
	cache := &stubGatewayCache{}
	store := NewOpenAIWSStateStore(cache)
	cfg := newOpenAIWSV2TestConfig()
	svc := &OpenAIGatewayService{
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{account}},
		cache:              cache,
		cfg:                cfg,
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
		openaiWSStateStore: store,
	}

	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_prev_capability", account.ID, time.Hour))

	selection, err := svc.selectAccountByPreviousResponseIDForCapability(
		ctx,
		&groupID,
		"resp_prev_capability",
		"text-embedding-3-small",
		nil,
		OpenAIEndpointCapabilityEmbeddings,
		false,
	)
	require.NoError(t, err)
	require.Nil(t, selection)
	boundAccountID, getErr := store.GetResponseAccount(ctx, groupID, "resp_prev_capability")
	require.NoError(t, getErr)
	require.Equal(t, account.ID, boundAccountID)
}

type previousResponseGroupRepo struct {
	GroupRepository
	group *Group
	err   error
}

func (r previousResponseGroupRepo) GetByID(context.Context, int64) (*Group, error) {
	return r.group, r.err
}

func (r previousResponseGroupRepo) GetByIDLite(context.Context, int64) (*Group, error) {
	return r.group, r.err
}

func TestOpenAIGatewayService_SelectAccountByPreviousResponseID_HonorsFreshGroupAndPrivacy(t *testing.T) {
	ctx := context.Background()
	groupID := int64(26)

	tests := []struct {
		name          string
		accountType   string
		freshGroupIDs []int64
		privacyMode   string
		group         *Group
		groupErr      error
		wantSelection bool
	}{
		{
			name:          "API key does not require provider privacy setting",
			accountType:   AccountTypeAPIKey,
			freshGroupIDs: []int64{groupID},
			group:         &Group{ID: groupID, RequirePrivacySet: true},
			wantSelection: true,
		},
		{
			name:          "account moved to another group",
			accountType:   AccountTypeAPIKey,
			freshGroupIDs: []int64{groupID + 1},
			group:         &Group{ID: groupID},
		},
		{
			name:          "privacy setting no longer satisfies group",
			accountType:   AccountTypeOAuth,
			freshGroupIDs: []int64{groupID},
			group:         &Group{ID: groupID, RequirePrivacySet: true},
		},
		{
			name:          "group policy lookup error fails closed",
			accountType:   AccountTypeOAuth,
			freshGroupIDs: []int64{groupID},
			privacyMode:   PrivacyModeTrainingOff,
			groupErr:      context.DeadlineExceeded,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stale := &Account{
				ID:          41,
				Platform:    PlatformOpenAI,
				Type:        tc.accountType,
				Status:      StatusActive,
				Schedulable: true,
				Concurrency: 1,
				GroupIDs:    []int64{groupID},
				Extra: map[string]any{
					"openai_apikey_responses_websockets_v2_enabled": true,
					"responses_websockets_v2_enabled":               true,
					"privacy_mode":                                  PrivacyModeTrainingOff,
				},
			}
			fresh := *stale
			fresh.GroupIDs = tc.freshGroupIDs
			fresh.Extra = map[string]any{
				"openai_apikey_responses_websockets_v2_enabled": true,
				"responses_websockets_v2_enabled":               true,
			}
			if tc.privacyMode != "" {
				fresh.Extra["privacy_mode"] = tc.privacyMode
			}

			repo := stubOpenAIAccountRepo{accounts: []Account{fresh}}
			cache := &stubGatewayCache{}
			store := NewOpenAIWSStateStore(cache)
			cfg := newOpenAIWSV2TestConfig()
			snapshot := &SchedulerSnapshotService{
				cache:     &openAISnapshotCacheStub{accountsByID: map[int64]*Account{stale.ID: stale}},
				groupRepo: previousResponseGroupRepo{group: tc.group, err: tc.groupErr},
			}
			svc := &OpenAIGatewayService{
				accountRepo:        repo,
				cache:              cache,
				cfg:                cfg,
				concurrencyService: NewConcurrencyService(stubConcurrencyCache{}),
				openaiWSStateStore: store,
				schedulerSnapshot:  snapshot,
			}
			responseID := "resp_prev_fresh_policy"
			require.NoError(t, store.BindResponseAccount(ctx, groupID, responseID, stale.ID, time.Hour))

			selection, err := svc.SelectAccountByPreviousResponseID(ctx, &groupID, responseID, "gpt-5.1", nil, false)
			require.NoError(t, err)
			if tc.wantSelection {
				require.NotNil(t, selection)
				require.Equal(t, stale.ID, selection.Account.ID)
				if selection.ReleaseFunc != nil {
					selection.ReleaseFunc()
				}
			} else {
				require.Nil(t, selection)
			}

			boundAccountID, getErr := store.GetResponseAccount(ctx, groupID, responseID)
			require.NoError(t, getErr)
			require.Equal(t, stale.ID, boundAccountID, "policy misses must preserve the response binding")
		})
	}
}

func newOpenAIWSV2TestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Gateway.OpenAIWS.Enabled = true
	cfg.Gateway.OpenAIWS.OAuthEnabled = true
	cfg.Gateway.OpenAIWS.APIKeyEnabled = true
	cfg.Gateway.OpenAIWS.ResponsesWebsocketsV2 = true
	cfg.Gateway.OpenAIWS.StickyResponseIDTTLSeconds = 3600
	return cfg
}

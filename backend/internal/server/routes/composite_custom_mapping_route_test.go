package routes

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type compositeMappingRouteUpstream struct {
	mu       sync.Mutex
	lastBody []byte
	calls    int
}

func (u *compositeMappingRouteUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	u.mu.Lock()
	u.lastBody = append([]byte(nil), body...)
	u.calls++
	u.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(bytes.NewReader([]byte(
			`{"id":"composite-route-mapping","object":"chat.completion","model":"provider-b","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`,
		))),
	}, nil
}

func (u *compositeMappingRouteUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, accountConcurrency)
}

func (u *compositeMappingRouteUpstream) requestBody() ([]byte, int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return append([]byte(nil), u.lastBody...), u.calls
}

type compositeMappingRouteAccountRepo struct {
	service.AccountRepository
	accounts []service.Account
}

func (r *compositeMappingRouteAccountRepo) all() []service.Account {
	if len(r.accounts) > 0 {
		return r.accounts
	}
	return nil
}

func (r *compositeMappingRouteAccountRepo) ListSchedulableByPlatform(_ context.Context, platform string) ([]service.Account, error) {
	var out []service.Account
	for _, account := range r.all() {
		if account.Platform == platform && account.IsSchedulable() {
			out = append(out, account)
		}
	}
	return out, nil
}

func (r *compositeMappingRouteAccountRepo) ListSchedulableByGroupIDAndPlatform(ctx context.Context, _ int64, platform string) ([]service.Account, error) {
	return r.ListSchedulableByPlatform(ctx, platform)
}

func (r *compositeMappingRouteAccountRepo) ListSchedulableUngroupedByPlatform(ctx context.Context, platform string) ([]service.Account, error) {
	return r.ListSchedulableByPlatform(ctx, platform)
}

func (r *compositeMappingRouteAccountRepo) GetByID(_ context.Context, id int64) (*service.Account, error) {
	for _, account := range r.all() {
		if account.ID == id {
			copy := account
			return &copy, nil
		}
	}
	return nil, nil
}

type compositeMappingRouteChannelRepo struct {
	service.ChannelRepository
	channels       []service.Channel
	groupPlatforms map[int64]string
}

func (r *compositeMappingRouteChannelRepo) ListAll(context.Context) ([]service.Channel, error) {
	result := make([]service.Channel, 0, len(r.channels))
	for i := range r.channels {
		result = append(result, *r.channels[i].Clone())
	}
	return result, nil
}

func (r *compositeMappingRouteChannelRepo) GetByID(_ context.Context, id int64) (*service.Channel, error) {
	for i := range r.channels {
		if r.channels[i].ID == id {
			return r.channels[i].Clone(), nil
		}
	}
	return nil, service.ErrChannelNotFound
}

func (r *compositeMappingRouteChannelRepo) Update(context.Context, *service.Channel) error {
	return nil
}

func (r *compositeMappingRouteChannelRepo) GetGroupIDs(_ context.Context, id int64) ([]int64, error) {
	channel, err := r.GetByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return append([]int64(nil), channel.GroupIDs...), nil
}

func (r *compositeMappingRouteChannelRepo) GetGroupPlatforms(_ context.Context, groupIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(groupIDs))
	for _, groupID := range groupIDs {
		if platform := r.groupPlatforms[groupID]; platform != "" {
			result[groupID] = platform
		}
	}
	return result, nil
}

type compositeMappingRouteConcurrencyCache struct{}

func (compositeMappingRouteConcurrencyCache) AcquireAccountSlot(context.Context, int64, int, string) (bool, error) {
	return true, nil
}
func (compositeMappingRouteConcurrencyCache) ReleaseAccountSlot(context.Context, int64, string) error {
	return nil
}
func (compositeMappingRouteConcurrencyCache) GetAccountConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}
func (compositeMappingRouteConcurrencyCache) GetAccountConcurrencyBatch(_ context.Context, ids []int64) (map[int64]int, error) {
	result := make(map[int64]int, len(ids))
	for _, id := range ids {
		result[id] = 0
	}
	return result, nil
}
func (compositeMappingRouteConcurrencyCache) IncrementAccountWaitCount(context.Context, int64, int) (bool, error) {
	return true, nil
}
func (compositeMappingRouteConcurrencyCache) DecrementAccountWaitCount(context.Context, int64) error {
	return nil
}
func (compositeMappingRouteConcurrencyCache) GetAccountWaitingCount(context.Context, int64) (int, error) {
	return 0, nil
}
func (compositeMappingRouteConcurrencyCache) AcquireUserSlot(context.Context, int64, int, string) (bool, error) {
	return true, nil
}
func (compositeMappingRouteConcurrencyCache) ReleaseUserSlot(context.Context, int64, string) error {
	return nil
}
func (compositeMappingRouteConcurrencyCache) GetUserConcurrency(context.Context, int64) (int, error) {
	return 0, nil
}
func (compositeMappingRouteConcurrencyCache) IncrementWaitCount(context.Context, int64, int) (bool, error) {
	return true, nil
}
func (compositeMappingRouteConcurrencyCache) DecrementWaitCount(context.Context, int64) error {
	return nil
}
func (compositeMappingRouteConcurrencyCache) GetAccountsLoadBatch(context.Context, []service.AccountWithConcurrency) (map[int64]*service.AccountLoadInfo, error) {
	return map[int64]*service.AccountLoadInfo{}, nil
}
func (compositeMappingRouteConcurrencyCache) GetUsersLoadBatch(context.Context, []service.UserWithConcurrency) (map[int64]*service.UserLoadInfo, error) {
	return map[int64]*service.UserLoadInfo{}, nil
}
func (compositeMappingRouteConcurrencyCache) CleanupExpiredAccountSlots(context.Context, int64) error {
	return nil
}
func (compositeMappingRouteConcurrencyCache) CleanupExpiredAccountSlotKeys(context.Context) error {
	return nil
}
func (compositeMappingRouteConcurrencyCache) CleanupStaleProcessSlots(context.Context, string) error {
	return nil
}

func newCompositeMappingRouteRouter(t *testing.T) (*gin.Engine, *compositeMappingRouteUpstream) {
	return newCompositeMappingRouteRouterWithMapping(t, map[string]map[string]string{
		service.PlatformCustom: {"public-a": "provider-b"},
	})
}

func newCompositeMappingRouteRouterWithMapping(t *testing.T, modelMapping map[string]map[string]string) (*gin.Engine, *compositeMappingRouteUpstream) {
	return newCompositeMappingRouteRouterWithAccountMapping(t, modelMapping, nil)
}

func newCompositeMappingRouteRouterWithAccountMapping(t *testing.T, modelMapping map[string]map[string]string, accountMapping map[string]string) (*gin.Engine, *compositeMappingRouteUpstream) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	groupID := int64(7311)
	credentials := map[string]any{
		"api_key":  "sk-composite-route-test",
		"base_url": "http://custom-upstream.test",
	}
	if len(accountMapping) > 0 {
		rawMapping := make(map[string]any, len(accountMapping))
		for source, target := range accountMapping {
			rawMapping[source] = target
		}
		credentials["model_mapping"] = rawMapping
	}
	accountRepo := &compositeMappingRouteAccountRepo{accounts: []service.Account{{
		ID:          7312,
		Name:        "composite-route-upstream",
		Platform:    service.PlatformCustom,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: credentials,
	}}}
	channelRepo := &compositeMappingRouteChannelRepo{
		channels: []service.Channel{{
			ID:           7313,
			Name:         "composite-route-channel",
			Status:       service.StatusActive,
			GroupIDs:     []int64{groupID},
			ModelMapping: modelMapping,
		}},
		groupPlatforms: map[int64]string{groupID: service.PlatformComposite},
	}
	channelService := service.NewChannelService(channelRepo, nil, nil, nil)

	cfg := &config.Config{RunMode: config.RunModeSimple}
	cfg.Gateway.MaxBodySize = 1024 * 1024
	cfg.Gateway.TextMaxBodySize = 1024 * 1024
	cfg.Default.RateMultiplier = 1
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	upstream := &compositeMappingRouteUpstream{}
	billingCache := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
	concurrencyService := service.NewConcurrencyService(compositeMappingRouteConcurrencyCache{})
	gateway := service.NewOpenAIGatewayService(
		accountRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		cfg,
		nil,
		concurrencyService,
		service.NewBillingService(cfg, nil),
		nil,
		billingCache,
		upstream,
		&service.DeferredService{},
		nil,
		nil,
		nil,
		channelService,
		nil,
		nil,
		nil,
	)
	apiKeyService := service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, cfg)
	openAIHandler := handler.NewOpenAIGatewayHandler(
		gateway,
		concurrencyService,
		billingCache,
		apiKeyService,
		nil,
		nil,
		nil,
		nil,
		cfg,
	)

	router := gin.New()
	apiKeyAuth := servermiddleware.APIKeyAuthMiddleware(func(c *gin.Context) {
		c.Set(string(servermiddleware.ContextKeyAPIKey), &service.APIKey{
			ID:      7314,
			GroupID: &groupID,
			User:    &service.User{ID: 7315, Status: service.StatusActive},
			Group:   &service.Group{ID: groupID, Platform: service.PlatformComposite, Status: service.StatusActive},
		})
		c.Set(string(servermiddleware.ContextKeyUser), servermiddleware.AuthSubject{UserID: 7315, Concurrency: 1})
		c.Next()
	})
	RegisterGatewayRoutes(
		router,
		&handler.Handlers{
			Gateway:       &handler.GatewayHandler{},
			OpenAIGateway: openAIHandler,
			AsyncImage:    handler.NewAsyncImageHandler(nil, openAIHandler),
		},
		apiKeyAuth,
		apiKeyService,
		nil,
		nil,
		nil,
		nil,
		cfg,
	)
	t.Cleanup(billingCache.Stop)
	return router, upstream
}

func TestCompositeGatewayRouteAppliesChannelModelMappingBeforeCustomUpstream(t *testing.T) {
	router, upstream := newCompositeMappingRouteRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(
		`{"model":"public-a","messages":[{"role":"user","content":"ping"}],"stream":false}`,
	))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	body, calls := upstream.requestBody()
	require.Equal(t, 1, calls)
	require.JSONEq(t, `{"model":"provider-b","messages":[{"role":"user","content":"ping"}],"stream":false}`, string(body))
}

func TestCompositeGatewayRouteReadsLegacyCompositeChannelMapping(t *testing.T) {
	router, upstream := newCompositeMappingRouteRouterWithMapping(t, map[string]map[string]string{
		service.PlatformComposite: {"public-a": "provider-b"},
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(
		`{"model":"public-a","messages":[{"role":"user","content":"ping"}],"stream":false}`,
	))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	body, calls := upstream.requestBody()
	require.Equal(t, 1, calls)
	require.JSONEq(t, `{"model":"provider-b","messages":[{"role":"user","content":"ping"}],"stream":false}`, string(body))
}

func TestCompositeGatewayRouteAppliesCustomAccountModelMapping(t *testing.T) {
	router, upstream := newCompositeMappingRouteRouterWithAccountMapping(t, nil, map[string]string{
		"public-a": "provider-b",
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(
		`{"model":"public-a","messages":[{"role":"user","content":"ping"}],"stream":false}`,
	))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	body, calls := upstream.requestBody()
	require.Equal(t, 1, calls)
	require.JSONEq(t, `{"model":"provider-b","messages":[{"role":"user","content":"ping"}],"stream":false}`, string(body))
}

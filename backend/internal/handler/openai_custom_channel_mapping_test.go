package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// customChannelMappingHTTPUpstream records the request body while returning a
// minimal Chat Completions response that the transparent Custom path can copy.
type customChannelMappingHTTPUpstream struct {
	service.HTTPUpstream
	mu        sync.Mutex
	lastBody  []byte
	callCount int
	accountID int64
	repo      *openAIWSUsageHandlerAccountRepoStub
}

func (u *customChannelMappingHTTPUpstream) Do(req *http.Request, _ string, accountID int64, _ int) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	u.mu.Lock()
	u.lastBody = append([]byte(nil), body...)
	u.callCount++
	u.accountID = accountID
	u.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(bytes.NewReader([]byte(
			`{"id":"chatcmpl-custom-mapping","object":"chat.completion","model":"k3","choices":[{"index":0,"message":{"role":"assistant","content":"4"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`,
		))),
	}, nil
}

func (u *customChannelMappingHTTPUpstream) requestBody() ([]byte, int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	return append([]byte(nil), u.lastBody...), u.callCount
}

func (u *customChannelMappingHTTPUpstream) selectedAccountID() int64 {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.accountID
}

func newCustomChannelMappingHandler(t *testing.T, mapping map[string]string, accountMappings ...map[string]string) (*OpenAIGatewayHandler, *customChannelMappingHTTPUpstream) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	groupID := int64(7311)
	credentials := map[string]any{
		"api_key":  "sk-custom-monitor-test",
		"base_url": "http://custom-upstream.test",
	}
	if len(accountMappings) > 0 && len(accountMappings[0]) > 0 {
		modelMapping := make(map[string]any, len(accountMappings[0]))
		for source, target := range accountMappings[0] {
			modelMapping[source] = target
		}
		credentials["model_mapping"] = modelMapping
	}
	account := service.Account{
		ID:          7312,
		Name:        "custom-kimi-upstream",
		Platform:    service.PlatformCustom,
		Type:        service.AccountTypeAPIKey,
		Status:      service.StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: credentials,
	}
	accountRepo := &openAIWSUsageHandlerAccountRepoStub{account: account}
	channelRepo := &openAIWSUsageHandlerChannelRepoStub{
		channels: []service.Channel{{
			ID:                 7313,
			Name:               "custom-kimi-channel",
			Status:             service.StatusActive,
			GroupIDs:           []int64{groupID},
			ModelMapping:       map[string]map[string]string{service.PlatformCustom: mapping},
			BillingModelSource: service.BillingModelSourceChannelMapped,
		}},
		groupPlatforms: map[int64]string{groupID: service.PlatformComposite},
	}
	channelService := service.NewChannelService(channelRepo, nil, nil, nil)

	cfg := &config.Config{RunMode: config.RunModeSimple}
	cfg.Default.RateMultiplier = 1
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	upstream := &customChannelMappingHTTPUpstream{repo: accountRepo}
	billingCache := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
	concurrencyCache := &concurrencyCacheMock{
		acquireUserSlotFn:    func(context.Context, int64, int, string) (bool, error) { return true, nil },
		acquireAccountSlotFn: func(context.Context, int64, int, string) (bool, error) { return true, nil },
	}
	concurrencyService := service.NewConcurrencyService(concurrencyCache)
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
	h := NewOpenAIGatewayHandler(
		gateway,
		concurrencyService,
		billingCache,
		service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, cfg),
		nil,
		nil,
		nil,
		nil,
		cfg,
	)
	t.Cleanup(billingCache.Stop)
	return h, upstream
}

func runCustomChannelMappingChat(t *testing.T, h *OpenAIGatewayHandler) *httptest.ResponseRecorder {
	t.Helper()
	groupID := int64(7311)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(
		`{"model":"kimi-k3","messages":[{"role":"user","content":"2+2?"}],"stream":false}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
		ID:      7314,
		GroupID: &groupID,
		User:    &service.User{ID: 7315, Status: service.StatusActive},
		Group:   &service.Group{ID: groupID, Platform: service.PlatformComposite, Status: service.StatusActive},
	})
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7315, Concurrency: 1})
	h.ChatCompletions(c)
	return recorder
}

func TestCustomGroupChatAppliesChannelModelMapping(t *testing.T) {
	h, upstream := newCustomChannelMappingHandler(t, map[string]string{"kimi-k3": "k3"})

	recorder := runCustomChannelMappingChat(t, h)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	body, calls := upstream.requestBody()
	require.Equal(t, 1, calls)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, "k3", decoded["model"], "Custom channel aliases must be mapped before transparent forwarding")
}

func TestCustomGroupChatWithoutChannelMappingPreservesModel(t *testing.T) {
	h, upstream := newCustomChannelMappingHandler(t, nil)

	recorder := runCustomChannelMappingChat(t, h)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	body, _ := upstream.requestBody()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, "kimi-k3", decoded["model"], "Unmapped Custom requests must remain transparent")
}

func TestCustomGroupChatChainsChannelAndAccountModelMappings(t *testing.T) {
	h, upstream := newCustomChannelMappingHandler(
		t,
		map[string]string{"kimi-k3": "channel-k3"},
		map[string]string{"channel-k3": "provider-k3"},
	)

	recorder := runCustomChannelMappingChat(t, h)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	body, calls := upstream.requestBody()
	require.Equal(t, 1, calls)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, "provider-k3", decoded["model"], "account mapping must apply after the channel alias")
}

func TestCustomGroupChatAppliesAccountModelMappingForMonitorAlias(t *testing.T) {
	h, upstream := newCustomChannelMappingHandler(
		t,
		nil,
		map[string]string{"monitor-alias": "provider-model"},
	)

	recorder := runCustomChannelMappingChatWithBody(t, h, `{"model":"monitor-alias","messages":[{"role":"user","content":"2+2?"}],"stream":false}`)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	body, calls := upstream.requestBody()
	require.Equal(t, 1, calls)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, "provider-model", decoded["model"], "Custom account model_mapping must reach the distributor")
}

func runCustomChannelMappingChatWithBody(t *testing.T, h *OpenAIGatewayHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	groupID := int64(7311)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
		ID:      7314,
		GroupID: &groupID,
		User:    &service.User{ID: 7315, Status: service.StatusActive},
		Group:   &service.Group{ID: groupID, Platform: service.PlatformComposite, Status: service.StatusActive},
	})
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7315, Concurrency: 1})
	h.ChatCompletions(c)
	return recorder
}

func TestCustomGroupChatSchedulesAgainstChannelMappedModel(t *testing.T) {
	h, upstream := newCustomChannelMappingHandler(t, map[string]string{"public-a": "provider-b"})
	repo := upstream.repo
	transparent := repo.account
	transparent.ID = 7316
	transparent.Priority = 0
	mapped := repo.account
	mapped.ID = 7317
	mapped.Priority = 10
	mapped.Credentials = map[string]any{
		"api_key":       "sk-mapped",
		"base_url":      "http://custom-upstream.test",
		"model_mapping": map[string]any{"provider-b": "provider-c"},
	}
	repo.accounts = []service.Account{transparent, mapped}

	recorder := runCustomChannelMappingChatWithBody(t, h, `{"model":"public-a","messages":[{"role":"user","content":"ping"}],"stream":false}`)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, mapped.ID, upstream.selectedAccountID(),
		"Custom scheduling must evaluate the channel target so the account mapped for provider-b wins")
	body, _ := upstream.requestBody()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, "provider-c", decoded["model"])
}

func TestCustomGroupChatPrefersAccountMappedAliasOverTransparentAccount(t *testing.T) {
	h, upstream := newCustomChannelMappingHandler(t, nil)
	repo := upstream.repo
	transparent := repo.account
	transparent.ID = 7316
	transparent.Priority = 0
	mapped := repo.account
	mapped.ID = 7317
	mapped.Priority = 10
	mapped.Credentials = map[string]any{
		"api_key":       "sk-mapped",
		"base_url":      "http://custom-upstream.test",
		"model_mapping": map[string]any{"monitor-alias": "provider-model"},
	}
	repo.accounts = []service.Account{transparent, mapped}

	recorder := runCustomChannelMappingChatWithBody(t, h, `{"model":"monitor-alias","messages":[{"role":"user","content":"ping"}],"stream":false}`)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, mapped.ID, upstream.selectedAccountID(),
		"a Custom account with the monitor alias must be selected before transparent fallbacks")
	body, _ := upstream.requestBody()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, "provider-model", decoded["model"])
}

func TestOpenAICompatibleSelectionModelUsesChannelTargetOnlyForCustom(t *testing.T) {
	mapping := service.ChannelMappingResult{Mapped: true, MappedModel: "provider-b"}

	require.Equal(t, "provider-b", openAICompatibleSelectionModel(service.PlatformCustom, "public-a", mapping))
	require.Equal(t, "provider-b", openAICompatibleSelectionModel(service.PlatformComposite, "public-a", mapping))
	require.Equal(t, "public-a", openAICompatibleSelectionModel(service.PlatformOpenAI, "public-a", mapping))
	require.Equal(t, "public-a", openAICompatibleSelectionModel(service.PlatformGrok, "public-a", mapping))
	require.Equal(t, "public-a", openAICompatibleSelectionModel(service.PlatformCustom, "public-a", service.ChannelMappingResult{}))

	account := &service.Account{
		Platform: service.PlatformCustom,
		Credentials: map[string]any{
			"model_mapping": map[string]any{"provider-b": "provider-c"},
		},
	}
	selectionModel := openAICompatibleSelectionModel(service.PlatformCustom, "public-a", mapping)
	require.Equal(t, "provider-c", account.GetMappedModel(selectionModel),
		"schedule state must use the channel target before applying account mapping")
}

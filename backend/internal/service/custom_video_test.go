//go:build unit

package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestForwardCustomVideoGenerationUsesStandardAsyncEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{
		"model":"seedance-2-5-720p",
		"prompt":"turn and smile",
		"duration":8,
		"ratio":"16:9",
		"referenceImages":["https://assets.example.test/input.jpg"],
		"referenceVideos":["https://assets.example.test/guide.mp4"],
		"referenceAudios":["https://assets.example.test/music.mp3"]
	}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	account := customVideoTestAccount()
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"id":"task-standard-id","request_id":"legacy-request-id","task_id":"legacy-task-id","status":"queued"}`,
		)),
	}}
	cache := &videoBindingCache{}
	svc := &OpenAIGatewayService{
		cfg:          customVideoTestConfig(),
		cache:        cache,
		httpUpstream: upstream,
	}

	result, err := svc.ForwardGrokMedia(
		context.Background(), c, account, GrokMediaEndpointVideosGenerations, "", body, "application/json",
	)
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, http.MethodPost, upstream.lastReq.Method)
	require.Equal(t, "https://custom.example.test/proxy/v1/videos", upstream.lastReq.URL.String())
	require.NotContains(t, upstream.lastReq.URL.Path, "/videos/generations")
	require.Equal(t, "Bearer custom-secret", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "application/json", upstream.lastReq.Header.Get("Content-Type"))
	require.JSONEq(t, string(body), string(upstream.lastBody))

	// The standard Sora response id wins over legacy aliases and is the value
	// persisted for subsequent status/content requests.
	require.Equal(t, "task-standard-id", result.ResponseID)
	require.Equal(t, "seedance-2-5-720p", result.BillingModel)
	require.Equal(t, 1, result.VideoCount)
	require.Equal(t, VideoBillingResolution720P, result.VideoResolution)
	require.Equal(t, 8, result.VideoDurationSeconds)
	require.InDelta(t, 1.4, result.VideoPriceMultiplier, 1e-12)
	require.Empty(t, recorder.Body.Bytes(), "creation success must remain buffered until task metadata is durable")

	groupID := int64(17)
	require.NoError(t, svc.BindGrokMediaVideoRequestAccount(
		context.Background(), &groupID, result.ResponseID, 21, 31, account.ID,
	))
	boundID, err := svc.ResolveGrokMediaVideoRequestAccount(
		context.Background(), &groupID, "task-standard-id", 21, 31,
	)
	require.NoError(t, err)
	require.Equal(t, account.ID, boundID)
	require.GreaterOrEqual(t, cache.lastTTL, 7*24*time.Hour)
	require.True(t, svc.WriteBufferedGrokMediaResponse(c, result))
	require.JSONEq(t, `{"id":"task-standard-id","request_id":"legacy-request-id","task_id":"legacy-task-id","status":"queued"}`, recorder.Body.String())
}

func TestForwardCustomVideoGenerationPreservesRequestedClipCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	body := []byte(`{"model":"videos-standard-720p","prompt":"two clips","n":2,"duration":4}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"id":"task-two-clips","status":"queued"}`)),
	}}
	svc := &OpenAIGatewayService{
		cfg:          customVideoTestConfig(),
		httpUpstream: upstream,
	}

	result, err := svc.ForwardGrokMedia(
		context.Background(), c, customVideoTestAccount(), GrokMediaEndpointVideosGenerations, "", body, "application/json",
	)
	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.JSONEq(t, string(body), string(upstream.lastBody))
	require.Equal(t, 2, result.VideoCount)
	require.Equal(t, VideoBillingResolution720P, result.VideoResolution)
	require.Equal(t, 4, result.VideoDurationSeconds)
}

func TestParseGrokMediaRequestAcceptsStringClipCount(t *testing.T) {
	info := ParseGrokMediaRequest("application/json", []byte(`{"model":"videos-standard-720p","n":"2"}`))
	require.Equal(t, 2, info.N)
}

func TestForwardCustomVideoStatusUsesBoundTaskURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task-abc123", nil)

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"id":"task-abc123","status":"processing"}`)),
	}}
	svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardGrokMedia(
		context.Background(), c, customVideoTestAccount(), GrokMediaEndpointVideoStatus, "task-abc123", nil, "",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, http.MethodGet, upstream.lastReq.Method)
	require.Equal(t, "https://custom.example.test/proxy/v1/videos/task-abc123", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer custom-secret", upstream.lastReq.Header.Get("Authorization"))
	require.Empty(t, upstream.lastBody)
	require.Empty(t, recorder.Body.Bytes())
	require.True(t, svc.WriteBufferedGrokMediaResponse(c, result))
	require.JSONEq(t, `{"id":"task-abc123","status":"processing"}`, recorder.Body.String())
}

func TestForwardCustomVideoStatusPreservesSignedURLs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task-abc123", nil)
	responseBody := `{"id":"task-abc123","status":"completed","url":"https://cdn.example.test/task-abc123.mp4?sig=one","video":{"url":"https://cdn.example.test/task-abc123-alt.mp4?sig=two"}}`

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}}
	svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardGrokMedia(
		context.Background(), c, customVideoTestAccount(), GrokMediaEndpointVideoStatus, "task-abc123", nil, "",
	)
	require.NoError(t, err)
	require.Empty(t, recorder.Body.Bytes())
	require.True(t, svc.WriteBufferedGrokMediaResponse(c, result))
	require.JSONEq(t, responseBody, recorder.Body.String())
}

func TestForwardCustomVideoContentFetchesContentDirectly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/videos/task-abc123/content", nil)

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode:    http.StatusOK,
		Header:        http.Header{"Content-Type": []string{"video/mp4"}},
		Body:          io.NopCloser(strings.NewReader("mp4-binary-content")),
		ContentLength: int64(len("mp4-binary-content")),
	}}
	svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}

	result, err := svc.ForwardGrokMedia(
		context.Background(), c, customVideoTestAccount(), GrokMediaEndpointVideoContent, "task-abc123", nil, "",
	)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Custom content retrieval is one direct request. Grok's preliminary status
	// lookup must not run for this platform.
	require.Len(t, upstream.requests, 1)
	require.Equal(t, http.MethodGet, upstream.lastReq.Method)
	require.Equal(t, "https://custom.example.test/proxy/v1/videos/task-abc123/content", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer custom-secret", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "*/*", upstream.lastReq.Header.Get("Accept"))
	require.Equal(t, "video/mp4", recorder.Header().Get("Content-Type"))
	require.Equal(t, "mp4-binary-content", recorder.Body.String())
}

func TestCustomVideoBillingResolutionFromModelSuffix(t *testing.T) {
	tests := []struct {
		model string
		want  string
	}{
		{model: "videos-mini-480p", want: VideoBillingResolution480P},
		{model: "videos-standard-720p", want: VideoBillingResolution720P},
		{model: "seedance-2-5-1080p", want: VideoBillingResolution1080P},
		{model: "provider/wan-3.0-4k", want: VideoBillingResolution4K},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			require.Equal(t, tt.want, NormalizeVideoBillingResolutionForModelOrDefault(tt.model, "", ""))
		})
	}
}

func TestCustomVideoBillingDurationUsesModelLimit(t *testing.T) {
	require.Equal(t, 30, NormalizeVideoBillingDurationSecondsForModelOrDefault("wan-3.0-720p", 30))
	require.Equal(t, 30, NormalizeVideoBillingDurationSecondsForModelOrDefault("provider/wan-3.0-pro-4k", 60))
	require.Equal(t, 15, NormalizeVideoBillingDurationSecondsForModelOrDefault("videos-mini-480p", 30))
	require.Equal(t, 15, NormalizeVideoBillingDurationSecondsForModelOrDefault("grok-imagine-video", 30))
}

func TestCustomVideoReferencePriceMultiplier(t *testing.T) {
	require.InDelta(t, 1.3, VideoReferencePriceMultiplier("videos-standard-720p", true), 1e-12)
	require.InDelta(t, 1.4, VideoReferencePriceMultiplier("provider/seedance-2-5-720p", true), 1e-12)
	require.InDelta(t, 1.0, VideoReferencePriceMultiplier("wan-3.0-720p", true), 1e-12)
	require.InDelta(t, 1.0, VideoReferencePriceMultiplier("videos-standard-720p", false), 1e-12)
}

func TestCustomVideoPreservesUpstreamErrorsAndMarksFailoverErrorsRaw(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"videos-mini-480p","prompt":"hello","duration":4}`)

	t.Run("non failover", func(t *testing.T) {
		responseBody := []byte(`{"error":{"type":"vendor_validation","message":"bad video input"},"vendor_code":19}`)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
		upstream := &httpUpstreamRecorder{resp: &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Header:     http.Header{"Content-Type": []string{"application/problem+json"}},
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}}
		svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}

		result, err := svc.ForwardGrokMedia(
			context.Background(), c, customVideoTestAccount(), GrokMediaEndpointVideosGenerations, "", body, "application/json",
		)

		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
		require.Equal(t, "application/problem+json", recorder.Header().Get("Content-Type"))
		require.Equal(t, responseBody, recorder.Body.Bytes())
	})

	t.Run("failover", func(t *testing.T) {
		responseBody := []byte(`{"error":{"type":"vendor_capacity","message":"busy"}}`)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", bytes.NewReader(body))
		upstream := &httpUpstreamRecorder{resp: &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		}}
		svc := &OpenAIGatewayService{cfg: customVideoTestConfig(), httpUpstream: upstream}

		result, err := svc.ForwardGrokMedia(
			context.Background(), c, customVideoTestAccount(), GrokMediaEndpointVideosGenerations, "", body, "application/json",
		)

		require.Nil(t, result)
		var failoverErr *UpstreamFailoverError
		require.True(t, errors.As(err, &failoverErr))
		require.True(t, failoverErr.RawResponsePassthrough)
		require.Equal(t, responseBody, failoverErr.ResponseBody)
		require.Empty(t, recorder.Body.Bytes())
	})
}

func TestCustomVideoChannelBillingModeControlsUnits(t *testing.T) {
	const (
		model      = "videos-standard-720p"
		unitPrice  = 0.25
		videoCount = 2
		duration   = 4
		groupID    = int64(710)
	)
	referenceMultiplier := VideoReferencePriceMultiplier(model, true)
	require.InDelta(t, 1.3, referenceMultiplier, 1e-12)

	tests := []struct {
		name       string
		mode       BillingMode
		wantTotal  float64
		wantActual float64
	}{
		{
			name:       "video mode bills each generated second",
			mode:       BillingModeVideo,
			wantTotal:  unitPrice * videoCount * duration,
			wantActual: unitPrice * videoCount * duration * 1.3,
		},
		{
			name:       "per request mode bills each completed clip",
			mode:       BillingModePerRequest,
			wantTotal:  unitPrice * videoCount,
			wantActual: unitPrice * videoCount * 1.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			billingService := NewBillingService(&config.Config{}, nil)
			resolver := newCustomVideoChannelPricingResolverForTest(groupID, model, tt.mode, unitPrice)
			svc := &OpenAIGatewayService{billingService: billingService, resolver: resolver}
			apiKey := &APIKey{
				GroupID: testInt64Pointer(groupID),
				Group:   &Group{ID: groupID, Platform: PlatformComposite},
			}
			result := &OpenAIForwardResult{
				VideoCount:           videoCount,
				VideoResolution:      VideoBillingResolution720P,
				VideoDurationSeconds: duration,
			}

			cost := svc.calculateOpenAIVideoCost(
				context.Background(), model, apiKey, result, referenceMultiplier,
			)
			require.NotNil(t, cost)
			require.Equal(t, string(tt.mode), cost.BillingMode)
			require.InDelta(t, tt.wantTotal, cost.TotalCost, 1e-12)
			require.InDelta(t, tt.wantActual, cost.ActualCost, 1e-12)
		})
	}
}

func TestValidateCustomVideoPricingRejectsUnpricedModels(t *testing.T) {
	const groupID = int64(711)
	apiKey := &APIKey{
		GroupID: testInt64Pointer(groupID),
		Group:   &Group{ID: groupID, Platform: PlatformComposite},
	}
	svc := &OpenAIGatewayService{}

	err := svc.ValidateCustomVideoPricing(context.Background(), apiKey, "videos-mini-720p", VideoBillingResolution720P)
	require.ErrorIs(t, err, ErrCustomVideoPricingUnavailable)
}

func TestValidateCustomVideoPricingAcceptsChannelAndGroupPrices(t *testing.T) {
	const (
		groupID = int64(712)
		model   = "videos-mini-720p"
	)
	apiKey := &APIKey{
		GroupID: testInt64Pointer(groupID),
		Group:   &Group{ID: groupID, Platform: PlatformComposite},
	}
	svc := &OpenAIGatewayService{
		resolver: newCustomVideoChannelPricingResolverForTest(groupID, model, BillingModeVideo, 0.25),
	}
	require.NoError(t, svc.ValidateCustomVideoPricing(context.Background(), apiKey, model, VideoBillingResolution720P))

	groupPrice := 0.2
	apiKey.Group.VideoPrice720P = &groupPrice
	svc.resolver = nil
	require.NoError(t, svc.ValidateCustomVideoPricing(context.Background(), apiKey, model, VideoBillingResolution720P))
}

func TestValidateCustomVideoPricingDoesNotUse480PGroupPriceFor4K(t *testing.T) {
	const groupID = int64(713)
	groupPrice := 0.2
	apiKey := &APIKey{
		GroupID: testInt64Pointer(groupID),
		Group: &Group{
			ID:             groupID,
			Platform:       PlatformComposite,
			VideoPrice480P: &groupPrice,
		},
	}
	svc := &OpenAIGatewayService{}

	err := svc.ValidateCustomVideoPricing(context.Background(), apiKey, "videos-ultra-4k", VideoBillingResolution4K)
	require.ErrorIs(t, err, ErrCustomVideoPricingUnavailable)
}

func customVideoTestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Security.URLAllowlist.Enabled = false
	return cfg
}

func customVideoTestAccount() *Account {
	return &Account{
		ID:          901,
		Name:        "custom-video",
		Platform:    PlatformCustom,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "custom-secret",
			"base_url": "https://custom.example.test/proxy/v1",
		},
	}
}

type videoBindingCache struct {
	bindings map[string]int64
	lastTTL  time.Duration
}

func (c *videoBindingCache) GetSessionAccountID(_ context.Context, _ int64, sessionHash string) (int64, error) {
	accountID, ok := c.bindings[sessionHash]
	if !ok {
		return 0, ErrStickySessionNotFound
	}
	return accountID, nil
}

func (c *videoBindingCache) SetSessionAccountID(_ context.Context, _ int64, sessionHash string, accountID int64, ttl time.Duration) error {
	if c.bindings == nil {
		c.bindings = make(map[string]int64)
	}
	c.bindings[sessionHash] = accountID
	c.lastTTL = ttl
	return nil
}

func (c *videoBindingCache) RefreshSessionTTL(_ context.Context, _ int64, _ string, ttl time.Duration) error {
	c.lastTTL = ttl
	return nil
}

func (c *videoBindingCache) DeleteSessionAccountID(_ context.Context, _ int64, sessionHash string) error {
	delete(c.bindings, sessionHash)
	return nil
}

func newCustomVideoChannelPricingResolverForTest(groupID int64, model string, mode BillingMode, price float64) *ModelPricingResolver {
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{
		groupID:  groupID,
		platform: PlatformCustom,
		model:    model,
	}] = &ChannelModelPricing{
		Platform:        PlatformCustom,
		Models:          []string{model},
		BillingMode:     mode,
		PerRequestPrice: testFloat64Pointer(price),
		Intervals: []PricingInterval{{
			TierLabel:       VideoBillingResolution720P,
			PerRequestPrice: testFloat64Pointer(price),
		}},
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformComposite
	cache.loadedAt = time.Now()

	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	return NewModelPricingResolver(channelService, NewBillingService(&config.Config{}, nil))
}

func testInt64Pointer(value int64) *int64 {
	return &value
}

func testFloat64Pointer(value float64) *float64 {
	return &value
}

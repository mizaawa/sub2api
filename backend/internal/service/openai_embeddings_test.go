package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type responseModelAuditSettingRepo struct {
	value string
}

func (r *responseModelAuditSettingRepo) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (r *responseModelAuditSettingRepo) GetValue(_ context.Context, key string) (string, error) {
	if key == SettingKeyResponseModelAuditBypass {
		return r.value, nil
	}
	return "", ErrSettingNotFound
}

func (r *responseModelAuditSettingRepo) Set(context.Context, string, string) error {
	panic("unexpected Set call")
}

func (r *responseModelAuditSettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (r *responseModelAuditSettingRepo) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (r *responseModelAuditSettingRepo) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (r *responseModelAuditSettingRepo) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestBuildOpenAIEmbeddingsURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		base string
		want string
	}{
		{"bare domain", "https://api.openai.com", "https://api.openai.com/v1/embeddings"},
		{"bare /v1", "https://api.openai.com/v1", "https://api.openai.com/v1/embeddings"},
		{"already embeddings", "https://api.openai.com/v1/embeddings", "https://api.openai.com/v1/embeddings"},
		{"third-party versioned path", "https://open.bigmodel.cn/api/paas/v4", "https://open.bigmodel.cn/api/paas/v4/embeddings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, buildOpenAIEmbeddingsURL(tt.base))
		})
	}
}

func TestForwardEmbeddings_APIKeyPassthroughRecordsUsageAndBatchInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reqBody := []byte(`{
		"model":"nowledge-embedding",
		"input":["hello","world"],
		"encoding_format":"float",
		"dimensions":256
	}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")

	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"X-Request-Id": []string{"emb-rid"},
		},
		Body: io.NopCloser(strings.NewReader(`{
			"object":"list",
			"data":[
				{"object":"embedding","index":0,"embedding":[0.1,0.2]},
				{"object":"embedding","index":1,"embedding":[0.3,0.4]}
			],
			"model":"jina-embeddings-v5-text-small",
			"usage":{"prompt_tokens":13,"total_tokens":13}
		}`)),
	}}
	svc := &OpenAIGatewayService{
		cfg:          &config.Config{},
		httpUpstream: upstream,
	}
	account := &Account{
		ID:       42,
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "sk-test",
			"base_url": "https://api.jina.ai",
			"model_mapping": map[string]any{
				"nowledge-embedding": "jina-embeddings-v5-text-small",
			},
		},
	}

	result, err := svc.ForwardEmbeddings(context.Background(), c, account, reqBody, "")

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, result)
	require.Equal(t, "emb-rid", result.RequestID)
	require.Equal(t, "nowledge-embedding", result.Model)
	require.Equal(t, "jina-embeddings-v5-text-small", result.BillingModel)
	require.Equal(t, "jina-embeddings-v5-text-small", result.UpstreamModel)
	require.Equal(t, "jina-embeddings-v5-text-small", result.UpstreamResponseModel)
	require.False(t, result.UpstreamResponseModelConflict)
	require.Equal(t, 13, result.Usage.InputTokens)
	require.Equal(t, 0, result.Usage.OutputTokens)
	require.Equal(t, "https://api.jina.ai/v1/embeddings", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer sk-test", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, "jina-embeddings-v5-text-small", gjson.GetBytes(upstream.lastBody, "model").String())
	require.Equal(t, int64(2), gjson.GetBytes(upstream.lastBody, "input.#").Int())
	require.Equal(t, "hello", gjson.GetBytes(upstream.lastBody, "input.0").String())
	require.Equal(t, "world", gjson.GetBytes(upstream.lastBody, "input.1").String())
	require.Equal(t, "float", gjson.GetBytes(upstream.lastBody, "encoding_format").String())
	require.Equal(t, int64(256), gjson.GetBytes(upstream.lastBody, "dimensions").Int())
	require.Equal(t, "nowledge-embedding", gjson.Get(rec.Body.String(), "model").String())
}

func TestForwardEmbeddings_ResponseModelAuditBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, tc := range []struct {
		name          string
		setting       string
		expectedModel string
	}{
		{name: "disabled preserves upstream runtime model", setting: "false", expectedModel: "runtime-embedding-model"},
		{name: "enabled restores original public model", setting: "true", expectedModel: "public-embedding-model"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"channel-embedding-model","input":"hello"}`)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(body))
			c.Request = c.Request.WithContext(WithRequestedPublicModel(c.Request.Context(), "public-embedding-model"))

			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(
					`{"object":"list","data":[],"model":"runtime-embedding-model","usage":{"prompt_tokens":1,"total_tokens":1}}`,
				)),
			}}
			svc := &OpenAIGatewayService{
				cfg:            &config.Config{},
				httpUpstream:   upstream,
				settingService: NewSettingService(&responseModelAuditSettingRepo{value: tc.setting}, &config.Config{}),
			}
			account := &Account{
				ID:       43,
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"api_key":  "sk-test",
					"base_url": "https://api.example.com",
					"model_mapping": map[string]any{
						"channel-embedding-model": "account-embedding-model",
					},
				},
			}

			result, err := svc.ForwardEmbeddings(c.Request.Context(), c, account, body, "")

			require.NoError(t, err)
			require.Equal(t, tc.expectedModel, gjson.Get(rec.Body.String(), "model").String())
			require.Equal(t, "account-embedding-model", gjson.GetBytes(upstream.lastBody, "model").String())
			require.Equal(t, "runtime-embedding-model", result.UpstreamResponseModel)
			require.False(t, result.UpstreamResponseModelConflict)
		})
	}
}

package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGatewayRoutesCustomGroupUsesOpenAICompatibleHandlers(t *testing.T) {
	router := newGatewayRoutesTestRouter(service.PlatformCustom)
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{method: http.MethodPost, path: "/v1/responses", body: `{"model":"private-model","input":"hi"}`},
		{method: http.MethodPost, path: "/v1/messages", body: `{"model":"private-model","messages":[]}`},
		{method: http.MethodPost, path: "/v1/chat/completions", body: `{"model":"private-model","messages":[]}`},
		{method: http.MethodPost, path: "/v1/embeddings", body: `{"model":"private-model","input":"hi"}`},
		{method: http.MethodPost, path: "/v1/images/generations", body: `{"model":"private-image-model","prompt":"hi"}`},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			require.NotEqual(t, http.StatusNotFound, w.Code, "path=%s should reach an OpenAI-compatible handler", tt.path)
			require.NotContains(t, w.Body.String(), "not supported for this platform")
		})
	}
}

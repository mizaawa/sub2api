package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorCreateRequestDoesNotRequireClientEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/channel-monitors",
		strings.NewReader(`{
			"provider":"custom",
			"group_id":42,
			"primary_model":"test-model",
			"interval_seconds":60
		}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	var request channelMonitorCreateRequest
	require.NoError(t, ctx.ShouldBindJSON(&request))
	require.Empty(t, request.Endpoint)
}

package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCustomDirectGroupUsesOpenAICompatiblePlatform(t *testing.T) {
	apiKey := &service.APIKey{Group: &service.Group{Platform: service.PlatformCustom}}
	require.Equal(t, service.PlatformCustom, openAICompatibleRequestPlatform(context.Background(), apiKey))
	require.Equal(t,
		service.OpenAIEndpointCapabilityResponses,
		openAIResponsesRequiredCapability(true, service.PlatformCustom),
	)

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	require.True(t, openAICompatibleTextTargetAllowed(c, apiKey, "private-model"))
}

func TestCompositeOpenAICompatibleTargetRejectsCustomPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	apiKey := &service.APIKey{Group: &service.Group{Platform: service.PlatformComposite}}

	require.False(t, openAICompatibleTextTargetAllowed(c, apiKey, "custom/private-model"))
}

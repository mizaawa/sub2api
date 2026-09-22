package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIImagesOAuthResponseModelAuditBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, enabled := range []bool{false, true} {
		expectedModel := "runtime-image-model"
		setting := "false"
		name := "disabled"
		if enabled {
			expectedModel = "public-image-model"
			setting = "true"
			name = "enabled"
		}

		t.Run(name+" non-streaming", func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)
			c.Request = c.Request.WithContext(WithRequestedPublicModel(c.Request.Context(), "public-image-model"))
			beginUpstreamResponseModelObservation(c)

			svc := &OpenAIGatewayService{
				cfg:            &config.Config{},
				settingService: NewSettingService(&responseModelAuditSettingRepo{value: setting}, &config.Config{}),
			}
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader(
					"data: {\"type\":\"response.completed\",\"response\":{\"created_at\":1710000001,\"tools\":[{\"type\":\"image_generation\",\"model\":\"runtime-image-model\"}],\"output\":[{\"type\":\"image_generation_call\",\"result\":\"ZmluYWw=\"}]}}\n\n",
				)),
			}

			_, imageCount, _, err := svc.handleOpenAIImagesOAuthNonStreamingResponse(
				resp, c, "b64_json", "account-image-model", "channel-image-model",
			)

			require.NoError(t, err)
			require.Equal(t, 1, imageCount)
			require.Equal(t, expectedModel, gjson.Get(rec.Body.String(), "model").String())
			require.Equal(t, "runtime-image-model", observedUpstreamResponseModel(c))
		})

		t.Run(name+" streaming", func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil)
			c.Request = c.Request.WithContext(WithRequestedPublicModel(c.Request.Context(), "public-image-model"))
			beginUpstreamResponseModelObservation(c)

			svc := &OpenAIGatewayService{
				cfg:            &config.Config{},
				settingService: NewSettingService(&responseModelAuditSettingRepo{value: setting}, &config.Config{}),
			}
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader(
					"data: {\"type\":\"response.created\",\"response\":{\"created_at\":1710000001,\"tools\":[{\"type\":\"image_generation\",\"model\":\"runtime-image-model\"}]}}\n\n" +
						"data: {\"type\":\"response.image_generation_call.partial_image\",\"partial_image_b64\":\"cGFydGlhbA==\",\"partial_image_index\":0}\n\n" +
						"data: {\"type\":\"response.completed\",\"response\":{\"created_at\":1710000001,\"tools\":[{\"type\":\"image_generation\",\"model\":\"runtime-image-model\"}],\"output\":[{\"type\":\"image_generation_call\",\"result\":\"ZmluYWw=\"}]}}\n\n",
				)),
			}

			_, imageCount, _, _, err := svc.handleOpenAIImagesOAuthStreamingResponse(
				resp, c, time.Now(), "b64_json", "image_generation", "account-image-model", "channel-image-model",
			)

			require.NoError(t, err)
			require.Equal(t, 1, imageCount)
			events := parseOpenAIImageTestSSEEvents(rec.Body.String())
			partial, ok := findOpenAIImageTestSSEEvent(events, "image_generation.partial_image")
			require.True(t, ok)
			require.Equal(t, expectedModel, gjson.Get(partial.Data, "model").String())
			completed, ok := findOpenAIImageTestSSEEvent(events, "image_generation.completed")
			require.True(t, ok)
			require.Equal(t, expectedModel, gjson.Get(completed.Data, "model").String())
			require.Equal(t, "runtime-image-model", observedUpstreamResponseModel(c))
		})
	}
}

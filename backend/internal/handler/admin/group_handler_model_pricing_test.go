//go:build unit

package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type modelPricingGroupAdminServiceStub struct {
	service.AdminService
	created *service.CreateGroupInput
	updated *service.UpdateGroupInput
}

func (s *modelPricingGroupAdminServiceStub) CreateGroup(_ context.Context, input *service.CreateGroupInput) (*service.Group, error) {
	s.created = input
	return &service.Group{ID: 9, ModelPricing: input.ModelPricing}, nil
}

func (s *modelPricingGroupAdminServiceStub) UpdateGroup(_ context.Context, _ int64, input *service.UpdateGroupInput) (*service.Group, error) {
	s.updated = input
	group := &service.Group{ID: 9}
	if input.ModelPricing != nil {
		group.ModelPricing = *input.ModelPricing
	}
	return group, nil
}

func TestGroupHandlerModelPricingRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			svc := &modelPricingGroupAdminServiceStub{}
			handler := NewGroupHandler(svc, nil, nil)
			router := gin.New()
			router.POST("/groups/9", handler.Create)
			router.PUT("/groups/:id", handler.Update)
			body := `{"name":"grok","platform":"grok","rate_multiplier":1,"model_pricing":[{"models":["grok-4.7"],"input_price":0.000002,"output_price":0.000006}]}`
			request := httptest.NewRequest(method, "/groups/9", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			var pricing []service.ChannelModelPricing
			if method == http.MethodPost {
				require.NotNil(t, svc.created)
				pricing = svc.created.ModelPricing
			} else {
				require.NotNil(t, svc.updated)
				require.NotNil(t, svc.updated.ModelPricing)
				pricing = *svc.updated.ModelPricing
			}
			require.Len(t, pricing, 1)
			require.Equal(t, []string{"grok-4.7"}, pricing[0].Models)
			require.Equal(t, service.BillingModeToken, pricing[0].BillingMode)
			require.Equal(t, 2e-6, *pricing[0].InputPrice)
			var response struct {
				Data struct {
					ModelPricing []service.ChannelModelPricing `json:"model_pricing"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
			require.Equal(t, pricing, response.Data.ModelPricing)
			require.Contains(t, recorder.Body.String(), `"input_price":0.000002`)
		})
	}
}

func TestGroupHandlerModelPricingOmittedAndClear(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{`{}`, `{"model_pricing":[]}`} {
		t.Run(body, func(t *testing.T) {
			svc := &modelPricingGroupAdminServiceStub{}
			router := gin.New()
			router.PUT("/groups/:id", NewGroupHandler(svc, nil, nil).Update)
			request := httptest.NewRequest(http.MethodPut, "/groups/9", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			require.NotNil(t, svc.updated)
			if body == `{}` {
				require.Nil(t, svc.updated.ModelPricing)
			} else {
				require.NotNil(t, svc.updated.ModelPricing)
				require.Empty(t, *svc.updated.ModelPricing)
			}
		})
	}
}

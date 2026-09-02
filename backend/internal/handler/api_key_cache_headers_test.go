package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSetAPIKeyResponseNoStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	setAPIKeyResponseNoStore(ctx)

	require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
}

func TestAPIKeyHandler_UserScopedResponsesAreNotCached(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &APIKeyHandler{}
	handlers := []struct {
		name string
		fn   func(*gin.Context)
	}{
		{"list", h.List},
		{"get", h.GetByID},
		{"create", h.Create},
		{"update", h.Update},
		{"delete", h.Delete},
		{"available-groups", h.GetAvailableGroups},
		{"group-rates", h.GetUserGroupRates},
	}

	for _, tc := range handlers {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/keys", nil)

			tc.fn(ctx)

			require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
		})
	}
}

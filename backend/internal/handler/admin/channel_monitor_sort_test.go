//go:build unit

package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type channelMonitorSortHandlerRepoStub struct {
	service.ChannelMonitorRepository
	updates []service.ChannelMonitorSortOrderUpdate
}

func (r *channelMonitorSortHandlerRepoStub) UpdateSortOrders(_ context.Context, updates []service.ChannelMonitorSortOrderUpdate) error {
	r.updates = append([]service.ChannelMonitorSortOrderUpdate(nil), updates...)
	return nil
}

func TestChannelMonitorUpdateSortOrderHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &channelMonitorSortHandlerRepoStub{}
	handler := NewChannelMonitorHandler(service.NewChannelMonitorService(repo, nil))
	router := gin.New()
	router.PUT("/api/v1/admin/channel-monitors/sort-order", handler.UpdateSortOrder)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/channel-monitors/sort-order", strings.NewReader(`{
		"updates":[{"id":9,"sort_order":20},{"id":3,"sort_order":10}]
	}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"code":0,"message":"success","data":{"message":"Sort order updated successfully"}}`, recorder.Body.String())
	require.Equal(t, []service.ChannelMonitorSortOrderUpdate{
		{ID: 9, SortOrder: 20},
		{ID: 3, SortOrder: 10},
	}, repo.updates)
}

func TestChannelMonitorUpdateSortOrderHandlerRejectsInvalidBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &channelMonitorSortHandlerRepoStub{}
	handler := NewChannelMonitorHandler(service.NewChannelMonitorService(repo, nil))
	router := gin.New()
	router.PUT("/api/v1/admin/channel-monitors/sort-order", handler.UpdateSortOrder)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/channel-monitors/sort-order", strings.NewReader(`{"updates":[{"id":0,"sort_order":10}]}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Empty(t, repo.updates)
}

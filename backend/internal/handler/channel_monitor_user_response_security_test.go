//go:build unit

package handler

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorUserResponsesHaveNoCredentialField(t *testing.T) {
	listPayload, err := json.Marshal(userMonitorViewToItem(&service.UserMonitorView{
		ID:       1,
		Name:     "monitor",
		Provider: service.MonitorProviderOpenAI,
	}))
	require.NoError(t, err)

	detailPayload, err := json.Marshal(userMonitorDetailToResponse(&service.UserMonitorDetail{
		ID:       1,
		Name:     "monitor",
		Provider: service.MonitorProviderOpenAI,
	}))
	require.NoError(t, err)

	require.NotContains(t, string(listPayload), `"api_key"`)
	require.NotContains(t, string(detailPayload), `"api_key"`)
}

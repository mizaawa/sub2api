//go:build unit

package admin

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorResponseNeverSerializesCredential(t *testing.T) {
	const secret = "canary-channel-monitor-secret"
	monitor := &service.ChannelMonitor{
		ID:       1,
		Name:     "monitor",
		Provider: service.MonitorProviderOpenAI,
		APIKey:   secret,
	}

	payload, err := json.Marshal(channelMonitorToResponse(monitor))

	require.NoError(t, err)
	require.NotContains(t, string(payload), secret)
	require.NotContains(t, string(payload), `"api_key"`)

	directPayload, err := json.Marshal(monitor)
	require.NoError(t, err)
	require.NotContains(t, string(directPayload), secret)
}

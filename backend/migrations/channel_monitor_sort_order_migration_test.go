package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelMonitorSortOrderMigration(t *testing.T) {
	content, err := FS.ReadFile("200_add_channel_monitor_sort_order.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0")
	require.Contains(t, sql, "UPDATE channel_monitors SET sort_order = id WHERE sort_order = 0")
	require.Contains(t, sql, "CREATE INDEX IF NOT EXISTS idx_channel_monitors_sort_order ON channel_monitors(sort_order)")
}

package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelMonitorGroupMigration(t *testing.T) {
	content, err := FS.ReadFile("201_channel_monitor_group_id.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS group_id BIGINT")
	require.Contains(t, sql, "ADD CONSTRAINT channel_monitors_group_id_fkey FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL")
	require.Contains(t, sql, "CREATE INDEX IF NOT EXISTS idx_channel_monitors_group_id ON channel_monitors(group_id)")
}

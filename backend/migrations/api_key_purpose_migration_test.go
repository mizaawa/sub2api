package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIKeyPurposeMigration(t *testing.T) {
	content, err := FS.ReadFile("202_api_key_purpose.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS purpose VARCHAR(32) NOT NULL DEFAULT ''")
	require.Contains(t, sql, "ADD CONSTRAINT api_keys_purpose_check CHECK (purpose IN ('', 'channel_monitor'))")
}

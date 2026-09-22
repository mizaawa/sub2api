package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomPlatformConstraintsMigration(t *testing.T) {
	content, err := FS.ReadFile("203_custom_platform_constraints.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.NotContains(t, sql, "UPDATE groups")
	require.NotContains(t, sql, "DELETE FROM account_groups")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check")
	require.Contains(t, sql, "CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom')) NOT VALID")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS composite_model_routes_target_platform_check")
	require.Contains(t, sql, "CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom')) NOT VALID")
	require.Equal(t, 2, strings.Count(sql, ") NOT VALID"))
}

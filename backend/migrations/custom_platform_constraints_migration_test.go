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
	require.Contains(t, sql, "UPDATE groups SET allow_image_generation = true WHERE platform = 'composite' AND allow_image_generation = false")
	require.Contains(t, sql, "DELETE FROM account_groups AS ag USING groups AS g, accounts AS a")
	require.Contains(t, sql, "g.platform = 'composite' AND a.platform <> 'custom'")
	require.Contains(t, sql, "g.platform <> 'composite' AND a.platform = 'custom'")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check")
	require.Contains(t, sql, "CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'))")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS composite_model_routes_target_platform_check")
	require.Contains(t, sql, "CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'))")
}

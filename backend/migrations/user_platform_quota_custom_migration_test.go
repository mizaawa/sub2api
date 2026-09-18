package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserPlatformQuotaCustomMigration(t *testing.T) {
	content, err := FS.ReadFile("209_user_platform_quotas_add_custom.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "position('custom' IN platform_constraint_def) = 0")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check")
	require.Contains(t, sql, "CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'))")
	require.NotContains(t, sql, "'composite'")
}

func TestCompositeRouteMigrationKeepsCustomExcluded(t *testing.T) {
	content, err := FS.ReadFile("172_composite_model_routes.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "composite_model_routes_target_platform_check")
	require.Contains(t, sql, "CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok'))")
	require.NotContains(t, sql, "target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom')")
}

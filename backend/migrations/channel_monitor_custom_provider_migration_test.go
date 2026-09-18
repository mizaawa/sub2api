package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelMonitorCustomProviderMigration(t *testing.T) {
	content, err := FS.ReadFile("207_channel_monitor_custom_provider.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "channel_monitors_provider_check")
	require.Contains(t, sql, "channel_monitor_request_templates_provider_check")
	require.Contains(t, sql, "CHECK (provider IN ('openai', 'anthropic', 'gemini', 'grok', 'custom'))")
	require.Contains(t, sql, "position('custom' IN monitor_constraint_def) = 0")
	require.Contains(t, sql, "position('custom' IN template_constraint_def) = 0")
	require.Contains(t, sql, "jsonb_array_elements(platforms)")
	require.Contains(t, sql, "platform_config.value->>'platform' = 'custom'")
}

func TestChannelMonitorV2FactoryConfigIncludesCustom(t *testing.T) {
	for _, name := range []string{
		"194_channel_monitor_v2.sql",
		"197_channel_monitor_v2_seed_popular_models.sql",
	} {
		content, err := FS.ReadFile(name)
		require.NoError(t, err)
		compact := strings.NewReplacer(" ", "", "\n", "", "\r", "", "\t", "").Replace(string(content))
		require.Contains(t, compact, `"platform":"custom"`, name)
	}
}

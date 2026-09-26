package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroupModelPricingMigration(t *testing.T) {
	content, err := FS.ReadFile("204_group_model_pricing.sql")
	require.NoError(t, err)
	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS model_pricing JSONB")
	require.Contains(t, sql, "CREATE OR REPLACE FUNCTION enqueue_group_auth_cache_invalidation()")
	require.Contains(t, sql, "OLD.model_pricing IS NOT DISTINCT FROM NEW.model_pricing")
	require.Contains(t, sql, "OLD.profit_control_enabled IS NOT DISTINCT FROM NEW.profit_control_enabled")
	require.Contains(t, sql, "INSERT INTO auth_cache_invalidation_outbox (cache_key)")
}

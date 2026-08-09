package repository

import (
	"context"
	"time"
)

// GetUserSpendingRank returns the 1-based rank for a user using the same
// deterministic ordering as the public leaderboard. Zero means no usage.
func (r *usageLogRepository) GetUserSpendingRank(ctx context.Context, startTime, endTime time.Time, userID int64) (int64, error) {
	const query = `
		WITH user_spend AS (
			SELECT u.user_id,
				COALESCE(SUM(u.actual_cost), 0) AS actual_cost,
				COALESCE(SUM(u.input_tokens + u.output_tokens + u.cache_creation_tokens + u.cache_read_tokens), 0) AS tokens
			FROM usage_logs u
			JOIN users us ON us.id = u.user_id AND us.deleted_at IS NULL
			WHERE u.created_at >= $1 AND u.created_at < $2
			GROUP BY u.user_id
		), ranked AS (
			SELECT user_id, ROW_NUMBER() OVER (ORDER BY actual_cost DESC, tokens DESC, user_id ASC) AS rank
			FROM user_spend
		)
		SELECT COALESCE((SELECT rank FROM ranked WHERE user_id = $3), 0)`
	rows, err := r.sql.QueryContext(ctx, query, startTime, endTime, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close() //nolint:errcheck
	if !rows.Next() {
		return 0, rows.Err()
	}
	var rank int64
	if err := rows.Scan(&rank); err != nil {
		return 0, err
	}
	return rank, rows.Err()
}

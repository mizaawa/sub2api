-- Users must explicitly opt in before their usage appears on the leaderboard.
CREATE TABLE IF NOT EXISTS user_leaderboard_preferences (
    user_id       BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    participating BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_leaderboard_preferences_participating
    ON user_leaderboard_preferences (participating);

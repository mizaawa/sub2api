-- Per-user deny list for public groups.
--
-- user_allowed_groups is an allow list for exclusive groups.  Keeping the
-- deny list separate means existing users retain the historical "all public
-- groups" behavior and an administrator can revoke one public group without
-- changing its visibility for every other user.
CREATE TABLE IF NOT EXISTS user_blocked_groups (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id   BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_user_blocked_groups_group_id
    ON user_blocked_groups (group_id);

-- A direct SQL change must evict every API-key auth snapshot for the user;
-- the application also invalidates these keys on the normal admin update path.
CREATE OR REPLACE FUNCTION enqueue_blocked_group_auth_cache_invalidation()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    target_user_id BIGINT;
BEGIN
    IF TG_OP = 'DELETE' THEN
        target_user_id := OLD.user_id;
    ELSE
        target_user_id := NEW.user_id;
    END IF;

    INSERT INTO auth_cache_invalidation_outbox (cache_key)
    SELECT encode(sha256(convert_to(k.key, 'UTF8')), 'hex')
    FROM api_keys AS k
    WHERE k.user_id = target_user_id
      AND k.deleted_at IS NULL
      AND k.key <> '';

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_user_blocked_groups_auth_cache_invalidation ON user_blocked_groups;
CREATE TRIGGER trg_user_blocked_groups_auth_cache_invalidation
AFTER INSERT OR UPDATE OR DELETE ON user_blocked_groups
FOR EACH ROW EXECUTE FUNCTION enqueue_blocked_group_auth_cache_invalidation();

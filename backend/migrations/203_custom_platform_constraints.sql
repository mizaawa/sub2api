-- Add the API-key-only Custom platform to persisted platform constraints.
-- The legacy group value remains "composite"; Custom is a concrete account
-- and route target used inside those groups.
--
-- Do not rewrite or delete production data during startup. NOT VALID still
-- enforces these constraints for new and updated rows, while avoiding a scan
-- of historical rows in the startup migration path.

ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

ALTER TABLE user_platform_quotas
    ADD CONSTRAINT user_platform_quotas_platform_check
    CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'))
    NOT VALID;

ALTER TABLE composite_model_routes
    DROP CONSTRAINT IF EXISTS composite_model_routes_target_platform_check;

ALTER TABLE composite_model_routes
    ADD CONSTRAINT composite_model_routes_target_platform_check
    CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'))
    NOT VALID;

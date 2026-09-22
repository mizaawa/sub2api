-- Add the API-key-only Custom platform to persisted platform constraints.
-- The legacy group value remains "composite"; Custom is a concrete account
-- and route target used inside those groups.

-- Video generation currently shares the media-generation capability gate.
-- Existing Composite groups predate that gate's Custom default, so preserve
-- functionality when they are presented as Custom after this migration.
UPDATE groups
SET allow_image_generation = true
WHERE platform = 'composite'
  AND allow_image_generation = false;

-- Composite groups previously accepted accounts from every provider. They
-- are now exposed as Custom groups and must contain only Custom API-key
-- accounts. Remove legacy cross-platform bindings in both directions before
-- the application starts enforcing the invariant on new writes.
DELETE FROM account_groups AS ag
USING groups AS g, accounts AS a
WHERE ag.group_id = g.id
  AND ag.account_id = a.id
  AND (
      (g.platform = 'composite' AND a.platform <> 'custom')
      OR (g.platform <> 'composite' AND a.platform = 'custom')
  );

ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

ALTER TABLE user_platform_quotas
    ADD CONSTRAINT user_platform_quotas_platform_check
    CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'));

ALTER TABLE composite_model_routes
    DROP CONSTRAINT IF EXISTS composite_model_routes_target_platform_check;

ALTER TABLE composite_model_routes
    ADD CONSTRAINT composite_model_routes_target_platform_check
    CHECK (target_platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'));

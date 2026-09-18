-- Allow Custom groups to use the same per-user platform quota machinery as
-- the other concrete platforms. Composite remains a routing-only platform and
-- is intentionally not part of this constraint.

DO $$
DECLARE
    platform_constraint_def TEXT;
BEGIN
    SELECT pg_get_constraintdef(c.oid)
      INTO platform_constraint_def
      FROM pg_constraint c
      JOIN pg_class t ON t.oid = c.conrelid
     WHERE t.relname = 'user_platform_quotas'
       AND c.conname = 'user_platform_quotas_platform_check';

    IF platform_constraint_def IS NULL OR position('custom' IN platform_constraint_def) = 0 THEN
        ALTER TABLE user_platform_quotas
            DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;
        ALTER TABLE user_platform_quotas
            ADD CONSTRAINT user_platform_quotas_platform_check
            CHECK (platform IN ('anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'));
    END IF;
END $$;

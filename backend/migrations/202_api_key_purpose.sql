-- Persist internal API-key ownership independently from mutable display names.
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS purpose VARCHAR(32) NOT NULL DEFAULT '';

COMMENT ON COLUMN api_keys.purpose IS
    'Internal key purpose; empty for ordinary user-managed keys';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'api_keys_purpose_check'
          AND conrelid = 'api_keys'::regclass
    ) THEN
        ALTER TABLE api_keys
            ADD CONSTRAINT api_keys_purpose_check
            CHECK (purpose IN ('', 'channel_monitor'));
    END IF;
END $$;

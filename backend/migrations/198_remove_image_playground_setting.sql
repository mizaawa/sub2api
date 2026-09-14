-- Remove the setting row left by the retired standalone image playground.
-- The feature and its API contract are no longer present, so this key is
-- otherwise an orphaned value that cannot be read or updated.
DELETE FROM settings
WHERE key = 'image_playground_enabled';

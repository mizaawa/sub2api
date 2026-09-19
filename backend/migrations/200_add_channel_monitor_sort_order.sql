-- Add persistent display ordering for channel monitors.
ALTER TABLE channel_monitors
    ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

-- Initialize existing rows with a deterministic ID-based order.
UPDATE channel_monitors
SET sort_order = id
WHERE sort_order = 0;

CREATE INDEX IF NOT EXISTS idx_channel_monitors_sort_order
    ON channel_monitors(sort_order);

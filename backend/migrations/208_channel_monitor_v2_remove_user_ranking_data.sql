-- Channel Monitor V2 is shipped without the per-user ranking feature.
-- Keep the historical tables for downgrade compatibility, but remove data that
-- is no longer read and could reveal per-user activity.

TRUNCATE TABLE
    channel_monitor_v2_user_metrics_1m,
    channel_monitor_v2_user_metrics_rollup;

DELETE FROM channel_monitor_v2_latency_histograms_1m
WHERE user_id <> 0;

DELETE FROM channel_monitor_v2_latency_histograms_rollup
WHERE user_id <> 0;

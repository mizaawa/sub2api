package service

import "sync/atomic"

// disableTempUnschedulableRuntime is the process-wide snapshot of the admin
// switch that controls transient account scheduling blockers.  Account values
// are also served from Redis and from repository queries, so keeping this
// small immutable bit in one place lets all those paths apply the same policy
// without adding a setting-service dependency to every scheduler call.
var disableTempUnschedulableRuntime atomic.Bool

// SetDisableTempUnschedulableRuntime publishes the current value of the admin
// switch.  It is intentionally cheap and lock-free because settings updates
// may happen while requests are being scheduled.
func SetDisableTempUnschedulableRuntime(enabled bool) {
	disableTempUnschedulableRuntime.Store(enabled)
}

// IsDisableTempUnschedulableEnabled reports whether transient scheduling
// blockers should be ignored.  Manual schedulable=false, inactive accounts,
// expiry auto-pause, and hard quota checks remain enforced by their callers.
func IsDisableTempUnschedulableEnabled() bool {
	return disableTempUnschedulableRuntime.Load()
}

// ShouldApplyTransientUnschedulableBlock is used by lower-level repository
// code to avoid importing or duplicating the setting key semantics.
func ShouldApplyTransientUnschedulableBlock() bool {
	return !IsDisableTempUnschedulableEnabled()
}

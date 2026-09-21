import type { MonitorStatus } from '@/api/admin/channelMonitor'
import {
  STATUS_DEGRADED,
  STATUS_ERROR,
  STATUS_FAILED,
  STATUS_OPERATIONAL,
} from '@/constants/channelMonitor'
import { firstTokenSeverity } from '@/utils/latencyHealth'

/** Apply the current latency thresholds while preserving explicit check failures. */
export function resolveMonitorStatus(
  status: MonitorStatus | '',
  latencyMs: number | null | undefined,
): MonitorStatus | '' {
  if (!status || status === STATUS_FAILED || status === STATUS_ERROR) return status
  if (latencyMs == null || !Number.isFinite(latencyMs)) return status

  switch (firstTokenSeverity(latencyMs)) {
    case 'critical':
      return STATUS_FAILED
    case 'warn':
    case 'slow':
      return STATUS_DEGRADED
    default:
      return STATUS_OPERATIONAL
  }
}

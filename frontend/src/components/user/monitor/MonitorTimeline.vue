<template>
  <div
    class="monitor-status-bar"
    role="img"
    :aria-label="timelineAriaLabel"
  >
    <span
      v-for="(bar, index) in displayBars"
      :key="index"
      class="monitor-status-bar__segment"
      :class="`monitor-status-bar__segment--${bar.status}`"
      :title="bar.title"
      aria-hidden="true"
    ></span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorTimelinePoint } from '@/api/channelMonitor'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'
import { MONITOR_TIMELINE_POINTS } from '@/constants/channelMonitor'
import { resolveMonitorStatus } from '@/utils/channelMonitorHealth'

type TimelineStatus = 'operational' | 'degraded' | 'failed' | 'error' | 'empty'
type ReportedTimelineStatus = Exclude<TimelineStatus, 'empty'>

interface TimelineBar {
  status: TimelineStatus
  title: string
}

const props = withDefaults(defineProps<{
  buckets?: MonitorTimelinePoint[]
  length?: number
  countdownSeconds?: number
  maintenance?: boolean
}>(), {
  buckets: () => [],
  length: MONITOR_TIMELINE_POINTS,
  countdownSeconds: 0,
  maintenance: false,
})

const { t } = useI18n()
const { statusLabel, formatLatency, formatRelativeTime } = useChannelMonitorFormat()

const safeLength = computed(() => Math.max(1, Math.floor(props.length)))
const realPointCount = computed(() => Math.min(props.buckets.length, safeLength.value))

const timelineAriaLabel = computed(() => {
  const label = t('channelStatus.timelineLabel', { n: realPointCount.value })
  const counts = new Map<ReportedTimelineStatus, number>()
  for (const point of props.buckets.slice(0, safeLength.value)) {
    const status = normalizeStatus(point)
    if (status !== 'empty') counts.set(status, (counts.get(status) ?? 0) + 1)
  }
  if (counts.size === 0) return label

  const summary = [...counts.entries()]
    .map(([status, count]) => `${statusLabel(status)}: ${count}`)
    .join('; ')
  return `${label}. ${summary}`
})

const displayBars = computed<TimelineBar[]>(() => {
  if (props.maintenance) {
    return Array.from({ length: safeLength.value }, () => ({
      status: 'empty' as const,
      title: t('monitorCommon.maintenancePaused'),
    }))
  }

  const points = [...props.buckets]
    .slice(0, safeLength.value)
    .reverse()
  const bars: TimelineBar[] = Array.from(
    { length: safeLength.value - points.length },
    () => ({ status: 'empty', title: '' }),
  )

  for (const point of points) {
    const status = normalizeStatus(point)
    const latency = point.latency_ms == null ? '' : ` · ${formatLatency(point.latency_ms)}ms`
    bars.push({
      status,
      title: `${formatRelativeTime(point.checked_at)} · ${statusLabel(status === 'empty' ? '' : status)}${latency}`,
    })
  }

  return bars
})

function normalizeStatus(point: MonitorTimelinePoint): TimelineStatus {
  const status = resolveMonitorStatus(point.status, point.latency_ms)
  if (
    status === 'operational' ||
    status === 'degraded' ||
    status === 'failed' ||
    status === 'error'
  ) {
    return status
  }
  return 'empty'
}
</script>

<style scoped>
.monitor-status-bar {
  display: grid;
  grid-template-columns: repeat(v-bind(safeLength), minmax(1px, 1fr));
  gap: 1px;
  width: 100%;
  height: 20px;
}

.monitor-status-bar__segment {
  width: 100%;
  min-width: 1px;
  height: 20px;
  border-radius: 1px;
  background: #e5e7eb;
  transition: filter 150ms ease;
}

.monitor-status-bar__segment:hover {
  filter: brightness(0.92);
}

.monitor-status-bar__segment--operational {
  background: #20bf8f;
}

.monitor-status-bar__segment--degraded {
  background: #f5b82e;
}

.monitor-status-bar__segment--failed,
.monitor-status-bar__segment--error {
  background: #ef6b5b;
}

.monitor-status-bar__segment--empty {
  background: #e5e7eb;
}

@media (max-width: 639px) {
  .monitor-status-bar {
    gap: 1px;
    height: 18px;
  }

  .monitor-status-bar__segment {
    height: 18px;
  }
}
</style>

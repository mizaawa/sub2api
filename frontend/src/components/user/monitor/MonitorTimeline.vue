<template>
  <div class="mt-4 pt-3 border-t border-gray-100 dark:border-dark-700/60">
    <div
      class="flex justify-between text-[10px] font-semibold uppercase tracking-widest text-gray-400 mb-2.5"
    >
      <span>{{ t('monitorCommon.history60pts', { n: length }) }}</span>
      <span class="tabular-nums">{{ t('monitorCommon.nextUpdateIn', { n: countdownSeconds }) }}</span>
    </div>

    <div
      v-if="maintenance"
      class="flex h-9 w-full items-center justify-center rounded-lg border border-dashed border-gray-300 dark:border-dark-600 text-[10px] uppercase tracking-widest text-gray-400"
    >
      {{ t('monitorCommon.maintenancePaused') }}
    </div>
    <div v-else class="timeline-container">
      <div
        v-for="(bar, idx) in displayBars"
        :key="idx"
        class="timeline-bar"
        :class="bar.colorClass"
        :style="{ height: bar.heightPct + '%' }"
        :title="bar.title"
      ></div>
    </div>

    <div
      class="mt-1.5 flex justify-between text-[9px] uppercase tracking-widest text-gray-400"
    >
      <span>{{ t('monitorCommon.past') }}</span>
      <span>{{ t('monitorCommon.now') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { MonitorTimelinePoint } from '@/api/channelMonitor'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'

const props = withDefaults(defineProps<{
  buckets?: MonitorTimelinePoint[]
  countdownSeconds: number
  length?: number
  maintenance?: boolean
}>(), {
  buckets: () => [],
  length: 30,
  maintenance: false,
})

const { t } = useI18n()
const { statusLabel, formatLatency, formatRelativeTime } = useChannelMonitorFormat()

interface Bar {
  colorClass: string
  heightPct: number
  title: string
}

// 初音未来主题配色：青色/绿色为正常，品红/粉色为警告，深色为错误
const STATUS_HEIGHT: Record<string, number> = {
  operational: 100,
  degraded: 62,
  failed: 38,
  error: 38,
  empty: 18,
}

const STATUS_COLOR: Record<string, string> = {
  operational: 'timeline-bar--operational',
  degraded: 'timeline-bar--degraded',
  failed: 'timeline-bar--failed',
  error: 'timeline-bar--error',
  empty: 'timeline-bar--empty',
}

const displayBars = computed<Bar[]>(() => {
  // Real points come newest-first; convert to oldest-first so the rightmost
  // bar represents "now". Pad the left with empty placeholders to keep the
  // bar count stable at `length`.
  const real = [...(props.buckets ?? [])]
    .slice(0, props.length)
    .reverse()

  const padCount = Math.max(0, props.length - real.length)
  const bars: Bar[] = []

  for (let i = 0; i < padCount; i += 1) {
    bars.push({
      colorClass: STATUS_COLOR.empty,
      heightPct: STATUS_HEIGHT.empty,
      title: '',
    })
  }

  for (const point of real) {
    const status = point.status as keyof typeof STATUS_HEIGHT
    const colorClass = STATUS_COLOR[status] ?? STATUS_COLOR.empty
    const heightPct = STATUS_HEIGHT[status] ?? STATUS_HEIGHT.empty
    const latency = formatLatency(point.latency_ms)
    const relative = formatRelativeTime(point.checked_at)
    const label = statusLabel(point.status)
    bars.push({
      colorClass,
      heightPct,
      title: `${relative} · ${label} · ${latency}ms`,
    })
  }

  return bars
})
</script>

<style scoped>
.timeline-container {
  display: grid;
  grid-template-columns: repeat(v-bind(length), 1fr);
  gap: 3px;
  height: 36px;
  width: 100%;
  align-items: end;
}

.timeline-bar {
  border-radius: 4px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
}

.timeline-bar:hover {
  transform: translateY(-2px);
  filter: brightness(1.15);
}

/* 初音未来配色方案 */
.timeline-bar--operational {
  background: linear-gradient(180deg, #39c5bb 0%, #2bb3a9 100%);
  box-shadow: 0 1px 3px rgba(57, 197, 187, 0.3);
}

.timeline-bar--degraded {
  background: linear-gradient(180deg, #e12885 0%, #c91f73 100%);
  box-shadow: 0 1px 3px rgba(225, 40, 133, 0.3);
}

.timeline-bar--failed {
  background: linear-gradient(180deg, #dc2626 0%, #b91c1c 100%);
  box-shadow: 0 1px 3px rgba(220, 38, 38, 0.3);
}

.timeline-bar--error {
  background: linear-gradient(180deg, #dc2626 0%, #b91c1c 100%);
  box-shadow: 0 1px 3px rgba(220, 38, 38, 0.3);
}

.timeline-bar--empty {
  background: #e5e7eb;
  box-shadow: none;
}

:global(.dark) .timeline-bar--operational {
  background: linear-gradient(180deg, #39c5bb 0%, #2bb3a9 100%);
  box-shadow: 0 1px 4px rgba(57, 197, 187, 0.4);
}

:global(.dark) .timeline-bar--degraded {
  background: linear-gradient(180deg, #e12885 0%, #c91f73 100%);
  box-shadow: 0 1px 4px rgba(225, 40, 133, 0.4);
}

:global(.dark) .timeline-bar--failed {
  background: linear-gradient(180deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 1px 4px rgba(239, 68, 68, 0.4);
}

:global(.dark) .timeline-bar--error {
  background: linear-gradient(180deg, #ef4444 0%, #dc2626 100%);
  box-shadow: 0 1px 4px rgba(239, 68, 68, 0.4);
}

:global(.dark) .timeline-bar--empty {
  background: #374151;
  box-shadow: none;
}
</style>

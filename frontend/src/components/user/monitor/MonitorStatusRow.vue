<template>
  <article
    class="px-4 py-5 sm:px-6"
    :data-testid="`monitor-status-row-${item.id}`"
  >
    <div class="flex min-w-0 flex-wrap items-start gap-x-4 gap-y-1">
      <div class="flex min-w-0 max-w-full flex-shrink-0 items-start gap-2.5">
        <span
          role="img"
          class="grid h-5 w-5 flex-shrink-0 place-items-center rounded-full text-white"
          :class="statusIconClass"
          :aria-label="statusLabel(currentStatus)"
          :title="statusLabel(currentStatus)"
        >
          <Icon :name="statusIcon" size="xs" :stroke-width="2.5" />
        </span>
        <div class="min-w-0 flex flex-wrap items-center gap-x-2 gap-y-1">
          <h2 class="min-w-0 break-words text-sm font-medium leading-5 text-gray-950 sm:text-base">
            {{ item.name }}
          </h2>
          <span
            v-if="groupRateLabel"
            class="inline-flex flex-shrink-0 items-center rounded-md bg-gray-100 px-1.5 py-0.5 text-xs font-medium text-gray-500"
            data-testid="monitor-status-group-rate"
          >
            {{ groupRateLabel }}
          </span>
        </div>
      </div>

      <span class="ml-auto flex-shrink-0 whitespace-nowrap text-sm text-gray-500 sm:text-base">
        {{ uptimeLabel }}
      </span>
    </div>

    <MonitorTimeline
      class="mt-3"
      :buckets="item.timeline"
      :length="MONITOR_TIMELINE_POINTS"
    />
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserMonitorView } from '@/api/channelMonitor'
import Icon from '@/components/icons/Icon.vue'
import { useChannelMonitorFormat } from '@/composables/useChannelMonitorFormat'
import { MONITOR_TIMELINE_POINTS } from '@/constants/channelMonitor'
import { formatMultiplier } from '@/utils/formatters'
import { resolveMonitorStatus } from '@/utils/channelMonitorHealth'
import MonitorTimeline from './MonitorTimeline.vue'

const props = defineProps<{
  item: UserMonitorView
}>()

const { t } = useI18n()
const { statusLabel } = useChannelMonitorFormat()

const groupRateLabel = computed(() => {
  const rate = props.item.group_rate_multiplier
  if (rate == null || !Number.isFinite(rate)) return ''
  return `${formatMultiplier(rate)}x`
})

const currentStatus = computed(() =>
  resolveMonitorStatus(props.item.primary_status, props.item.primary_latency_ms)
)

const statusIcon = computed<'check' | 'exclamationTriangle' | 'x' | 'infoCircle'>(() => {
  switch (currentStatus.value) {
    case 'operational':
      return 'check'
    case 'degraded':
      return 'exclamationTriangle'
    case 'failed':
    case 'error':
      return 'x'
    default:
      return 'infoCircle'
  }
})

const statusIconClass = computed(() => {
  switch (currentStatus.value) {
    case 'operational':
      return 'bg-emerald-500'
    case 'degraded':
      return 'bg-amber-500'
    case 'failed':
    case 'error':
      return 'bg-red-500'
    default:
      return 'bg-gray-400'
  }
})

const hasHistory = computed(() =>
  Boolean(props.item.primary_status) || (props.item.timeline?.length ?? 0) > 0
)

const uptimeLabel = computed(() => {
  if (!hasHistory.value || !Number.isFinite(props.item.availability_7d)) {
    return t('channelStatus.uptimeUnavailable')
  }

  const value = Math.max(0, Math.min(100, props.item.availability_7d))
  const formatted = new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(value)
  return t('channelStatus.uptime', { value: formatted })
})
</script>

<template>
  <AppLayout>
    <div class="mx-auto w-full max-w-5xl pb-8">
      <section
        class="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm"
        aria-labelledby="channel-system-status-title"
      >
        <header class="border-b border-gray-200 px-4 py-5 sm:px-6">
          <h1
            id="channel-system-status-title"
            class="text-xl font-semibold text-gray-950"
          >
            {{ t('channelStatus.systemStatus') }}
          </h1>
        </header>

        <MonitorCardGrid :items="items" :loading="loading" />
      </section>

      <p class="mt-4 px-1 text-xs leading-5 text-gray-500">
        {{ t('channelStatus.metricsDisclaimer') }}
      </p>

      <footer class="mt-5 border-t border-gray-200 pt-4 text-center text-xs text-gray-500">
        <span>{{ t('channelStatus.poweredBy') }}</span>
        <a
          href="https://mizaawa.com"
          target="_blank"
          rel="noopener noreferrer"
          class="ml-1 font-medium text-gray-700 underline decoration-gray-300 underline-offset-4 transition-colors hover:text-gray-950"
        >
          mizaawa.com
        </a>
      </footer>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  list as listChannelMonitorViews,
  type UserMonitorView,
} from '@/api/channelMonitor'
import AppLayout from '@/components/layout/AppLayout.vue'
import MonitorCardGrid from '@/components/user/monitor/MonitorCardGrid.vue'
import { DEFAULT_INTERVAL_SECONDS } from '@/constants/channelMonitor'
import { useAutoRefresh } from '@/composables/useAutoRefresh'

const { t } = useI18n()
const appStore = useAppStore()

const items = ref<UserMonitorView[]>([])
const loading = ref(false)
let abortController: AbortController | null = null

const autoRefresh = useAutoRefresh({
  storageKey: 'channel-status-auto-refresh',
  intervals: [30, 60, 120] as const,
  defaultInterval: DEFAULT_INTERVAL_SECONDS,
  onRefresh: () => reload(true),
  shouldPause: () => document.hidden || loading.value,
})

async function reload(silent = false) {
  abortController?.abort()
  const ctrl = new AbortController()
  abortController = ctrl
  if (!silent) loading.value = true

  try {
    const response = await listChannelMonitorViews({ signal: ctrl.signal })
    if (ctrl.signal.aborted || abortController !== ctrl) return
    items.value = response.items || []
  } catch (err: unknown) {
    const requestError = err as { name?: string; code?: string }
    if (requestError?.name === 'AbortError' || requestError?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(err, t('channelStatus.loadError')))
  } finally {
    if (abortController === ctrl) {
      if (!silent) loading.value = false
      autoRefresh.resetCountdown()
      abortController = null
    }
  }
}

watch(
  () => appStore.cachedPublicSettings?.channel_monitor_enabled,
  (enabled) => {
    if (enabled === false) autoRefresh.stop()
    else if (autoRefresh.enabled.value) autoRefresh.start()
  },
)

onMounted(() => {
  void reload(false)
  if (appStore.cachedPublicSettings?.channel_monitor_enabled !== false) {
    autoRefresh.setEnabled(true)
  }
})

onBeforeUnmount(() => {
  abortController?.abort()
})
</script>

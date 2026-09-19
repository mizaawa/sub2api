<template>
  <div class="divide-y divide-gray-200">
    <section
      v-for="provider in providerSections"
      :key="provider.value"
      :data-testid="`monitor-provider-${provider.value}`"
    >
      <button
        type="button"
        class="flex w-full items-center justify-between gap-4 px-4 py-4 text-left transition-colors hover:bg-gray-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-primary-500 sm:px-6"
        :aria-expanded="expanded[provider.value]"
        :aria-controls="`monitor-provider-content-${provider.value}`"
        :data-testid="`monitor-provider-toggle-${provider.value}`"
        @click="toggleProvider(provider.value)"
      >
        <span class="flex min-w-0 items-center gap-2.5">
          <span class="truncate text-base font-semibold text-gray-950">
            {{ provider.label }}
          </span>
          <span class="whitespace-nowrap text-sm text-gray-500">
            {{ t('channelStatus.componentCount', { n: provider.items.length }) }}
          </span>
        </span>

        <Icon
          name="chevronDown"
          size="sm"
          class="flex-shrink-0 text-gray-400 transition-transform duration-200"
          :class="expanded[provider.value] ? 'rotate-180' : ''"
        />
      </button>

      <div
        v-show="expanded[provider.value]"
        :id="`monitor-provider-content-${provider.value}`"
        class="border-t border-gray-100"
      >
        <div
          v-if="loading"
          class="space-y-5 px-4 py-5 sm:px-6"
          :data-testid="`monitor-provider-loading-${provider.value}`"
          :aria-label="t('channelStatus.loadingProvider', { provider: provider.label })"
        >
          <div class="flex animate-pulse items-center gap-3">
            <div class="h-5 w-5 flex-shrink-0 rounded-full bg-gray-200"></div>
            <div class="h-4 w-36 rounded bg-gray-200"></div>
            <div class="ml-auto h-4 w-20 rounded bg-gray-200"></div>
          </div>
          <div class="h-5 animate-pulse rounded-sm bg-gray-100"></div>
        </div>

        <p
          v-else-if="provider.items.length === 0"
          class="px-4 py-5 text-sm text-gray-500 sm:px-6"
          :data-testid="`monitor-provider-empty-${provider.value}`"
        >
          {{ t('channelStatus.emptyProvider') }}
        </p>

        <div v-else class="divide-y divide-gray-100">
          <MonitorStatusRow
            v-for="item in provider.items"
            :key="item.id"
            :item="item"
          />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import type { UserMonitorView } from '@/api/channelMonitor'
import Icon from '@/components/icons/Icon.vue'
import MonitorStatusRow from './MonitorStatusRow.vue'

type ProviderKey = 'openai' | 'anthropic' | 'gemini' | 'grok' | 'custom'

const PROVIDER_KEYS: readonly ProviderKey[] = [
  'openai',
  'anthropic',
  'gemini',
  'grok',
  'custom',
]

const props = defineProps<{
  items: UserMonitorView[]
  loading: boolean
}>()

const { t } = useI18n()

const expanded = reactive<Record<ProviderKey, boolean>>({
  openai: true,
  anthropic: true,
  gemini: true,
  grok: true,
  custom: true,
})

const providerSections = computed(() => PROVIDER_KEYS.map((value) => ({
  value,
  label: t(`channelStatus.providers.${value}`),
  items: props.items.filter((item) => String(item.provider) === value),
})))

function toggleProvider(provider: ProviderKey) {
  expanded[provider] = !expanded[provider]
}
</script>

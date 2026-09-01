<template>
  <AppLayout>
    <section class="image-playground-page">
      <header class="playground-toolbar">
        <div class="min-w-0">
          <h1 class="text-lg font-semibold text-gray-900">{{ t('imagePlayground.title') }}</h1>
        </div>

        <div v-if="imageKeys.length" class="playground-controls">
          <label class="control-group">
            <span class="control-label">{{ t('imagePlayground.keyLabel') }}</span>
            <Select
              v-model="selectedKeyId"
              :options="keyOptions"
              :disabled="loadingKeys"
              class="w-full"
              searchable
              @change="handleKeyChange"
            />
          </label>

          <label class="control-group">
            <span class="control-label">{{ t('imagePlayground.modelLabel') }}</span>
            <Select
              v-model="selectedModel"
              :options="modelOptions"
              :disabled="loadingModels"
              class="w-full"
              searchable
            />
          </label>

          <div class="balance-summary" aria-live="polite">
            <span class="balance-label">{{ t('imagePlayground.availableBalance') }}</span>
            <strong class="balance-value">{{ formatMoney(authStore.user?.balance) }}</strong>
            <span v-if="frozenBalance > 0" class="frozen-balance">
              {{ t('imagePlayground.frozenBalance') }} {{ formatMoney(frozenBalance) }}
            </span>
          </div>

          <button
            type="button"
            class="refresh-button"
            :disabled="refreshing"
            :title="t('imagePlayground.refresh')"
            :aria-label="t('imagePlayground.refresh')"
            @click="refreshAll"
          >
            <Icon name="refresh" size="md" :class="refreshing ? 'animate-spin' : ''" />
          </button>
        </div>
      </header>

      <p v-if="modelsError && imageKeys.length" class="playground-notice" role="status">
        {{ t('imagePlayground.modelsFailed') }}
      </p>

      <div v-if="loadingKeys" class="playground-state">
        <LoadingSpinner />
        <p>{{ t('imagePlayground.loading') }}</p>
      </div>

      <div v-else-if="loadError" class="playground-state playground-state-error" role="alert">
        <Icon name="exclamationCircle" size="xl" />
        <p>{{ t('imagePlayground.loadFailed') }}</p>
        <button type="button" class="btn btn-secondary" @click="refreshAll">
          <Icon name="refresh" size="sm" class="mr-2" />
          {{ t('common.retry') }}
        </button>
      </div>

      <div v-else-if="!imageKeys.length" class="playground-state">
        <Icon name="key" size="xl" class="text-gray-400" />
        <div class="text-center">
          <h2 class="text-base font-semibold text-gray-900">{{ t('imagePlayground.noAccessTitle') }}</h2>
          <p class="mt-1 max-w-lg text-sm text-gray-500">{{ t('imagePlayground.noAccessDescription') }}</p>
        </div>
        <RouterLink :to="{ path: '/keys', query: { returnTo: '/image-playground' } }" class="btn btn-primary">
          {{ t('imagePlayground.manageKeys') }}
        </RouterLink>
      </div>

      <div v-else-if="selectedKey && selectedModel" class="playground-frame-shell">
        <div v-if="loadingModels" class="frame-loading" aria-live="polite">
          <LoadingSpinner />
          <span>{{ t('imagePlayground.loadingModels') }}</span>
        </div>
        <iframe
          ref="frameRef"
          :key="frameIdentity"
          :name="frameName"
          class="playground-frame"
          src="/image-playground-app/index.html"
          :title="t('imagePlayground.playgroundTitle')"
          allow="clipboard-read; clipboard-write"
          @load="clearFrameNameAttribute"
        ></iframe>
      </div>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select, { type SelectOption } from '@/components/common/Select.vue'
import { keysAPI } from '@/api/keys'
import { buildGatewayUrl } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { ApiKey } from '@/types'

interface GatewayModel {
  id?: string
  display_name?: string
}

const { t } = useI18n()
const authStore = useAuthStore()
const imageKeys = ref<ApiKey[]>([])
const selectedKeyId = ref<number | null>(null)
const models = ref<GatewayModel[]>([])
const selectedModel = ref<string | null>(null)
const loadingKeys = ref(true)
const loadingModels = ref(false)
const refreshing = ref(false)
const loadError = ref(false)
const modelsError = ref(false)
const frameRef = ref<HTMLIFrameElement | null>(null)
let balanceRefreshTimer: number | undefined

const selectedKey = computed(() =>
  imageKeys.value.find((key) => key.id === selectedKeyId.value) ?? null,
)

const frozenBalance = computed(() => Number(authStore.user?.frozen_balance || 0))

const keyOptions = computed<SelectOption[]>(() => imageKeys.value.map((key) => ({
  value: key.id,
  label: t('imagePlayground.keyOption', {
    group: key.group?.name || '-',
    key: key.name,
  }),
})))

const modelOptions = computed<SelectOption[]>(() => models.value.map((model) => ({
  value: model.id || '',
  label: model.display_name && model.display_name !== model.id
    ? `${model.display_name} (${model.id})`
    : model.id || '',
})).filter((option) => option.value))

const frameIdentity = computed(() => `${selectedKeyId.value || ''}:${selectedModel.value || ''}`)

const frameName = computed(() => {
  if (!selectedKey.value || !selectedModel.value) return ''
  return `sub2api-image-playground:${JSON.stringify({
    customProviders: [],
    profiles: [{
      id: 'sub2api-integrated',
      name: `${selectedKey.value.group?.name || 'Sub2API'} · ${selectedKey.value.name}`,
      provider: 'sb2api-async',
      baseUrl: buildGatewayUrl('/v1'),
      apiKey: selectedKey.value.key,
      model: selectedModel.value,
      timeout: 600,
      transparentBackgroundMethod: 'api',
    }],
  })}`
})

function keyAllowsImageGeneration(key: ApiKey): boolean {
  const platform = key.group?.platform
  return key.status === 'active'
    && key.group?.allow_image_generation === true
    && (platform === 'openai' || platform === 'grok' || platform === 'composite')
}

function defaultModelForKey(key: ApiKey): string {
  return key.group?.platform === 'grok' ? 'grok-imagine' : 'gpt-image-2'
}

function chooseDefaultModel(key: ApiKey, available: GatewayModel[]): string {
  const ids = available.map((model) => model.id || '').filter(Boolean)
  const preferred = key.group?.platform === 'grok'
    ? ['grok-imagine-image-quality', 'grok-imagine-image', 'grok-imagine']
    : ['gpt-image-2', 'gpt-image-1.5', 'gpt-image-1', 'dall-e-3']
  return preferred.find((model) => ids.includes(model)) || ids[0] || defaultModelForKey(key)
}

async function loadKeys(): Promise<void> {
  loadingKeys.value = true
  loadError.value = false
  try {
    const keys: ApiKey[] = []
    let page = 1
    while (true) {
      const response = await keysAPI.list(page, 100, {
        status: 'active',
        sort_by: 'created_at',
        sort_order: 'desc',
      })
      keys.push(...(response.items || []).filter(keyAllowsImageGeneration))
      if (page >= response.pages || !response.items?.length) break
      page += 1
    }
    imageKeys.value = keys
    if (!keys.some((key) => key.id === selectedKeyId.value)) {
      selectedKeyId.value = keys[0]?.id ?? null
    }
    if (selectedKey.value) await loadModels()
  } catch (error) {
    console.error('Failed to load image playground keys:', error)
    imageKeys.value = []
    selectedKeyId.value = null
    loadError.value = true
  } finally {
    loadingKeys.value = false
  }
}

async function loadModels(): Promise<void> {
  const key = selectedKey.value
  if (!key) {
    models.value = []
    selectedModel.value = null
    return
  }

  loadingModels.value = true
  modelsError.value = false
  try {
    const response = await fetch(buildGatewayUrl('/v1/models'), {
      headers: { Authorization: `Bearer ${key.key}` },
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const body = await response.json() as { data?: GatewayModel[] }
    const available = (body.data || []).filter((model) => {
      const id = String(model.id || '').toLowerCase()
      if (key.group?.platform === 'grok') {
        return id.startsWith('grok-imagine') && !id.includes('video')
      }
      if (key.group?.platform === 'openai') {
        return id.includes('image') || id.includes('dall-e')
      }
      return (id.includes('image') || id.includes('imagine') || id.includes('dall-e')) && !id.includes('video')
    })
    models.value = available.length ? available : [{ id: defaultModelForKey(key) }]
    selectedModel.value = chooseDefaultModel(key, models.value)
  } catch (error) {
    console.warn('Failed to load image playground models:', error)
    models.value = [{ id: defaultModelForKey(key) }]
    selectedModel.value = models.value[0].id || null
    modelsError.value = true
  } finally {
    loadingModels.value = false
  }
}

async function handleKeyChange(): Promise<void> {
  selectedModel.value = null
  await loadModels()
}

async function refreshAll(): Promise<void> {
  if (refreshing.value) return
  refreshing.value = true
  try {
    await Promise.all([loadKeys(), authStore.refreshUser()])
  } finally {
    refreshing.value = false
  }
}

function clearFrameNameAttribute(): void {
  frameRef.value?.removeAttribute('name')
}

function formatMoney(value: number | null | undefined): string {
  return new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 4,
  }).format(Number(value || 0))
}

onMounted(async () => {
  await Promise.all([loadKeys(), authStore.refreshUser().catch(() => undefined)])
  balanceRefreshTimer = window.setInterval(() => {
    void authStore.refreshUser().catch(() => undefined)
  }, 15_000)
})

onBeforeUnmount(() => {
  if (balanceRefreshTimer !== undefined) window.clearInterval(balanceRefreshTimer)
})
</script>

<style scoped>
.image-playground-page {
  display: flex;
  min-height: calc(100dvh - 8rem);
  flex-direction: column;
  gap: 0.75rem;
}

.playground-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: 1rem;
  background: transparent;
}

.playground-controls {
  display: grid;
  width: min(100%, 58rem);
  grid-template-columns: minmax(15rem, 1.5fr) minmax(12rem, 1fr) minmax(8.5rem, auto) 2.75rem;
  align-items: end;
  gap: 0.75rem;
}

.control-group {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.35rem;
}

.control-label,
.balance-label {
  color: #6b7280;
  font-size: 0.75rem;
  font-weight: 500;
}

.balance-summary {
  display: flex;
  min-height: 2.75rem;
  flex-direction: column;
  justify-content: center;
  padding: 0 0.25rem;
}

.balance-value {
  color: #047857;
  font-size: 0.95rem;
  line-height: 1.25rem;
}

.frozen-balance {
  color: #92400e;
  font-size: 0.7rem;
}

.refresh-button {
  display: inline-flex;
  width: 2.75rem;
  height: 2.75rem;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 0.375rem;
  background: #f3f4f6;
  color: #4b5563;
  transition: background-color 150ms ease, color 150ms ease;
}

.refresh-button:hover:not(:disabled) {
  background: #e5e7eb;
  color: #111827;
}

.refresh-button:focus-visible {
  outline: 2px solid rgb(20 184 166 / 35%);
  outline-offset: 2px;
}

.refresh-button:disabled {
  cursor: wait;
  opacity: 0.55;
}

.playground-notice {
  color: #92400e;
  font-size: 0.75rem;
  text-align: right;
}

.playground-state {
  display: flex;
  min-height: 28rem;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  background: transparent;
  color: #6b7280;
}

.playground-state-error {
  color: #b91c1c;
}

.playground-frame-shell {
  position: relative;
  min-height: 42rem;
  flex: 1;
  overflow: hidden;
  border: 0;
  border-radius: 0.375rem;
  background: #f8fafc;
}

.playground-frame {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 42rem;
  border: 0;
  background: transparent;
}

.frame-loading {
  position: absolute;
  inset: 0;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  background: rgb(248 250 252 / 88%);
  color: #6b7280;
  font-size: 0.875rem;
}

@media (max-width: 1023px) {
  .image-playground-page {
    min-height: calc(100dvh - 7rem);
  }

  .playground-controls {
    width: 100%;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto 2.75rem;
  }
}

@media (max-width: 767px) {
  .playground-controls {
    grid-template-columns: minmax(0, 1fr) 2.75rem;
  }

  .control-group,
  .balance-summary {
    grid-column: 1 / -1;
  }

  .balance-summary {
    min-height: auto;
  }

  .refresh-button {
    grid-column: 2;
    grid-row: 3;
  }

  .playground-frame-shell,
  .playground-frame {
    min-height: 36rem;
  }
}
</style>

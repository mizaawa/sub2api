<template>
  <main class="image-playground-launcher">
    <div v-if="loadingKeys" class="launcher-state" aria-live="polite">
      <LoadingSpinner />
      <p>{{ t('imagePlayground.loading') }}</p>
    </div>

    <div v-else-if="loadError" class="launcher-state launcher-state-error" role="alert">
      <Icon name="exclamationCircle" size="xl" />
      <p>{{ t('imagePlayground.loadFailed') }}</p>
      <button type="button" class="btn btn-secondary" @click="refreshAll">
        <Icon name="refresh" size="sm" class="mr-2" />
        {{ t('common.retry') }}
      </button>
    </div>

    <div v-else-if="!imageKeys.length" class="launcher-state">
      <Icon name="key" size="xl" class="text-gray-400" />
      <div class="text-center">
        <h1 class="text-base font-semibold text-gray-900">{{ t('imagePlayground.noAccessTitle') }}</h1>
        <p class="mt-1 max-w-lg text-sm text-gray-500">{{ t('imagePlayground.noAccessDescription') }}</p>
      </div>
      <RouterLink :to="{ path: '/keys', query: { returnTo: '/image-playground' } }" class="btn btn-primary">
        {{ t('imagePlayground.manageKeys') }}
      </RouterLink>
    </div>

    <div v-else class="launcher-state" aria-live="polite">
      <LoadingSpinner />
      <p>{{ loadingModels ? t('imagePlayground.loadingModels') : t('imagePlayground.opening') }}</p>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { keysAPI } from '@/api/keys'
import { buildGatewayUrl } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { ApiKey } from '@/types'

interface GatewayModel {
  id?: string
  display_name?: string
}

const DEFAULT_IMAGE_PLAYGROUND_URL = 'https://api.relaycat.top/image-playground/'
const imagePlaygroundUrl = String(import.meta.env.VITE_IMAGE_PLAYGROUND_URL || DEFAULT_IMAGE_PLAYGROUND_URL).trim()

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
const redirecting = ref(false)
let balanceRefreshTimer: number | undefined

const selectedKey = computed(() =>
  imageKeys.value.find((key) => key.id === selectedKeyId.value) ?? null,
)

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
  } finally {
    loadingModels.value = false
    if (selectedKey.value && selectedModel.value) redirectToStandalone()
  }
}

function buildStandaloneUrl(key: ApiKey, model: string): string {
  const target = new URL(imagePlaygroundUrl, window.location.origin)
  const settings = {
    customProviders: [],
    profiles: [{
      id: 'sub2api-integrated',
      name: `${key.group?.name || 'Sub2API'} · ${key.name}`,
      provider: 'sb2api-async',
      baseUrl: buildGatewayUrl('/v1'),
      apiKey: key.key,
      model,
      timeout: 600,
      apiMode: 'images',
      transparentBackgroundMethod: 'api',
    }],
  }
  target.searchParams.set('settings', JSON.stringify(settings))
  return target.toString()
}

function redirectToStandalone(): void {
  if (redirecting.value || !selectedKey.value || !selectedModel.value) return
  try {
    redirecting.value = true
    window.location.replace(buildStandaloneUrl(selectedKey.value, selectedModel.value))
  } catch (error) {
    console.error('Failed to open standalone image playground:', error)
    redirecting.value = false
    loadError.value = true
  }
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
.image-playground-launcher {
  display: grid;
  min-height: 100dvh;
  place-items: center;
  padding: 2rem;
  background: #f8fafc;
}

.launcher-state {
  display: flex;
  max-width: 32rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  color: #6b7280;
  text-align: center;
}

.launcher-state-error {
  color: #b91c1c;
}

@media (prefers-color-scheme: dark) {
  .image-playground-launcher {
    background: #111827;
  }
}
</style>

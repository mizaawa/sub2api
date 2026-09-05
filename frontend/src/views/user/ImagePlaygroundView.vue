<template>
  <main class="image-playground-launcher">
    <div v-if="!imagePlaygroundEnabled" class="launcher-state" role="status">
      <Icon name="lock" size="xl" class="text-gray-400" />
      <p>{{ t('imagePlayground.disabled') }}</p>
    </div>

    <div v-else-if="loadingKeys" class="launcher-state" aria-live="polite">
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
import { userGroupsAPI } from '@/api/groups'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { FeatureFlags, isFeatureFlagEnabled } from '@/utils/featureFlags'
import type { ApiKey } from '@/types'

interface GatewayModel {
  id?: string
  display_name?: string
}

interface LoadContext {
  generation: number
  userId: number
  userEmail: string
  tokenPresent: boolean
  controller: AbortController
}

const STANDALONE_IMAGE_PLAYGROUND_PATH = '/image-playground/'
const STANDALONE_CONFIG_PREFIX = 'sub2api-image-playground:'
const STANDALONE_CONFIG_STORAGE_KEY = 'sub2api-image-playground:bootstrap'
// Mobile browsers may create a fresh sessionStorage partition when the
// standalone route is opened from an installed/PWA context. Keep a short-lived
// same-origin fallback so the selected key profile survives that handoff.
const STANDALONE_CONFIG_FALLBACK_STORAGE_KEY = `${STANDALONE_CONFIG_STORAGE_KEY}:fallback`
// Browser requests stay same-origin. When Sub2API is deployed at the zayu
// hostname this still resolves to the zayu `/v1` endpoint, while local/dev
// deployments avoid a cross-origin preflight for the auth headers.
const IMAGE_PLAYGROUND_API_BASE_URL = '/v1'
const STANDALONE_SERVICE_WORKER_PATH = '/image-playground/'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const imageKeys = ref<ApiKey[]>([])
const selectedKeyId = ref<number | null>(null)
const models = ref<GatewayModel[]>([])
const modelsByKeyId = ref<Record<number, string[]>>({})
const selectedModel = ref<string | null>(null)
const loadingKeys = ref(true)
const loadingModels = ref(false)
const refreshing = ref(false)
const loadError = ref(false)
const redirecting = ref(false)
let balanceRefreshTimer: number | undefined
let loadGeneration = 0
let activeLoadController: AbortController | null = null

const selectedKey = computed(() =>
  imageKeys.value.find((key) => key.id === selectedKeyId.value) ?? null,
)
const imagePlaygroundEnabled = computed(() => isFeatureFlagEnabled(FeatureFlags.imagePlayground))

function normalizeNumericId(value: unknown): number | null {
  const id = typeof value === 'number' ? value : typeof value === 'string' ? Number(value.trim()) : NaN
  return Number.isSafeInteger(id) && id > 0 ? id : null
}

function readPersistedIdentity(): { userId: number | null; userEmail: string | null } {
  try {
    const raw = localStorage.getItem('auth_user')
    if (!raw) return { userId: null, userEmail: null }
    const parsed = JSON.parse(raw) as { id?: unknown; email?: unknown }
    const email = typeof parsed.email === 'string' ? parsed.email.trim().toLowerCase() : ''
    return {
      userId: normalizeNumericId(parsed.id),
      userEmail: email || null,
    }
  } catch {
    return { userId: null, userEmail: null }
  }
}

function currentAuthToken(): string | null {
  const inMemoryToken = typeof authStore.token === 'string' ? authStore.token.trim() : ''
  if (inMemoryToken) return inMemoryToken
  try {
    const persistedToken = localStorage.getItem('auth_token')?.trim() || ''
    return persistedToken || null
  } catch {
    return null
  }
}

function keyAllowsImageGeneration(key: ApiKey): boolean {
  return typeof key.key === 'string'
    && key.key.trim().length > 0
    && String(key.status).toLowerCase() === 'active'
    // The image gateway resolves its handler and upstream platform from the
    // bound group. An ungrouped key has no platform target and would be
    // accepted by the launcher only to fail with a 404 in `/v1/images/*`.
    && key.group_id != null
    && key.group?.allow_image_generation === true
}

function currentUserId(): number | null {
  return normalizeNumericId(authStore.user?.id) ?? readPersistedIdentity().userId
}

function currentUserEmail(): string | null {
  const email = authStore.user?.email
  if (typeof email === 'string' && email.trim()) return email.trim().toLowerCase()
  return readPersistedIdentity().userEmail
}

async function waitForSessionIdentity(timeoutMs = 1500): Promise<void> {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline) {
    if (currentUserId() !== null && currentUserEmail() !== null && currentAuthToken() !== null) return
    await new Promise<void>((resolve) => window.setTimeout(resolve, 50))
  }
}

function beginLoad(): LoadContext | null {
  activeLoadController?.abort()
  const userId = currentUserId()
  const userEmail = currentUserEmail()
  if (!userId || !userEmail) return null
  const controller = new AbortController()
  activeLoadController = controller
  loadGeneration += 1
  return {
    generation: loadGeneration,
    userId,
    userEmail,
    tokenPresent: Boolean(currentAuthToken()),
    controller,
  }
}

function isCurrentLoad(context: LoadContext): boolean {
  return context.generation === loadGeneration
    && !context.controller.signal.aborted
    && currentUserId() === context.userId
    && currentUserEmail() === context.userEmail
    && Boolean(currentAuthToken()) === context.tokenPresent
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError'
}

async function invalidateStandaloneServiceWorker(): Promise<void> {
  try {
    if ('serviceWorker' in navigator) {
      const registrations = await navigator.serviceWorker.getRegistrations()
      await Promise.all(registrations
        .filter((registration) => {
          try {
            return new URL(registration.scope).pathname.startsWith(STANDALONE_SERVICE_WORKER_PATH)
          } catch {
            return false
          }
        })
        .map((registration) => registration.unregister()))
    }
    if ('caches' in window) {
      const keys = await caches.keys()
      await Promise.all(keys
        .filter((key) => /^(gpt-image-playground|zayu-image-playground)/i.test(key))
        .map((key) => caches.delete(key)))
    }
  } catch (error) {
    // Cache and Service Worker APIs may be unavailable in private browsing;
    // the server-side feature gate remains authoritative in that case.
    console.warn('Failed to invalidate the standalone image playground cache:', error)
  }
}

function defaultModelForKey(key: ApiKey): string {
  return key.group?.platform === 'grok' ? 'grok-imagine' : 'gpt-image-2'
}

function normalizeApiKeyRecord(raw: unknown): ApiKey | null {
  if (!raw || typeof raw !== 'object' || Array.isArray(raw)) return null
  const record = raw as ApiKey & { id?: unknown; api_key?: unknown; apiKey?: unknown; group?: { id?: unknown } }
  const id = normalizeNumericId(record.id)
  if (id === null) return null
  const key = typeof record.key === 'string' && record.key.trim()
    ? record.key.trim()
    : typeof record.api_key === 'string'
      ? record.api_key.trim()
      : typeof record.apiKey === 'string'
        ? record.apiKey.trim()
        : ''
  const groupId = record.group_id == null
    ? normalizeNumericId(record.group?.id)
    : normalizeNumericId(record.group_id)
  return { ...record, id, key, group_id: groupId }
}

function keyHasRemainingQuota(key: ApiKey): boolean {
  const quota = Number(key.quota)
  const used = Number(key.quota_used)
  const hasLimit = Number.isFinite(quota) && quota > 0
  return !hasLimit || !Number.isFinite(used) || used < quota
}

function chooseDefaultModel(key: ApiKey, available: GatewayModel[]): string {
  const ids = available.map((model) => model.id || '').filter(Boolean)
  const preferred = key.group?.platform === 'grok'
    ? ['grok-imagine-image-quality', 'grok-imagine-image', 'grok-imagine']
    : ['gpt-image-2', 'gpt-image-1.5', 'gpt-image-1', 'dall-e-3']
  return preferred.find((model) => ids.includes(model)) || ids[0] || defaultModelForKey(key)
}

async function loadKeys(context: LoadContext | null = beginLoad()): Promise<void> {
  if (!context) {
    loadingKeys.value = false
    imageKeys.value = []
    selectedKeyId.value = null
    loadError.value = true
    return
  }
  loadingKeys.value = true
  loadError.value = false
  try {
    const keys: ApiKey[] = []
    let page = 1
    const availableGroups = await userGroupsAPI.getAvailable({ signal: context.controller.signal })
    if (!isCurrentLoad(context)) return
    const groupsById = new Map(availableGroups.map((group) => [String(group.id), group]))
    while (true) {
      const response = await keysAPI.list(page, 100, {
        status: 'active',
        sort_by: 'created_at',
        sort_order: 'desc',
      }, { signal: context.controller.signal })
      if (!isCurrentLoad(context)) return
      const responseRecord = response as unknown as {
        items?: unknown
        data?: { items?: unknown; pages?: unknown } | unknown[]
        pages?: unknown
      }
      const rawItems = Array.isArray(responseRecord.items)
        ? responseRecord.items
        : Array.isArray(responseRecord.data)
          ? responseRecord.data
          : responseRecord.data && !Array.isArray(responseRecord.data) && Array.isArray(responseRecord.data.items)
          ? responseRecord.data.items
          : []
      const normalizedItems = rawItems
        .map((raw) => normalizeApiKeyRecord(raw))
        .filter((key): key is ApiKey => key !== null)
        .map((key) => ({
          ...key,
          // Keep the embedded group as a fallback. Some mobile WebViews
          // return the group ID as a string while the group catalogue uses a
          // number, and dropping it would make a valid key disappear.
          group: key.group_id != null ? groupsById.get(String(key.group_id)) ?? key.group : key.group,
        }))
        .filter((key) => keyAllowsImageGeneration(key))
      keys.push(...normalizedItems)
      const pages = Number(responseRecord.pages ?? (
        responseRecord.data && !Array.isArray(responseRecord.data) ? responseRecord.data.pages : undefined
      ))
      if (page >= (Number.isFinite(pages) && pages > 0 ? pages : 1) || !rawItems.length) break
      page += 1
    }
    if (!isCurrentLoad(context)) return
    // Keep every active image-enabled key available to the standalone profile
    // selector. The backend remains authoritative for ownership, quota, and billing.
    imageKeys.value = keys
      .sort((a, b) => Number(keyHasRemainingQuota(b)) - Number(keyHasRemainingQuota(a)))
    if (!imageKeys.value.some((key) => key.id === selectedKeyId.value)) {
      selectedKeyId.value = imageKeys.value[0]?.id ?? null
    }
    if (selectedKey.value) await loadModels(context)
  } catch (error) {
    if (!isCurrentLoad(context) || isAbortError(error)) return
    console.error('Failed to load image playground keys:', error)
    imageKeys.value = []
    selectedKeyId.value = null
    loadError.value = true
  } finally {
    if (isCurrentLoad(context)) {
      loadingKeys.value = false
      if (activeLoadController === context.controller) activeLoadController = null
    }
  }
}

async function loadModels(context: LoadContext): Promise<void> {
  const key = selectedKey.value
  if (!key) {
    models.value = []
    selectedModel.value = null
    return
  }

  loadingModels.value = true
  try {
    // Only the selected key is sent to the provider. Other keys are carried to
    // the local standalone selector but are never probed in parallel.
    const response = await fetch(`${IMAGE_PLAYGROUND_API_BASE_URL}/models`, {
      headers: {
        Authorization: `Bearer ${key.key}`,
        'X-API-Key': key.key,
        'X-Sub2API-User-Email': context.userEmail,
      },
      cache: 'no-store',
      credentials: 'include',
      signal: context.controller.signal,
    })
    if (!isCurrentLoad(context) || selectedKey.value?.id !== key.id) return
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const body = await response.json() as { data?: GatewayModel[]; models?: GatewayModel[] } | GatewayModel[]
    const upstreamModels = Array.isArray(body) ? body : body.data || body.models || []
    const available = upstreamModels.filter((model) => {
      const id = String(model.id || '').toLowerCase()
      if (key.group?.platform === 'grok') {
        return id.startsWith('grok-imagine') && !id.includes('video')
      }
      if (key.group?.platform === 'openai') {
        return id.includes('image') || id.includes('dall-e')
      }
      return (id.includes('image') || id.includes('imagine') || id.includes('dall-e')) && !id.includes('video')
    })
    const ids = [...new Set(available.map((model) => model.id).filter((id): id is string => Boolean(id)))]
    const selectedAvailable = ids.length ? ids : [defaultModelForKey(key)]
    // Keep model lists for every key so switching profiles in the standalone
    // workbench never discards models fetched for another API key.
    modelsByKeyId.value = {
      ...modelsByKeyId.value,
      [key.id]: selectedAvailable,
    }
    models.value = selectedAvailable.map((id) => ({ id }))
    selectedModel.value = chooseDefaultModel(key, models.value)
  } catch (error) {
    if (!isCurrentLoad(context) || isAbortError(error)) return
    console.warn('Failed to load image playground models:', error)
    models.value = [{ id: defaultModelForKey(key) }]
    selectedModel.value = models.value[0].id || null
  } finally {
    if (isCurrentLoad(context)) {
      loadingModels.value = false
      if (selectedKey.value?.id === key.id && selectedModel.value) redirectToStandalone(context, key, selectedModel.value)
      if (activeLoadController === context.controller) activeLoadController = null
    }
  }
}

function buildStandaloneSettings(context: LoadContext, key: ApiKey, model: string): Record<string, unknown> {
  const profiles = imageKeys.value.map((candidate) => ({
    id: `sub2api-key-${candidate.id}`,
    name: candidate.name || `Sub2API API Key ${candidate.id}`,
    description: `Sub2API 分组 · ${candidate.group?.name || candidate.group_id || 'Image'}`,
    groupId: candidate.group_id ?? candidate.group?.id,
    provider: 'sb2api-async',
    baseUrl: IMAGE_PLAYGROUND_API_BASE_URL,
    apiKey: candidate.key,
    userEmail: context.userEmail,
    model: candidate.id === key.id ? model : defaultModelForKey(candidate),
    modelOptions: modelsByKeyId.value[candidate.id] || [defaultModelForKey(candidate)],
    isDefault: candidate.id === key.id,
    timeout: 600,
    apiMode: 'images',
    transparentBackgroundMethod: 'api',
  }))

  return {
    customProviders: [],
    profiles,
    activeProfileId: `sub2api-key-${key.id}`,
  }
}

function redirectToStandalone(context: LoadContext, key: ApiKey, model: string): void {
  if (redirecting.value || !isCurrentLoad(context) || currentUserId() !== context.userId) return
  try {
    redirecting.value = true
    const settings = buildStandaloneSettings(context, key, model)
    const payload = `${STANDALONE_CONFIG_PREFIX}${JSON.stringify({
      version: 1,
      issuedAt: Date.now(),
      userId: context.userId,
      userEmail: context.userEmail,
      settings,
    })}`
    let stored = false
    // Use same-origin session storage as the primary one-time carrier. Unlike
    // window.name, it is scoped to this tab and never travels with a referrer.
    try {
      sessionStorage.setItem(STANDALONE_CONFIG_STORAGE_KEY, payload)
      stored = true
    } catch (error) {
      console.warn('Failed to persist image playground session bootstrap:', error)
    }
    // A short-lived local fallback covers mobile WebViews/PWA transitions that
    // recreate the sessionStorage namespace during navigation.
    try {
      localStorage.setItem(STANDALONE_CONFIG_FALLBACK_STORAGE_KEY, payload)
      stored = true
    } catch (error) {
      console.warn('Failed to persist image playground fallback bootstrap:', error)
    }
    if (!stored) throw new Error('Browser storage is unavailable for image playground bootstrap')
    window.name = ''
    window.location.replace(STANDALONE_IMAGE_PLAYGROUND_PATH)
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
    if (!authStore.token || !authStore.user) authStore.checkAuth()
    await waitForSessionIdentity()
    // Refresh the authenticated identity before building the one-time
    // bootstrap payload. This prevents a stale/empty user id from being
    // carried when a session was restored just before opening the launcher.
    await authStore.refreshUser().catch(() => undefined)
    // Capture the context only after refreshUser settles. Creating it before
    // the refresh allows a concurrent account/token update to invalidate the
    // snapshot while the key requests are still starting.
    const context = beginLoad()
    await loadKeys(context)
  } finally {
    refreshing.value = false
  }
}

onMounted(async () => {
  // Remove a worker from an older standalone build before any navigation can
  // be redirected back to the static gallery entry.
  await invalidateStandaloneServiceWorker()
  await appStore.fetchPublicSettings()
  if (!imagePlaygroundEnabled.value) {
    loadingKeys.value = false
    return
  }
  if (!authStore.token || !authStore.user) authStore.checkAuth()
  await waitForSessionIdentity()
  // The standalone page binds its local cache to this identity. Resolve the
  // current user first so the bootstrap cannot be attributed to an older
  // session during a refresh or account switch.
  await authStore.refreshUser().catch(() => undefined)
  await loadKeys()
  balanceRefreshTimer = window.setInterval(() => {
    void authStore.refreshUser().catch(() => undefined)
  }, 15_000)
})

onBeforeUnmount(() => {
  activeLoadController?.abort()
  activeLoadController = null
  loadGeneration += 1
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

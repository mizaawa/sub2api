import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { createAccountMock, probeUpstreamBillingMock } = vi.hoisted(() => ({
  createAccountMock: vi.fn(),
  probeUpstreamBillingMock: vi.fn(),
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
    showWarning: vi.fn(),
    showInfo: vi.fn(),
  }),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ isSimpleMode: true }),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      create: createAccountMock,
      probeUpstreamBilling: probeUpstreamBillingMock,
      checkMixedChannelRisk: vi.fn().mockResolvedValue({ has_risk: false }),
    },
    settings: {
      getWebSearchEmulationConfig: vi.fn().mockResolvedValue({ enabled: false, providers: [] }),
      getSettings: vi.fn().mockResolvedValue({}),
    },
    tlsFingerprintProfiles: {
      list: vi.fn().mockResolvedValue([]),
    },
  },
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

import CreateAccountModal from '../CreateAccountModal.vue'

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: { show: { type: Boolean, default: false } },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

const ModelWhitelistSelectorStub = defineComponent({
  name: 'ModelWhitelistSelector',
  template: '<div data-testid="model-whitelist" />',
})

function mountModal() {
  return mount(CreateAccountModal, {
    props: { show: true, proxies: [], groups: [] },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        ConfirmDialog: true,
        Select: true,
        Icon: true,
        PlatformIcon: true,
        ProxySelector: true,
        ProxyAdBanner: true,
        GroupSelector: true,
        ModelWhitelistSelector: ModelWhitelistSelectorStub,
        QuotaLimitCard: true,
      },
    },
  })
}

async function selectPlatform(wrapper: ReturnType<typeof mountModal>, label: string) {
  const button = wrapper.findAll('button').find((candidate) => candidate.text().includes(label))
  expect(button, `platform button ${label}`).toBeDefined()
  await button?.trigger('click')
}

async function submitAPIKeyAccount(
  platform: 'anthropic' | 'openai' | 'gemini' | 'antigravity' | 'grok' | 'custom',
) {
  const wrapper = mountModal()
  const labels: Record<typeof platform, string> = {
    anthropic: 'Anthropic',
    openai: 'OpenAI',
    gemini: 'Gemini',
    antigravity: 'Antigravity',
    grok: 'Grok',
    custom: 'Custom',
  }
  await selectPlatform(wrapper, labels[platform])
  await wrapper.get('[data-tour="account-form-name"]').setValue(`${platform} account`)
  if (platform === 'custom') {
    await wrapper.get('input[placeholder="https://api.example.com/v1"]').setValue('https://relay.example/v1')
  }
  await wrapper.get('input[type="password"]').setValue('test-api-key')
  await wrapper.get('form#create-account-form').trigger('submit.prevent')
  await flushPromises()
  return wrapper
}

describe('CreateAccountModal API-key-only creation', () => {
  beforeEach(() => {
    createAccountMock.mockReset().mockImplementation(async (payload) => ({
      id: 42,
      platform: payload.platform,
      type: payload.type,
    }))
    probeUpstreamBillingMock.mockReset().mockResolvedValue({})
  })

  it('offers all six platforms without an account-type chooser', () => {
    const wrapper = mountModal()
    for (const label of ['Anthropic', 'OpenAI', 'Gemini', 'Antigravity', 'Grok', 'Custom']) {
      expect(wrapper.findAll('button').some((button) => button.text().includes(label))).toBe(true)
    }
    expect(wrapper.find('[data-tour="account-form-type"]').exists()).toBe(false)
    expect(wrapper.find('input[type="radio"]').exists()).toBe(false)
  })

  it.each(['anthropic', 'openai', 'gemini', 'antigravity', 'grok', 'custom'] as const)(
    'submits %s through the API Key account path',
    async (platform) => {
      await submitAPIKeyAccount(platform)

      expect(createAccountMock.mock.calls[0]?.[0]).toMatchObject({
        platform,
        type: 'apikey',
      })
    },
  )

  it('uses API Key, model whitelist, and unlimited defaults', async () => {
    const wrapper = await submitAPIKeyAccount('openai')

    expect(wrapper.find('[data-testid="model-whitelist"]').exists()).toBe(true)
    expect(createAccountMock).toHaveBeenCalledTimes(1)
    const payload = createAccountMock.mock.calls[0]?.[0]
    expect(payload).toMatchObject({
      platform: 'openai',
      type: 'apikey',
      concurrency: 0,
      load_factor: 0,
      upstream_billing_probe_enabled: true,
      upstream_billing_rate_sync_enabled: true,
    })
    expect(payload).not.toHaveProperty('auto_pause_on_expired')
    expect(probeUpstreamBillingMock).toHaveBeenCalledWith(42)
  })

  it('creates Custom as an isolated OpenAI-compatible API-key account', async () => {
    await submitAPIKeyAccount('custom')

    expect(createAccountMock.mock.calls[0]?.[0]).toMatchObject({
      platform: 'custom',
      type: 'apikey',
      credentials: {
        base_url: 'https://relay.example/v1',
        api_key: 'test-api-key',
      },
    })
    expect(createAccountMock.mock.calls[0]?.[0]?.credentials).not.toHaveProperty('model_mapping')
  })

  it('requires a Base URL for Custom accounts', async () => {
    const wrapper = mountModal()
    await selectPlatform(wrapper, 'Custom')
    await wrapper.get('[data-tour="account-form-name"]').setValue('custom account')
    await wrapper.get('input[type="password"]').setValue('test-api-key')
    await wrapper.get('form#create-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(createAccountMock).not.toHaveBeenCalled()
  })

  it('uses the unified API Key form for Antigravity', async () => {
    const wrapper = await submitAPIKeyAccount('antigravity')

    expect(createAccountMock.mock.calls[0]?.[0]).toMatchObject({
      platform: 'antigravity',
      type: 'apikey',
      credentials: {
        base_url: 'https://cloudcode-pa.googleapis.com',
        api_key: 'test-api-key',
      },
    })
    expect(wrapper.find('[data-testid="antigravity-project-id-input"]').exists()).toBe(false)
    expect(wrapper.find('oauth-authorization-flow-stub').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('admin.accounts.antigravityProjectIdLabel')
    expect(wrapper.text()).not.toContain('admin.accounts.types.oauth')
  })

  it('ties probing to the only visible upstream rate-sync switch', async () => {
    const wrapper = mountModal()
    await selectPlatform(wrapper, 'OpenAI')
    await wrapper.get('[data-tour="account-form-name"]').setValue('OpenAI account')
    await wrapper.get('input[type="password"]').setValue('test-api-key')
    await wrapper.get('[data-testid="upstream-billing-rate-sync"]').trigger('click')
    await wrapper.get('form#create-account-form').trigger('submit.prevent')
    await flushPromises()

    const payload = createAccountMock.mock.calls[0]?.[0]
    expect(payload?.upstream_billing_probe_enabled).toBe(false)
    expect(payload?.upstream_billing_rate_sync_enabled).toBe(false)
    expect(wrapper.find('[data-testid="upstream-billing-auto-probe"]').exists()).toBe(false)
    expect(probeUpstreamBillingMock).not.toHaveBeenCalled()
  })

  it('keeps the OpenAI long-context setting on API Key accounts', async () => {
    const wrapper = mountModal()
    await selectPlatform(wrapper, 'OpenAI')
    await wrapper.get('[data-tour="account-form-name"]').setValue('OpenAI account')
    await wrapper.get('input[type="password"]').setValue('test-api-key')
    await wrapper.get('[data-testid="openai-long-context-billing-toggle"]').trigger('click')
    await wrapper.get('form#create-account-form').trigger('submit.prevent')
    await flushPromises()

    expect(createAccountMock.mock.calls[0]?.[0]?.extra?.openai_long_context_billing_enabled).toBe(true)
  })
})

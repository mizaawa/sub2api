import { defineComponent } from 'vue'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { UserMonitorView } from '@/api/channelMonitor'
import ChannelStatusView from '@/views/user/ChannelStatusView.vue'

const {
  listMonitors,
  showError,
  setAutoRefreshEnabled,
  stopAutoRefresh,
  startAutoRefresh,
  resetCountdown,
} = vi.hoisted(() => ({
  listMonitors: vi.fn(),
  showError: vi.fn(),
  setAutoRefreshEnabled: vi.fn(),
  stopAutoRefresh: vi.fn(),
  startAutoRefresh: vi.fn(),
  resetCountdown: vi.fn(),
}))

vi.mock('@/api/channelMonitor', () => ({
  list: listMonitors,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: { channel_monitor_enabled: true, site_name: 'Monitor Test Site' },
    siteName: 'Fallback Site',
    showError,
  }),
}))

vi.mock('@/composables/useAutoRefresh', async () => {
  const { ref } = await import('vue')
  return {
    useAutoRefresh: () => ({
      enabled: ref(false),
      intervalSeconds: ref(60),
      countdown: ref(60),
      fetching: ref(false),
      intervals: [30, 60, 120],
      setEnabled: setAutoRefreshEnabled,
      setInterval: vi.fn(),
      resetCountdown,
      start: startAutoRefresh,
      stop: stopAutoRefresh,
    }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  const translations: Record<string, string> = {
    'channelStatus.title': 'Channel Status',
    'channelStatus.poweredBy': 'Powered by',
    'channelStatus.metricsDisclaimer': 'Availability metrics are aggregated from channel monitoring checks.',
    'channelStatus.emptyProvider': 'No monitored channels configured',
    'channelStatus.uptimeUnavailable': '-- uptime',
    'channelStatus.providers.openai': 'OpenAI',
    'channelStatus.providers.anthropic': 'Anthropic',
    'channelStatus.providers.gemini': 'Gemini',
    'channelStatus.providers.grok': 'Grok',
    'channelStatus.providers.custom': 'Custom',
    'monitorCommon.status.operational': 'Operational',
    'monitorCommon.status.degraded': 'Degraded',
    'monitorCommon.status.failed': 'Failed',
    'monitorCommon.status.error': 'Error',
    'monitorCommon.status.unknown': 'Unknown',
  }
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'channelStatus.componentCount') return `${params?.n} components`
        if (key === 'channelStatus.loadingProvider') return `Loading ${params?.provider}`
        if (key === 'channelStatus.uptime') return `${params?.value}% uptime`
        if (key === 'channelStatus.timelineLabel') return `${params?.n} recent status checks`
        return translations[key] ?? key
      },
    }),
  }
})

const AppLayoutStub = defineComponent({
  template: '<main><slot /></main>',
})

function makeMonitor(overrides: Partial<UserMonitorView> = {}): UserMonitorView {
  return {
    id: 1,
    sort_order: 0,
    name: 'Chat Completions',
    provider: 'openai',
    group_name: '',
    primary_model: 'gpt-4o-mini',
    primary_status: 'operational',
    primary_latency_ms: 120,
    primary_ping_latency_ms: 20,
    availability_7d: 100,
    extra_models: [],
    timeline: [
      {
        status: 'operational',
        latency_ms: 120,
        ping_latency_ms: 20,
        checked_at: '2026-09-19T00:00:00Z',
      },
    ],
    ...overrides,
  }
}

function mountView(): VueWrapper {
  return mount(ChannelStatusView, {
    global: {
      stubs: { AppLayout: AppLayoutStub },
    },
  })
}

describe('ChannelStatusView', () => {
  beforeEach(() => {
    for (const mock of [
      listMonitors,
      showError,
      setAutoRefreshEnabled,
      stopAutoRefresh,
      startAutoRefresh,
      resetCountdown,
    ]) {
      mock.mockReset()
    }
  })

  it('groups monitors into five fixed provider sections and renders the status-page contract', async () => {
    listMonitors.mockResolvedValue({
      items: [
        makeMonitor({ id: 5, name: 'Custom Gateway', provider: 'custom' as UserMonitorView['provider'], availability_7d: 99.95 }),
        makeMonitor({ id: 6, name: 'Custom Backup', provider: 'custom' as UserMonitorView['provider'] }),
        makeMonitor({ id: 7, name: 'Custom Third', provider: 'custom' as UserMonitorView['provider'] }),
        makeMonitor({ id: 2, name: 'Responses', provider: 'openai', availability_7d: 99.5 }),
        makeMonitor({ id: 4, name: 'Grok Chat', provider: 'grok' }),
        makeMonitor({ id: 1, name: 'Chat Completions', provider: 'openai' }),
      ],
    })

    const wrapper = mountView()
    await flushPromises()

    const sections = wrapper.findAll('section[data-testid^="monitor-provider-"]')
    expect(sections).toHaveLength(5)
    expect(sections.map(section => section.attributes('data-testid'))).toEqual([
      'monitor-provider-openai',
      'monitor-provider-anthropic',
      'monitor-provider-gemini',
      'monitor-provider-grok',
      'monitor-provider-custom',
    ])
    expect(sections.map(section => section.find('button').text())).toEqual([
      'OpenAI2 components',
      'Anthropic0 components',
      'Gemini0 components',
      'Grok1 components',
      'Custom3 components',
    ])

    const providerGrid = wrapper.get('[data-testid="monitor-provider-grid"]')
    expect(providerGrid.classes()).toContain('lg:grid-cols-2')
    expect(wrapper.get('[data-testid="monitor-provider-custom"]').classes()).toContain('lg:col-span-2')

    const openAISection = wrapper.get('[data-testid="monitor-provider-openai"]')
    expect(openAISection.text()).toContain('Chat Completions')
    expect(openAISection.text()).toContain('Responses')
    expect(openAISection.text()).not.toContain('Custom Gateway')
    expect(openAISection.findAll('[data-testid^="monitor-status-row-"]').map(row => row.text()))
      .toEqual([
        expect.stringContaining('Responses'),
        expect.stringContaining('Chat Completions'),
      ])
    const customItems = wrapper.get('[data-testid="monitor-provider-items-custom"]')
    expect(customItems.classes()).toContain('monitor-items--custom')
    expect(customItems.findAll('.monitor-item').map(item => item.text())).toEqual([
      expect.stringContaining('Custom Gateway'),
      expect.stringContaining('Custom Backup'),
      expect.stringContaining('Custom Third'),
    ])
    expect(wrapper.get('[data-testid="monitor-status-row-1"]').text()).toContain('100% uptime')
    expect(wrapper.get('[data-testid="monitor-status-row-5"]').text()).toContain('99.95% uptime')

    expect(wrapper.text()).not.toContain('View history')
    expect(wrapper.text()).not.toContain('7 days')
    expect(wrapper.text()).not.toContain('System status')
    const notice = wrapper.get('[data-testid="channel-status-notice"]')
    expect(notice.get('h1').text()).toBe('Monitor Test Site')
    expect(notice.text()).toContain('Availability metrics are aggregated from channel monitoring checks.')
    expect(wrapper.get('[data-testid="channel-status-grid"]').text()).not.toContain(
      'Availability metrics are aggregated from channel monitoring checks.'
    )
    expect(wrapper.get('[data-testid="monitor-status-row-1"]').findAll('.monitor-status-bar__segment')).toHaveLength(120)
    const poweredBy = wrapper.get('a[href="https://mizaawa.com"]')
    expect(poweredBy.text()).toBe('mizaawa.com')
    expect(poweredBy.attributes('target')).toBe('_blank')
    expect(poweredBy.attributes('rel')).toBe('noopener noreferrer')

    expect(listMonitors).toHaveBeenCalledTimes(1)
    expect(listMonitors).toHaveBeenCalledWith({ signal: expect.any(AbortSignal) })
    expect(setAutoRefreshEnabled).toHaveBeenCalledWith(true)
    wrapper.unmount()
  })

  it('starts expanded and collapses only the selected provider', async () => {
    listMonitors.mockResolvedValue({ items: [makeMonitor()] })
    const wrapper = mountView()
    await flushPromises()

    const openAIToggle = wrapper.get('[data-testid="monitor-provider-toggle-openai"]')
    const openAIContent = wrapper.get('#monitor-provider-content-openai')
    const customContent = wrapper.get('#monitor-provider-content-custom')
    expect(openAIToggle.attributes('aria-expanded')).toBe('true')
    expect(openAIContent.isVisible()).toBe(true)
    expect(customContent.isVisible()).toBe(true)

    await openAIToggle.trigger('click')
    expect(openAIToggle.attributes('aria-expanded')).toBe('false')
    expect(wrapper.get('#monitor-provider-content-openai').attributes('style')).toContain('display: none')
    expect(customContent.isVisible()).toBe(true)

    await openAIToggle.trigger('click')
    expect(openAIToggle.attributes('aria-expanded')).toBe('true')
    expect(wrapper.get('#monitor-provider-content-openai').attributes('style') || '').not.toContain('display: none')
    wrapper.unmount()
  })

  it('settles into the five empty provider sections after a load failure', async () => {
    listMonitors.mockRejectedValue(new Error('status unavailable'))
    const wrapper = mountView()
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('status unavailable')
    expect(wrapper.findAll('[data-testid^="monitor-provider-empty-"]')).toHaveLength(5)
    expect(wrapper.findAll('[data-testid^="monitor-provider-loading-"]')).toHaveLength(0)
    wrapper.unmount()
  })
})

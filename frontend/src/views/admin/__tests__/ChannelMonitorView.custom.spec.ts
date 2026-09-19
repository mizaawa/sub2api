import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import MonitorFiltersBar from '@/components/admin/monitor/MonitorFiltersBar.vue'
import MonitorFormDialog from '@/components/admin/monitor/MonitorFormDialog.vue'
import MonitorKeyPickerDialog from '@/components/admin/monitor/MonitorKeyPickerDialog.vue'
import Select from '@/components/common/Select.vue'
import type { ApiKey, GroupPlatform } from '@/types'
import {
  PROVIDERS,
  PROVIDER_CUSTOM,
} from '@/constants/channelMonitor'

const { listTemplates, listKeys, getUserGroupRates } = vi.hoisted(() => ({
  listTemplates: vi.fn(),
  listKeys: vi.fn(),
  getUserGroupRates: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    channelMonitor: {
      create: vi.fn(),
      update: vi.fn(),
    },
    channelMonitorTemplate: {
      list: listTemplates,
    },
  },
}))

vi.mock('@/api/keys', () => ({
  keysAPI: { list: listKeys },
}))

vi.mock('@/api/groups', () => ({
  userGroupsAPI: { getUserGroupRates },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    cachedPublicSettings: null,
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const BaseDialogStub = defineComponent({
  props: { show: { type: Boolean, default: false } },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

const MonitorKeyPickerDialogStub = defineComponent({
  name: 'MonitorKeyPickerDialog',
  props: {
    show: Boolean,
    loading: Boolean,
    keys: { type: Array, default: () => [] },
    provider: String,
  },
  template: '<div data-testid="key-picker-stub" />',
})

function mountDialog() {
  return mount(MonitorFormDialog, {
    props: { show: true, monitor: null },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Toggle: true,
        Select: true,
        ModelTagInput: true,
        MonitorKeyPickerDialog: MonitorKeyPickerDialogStub,
        MonitorAdvancedRequestConfig: true,
      },
    },
  })
}

function key(id: number, platform: GroupPlatform): ApiKey {
  return {
    id,
    name: `${platform} key`,
    key: `sk-${platform}`,
    status: 'active',
    expires_at: null,
    group: {
      id,
      name: `${platform} group`,
      platform,
      subscription_type: 'standard',
      rate_multiplier: 1,
    },
  } as ApiKey
}

describe('channel monitor Custom provider', () => {
  beforeEach(() => {
    listTemplates.mockReset().mockResolvedValue({ items: [] })
    listKeys.mockReset().mockResolvedValue({ items: [], page: 1, page_size: 100, pages: 1, total: 0 })
    getUserGroupRates.mockReset().mockResolvedValue({})
  })

  it('offers Custom as the fifth provider with the supplied seven-node icon', async () => {
    const wrapper = mountDialog()
    await flushPromises()

    expect(PROVIDERS).toContain(PROVIDER_CUSTOM)
    const providerButtons = wrapper.findAll('[data-testid^="monitor-provider-"]')
    expect(providerButtons).toHaveLength(5)
    expect(providerButtons[0].element.parentElement?.className).toContain('sm:grid-cols-3')
    expect(providerButtons[0].element.parentElement?.className).toContain('md:grid-cols-5')

    const customButton = wrapper.get('[data-testid="monitor-provider-custom"]')
    expect(customButton.text()).toContain('monitorCommon.providers.custom')
    expect(customButton.get('svg').attributes('viewBox')).toBe('0 0 48 48')
    expect(customButton.findAll('circle')).toHaveLength(7)
  })

  it('includes Custom in the admin provider filter', () => {
    const wrapper = mount(MonitorFiltersBar, {
      props: {
        loading: false,
        search: '',
        provider: '',
        enabled: '',
        'onUpdate:search': () => undefined,
        'onUpdate:provider': () => undefined,
        'onUpdate:enabled': () => undefined,
      },
    })

    const options = wrapper.findAllComponents(Select)[0].props('options') as Array<{
      value: string
      label: string
    }>
    expect(options).toContainEqual({
      value: PROVIDER_CUSTOM,
      label: 'monitorCommon.providers.custom',
    })
  })

  it('shows keys from every group platform when Custom is selected', () => {
    const keys = [
      key(1, 'openai'),
      key(2, 'anthropic'),
      key(3, 'gemini'),
      key(4, 'grok'),
      key(5, 'antigravity'),
      key(6, 'composite'),
    ]
    const wrapper = mount(MonitorKeyPickerDialog, {
      props: {
        show: true,
        loading: false,
        keys,
        provider: PROVIDER_CUSTOM,
      },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          GroupBadge: { template: '<span />' },
        },
      },
    })

    expect(wrapper.findAll('tbody tr')).toHaveLength(keys.length)
  })

  it('loads every active-key page before opening the Custom key picker', async () => {
    listKeys
      .mockResolvedValueOnce({
        items: [key(1, 'openai')],
        page: 1,
        page_size: 100,
        pages: 2,
        total: 2,
      })
      .mockResolvedValueOnce({
        items: [key(2, 'composite')],
        page: 2,
        page_size: 100,
        pages: 2,
        total: 2,
      })

    const wrapper = mountDialog()
    await flushPromises()
    await wrapper.get('[data-testid="monitor-provider-custom"]').trigger('click')
    const useKeyButton = wrapper.findAll('button').find(button =>
      button.text().includes('admin.channelMonitor.form.useMyKey'),
    )
    expect(useKeyButton).toBeDefined()
    await useKeyButton!.trigger('click')
    await flushPromises()

    expect(listKeys).toHaveBeenNthCalledWith(1, 1, 100, { status: 'active' })
    expect(listKeys).toHaveBeenNthCalledWith(2, 2, 100, { status: 'active' })
    const picker = wrapper.getComponent(MonitorKeyPickerDialogStub)
    expect(picker.props('provider')).toBe(PROVIDER_CUSTOM)
    expect((picker.props('keys') as ApiKey[]).map(item => item.id)).toEqual([1, 2])
  })
})

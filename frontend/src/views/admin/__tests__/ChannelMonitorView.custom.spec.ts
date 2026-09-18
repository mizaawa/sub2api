import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import MonitorFiltersBar from '@/components/admin/monitor/MonitorFiltersBar.vue'
import MonitorFormDialog from '@/components/admin/monitor/MonitorFormDialog.vue'
import Select from '@/components/common/Select.vue'
import {
  API_MODE_RESPONSES,
  PROVIDERS,
  PROVIDER_CUSTOM,
} from '@/constants/channelMonitor'

const { createMonitor, listTemplates } = vi.hoisted(() => ({
  createMonitor: vi.fn(),
  listTemplates: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    channelMonitor: {
      create: createMonitor,
      update: vi.fn(),
    },
    channelMonitorTemplate: {
      list: listTemplates,
    },
  },
}))

vi.mock('@/api/keys', () => ({
  keysAPI: { list: vi.fn() },
}))

vi.mock('@/api/groups', () => ({
  userGroupsAPI: { getUserGroupRates: vi.fn() },
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

function mountDialog() {
  return mount(MonitorFormDialog, {
    props: { show: true, monitor: null },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        Toggle: true,
        Select: true,
        ModelTagInput: true,
        MonitorKeyPickerDialog: true,
        MonitorAdvancedRequestConfig: true,
      },
    },
  })
}

describe('channel monitor Custom provider', () => {
  beforeEach(() => {
    createMonitor.mockReset().mockResolvedValue({})
    listTemplates.mockReset().mockResolvedValue({ items: [] })
  })

  it('offers Custom with both OpenAI-compatible API modes and submits the selected mode', async () => {
    const wrapper = mountDialog()
    await flushPromises()

    expect(PROVIDERS).toContain(PROVIDER_CUSTOM)
    const customButton = wrapper.get('[data-testid="monitor-provider-custom"]')
    expect(customButton.text()).toContain('monitorCommon.providers.custom')
    expect(customButton.get('svg').attributes('viewBox')).toBe('0 0 48 48')

    await customButton.trigger('click')
    expect(wrapper.get('[data-testid="monitor-api-mode-chat_completions"]').exists()).toBe(true)
    const responsesButton = wrapper.get('[data-testid="monitor-api-mode-responses"]')
    await responsesButton.trigger('click')

    await wrapper.get('input[placeholder="admin.channelMonitor.form.namePlaceholder"]').setValue('Custom upstream')
    await wrapper.get('[data-testid="monitor-endpoint"]').setValue('https://custom.example.com')
    await wrapper.get('input[type="password"]').setValue('sk-custom')
    await wrapper.get('[data-testid="monitor-primary-model"]').setValue('custom-model')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(createMonitor).toHaveBeenCalledWith(expect.objectContaining({
      provider: PROVIDER_CUSTOM,
      api_mode: API_MODE_RESPONSES,
      endpoint: 'https://custom.example.com',
      primary_model: 'custom-model',
    }))
  })

  it('includes Custom in the provider filter', () => {
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

    const providerOptions = wrapper.findAllComponents(Select)[0].props('options') as Array<{
      value: string
      label: string
    }>
    expect(providerOptions).toContainEqual({
      value: PROVIDER_CUSTOM,
      label: 'monitorCommon.providers.custom',
    })
  })
})

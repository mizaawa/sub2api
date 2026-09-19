import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import MonitorAdvancedRequestConfig from './MonitorAdvancedRequestConfig.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

function mountConfig() {
  return mount(MonitorAdvancedRequestConfig, {
    props: {
      provider: 'custom',
      apiMode: 'chat_completions',
      extraHeaders: {},
      bodyOverrideMode: 'off',
      bodyOverride: null,
    },
  })
}

describe('MonitorAdvancedRequestConfig header protection', () => {
  it.each([
    'Authorization',
    'Proxy-Authorization',
    'X-API-Key',
    'x-goog-api-key',
    'X-Sub2API-Monitor-Timestamp',
    'x-sub2api-monitor-signature',
  ])('rejects the managed header %s case-insensitively', async (headerName) => {
    const wrapper = mountConfig()
    const input = wrapper.findAll('input')[0]

    await input.setValue(headerName)
    await input.trigger('blur')

    expect(wrapper.text()).toContain('admin.channelMonitor.advanced.headerNameForbidden')
    expect(wrapper.emitted('update:extraHeaders')).toBeUndefined()
  })

  it('still emits ordinary custom headers', async () => {
    const wrapper = mountConfig()
    const [name, value] = wrapper.findAll('input')

    await name.setValue('X-Monitor-Client')
    await value.setValue('status-probe')
    await value.trigger('blur')

    expect(wrapper.emitted('update:extraHeaders')?.at(-1)).toEqual([
      { 'X-Monitor-Client': 'status-probe' },
    ])
  })
})

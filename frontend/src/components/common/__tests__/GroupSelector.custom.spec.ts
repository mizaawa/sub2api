import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import GroupSelector from '@/components/common/GroupSelector.vue'
import type { AccountPlatform, AdminGroup, GroupPlatform } from '@/types'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const GroupBadgeStub = defineComponent({
  props: { name: { type: String, required: true } },
  template: '<span>{{ name }}</span>',
})

const groups = [
  { id: 1, name: 'Custom group', platform: 'composite', rate_multiplier: 1 },
  { id: 2, name: 'OpenAI group', platform: 'openai', rate_multiplier: 1 },
  { id: 3, name: 'Anthropic group', platform: 'anthropic', rate_multiplier: 1 },
  { id: 4, name: 'Antigravity group', platform: 'antigravity', rate_multiplier: 1 },
  { id: 5, name: 'Gemini group', platform: 'gemini', rate_multiplier: 1 },
] as AdminGroup[]

function mountSelector(platform: AccountPlatform | GroupPlatform, mixedScheduling = false) {
  return mount(GroupSelector, {
    props: {
      modelValue: [],
      groups,
      platform,
      mixedScheduling,
      searchable: false,
    },
    global: {
      stubs: {
        GroupBadge: GroupBadgeStub,
        Icon: true,
      },
    },
  })
}

describe('GroupSelector Custom isolation', () => {
  it('offers Custom accounts only composite-backed Custom groups', () => {
    const wrapper = mountSelector('custom')

    expect(wrapper.text()).toContain('Custom group')
    expect(wrapper.text()).not.toContain('OpenAI group')
    expect(wrapper.text()).not.toContain('Anthropic group')
  })

  it('does not expose Custom groups to regular account platforms', () => {
    const wrapper = mountSelector('openai')

    expect(wrapper.text()).toContain('OpenAI group')
    expect(wrapper.text()).not.toContain('Custom group')
  })

  it('does not expose Custom groups to Antigravity mixed scheduling', () => {
    const wrapper = mountSelector('antigravity', true)

    expect(wrapper.text()).toContain('Antigravity group')
    expect(wrapper.text()).toContain('Anthropic group')
    expect(wrapper.text()).toContain('Gemini group')
    expect(wrapper.text()).not.toContain('Custom group')
  })
})

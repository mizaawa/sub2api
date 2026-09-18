import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import PlatformIcon from '@/components/common/PlatformIcon.vue'

describe('PlatformIcon Custom platform', () => {
  it('renders the Custom mark with theme-aware strokes', () => {
    const wrapper = mount(PlatformIcon, {
      props: { platform: 'custom', size: 'lg' },
    })

    const svg = wrapper.get('svg')
    expect(svg.attributes('viewBox')).toBe('0 0 48 48')
    expect(svg.attributes('stroke')).toBe('currentColor')
    expect(svg.classes()).toEqual(expect.arrayContaining(['w-5', 'h-5']))
    expect(svg.findAll('path')).toHaveLength(2)
  })
})

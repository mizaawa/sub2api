import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import PlatformIcon from '@/components/common/PlatformIcon.vue'

describe('PlatformIcon Custom mark', () => {
  it.each(['custom', 'composite'] as const)(
    'renders the supplied seven-node mark for %s',
    (platform) => {
      const wrapper = mount(PlatformIcon, {
        props: { platform, size: 'lg' },
      })

      const svg = wrapper.get('svg')
      expect(svg.attributes('viewBox')).toBe('0 0 48 48')
      expect(svg.classes()).toEqual(expect.arrayContaining(['w-5', 'h-5']))
      expect(svg.findAll('path')).toHaveLength(7)
      for (const path of svg.findAll('path')) {
        expect(path.attributes('stroke')).toBe('currentColor')
      }
    },
  )
})

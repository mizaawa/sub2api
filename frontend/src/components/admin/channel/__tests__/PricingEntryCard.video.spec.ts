import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import IntervalRow from '../IntervalRow.vue'
import PricingEntryCard from '../PricingEntryCard.vue'
import type { IntervalFormEntry, PricingFormEntry } from '../types'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const emptyInterval = (tierLabel = ''): IntervalFormEntry => ({
  min_tokens: 0,
  max_tokens: null,
  tier_label: tierLabel,
  input_price: null,
  output_price: null,
  cache_write_price: null,
  cache_read_price: null,
  per_request_price: null,
  sort_order: 0,
})

const videoEntry = (intervals: IntervalFormEntry[] = []): PricingFormEntry => ({
  models: ['videos-mini-480p'],
  billing_mode: 'video',
  input_price: null,
  output_price: null,
  cache_write_price: null,
  cache_read_price: null,
  image_input_price: null,
  image_output_price: null,
  per_request_price: null,
  intervals,
})

describe('channel video pricing controls', () => {
  it('creates the four standard video resolution tiers in order', async () => {
    const wrapper = mount(PricingEntryCard, {
      props: { entry: videoEntry(), platform: 'custom' },
      global: {
        stubs: {
          Icon: true,
          Select: true,
          ModelTagInput: true,
          IntervalRow: defineComponent({ template: '<div />' }),
        },
      },
    })

    const addTier = () => wrapper.findAll('button').find(button => (
      button.text().includes('admin.channels.form.addTier')
    ))

    for (const expected of ['480p', '720p', '1080p', '4k']) {
      const button = addTier()
      if (!button) throw new Error('Add tier button not found')
      await button.trigger('click')
      const updates = wrapper.emitted<PricingFormEntry[]>('update') ?? []
      const updated = updates.at(-1)?.[0]
      expect(updated?.intervals.at(-1)?.tier_label).toBe(expected)
      await wrapper.setProps({ entry: updated })
    }
  })

  it('labels video tiers as resolutions and shows all standard placeholders', () => {
    const wrapper = mount(IntervalRow, {
      props: { interval: emptyInterval('480p'), mode: 'video' },
      global: { stubs: { Icon: true } },
    })

    expect(wrapper.text()).toContain('admin.channels.form.resolution')
    expect(wrapper.get('input[type="text"]').attributes('placeholder')).toBe(
      '480p / 720p / 1080p / 4k',
    )
  })
})

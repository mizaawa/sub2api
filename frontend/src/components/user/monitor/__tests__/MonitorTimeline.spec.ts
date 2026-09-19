import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import type { MonitorTimelinePoint } from '@/api/channelMonitor'
import MonitorTimeline from '@/components/user/monitor/MonitorTimeline.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        if (key === 'channelStatus.timelineLabel') return `${params?.n} recent status checks`
        if (key.startsWith('monitorCommon.status.')) return key.split('.').at(-1)
        return key
      },
    }),
  }
})

function point(status: MonitorTimelinePoint['status'], minute: number): MonitorTimelinePoint {
  return {
    status,
    latency_ms: 100 + minute,
    ping_latency_ms: 20,
    checked_at: `2026-09-19T00:${String(minute).padStart(2, '0')}:00Z`,
  }
}

describe('MonitorTimeline', () => {
  it('pads to a stable length and renders each status at the same segment height', () => {
    const wrapper = mount(MonitorTimeline, {
      props: {
        length: 8,
        buckets: [
          point('operational', 4),
          point('degraded', 3),
          point('failed', 2),
          point('error', 1),
        ],
      },
    })

    const segments = wrapper.findAll('.monitor-status-bar__segment')
    expect(segments).toHaveLength(8)
    expect(segments.slice(0, 4).every(segment => segment.classes().includes('monitor-status-bar__segment--empty'))).toBe(true)
    expect(segments.slice(4).map(segment => segment.classes().find(name => name.includes('--')))).toEqual([
      'monitor-status-bar__segment--error',
      'monitor-status-bar__segment--failed',
      'monitor-status-bar__segment--degraded',
      'monitor-status-bar__segment--operational',
    ])
    expect(segments.every(segment => segment.attributes('style') === undefined)).toBe(true)
    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe(
      '4 recent status checks. operational: 1; degraded: 1; failed: 1; error: 1',
    )
  })

  it('renders an empty fixed-width timeline without history', () => {
    const wrapper = mount(MonitorTimeline, { props: { length: 6, buckets: [] } })
    const segments = wrapper.findAll('.monitor-status-bar__segment--empty')
    expect(segments).toHaveLength(6)
    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe('0 recent status checks')
  })
})

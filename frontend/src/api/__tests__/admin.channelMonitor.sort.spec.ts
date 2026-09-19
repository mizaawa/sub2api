import { beforeEach, describe, expect, it, vi } from 'vitest'

const { put } = vi.hoisted(() => ({
  put: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  apiClient: { put },
}))

import { updateSortOrder } from '@/api/admin/channelMonitor'

describe('admin channel monitor sort API', () => {
  beforeEach(() => {
    put.mockReset()
    put.mockResolvedValue({ data: { message: 'ok' } })
  })

  it('sends the complete order to the dedicated sort endpoint', async () => {
    const updates = [
      { id: 8, sort_order: 0 },
      { id: 3, sort_order: 10 },
    ]

    await expect(updateSortOrder(updates)).resolves.toEqual({ message: 'ok' })
    expect(put).toHaveBeenCalledWith('/admin/channel-monitors/sort-order', { updates })
  })
})

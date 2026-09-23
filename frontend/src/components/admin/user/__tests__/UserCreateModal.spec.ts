import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

const { createUser, showError, showSuccess, stepUpRun } = vi.hoisted(() => ({
  createUser: vi.fn(),
  showError: vi.fn(),
  showSuccess: vi.fn(),
  stepUpRun: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    users: { create: createUser }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess })
}))

vi.mock('@/composables/useStepUp', () => ({
  useStepUp: () => ({ run: stepUpRun }),
  isStepUpBlocked: () => false,
  isStepUpCancelled: () => false,
  stepUpBlockReason: () => ''
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key })
  }
})

import UserCreateModal from '../UserCreateModal.vue'

const mountModal = () => mount(UserCreateModal, {
  props: { show: true },
  global: {
    stubs: {
      BaseDialog: {
        props: ['show'],
        template: '<div v-if="show"><slot /><slot name="footer" /></div>'
      },
      Icon: true,
      TotpStepUpDialog: true
    }
  }
})

describe('UserCreateModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    stepUpRun.mockImplementation((operation: () => Promise<unknown>) => operation())
    createUser.mockResolvedValue({})
  })

  it('shows the six-character requirement and blocks shorter passwords', async () => {
    const wrapper = mountModal()
    const passwordInput = wrapper.get('input[type="text"]')

    expect(passwordInput.attributes('minlength')).toBe('6')
    expect(wrapper.text()).toContain('common.passwordMinLength')

    await passwordInput.setValue('abcde')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('common.passwordMinLength')
    expect(stepUpRun).not.toHaveBeenCalled()
    expect(createUser).not.toHaveBeenCalled()
  })

  it('submits a password with at least six characters', async () => {
    const wrapper = mountModal()

    await wrapper.get('input[type="email"]').setValue('new-user@example.com')
    await wrapper.get('input[type="text"]').setValue('secret1')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(createUser).toHaveBeenCalledWith(expect.objectContaining({
      email: 'new-user@example.com',
      password: 'secret1'
    }))
  })
})

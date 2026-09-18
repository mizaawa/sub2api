import { defineComponent, nextTick } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Channel } from '@/api/admin/channels'
import type { AdminGroup } from '@/types'

const {
  listChannels,
  getAllGroups,
  getWebSearchConfig,
  syncPricingModels,
  updateChannel,
} = vi.hoisted(() => ({
  listChannels: vi.fn(),
  getAllGroups: vi.fn(),
  getWebSearchConfig: vi.fn(),
  syncPricingModels: vi.fn(),
  updateChannel: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    channels: {
      list: listChannels,
      create: vi.fn(),
      update: updateChannel,
      remove: vi.fn(),
      syncPricingModels,
    },
    groups: {
      getAll: getAllGroups,
    },
    settings: {
      getWebSearchEmulationConfig: getWebSearchConfig,
    },
    accounts: {
      list: vi.fn(),
      getById: vi.fn(),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn(),
  }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => ({
        'admin.channels.createChannel': 'Create Channel',
        'admin.groups.platforms.anthropic': 'Anthropic',
        'admin.groups.platforms.openai': 'OpenAI',
        'admin.groups.platforms.gemini': 'Gemini',
        'admin.groups.platforms.antigravity': 'Antigravity',
        'admin.groups.platforms.grok': 'Grok',
        'admin.groups.platforms.custom': 'Custom',
      }[key] ?? key),
    }),
  }
})

import ChannelsView from '@/views/admin/ChannelsView.vue'

const groups = [
  { id: 11, name: 'Custom group', platform: 'custom', rate_multiplier: 1, account_count: 1 },
  { id: 12, name: 'Composite group', platform: 'composite', rate_multiplier: 1, account_count: 2 },
  { id: 13, name: 'OpenAI group', platform: 'openai', rate_multiplier: 1, account_count: 3 },
] as AdminGroup[]

const customChannel: Channel = {
  id: 77,
  name: 'Custom pricing',
  description: 'Custom upstream pricing',
  status: 'active',
  billing_model_source: 'channel_mapped',
  restrict_models: true,
  features_config: {},
  group_ids: [11],
  model_pricing: [{
    platform: 'custom',
    models: ['public-model'],
    billing_mode: 'token',
    input_price: 0.000001,
    output_price: 0.000002,
    cache_write_price: null,
    cache_read_price: null,
    image_input_price: null,
    image_output_price: null,
    per_request_price: null,
    intervals: [],
  }],
  model_mapping: { custom: { 'public-model': 'upstream-model' } },
  apply_pricing_to_account_stats: false,
  account_stats_pricing_rules: [],
  created_at: '2026-09-17T00:00:00Z',
  updated_at: '2026-09-17T00:00:00Z',
}

const AppLayoutStub = defineComponent({
  template: '<main><slot /></main>',
})

const TablePageLayoutStub = defineComponent({
  template: '<section><slot name="filters" /><slot name="table" /><slot name="pagination" /></section>',
})

const DataTableStub = defineComponent({
  props: {
    data: { type: Array, default: () => [] },
  },
  template: '<div data-testid="channels-table"><div v-for="row in data" :key="row.id"><slot name="cell-actions" :row="row" /></div><slot v-if="data.length === 0" name="empty" /></div>',
})

const BaseDialogStub = defineComponent({
  props: { show: { type: Boolean, default: false } },
  template: '<div v-if="show" data-testid="channel-dialog"><slot /><slot name="footer" /></div>',
})

const SelectStub = defineComponent({
  props: {
    modelValue: { type: [String, Number], default: '' },
    options: { type: Array, default: () => [] },
  },
  template: '<select><option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option></select>',
})

const ToggleStub = defineComponent({
  props: { modelValue: { type: Boolean, default: false } },
  template: '<input type="checkbox" :checked="modelValue" />',
})

const PricingEntryCardStub = defineComponent({
  props: {
    entry: { type: Object, required: true },
    platform: { type: String, default: '' },
  },
  template: '<div data-testid="pricing-entry" :data-platform="platform">{{ entry.models.join(",") }}</div>',
})

function mountView() {
  return mount(ChannelsView, {
    global: {
      stubs: {
        AppLayout: AppLayoutStub,
        TablePageLayout: TablePageLayoutStub,
        DataTable: DataTableStub,
        Pagination: true,
        BaseDialog: BaseDialogStub,
        ConfirmDialog: true,
        EmptyState: true,
        Select: SelectStub,
        Icon: true,
        PlatformIcon: true,
        Toggle: ToggleStub,
        PricingEntryCard: PricingEntryCardStub,
      },
    },
  })
}

async function openCreateDialog() {
  const wrapper = mountView()
  await flushPromises()
  const createButton = wrapper
    .findAll('button')
    .find((button) => button.text().includes('Create Channel'))
  expect(createButton).toBeDefined()
  await createButton?.trigger('click')
  await flushPromises()
  return wrapper
}

async function enableCustomPlatform(wrapper: ReturnType<typeof mountView>) {
  const customPlatformLabel = wrapper
    .findAll('label')
    .find((label) => label.text().includes('Custom'))
  expect(customPlatformLabel, 'Custom platform option should be visible').toBeDefined()
  if (!customPlatformLabel) throw new Error('Custom platform option is missing')

  await customPlatformLabel.get('input[type="checkbox"]').setValue(true)
  await nextTick()

  const customTab = wrapper
    .findAll('button.channel-tab')
    .find((button) => button.text().includes('Custom'))
  expect(customTab, 'Custom platform tab should be visible after enabling it').toBeDefined()
  if (!customTab) throw new Error('Custom platform tab is missing')

  await customTab.trigger('click')
  await nextTick()
}

describe('ChannelsView Custom platform pricing configuration', () => {
  beforeEach(() => {
    listChannels.mockReset().mockResolvedValue({ items: [], total: 0 })
    getAllGroups.mockReset().mockResolvedValue(groups)
    getWebSearchConfig.mockReset().mockResolvedValue({ enabled: false, providers: [] })
    syncPricingModels.mockReset().mockResolvedValue({ models: [] })
    updateChannel.mockReset().mockResolvedValue(customChannel)
  })

  it('shows Custom as an available platform option and only exposes Custom plus Composite groups', async () => {
    const wrapper = await openCreateDialog()
    await enableCustomPlatform(wrapper)

    expect(wrapper.text()).toContain('Custom group')
    expect(wrapper.text()).toContain('Composite group')
    expect(wrapper.text()).not.toContain('OpenAI group')
  })

  it('does not offer LiteLLM model synchronization for Custom pricing', async () => {
    const wrapper = await openCreateDialog()
    await enableCustomPlatform(wrapper)

    const syncButtons = wrapper
      .findAll('button')
      .filter((button) => button.text().includes('admin.channels.form.syncLatestModels'))
    expect(syncButtons).toHaveLength(0)
    expect(syncPricingModels).not.toHaveBeenCalled()
  })

  it('round-trips Custom groups, model mapping, and pricing when editing', async () => {
    listChannels.mockResolvedValue({ items: [customChannel], total: 1 })
    const wrapper = mountView()
    await flushPromises()

    const editButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('common.edit'))
    expect(editButton).toBeDefined()
    await editButton?.trigger('click')
    await flushPromises()

    const customTab = wrapper
      .findAll('button.channel-tab')
      .find((button) => button.text().includes('Custom'))
    expect(customTab).toBeDefined()
    await customTab?.trigger('click')
    await nextTick()

    expect(wrapper.get('[data-testid="pricing-entry"]').attributes('data-platform')).toBe('custom')
    expect(wrapper.get('[data-testid="pricing-entry"]').text()).toContain('public-model')

    await wrapper.get('#channel-form').trigger('submit')
    await flushPromises()

    expect(updateChannel).toHaveBeenCalledWith(77, expect.objectContaining({
      group_ids: [11],
      model_mapping: { custom: { 'public-model': 'upstream-model' } },
      model_pricing: [expect.objectContaining({
        platform: 'custom',
        models: ['public-model'],
      })],
    }))
  })
})

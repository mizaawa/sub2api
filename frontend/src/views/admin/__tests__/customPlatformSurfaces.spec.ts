import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const srcRoot = resolve(__dirname, '../../..')

function read(relativePath: string) {
  return readFileSync(resolve(srcRoot, relativePath), 'utf8')
}

describe('Custom platform admin surfaces', () => {
  it('exposes Custom in account and group platform selectors', () => {
    const accountFilters = read('components/admin/account/AccountTableFilters.vue')
    const groupsView = read('views/admin/GroupsView.vue')

    expect(accountFilters).toContain("{ value: 'custom', label: 'Custom' }")
    expect(groupsView).toMatch(/platformOptions[\s\S]*?value: \"custom\"[\s\S]*?value: \"composite\"/)
    expect(groupsView).toMatch(/platformFilterOptions[\s\S]*?value: \"custom\"[\s\S]*?value: \"composite\"/)
  })

  it('labels Custom groups and monitor configuration rows consistently', () => {
    const monitorSettings = read('features/channel-monitor-v2/MonitorSettingsPanel.vue')
    const types = read('types/index.ts')

    expect(monitorSettings).toContain("custom: 'Custom'")
    expect(types).toMatch(/AccountPlatform[^\n]+\| 'custom'/)
    expect(types).toMatch(/GroupPlatform[^\n]+\| 'custom'/)
  })
})

import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const viewPath = resolve(dirname(fileURLToPath(import.meta.url)), '../RedeemView.vue')
const viewSource = readFileSync(viewPath, 'utf8')

describe('RedeemView balance tone', () => {
  it('uses the shared balance thresholds for the current balance', () => {
    expect(viewSource).toContain(':class="balanceToneClass(user?.balance)"')
    expect(viewSource).toContain("amount < 0) return 'balance-tone-negative'")
    expect(viewSource).toContain("amount <= 10) return 'balance-tone-low'")
    expect(viewSource).toContain("amount > 100) return 'balance-tone-premium'")
  })
})

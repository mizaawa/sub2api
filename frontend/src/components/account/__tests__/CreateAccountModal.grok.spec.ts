import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = readFileSync(
  resolve(process.cwd(), 'src/components/account/CreateAccountModal.vue'),
  'utf8',
)

describe('CreateAccountModal unified API Key flow', () => {
  it('keeps the official Grok defaults in the shared API Key form', () => {
    expect(source).toContain("? 'https://api.x.ai/v1'")
    expect(source).toContain("? 'xai-...'")
    expect(source).toContain('v-if="form.type === \'apikey\'"')
  })

  it('submits rate sync and omits expiry auto-pause', () => {
    expect(source).toContain('upstream_billing_rate_sync_enabled: upstreamBillingRateSyncEnabled.value')
    expect(source).not.toContain('auto_pause_on_expired:')
  })
})

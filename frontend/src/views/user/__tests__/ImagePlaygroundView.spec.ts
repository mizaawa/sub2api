import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../ImagePlaygroundView.vue')
const componentSource = readFileSync(componentPath, 'utf8')

describe('Image playground launcher', () => {
  it('opens the standalone project instead of embedding a local workbench', () => {
    expect(componentSource).not.toContain('<AppLayout>')
    expect(componentSource).not.toContain('<iframe')
    expect(componentSource).not.toContain('/image-playground-app/')
    expect(componentSource).toContain('https://api.relaycat.top/image-playground/')
    expect(componentSource).toContain("window.location.replace(buildStandaloneUrl")
  })

  it('passes the selected Sub2API async profile to the standalone project', () => {
    expect(componentSource).toContain("provider: 'sb2api-async'")
    expect(componentSource).toContain("baseUrl: buildGatewayUrl('/v1')")
    expect(componentSource).toContain('apiKey: key.key')
    expect(componentSource).toContain("target.searchParams.set('settings'")
  })

  it('keeps the API key creation return path', () => {
    expect(componentSource).toContain("path: '/keys', query: { returnTo: '/image-playground' }")
  })
})

import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../ImagePlaygroundView.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const standaloneIndexPath = resolve(dirname(componentPath), '../../../public/image-playground/index.html')

describe('Image playground launcher', () => {
  it('opens the standalone project instead of embedding a local workbench', () => {
    expect(existsSync(standaloneIndexPath)).toBe(true)
    expect(componentSource).not.toContain('<AppLayout>')
    expect(componentSource).not.toContain('<iframe')
    expect(componentSource).not.toContain('/image-playground-app/')
    expect(componentSource).not.toContain('api.relaycat.top')
    expect(componentSource).toContain("const STANDALONE_IMAGE_PLAYGROUND_PATH = '/image-playground/'")
    expect(componentSource).toContain("const STANDALONE_CONFIG_PREFIX = 'sub2api-image-playground:'")
    expect(componentSource).toContain('window.name = `${STANDALONE_CONFIG_PREFIX}${JSON.stringify(settings)}`')
    expect(componentSource).toContain('window.location.replace(STANDALONE_IMAGE_PLAYGROUND_PATH)')
  })

  it('passes the selected Sub2API async profile to the standalone project', () => {
    expect(componentSource).toContain("provider: 'sb2api-async'")
    expect(componentSource).toContain("baseUrl: buildGatewayUrl('/v1')")
    expect(componentSource).toContain('apiKey: candidate.key')
    expect(componentSource).toContain('function buildStandaloneSettings')
  })

  it('keeps the API key creation return path', () => {
    expect(componentSource).toContain("path: '/keys', query: { returnTo: '/image-playground' }")
  })
})

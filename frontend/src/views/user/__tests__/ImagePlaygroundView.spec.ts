import { existsSync, readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const componentPath = resolve(dirname(fileURLToPath(import.meta.url)), '../ImagePlaygroundView.vue')
const componentSource = readFileSync(componentPath, 'utf8')
const standaloneIndexPath = resolve(dirname(componentPath), '../../../public/image-playground/index.html')
const standaloneIndex = readFileSync(standaloneIndexPath, 'utf8')
const standaloneScriptPath = resolve(
  dirname(standaloneIndexPath),
  'assets',
  standaloneIndex.match(/src="\.\/assets\/(index-[^"]+\.js)"/)?.[1] ?? '',
)
const standaloneScript = readFileSync(standaloneScriptPath, 'utf8')

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
    expect(componentSource).toContain('userGroupsAPI.getAvailable()')
    expect(componentSource).toContain("const IMAGE_PLAYGROUND_API_BASE_URL = 'https://api.zayuapi.com/v1'")
    expect(componentSource).toContain('imagePlaygroundEnabled')
  })

  it('imports window configuration and clears the credential carrier in the standalone build', () => {
    expect(standaloneIndex).toContain('assets/index-')
    expect(standaloneScript).toContain('sub2api-image-playground:')
    expect(standaloneScript).toContain('Failed to import Sub2API window configuration')
    expect(standaloneScript).toContain('window.name=""')
  })

  it('passes the selected Sub2API async profile to the standalone project', () => {
    expect(componentSource).toContain("provider: 'sb2api-async'")
    expect(componentSource).toContain("baseUrl: IMAGE_PLAYGROUND_API_BASE_URL")
    expect(componentSource).toContain('apiKey: candidate.key')
    expect(componentSource).toContain('modelOptions: modelsByKeyId.value[candidate.id]')
    expect(componentSource).toContain('id: `sub2api-key-${candidate.id}`')
    expect(componentSource).toContain('name: candidate.name')
    expect(componentSource).toContain('function buildStandaloneSettings')
  })

  it('keeps the API key creation return path', () => {
    expect(componentSource).toContain("path: '/keys', query: { returnTo: '/image-playground' }")
  })
})

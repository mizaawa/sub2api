import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

const stylePath = resolve(dirname(fileURLToPath(import.meta.url)), '../../../style.css')
const styleSource = readFileSync(stylePath, 'utf8')

describe('shared dialog surface', () => {
  it('clips child backgrounds to the panel and rounds the footer edge', () => {
    const contentBlock = styleSource.match(/\.modal-content\s*\{[\s\S]*?\n {2}\}/)?.[0]
    const footerBlock = styleSource.match(/\.modal-footer\s*\{[\s\S]*?\n {2}\}/)?.[0]

    expect(contentBlock).toContain('@apply overflow-hidden rounded-4xl shadow-card-hover;')
    expect(footerBlock).toContain('@apply rounded-b-4xl;')
  })
})

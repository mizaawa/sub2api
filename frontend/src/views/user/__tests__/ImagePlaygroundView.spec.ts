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
const responsiveOverridePath = resolve(dirname(standaloneIndexPath), 'assets', 'workbench-overrides.css')
const responsiveOverride = readFileSync(responsiveOverridePath, 'utf8')

describe('Image playground launcher', () => {
  it('opens the standalone project instead of embedding a local workbench', () => {
    expect(existsSync(standaloneIndexPath)).toBe(true)
    expect(componentSource).not.toContain('<AppLayout>')
    expect(componentSource).not.toContain('<iframe')
    expect(componentSource).not.toContain('/image-playground-app/')
    expect(componentSource).not.toContain('api.relaycat.top')
    expect(componentSource).toContain("const STANDALONE_IMAGE_PLAYGROUND_PATH = '/image-playground/'")
    expect(componentSource).toContain("const STANDALONE_CONFIG_PREFIX = 'sub2api-image-playground:'")
    expect(componentSource).toContain("const STANDALONE_CONFIG_STORAGE_KEY = 'sub2api-image-playground:bootstrap'")
    expect(componentSource).toContain('sessionStorage.setItem(STANDALONE_CONFIG_STORAGE_KEY')
    expect(componentSource).not.toContain('window.name = `${STANDALONE_CONFIG_PREFIX}${JSON.stringify(settings)}`')
    expect(componentSource).toContain('window.location.replace(STANDALONE_IMAGE_PLAYGROUND_PATH)')
    expect(componentSource).toContain('userGroupsAPI.getAvailable({ signal: context.controller.signal })')
    expect(componentSource).toContain('key.group_id != null')
    expect(componentSource).toContain('key.group?.allow_image_generation === true')
    expect(componentSource).toContain("const IMAGE_PLAYGROUND_API_BASE_URL = '/v1'")
    expect(componentSource).toContain('imagePlaygroundEnabled')
  })

  it('imports window configuration and clears the credential carrier in the standalone build', () => {
    expect(standaloneIndex).toContain('assets/index-')
    expect(standaloneScript).toContain('sub2api-image-playground:')
    expect(standaloneScript).toContain('sub2api-image-playground:bootstrap')
    expect(standaloneScript).toContain('sessionStorage.removeItem')
    expect(standaloneScript).not.toContain('window.name')
    expect(standaloneScript).not.toContain('AgentWorkspace')
    expect(standaloneScript).not.toContain('创建新配置')
    expect(standaloneScript).not.toContain('习惯配置')
  })

  it('retains the managed workbench capabilities while keeping removed surfaces out', () => {
    for (const capability of [
      '输入提示词开始生成图片',
      '参考图',
      '编辑遮罩',
      '下载图片',
      '收藏夹',
      '查看',
      'aria-label":"当前配置"',
      'aria-label":"选择模型"',
      '生成图像',
    ]) {
      expect(standaloneScript).toContain(capability)
    }
    expect(standaloneScript).not.toContain('AgentWorkspace')
    expect(standaloneScript).not.toContain('习惯配置')
    expect(standaloneScript).not.toContain('审核')
  })

  it('binds the standalone cache to user identity without treating token rotation as logout', () => {
    const sessionComponent = standaloneScript.slice(
      standaloneScript.indexOf('function oj(){'),
      standaloneScript.indexOf('const cj="width=device-width'),
    )

    expect(sessionComponent).toContain('const g=wk(),v=_o()')
    expect(sessionComponent).toContain('g.userId===x')
    expect(sessionComponent).toContain('Xl(g.userEmail)===Xl(y)')
    expect(sessionComponent).toContain('x=g?.userId??v.userId')
    expect(sessionComponent).toContain('sessionValid=')
    expect(sessionComponent).toContain('localUserPresent=')
    expect(sessionComponent).toContain('S.key===null||S.key==="auth_user"')
    expect(sessionComponent).toContain('invalidSince=0')
    expect(sessionComponent).toContain('工作台已跳过损坏的本地任务，可继续使用')
    expect(sessionComponent).not.toContain('ij()')
    expect(sessionComponent).not.toContain('auth_token')
    expect(sessionComponent).not.toContain('refresh_token')
    expect(sessionComponent).not.toContain('token_expires_at')
  })

  it('keeps a failed bootstrap available for a later standalone mount', () => {
    expect(standaloneScript).toContain('function clearImagePlaygroundBootstrap')
    expect(standaloneScript).toContain('Promise.allSettled(s.map')
    expect(standaloneScript).toContain('Failed to persist image data; using memory cache')
    expect(standaloneScript).toContain('Failed to persist image task:')
    expect(standaloneScript).toContain('return Yl(s,a)')
    expect(standaloneScript).toContain('async function yb(')
    expect(standaloneScript).toContain('window.sessionStorage.getItem(zx)')
    expect(standaloneScript).not.toContain('window.sessionStorage.getItem(zx),window.sessionStorage.removeItem(zx)')
    expect(standaloneScript).toContain('g&&clearImagePlaygroundBootstrap()')
  })

  it('passes the selected Sub2API async profile to the standalone project', () => {
    expect(componentSource).toContain("provider: 'sb2api-async'")
    expect(componentSource).toContain("baseUrl: IMAGE_PLAYGROUND_API_BASE_URL")
    expect(componentSource).toContain('apiKey: candidate.key')
    expect(componentSource).toContain('modelOptions: modelsByKeyId.value[candidate.id]')
    expect(componentSource).toContain('...modelsByKeyId.value')
    expect(componentSource).toContain("'X-Sub2API-User-Email': context.userEmail")
    expect(componentSource).toContain('id: `sub2api-key-${candidate.id}`')
    expect(componentSource).toContain('name: candidate.name')
    expect(componentSource).toContain('function buildStandaloneSettings')
    expect(componentSource).toContain('Only the selected key is sent to the provider')
    expect(componentSource).not.toContain('Promise.all(imageKeys.value.map')
    expect(standaloneScript).toContain('OpenAI 兼容接口')
    expect(standaloneScript).toContain('https://api.zayuapi.com/v1')
    expect(standaloneScript).toContain('window.location.origin}/v1')
    expect(standaloneScript).toContain('async function v4')
    for (const helper of [
      'function b4(',
      'function w4(',
      'function Cx(',
      'function k4(',
      'function S4(',
      'function Qs(',
      'function j4(',
      'function T4(',
      'async function N4(',
    ]) {
      expect(standaloneScript).toContain(helper)
    }
    expect(standaloneScript).toContain('output_compression==null&&delete')
    expect(standaloneScript).toContain('Failed to persist image task:')
    expect(standaloneScript).toContain('Failed to persist image data; using memory cache')
    expect(standaloneScript).toContain('Failed to read image data:')
    expect(standaloneScript).toContain('aria-label":"当前配置"')
    expect(standaloneScript).toContain('aria-label":"选择模型"')
    expect(standaloneScript).toContain('onClick:()=>mr&&mr()')
    expect(standaloneScript).toContain('N=_r(tn(z.getState().settings)),ee=N.profiles.find')
    expect(standaloneScript).toContain('x=z(j=>j.settings),j=z(j=>j.setSettings),y=z(j=>j.setShowSettings)')
    expect(standaloneScript).toContain('[modelPulling,setModelPulling]=b.useState(!1)')
    expect(standaloneScript).toContain('const modelProfiles=b.useMemo(()=>tn(x),[x])')
    expect(standaloneScript).not.toContain('const modelProfiles=b.useMemo(()=>tn(J),[J])')
    expect(standaloneScript).toContain('onModelRefresh:pullModels')
    expect(standaloneScript).toContain('rawImageUrls:m')
    expect(standaloneScript).not.toContain('moderation')
    expect(standaloneScript).not.toContain('审核')
  })

  it('uses the synchronous zayu image endpoints and keeps mobile controls inside the viewport', () => {
    expect(standaloneScript).toContain('path:"images/generations"')
    expect(standaloneScript).not.toContain('path:"images/generations/async"')
    expect(standaloneScript).toContain('data-image-actions')
    expect(standaloneScript).toContain('data-image-params')
    expect(responsiveOverride).toContain('@media (max-width: 639px)')
    expect(responsiveOverride).toContain('grid-template-columns: repeat(2, minmax(0, 1fr))')
    expect(responsiveOverride).toContain('[data-image-params] > .grid > *')
    expect(responsiveOverride).toContain('width: 100%')
  })

  it('keeps the API key creation return path', () => {
    expect(componentSource).toContain("path: '/keys', query: { returnTo: '/image-playground' }")
  })
})

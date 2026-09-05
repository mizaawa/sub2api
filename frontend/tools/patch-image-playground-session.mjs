import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const bundlePath = resolve(root, 'public/image-playground/assets/index-rvKuKMn_.js')
const source = readFileSync(bundlePath, 'utf8')
// Repair an intermediate bundle created by older versions of this patch. The
// thumbnail loader is asynchronous; an omitted keyword leaves `await` at
// top-level inside the function and prevents the entire workbench from
// parsing. Normalize it before applying the idempotent replacements below.
const normalizedSource = source.replace('}function yb(a){', '}async function yb(a){')
const startMarker = 'function oj(){'
const endMarker = 'const cj="width=device-width'
const start = normalizedSource.indexOf(startMarker)
const end = normalizedSource.indexOf(endMarker, start)

if (start < 0 || end < 0 || normalizedSource.indexOf(startMarker, start + startMarker.length) >= 0) {
  throw new Error('Image playground session component markers are missing or ambiguous')
}

// The Go static server already rejects the standalone HTML and every asset when
// the feature is disabled. At runtime, bind local data to the bootstrap identity
// once, then observe auth_user and token presence. Access/refresh token rotation
// writes several localStorage keys in sequence and must not look like an account change.
// The same-tab bootstrap is authoritative during the handoff from the main
// app; direct standalone visits still require the persisted auth session.
const replacement = `function oj(){const a=z(g=>g.filterFavorite),l=z(g=>g.activeFavoriteCollectionId),[s,i]=b.useState(!1),[d,f]=b.useState(!1),[m,p]=b.useState(null);return AS(),b.useEffect(()=>{if(vy)return;vy=!0;const g=wk(),v=_o(),x=v.userId,y=v.tokenPresent,S=bk(),E=Bx(),A=!!(g&&x&&y&&g.userId===x&&g.userEmail&&Xl(g.userEmail)===v.userEmail),R=!!(!g&&x&&y&&S===x&&E===v.userEmail),_=A||R,U=x===null||!y||S!==x||E!==v.userEmail,O=A&&(S!==x||E!==v.userEmail)||!_&&(!!g||U),D=O?zs():Promise.resolve();_&&x?(yk(x,v.userEmail),p(x),f(!0),nr(!0)):(p(null),f(!1),nr(!1),!A&&O&&Lx()),D.then(async()=>{if(!_)return f(!1),nr(!1),await zs(),!1;nr(!0),await hk();return!0}).then(H=>{if(!H)return;const ee=z.getState();z.setState({appMode:"gallery"}),ee.setSettings(_r(A?g==null?void 0:g.settings:ee.settings)),g&&clearImagePlaygroundBootstrap()}).catch(H=>{console.warn("Failed to initialize image playground:",H),f(!1),nr(!1),zs()}).finally(()=>i(!0))},[]),b.useEffect(()=>{if(!s||m===null||!d)return;let g=!1;const v=()=>{const S=_o();S.userId===m&&S.userEmail===Bx()||g||(g=!0,f(!1),nr(!1),Ab(),zs(),Lx())},x=S=>{(S.key===null||S.key==="auth_user")&&v()};window.addEventListener("storage",x);const y=window.setInterval(v,2e3);return v(),()=>{window.removeEventListener("storage",x),window.clearInterval(y)}},[s,m,d]),b.useEffect(()=>{const g=v=>{var x;(x=v.target)!=null&&x.closest("img")&&v.preventDefault()};return document.addEventListener("dragstart",g),()=>document.removeEventListener("dragstart",g)},[]),s?d?o.jsxs(o.Fragment,{children:[o.jsx(Yk,{}),o.jsx("main",{"data-home-main":!0,"data-drag-select-surface":!0,className:"pb-48",children:o.jsxs("div",{className:"safe-area-x max-w-7xl mx-auto",children:[o.jsx("div",{"data-image-cleanup-banner":!0,role:"status",className:"mb-3 flex items-center border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-900 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-100",children:"云端图片每天 00:00、03:00、06:00、09:00、12:00、15:00、18:00、21:00 自动清理，请及时下载。"}),o.jsx(Xk,{}),a&&!l?o.jsx(rj,{}):o.jsx(Gk,{})]})}),o.jsx(SS,{}),o.jsx(MS,{}),o.jsx(DS,{}),o.jsx(OS,{}),o.jsx(zS,{}),o.jsx(QS,{}),o.jsx(lj,{}),o.jsx(sj,{}),o.jsx(LS,{}),o.jsx(KS,{})]}):o.jsx("main",{className:"flex min-h-screen items-center justify-center p-6 text-center text-sm text-gray-500 dark:text-gray-400",children:"登录身份已失效，请返回主应用重新进入生图工作台。"}):o.jsx("main",{className:"min-h-screen","aria-busy":"true"})}`

// Logout can remove only auth_token before auth_user is cleared. Keep the
// standalone page bound to a live session while still allowing token rotation.
const sessionIdentityReplacement = replacement.replace(
  'const g=wk(),v=_o(),x=v.userId,y=v.tokenPresent,S=bk(),E=Bx(),A=!!(g&&x&&y&&g.userId===x&&g.userEmail&&Xl(g.userEmail)===v.userEmail),R=!!(!g&&x&&y&&S===x&&E===v.userEmail),_=A||R,U=x===null||!y||S!==x||E!==v.userEmail,O=A&&(S!==x||E!==v.userEmail)||!_&&(!!g||U),D=O?zs():Promise.resolve();',
  'const g=wk(),v=_o(),x=g?.userId??v.userId,y=g?.userEmail??v.userEmail,S=bk(),E=Bx(),A=!!(g&&g.userId===x&&g.userEmail&&Xl(g.userEmail)===Xl(y)),R=!!(!g&&x&&v.tokenPresent&&S===x&&E===v.userEmail),_=A||R,U=!x||!y||(!g&&(S!==x||E!==v.userEmail)),O=!_&&U,D=O?zs():Promise.resolve();',
)
  .replace('_&&x?(yk(x,v.userEmail)', '_&&x?(yk(x,y)')
  .replace(
    'const S=_o();S.userId===m&&S.userEmail===Bx()||g||(g=!0',
    'const S=_o(),ownerEmail=Bx(),localUserPresent=!!window.localStorage.getItem("auth_user"),sessionValid=S.userId===m&&(!localUserPresent||S.tokenPresent)&&(!S.userEmail||!ownerEmail||S.userEmail===ownerEmail)||(!localUserPresent&&bk()===m&&ownerEmail===Bx());sessionValid||g||(g=!0',
  )

const replacementWithVersionModal = sessionIdentityReplacement
  .replace('o.jsx(lj,{}),o.jsx(LS,{})', 'o.jsx(lj,{}),o.jsx(sj,{}),o.jsx(LS,{})')
  .replace('o.jsx(KS,{})]}):o.jsx("main"', 'o.jsx(KS,{}),o.jsx(VS,{})]}):o.jsx("main"')
const auditedReplacement = replacementWithVersionModal.replace(
  'S.userId===m&&S.userEmail===Bx()',
  'S.tokenPresent&&S.userId===m&&S.userEmail===Bx()',
)

let patched = normalizedSource.slice(0, start) + auditedReplacement + normalizedSource.slice(end)

// Older patch runs may have spliced an async function's prefix into the
// replacement boundary, leaving `function Qy/yb` with an illegal `await`.
// Repair those generated artifacts before applying idempotent replacements.
for (const [broken, fixed] of [
  ['function Qy(a,l="upload"){return(await $y(a,l)).id}', 'async function Qy(a,l="upload"){return(await $y(a,l)).id}'],
  ['function yb(a){const l=_0(a);', 'async function yb(a){const l=_0(a);'],
]) {
  if (patched.includes(broken)) replaceOnce('repair async image helper prefix', broken, fixed)
}
patched = patched.replace(/(?:async\s+){2,}function fr\(/g, 'async function fr(')
patched = patched.replace(/(?:async\s+){2,}function Qy\(/g, 'async function Qy(')
patched = patched.replace(/(?:async\s+){2,}function Tn\(/g, 'async function Tn(')
patched = patched.replace(/(?:async\s+){2,}function yb\(/g, 'async function yb(')

// Local IndexedDB is auxiliary state, not a prerequisite for sending an
// image request. A corrupted/locked database used to reject the first task's
// persistence promise before Rb() could reach the provider, leaving every
// later click stuck in "生成中". Keep persistence best-effort so the network
// request and its error handling remain usable.
const resilientTaskPersistence = 'async function fr(a,l=Vn()){if(!He(l))return"";let i="";try{i=await O5(ck(a))}catch(d){console.warn("Failed to persist image task:",d);return""}if(!He(l)&&!z.getState().tasks.some(f=>f.id===a.id))try{return await v0(a.id)}catch(d){console.warn("Failed to remove stale image task:",d)}return i}'
replaceFunctionOnce(
  'best-effort image task persistence',
  'async function fr(',
  'function uk(',
  resilientTaskPersistence,
  'async function fr(',
)

// Generated and uploaded image bytes are also auxiliary persistence. Keep a
// bounded in-memory copy when IndexedDB is unavailable so a successful
// provider response can still be rendered in the current session.
const resilientImageStorage = 'async function $y(a,l="upload"){const s=await B5(a);let i=null;try{i=await ai(s)}catch(d){console.warn("Failed to read image cache:",d)}if(i){i.dataUrl&&Yl(s,i.dataUrl);return{id:s,width:i.width,height:i.height}}const d=await Vy(a);try{await Py({id:s,dataUrl:a,createdAt:Date.now(),source:l,width:d.width,height:d.height}),d.thumbnailDataUrl&&await x0({id:s,thumbnailDataUrl:d.thumbnailDataUrl,width:d.width,height:d.height,thumbnailVersion:ni})}catch(f){console.warn("Failed to persist image data; using memory cache:",f)}return Yl(s,a),{id:s,width:d.width,height:d.height}}async '
replaceFunctionOnce(
  'best-effort image byte persistence',
  'async function $y(',
  'function Qy(',
  resilientImageStorage,
  'Failed to persist image data; using memory cache',
)

const resilientImageRead = 'async function Tn(a){const l=Zs(a);if(l)return l;try{const s=await ai(a);if(s)return s.dataUrl&&Yl(a,s.dataUrl),s.dataUrl}catch(i){console.warn("Failed to read image data:",i)}return void 0}async '
replaceFunctionOnce(
  'best-effort image byte reads',
  'async function Tn(',
  'function yb(',
  resilientImageRead,
  'Failed to read image data:',
)

// Restoring local workbench state is best-effort. A stale task or an image
// record from an older IndexedDB schema must not turn an otherwise valid
// bootstrap into the misleading "login expired" screen. Keep the identity
// session alive and skip only the damaged record.
const resilientTaskRestore = `async function hk(){const a=Vn();if(!He(a))return;let l=[];try{l=await I5()}catch(v){console.warn("Failed to load local image tasks:",v)}if(!He(a))return;const s=l.filter(Dx);s.length&&await Promise.allSettled(s.map(async v=>{try{await v0(v.id)}catch(x){console.warn("Failed to remove stale image task:",x)}}));if(!He(a))return;let i=[],d=[];try{const restored=ek(l.filter(v=>!Dx(v)),Date.now());i=restored.tasks||[],d=restored.interruptedTasks||[]}catch(v){console.warn("Failed to restore local image tasks:",v)}d.length&&await Promise.allSettled(d.map(async v=>{try{await fr(v)}catch(x){console.warn("Failed to persist interrupted image task:",x)}}));if(!He(a))return;const f=z.getState();let m;try{m=Z4(i,f.favoriteCollections,f.defaultFavoriteCollectionId)}catch(v){console.warn("Failed to normalize local image tasks:",v),m={tasks:[],collections:f.favoriteCollections,defaultFavoriteCollectionId:f.defaultFavoriteCollectionId,changed:!1}}z.setState({settings:tn(f.settings),tasks:m.tasks||[],favoriteCollections:m.collections||f.favoriteCollections,defaultFavoriteCollectionId:m.defaultFavoriteCollectionId,appMode:"gallery"});if(m.changed)await Promise.allSettled((m.tasks||[]).map(async v=>{try{await fr(v)}catch(x){console.warn("Failed to persist normalized image task:",x)}}));if(!He(a))return;const p=[];for(const v of z.getState().inputImages){if(!He(a))return;if(v.dataUrl){p.push(v),Yl(v.id,v.dataUrl);continue}let x=null;try{x=await Tn(v.id)}catch(S){console.warn("Failed to restore input image:",S)}if(!He(a))return;x&&p.push({...v,dataUrl:x})}if(!He(a))return;p.length!==z.getState().inputImages.length&&z.setState({inputImages:p});const g=new Set;z.getState().inputImages.forEach(v=>g.add(v.id)),(m.tasks||[]).forEach(v=>{var x,y;v.inputImageIds.forEach(S=>g.add(S)),v.outputImages.forEach(S=>g.add(S)),v.maskImageId&&g.add(v.maskImageId),(x=v.transparentOriginalImages)==null||x.forEach(S=>S&&g.add(S)),(y=v.streamPartialImageIds)==null||y.forEach(S=>g.add(S))});try{for(const v of await z5()){if(!He(a))return;g.has(v)||await Ky(v)}}catch(v){console.warn("Failed to clean orphaned image data:",v)}}`
replaceFunctionOnce(
  'resilient local task restore',
  'async function hk(){',
  'async function Ob',
  resilientTaskRestore,
  'Promise.allSettled(s.map',
)

// Initialization failures after identity validation are recoverable. Keep the
// gallery mounted so the user can retry generation and inspect a useful toast;
// reserve the signed-out screen for the explicit identity gate above.
const initializationCatch = '.catch(H=>{console.warn("Failed to initialize image playground:",H),f(!1),nr(!1),zs()})'
const initializationCatchFixed = '.catch(H=>{console.warn("Failed to initialize image playground:",H);if(!_){f(!1),nr(!1),zs();return}const ee=z.getState();A&&g&&ee.setSettings(_r(g.settings)),z.setState({appMode:"gallery"}),f(!0),nr(!0),ee.showToast&&ee.showToast("工作台已跳过损坏的本地任务，可继续使用","warning"),g&&clearImagePlaygroundBootstrap()})'
if (patched.includes(initializationCatch)) replaceOnce('recoverable initialization error', initializationCatch, initializationCatchFixed)
else if (!patched.includes(initializationCatchFixed)) throw new Error('recoverable initialization error marker is missing')

function replaceOnce(label, from, to) {
  const occurrences = patched.split(from).length - 1
  if (occurrences !== 1) {
    throw new Error(`${label}: expected one occurrence, found ${occurrences}`)
  }
  patched = patched.replace(from, to)
}

function removeOnce(label, from) {
  const occurrences = patched.split(from).length - 1
  if (occurrences === 0) return
  if (occurrences !== 1) {
    throw new Error(`${label}: expected one occurrence, found ${occurrences}`)
  }
  patched = patched.replace(from, '')
}

function removeCount(label, from, expected) {
  const occurrences = patched.split(from).length - 1
  if (occurrences === 0) return
  if (occurrences !== expected) {
    throw new Error(`${label}: expected ${expected} occurrences, found ${occurrences}`)
  }
  patched = patched.split(from).join('')
}

function replaceFunctionOnce(label, startMarker, endMarker, replacement, alreadyMarker) {
  const start = patched.indexOf(startMarker)
  const end = patched.indexOf(endMarker, start)
  if (start < 0 || end < 0) throw new Error(`${label}: function markers are missing`)
  const current = patched.slice(start, end)
  if (current.includes(alreadyMarker)) return
  patched = patched.slice(0, start) + replacement + patched.slice(end)
}

// Browser auth refreshes update several storage keys in sequence. Debounce a
// missing token/parse result, but react immediately when a different user is
// detected. This prevents a transient refresh gap from ejecting the gallery.
const sessionWatch = 'b.useEffect(()=>{if(!s||m===null||!d)return;let g=!1;const v=()=>{const S=_o(),ownerEmail=Bx(),localUserPresent=!!window.localStorage.getItem("auth_user"),sessionValid=S.userId===m&&(!localUserPresent||S.tokenPresent)&&(!S.userEmail||!ownerEmail||S.userEmail===ownerEmail)||(!localUserPresent&&bk()===m&&ownerEmail===Bx());sessionValid||g||(g=!0,f(!1),nr(!1),Ab(),zs(),Lx())},x=S=>{(S.key===null||S.key==="auth_user")&&v()};window.addEventListener("storage",x);const y=window.setInterval(v,2e3);return v(),()=>{window.removeEventListener("storage",x),window.clearInterval(y)}},[s,m,d])'
const sessionWatchFixed = 'b.useEffect(()=>{if(!s||m===null||!d)return;let g=!1,invalidSince=0;const v=()=>{const S=_o(),ownerEmail=Bx(),localUserPresent=!!window.localStorage.getItem("auth_user"),identityMismatch=S.userId!==null&&S.userId!==m||!!S.userEmail&&!!ownerEmail&&S.userEmail!==ownerEmail,sessionValid=!identityMismatch&&(S.userId===m&&(!localUserPresent||S.tokenPresent)||!localUserPresent&&S.tokenPresent&&bk()===m&&ownerEmail===Bx());if(sessionValid){invalidSince=0;return}if(!identityMismatch){if(!invalidSince)invalidSince=Date.now();if(Date.now()-invalidSince<8e3)return}g||(g=!0,f(!1),nr(!1),Ab(),zs(),Lx())},x=S=>{(S.key===null||S.key==="auth_user")&&v()};window.addEventListener("storage",x);const y=window.setInterval(v,2e3);return v(),()=>{window.removeEventListener("storage",x),window.clearInterval(y)}},[s,m,d])'
if (patched.includes(sessionWatch)) replaceOnce('debounced standalone session watch', sessionWatch, sessionWatchFixed)
else if (!patched.includes(sessionWatchFixed)) throw new Error('debounced standalone session watch marker is missing')

// Keep provider requests same-origin so browser preflight policy cannot block
// the Authorization and owner-email headers. Deployments on the zayu hostname
// still resolve this to the same zayu `/v1` endpoint.
const remoteBaseUrl = 'Yo="https://api.zayuapi.com/v1"'
const sameOriginBaseUrl = 'Yo=typeof window>"u"?"https://api.zayuapi.com/v1":`${window.location.origin}/v1`'
if (patched.includes(remoteBaseUrl)) replaceOnce('same-origin image API base URL', remoteBaseUrl, sameOriginBaseUrl)
const staleSameOriginBaseUrl = 'Yo=typeof window<"u"?"https://api.zayuapi.com/v1":`${window.location.origin}/v1`'
if (patched.includes(staleSameOriginBaseUrl)) replaceOnce('same-origin image API base URL condition', staleSameOriginBaseUrl, sameOriginBaseUrl)

// The vendored build references v4 from the task runner but the function was
// omitted by the upstream production bundle. Dispatch managed image profiles
// through the HTTP Images API and retain the Responses path for other modes.
const imageDispatcher = 'async function v4(a,l,s){if(s&&s.submit)return a.params.n>1?M4(a,l,s,a.params.n):vb(a,l,s);return A4(a,l)}'
if (!patched.includes('async function v4(')) {
  replaceOnce('image request dispatcher', 'const lk={', imageDispatcher + 'const lk={')
}

// The managed bundle was produced from a gallery-only build where the custom
// provider helpers were tree-shaken even though the request functions still
// reference them. Restore the complete helper set before the provider
// definition so model loading and image generation cannot fail with a runtime
// ReferenceError. Keep this in the patch script because the vendored bundle is
// regenerated from upstream periodically.
const customProviderHelpers = `function b4(a,l){return new Promise((s,i)=>{if(l.aborted){i(new DOMException("Aborted","AbortError"));return}const d=setTimeout(s,a);l.addEventListener("abort",()=>{clearTimeout(d),i(new DOMException("Aborted","AbortError"))},{once:!0})})}function w4(a,l){const s=zr(a,l.statusPath),i=typeof s=="string"?s:String(s??"");return l.successValues.includes(i)?"success":l.failureValues.includes(i)?"failure":"pending"}function Cx(a){if(typeof DOMException<"u"&&a instanceof DOMException&&a.name==="AbortError")return!0;const l=a instanceof Error?a.message:String(a);return/abort|network|failed to fetch|fetch failed|load failed|timeout|连接|断开|中断/i.test(l)}function k4(a){return a===408||a===429||a>=500}function S4(a,l){return a.replace(/\\{task_id\\}/g,encodeURIComponent(l)).replace(/\\{taskId\\}/g,encodeURIComponent(l))}function Qs(a,l){if(typeof a=="string"&&a.startsWith("$"))return zr(l,a.slice(1));if(Array.isArray(a))return a.map(s=>Qs(s,l)).filter(s=>s!=null);if(a&&typeof a=="object"){const s=Object.entries(a).map(([i,d])=>[i,Qs(d,l)]).filter(([,i])=>i!=null&&(!Array.isArray(i)||i.length>0));return Object.fromEntries(s)}return a}function j4(a,l){const s=l.codexCli&&!a.skipCodexCliSizePrompt?R0(a.prompt,a.params.size):a.prompt,i=l.codexCli&&!a.settings.allowPromptRewrite?\`\${T0}\\n\${s}\`:s,d={...a.params,...l.codexCli?{size:void 0,quality:void 0}:{},...a.nativeTransparentBackground?{background:"transparent"}:{}};return{profile:l,prompt:i,params:d,inputImages:{dataUrls:a.inputImageDataUrls.length?a.inputImageDataUrls:void 0,count:a.inputImageDataUrls.length},mask:{dataUrl:a.maskDataUrl}}}function T4(a,l){if(!a)return;const s=Object.entries(a).map(([i,d])=>[i,Qs(d,l)]).filter(([,i])=>i!=null&&String(i)!=="").map(([i,d])=>[i,String(d)]);return s.length?Object.fromEntries(s):void 0}async function N4(a,l,s){var v,x,y;const i=new FormData,d=Qs(a.body??{},s);if(d&&typeof d=="object"&&!Array.isArray(d)){for(const[S,E]of Object.entries(d))if(E!=null)if(Array.isArray(E))for(const A of E)i.append(S,String(A));else i.append(S,String(E))}const f=(v=a.files)==null?void 0:v.some(S=>S.source==="inputImages"),m=(x=a.files)==null?void 0:x.some(S=>S.source==="mask"),p=[];if(f)for(let S=0;S<l.inputImageDataUrls.length;S++){const E=l.inputImageDataUrls[S],A=l.maskDataUrl&&S===0?await b0(E):await y0(E);p.push(A)}const g=m&&l.maskDataUrl?await Zy(l.maskDataUrl):null;l.maskDataUrl&&(f||m)&&(Bl("遮罩主图文件",((y=p[0])==null?void 0:y.size)??0),Bl("遮罩文件",(g==null?void 0:g.size)??0)),Go(p.reduce((S,E)=>S+E.size,0)+((g==null?void 0:g.size)??0));for(const S of a.files??[])if(S.source==="inputImages")for(let E=0;E<p.length;E++){const A=p[E],R=A.type.split("/")[1]||"png";i.append(S.field,A,\`input-\${E+1}.\${R}\`)}else S.source==="mask"&&g&&i.append(S.field,g,"mask.png");return i}`
if (!patched.includes('function j4(')) {
  replaceOnce('custom provider request helpers', 'const lk={', customProviderHelpers + 'const lk={')
}

// zayu rejects an explicit null output_compression field. Omit optional nulls
// from the resolved JSON body while preserving zero and positive values.
const jsonBodyBuild = 'const A=Qs(a.body??{},p);s.responseFormatB64Json&&A&&typeof A=="object"&&!Array.isArray(A)&&(A.response_format="b64_json"),S=JSON.stringify(A)'
const jsonBodyBuildFixed = 'const A=Qs(a.body??{},p);A&&typeof A=="object"&&!Array.isArray(A)&&A.output_compression==null&&delete A.output_compression,s.responseFormatB64Json&&A&&typeof A=="object"&&!Array.isArray(A)&&(A.response_format="b64_json"),S=JSON.stringify(A)'
if (patched.includes(jsonBodyBuild)) replaceOnce('omit null image compression', jsonBodyBuild, jsonBodyBuildFixed)

// Keep the bootstrap carrier until the standalone app has finished restoring
// its settings. A malformed or interrupted first mount must be retryable.
const bootstrapReaderReplacement = 'function wk(){if(typeof window>"u")return null;let a=null;try{a=window.sessionStorage.getItem(zx)}catch{return null}if(!a||!a.startsWith(_x))return null;try{const l=JSON.parse(a.slice(_x.length));if(!l||typeof l!=="object"||Array.isArray(l))return null;const s=F0(l.userId),i=l.settings;if(!s||!i||typeof i!=="object"||Array.isArray(i))return null;const d=Xl(l.userEmail);return{settings:i,userId:s,...d?{userEmail:d}:{}}}catch{return null}}function clearImagePlaygroundBootstrap(){try{window.sessionStorage.removeItem(zx)}catch{}}'
replaceFunctionOnce(
  'durable standalone bootstrap carrier',
  'function wk(){',
  'const Yb=',
  bootstrapReaderReplacement,
  'function clearImagePlaygroundBootstrap',
)

// The scroll lock helper must receive the dialog container so wheel and touch
// gestures can reach the inner overflow-y-auto settings panel.
const settingsScrollLock = 'Aa(a,()=>i(!1)),aa(a),!a)return'
const settingsScrollLockFixed = 'Aa(a,()=>i(!1)),aa(a,y),!a)return'
if (patched.includes(settingsScrollLock)) replaceOnce('settings scroll lock container', settingsScrollLock, settingsScrollLockFixed)
else if (!patched.includes(settingsScrollLockFixed)) throw new Error('settings scroll lock container: marker is missing')
const settingsDialog = 'o.jsxs("section",{role:"dialog","aria-modal":"true","aria-label":"设置",className:'
const settingsDialogFixed = 'o.jsxs("section",{ref:y,role:"dialog","aria-modal":"true","aria-label":"设置",className:'
if (patched.includes(settingsDialog)) replaceOnce('settings dialog scroll ref', settingsDialog, settingsDialogFixed)
else if (!patched.includes(settingsDialogFixed)) throw new Error('settings dialog scroll ref: marker is missing')

// Always refresh the selected profile's image models when the main model
// selector receives focus. The launcher may already have a default list, but
// that list is not authoritative after provider-side model changes.
const modelRefreshGuard = 'if(modelPulling)return;const M=_r(tn(z.getState().settings)),P=M.profiles.find(le=>le.id===M.activeProfileId);if(!P||!P.apiKey.trim()||(P.modelOptions??[]).length>1)return;'
const modelRefreshGuardFixed = 'if(modelPulling)return;const M=_r(tn(z.getState().settings)),P=M.profiles.find(le=>le.id===M.activeProfileId);if(!P||!P.apiKey.trim())return;'
if (patched.includes(modelRefreshGuard)) replaceOnce('main model selector refresh guard', modelRefreshGuard, modelRefreshGuardFixed)

// The public zayu endpoint exposes the synchronous Images API. The previous
// async template requires optional object storage and returns 404 when that
// feature is disabled, making every standalone generation fail before polling.
const asyncGenerationPath = 'path:"images/generations/async"'
const syncGenerationPath = 'path:"images/generations"'
if (patched.includes(asyncGenerationPath)) replaceOnce('sync image generation endpoint', asyncGenerationPath, syncGenerationPath)
const asyncEditPath = 'path:"images/edits/async"'
const syncEditPath = 'path:"images/edits"'
if (patched.includes(asyncEditPath)) replaceOnce('sync image edit endpoint', asyncEditPath, syncEditPath)
const taskIdGeneration = 'n:"$params.n"},taskIdPath:"task_id",result:'
const syncGenerationResult = 'n:"$params.n"},result:'
if (patched.includes(taskIdGeneration)) replaceOnce('sync image generation response', taskIdGeneration, syncGenerationResult)
const taskIdEdit = 'field:"mask",source:"mask"}],taskIdPath:"task_id",result:'
const syncEditResult = 'field:"mask",source:"mask"}],result:'
if (patched.includes(taskIdEdit)) replaceOnce('sync image edit response', taskIdEdit, syncEditResult)

// The managed build never sends moderation controls. Keep these replacements
// explicit so a refreshed vendor bundle fails loudly instead of silently
// retaining a request field or stale UI.
removeOnce('default params moderation', ',moderation:"auto"')
removeOnce('normalized params moderation', ',moderation:a.moderation==="low"?"low":"auto"')
removeOnce('actual params moderation', ',(l.moderation==="auto"||l.moderation==="low")&&(s.moderation=l.moderation)')
removeOnce('responses moderation', ',moderation:a.moderation')
removeOnce('stream result moderation', ',moderation:en(a,"moderation")')
removeOnce('multipart moderation', ',O.append("moderation",i.moderation)')
removeCount('sync moderation', ',moderation:i.moderation', 1)
removeCount('async moderation', ',moderation:"$params.moderation"', 2)

const toolbarStart = patched.indexOf('function kS(')
const toolbarEnd = patched.indexOf('const Ys=16', toolbarStart)
if (toolbarStart < 0 || toolbarEnd < 0) {
  throw new Error('toolbar component markers are missing')
}
let toolbar = patched.slice(toolbarStart, toolbarEnd)
const toolbarSignature = 'activeProfile:i,isFalProvider:d,isFalTextToImage:f,displaySize:m,qualityOptions:p,selectClass:g,transparentOutputAvailable:v,showTransparentOutputControl:x,transparentOutputEnabled:y,transparentOutputHint:S,onTransparentOutputMenuOpenChange:E,compressionHint:A,compressionDisabled:R,outputCompressionInput:_,setOutputCompressionInput:U,commitOutputCompression:O,moderationHint:D,moderationDisabled:N,outputImageLimit:H'
if (toolbar.includes(toolbarSignature)) {
  toolbar = toolbar.replace(toolbarSignature, 'activeProfile:i,profileOptions:po,onProfileChange:pc,onModelChange:mc,onModelRefresh:mr,isFalProvider:d,isFalTextToImage:f,displaySize:m,qualityOptions:p,selectClass:g,transparentOutputAvailable:v,showTransparentOutputControl:x,transparentOutputEnabled:y,transparentOutputHint:S,onTransparentOutputMenuOpenChange:E,compressionHint:A,compressionDisabled:R,outputCompressionInput:_,setOutputCompressionInput:U,commitOutputCompression:O,outputImageLimit:H')
} else if (!toolbar.includes('profileOptions:po,onProfileChange:pc,onModelChange:mc,onModelRefresh:mr')) {
  throw new Error('toolbar signature marker is missing')
}
const moderationControl = 'o.jsxs("label",{className:"relative flex flex-col gap-0.5",onMouseEnter:D.show,onMouseLeave:D.hide,onTouchStart:D.startTouch,onTouchEnd:D.clearTimer,onTouchCancel:D.hide,onClick:D.show,children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"审核"}),o.jsx(Gs,{value:N?"auto":l.moderation,onChange:V=>{N||s({moderation:V})},options:[{label:"auto",value:"auto"},{label:"low",value:"low"}],disabled:N,showValueTooltips:!1,className:N?"px-3 py-1.5 rounded-xl border border-gray-200/60 dark:border-white/[0.08] bg-gray-100/50 dark:bg-white/[0.05] opacity-50 cursor-not-allowed text-xs transition-all duration-200 shadow-sm":g}),o.jsx(rr,{visible:N&&D.visible,text:"fal.ai 不支持审核参数"})]})'
const modelControl = 'o.jsxs("div",{className:"relative flex min-w-0 flex-col gap-0.5",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"配置 / 模型"}),o.jsxs("div",{className:"grid min-w-0 grid-cols-1 gap-1 sm:grid-cols-2",children:[o.jsx("select",{value:i.id,onChange:V=>pc&&pc(V.target.value),className:`${g} min-w-0 w-full`,"aria-label":"当前配置",children:(po??[]).map(V=>o.jsx("option",{value:V.id,children:V.name},V.id))}),o.jsx("select",{value:i.model,onFocus:()=>mr&&mr(),onClick:()=>mr&&mr(),onChange:V=>mc&&mc(V.target.value),className:`${g} min-w-0 w-full`,"aria-label":"选择模型",children:[...new Set((i.modelOptions??[]).concat(i.model||[]))].filter(V=>V).map(V=>o.jsx("option",{value:V,children:V},V))})]})]})'
const modelControlWithoutClick = modelControl.replace('onClick:()=>mr&&mr(),', '')
const legacyModelControl = 'o.jsxs("div",{className:"relative flex flex-col gap-0.5",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"配置 / 模型"}),o.jsxs("div",{className:"flex gap-1",children:[o.jsx("select",{value:i.id,onChange:V=>pc&&pc(V.target.value),className:g,"aria-label":"当前配置",children:(po??[]).map(V=>o.jsx("option",{value:V.id,children:V.name},V.id))}),o.jsx("select",{value:i.model,onFocus:()=>mr&&mr(),onClick:()=>mr&&mr(),onChange:V=>mc&&mc(V.target.value),className:g,"aria-label":"选择模型",children:[...new Set((i.modelOptions??[]).concat(i.model||[]))].filter(V=>V).map(V=>o.jsx("option",{value:V,children:V},V))})]})]})'
if (toolbar.includes(legacyModelControl)) toolbar = toolbar.replace(legacyModelControl, modelControl)
else if (toolbar.includes(modelControlWithoutClick)) toolbar = toolbar.replace(modelControlWithoutClick, modelControl)
if (toolbar.includes(moderationControl)) toolbar = toolbar.replace(moderationControl, modelControl)
else if (!toolbar.includes(modelControl)) throw new Error('toolbar model control marker is missing')
patched = patched.slice(0, toolbarStart) + toolbar + patched.slice(toolbarEnd)

const modelFocusOnly = 'onFocus:()=>mr&&mr(),onChange:V=>mc&&mc(V.target.value)'
const modelFocusAndClick = 'onFocus:()=>mr&&mr(),onClick:()=>mr&&mr(),onChange:V=>mc&&mc(V.target.value)'
if (patched.includes(modelFocusOnly)) replaceOnce('model selector click refresh', modelFocusOnly, modelFocusAndClick)

// Add stable hooks for the responsive action and parameter rows. The vendor
// bundle is minified, so semantic data attributes keep CSS overrides readable.
const actionRow = 'o.jsxs("div",{className:"mt-3 flex items-end gap-2",children:'
const actionRowFixed = 'o.jsxs("div",{"data-image-actions":!0,className:"mt-3 flex items-end gap-2",children:'
if (patched.includes(actionRow)) replaceOnce('responsive action row hook', actionRow, actionRowFixed)
const desktopParams = 'o.jsx("div",{className:"hidden min-w-0 flex-1 sm:block",children:'
const desktopParamsFixed = 'o.jsx("div",{"data-image-params":!0,className:"hidden min-w-0 flex-1 sm:block",children:'
if (patched.includes(desktopParams)) replaceOnce('desktop parameter row hook', desktopParams, desktopParamsFixed)
const mobileParams = 'o.jsx("div",{className:"mt-2 sm:hidden",children:'
const mobileParamsFixed = 'o.jsx("div",{"data-image-params":!0,className:"mt-2 sm:hidden",children:'
if (patched.includes(mobileParams)) replaceOnce('mobile parameter row hook', mobileParams, mobileParamsFixed)

// Remove the old moderation details block from the task inspector as well.
const detailsModeration = ',o.jsxs("div",{className:"bg-gray-50 dark:bg-white/[0.03] rounded-lg px-3 py-2 min-w-0 overflow-hidden",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500",children:"审核"}),o.jsx("br",{}),o.jsx("div",{className:"mt-0.5 overflow-x-auto hide-scrollbar whitespace-nowrap mask-edge-r pr-2",children:o.jsx(Bs,{task:w,paramKey:"moderation",className:"font-medium",actualParams:Qe})})]})'
removeOnce('task moderation details', detailsModeration)
removeOnce('toolbar moderation props', ',moderationHint:de,moderationDisabled:!1')

// A provider may return a valid image URL without CORS headers. In that case
// an image element can still render the URL, while fetch/blob conversion cannot.
// Keep the URL as the stored image source instead of turning a successful task
// into a misleading "load failed" error.
const syncImageFallback = 'async function pb(a,l,s,i){const d=[],f=(l.imageUrlPaths??[]).flatMap(p=>Nx(a,p).filter(g=>zl(g)||Po(g))),m=f.filter(zl);for(const p of l.b64JsonPaths??[])for(const g of Nx(a,p))typeof g==="string"&&g.trim()&&d.push(Ll(g,s));for(const p of f)try{d.push(await E0(p,s,i))}catch(g){zl(p)?d.push(p):g instanceof Error&&(g.rawImageUrls=m);if(!zl(p))throw g}if(!d.length){const p=new Error("接口没有返回可识别的图片数据，请查看接口实际返回的数据结构，并根据 API 文档调整「自定义服务商」配置中的结果提取路径。" );throw p.rawResponsePayload=JSON.stringify(a,null,2),p}return{images:d,...m.length?{rawImageUrls:m}:{}}}'
if (patched.includes('async function pb(')) {
  replaceFunctionOnce('sync image result fallback', 'async function pb(', 'async function C4', syncImageFallback, 'zl(p)?d.push(p):g instanceof Error&&(g.rawImageUrls=m)')
} else {
  const c4 = patched.indexOf('async function C4')
  if (c4 < 0) throw new Error('sync image result fallback: C4 marker is missing')
  patched = patched.slice(0, c4) + syncImageFallback + patched.slice(c4)
}
replaceFunctionOnce(
  'stream image result fallback',
  'async function hb(',
  'async function pb',
  'async function hb(a,l,s){const i=a.data;if(!Array.isArray(i)||!i.length){const g=new Error("接口没有返回图片数据，请查看接口实际返回的数据结构，并根据 API 文档调整「自定义服务商」配置中的结果提取路径。" );throw g.rawResponsePayload=JSON.stringify(a,null,2),g}const d=[],f=i.map(g=>g.url).filter(zl),m=[];for(const g of i){const v=g.b64_json;if(v){d.push(Ll(v,l)),m.push(typeof g.revised_prompt==="string"?g.revised_prompt:void 0);continue}if(zl(g.url)||Po(g.url))try{d.push(await E0(g.url,l,s)),m.push(typeof g.revised_prompt==="string"?g.revised_prompt:void 0)}catch(x){if(zl(g.url)){d.push(g.url),m.push(typeof g.revised_prompt==="string"?g.revised_prompt:void 0);continue}throw x}}if(!d.length){const p=new Error("接口没有返回可识别的图片数据，请查看接口实际返回的数据结构，并根据 API 文档调整「自定义服务商」配置中的结果提取路径。" );throw p.rawResponsePayload=JSON.stringify(a,null,2),p}return{images:d,...m.length?{rawImageUrls:f}:{}}}',
  'zl(g.url)){d.push(g.url),m.push',
)

// Read the latest Zustand settings for profile/model changes. This keeps the
// launcher bootstrap, toolbar, and settings dialog on one browser-local object.
if (patched.includes('U=D=>{var H;const N=A.profiles.find(ee=>ee.id===D);')) {
  replaceOnce(
    'settings profile switch',
    'U=D=>{var H;const N=A.profiles.find(ee=>ee.id===D);N&&(S.current+=1,(H=y.current)==null||H.abort(),y.current=null,x(!1),g(!1),s(_r({...A,activeProfileId:N.id})))}',
    'U=D=>{var H;const N=_r(tn(z.getState().settings)),ee=N.profiles.find($=>$.id===D);ee&&(S.current+=1,(H=y.current)==null||H.abort(),y.current=null,x(!1),g(!1),s(_r({...N,activeProfileId:ee.id})))}',
  )
} else if (!patched.includes('N=_r(tn(z.getState().settings)),ee=N.profiles.find')) {
  throw new Error('settings profile switch marker is missing')
}

// The model controls live in SS, the input toolbar component. Keep their
// Zustand state and callbacks in that same lexical scope; placing them in the
// gallery component makes the selectors fail at runtime with undefined names.
const ssSettingsMarker = 'x=z(j=>j.settings),y=z(j=>j.setShowSettings)'
const staleSsSettingsMarker = 'x=z(j=>j.settings),j=z(j=>j.setSettings),I=z(j=>j.showToast),y=z(j=>j.setShowSettings)'
if (patched.includes(staleSsSettingsMarker)) {
  replaceOnce(
    'stale input toolbar settings state',
    staleSsSettingsMarker,
    'x=z(j=>j.settings),j=z(j=>j.setSettings),y=z(j=>j.setShowSettings)',
  )
}
if (patched.includes(ssSettingsMarker) && !patched.includes('x=z(j=>j.settings),j=z(j=>j.setSettings),y=z(j=>j.setShowSettings)')) {
  replaceOnce(
    'input toolbar settings state',
    ssSettingsMarker,
    'x=z(j=>j.settings),j=z(j=>j.setSettings),y=z(j=>j.setShowSettings)',
  )
}
if (patched.includes('p=z(M=>M.setConfirmDialog),j=z(M=>M.setSettings),I=z(M=>M.showToast),J=z(M=>M.settings),g=z(M=>M.selectedTaskIds)')) {
  replaceOnce(
    'unused gallery model state',
    'p=z(M=>M.setConfirmDialog),j=z(M=>M.setSettings),I=z(M=>M.showToast),J=z(M=>M.settings),g=z(M=>M.selectedTaskIds)',
    'p=z(M=>M.setConfirmDialog),g=z(M=>M.selectedTaskIds)',
  )
}
if (patched.includes('re=b.useRef([]),[modelPulling,setModelPulling]=b.useState(!1),ge=')) {
  replaceOnce(
    'unused gallery model loading state',
    're=b.useRef([]),[modelPulling,setModelPulling]=b.useState(!1),ge=',
    're=b.useRef([]),ge=',
  )
}
const ssModelStateMarker = '[Ke,_t]=b.useState(!1),[Ee,lt]=b.useState(!0)'
if (patched.includes(ssModelStateMarker) && !patched.includes('[modelPulling,setModelPulling]=b.useState(!1),[Ke,_t]')) {
  replaceOnce(
    'input toolbar model loading state',
    ssModelStateMarker,
    '[modelPulling,setModelPulling]=b.useState(!1),[Ke,_t]=b.useState(!1),[Ee,lt]=b.useState(!0)',
  )
}
const modelHooks = 'const modelProfiles=b.useMemo(()=>tn(x),[x]),changeProfile=b.useCallback(M=>{const P=_r(tn(z.getState().settings));P.profiles.some(le=>le.id===M)&&j(_r({...P,activeProfileId:M}))},[j]),changeModel=b.useCallback(M=>{const P=_r(tn(z.getState().settings)),le=P.profiles.find(V=>V.id===P.activeProfileId);le&&j(_r({...P,profiles:P.profiles.map(V=>V.id===le.id?{...V,model:M,modelOptions:[...new Set((V.modelOptions??[]).concat(M))]}:V)}))},[j]),pullModels=b.useCallback(async()=>{if(modelPulling)return;const M=_r(tn(z.getState().settings)),P=M.profiles.find(le=>le.id===M.activeProfileId);if(!P||!P.apiKey.trim())return;const V=_o().userEmail;if(!V){S("当前登录用户缺少邮箱信息","error");return}setModelPulling(!0);try{const ye=await fetch(`${Yo}/models`,{headers:{Authorization:`Bearer ${P.apiKey}`,"X-Sub2API-User-Email":V},cache:"no-store"});if(!ye.ok)throw new Error(`HTTP ${ye.status}`);const T=await ye.json(),C=Array.isArray(T)?T:T.data??T.models??[],B=[...new Set(C.map(G=>typeof(G==null?void 0:G.id)==="string"?G.id.trim():"").filter(G=>G&&!/video/i.test(G)&&/image|imagine|dall[-_ ]?e/i.test(G)))];if(!B.length)throw new Error("接口没有返回可用的图片模型");const G=_r(tn(z.getState().settings)),Y=G.profiles.find(X=>X.id===P.id);Y&&Y.apiKey===P.apiKey&&j(_r({...G,profiles:G.profiles.map(X=>X.id===P.id?{...X,modelOptions:B,model:B.includes(X.model)?X.model:B[0]}:X)})),S(`已拉取 ${B.length} 个图片模型`,`success`)}catch(M){console.warn("Failed to pull image models:",M),S(M instanceof Error?`模型拉取失败：${M.message}`:"模型拉取失败","error")}finally{setModelPulling(!1)}},[S,j,modelPulling]);'
const modelHooksWithEmailFallback = modelHooks.replace('const V=_o().userEmail;', 'const V=_o().userEmail||P.userEmail;')
const toolbarHookMarker = 'const ra=async()=>{if(!he){if(!gn){y(!0);return}if(a.trim()){se(!0);try{await Ob()}finally{se(!1)}}}},la=()=>{'
for (const staleHook of [
  'const modelProfiles=b.useMemo(()=>tn(J),[J])',
  'const modelProfiles=b.useMemo(()=>tn(x),[x])',
]) {
  const staleStart = patched.indexOf(staleHook)
  if (staleStart < 0 || patched.includes(modelHooksWithEmailFallback)) continue
  const staleEnd = patched.indexOf('const ra=', staleStart)
  if (staleEnd < 0) throw new Error('toolbar model hooks are malformed')
  patched = patched.slice(0, staleStart) + modelHooksWithEmailFallback + patched.slice(staleEnd)
}
if (!patched.includes(modelHooksWithEmailFallback) && patched.includes(toolbarHookMarker)) {
  replaceOnce('toolbar model hooks', toolbarHookMarker, modelHooksWithEmailFallback + toolbarHookMarker)
} else if (!patched.includes(modelHooksWithEmailFallback)) {
  throw new Error('toolbar model hook marker is missing')
}
if (patched.includes('params:Ol(g),setParams:v,activeProfile:nt,isFalProvider:!1')) {
  replaceOnce(
    'toolbar model props',
    'params:Ol(g),setParams:v,activeProfile:nt,isFalProvider:!1',
    'params:Ol(g),setParams:v,activeProfile:nt,profileOptions:modelProfiles.profiles,onProfileChange:changeProfile,onModelChange:changeModel,onModelRefresh:pullModels,isFalProvider:!1',
  )
}
if (patched.includes('const be=await ge.json(),Z=Array.isArray(be)?be:be.data??[]')) {
  replaceOnce('settings model response shape', 'const be=await ge.json(),Z=Array.isArray(be)?be:be.data??[]', 'const be=await ge.json(),Z=Array.isArray(be)?be:be.data??be.models??[]')
}
if (patched.includes('const T=await ye.json(),C=Array.isArray(T)?T:T.data??[]')) {
  replaceOnce('toolbar model response shape', 'const T=await ye.json(),C=Array.isArray(T)?T:T.data??[]', 'const T=await ye.json(),C=Array.isArray(T)?T:T.data??T.models??[]')
}

// The bootstrap already carries the authenticated email. Use it when the
// host application's localStorage snapshot is still settling after login.
const settingsModelEmail = 'const D=R.id,N=R.apiKey,H=_o().userEmail;'
const settingsModelEmailFixed = 'const D=R.id,N=R.apiKey,H=_o().userEmail||R.userEmail;'
if (patched.includes(settingsModelEmail)) replaceOnce('settings model email fallback', settingsModelEmail, settingsModelEmailFixed)

if (patched.includes('moderation')) {
  let cursor = 0
  while ((cursor = patched.indexOf('moderation', cursor)) >= 0) {
    console.error(patched.slice(Math.max(0, cursor - 100), cursor + 180))
    cursor += 10
  }
  throw new Error('moderation controls or request fields remain in the standalone bundle')
}

writeFileSync(bundlePath, patched)

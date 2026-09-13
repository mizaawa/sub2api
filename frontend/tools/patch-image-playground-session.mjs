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
const replacementWithInitReenable = replacement.replace(
  '.then(H=>{if(!H)return;const ee=z.getState();',
  '.then(H=>{if(!H)return;nr(!0);const ee=z.getState();',
)
const sessionIdentityReplacement = replacementWithInitReenable.replace(
  'const g=wk(),v=_o(),x=v.userId,y=v.tokenPresent,S=bk(),E=Bx(),A=!!(g&&x&&y&&g.userId===x&&g.userEmail&&Xl(g.userEmail)===v.userEmail),R=!!(!g&&x&&y&&S===x&&E===v.userEmail),_=A||R,U=x===null||!y||S!==x||E!==v.userEmail,O=A&&(S!==x||E!==v.userEmail)||!_&&(!!g||U),D=O?zs():Promise.resolve();',
  'const g=wk(),v=_o(),localIdentityPresent=v.userId!==null||v.userEmail!==null,x=v.userId??g?.userId,y=v.userEmail??g?.userEmail,S=bk(),E=Bx(),bootstrapMatches=!!(g&&(v.userId===null||g.userId===v.userId)&&(v.userEmail===null||Xl(g.userEmail)===v.userEmail)),ownerMismatch=!g&&x!==null&&y!==null&&((S!==null&&S!==x)||(E!==null&&E!==y)),A=!!(g&&bootstrapMatches&&v.tokenPresent&&g.userId===x&&g.userEmail&&Xl(g.userEmail)===Xl(y)),R=!!(!g&&x&&y&&v.tokenPresent&&(S===null||S===x)&&(E===null||E===y)),_=A||R,U=!x||!y||!v.tokenPresent||!!(g&&!bootstrapMatches),O=!_&&U,D=O||ownerMismatch?zs().then(()=>_):Promise.resolve(!0);',
)
  .replace('_&&x?(yk(x,v.userEmail)', '_&&x?(yk(x,y)')
  .replace(
    'const S=_o();S.userId===m&&S.userEmail===Bx()||g||(g=!0',
    'const S=_o(),ownerEmail=Bx(),localUserPresent=!!window.localStorage.getItem("auth_user"),identityMismatch=S.userId!==null&&S.userId!==m||!!S.userEmail&&!!ownerEmail&&S.userEmail!==ownerEmail,ownerMatches=bk()===m&&(!ownerEmail||!S.userEmail||S.userEmail===ownerEmail),sessionValid=!identityMismatch&&S.tokenPresent&&(S.userId===m||ownerMatches);sessionValid||g||(g=!0',
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
patched = patched.replace(/(?:async\s+){2,}function e1\(/g, 'async function e1(')

// The image input flow uploads a data URL through Io before adding it to the
// Zustand input list. Some generated bundles lost this bridge while replacing
// the session component, so the plus-button path raised a ReferenceError.
// Keep the bridge adjacent to the task/image cleanup helpers it coordinates.
const imagePersistenceBridge = 'async function Io(a,l="upload",s=Vn()){if(!He(s))return null;const i=await Qy(a,l);return He(s)?i:(await Br(i),null)}'
if (!patched.includes('async function Io(')) {
  const bridgeMarker = 'const Jf='
  const bridgeIndex = patched.indexOf(bridgeMarker)
  if (bridgeIndex < 0) throw new Error('image persistence bridge marker is missing')
  patched = `${patched.slice(0, bridgeIndex)}${imagePersistenceBridge}${patched.slice(bridgeIndex)}`
}

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

// A generated task can retain a provider URL when the provider does not allow
// CORS blob reads. Prefer the in-memory/IndexedDB value first, then try an
// optional same-origin proxy (configured by the host), and finally let the
// browser navigate to the original URL so a server Content-Disposition header
// can still trigger a download. The navigation marker is handled by both the
// single-image and ZIP download paths below.
const imageDownloadFallback = `function sub2apiImageIsSameOrigin(a){if(typeof window==="undefined")return!1;try{return new URL(a,window.location.href).origin===window.location.origin}catch{return!1}}function sub2apiImageProxyUrl(a){if(typeof window==="undefined")return"";try{const l=globalThis.__SUB2API_IMAGE_DOWNLOAD_PROXY__,s=typeof l==="function"?l(a):typeof l==="string"?l.includes("{url}")?l.replace("{url}",encodeURIComponent(a)):l:"";if(s){const i=new URL(s,window.location.href);if(i.origin===window.location.origin)return i.href}const i=new URL(a,window.location.href);return i.origin!==window.location.origin&&i.pathname.startsWith("/v1/")?new URL(i.pathname+i.search,window.location.origin).href:""}catch{return""}}function sub2apiImageExtension(a){try{const l=new URL(a,typeof window!=="undefined"?window.location.href:void 0).pathname.match(/\\.([a-z0-9]+)$/i),s=(l==null?void 0:l[1]??"").toLowerCase();return s==="jpeg"?"jpg":["png","jpg","webp","gif"].includes(s)?s:"png"}catch{return"png"}}function sub2apiImageDownloadByNavigation(a,l){if(typeof document==="undefined"||!a)return!1;const s=document.createElement("a"),i=document.body||document.documentElement;if(!i)return!1;s.href=a,s.download=l,s.target="_blank",s.rel="noopener noreferrer",s.style.display="none",i.appendChild(s);try{s.click();return!0}catch{return!1}finally{s.remove()}}function sub2apiImageNavigationError(a,l){const s=new Error(l instanceof Error?l.message:String(l));return s.downloadNavigationUrl=a,s}async function sub2apiImageResponse(a,l){try{const s=await fetch(a,{cache:"no-store",credentials:l?"include":"omit"});return s.ok||a.startsWith("data:")?s:null}catch{return null}}`
replaceFunctionOnce(
  'image download fallback helpers',
  'async function e1(',
  'function t1(',
  imageDownloadFallback + 'async function e1(',
  'sub2apiImageNavigationError',
)

const imageDownloadReader = 'async function e1(a){let l=a;if(typeof a==="string"&&!a.startsWith("data:")&&!a.startsWith("http://")&&!a.startsWith("https://")){const s=await Tn(a);s&&(l=s)}if(typeof l!=="string"||!l)throw new Error("图片地址为空");const s=l.startsWith("data:"),i=/^https?:\\/\\//i.test(l);let d=await sub2apiImageResponse(l,i&&sub2apiImageIsSameOrigin(l));if(!d&&i){const f=sub2apiImageProxyUrl(l);f&&f!==l&&(d=await sub2apiImageResponse(f,!0))}if(!d){if(i)throw sub2apiImageNavigationError(l,new Error("图片 URL 下载失败：跨域或网络异常"));throw new Error("读取图片失败："+a)}const f=await d.blob();if(!f.size&&!f.type&&i)throw sub2apiImageNavigationError(l,new Error("图片 URL 返回空内容"));if(!d.ok&&!s)throw new Error("读取图片失败："+a);return f}'
replaceFunctionOnce(
  'cache-first image download reader',
  'async function e1(',
  'function t1(',
  imageDownloadReader,
  'sub2apiImageNavigationError',
)

const imageDownloadSingle = 'async function or(a,l="images"){if(a.length===0)return{successCount:0,failCount:0};let s=0,i=0;const d=a.length>1;for(let f=0;f<a.length;f++){const m=a[f],p=typeof m==="string"?m:m==null?"":m.imageId,g=typeof m==="string"?"":m==null?"":m.downloadUrl;try{const v=await e1(p),x=String(f+1).padStart(2,"0"),y=d?l+"-"+x+"."+s0(v):l+"."+s0(v);t1(v,y),s++,d&&await mS(100)}catch(v){const x=typeof(v==null?void 0:v.downloadNavigationUrl)==="string"?v.downloadNavigationUrl:g;if(typeof x==="string"&&x){const y=String(f+1).padStart(2,"0"),S=d?l+"-"+y+"."+sub2apiImageExtension(x):l+"."+sub2apiImageExtension(x);if(sub2apiImageDownloadByNavigation(x,S)){s++,d&&await mS(100);continue}}console.error(v),i++}}return{successCount:s,failCount:i}}'
replaceFunctionOnce(
  'single image download navigation fallback',
  'async function or(',
  'async function Ws(',
  imageDownloadSingle,
  'const m=a[f],p=typeof m==="string"?m:m==null?"":m.imageId',
)

const imageDownloadZip = 'async function Ws(a,l="images"){if(a.length===0)return{successCount:0,failCount:0};let s=0,i=0,direct=0;const d={},f=new Set;for(let m=0;m<a.length;m++){const p=a[m],g=typeof p==="string"?p:p==null?"":p.imageId,v=typeof p==="string"?"":p==null?"":p.downloadUrl;try{const x=await e1(g),y=String(m+1).padStart(2,"0"),S=ly(typeof p==="string"?"":p==null?"":p.fileNameBase||("image-"+y))||("image-"+y),E=s0(x);let A=S+"."+E,R=2;for(;f.has(A);)A=S+"-"+String(R).padStart(2,"0")+"."+E,R++;f.add(A),d[A]=[new Uint8Array(await x.arrayBuffer()),{mtime:new Date}],s++}catch(x){const y=typeof(x==null?void 0:x.downloadNavigationUrl)==="string"?x.downloadNavigationUrl:v;if(typeof y==="string"&&y){const S=ly(typeof p==="string"?"":p==null?"":p.fileNameBase||("image-"+String(m+1).padStart(2,"0")))||("image-"+String(m+1).padStart(2,"0"));if(sub2apiImageDownloadByNavigation(y,S+"."+sub2apiImageExtension(y))){direct++;continue}}console.error(x),i++}}if(s>0){const m=uS(d,{level:6}),p=m.buffer.slice(m.byteOffset,m.byteOffset+m.byteLength);t1(new Blob([p],{type:"application/zip"}),(ly(l)||"images")+".zip")}return{successCount:s+direct,failCount:i}}'
replaceFunctionOnce(
  'ZIP image download navigation fallback',
  'async function Ws(',
  'function sy(',
  imageDownloadZip,
  'const p=a[m],g=typeof p==="string"?p:p==null?"":p.imageId',
)

// Keep the provider URL alongside each cached image id. If an older task was
// restored without its IndexedDB bytes, the download action can still use the
// URL returned by the provider instead of failing on the synthetic id.
const imageDownloadMetadata = 'function sy(a){return[...a].sort((l,s)=>s.createdAt-l.createdAt).flatMap(l=>Bo(l.outputImages||[],`task-${l.id}`))}function Bo(a,l="image"){return a.map((s,i)=>({imageId:s,fileNameBase:dS(l,i,a.length)}))}'
const imageDownloadMetadataFixed = 'function sy(a){return[...a].sort((l,s)=>s.createdAt-l.createdAt).flatMap(l=>Bo(l.outputImages||[],`task-${l.id}`,l.rawImageUrls??[]))}function Bo(a,l="image",s=[]){return a.map((i,d)=>({imageId:i,fileNameBase:dS(l,d,a.length),downloadUrl:s[d]}))}'
if (patched.includes(imageDownloadMetadata)) replaceOnce('image download URL metadata', imageDownloadMetadata, imageDownloadMetadataFixed)
else if (!patched.includes(imageDownloadMetadataFixed)) throw new Error('image download URL metadata marker is missing')

const imageDownloadCallFallbacks = [
  ['or([Ke],`task-${w.id}`)', 'or([{imageId:Ke,downloadUrl:xt[_t]}],`task-${w.id}`)'],
  ['Ws(Bo(w.outputImages,ve),ve)', 'Ws(Bo(w.outputImages,ve,w.rawImageUrls??[]),ve)'],
  ['or(w.outputImages,ve)', 'or(Bo(w.outputImages,ve,w.rawImageUrls??[]),ve)'],
  ['or(ne,ve)', 'or(sy(j),ve)'],
  ['or(je.map(ke=>ke.imageId),`favorites-${ve.name}`)', 'or(je,`favorites-${ve.name}`)'],
]
for (const [from, to] of imageDownloadCallFallbacks) {
  if (patched.includes(from)) replaceOnce('image download URL fallback call', from, to)
}

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

// Cached reference images are data URLs. Calling fetch(dataUrl) is needlessly
// dependent on the browser's URL loader and can fail with a generic
// "Failed to fetch" for large images before the multipart request is sent.
// Decode data URLs locally and keep fetch only for blob/http URLs.
const dataUrlBlobReader = 'async function y0(a,l="image/png"){if(typeof a==="string"&&/^data:/i.test(a)){const i=a.indexOf(",");if(i<0)throw new Error("图片 data URL 格式无效");const d=a.slice(5,i),f=a.slice(i+1),m=(d.match(/^([^;,\\s]+)/)?.[1]||l).toLowerCase();try{if(/(?:^|;)base64$/i.test(d)){const p=f.replace(/\\s+/g,"");const g=atob(p),v=new Uint8Array(g.length);for(let x=0;x<g.length;x++)v[x]=g.charCodeAt(x);return new Blob([v],{type:m})}return new Blob([decodeURIComponent(f)],{type:m})}catch{throw new Error("图片 data URL 解码失败")}}const i=await(await fetch(a)).blob();return i.type?i:new Blob([await i.arrayBuffer()],{type:l})}'
replaceFunctionOnce(
  'local data URL image decoding',
  'async function y0(',
  'async function b0(',
  dataUrlBlobReader,
  '图片 data URL 解码失败',
)

// Browser auth refreshes update several storage keys in sequence. Debounce a
// missing token/parse result, but react immediately when a different user is
// detected. This prevents a transient refresh gap from ejecting the gallery.
const sessionWatch = 'b.useEffect(()=>{if(!s||m===null||!d)return;let g=!1;const v=()=>{const S=_o(),ownerEmail=Bx(),localUserPresent=!!window.localStorage.getItem("auth_user"),sessionValid=S.userId===m&&(!localUserPresent||S.tokenPresent)&&(!S.userEmail||!ownerEmail||S.userEmail===ownerEmail)||(!localUserPresent&&bk()===m&&ownerEmail===Bx());sessionValid||g||(g=!0,f(!1),nr(!1),Ab(),zs(),Lx())},x=S=>{(S.key===null||S.key==="auth_user")&&v()};window.addEventListener("storage",x);const y=window.setInterval(v,2e3);return v(),()=>{window.removeEventListener("storage",x),window.clearInterval(y)}},[s,m,d])'
const sessionWatchPreviousFixed = 'b.useEffect(()=>{if(!s||m===null||!d)return;let g=!1,invalidSince=0;const v=()=>{const S=_o(),ownerEmail=Bx(),localUserPresent=!!window.localStorage.getItem("auth_user"),identityMismatch=S.userId!==null&&S.userId!==m||!!S.userEmail&&!!ownerEmail&&S.userEmail!==ownerEmail,sessionValid=!identityMismatch&&(S.userId===m&&(!localUserPresent||S.tokenPresent)||!localUserPresent&&S.tokenPresent&&bk()===m&&ownerEmail===Bx());if(sessionValid){invalidSince=0;return}if(!identityMismatch){if(!invalidSince)invalidSince=Date.now();if(Date.now()-invalidSince<8e3)return}g||(g=!0,f(!1),nr(!1),Ab(),zs(),Lx())},x=S=>{(S.key===null||S.key==="auth_user")&&v()};window.addEventListener("storage",x);const y=window.setInterval(v,2e3);return v(),()=>{window.removeEventListener("storage",x),window.clearInterval(y)}},[s,m,d])'
const sessionWatchLegacyFixed = 'b.useEffect(()=>{if(!s||m===null||!d)return;let g=!1,invalidSince=0;const v=()=>{const S=_o(),ownerEmail=Bx(),localUserPresent=!!window.localStorage.getItem("auth_user"),identityMismatch=S.userId!==null&&S.userId!==m||!!S.userEmail&&!!ownerEmail&&S.userEmail!==ownerEmail,sessionValid=!identityMismatch&&(S.userId===m&&(!localUserPresent||S.tokenPresent)||!localUserPresent&&S.tokenPresent&&bk()===m&&ownerEmail===Bx());if(sessionValid){invalidSince=0;return}if(!identityMismatch){if(!invalidSince)invalidSince=Date.now();if(Date.now()-invalidSince<8e3)return}g||(g=!0,f(!1),nr(!1),Ab(),zs(),Lx())},x=S=>{(S.key===null||S.key==="auth_user")&&v()};window.addEventListener("storage",x);const y=window.setInterval(v,2e3);return v(),()=>{window.removeEventListener("storage",x),window.clearInterval(y)}},[s,m,d])'
const sessionWatchFixed = 'b.useEffect(()=>{if(!s||m===null||!d)return;let g=!1,invalidSince=0;const v=()=>{const S=_o(),ownerEmail=Bx(),localUserPresent=!!window.localStorage.getItem("auth_user"),identityMismatch=S.userId!==null&&S.userId!==m||!!S.userEmail&&!!ownerEmail&&S.userEmail!==ownerEmail,ownerMatches=bk()===m&&(!ownerEmail||!S.userEmail||S.userEmail===ownerEmail),sessionValid=!identityMismatch&&S.tokenPresent&&(S.userId===m||ownerMatches);if(sessionValid){invalidSince=0;return}if(!identityMismatch){if(!invalidSince)invalidSince=Date.now();if(Date.now()-invalidSince<8e3)return}g||(g=!0,f(!1),nr(!1),Ab(),zs(),Lx())},x=S=>{(S.key===null||S.key==="auth_user")&&v()};window.addEventListener("storage",x);const y=window.setInterval(v,2e3);return v(),()=>{window.removeEventListener("storage",x),window.clearInterval(y)}},[s,m,d])'
if (patched.includes(sessionWatch)) replaceOnce('debounced standalone session watch', sessionWatch, sessionWatchFixed)
else if (patched.includes(sessionWatchPreviousFixed)) replaceOnce('updated standalone session watch', sessionWatchPreviousFixed, sessionWatchFixed)
else if (patched.includes(sessionWatchLegacyFixed)) replaceOnce('legacy standalone session watch', sessionWatchLegacyFixed, sessionWatchFixed)
else if (!patched.includes(sessionWatchFixed)) {
  // Accept minor vendor/minifier differences by replacing the effect bounded
  // by its stable dependency tuple rather than relying on one exact body.
  const watchStart = patched.indexOf('b.useEffect(()=>{if(!s||m===null||!d)return;')
  const watchEnd = patched.indexOf('},[s,m,d])', watchStart)
  if (watchStart < 0 || watchEnd < watchStart) throw new Error('debounced standalone session watch marker is missing')
  patched = `${patched.slice(0, watchStart)}${sessionWatchFixed}${patched.slice(watchEnd + '},[s,m,d])'.length)}`
}

// Keep provider requests same-origin so browser preflight policy cannot block
// the Authorization and owner-email headers. Deployments on the zayu hostname
// still resolve this to the same zayu `/v1` endpoint.
const remoteBaseUrl = 'Yo="https://api.zayuapi.com/v1"'
const sameOriginBaseUrl = 'Yo=typeof window>"u"?"https://api.zayuapi.com/v1":`${window.location.origin}/v1`'
if (patched.includes(remoteBaseUrl)) replaceOnce('same-origin image API base URL', remoteBaseUrl, sameOriginBaseUrl)
const staleSameOriginBaseUrl = 'Yo=typeof window<"u"?"https://api.zayuapi.com/v1":`${window.location.origin}/v1`'
if (patched.includes(staleSameOriginBaseUrl)) replaceOnce('same-origin image API base URL condition', staleSameOriginBaseUrl, sameOriginBaseUrl)

// Some mobile WebViews drop an Authorization header when a navigation or
// embedded fetch crosses their request boundary. The gateway accepts the
// equivalent X-API-Key header, so emit both from the same selected profile.
const bearerHeaders = 'const l={Authorization:`Bearer ${a.apiKey}`}'
const bearerHeadersFixed = 'const l={Authorization:`Bearer ${a.apiKey}`,"X-API-Key":a.apiKey}'
if (patched.includes(bearerHeaders)) replaceOnce('dual image API key headers', bearerHeaders, bearerHeadersFixed)

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
// Preserve the managed profile's Base64 response preference while normalizing
// settings. Without this field j5() silently drops the launcher hint and the
// browser falls back to downloading a provider URL across origins.
const profileBase64Response = 'transparentBackgroundMethod:"api"}}function tn('
const profileBase64ResponseFixed = 'transparentBackgroundMethod:"api",responseFormatB64Json:s.responseFormatB64Json===true}}function tn('
if (patched.includes(profileBase64Response)) replaceOnce('preserve Base64 image response preference', profileBase64Response, profileBase64ResponseFixed)
else if (!patched.includes(profileBase64ResponseFixed)) throw new Error('Base64 image response preference marker is missing')

const jsonBodyBuild = 'const A=Qs(a.body??{},p);s.responseFormatB64Json&&A&&typeof A=="object"&&!Array.isArray(A)&&(A.response_format="b64_json"),S=JSON.stringify(A)'
const jsonBodyBuildFixed = 'const A=Qs(a.body??{},p);A&&typeof A=="object"&&!Array.isArray(A)&&A.output_compression==null&&delete A.output_compression,s.responseFormatB64Json&&A&&typeof A=="object"&&!Array.isArray(A)&&(A.response_format="b64_json"),S=JSON.stringify(A)'
if (patched.includes(jsonBodyBuild)) replaceOnce('omit null image compression', jsonBodyBuild, jsonBodyBuildFixed)

// Keep the bootstrap carrier until the standalone app has finished restoring
// its settings. Mobile WebViews/PWA launches can recreate sessionStorage, so
// the launcher also writes a short-lived same-origin localStorage fallback.
// Prefer a candidate matching the local auth snapshot when both stores contain
// data; this prevents an older tab's carrier from hiding a fresh handoff.
const bootstrapReaderReplacement = 'function wk(){if(typeof window>"u")return null;const a=[];try{const l=window.sessionStorage.getItem(zx);l&&a.push({value:l,fallback:!1})}catch{}try{const l=window.localStorage.getItem(zx+":fallback");l&&a.push({value:l,fallback:!0})}catch{}const l=_o();for(const s of a){const i=s.value;if(!i.startsWith(_x))continue;try{const d=JSON.parse(i.slice(_x.length));if(!d||typeof d!=="object"||Array.isArray(d))continue;const f=F0(d.userId),m=d.settings;if(!f||!m||typeof m!=="object"||Array.isArray(m))continue;const p=Number(d.issuedAt);if(s.fallback&&(!Number.isFinite(p)||Date.now()-p>120000||Date.now()-p< -30000))continue;if(l.userId!==null&&l.userId!==f)continue;const g=Xl(d.userEmail);if(l.userEmail!==null&&g!==l.userEmail)continue;return{settings:m,userId:f,...g?{userEmail:g}:{}}}catch{}}return null}function clearImagePlaygroundBootstrap(){try{window.sessionStorage.removeItem(zx)}catch{}try{window.localStorage.removeItem(zx+":fallback")}catch{}}'
replaceFunctionOnce(
  'durable standalone bootstrap carrier',
  'function wk(){',
  'const Yb=',
  bootstrapReaderReplacement,
  'localStorage.getItem(zx+":fallback")',
)

// The scroll lock helper must receive the dialog container so wheel and touch
// gestures can reach the inner overflow-y-auto settings panel.
const settingsScrollLock = 'Aa(a,()=>i(!1)),aa(a),!a)return'
const settingsScrollLockFixed = 'Aa(a,()=>i(!1)),aa(a,y),!a)return'
if (patched.includes(settingsScrollLock)) replaceOnce('settings scroll lock container', settingsScrollLock, settingsScrollLockFixed)
else if (!patched.includes(settingsScrollLockFixed) && !patched.includes('data-image-settings-dialog')) throw new Error('settings scroll lock container: marker is missing')
const settingsDialog = 'o.jsxs("section",{role:"dialog","aria-modal":"true","aria-label":"设置",className:'
const settingsDialogFixed = 'o.jsxs("section",{ref:y,role:"dialog","aria-modal":"true","aria-label":"设置",className:'
if (patched.includes(settingsDialog)) replaceOnce('settings dialog scroll ref', settingsDialog, settingsDialogFixed)
else if (!patched.includes(settingsDialogFixed) && !patched.includes('data-image-settings-dialog')) throw new Error('settings dialog scroll ref: marker is missing')

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
const modelControlBase = 'o.jsxs("div",{"data-image-profile-model":!0,className:"relative min-w-0 grid grid-cols-1 gap-1 sm:grid-cols-2",children:[o.jsxs("label",{className:"flex min-w-0 flex-col gap-0.5",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"配置"}),o.jsx(Gs,{value:i.id,onChange:V=>pc&&pc(V.target.value),options:(po??[]).map(V=>({label:V.name,value:V.id})),showValueTooltips:!1,className:g})]}),o.jsxs("label",{className:"flex min-w-0 flex-col gap-0.5",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"模型"}),o.jsx(Gs,{value:i.model,onChange:V=>mc&&mc(V.target.value),onOpenChange:V=>{V&&mr&&mr()},options:[...new Set((i.modelOptions??[]).concat(i.model||[]))].filter(V=>V).map(V=>({label:V,value:V})),showValueTooltips:!1,className:g})]})]})'
const modelControl = modelControlBase
const modelControlWithoutClick = modelControl.replace('onClick:()=>mr&&mr(),', '')
const modelControlWithoutHook = modelControl.replace('"data-image-profile-model":!0,', '')
const legacyPatchedModelControl = 'o.jsxs("div",{"data-image-profile-model":!0,className:"relative flex min-w-0 flex-col gap-0.5",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"配置 / 模型"}),o.jsxs("div",{className:"grid min-w-0 grid-cols-1 gap-1 sm:grid-cols-2",children:[o.jsx("select",{value:i.id,onChange:V=>pc&&pc(V.target.value),className:`${g} min-w-0 w-full`,"aria-label":"当前配置",children:(po??[]).map(V=>o.jsx("option",{value:V.id,children:V.name},V.id))}),o.jsx("select",{value:i.model,onFocus:()=>mr&&mr(),onClick:()=>mr&&mr(),onChange:V=>mc&&mc(V.target.value),className:`${g} min-w-0 w-full`,"aria-label":"选择模型",children:[...new Set((i.modelOptions??[]).concat(i.model||[]))].filter(V=>V).map(V=>o.jsx("option",{value:V,children:V},V))})]})]})'
const legacyNativeSplitModelControl = 'o.jsxs("div",{"data-image-profile-model":!0,className:"relative min-w-0 grid grid-cols-1 gap-1 sm:grid-cols-2",children:[o.jsxs("label",{className:"flex min-w-0 flex-col gap-0.5",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"配置"}),o.jsx("select",{value:i.id,onChange:V=>pc&&pc(V.target.value),className:`${g} min-w-0 w-full`,"aria-label":"当前配置",children:(po??[]).map(V=>o.jsx("option",{value:V.id,children:V.name},V.id))})]}),o.jsxs("label",{className:"flex min-w-0 flex-col gap-0.5",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"模型"}),o.jsx("select",{value:i.model,onFocus:()=>mr&&mr(),onClick:()=>mr&&mr(),onChange:V=>mc&&mc(V.target.value),className:`${g} min-w-0 w-full`,"aria-label":"选择模型",children:[...new Set((i.modelOptions??[]).concat(i.model||[]))].filter(V=>V).map(V=>o.jsx("option",{value:V,children:V},V))})]})]})'
const legacyModelControl = 'o.jsxs("div",{className:"relative flex flex-col gap-0.5",children:[o.jsx("span",{className:"text-gray-400 dark:text-gray-500 ml-1",children:"配置 / 模型"}),o.jsxs("div",{className:"flex gap-1",children:[o.jsx("select",{value:i.id,onChange:V=>pc&&pc(V.target.value),className:g,"aria-label":"当前配置",children:(po??[]).map(V=>o.jsx("option",{value:V.id,children:V.name},V.id))}),o.jsx("select",{value:i.model,onFocus:()=>mr&&mr(),onClick:()=>mr&&mr(),onChange:V=>mc&&mc(V.target.value),className:g,"aria-label":"选择模型",children:[...new Set((i.modelOptions??[]).concat(i.model||[]))].filter(V=>V).map(V=>o.jsx("option",{value:V,children:V},V))})]})]})'
if (toolbar.includes(legacyModelControl)) toolbar = toolbar.replace(legacyModelControl, modelControl)
else if (toolbar.includes(legacyPatchedModelControl)) toolbar = toolbar.replace(legacyPatchedModelControl, modelControl)
else if (toolbar.includes(legacyNativeSplitModelControl)) toolbar = toolbar.replace(legacyNativeSplitModelControl, modelControl)
else if (toolbar.includes(modelControlWithoutClick)) toolbar = toolbar.replace(modelControlWithoutClick, modelControl)
else if (toolbar.includes(modelControlWithoutHook)) toolbar = toolbar.replace(modelControlWithoutHook, modelControl)
// Also recognise the already-patched form where onChange handlers pass the value
// directly (V) instead of via V.target.value — produced by the onChange fix patches.
const modelControlFixed = modelControl
  .replace('onChange:V=>pc&&pc(V.target.value)', 'onChange:V=>pc&&pc(V)')
  .replace('onChange:V=>mc&&mc(V.target.value),onOpenChange:V=>{V&&mr&&mr()}', 'onChange:V=>mc&&mc(V),onOpenChange:V=>{V&&(i.modelOptions??[]).filter(m=>m).length<=1&&mr&&mr()}')
const modelControlAlreadyPatched = toolbar.includes('data-image-profile-model":!0') &&
  toolbar.includes('onChange:V=>pc&&pc(V)') &&
  toolbar.includes('onChange:V=>mc&&mc(V)')
if (toolbar.includes(moderationControl)) toolbar = toolbar.replace(moderationControl, modelControl)
else if (!toolbar.includes(modelControl) && !toolbar.includes(modelControlFixed) && !modelControlAlreadyPatched) throw new Error('toolbar model control marker is missing')

// Model identifiers are often prefixed with a provider path. Mark only the
// model trigger so its constrained label uses a leading ellipsis and keeps the
// useful suffix visible (for example, the actual model family and version).
const modelLabelMarker = 'children:"模型"}),o.jsx(Gs,{value:i.model,'
const modelLabelStart = toolbar.indexOf(modelLabelMarker)
if (modelLabelStart < 0) throw new Error('model selector label marker is missing')
const modelLabelClassMarker = 'showValueTooltips:!1,className:g'
const modelLabelClassStart = toolbar.indexOf(modelLabelClassMarker, modelLabelStart)
if (modelLabelClassStart < 0) throw new Error('model selector class marker is missing')
const modelLabelClassFixed = `${modelLabelClassMarker}+" image-model-select"`
if (!toolbar.includes(modelLabelClassFixed, modelLabelClassStart)) {
  toolbar = `${toolbar.slice(0, modelLabelClassStart)}${modelLabelClassFixed}${toolbar.slice(modelLabelClassStart + modelLabelClassMarker.length)}`
}
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
} else if (!patched.includes('N=_r(tn(z.getState().settings)),ee=N.profiles.find') && !patched.includes('data-image-settings-dialog')) {
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
if (patched.includes(ssModelStateMarker) && !patched.includes('[modelPulling,setModelPulling]=b.useState(!1)')) {
  replaceOnce(
    'input toolbar model loading state',
    ssModelStateMarker,
    '[modelPulling,setModelPulling]=b.useState(!1),[Ke,_t]=b.useState(!1),[Ee,lt]=b.useState(!0)',
  )
}
const inputCollapseStateMarker = '[modelPulling,setModelPulling]=b.useState(!1),[Ke,_t]=b.useState(!1)'
const inputCollapseStateFixed = '[modelPulling,setModelPulling]=b.useState(!1),[inputPanelCollapsed,setInputPanelCollapsed]=b.useState(!1),[Ke,_t]=b.useState(!1)'
if (!patched.includes('[inputPanelCollapsed,setInputPanelCollapsed]=b.useState(!1)')) {
  if (patched.includes(inputCollapseStateMarker)) {
    replaceOnce('input panel collapse state', inputCollapseStateMarker, inputCollapseStateFixed)
  } else {
    throw new Error('input panel collapse state marker is missing')
  }
}
const modelHooks = 'const modelProfiles=b.useMemo(()=>tn(x),[x]),changeProfile=b.useCallback(M=>{const P=_r(tn(z.getState().settings));P.profiles.some(le=>le.id===M)&&j(_r({...P,activeProfileId:M}))},[j]),changeModel=b.useCallback(M=>{const P=_r(tn(z.getState().settings)),le=P.profiles.find(V=>V.id===P.activeProfileId);le&&j(_r({...P,profiles:P.profiles.map(V=>V.id===le.id?{...V,model:M,modelOptions:[...new Set((V.modelOptions??[]).concat(M))]}:V)}))},[j]),pullModels=b.useCallback(async()=>{if(modelPulling)return;const M=_r(tn(z.getState().settings)),P=M.profiles.find(le=>le.id===M.activeProfileId);if(!P||!P.apiKey.trim())return;const V=_o().userEmail;if(!V){S("当前登录用户缺少邮箱信息","error");return}setModelPulling(!0);try{const ye=await fetch(`${Yo}/models`,{headers:{Authorization:`Bearer ${P.apiKey}`,"X-API-Key":P.apiKey,"X-Sub2API-User-Email":V},cache:"no-store",credentials:"include"});if(!ye.ok)throw new Error(`HTTP ${ye.status}`);const T=await ye.json(),C=Array.isArray(T)?T:T.data??T.models??[],B=[...new Set(C.map(G=>typeof(G==null?void 0:G.id)==="string"?G.id.trim():"").filter(G=>G&&!/video/i.test(G)&&/image|imagine|dall[-_ ]?e/i.test(G)))];if(!B.length)throw new Error("接口没有返回可用的图片模型");const G=_r(tn(z.getState().settings)),Y=G.profiles.find(X=>X.id===P.id);Y&&Y.apiKey===P.apiKey&&j(_r({...G,profiles:G.profiles.map(X=>X.id===P.id?{...X,modelOptions:B,model:B.includes(X.model)?X.model:B[0]}:X)})),S(`已拉取 ${B.length} 个图片模型`,`success`)}catch(M){console.warn("Failed to pull image models:",M),S(M instanceof Error?`模型拉取失败：${M.message}`:"模型拉取失败","error")}finally{setModelPulling(!1)}},[S,j,modelPulling]);'
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

// The managed workbench exposes API-key profiles through the home toolbar.
// The settings dialog is intentionally limited to browser-local preferences;
// keeping a second API configuration editor there made mobile navigation
// ambiguous and allowed stale credentials to be selected accidentally.
const settingsStart = patched.indexOf('function OS(){')
const settingsEnd = patched.indexOf('function i1(', settingsStart)
if (settingsStart < 0 || settingsEnd < 0) throw new Error('settings dialog markers are missing')
const settingsDialogReplacement = `function OS(){const a=z(D=>D.showSettings),l=z(D=>D.settings),s=z(D=>D.setSettings),i=z(D=>D.setShowSettings),[f,m]=b.useState("local"),y=b.useRef(null),A=b.useMemo(()=>_r(tn(l)),[l]);return Aa(a,()=>i(!1)),aa(a,y),!a?null:vr.createPortal(o.jsxs("div",{"data-no-drag-select":!0,"data-image-settings-dialog":!0,className:"fixed inset-0 z-[100] flex items-center justify-center p-4",onClick:()=>i(!1),children:[o.jsx("div",{className:"absolute inset-0 bg-black/30 backdrop-blur-sm"}),o.jsxs("section",{ref:y,role:"dialog","aria-modal":"true","aria-label":"设置",className:"relative z-10 flex max-h-[88vh] w-full max-w-2xl flex-col overflow-hidden rounded-3xl border border-white/50 bg-white/95 shadow-2xl dark:border-white/[0.08] dark:bg-gray-900/95",onClick:D=>D.stopPropagation(),children:[o.jsxs("header",{className:"flex items-center justify-between border-b border-gray-200/80 px-5 py-4 dark:border-white/[0.08]",children:[o.jsx("h2",{className:"text-base font-semibold text-gray-800 dark:text-gray-100",children:"设置"}),o.jsx("button",{type:"button",onClick:()=>i(!1),className:"rounded-lg p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-white/[0.06]","aria-label":"关闭",children:"×"})]}),o.jsxs("div",{className:"flex min-h-0 flex-1 flex-col sm:flex-row",children:[o.jsx("nav",{className:"flex shrink-0 gap-1 overflow-x-auto border-b border-gray-200/80 p-2 sm:w-40 sm:flex-col sm:border-b-0 sm:border-r dark:border-white/[0.08]","aria-label":"设置分类",children:[["local","本地数据"],["about","关于"]].map(([D,N])=>o.jsx("button",{type:"button",onClick:()=>m(D),className:\`whitespace-nowrap rounded-xl px-3 py-2.5 text-left text-sm transition-colors \${f===D?"bg-blue-50 font-medium text-blue-600 dark:bg-blue-500/10 dark:text-blue-400":"text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-white/[0.04]"}\`,children:N},D))}),o.jsx("div",{className:"min-h-0 flex-1 overflow-y-auto p-5",children:[f==="local"&&o.jsxs("div",{className:"space-y-4 text-sm text-gray-600 dark:text-gray-300",children:[o.jsx("p",{className:"rounded-xl bg-gray-50 p-4 leading-relaxed dark:bg-white/[0.03]",children:"配置、任务记录和图片仅写入当前浏览器的本地存储，不会创建云端配置记录。"}),o.jsxs("label",{className:"flex items-center justify-between gap-4 rounded-xl border border-gray-200/70 p-4 dark:border-white/[0.08]",children:[o.jsx("span",{children:"提交后清空提示词"}),o.jsx("input",{type:"checkbox",checked:A.clearInputAfterSubmit,onChange:D=>s({clearInputAfterSubmit:D.target.checked})})]}),o.jsxs("label",{className:"flex items-center justify-between gap-4 rounded-xl border border-gray-200/70 p-4 dark:border-white/[0.08]",children:[o.jsx("span",{children:"重新打开页面时保留输入内容"}),o.jsx("input",{type:"checkbox",checked:A.persistInputOnRestart,onChange:D=>s({persistInputOnRestart:D.target.checked})})]})]}),f==="about"&&o.jsxs("div",{className:"space-y-3 text-sm text-gray-600 dark:text-gray-300",children:[o.jsx("h3",{className:"text-base font-semibold text-gray-800 dark:text-gray-100",children:"小杂鱼の生图"}),o.jsx("p",{children:"仅提供图片画廊和图片生成能力。"}),o.jsx("p",{className:"text-xs text-gray-500 dark:text-gray-400",children:"基于 GPT Image Playground（MIT License）构建。"})]})]})]})]})]}),document.body)}`
if (patched.slice(settingsStart, settingsEnd) !== settingsDialogReplacement) {
  patched = patched.slice(0, settingsStart) + settingsDialogReplacement + patched.slice(settingsEnd)
}

// Keep the prompt compact and predictable on touch screens. The full-screen
// toggle consumed the same row as the prompt and frequently obscured text.
const hasCompactPromptRow = patched.includes('data-image-prompt-row')
const promptStart = hasCompactPromptRow ? -1 : patched.indexOf('o.jsxs("div",{className:"relative",children:[rn&&', patched.indexOf('function SS(){'))
const actionStart = hasCompactPromptRow ? -1 : patched.indexOf('o.jsxs("div",{"data-image-actions"', promptStart)
if (!hasCompactPromptRow && (promptStart < 0 || actionStart < 0)) throw new Error('image prompt markers are missing')
const promptBlock = hasCompactPromptRow ? '' : patched.slice(promptStart, actionStart)
const expandButton = 'o.jsx("button",{type:"button",className:"absolute bottom-2 right-2 rounded p-1 text-gray-400 hover:bg-gray-100",onClick:()=>le(j=>!j),"aria-label":P?"收起输入框":"展开输入框",children:P?o.jsx(Ok,{className:"h-4 w-4"}):o.jsx(Ik,{className:"h-4 w-4"})})'
const addButton = 'o.jsx("button",{"data-image-add":!0,type:"button",className:"flex h-[42px] w-[42px] shrink-0 items-center justify-center rounded-2xl border border-gray-200/60 bg-white/50 text-xl leading-none text-gray-500 shadow-sm transition hover:bg-white hover:text-gray-700 dark:border-white/[0.1] dark:bg-white/[0.03] dark:text-gray-300 dark:hover:bg-white/[0.08]","aria-label":"添加图片",onClick:()=>{var j;return(j=ft.current)==null?void 0:j.click()},children:"+"})'
// Older runs could leave a separator after the removed expand button. Repair
// that exact prompt-only sequence before validating the generated JavaScript.
if (patched.includes('data-image-prompt-row') && patched.includes('})}),]})')) {
  patched = patched.replace('})}),]})', '})})]})')
}
if (!hasCompactPromptRow) {
  let compactPrompt = promptBlock
  if (compactPrompt.includes(expandButton)) compactPrompt = compactPrompt.replace(expandButton, '')
  compactPrompt = compactPrompt.replace(',]})', ']})')
  compactPrompt = compactPrompt.replace('})}),]})', '})})]})')
  const trailingComma = compactPrompt.endsWith(',') ? ',' : ''
  if (trailingComma) compactPrompt = compactPrompt.slice(0, -1)
  const wrappedPrompt = `o.jsxs("div",{"data-image-prompt-row":!0,className:"flex min-w-0 items-stretch gap-2",children:[${compactPrompt},${addButton}]})${trailingComma}`
  patched = patched.slice(0, promptStart) + wrappedPrompt + patched.slice(actionStart)
}
const promptContainer = 'o.jsxs("div",{className:"relative",children:[rn&&'
const flexiblePromptContainer = 'o.jsxs("div",{className:"relative min-w-0 flex-1",children:[rn&&'
if (patched.includes(promptContainer)) replaceOnce('flexible image prompt container', promptContainer, flexiblePromptContainer)
const legacyPrompt = '描述你想生成的图片，可输入 @ 来指定参考图...'
if (patched.includes(legacyPrompt)) replaceOnce('image prompt placeholder', legacyPrompt, '描述你所想的图片')
const legacyTransparentHint = '实现方式可在设置的 API 配置中选择'
if (patched.includes(legacyTransparentHint)) replaceOnce('transparent output hint', legacyTransparentHint, '透明背景由当前分组接口配置决定')

// The old action row had a second settings button and a wide reference-image
// button. Both are now represented by the compact add-image control beside
// the prompt; leave the generation action and home parameter selectors here.
const toolbarSettingsButton = 'o.jsx("button",{type:"button",className:"rounded-xl border border-gray-200 p-2.5 text-gray-600 hover:bg-gray-100 dark:border-white/[0.1] dark:text-gray-300",onClick:()=>y(!0),"aria-label":"打开设置",children:o.jsx(Gb,{className:"h-5 w-5"})}),'
if (patched.includes(toolbarSettingsButton)) replaceOnce('input toolbar settings button', toolbarSettingsButton, '')
const toolbarAddButton = 'o.jsx("button",{type:"button",className:"rounded-xl border border-gray-200 p-2.5 text-gray-600 hover:bg-gray-100 dark:border-white/[0.1] dark:text-gray-300",onClick:()=>{var j;return(j=ft.current)==null?void 0:j.click()},"aria-label":"添加参考图",children:"＋"}),'
if (patched.includes(toolbarAddButton)) replaceOnce('input toolbar add button', toolbarAddButton, '')
const generateButtonPrefix = 'o.jsx("button",{type:"button",className:"flex-1 rounded-xl bg-blue-600 px-4 py-3'
const generateButtonFixedPrefix = 'o.jsx("button",{"data-image-generate":!0,type:"button",className:"flex-1 rounded-xl bg-blue-600 px-4 py-3'
if (patched.includes(generateButtonPrefix)) replaceOnce('generation action hook', generateButtonPrefix, generateButtonFixedPrefix)
const staleToolbarSettingsState = 'x=z(j=>j.settings),j=z(j=>j.setSettings),y=z(j=>j.setShowSettings),S='
const compactToolbarSettingsState = 'x=z(j=>j.settings),j=z(j=>j.setSettings),S='
if (patched.includes(staleToolbarSettingsState)) replaceOnce('unused input toolbar settings state', staleToolbarSettingsState, compactToolbarSettingsState)

// A missing profile/model selector must remain visible on desktop. Give the
// profile/model pair two grid tracks while keeping the other controls compact.
const actionClass = 'className:"mt-3 flex items-end gap-2"'
const responsiveActionClass = 'className:"mt-3 flex flex-wrap items-end gap-2"'
if (patched.includes(actionClass)) replaceOnce('responsive action row wrapping', actionClass, responsiveActionClass)

// Keep a compact control at the input surface's upper-left corner so the
// prompt, parameters, and generation action can be hidden without losing the
// in-memory draft. The button remains mounted while the panel is collapsed.
const inputPanel = 'o.jsxs("div",{className:"rounded-2xl border border-white/50 bg-white/80 p-3 shadow-[0_8px_30px_rgb(0,0,0,0.08)] backdrop-blur-2xl dark:border-white/[0.08] dark:bg-gray-900/80 sm:rounded-3xl sm:p-4",children:'
const collapsibleInputPanel = 'o.jsx("button",{"data-input-collapse-toggle":!0,type:"button",onClick:()=>setInputPanelCollapsed(V=>!V),"aria-expanded":!inputPanelCollapsed,"aria-controls":"image-playground-input-panel","aria-label":inputPanelCollapsed?"展开输入栏":"折叠输入栏",title:inputPanelCollapsed?"展开输入栏":"折叠输入栏",className:"mb-2 flex h-[42px] w-[42px] shrink-0 items-center justify-center rounded-xl border border-gray-200/70 bg-white/90 text-gray-500 shadow-sm transition hover:bg-gray-50 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500/30 dark:border-white/[0.1] dark:bg-gray-900/90 dark:text-gray-300 dark:hover:bg-white/[0.08]",children:o.jsx(_k,{className:`h-4 w-4 transition-transform duration-200 ${inputPanelCollapsed?"rotate-180":""}`})}),!inputPanelCollapsed&&o.jsxs("div",{id:"image-playground-input-panel","data-input-panel-content":!0,className:"rounded-2xl border border-white/50 bg-white/80 p-3 shadow-[0_8px_30px_rgb(0,0,0,0.08)] backdrop-blur-2xl dark:border-white/[0.08] dark:bg-gray-900/80 sm:rounded-3xl sm:p-4",children:'
if (patched.includes(inputPanel)) {
  replaceOnce('collapsible image input panel', inputPanel, collapsibleInputPanel)
} else if (!patched.includes('data-input-collapse-toggle')) {
  throw new Error('collapsible image input panel marker is missing')
}

// Requests with incomplete managed profile data should show a toast instead
// of opening a settings page that no longer contains API configuration.
const openSettingsOnInvalidApi = 'l.showToast(`请先完善请求 API 配置：${f}`,"error"),l.setShowSettings(!0);'
const invalidManagedProfileToast = 'l.showToast(`当前分组密钥不可用：${f}`,"error");'
if (patched.includes(openSettingsOnInvalidApi)) replaceOnce('invalid profile settings redirect', openSettingsOnInvalidApi, invalidManagedProfileToast)
const staleInvalidManagedProfileToast = 'l.showToast(`请先完善请求 API 配置：${f}`,"error");'
if (patched.includes(staleInvalidManagedProfileToast)) replaceOnce('invalid profile toast copy', staleInvalidManagedProfileToast, invalidManagedProfileToast)

if (patched.includes('aria-label":"打开设置"')) throw new Error('input toolbar settings action remains in standalone bundle')
if (patched.includes(legacyPrompt)) throw new Error('legacy image prompt placeholder remains in standalone bundle')
if (patched.includes('aria-label":P?"收起输入框":"展开输入框"')) throw new Error('fullscreen prompt action remains in standalone bundle')
if (!patched.includes('描述你所想的图片')) throw new Error('new image prompt placeholder is missing')
if (!patched.includes('data-image-prompt-row') || !patched.includes('data-image-add')) throw new Error('compact image prompt row is missing')
if (!patched.includes('data-image-profile-model')) throw new Error('profile and model layout hook is missing')
if (!patched.includes('data-image-generate')) throw new Error('generation action hook is missing')
if (!patched.includes('data-input-collapse-toggle') || !patched.includes('data-input-panel-content')) throw new Error('input panel collapse controls are missing')
if (patched.includes('children:"＋"')) throw new Error('legacy wide image add action remains in standalone bundle')
if (patched.includes('["api","API 配置"]')) throw new Error('API configuration tab remains in settings dialog')

if (patched.includes('moderation')) {
  let cursor = 0
  while ((cursor = patched.indexOf('moderation', cursor)) >= 0) {
    console.error(patched.slice(Math.max(0, cursor - 100), cursor + 180))
    cursor += 10
  }
  throw new Error('moderation controls or request fields remain in the standalone bundle')
}

// Remove the upstream version-check badge. The embedded workbench is vendored
// at a fixed commit; the upstream release tag is irrelevant here and the badge
// makes users think the whole sub2api deployment is outdated.
const versionCheckFn = 'function jk(){const[a,l]=b.useState(null),[s,i]=b.useState(()=>sessionStorage.getItem("version-dismissed")==="true");return b.useEffect(()=>{let m=!1;return fetch(kk,{headers:{Accept:"application/vnd.github.v3+json"}}).then(p=>{if(!p.ok)throw new Error(`HTTP ${p.status}`);return p.json()}).then(p=>{if(m)return;const g=p.tag_name??"",v=g.replace(/^v/,"");v&&Sk(v,"0.7.8")>0&&l({tag:g,url:p.html_url??`https://github.com/${Yb}/releases/latest`})}).catch(()=>{}),()=>{m=!0}},[]),{hasUpdate:a!==null&&!s,latestRelease:a,dismiss:()=>{i(!0),sessionStorage.setItem("version-dismissed","true")}}}'
const versionCheckStub = 'function jk(){return{hasUpdate:!1,latestRelease:null,dismiss:()=>{}}}'
if (patched.includes(versionCheckFn)) replaceOnce('version check stub', versionCheckFn, versionCheckStub)
else if (!patched.includes(versionCheckStub) && !patched.includes('function jk(){return{hasUpdate:!1')) throw new Error('version check function marker is missing')

// The toolbar model/profile selectors use the custom Gs (Select) component which
// calls onChange(optionValue) directly — not onChange(event). The wiring
// onChange:V=>pc&&pc(V.target.value) therefore passes undefined to changeProfile/
// changeModel, breaking both selectors silently. Fix to pass V directly.
const profileOnChangeBroken = 'onChange:V=>pc&&pc(V.target.value)'
const profileOnChangeFixed = 'onChange:V=>pc&&pc(V)'
if (patched.includes(profileOnChangeBroken)) replaceOnce('profile selector onChange', profileOnChangeBroken, profileOnChangeFixed)
else if (!patched.includes(profileOnChangeFixed)) throw new Error('profile selector onChange marker is missing')

const modelOnChangeBroken = 'onChange:V=>mc&&mc(V.target.value)'
const modelOnChangeFixed = 'onChange:V=>mc&&mc(V)'
if (patched.includes(modelOnChangeBroken)) replaceOnce('model selector onChange', modelOnChangeBroken, modelOnChangeFixed)
else if (!patched.includes(modelOnChangeFixed)) throw new Error('model selector onChange marker is missing')

// The model selector onOpenChange fires pullModels on every dropdown open, causing
// repeated API calls for the same profile. Guard it so models are only fetched when
// the current profile has no cached options yet (i.e. modelOptions list is empty or
// contains only the one bootstrap placeholder model).
const modelOnOpenBroken = 'onOpenChange:V=>{V&&mr&&mr()}'
const modelOnOpenFixed = 'onOpenChange:V=>{V&&(i.modelOptions??[]).filter(m=>m).length<=1&&mr&&mr()}'
if (patched.includes(modelOnOpenBroken)) replaceOnce('model selector onOpenChange guard', modelOnOpenBroken, modelOnOpenFixed)
else if (!patched.includes(modelOnOpenFixed)) throw new Error('model selector onOpenChange marker is missing')

writeFileSync(bundlePath, patched)

import { readFileSync, writeFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const bundlePath = resolve(root, 'public/image-playground/assets/index-rvKuKMn_.js')
const source = readFileSync(bundlePath, 'utf8')
const startMarker = 'function oj(){'
const endMarker = 'const cj="width=device-width'
const start = source.indexOf(startMarker)
const end = source.indexOf(endMarker, start)

if (start < 0 || end < 0 || source.indexOf(startMarker, start + startMarker.length) >= 0) {
  throw new Error('Image playground session component markers are missing or ambiguous')
}

// The Go static server already rejects the standalone HTML and every asset when
// the feature is disabled. At runtime, bind local data to the bootstrap identity
// once, then observe auth_user and token presence. Access/refresh token rotation
// writes several localStorage keys in sequence and must not look like an account change.
// A missing auth token still invalidates the standalone session, even when
// auth_user remains in localStorage during logout.
const replacement = `function oj(){const a=z(g=>g.filterFavorite),l=z(g=>g.activeFavoriteCollectionId),[s,i]=b.useState(!1),[d,f]=b.useState(!1),[m,p]=b.useState(null);return AS(),b.useEffect(()=>{if(vy)return;vy=!0;const g=wk(),v=_o(),x=v.userId,y=v.tokenPresent,S=bk(),E=Bx(),A=!!(g&&x&&y&&g.userId===x&&g.userEmail&&Xl(g.userEmail)===v.userEmail),R=!!(!g&&x&&y&&S===x&&E===v.userEmail),_=A||R,U=x===null||!y||S!==x||E!==v.userEmail,O=A&&(S!==x||E!==v.userEmail)||!_&&(!!g||U),D=O?zs():Promise.resolve();_&&x?(yk(x,v.userEmail),p(x),f(!0),nr(!0)):(p(null),f(!1),nr(!1),!A&&O&&Lx()),D.then(async()=>{if(!_)return f(!1),nr(!1),await zs(),!1;nr(!0),await hk();return!0}).then(H=>{if(!H)return;const ee=z.getState();z.setState({appMode:"gallery"}),ee.setSettings(_r(A?g==null?void 0:g.settings:ee.settings))}).catch(H=>{console.warn("Failed to initialize image playground:",H),f(!1),nr(!1),zs()}).finally(()=>i(!0))},[]),b.useEffect(()=>{if(!s||m===null||!d)return;let g=!1;const v=()=>{const S=_o();S.userId===m&&S.userEmail===Bx()||g||(g=!0,f(!1),nr(!1),Ab(),zs(),Lx())},x=S=>{(S.key===null||S.key==="auth_user")&&v()};window.addEventListener("storage",x);const y=window.setInterval(v,2e3);return v(),()=>{window.removeEventListener("storage",x),window.clearInterval(y)}},[s,m,d]),b.useEffect(()=>{const g=v=>{var x;(x=v.target)!=null&&x.closest("img")&&v.preventDefault()};return document.addEventListener("dragstart",g),()=>document.removeEventListener("dragstart",g)},[]),s?d?o.jsxs(o.Fragment,{children:[o.jsx(Yk,{}),o.jsx("main",{"data-home-main":!0,"data-drag-select-surface":!0,className:"pb-48",children:o.jsxs("div",{className:"safe-area-x max-w-7xl mx-auto",children:[o.jsx("div",{"data-image-cleanup-banner":!0,role:"status",className:"mb-3 flex items-center border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-900 dark:border-amber-400/20 dark:bg-amber-400/10 dark:text-amber-100",children:"云端图片每天 00:00、03:00、06:00、09:00、12:00、15:00、18:00、21:00 自动清理，请及时下载。"}),o.jsx(Xk,{}),a&&!l?o.jsx(rj,{}):o.jsx(Gk,{})]})}),o.jsx(SS,{}),o.jsx(MS,{}),o.jsx(DS,{}),o.jsx(OS,{}),o.jsx(zS,{}),o.jsx(QS,{}),o.jsx(lj,{}),o.jsx(sj,{}),o.jsx(LS,{}),o.jsx(KS,{}),o.jsx(VS,{})]}):o.jsx("main",{className:"flex min-h-screen items-center justify-center p-6 text-center text-sm text-gray-500 dark:text-gray-400",children:"登录身份已失效，请返回主应用重新进入生图工作台。"}):o.jsx("main",{className:"min-h-screen","aria-busy":"true"})}`

// Logout can remove only auth_token before auth_user is cleared. Keep the
// standalone page bound to a live session while still allowing token rotation.
const auditedReplacement = replacement.replace(
  'S.userId===m&&S.userEmail===Bx()',
  'S.tokenPresent&&S.userId===m&&S.userEmail===Bx()',
)

writeFileSync(bundlePath, source.slice(0, start) + auditedReplacement + source.slice(end))

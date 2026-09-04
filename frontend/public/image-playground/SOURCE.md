# 小杂鱼の生图

This directory contains the production build of [CookSleep/gpt_image_playground](https://github.com/CookSleep/gpt_image_playground), embedded under the MIT license.

- Source commit: `dca69294c0ae17cb4d1c2d17722b6a013a61971b`
- Integration provider: `sb2api-async` (displayed as `OpenAI 兼容接口`)
- Build flags: `VITE_SHOW_PRESET_CONFIG_ONLY=true` and the managed gallery-only build.
- The launcher supplies profiles for active, user-owned API keys whose groups allow image generation. Profile names are the API Key creation names.
- The standalone settings screen has no Agent or habitual-preference tabs, no profile creation/deletion, and no editable provider, URL, or key fields. Browser requests use the current Sub2API origin's `/v1` endpoint, which is the zayu `/v1` endpoint when the app is deployed on the zayu hostname.
- Model refresh requests use only the currently selected key and the same-origin `/v1` endpoint. Generation uses the synchronous `images/generations` endpoint so the workbench does not depend on optional async object storage.
- The workbench toolbar and settings dialog use the same persisted `profiles` object. The profile selector changes the active API Key, and focusing the model selector refreshes the current profile's image-capable models.
- Moderation controls and request fields are removed; generation requests rely on the provider's default moderation behavior.
- When a provider image URL does not allow CORS blob fetching, the result keeps the original URL as its browser image source so a successful generation is not reported as `load failed`.
- Credentials are carried once through same-origin `sessionStorage`, consumed at startup, and never placed in the URL or `window.name`. Settings remain browser-local; no cloud configuration record is written.
- The Go static server is the authoritative feature gate for the HTML and every asset. The standalone runtime binds browser data to the bootstrap user ID/email and observes `auth_user` plus token presence; normal access/refresh token rotation does not invalidate that identity.
- `frontend/tools/patch-image-playground-session.mjs` reapplies the audited session component patch if this vendored bundle is refreshed from upstream.

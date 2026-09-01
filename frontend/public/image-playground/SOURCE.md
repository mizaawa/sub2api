# GPT Image Playground

This directory contains the production build of [CookSleep/gpt_image_playground](https://github.com/CookSleep/gpt_image_playground), embedded under the MIT license.

- Source commit: `dca69294c0ae17cb4d1c2d17722b6a013a61971b`
- Integration provider: `sb2api-async`
- Build flags: `VITE_SHOW_PRESET_CONFIG_ONLY=true`
- The launcher supplies one preconfigured profile per available image-enabled Sub2API group through the standalone window name. Each profile contains the group's active user key, gateway Base URL, and the model catalogue returned by `/v1/models`.
- The standalone settings screen exposes group and model selection while hiding manual Provider, Base URL, and API Key entry. Generation still uses `images/generations/async`, so Sub2API remains authoritative for balance checks, quota checks, routing, and billing.
- Credentials are not placed in the URL; the app clears `window.name` after importing it.

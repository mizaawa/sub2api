<template>
  <div v-if="hasHomeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else-if="compactHomeEnabled"
    data-testid="compact-home"
    class="compact-home-page flex min-h-screen flex-col"
  >
    <header class="compact-header px-4 py-4 sm:px-6">
      <nav class="mx-auto flex max-w-5xl flex-wrap items-center justify-between gap-3 sm:gap-4">
        <div class="flex min-w-0 flex-1 items-center gap-3">
          <img
            :src="siteLogo || '/logo.svg'"
            alt="Logo"
            class="h-10 w-10 shrink-0 rounded-2xl object-contain"
          />
          <span class="min-w-0 truncate text-base font-semibold">{{ siteName }}</span>
        </div>
        <div class="flex max-w-full shrink-0 flex-wrap items-center justify-end gap-2">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="compact-icon-button"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <button
            class="compact-icon-button"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="compact-primary-button"
          >
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main class="flex min-w-0 flex-1 items-center justify-center px-4 py-16 sm:px-6">
      <div class="min-w-0 max-w-2xl text-center">
        <img
          :src="siteLogo || '/logo.svg'"
          alt="Logo"
          class="mx-auto mb-7 h-20 w-20 rounded-3xl object-contain"
        />
        <h1 class="[overflow-wrap:anywhere] text-4xl font-semibold md:text-5xl">{{ siteName }}</h1>
        <p class="mt-5 whitespace-pre-wrap [overflow-wrap:anywhere] text-base text-neutral-600 dark:text-dark-300">
          {{ siteSubtitle }}
        </p>
        <router-link
          :to="isAuthenticated ? dashboardPath : '/login'"
          class="compact-primary-button mt-8 px-6 py-3"
        >
          {{ isAuthenticated ? t('home.goToDashboard') : t('home.login') }}
        </router-link>
      </div>
    </main>

    <footer class="compact-footer px-4 py-5 text-center text-sm sm:px-6">
      &copy; {{ currentYear }} {{ siteName }}
    </footer>
  </div>

  <div v-else class="landing-page">
    <header class="landing-header">
      <nav class="landing-nav">
        <div class="landing-brand">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" />
          <span>{{ siteName }}</span>
        </div>

        <div class="landing-actions">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="landing-icon-button"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <button
            class="landing-icon-button"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            @click="toggleTheme"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link
            :to="isAuthenticated ? dashboardPath : '/login'"
            class="landing-login"
          >
            <span v-if="isAuthenticated" class="landing-avatar">{{ userInitial }}</span>
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main>
      <section class="landing-hero">
        <div class="hero-content">
          <p class="hero-kicker">{{ siteSubtitle }}</p>
          <h1>{{ siteName }}</h1>
          <p class="hero-description">{{ t('home.providers.description') }}</p>
          <div class="hero-actions">
            <router-link
              :to="isAuthenticated ? dashboardPath : '/login'"
              class="hero-primary-action"
            >
              {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
              <Icon name="arrowRight" size="md" :stroke-width="2" />
            </router-link>
            <a
              v-if="docUrl"
              :href="docUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="hero-secondary-action"
            >
              {{ t('home.docs') }}
            </a>
          </div>
          <div class="hero-capabilities" aria-label="Core capabilities">
            <span>{{ t('home.tags.subscriptionToApi') }}</span>
            <span>{{ t('home.tags.stickySession') }}</span>
            <span>{{ t('home.tags.realtimeBilling') }}</span>
          </div>
        </div>
      </section>

      <section class="product-stage">
        <div class="product-stage-inner">
          <div class="product-stage-heading">
            <p>{{ t('home.features.unifiedGateway') }}</p>
            <h2>{{ t('home.features.unifiedGatewayDesc') }}</h2>
          </div>

          <div class="terminal-container api-console" aria-label="API request preview">
            <div class="api-console-bar">
              <div class="api-console-dots" aria-hidden="true">
                <span></span><span></span><span></span>
              </div>
              <span>API Console</span>
              <span class="api-console-status">Live</span>
            </div>
            <div class="api-console-body">
              <div class="console-line console-line-1">
                <span class="console-prompt">$</span>
                <span class="console-command">curl</span>
                <span class="console-option">-X POST</span>
                <span class="console-path">/v1/messages</span>
              </div>
              <div class="console-line console-line-2">
                <span class="console-note"># Routing to the best available upstream</span>
              </div>
              <div class="console-line console-line-3">
                <span class="console-success">200 OK</span>
                <span class="console-response">{ "content": "Hello!" }</span>
              </div>
              <div class="console-line console-line-4">
                <span class="console-prompt">$</span>
                <span class="console-cursor"></span>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="capabilities-section">
        <div class="section-heading">
          <p>{{ t('home.tags.subscriptionToApi') }}</p>
          <h2>{{ t('home.features.multiAccount') }}</h2>
        </div>

        <div class="capabilities-grid">
          <article>
            <div class="capability-icon capability-icon-coral">
              <Icon name="server" size="lg" />
            </div>
            <h3>{{ t('home.features.unifiedGateway') }}</h3>
            <p>{{ t('home.features.unifiedGatewayDesc') }}</p>
          </article>
          <article>
            <div class="capability-icon capability-icon-green">
              <Icon name="shield" size="lg" />
            </div>
            <h3>{{ t('home.features.multiAccount') }}</h3>
            <p>{{ t('home.features.multiAccountDesc') }}</p>
          </article>
          <article>
            <div class="capability-icon capability-icon-blue">
              <Icon name="chart" size="lg" />
            </div>
            <h3>{{ t('home.features.balanceQuota') }}</h3>
            <p>{{ t('home.features.balanceQuotaDesc') }}</p>
          </article>
        </div>
      </section>

      <section class="providers-section">
        <div class="providers-inner">
          <div class="section-heading providers-heading">
            <p>{{ t('home.providers.title') }}</p>
            <h2>{{ t('home.providers.description') }}</h2>
          </div>
          <div class="provider-list">
            <div class="provider-item">
              <span class="provider-mark provider-mark-coral">C</span>
              <span>{{ t('home.providers.claude') }}</span>
              <small>{{ t('home.providers.supported') }}</small>
            </div>
            <div class="provider-item">
              <span class="provider-mark provider-mark-green">G</span>
              <span>GPT</span>
              <small>{{ t('home.providers.supported') }}</small>
            </div>
            <div class="provider-item">
              <span class="provider-mark provider-mark-blue">G</span>
              <span>{{ t('home.providers.gemini') }}</span>
              <small>{{ t('home.providers.supported') }}</small>
            </div>
            <div class="provider-item">
              <span class="provider-mark provider-mark-violet">A</span>
              <span>{{ t('home.providers.antigravity') }}</span>
              <small>{{ t('home.providers.supported') }}</small>
            </div>
            <div class="provider-item provider-item-muted">
              <span class="provider-mark">+</span>
              <span>{{ t('home.providers.more') }}</span>
              <small>{{ t('home.providers.soon') }}</small>
            </div>
          </div>
        </div>
      </section>
    </main>

    <footer class="landing-footer">
      <div>
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <nav>
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">
            {{ t('home.docs') }}
          </a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a>
        </nav>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(
  appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '',
  { allowRelative: true, allowDataUrl: true },
))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const hasHomeContent = computed(() => homeContent.value.trim().length > 0)
const compactHomeEnabled = computed(() => appStore.cachedPublicSettings?.compact_home_enabled === true)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const githubUrl = 'https://github.com/mizaawa/sub2api'
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => authStore.user?.email?.charAt(0).toUpperCase() || '')
const currentYear = computed(() => new Date().getFullYear())

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.compact-home-page,
.landing-page {
  min-height: 100vh;
  background: var(--md-sys-color-background);
  color: var(--md-sys-color-on-surface);
}

.compact-header,
.compact-footer {
  border-color: var(--md-sys-color-outline);
}

.compact-header { border-bottom-width: 1px; }
.compact-footer { border-top-width: 1px; color: var(--md-sys-color-on-surface-variant); }

.compact-icon-button,
.landing-icon-button {
  display: inline-flex;
  width: 2.75rem;
  height: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  color: var(--md-sys-color-on-surface-variant);
  transition: background-color 180ms ease, color 180ms ease;
}

.compact-icon-button:hover,
.landing-icon-button:hover { background: var(--md-sys-color-surface-container); color: var(--md-sys-color-on-surface); }

.compact-primary-button,
.landing-login,
.hero-primary-action {
  display: inline-flex;
  min-height: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  background: var(--md-sys-color-on-surface);
  color: var(--md-sys-color-background);
  font-size: 0.875rem;
  font-weight: 600;
  transition: opacity 180ms ease, transform 180ms ease;
}

.compact-primary-button { padding: 0.625rem 1.125rem; }
.compact-primary-button:hover,
.landing-login:hover,
.hero-primary-action:hover { opacity: 0.86; }

.landing-header {
  position: relative;
  z-index: 20;
  border-bottom: 1px solid var(--md-sys-color-outline);
  background: var(--md-sys-color-background);
}

.landing-nav {
  display: flex;
  width: min(100% - 3rem, 76rem);
  min-height: 4.75rem;
  margin: 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.landing-brand { display: flex; min-width: 0; align-items: center; gap: 0.75rem; font-weight: 650; }
.landing-brand img { width: 2.5rem; height: 2.5rem; border-radius: 0.875rem; object-fit: contain; }
.landing-brand span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.landing-actions { display: flex; align-items: center; gap: 0.375rem; }
.landing-login { gap: 0.5rem; margin-left: 0.375rem; padding: 0.625rem 1.125rem; }
.landing-avatar { display: grid; width: 1.5rem; height: 1.5rem; place-items: center; border-radius: 50%; background: var(--md-sys-color-primary); color: var(--md-sys-color-on-primary); font-size: 0.6875rem; }

.landing-hero {
  display: flex;
  min-height: min(38rem, calc(100svh - 9.5rem));
  align-items: center;
  justify-content: center;
  padding: 5rem 1.5rem 6.5rem;
  text-align: center;
}

.hero-content { width: min(100%, 58rem); animation: hero-enter 700ms cubic-bezier(0.2, 0, 0, 1) both; }
.hero-kicker { margin-bottom: 1.5rem; color: var(--md-sys-color-primary); font-size: 1rem; font-weight: 650; }
.hero-content h1 { margin: 0; font-size: 6rem; line-height: 0.98; font-weight: 650; letter-spacing: 0; overflow-wrap: anywhere; }
.hero-description { max-width: 42rem; margin: 1.75rem auto 0; color: var(--md-sys-color-on-surface-variant); font-size: 1.25rem; line-height: 1.65; }
.hero-actions { display: flex; flex-wrap: wrap; align-items: center; justify-content: center; gap: 0.75rem; margin-top: 2.25rem; }
.hero-primary-action { min-height: 3.25rem; gap: 0.625rem; padding: 0.75rem 1.5rem; border-radius: 1.125rem; background: var(--md-sys-color-primary); color: var(--md-sys-color-on-primary); font-size: 1rem; }
.hero-primary-action:hover { transform: translateY(-2px); }
.hero-secondary-action { display: inline-flex; min-height: 3.25rem; align-items: center; padding: 0.75rem 1.5rem; border: 1px solid var(--md-sys-color-outline); border-radius: 1.125rem; font-weight: 600; transition: background-color 180ms ease; }
.hero-secondary-action:hover { background: var(--md-sys-color-surface-container); }
.hero-capabilities { display: flex; flex-wrap: wrap; justify-content: center; gap: 0; margin-top: 2.5rem; color: var(--md-sys-color-on-surface-variant); font-size: 0.875rem; }
.hero-capabilities span { display: inline-flex; align-items: center; }
.hero-capabilities span + span::before { width: 3px; height: 3px; margin: 0 0.875rem; border-radius: 50%; background: var(--md-sys-color-outline); content: ''; }

.product-stage { padding: 6.5rem 1.5rem 7rem; background: #191918; color: #f7f6f2; }
.product-stage-inner { width: min(100%, 70rem); margin: 0 auto; }
.product-stage-heading { max-width: 48rem; margin: 0 auto 3rem; text-align: center; }
.product-stage-heading p,
.section-heading > p { margin-bottom: 1rem; color: #e38361; font-size: 0.875rem; font-weight: 700; }
.product-stage-heading h2 { margin: 0; color: #f7f6f2; font-size: 2.5rem; line-height: 1.2; font-weight: 550; letter-spacing: 0; }

.api-console { overflow: hidden; border: 1px solid #4b4a46; border-radius: 2rem; background: #232321; animation: stage-enter 700ms 140ms cubic-bezier(0.2, 0, 0, 1) both; }
.api-console-bar { display: grid; grid-template-columns: 1fr auto 1fr; min-height: 4rem; align-items: center; padding: 0 1.5rem; border-bottom: 1px solid #4b4a46; color: #aaa79f; font-size: 0.8125rem; }
.api-console-dots { display: flex; gap: 0.5rem; }
.api-console-dots span { width: 0.75rem; height: 0.75rem; border-radius: 50%; }
.api-console-dots span:nth-child(1) { background: #e56b61; }
.api-console-dots span:nth-child(2) { background: #e1ae4b; }
.api-console-dots span:nth-child(3) { background: #5ca970; }
.api-console-status { justify-self: end; color: #8fd09f; }
.api-console-status::before { display: inline-block; width: 0.5rem; height: 0.5rem; margin-right: 0.5rem; border-radius: 50%; background: #5ca970; content: ''; }
.api-console-body { min-height: 23rem; padding: 4rem clamp(1.5rem, 8%, 6rem); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 1.125rem; line-height: 2.4; }
.console-line { display: flex; flex-wrap: wrap; align-items: center; gap: 0.75rem; opacity: 0; animation: line-enter 420ms ease forwards; }
.console-line-1 { animation-delay: 500ms; }
.console-line-2 { animation-delay: 850ms; }
.console-line-3 { animation-delay: 1200ms; }
.console-line-4 { animation-delay: 1550ms; }
.console-prompt { color: #8fd09f; font-weight: 700; }
.console-command { color: #efa98d; }
.console-option { color: #9eb8e4; }
.console-path { color: #f4d35e; }
.console-note { color: #85827b; font-style: italic; }
.console-success { padding: 0 0.625rem; border-radius: 0.625rem; background: #22482c; color: #9ddeab; font-weight: 700; }
.console-response { color: #e5c07b; }
.console-cursor { width: 0.625rem; height: 1.25rem; background: #8fd09f; animation: cursor-blink 1s step-end infinite; }

.capabilities-section { padding: 7rem 1.5rem; background: var(--md-sys-color-surface); }
.section-heading { width: min(100%, 46rem); margin: 0 auto 4rem; text-align: center; }
.section-heading h2 { margin: 0; color: var(--md-sys-color-on-surface); font-size: 2.75rem; line-height: 1.15; font-weight: 600; letter-spacing: 0; }
.capabilities-grid { display: grid; width: min(100%, 72rem); margin: 0 auto; grid-template-columns: repeat(3, minmax(0, 1fr)); }
.capabilities-grid article { padding: 0 2.5rem; }
.capabilities-grid article + article { border-left: 1px solid var(--md-sys-color-outline); }
.capability-icon { display: grid; width: 3.5rem; height: 3.5rem; margin-bottom: 1.75rem; place-items: center; border-radius: 1.25rem; }
.capability-icon-coral { background: #fce9de; color: #a54a33; }
.capability-icon-green { background: #e5f2e8; color: #36734a; }
.capability-icon-blue { background: #e8eef9; color: #365d9d; }
:global(.dark) .capability-icon-coral { background: #4a2a21; color: #efa98d; }
:global(.dark) .capability-icon-green { background: #213a29; color: #8fd09f; }
:global(.dark) .capability-icon-blue { background: #24334c; color: #9eb8e4; }
.capabilities-grid h3 { margin-bottom: 0.75rem; font-size: 1.25rem; font-weight: 650; }
.capabilities-grid p { color: var(--md-sys-color-on-surface-variant); line-height: 1.75; }

.providers-section { padding: 7rem 1.5rem; background: #eef1f4; }
:global(.dark) .providers-section { background: #1c2024; }
.providers-inner { width: min(100%, 72rem); margin: 0 auto; }
.providers-heading { margin-bottom: 3rem; }
.provider-list { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 0.75rem; }
.provider-item { display: grid; min-height: 11rem; place-items: center; align-content: center; gap: 0.75rem; padding: 1.25rem; border: 1px solid transparent; border-radius: 1.75rem; background: var(--md-sys-color-surface); font-weight: 650; transition: border-color 180ms ease, transform 180ms ease; }
.provider-item:hover { border-color: #b7bdc5; transform: translateY(-3px); }
.provider-mark { display: grid; width: 3rem; height: 3rem; place-items: center; border-radius: 1rem; background: var(--md-sys-color-surface-container); font-weight: 750; }
.provider-mark-coral { background: #fce9de; color: #a54a33; }
.provider-mark-green { background: #e5f2e8; color: #36734a; }
.provider-mark-blue { background: #e8eef9; color: #365d9d; }
.provider-mark-violet { background: #eee8f7; color: #6b4d8f; }
.provider-item small { color: #3f7550; font-size: 0.6875rem; font-weight: 700; }
.provider-item-muted { opacity: 0.62; }

.landing-footer { padding: 2.5rem 1.5rem; border-top: 1px solid var(--md-sys-color-outline); background: var(--md-sys-color-background); color: var(--md-sys-color-on-surface-variant); }
.landing-footer > div { display: flex; width: min(100%, 72rem); margin: 0 auto; align-items: center; justify-content: space-between; gap: 1rem; font-size: 0.875rem; }
.landing-footer nav { display: flex; gap: 1.5rem; }
.landing-footer a:hover { color: var(--md-sys-color-on-surface); }

@keyframes hero-enter { from { opacity: 0; transform: translateY(1.25rem); } to { opacity: 1; transform: translateY(0); } }
@keyframes stage-enter { from { opacity: 0; transform: translateY(1.5rem); } to { opacity: 1; transform: translateY(0); } }
@keyframes line-enter { from { opacity: 0; transform: translateY(0.5rem); } to { opacity: 1; transform: translateY(0); } }
@keyframes cursor-blink { 0%, 50% { opacity: 1; } 51%, 100% { opacity: 0; } }

@media (max-width: 900px) {
  .hero-content h1 { font-size: 4.5rem; }
  .capabilities-grid { grid-template-columns: 1fr; gap: 2.5rem; }
  .capabilities-grid article { padding: 0; }
  .capabilities-grid article + article { padding-top: 2.5rem; border-top: 1px solid var(--md-sys-color-outline); border-left: 0; }
  .provider-list { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .provider-item:last-child { grid-column: 1 / -1; }
}

@media (max-width: 640px) {
  .landing-nav { width: min(100% - 2rem, 76rem); min-height: 4.25rem; }
  .landing-brand span { display: none; }
  .landing-actions { gap: 0; }
  .landing-actions :deep(.relative button) { padding-inline: 0.5rem; }
  .landing-login { min-height: 2.5rem; padding: 0.5rem 0.875rem; }
  .landing-icon-button { width: 2.5rem; height: 2.5rem; }
  .landing-hero { min-height: 39rem; padding: 4.5rem 1rem 5rem; }
  .hero-kicker { margin-bottom: 1.25rem; }
  .hero-content h1 { font-size: 3.5rem; line-height: 1.02; }
  .hero-description { margin-top: 1.5rem; font-size: 1rem; }
  .hero-capabilities { flex-direction: column; gap: 0.5rem; }
  .hero-capabilities span { justify-content: center; }
  .hero-capabilities span + span::before { display: none; }
  .product-stage { padding: 5rem 1rem; }
  .product-stage-heading h2,
  .section-heading h2 { font-size: 2rem; }
  .api-console { border-radius: 1.5rem; }
  .api-console-bar { padding: 0 1rem; }
  .api-console-body { min-height: 18rem; padding: 2.5rem 1rem; font-size: 0.8125rem; line-height: 2.15; }
  .console-line { gap: 0.45rem; }
  .capabilities-section,
  .providers-section { padding: 5rem 1rem; }
  .section-heading { margin-bottom: 3rem; }
  .provider-list { grid-template-columns: 1fr; }
  .provider-item:last-child { grid-column: auto; }
  .provider-item { min-height: 8.5rem; }
  .landing-footer > div { flex-direction: column; text-align: center; }
}

@media (prefers-reduced-motion: reduce) {
  .hero-content,
  .api-console,
  .console-line { animation: none; opacity: 1; transform: none; }
  .console-cursor { animation: none; }
  .hero-primary-action:hover,
  .provider-item:hover { transform: none; }
}
</style>

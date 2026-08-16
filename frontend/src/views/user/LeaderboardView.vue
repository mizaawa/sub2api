<template>
  <AppLayout>
    <div class="leaderboard-page mx-auto w-full max-w-6xl space-y-5">
      <section class="leaderboard-hero overflow-hidden rounded-3xl border">
        <div class="hero-content flex flex-wrap items-end justify-between gap-5 px-5 py-6 sm:px-7 sm:py-7">
          <div class="min-w-0">
            <p class="hero-period text-xs font-semibold uppercase tracking-[0.12em]">{{ periodLabel }}</p>
            <h1 class="mt-1 text-2xl font-bold tracking-tight sm:text-3xl">{{ t('leaderboard.title') }}</h1>
            <p class="hero-description mt-2 max-w-xl text-sm leading-6">{{ t('leaderboard.description') }}</p>
          </div>
          <div class="rank-summary shrink-0 rounded-2xl border px-4 py-3 text-right">
            <p class="hero-period text-xs">{{ t('leaderboard.myRank') }}</p>
            <p class="mt-1 text-2xl font-bold tabular-nums">{{ myRankLabel }}</p>
          </div>
        </div>
      </section>

      <section class="leaderboard-card overflow-hidden rounded-3xl border">
        <div class="card-toolbar flex flex-wrap items-center justify-between gap-3 border-b px-4 py-4 sm:px-6">
          <div class="min-w-0">
            <h2 class="text-base font-semibold">{{ t('leaderboard.title') }}</h2>
            <p class="muted mt-1 text-xs">{{ t('leaderboard.periodTimezone') }}</p>
          </div>
          <div class="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
            <div class="flex items-center gap-2">
              <div class="period-switcher flex min-w-0 flex-1 rounded-xl border p-1 sm:flex-none" role="tablist">
                <button
                  v-for="option in periodOptions"
                  :key="option.value"
                  type="button"
                  role="tab"
                  :aria-selected="selectedPeriod === option.value"
                  :disabled="loading"
                  class="period-option min-w-0 flex-1 rounded-lg px-2.5 py-1.5 text-xs font-semibold transition-colors sm:flex-none sm:px-3"
                  :class="selectedPeriod === option.value ? 'period-option-active' : 'period-option-idle'"
                  @click="selectPeriod(option.value)"
                >
                  {{ option.label }}
                </button>
              </div>
              <button class="btn btn-secondary btn-sm shrink-0" type="button" :disabled="loading" :aria-label="t('common.refresh')" @click="load">
                <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
                <span class="hidden sm:inline">{{ t('common.refresh') }}</span>
              </button>
            </div>
            <button
              type="button"
              class="participation-toggle w-full rounded-xl border px-3 py-2 text-xs font-semibold transition-colors sm:w-auto"
              :class="participating ? 'participation-enabled' : 'participation-disabled'"
              :disabled="loading || participationLoading"
              :aria-pressed="participating"
              @click="toggleParticipation"
            >
              {{ participating ? t('leaderboard.participating') : t('leaderboard.notParticipating') }}
            </button>
          </div>
        </div>

        <div v-if="error" class="px-5 py-10 text-center text-sm text-rose-600 dark:text-rose-300">{{ error }}</div>
        <div v-else-if="loading" class="muted px-5 py-10 text-center text-sm">{{ t('common.loading') }}</div>
        <div v-else-if="!entries.length" class="muted px-5 py-8 text-center text-sm">{{ t('leaderboard.noData') }}</div>
        <div v-else class="overflow-x-auto">
          <table class="leaderboard-table w-full table-fixed text-left text-sm">
            <colgroup>
              <col class="rank-column" />
              <col class="user-column" />
              <col class="metric-column" />
              <col class="metric-column optional-column" />
              <col class="metric-column optional-column" />
            </colgroup>
            <thead class="table-heading text-xs uppercase tracking-wide">
              <tr>
                <th class="px-3 py-3 font-semibold sm:px-5">{{ t('leaderboard.rank') }}</th>
                <th class="px-3 py-3 font-semibold sm:px-5">{{ t('leaderboard.user') }}</th>
                <th class="px-3 py-3 text-right font-semibold sm:px-5">{{ t('leaderboard.cost') }}</th>
                <th class="hidden px-3 py-3 text-right font-semibold sm:table-cell sm:px-5">{{ t('leaderboard.requests') }}</th>
                <th class="hidden px-3 py-3 text-right font-semibold sm:table-cell sm:px-5">{{ t('leaderboard.tokens') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="entry in entries" :key="entry.rank" class="table-row transition-colors">
                <td class="px-3 py-3 sm:px-5">
                  <span class="rank-badge inline-flex min-w-8 items-center justify-center rounded-lg border px-2 py-1 font-bold tabular-nums" :class="rankClass(entry.rank)">
                    {{ entry.rank }}
                  </span>
                </td>
                <td class="truncate px-3 py-3 font-medium sm:px-5" :title="entry.display_name">{{ entry.display_name }}</td>
                <td class="px-3 py-3 text-right tabular-nums sm:px-5">${{ entry.actual_cost.toFixed(2) }}</td>
                <td class="hidden px-3 py-3 text-right tabular-nums sm:table-cell sm:px-5">{{ formatNumber(entry.requests) }}</td>
                <td class="hidden px-3 py-3 text-right tabular-nums sm:table-cell sm:px-5">{{ formatNumber(entry.tokens) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import leaderboardAPI, { type LeaderboardEntry, type LeaderboardResponse } from '@/api/leaderboard'

type LeaderboardPeriod = 'today' | 'week' | 'month'

const { t } = useI18n()
const entries = ref<LeaderboardEntry[]>([])
const data = ref<LeaderboardResponse | null>(null)
const loading = ref(false)
const error = ref('')
const selectedPeriod = ref<LeaderboardPeriod>('today')
const participating = ref(false)
const participationLoading = ref(false)

const periodOptions = computed(() => [
  { value: 'today' as const, label: t('leaderboard.periodToday') },
  { value: 'week' as const, label: t('leaderboard.periodWeek') },
  { value: 'month' as const, label: t('leaderboard.periodMonth') },
])

const periodLabel = computed(() => {
  const key = selectedPeriod.value === 'week' ? 'leaderboard.periodWeek' : selectedPeriod.value === 'month' ? 'leaderboard.periodMonth' : 'leaderboard.periodToday'
  return t(key)
})

const myRankLabel = computed(() => data.value?.my_rank ? `#${data.value.my_rank}` : t('leaderboard.unranked'))

function formatNumber(value: number): string {
  return new Intl.NumberFormat().format(value)
}

function rankClass(rank: number): string {
  if (rank === 1) return 'rank-gold'
  if (rank === 2) return 'rank-silver'
  if (rank === 3) return 'rank-bronze'
  return 'rank-regular'
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    data.value = await leaderboardAPI.get(selectedPeriod.value)
    participating.value = data.value.participating
    entries.value = data.value.entries.slice(0, 25)
  } catch {
    error.value = t('leaderboard.loadFailed')
  } finally {
    loading.value = false
  }
}

async function toggleParticipation(): Promise<void> {
  participationLoading.value = true
  error.value = ''
  try {
    const result = await leaderboardAPI.setParticipation(!participating.value)
    participating.value = result.participating
    await load()
  } catch {
    error.value = t('leaderboard.participationFailed')
  } finally {
    participationLoading.value = false
  }
}

function selectPeriod(period: LeaderboardPeriod): void {
  if (period === selectedPeriod.value || loading.value) return
  selectedPeriod.value = period
  void load()
}

onMounted(load)
</script>

<style scoped>
.leaderboard-hero,
.leaderboard-card {
  border-color: color-mix(in srgb, var(--md-sys-color-outline) 80%, transparent);
  background: var(--md-sys-color-surface);
  color: var(--md-sys-color-on-surface);
  box-shadow: 0 12px 30px color-mix(in srgb, var(--md-sys-color-primary) 9%, transparent);
}

.hero-content {
  background: color-mix(in srgb, var(--md-sys-color-primary) 12%, var(--md-sys-color-surface));
}

.hero-period { color: var(--md-sys-color-primary); }
.hero-description, .muted { color: var(--md-sys-color-on-surface-variant); }

.rank-summary {
  border-color: color-mix(in srgb, var(--md-sys-color-primary) 36%, var(--md-sys-color-outline));
  background: color-mix(in srgb, var(--md-sys-color-primary) 10%, var(--md-sys-color-surface));
}

.card-toolbar { border-color: color-mix(in srgb, var(--md-sys-color-outline) 70%, transparent); }
.period-switcher { border-color: var(--md-sys-color-outline); background: var(--md-sys-color-surface-container); }
.period-option-active { background: var(--md-sys-color-primary); color: var(--md-sys-color-on-primary); }
.period-option-idle { color: var(--md-sys-color-on-surface-variant); }
.period-option-idle:hover { background: color-mix(in srgb, var(--md-sys-color-primary) 12%, transparent); color: var(--md-sys-color-on-surface); }
/* 参与按钮：未参与 = 青绿色；已参与 = 适配主题的红色 */
.participation-toggle { white-space: nowrap; }
.participation-enabled {
  border-color: transparent;
  background: color-mix(in srgb, #e53935 85%, var(--md-sys-color-surface));
  color: #fff;
}
.participation-enabled:hover:not(:disabled) {
  background: color-mix(in srgb, #c62828 85%, var(--md-sys-color-surface));
}
.participation-disabled {
  border-color: transparent;
  background: #009688;
  color: #fff;
}
.participation-disabled:hover:not(:disabled) {
  background: #00796b;
}

.table-heading { background: var(--md-sys-color-surface-container); color: var(--md-sys-color-on-surface-variant); }
.table-row { border-top: 1px solid color-mix(in srgb, var(--md-sys-color-outline) 55%, transparent); }
.table-row:hover { background: color-mix(in srgb, var(--md-sys-color-primary) 8%, transparent); }
.rank-column { width: 4.25rem; }
.user-column { width: 34%; }
.metric-column { width: 20%; }
.rank-badge { border-color: transparent; }
.rank-regular { background: color-mix(in srgb, var(--md-sys-color-primary) 10%, transparent); color: var(--md-sys-color-primary); }

/* Miku-compatible metallic accents: restrained saturation keeps both themes readable. */
.rank-gold { border-color: #c7a34a; background: color-mix(in srgb, #e0bd63 26%, var(--md-sys-color-surface)); color: color-mix(in srgb, #9a7212 72%, var(--md-sys-color-on-surface)); }
.rank-silver { border-color: #93aeb0; background: color-mix(in srgb, #b9d2d3 25%, var(--md-sys-color-surface)); color: color-mix(in srgb, #587476 72%, var(--md-sys-color-on-surface)); }
.rank-bronze { border-color: #b98262; background: color-mix(in srgb, #c99572 25%, var(--md-sys-color-surface)); color: color-mix(in srgb, #87543c 72%, var(--md-sys-color-on-surface)); }

@media (max-width: 639px) {
  .leaderboard-page { max-width: 100%; }
  .hero-content { gap: 1rem; }
  .rank-summary { display: flex; width: 100%; align-items: center; justify-content: space-between; text-align: left; }
  .rank-column { width: 3.75rem; }
  .user-column { width: 40%; }
  .metric-column { width: 28%; }
  .leaderboard-table col.optional-column { display: none; }
}
</style>

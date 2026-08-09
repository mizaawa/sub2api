<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-6">
      <section class="card overflow-hidden">
        <div class="relative overflow-hidden bg-gradient-to-br from-teal-600 via-emerald-600 to-cyan-700 px-6 py-8 text-white md:px-8">
          <div class="relative flex flex-wrap items-end justify-between gap-5">
            <div>
              <p class="text-sm font-medium text-teal-100">{{ t('leaderboard.period') }}</p>
              <h1 class="mt-1 text-3xl font-bold tracking-tight">{{ t('leaderboard.title') }}</h1>
              <p class="mt-2 max-w-xl text-sm text-teal-50">{{ t('leaderboard.description') }}</p>
            </div>
            <div class="rounded-2xl border border-white/25 bg-white/10 px-4 py-3 text-right backdrop-blur-sm">
              <p class="text-xs text-teal-100">{{ t('leaderboard.myRank') }}</p>
              <p class="mt-1 text-2xl font-bold">{{ myRankLabel }}</p>
            </div>
          </div>
        </div>
      </section>

      <section class="card overflow-hidden">
        <div class="flex items-center justify-between gap-4 border-b border-gray-100 px-5 py-4 dark:border-dark-700 md:px-6">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('leaderboard.title') }}</h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('leaderboard.masked') }}</p>
          </div>
          <button class="btn btn-secondary btn-sm" type="button" :disabled="loading" @click="load">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            <span>{{ t('common.refresh') }}</span>
          </button>
        </div>

        <div v-if="error" class="px-6 py-10 text-center text-sm text-rose-600 dark:text-rose-300">{{ error }}</div>
        <div v-else-if="loading" class="px-6 py-10 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('common.loading') }}</div>
        <div v-else-if="!entries.length" class="px-6 py-12 text-center text-sm text-gray-500 dark:text-dark-400">{{ t('leaderboard.noData') }}</div>
        <div v-else class="overflow-x-auto">
          <table class="min-w-full text-left text-sm">
            <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-800/70 dark:text-dark-300">
              <tr>
                <th class="px-5 py-3 font-medium">{{ t('leaderboard.rank') }}</th>
                <th class="px-5 py-3 font-medium">{{ t('leaderboard.user') }}</th>
                <th class="px-5 py-3 text-right font-medium">{{ t('leaderboard.cost') }}</th>
                <th class="px-5 py-3 text-right font-medium">{{ t('leaderboard.requests') }}</th>
                <th class="px-5 py-3 text-right font-medium">{{ t('leaderboard.tokens') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="entry in entries" :key="entry.rank" class="transition-colors hover:bg-teal-50/60 dark:hover:bg-teal-900/15">
                <td class="px-5 py-3 font-semibold text-teal-700 dark:text-teal-300">{{ entry.rank }}</td>
                <td class="px-5 py-3 font-medium text-gray-800 dark:text-gray-100">{{ entry.display_name }}</td>
                <td class="px-5 py-3 text-right tabular-nums text-gray-700 dark:text-dark-200">${{ entry.actual_cost.toFixed(2) }}</td>
                <td class="px-5 py-3 text-right tabular-nums text-gray-600 dark:text-dark-300">{{ formatNumber(entry.requests) }}</td>
                <td class="px-5 py-3 text-right tabular-nums text-gray-600 dark:text-dark-300">{{ formatNumber(entry.tokens) }}</td>
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

const { t } = useI18n()
const entries = ref<LeaderboardEntry[]>([])
const data = ref<LeaderboardResponse | null>(null)
const loading = ref(false)
const error = ref('')

const myRankLabel = computed(() => data.value?.my_rank ? `#${data.value.my_rank}` : t('leaderboard.unranked'))

function formatNumber(value: number): string {
  return new Intl.NumberFormat().format(value)
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    data.value = await leaderboardAPI.get()
    entries.value = data.value.entries.slice(0, 25)
  } catch {
    error.value = t('leaderboard.loadFailed')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

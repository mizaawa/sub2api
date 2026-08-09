import { apiClient } from './client'

export interface LeaderboardEntry {
  rank: number
  display_name: string
  actual_cost: number
  requests: number
  tokens: number
}

export interface LeaderboardResponse {
  period: 'today' | 'week' | 'month'
  period_days: number
  entries: LeaderboardEntry[]
  my_rank: number | null
  my_actual_cost: number
  my_requests: number
  my_tokens: number
}

export async function getLeaderboard(period: 'today' | 'week' | 'month' = 'today', options?: { signal?: AbortSignal }): Promise<LeaderboardResponse> {
  const { data } = await apiClient.get<LeaderboardResponse>('/leaderboard', {
    params: { period },
    signal: options?.signal,
  })
  return data
}

export const leaderboardAPI = { get: getLeaderboard }
export default leaderboardAPI

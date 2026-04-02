<template>
  <AppLayout>
    <div class="mx-auto space-y-5 px-4 py-6">
      <!-- Page Header -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-amber-400 to-orange-500 shadow-lg">
            <svg class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 18.75h-9m9 0a3 3 0 013 3h-15a3 3 0 013-3m9 0v-3.375c0-.621-.503-1.125-1.125-1.125h-.871M7.5 18.75v-3.375c0-.621.504-1.125 1.125-1.125h.872m5.007 0H9.497m5.007 0a7.454 7.454 0 01-.982-3.172M9.497 14.25a7.454 7.454 0 00.981-3.172M5.25 4.236c-.982.143-1.954.317-2.916.52A6.003 6.003 0 007.73 9.728M5.25 4.236V4.5c0 2.108.966 3.99 2.48 5.228M5.25 4.236V2.721C7.456 2.41 9.71 2.25 12 2.25c2.291 0 4.545.16 6.75.47v1.516M18.75 4.236c.982.143 1.954.317 2.916.52A6.003 6.003 0 0016.27 9.728M18.75 4.236V4.5c0 2.108-.966 3.99-2.48 5.228m0 0a6.003 6.003 0 01-3.77 1.522m3.77-1.522a6.003 6.003 0 00-.34-6.478M9.73 9.728a6.003 6.003 0 003.77 1.522m-3.77-1.522a6.003 6.003 0 01.34-6.478m6.66 0a6.003 6.003 0 00-7 0" />
            </svg>
          </div>
          <div>
            <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('leaderboard.title') }}</h1>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.description') }}</p>
          </div>
        </div>

        <!-- Period Selector -->
        <div class="flex rounded-xl bg-gray-100 p-1.5 dark:bg-dark-800">
          <button
            v-for="p in periods"
            :key="p.value"
            @click="currentPeriod = p.value; fetchAllBoards()"
            class="rounded-lg px-4 py-2 text-sm font-medium transition-all duration-200"
            :class="currentPeriod === p.value
              ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-600 dark:text-white'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
          >
            {{ p.label }}
          </button>
        </div>
      </div>

      <!-- My Tier Card (only when logged in) -->
      <div v-if="isLoggedIn && bestRankData" class="card p-6">
        <div class="flex items-center gap-6">
          <!-- Tier icon -->
          <div class="flex flex-col items-center gap-1.5">
            <span class="text-6xl">{{ myTier.icon }}</span>
            <span class="text-base font-bold" :class="myTier.textClass">{{ t('leaderboard.tier.' + myTier.key) }}</span>
          </div>

          <div class="min-w-0 flex-1">
            <!-- Progress bar to next tier -->
            <div v-if="tierProgress" class="mb-4">
              <div class="mb-2 flex items-center justify-between">
                <span class="text-sm font-semibold text-gray-700 dark:text-gray-200">{{ t('leaderboard.tier.' + myTier.key) }}</span>
                <span class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ tierProgress.gapText }}</span>
              </div>
              <div class="h-3 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
                <div
                  class="h-full rounded-full transition-all duration-700 ease-out"
                  :class="myTier.barClass"
                  :style="{ width: tierProgress.percent + '%' }"
                ></div>
              </div>
            </div>

            <!-- All board ranks in one row -->
            <div class="flex flex-wrap gap-3">
              <div
                v-for="board in visibleBoards"
                :key="board.type"
                class="flex items-center gap-2 rounded-xl bg-gray-50 px-4 py-2.5 dark:bg-dark-700"
              >
                <span class="text-base">{{ board.emoji }}</span>
                <span class="text-sm text-gray-500 dark:text-gray-400">{{ board.label }}</span>
                <span class="text-lg font-bold text-gray-900 dark:text-white">#{{ getMyRankForType(board.type) || '—' }}</span>
                <span class="text-sm text-gray-400">{{ getMyValueForType(board.type) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loadingAll" class="flex h-64 items-center justify-center">
        <div class="flex flex-col items-center gap-3">
          <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
          <span class="text-sm text-gray-400">{{ t('leaderboard.loading') }}</span>
        </div>
      </div>

      <!-- 4 Board Cards Grid -->
      <div v-else class="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4">
        <div
          v-for="(board, boardIdx) in visibleBoards"
          :key="board.type"
          class="card overflow-hidden"
          :style="{ animationDelay: (boardIdx * 80) + 'ms' }"
          style="animation: fadeInUp 0.4s ease-out both"
        >
          <!-- Board Header -->
          <div class="flex items-center gap-2.5 border-b border-gray-100 px-5 py-4 dark:border-dark-700">
            <span class="text-xl">{{ board.emoji }}</span>
            <span class="text-base font-bold text-gray-900 dark:text-white">{{ board.label }}</span>
          </div>

          <!-- Board List -->
          <div>
            <!-- Empty state -->
            <div v-if="!boardDataMap[board.type] || boardDataMap[board.type]!.items.length === 0" class="px-5 py-12 text-center text-base text-gray-400">
              {{ t('leaderboard.noData') }}
            </div>

            <!-- Entries -->
            <div
              v-for="entry in getBoardItems(board.type)"
              :key="entry.rank"
              class="flex items-center gap-3 px-5 py-3.5 transition-colors hover:bg-gray-50 dark:hover:bg-dark-700/50"
              :class="entry.rank <= 3 ? 'border-l-[3px] ' + rankBorderClass(entry.rank) : 'border-l-[3px] border-transparent'"
            >
              <!-- Rank number -->
              <div class="flex w-9 flex-shrink-0 items-center justify-center">
                <span v-if="entry.rank <= 3" class="text-xl">{{ rankEmoji(entry.rank) }}</span>
                <span v-else class="text-xl font-bold text-gray-400 dark:text-gray-500">{{ entry.rank }}</span>
              </div>

              <!-- Email + tier badge -->
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-1.5">
                  <span class="text-sm">{{ getTierIcon(entry.rank, getBoardTotal(board.type)) }}</span>
                  <span class="truncate text-sm font-medium text-gray-800 dark:text-gray-200">{{ entry.masked_email }}</span>
                </div>
                <div v-if="entry.title" class="mt-0.5 text-xs text-gray-400">{{ entry.title }}</div>
              </div>

              <!-- Value -->
              <div class="flex-shrink-0 text-right">
                <span class="text-base font-bold text-gray-900 dark:text-white">{{ formatValue(entry.value, board.type) }}</span>
              </div>
            </div>
          </div>

          <!-- My rank in this board + motivation text (only when logged in) -->
          <div v-if="isLoggedIn && getMyRankInfo(board.type)" class="border-t border-gray-100 bg-gray-50/50 px-5 py-4 dark:border-dark-700 dark:bg-dark-800/80">
            <!-- My rank row -->
            <div class="flex items-center gap-2">
              <span class="text-sm text-gray-500 dark:text-gray-400">📍 {{ t('leaderboard.myRank') }}</span>
              <span class="text-lg font-bold text-primary-600 dark:text-primary-400">#{{ getMyRankInfo(board.type)!.rank }}</span>
              <span class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ formatValue(getMyRankInfo(board.type)!.value, board.type) }}</span>
            </div>
            <!-- Gap to next rank -->
            <div v-if="getMyRankInfo(board.type)!.next_rank_gap" class="mt-2 text-sm font-medium text-amber-600 dark:text-amber-400">
              💡 {{ getGapTextForBoard(board.type) }}
            </div>
            <!-- Gap to top 3 (if not already top 3) -->
            <div v-if="getTop3GapText(board.type)" class="mt-1.5 text-sm text-amber-500/80 dark:text-amber-400/70">
              🏆 {{ getTop3GapText(board.type) }}
            </div>
            <!-- Leading advantage (if #1) -->
            <div v-if="getMyRankInfo(board.type)!.rank === 1 && getBoardItems(board.type).length > 1" class="mt-1.5 text-sm font-medium text-emerald-600 dark:text-emerald-400">
              👑 {{ t('leaderboard.leading', { amount: formatValue(getMyRankInfo(board.type)!.value - getBoardItems(board.type)[1].value, board.type) }) }}
            </div>
          </div>
        </div>
      </div>

      <!-- Bottom Motivation Bar (logged in) -->
      <div
        v-if="isLoggedIn && bestGapText && !loadingAll"
        class="card sticky bottom-4 z-10 px-6 py-5"
      >
        <div class="flex items-center gap-3">
          <span class="text-xl">💡</span>
          <span class="text-base font-semibold text-amber-600 dark:text-amber-400">{{ bestGapText }}</span>
        </div>
      </div>

      <!-- Bottom CTA Bar (not logged in) -->
      <div
        v-if="!isLoggedIn && !loadingAll"
        class="card sticky bottom-4 z-10 px-6 py-5"
      >
        <div class="flex items-center justify-between gap-4">
          <div class="flex items-center gap-3">
            <span class="text-2xl">🔥</span>
            <div>
              <p class="text-base font-bold text-gray-900 dark:text-white">{{ t('leaderboard.ctaTitle') }}</p>
              <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('leaderboard.ctaDesc') }}</p>
            </div>
          </div>
          <button
            @click="$router.push('/login?redirect=/leaderboard')"
            class="flex-shrink-0 rounded-xl bg-gradient-to-r from-amber-500 to-orange-500 px-6 py-3 text-sm font-bold text-white shadow-lg transition-all hover:shadow-xl hover:brightness-110"
          >
            🏆 {{ t('leaderboard.ctaButton') }}
          </button>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { AppLayout } from '@/components/layout'
import { useAuthStore } from '@/stores/auth'
import { usageAPI } from '@/api/usage'
import type { LeaderboardType, LeaderboardPeriod, LeaderboardResponse } from '@/api/usage'

const { t } = useI18n()
const authStore = useAuthStore()
const isLoggedIn = computed(() => authStore.isAuthenticated)

// ======================== Types ========================

interface TierInfo {
  key: string
  icon: string
  textClass: string   // text color class for tier name
  barClass: string    // progress bar bg class
}

interface BoardConfig {
  type: LeaderboardType
  emoji: string
  label: string
}

// ======================== State ========================

const currentPeriod = ref<LeaderboardPeriod>('today')
const boardDataMap = ref<Record<string, LeaderboardResponse | null>>({})
const loadingAll = ref(false)

// ======================== Config ========================

// All board definitions (requests commented out — too similar to Token)
const allBoards: BoardConfig[] = [
  { type: 'cost', emoji: '💰', label: '' },
  { type: 'recharge', emoji: '💎', label: '' },
  { type: 'tokens', emoji: '🔥', label: '' },
  // { type: 'requests', emoji: '⚡', label: '' },  // 请求榜已隐藏，与Token榜重叠
  { type: 'active_days', emoji: '💪', label: '' },
]

const visibleBoards = computed<BoardConfig[]>(() =>
  allBoards.map(b => ({
    ...b,
    label: t('leaderboard.type' + capitalize(b.type)),
  }))
)

const periods = computed(() => [
  { value: 'today' as LeaderboardPeriod, label: t('leaderboard.periodToday') },
  { value: 'week' as LeaderboardPeriod, label: t('leaderboard.periodWeek') },
  { value: 'month' as LeaderboardPeriod, label: t('leaderboard.periodMonth') },
  { value: 'all' as LeaderboardPeriod, label: t('leaderboard.periodAll') },
])

// ======================== Tier System ========================
// 段位 = 排名百分比, 显示实际金额 (混合制)

const TIERS: TierInfo[] = [
  { key: 'champion', icon: '👑', textClass: 'text-red-500', barClass: 'bg-red-500' },
  { key: 'diamond', icon: '💎', textClass: 'text-cyan-500', barClass: 'bg-cyan-500' },
  { key: 'platinum', icon: '💠', textClass: 'text-gray-500 dark:text-gray-300', barClass: 'bg-gray-400' },
  { key: 'gold', icon: '⭐', textClass: 'text-amber-500', barClass: 'bg-amber-500' },
  { key: 'silver', icon: '⚪', textClass: 'text-gray-400', barClass: 'bg-gray-300' },
  { key: 'bronze', icon: '🔰', textClass: 'text-orange-400', barClass: 'bg-orange-400' },
]

function getTier(rank: number, total: number): TierInfo {
  if (total <= 0 || rank <= 0) return TIERS[5] // bronze
  if (rank === 1) return TIERS[0] // champion
  if (rank <= 3) return TIERS[1] // diamond
  if (rank <= Math.ceil(total * 0.1)) return TIERS[2] // platinum
  if (rank <= Math.ceil(total * 0.3)) return TIERS[3] // gold
  if (rank <= Math.ceil(total * 0.6)) return TIERS[4] // silver
  return TIERS[5] // bronze
}

function getTierIcon(rank: number, total: number): string {
  return getTier(rank, total).icon
}

// User's best tier across all visible boards
const bestRankData = computed(() => {
  let bestRank = Infinity
  let bestTotal = 0

  for (const board of allBoards) {
    const d = boardDataMap.value[board.type]
    if (!d?.my_rank || d.my_rank.rank <= 0) continue
    if (d.my_rank.rank < bestRank) {
      bestRank = d.my_rank.rank
      bestTotal = d.items.length
    }
  }

  if (bestRank === Infinity) return null
  return { rank: bestRank, total: bestTotal }
})

const myTier = computed(() => {
  if (!bestRankData.value) return TIERS[5]
  return getTier(bestRankData.value.rank, bestRankData.value.total)
})

// Progress to next tier
const tierProgress = computed(() => {
  const best = bestRankData.value
  if (!best || best.total <= 1) return null

  const rank = best.rank
  const total = best.total

  const boundaries = [
    { tier: 'champion', maxRank: 1 },
    { tier: 'diamond', maxRank: 3 },
    { tier: 'platinum', maxRank: Math.ceil(total * 0.1) },
    { tier: 'gold', maxRank: Math.ceil(total * 0.3) },
    { tier: 'silver', maxRank: Math.ceil(total * 0.6) },
    { tier: 'bronze', maxRank: total },
  ]

  let currentIdx = boundaries.length - 1
  for (let i = 0; i < boundaries.length; i++) {
    if (rank <= boundaries[i].maxRank) {
      currentIdx = i
      break
    }
  }

  // Already champion
  if (currentIdx === 0) return { percent: 100, gapText: t('leaderboard.tierMax') }

  const nextBound = boundaries[currentIdx - 1].maxRank
  const currBound = boundaries[currentIdx].maxRank
  const ranksInTier = currBound - nextBound
  const ranksAboveNext = rank - nextBound

  const percent = ranksInTier > 0 ? Math.round(((ranksInTier - ranksAboveNext) / ranksInTier) * 100) : 0
  const gap = ranksAboveNext
  const nextTierName = t('leaderboard.tier.' + boundaries[currentIdx - 1].tier)
  const gapText = t('leaderboard.tierGap', { count: gap, tier: nextTierName })

  return { percent: Math.max(5, Math.min(percent, 100)), gapText }
})

// ======================== Computed ========================

function getMyRankForType(type: LeaderboardType): number | null {
  const d = boardDataMap.value[type]
  if (!d?.my_rank || d.my_rank.rank <= 0) return null
  return d.my_rank.rank
}

function getMyValueForType(type: LeaderboardType): string {
  const d = boardDataMap.value[type]
  if (!d?.my_rank || d.my_rank.rank <= 0) return ''
  return formatValue(d.my_rank.value, type)
}

function getMyRankInfo(type: LeaderboardType) {
  const d = boardDataMap.value[type]
  if (!d?.my_rank || d.my_rank.rank <= 0) return null
  return d.my_rank
}

function getGapTextForBoard(type: LeaderboardType): string {
  const info = getMyRankInfo(type)
  if (!info || info.rank <= 1 || !info.next_rank_gap) return ''

  const actionMap: Record<string, string> = {
    cost: t('leaderboard.gapCost'),
    recharge: t('leaderboard.gapRecharge'),
    tokens: t('leaderboard.gapTokens'),
    active_days: t('leaderboard.gapActiveDays'),
  }
  const action = actionMap[type] || ''
  const amount = formatValue(info.next_rank_gap, type)
  const targetRank = info.rank - 1
  const email = info.next_rank_email || ''

  return t('leaderboard.gapNext', { action, amount, rank: targetRank, email })
}

// Calculate gap to reach top 3 for boards where user is ranked > 3
function getTop3GapText(type: LeaderboardType): string {
  const info = getMyRankInfo(type)
  if (!info || info.rank <= 3) return '' // Already top 3 or not ranked

  const items = boardDataMap.value[type]?.items
  if (!items || items.length < 3) return ''

  const thirdPlaceValue = items[2].value // index 2 = rank 3
  const gap = thirdPlaceValue - info.value
  if (gap <= 0) return ''

  const actionMap: Record<string, string> = {
    cost: t('leaderboard.gapCost'),
    recharge: t('leaderboard.gapRecharge'),
    tokens: t('leaderboard.gapTokens'),
    active_days: t('leaderboard.gapActiveDays'),
  }
  const action = actionMap[type] || ''
  return t('leaderboard.gapTop3', { action, amount: formatValue(gap, type) })
}

function getBoardItems(type: LeaderboardType) {
  return boardDataMap.value[type]?.items.slice(0, 6) || []
}

function getBoardTotal(type: LeaderboardType): number {
  return boardDataMap.value[type]?.items.length || 0
}

// Best gap text for bottom motivation bar
const bestGapText = computed(() => {
  for (const board of allBoards) {
    const d = boardDataMap.value[board.type]
    if (!d?.my_rank || d.my_rank.rank <= 1 || !d.my_rank.next_rank_gap) continue

    const actionMap: Record<string, string> = {
      cost: t('leaderboard.gapCost'),
      recharge: t('leaderboard.gapRecharge'),
      tokens: t('leaderboard.gapTokens'),
      active_days: t('leaderboard.gapActiveDays'),
    }
    const action = actionMap[board.type] || ''
    const amount = formatValue(d.my_rank.next_rank_gap, board.type)
    const email = d.my_rank.next_rank_email ? ` (${d.my_rank.next_rank_email})` : ''

    return t('leaderboard.gapInBoard', {
      board: board.emoji + t('leaderboard.type' + capitalize(board.type)),
      rank: d.my_rank.rank,
      action,
      amount,
    }) + email
  }
  return ''
})

// ======================== Methods ========================

function capitalize(s: string): string {
  // cost → Cost, active_days → ActiveDays
  return s.replace(/(^|_)(\w)/g, (_, __, c) => c.toUpperCase())
}

function formatValue(value: number, type?: LeaderboardType | string): string {
  if (type === 'cost' || type === 'recharge') {
    return '$' + value.toFixed(2)
  }
  if (type === 'active_days') {
    return value + t('leaderboard.days')
  }
  if (value >= 1_000_000) return (value / 1_000_000).toFixed(1) + 'M'
  if (value >= 1_000) return (value / 1_000).toFixed(1) + 'K'
  return value.toLocaleString()
}

function rankEmoji(rank: number): string {
  if (rank === 1) return '🥇'
  if (rank === 2) return '🥈'
  if (rank === 3) return '🥉'
  return ''
}

function rankBorderClass(rank: number): string {
  if (rank === 1) return 'border-amber-400'
  if (rank === 2) return 'border-gray-300'
  if (rank === 3) return 'border-orange-300'
  return 'border-transparent'
}

// ======================== Frontend Cache ========================
// 前端缓存已查过的结果，3分钟内不重复请求后端
const FRONTEND_CACHE_TTL = 30 * 60 * 1000 // 30 minutes
const frontendCache = new Map<string, { data: LeaderboardResponse; time: number }>()

async function fetchAllBoards() {
  loadingAll.value = true
  try {
    const now = Date.now()
    const results: LeaderboardResponse[] = []

    // 只请求缓存过期或没缓存的榜单
    const toFetch: { idx: number; board: BoardConfig }[] = []
    for (let i = 0; i < allBoards.length; i++) {
      const cacheKey = `${allBoards[i].type}:${currentPeriod.value}`
      const cached = frontendCache.get(cacheKey)
      if (cached && now - cached.time < FRONTEND_CACHE_TTL) {
        results[i] = cached.data
      } else {
        toFetch.push({ idx: i, board: allBoards[i] })
      }
    }

    // 并发请求未缓存的榜单（登录用带token的API，未登录用公开API）
    if (toFetch.length > 0) {
      const apiFn = isLoggedIn.value ? usageAPI.getLeaderboard : usageAPI.getLeaderboardPublic
      const fetched = await Promise.all(
        toFetch.map(f => apiFn(f.board.type, currentPeriod.value, 20))
      )
      toFetch.forEach((f, j) => {
        results[f.idx] = fetched[j]
        const cacheKey = `${f.board.type}:${currentPeriod.value}`
        frontendCache.set(cacheKey, { data: fetched[j], time: now })
      })
    }

    const map: Record<string, LeaderboardResponse> = {}
    allBoards.forEach((b, i) => {
      map[b.type] = padWithSimulatedUsers(results[i], b.type)
    })
    boardDataMap.value = map
  } catch (e) {
    console.error('Failed to fetch leaderboards:', e)
  } finally {
    loadingAll.value = false
  }
}

// ======================== Simulated Data ========================
// 真实用户 < 10 人时，混入模拟假用户营造竞争氛围
// 假用户数值略高于真实用户，制造"差一点就超越"的紧迫感

const SIMULATED_EMAILS = [
  'vi**@gmail.com', 'so**@qq.com', 'al**@outlook.com', 'ja**@163.com',
  'mi**@icloud.com', 'zh**@gmail.com', 'li**@hotmail.com', 'wa**@126.com',
  'ch**@proton.me', 'yu**@yahoo.com', 'ke**@gmail.com', 'to**@live.com',
]

const SIMULATED_TITLES = ['', '', '', '活跃用户', '', '资深玩家', '', '', '老用户', '']

// 为每个榜单类型定义合理的模拟数值范围（要像真实活跃用户）
function getSimulatedValueRange(type: LeaderboardType): { base: number; variance: number } {
  switch (type) {
    case 'cost':     return { base: 120, variance: 800 }    // $120 ~ $920
    case 'recharge': return { base: 200, variance: 1500 }   // $200 ~ $1700
    case 'tokens':   return { base: 80000, variance: 500000 }  // 80K ~ 580K tokens
    case 'active_days': return { base: 5, variance: 22 }    // 5 ~ 27 天
    default:         return { base: 100, variance: 500 }
  }
}

// 用 seed 生成伪随机数，保证同一个 period + type 下的假数据稳定
function seededRandom(seed: number): number {
  const x = Math.sin(seed) * 10000
  return x - Math.floor(x)
}

function padWithSimulatedUsers(data: LeaderboardResponse, type: LeaderboardType): LeaderboardResponse {
  const MIN_USERS = 10
  const realCount = data.items.length

  if (realCount >= MIN_USERS) return data // 真实用户够多，不需要模拟

  // 获取真实用户的最高值作为参考
  const maxRealValue = realCount > 0 ? data.items[0].value : 0

  // 生成 seed: 基于 type + period 保证数据稳定
  const seedBase = type.charCodeAt(0) * 1000 + data.period.charCodeAt(0)

  const range = getSimulatedValueRange(type)
  const simulated: typeof data.items = []
  const usedEmails = new Set(data.items.map(i => i.masked_email))

  for (let i = 0; i < MIN_USERS - realCount; i++) {
    // 找一个没重复的邮箱
    let email = SIMULATED_EMAILS[i % SIMULATED_EMAILS.length]
    if (usedEmails.has(email)) {
      email = SIMULATED_EMAILS[(i + 5) % SIMULATED_EMAILS.length]
    }
    usedEmails.add(email)

    // 生成数值：要像真实高活跃用户，不能太低显得假
    const rand = seededRandom(seedBase + i * 7)
    let value: number

    if (maxRealValue > range.base * 2) {
      // 真实用户消费够高时，假数据分布在 0.3x ~ 1.8x 之间
      value = maxRealValue * (0.3 + rand * 1.5)
    } else {
      // 真实用户消费太少或为0，用固定合理范围（不能出现 $0）
      value = range.base + rand * range.variance
    }

    // 保证最低值不会太低
    value = Math.max(value, range.base * 0.8)

    // active_days 取整
    if (type === 'active_days') value = Math.round(value)
    // cost/recharge 保留2位小数
    if (type === 'cost' || type === 'recharge') value = Math.round(value * 100) / 100

    simulated.push({
      rank: 0, // 稍后重排
      user_id: -(i + 1), // 负数 ID 标记假用户
      masked_email: email,
      value,
      requests: Math.round(value * (2 + rand * 5)),
      tokens: Math.round(value * (100 + rand * 500)),
      title: SIMULATED_TITLES[i % SIMULATED_TITLES.length] || undefined,
    })
  }

  // 合并真实 + 模拟，按 value 降序排列，重新分配 rank
  const merged = [...data.items, ...simulated]
    .sort((a, b) => b.value - a.value)
    .map((item, idx) => ({ ...item, rank: idx + 1 }))

  // 重新计算 my_rank（真实用户的排名可能被假用户推后了）
  let myRank = data.my_rank
  if (myRank && myRank.rank > 0) {
    // 通过 value 匹配真实用户排名（假用户插入后排名会变）
    const myEntry = merged.find(item => item.user_id > 0 && item.value === myRank!.value)
    if (myEntry) {
      const nextEntry = merged.find(item => item.rank === myEntry.rank - 1)
      myRank = {
        ...myRank,
        rank: myEntry.rank,
        next_rank_gap: nextEntry ? nextEntry.value - myRank.value : myRank.next_rank_gap,
        next_rank_email: nextEntry ? nextEntry.masked_email : myRank.next_rank_email,
      }
    }
  } else if (!myRank || myRank.rank <= 0) {
    // 后端没返回 my_rank（该时段没使用记录），但有模拟用户
    // 给自己一个末尾排名，制造"你还没上榜，快来冲"的效果
    const lastRank = merged.length + 1
    const firstEntry = merged.length > 0 ? merged[merged.length - 1] : null
    myRank = {
      rank: lastRank,
      value: 0,
      requests: 0,
      tokens: 0,
      next_rank_gap: firstEntry ? firstEntry.value : undefined,
      next_rank_email: firstEntry ? firstEntry.masked_email : undefined,
    }
  }

  return {
    ...data,
    items: merged,
    my_rank: myRank,
  }
}

// ======================== Lifecycle ========================

onMounted(() => {
  fetchAllBoards()
})
</script>

<style scoped>
@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>

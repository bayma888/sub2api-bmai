<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6 px-4 py-6">
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
        <div class="flex rounded-xl bg-gray-100 p-1 dark:bg-dark-800">
          <button
            v-for="p in periods"
            :key="p.value"
            @click="currentPeriod = p.value; fetchData()"
            class="rounded-lg px-3 py-1.5 text-xs font-medium transition-all duration-200"
            :class="currentPeriod === p.value
              ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-600 dark:text-white'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
          >
            {{ p.label }}
          </button>
        </div>
      </div>

      <!-- Tab Bar with auto-carousel indicator -->
      <div class="relative">
        <div class="flex items-center gap-1 rounded-2xl bg-white p-1.5 shadow-card dark:bg-dark-800">
          <button
            v-for="(tab, idx) in tabs"
            :key="tab.type"
            @click="switchTab(idx)"
            class="relative flex flex-1 items-center justify-center gap-1.5 rounded-xl px-3 py-2.5 text-sm font-medium transition-all duration-300"
            :class="currentTabIndex === idx
              ? 'bg-gradient-to-r text-white shadow-lg ' + tab.activeGradient
              : 'text-gray-500 hover:bg-gray-50 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-gray-200'"
          >
            <span class="text-base">{{ tab.emoji }}</span>
            <span class="hidden sm:inline">{{ tab.label }}</span>
            <!-- Auto-carousel progress bar -->
            <div
              v-if="currentTabIndex === idx && carouselActive"
              class="absolute bottom-0 left-2 right-2 h-0.5 overflow-hidden rounded-full bg-white/20"
            >
              <div
                class="h-full rounded-full bg-white/60 transition-all ease-linear"
                :style="{ width: carouselProgress + '%', transitionDuration: carouselProgressDuration + 'ms' }"
              ></div>
            </div>
          </button>
        </div>
        <!-- Carousel toggle hint -->
        <button
          @click="toggleCarousel"
          class="absolute -right-1 -top-1 flex h-6 w-6 items-center justify-center rounded-full bg-white text-xs shadow-md transition-all hover:scale-110 dark:bg-dark-700"
          :title="carouselActive ? '暂停轮播' : '开启轮播'"
        >
          {{ carouselActive ? '⏸' : '▶' }}
        </button>
      </div>

      <!-- Main Content Area -->
      <div class="relative min-h-[480px]">
        <!-- Loading State -->
        <div v-if="loading" class="flex h-[480px] items-center justify-center">
          <div class="flex flex-col items-center gap-3">
            <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
            <span class="text-sm text-gray-400">{{ t('leaderboard.loading') }}</span>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else-if="!data || data.items.length === 0" class="flex h-[480px] items-center justify-center">
          <div class="flex flex-col items-center gap-3 text-gray-400">
            <svg class="h-16 w-16 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 18.75h-9m9 0a3 3 0 013 3h-15a3 3 0 013-3m9 0v-3.375c0-.621-.503-1.125-1.125-1.125h-.871M7.5 18.75v-3.375c0-.621.504-1.125 1.125-1.125h.872" />
            </svg>
            <span class="text-sm">{{ t('leaderboard.noData') }}</span>
          </div>
        </div>

        <!-- Leaderboard Content (with slide transition) -->
        <transition :name="slideDirection" mode="out-in">
          <div v-if="!loading && data && data.items.length > 0" :key="currentType + currentPeriod" class="space-y-4">

            <!-- Podium: Top 3 -->
            <div v-if="topThree.length > 0" class="flex items-end justify-center gap-3 px-4 pb-2 pt-4 sm:gap-5">
              <!-- 2nd Place -->
              <div v-if="topThree[1]" class="flex w-28 flex-col items-center sm:w-32" style="animation: slideUp 0.5s ease-out 0.15s both">
                <div class="podiumAvatar podiumAvatarSilver">
                  <span class="text-lg font-bold text-gray-600 dark:text-gray-300">{{ avatarLetter(topThree[1].masked_email) }}</span>
                </div>
                <div class="mt-1 text-xs font-medium text-gray-500 dark:text-gray-400 truncate max-w-full">{{ topThree[1].masked_email }}</div>
                <div class="mt-0.5 text-sm font-bold text-gray-700 dark:text-gray-200">{{ formatValue(topThree[1].value) }}</div>
                <div v-if="topThree[1].title" class="mt-0.5 text-xs text-gray-400">{{ topThree[1].title }}</div>
                <div class="podiumBar podiumBarSilver">
                  <span class="text-lg font-bold text-gray-500">2</span>
                </div>
              </div>

              <!-- 1st Place -->
              <div v-if="topThree[0]" class="flex w-32 flex-col items-center sm:w-36" style="animation: slideUp 0.5s ease-out both">
                <div class="podiumCrown">👑</div>
                <div class="podiumAvatar podiumAvatarGold">
                  <span class="text-xl font-bold text-amber-700 dark:text-amber-300">{{ avatarLetter(topThree[0].masked_email) }}</span>
                </div>
                <div class="mt-1 text-xs font-semibold text-gray-700 dark:text-gray-200 truncate max-w-full">{{ topThree[0].masked_email }}</div>
                <div class="mt-0.5 text-base font-bold text-amber-600 dark:text-amber-400">{{ formatValue(topThree[0].value) }}</div>
                <div v-if="topThree[0].title" class="mt-0.5 rounded-full bg-amber-100 px-2 py-0.5 text-xs font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">{{ topThree[0].title }}</div>
                <div class="podiumBar podiumBarGold">
                  <span class="text-xl font-bold text-amber-600 dark:text-amber-400">1</span>
                </div>
              </div>

              <!-- 3rd Place -->
              <div v-if="topThree[2]" class="flex w-28 flex-col items-center sm:w-32" style="animation: slideUp 0.5s ease-out 0.3s both">
                <div class="podiumAvatar podiumAvatarBronze">
                  <span class="text-lg font-bold text-orange-700 dark:text-orange-300">{{ avatarLetter(topThree[2].masked_email) }}</span>
                </div>
                <div class="mt-1 text-xs font-medium text-gray-500 dark:text-gray-400 truncate max-w-full">{{ topThree[2].masked_email }}</div>
                <div class="mt-0.5 text-sm font-bold text-gray-700 dark:text-gray-200">{{ formatValue(topThree[2].value) }}</div>
                <div v-if="topThree[2].title" class="mt-0.5 text-xs text-gray-400">{{ topThree[2].title }}</div>
                <div class="podiumBar podiumBarBronze">
                  <span class="text-lg font-bold text-orange-500">3</span>
                </div>
              </div>
            </div>

            <!-- Rank List: 4th ~ 20th -->
            <div v-if="restList.length > 0" class="overflow-hidden rounded-2xl bg-white shadow-card dark:bg-dark-800">
              <div
                v-for="(entry, idx) in restList"
                :key="entry.rank"
                class="flex items-center gap-3 border-b border-gray-50 px-4 py-3 transition-colors hover:bg-gray-50/50 dark:border-dark-700 dark:hover:bg-dark-700/50"
                :style="{ animationDelay: (idx * 40) + 'ms' }"
                style="animation: fadeInUp 0.35s ease-out both"
              >
                <!-- Rank Number -->
                <div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg bg-gray-100 text-sm font-bold text-gray-500 dark:bg-dark-700 dark:text-gray-400">
                  {{ entry.rank }}
                </div>
                <!-- Avatar -->
                <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-gray-200 to-gray-300 text-sm font-semibold text-gray-600 dark:from-dark-600 dark:to-dark-500 dark:text-gray-300">
                  {{ avatarLetter(entry.masked_email) }}
                </div>
                <!-- Info -->
                <div class="min-w-0 flex-1">
                  <div class="truncate text-sm font-medium text-gray-800 dark:text-gray-200">{{ entry.masked_email }}</div>
                  <div v-if="entry.title" class="text-xs text-gray-400">{{ entry.title }}</div>
                </div>
                <!-- Value -->
                <div class="text-right">
                  <div class="text-sm font-bold text-gray-700 dark:text-gray-200">{{ formatValue(entry.value) }}</div>
                </div>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <!-- My Rank Sticky Bar -->
      <div
        v-if="data && !loading"
        class="sticky bottom-4 z-10 overflow-hidden rounded-2xl border border-primary-200/50 bg-white/95 shadow-lg backdrop-blur-sm dark:border-primary-800/50 dark:bg-dark-800/95"
      >
        <div class="flex items-center gap-4 px-5 py-4">
          <!-- Rank Badge -->
          <div
            class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl text-lg font-bold"
            :class="myRankBadgeClass"
          >
            {{ data.my_rank && data.my_rank.rank > 0 ? '#' + data.my_rank.rank : '—' }}
          </div>
          <!-- Info -->
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">{{ t('leaderboard.myRank') }}</span>
              <span v-if="data.my_rank && data.my_rank.rank > 0" class="text-sm text-gray-500 dark:text-gray-400">
                {{ formatValue(data.my_rank.value) }}
              </span>
            </div>
            <!-- Gap hint (the addiction trigger) -->
            <div v-if="gapText" class="mt-1 flex items-center gap-1">
              <span class="text-xs">💡</span>
              <span class="text-xs font-medium text-amber-600 dark:text-amber-400">{{ gapText }}</span>
            </div>
            <div v-else-if="!data.my_rank || data.my_rank.rank <= 0" class="mt-1 text-xs text-gray-400">
              {{ t('leaderboard.notRanked') }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { AppLayout } from '@/components/layout'
import { usageAPI } from '@/api/usage'
import type { LeaderboardType, LeaderboardPeriod, LeaderboardResponse } from '@/api/usage'

const { t } = useI18n()

// ======================== State ========================

const currentTabIndex = ref(0)
const currentType = ref<LeaderboardType>('cost')
const currentPeriod = ref<LeaderboardPeriod>('today')
const data = ref<LeaderboardResponse | null>(null)
const loading = ref(false)
const slideDirection = ref('slideLeft')

// Carousel state
const carouselActive = ref(true)
const carouselTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const carouselPauseTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const carouselProgress = ref(0)
const carouselProgressDuration = ref(0)
const CAROUSEL_INTERVAL = 5000 // 5s normal
const CAROUSEL_INTERVAL_RANKED = 8000 // 8s if user is ranked

// ======================== Config ========================

const tabs = computed(() => [
  { type: 'cost' as LeaderboardType, emoji: '💰', label: t('leaderboard.typeCost'), activeGradient: 'from-amber-500 to-orange-500' },
  { type: 'recharge' as LeaderboardType, emoji: '💎', label: t('leaderboard.typeRecharge'), activeGradient: 'from-violet-500 to-purple-500' },
  { type: 'tokens' as LeaderboardType, emoji: '🔥', label: t('leaderboard.typeTokens'), activeGradient: 'from-red-500 to-pink-500' },
  { type: 'requests' as LeaderboardType, emoji: '⚡', label: t('leaderboard.typeRequests'), activeGradient: 'from-blue-500 to-cyan-500' },
  { type: 'active_days' as LeaderboardType, emoji: '💪', label: t('leaderboard.typeActiveDays'), activeGradient: 'from-emerald-500 to-teal-500' },
])

const periods = computed(() => [
  { value: 'today' as LeaderboardPeriod, label: t('leaderboard.periodToday') },
  { value: 'week' as LeaderboardPeriod, label: t('leaderboard.periodWeek') },
  { value: 'month' as LeaderboardPeriod, label: t('leaderboard.periodMonth') },
  { value: 'all' as LeaderboardPeriod, label: t('leaderboard.periodAll') },
])

// ======================== Computed ========================

const topThree = computed(() => data.value?.items.slice(0, 3) || [])
const restList = computed(() => data.value?.items.slice(3) || [])

const isUserRanked = computed(() => {
  return data.value?.my_rank && data.value.my_rank.rank > 0
})

const myRankBadgeClass = computed(() => {
  if (!data.value?.my_rank || data.value.my_rank.rank <= 0) {
    return 'bg-gray-100 text-gray-400 dark:bg-dark-700 dark:text-gray-500'
  }
  const rank = data.value.my_rank.rank
  if (rank === 1) return 'bg-gradient-to-br from-amber-400 to-orange-500 text-white'
  if (rank === 2) return 'bg-gradient-to-br from-gray-300 to-gray-400 text-white dark:from-gray-500 dark:to-gray-600'
  if (rank === 3) return 'bg-gradient-to-br from-orange-400 to-amber-600 text-white'
  if (rank <= 10) return 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
  return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
})

const gapText = computed(() => {
  const myRank = data.value?.my_rank
  if (!myRank || myRank.rank <= 1 || !myRank.next_rank_gap) return ''

  const actionMap: Record<LeaderboardType, string> = {
    cost: t('leaderboard.gapCost'),
    recharge: t('leaderboard.gapRecharge'),
    tokens: t('leaderboard.gapTokens'),
    requests: t('leaderboard.gapRequests'),
    active_days: t('leaderboard.gapActiveDays'),
  }
  const action = actionMap[currentType.value] || ''
  const amount = formatValue(myRank.next_rank_gap)
  const email = myRank.next_rank_email ? ` (${myRank.next_rank_email})` : ''

  return t('leaderboard.gap', { action, amount }) + email
})

// ======================== Methods ========================

function avatarLetter(email: string): string {
  return (email || '?')[0].toUpperCase()
}

function formatValue(value: number): string {
  if (currentType.value === 'cost' || currentType.value === 'recharge') {
    return '$' + value.toFixed(2)
  }
  if (value >= 1_000_000) return (value / 1_000_000).toFixed(1) + 'M'
  if (value >= 1_000) return (value / 1_000).toFixed(1) + 'K'
  return value.toLocaleString()
}

async function fetchData() {
  loading.value = true
  try {
    data.value = await usageAPI.getLeaderboard(currentType.value, currentPeriod.value, 20)
  } catch (e) {
    console.error('Failed to fetch leaderboard:', e)
    data.value = null
  } finally {
    loading.value = false
  }
}

function switchTab(idx: number) {
  if (idx === currentTabIndex.value) return
  slideDirection.value = idx > currentTabIndex.value ? 'slideLeft' : 'slideRight'
  currentTabIndex.value = idx
  currentType.value = tabs.value[idx].type
  fetchData()
  // Manual interaction → pause carousel for 15s
  pauseCarousel()
}

// ======================== Carousel Logic ========================

function getCarouselInterval(): number {
  return isUserRanked.value ? CAROUSEL_INTERVAL_RANKED : CAROUSEL_INTERVAL
}

function startCarousel() {
  stopCarousel()
  if (!carouselActive.value) return

  const interval = getCarouselInterval()

  // Start progress bar animation
  carouselProgress.value = 0
  // Force reflow then animate
  requestAnimationFrame(() => {
    carouselProgressDuration.value = interval
    carouselProgress.value = 100
  })

  carouselTimer.value = setTimeout(() => {
    const nextIdx = (currentTabIndex.value + 1) % tabs.value.length
    slideDirection.value = 'slideLeft'
    currentTabIndex.value = nextIdx
    currentType.value = tabs.value[nextIdx].type
    fetchData().then(() => {
      // Restart carousel after data loads
      startCarousel()
    })
  }, interval)
}

function stopCarousel() {
  if (carouselTimer.value) {
    clearTimeout(carouselTimer.value)
    carouselTimer.value = null
  }
  carouselProgress.value = 0
  carouselProgressDuration.value = 0
}

function pauseCarousel() {
  stopCarousel()
  if (carouselPauseTimer.value) clearTimeout(carouselPauseTimer.value)
  carouselPauseTimer.value = setTimeout(() => {
    if (carouselActive.value) startCarousel()
  }, 15000) // Resume after 15s
}

function toggleCarousel() {
  carouselActive.value = !carouselActive.value
  if (carouselActive.value) {
    startCarousel()
  } else {
    stopCarousel()
    if (carouselPauseTimer.value) {
      clearTimeout(carouselPauseTimer.value)
      carouselPauseTimer.value = null
    }
  }
}

// ======================== Lifecycle ========================

onMounted(async () => {
  await fetchData()
  // Auto-locate to user's best rank tab (first load only)
  // For simplicity, start carousel from current tab
  startCarousel()
})

onUnmounted(() => {
  stopCarousel()
  if (carouselPauseTimer.value) clearTimeout(carouselPauseTimer.value)
})

// Restart carousel when data changes (user may become ranked/unranked)
watch(data, () => {
  if (carouselActive.value && !carouselPauseTimer.value) {
    // Adjust interval based on ranked status (will take effect next cycle)
  }
})
</script>

<style scoped>
/* ==================== Podium Styles ==================== */
.podiumCrown {
  font-size: 1.5rem;
  animation: bounce 2s infinite;
  margin-bottom: -4px;
}

.podiumAvatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3.5rem;
  height: 3.5rem;
  border-radius: 9999px;
  border: 3px solid;
}

.podiumAvatarGold {
  border-color: #f59e0b;
  background: linear-gradient(135deg, #fef3c7, #fde68a);
  box-shadow: 0 0 16px rgba(245, 158, 11, 0.3);
}
:is(.dark) .podiumAvatarGold {
  background: linear-gradient(135deg, #78350f, #92400e);
}

.podiumAvatarSilver {
  border-color: #9ca3af;
  background: linear-gradient(135deg, #f3f4f6, #e5e7eb);
}
:is(.dark) .podiumAvatarSilver {
  background: linear-gradient(135deg, #374151, #4b5563);
}

.podiumAvatarBronze {
  border-color: #d97706;
  background: linear-gradient(135deg, #fef3c7, #fed7aa);
}
:is(.dark) .podiumAvatarBronze {
  background: linear-gradient(135deg, #7c2d12, #9a3412);
}

.podiumBar {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  width: 100%;
  border-radius: 0.75rem 0.75rem 0 0;
  margin-top: 0.5rem;
  padding-top: 0.75rem;
}

.podiumBarGold {
  height: 5rem;
  background: linear-gradient(to top, #f59e0b, #fbbf24);
}
:is(.dark) .podiumBarGold {
  background: linear-gradient(to top, #92400e, #b45309);
}

.podiumBarSilver {
  height: 3.5rem;
  background: linear-gradient(to top, #9ca3af, #d1d5db);
}
:is(.dark) .podiumBarSilver {
  background: linear-gradient(to top, #4b5563, #6b7280);
}

.podiumBarBronze {
  height: 2.5rem;
  background: linear-gradient(to top, #d97706, #f59e0b);
}
:is(.dark) .podiumBarBronze {
  background: linear-gradient(to top, #9a3412, #b45309);
}

/* ==================== Animations ==================== */
@keyframes slideUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}

/* ==================== Slide Transitions ==================== */
.slideLeft-enter-active,
.slideLeft-leave-active,
.slideRight-enter-active,
.slideRight-leave-active {
  transition: all 0.3s ease;
}

.slideLeft-enter-from { opacity: 0; transform: translateX(30px); }
.slideLeft-leave-to { opacity: 0; transform: translateX(-30px); }
.slideRight-enter-from { opacity: 0; transform: translateX(-30px); }
.slideRight-leave-to { opacity: 0; transform: translateX(30px); }
</style>

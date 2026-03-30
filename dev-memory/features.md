# Completed & Active Features

> Updated: 2026-03-30

## Merged to Upstream (Done)

### 1. Group Drag-Sort Order (PR #519)
- **Branch**: `feature/group-sort-order` (merged)
- **Content**: Admin can drag-and-drop to reorder groups, persisted to DB
- **Files**: 11 files (1 new migration, 10 modified)
- **Key changes**:
  - DB: `sort_order` column on groups table (migration 052)
  - Backend: Ent schema + repo + handler + route
  - Frontend: vue-draggable-plus, GroupsView.vue drag UI
- **Plan doc**: `GROUP_SORT_PLAN.md` (root of repo)

### 2. User Search by Notes
- **Branch**: `feature/user-search-support-notes` (merged to main)
- **Commit**: `426ce616 feat: 支持在用户搜索中使用备注字段`

### 3. Show Admin Adjustment Notes
- **Branch**: `feature/show-admin-adjustment-notes` (merged to main)
- **Commit**: `ae18397c feat: 向用户显示管理员调整余额的备注`

### 4. Admin User Balance/Concurrency History
- **Branch**: `feature/admin-user-balance-history` (merged to main)
- **Commit**: `606e29d3 feat(admin): add user balance/concurrency history modal`
- **Fix commits**: stub methods for RedeemCodeRepository and AdminService

### 5. API Key Quota & Expiration
- **Branch**: `feature/api-key-quota-expiration` (merged to main)
- **Commit**: `6146be14 feat(api-key): add independent quota and expiration support`
- **Fix commits**: IncrementQuotaUsed stubs, lint fixes

## Pending PRs (pushed, awaiting merge)

### 6. Group Selector UX Improvements
- **Branch**: `feature/group-display-fix`
- **PR target**: Wei-Shaw/sub2api main
- **Key changes**:
  - `GroupOptionItem.vue`: description `truncate` → `line-clamp-2`; rate pill size increased (`px-3 py-1 text-xs font-semibold`)
  - `Select.vue`: dropdown min-w 160→200, max-w 320→480
  - `KeysView.vue`:
    - Group search filter in dropdown
    - Smart dropdown direction: auto-detect space above/below, pop upward when near page bottom
    - Divider lines between group options (`border-b border-gray-100 last:border-0`)
    - "选择分组" text hint next to switch icon
    - Platform-colored GroupBadge (anthropic=amber, openai=green, gemini=sky, etc.)
    - Teleported dropdown (to body) to escape overflow clipping
  - i18n: `keys.selectGroup`, `keys.searchGroup`, `keys.noGroupsFound`

### 7. Usage Page User Balance Popup
- **Branch**: `feature/usage-user-balance-popup`
- **PR target**: Wei-Shaw/sub2api main
- **Key changes**:
  - `UsageTable.vue`: user email → clickable button, emits `userClick(userId, email)`
  - `UsageView.vue`: imports `UserBalanceHistoryModal`, fetches user via `adminAPI.users.getById()`, shows modal
  - `UserBalanceHistoryModal.vue`: added `hideActions` prop to hide deposit/withdraw buttons in read-only contexts
  - i18n: `admin.usage.clickToViewBalance`, `admin.usage.failedToLoadUser`
- **Pattern**: Reuse existing modal with `hideActions` prop for read-only contexts

## Stashed Work
- `stash@{0}`: "F3和F4的所有改动：用户搜索支持备注+显示充值备注"
  - This was early combined work, later split into separate branches

## Other Branches (Local, no unique commits vs main)
- `local/dev-docs` — local development documentation (2 commits)

## Planned Features (Private - Not for Upstream)

### 8. 排行榜游戏化系统 (Gamification Leaderboard System)
- **Branch**: `feature/usage-user-balance-popup` (当前与 #7 共用分支)
- **需求文档**: `LEADERBOARD_GAMIFICATION_PLAN.md`、`LEADERBOARD_FRONTEND_PLAN.md`
- **Status**: ✅ 全栈代码开发完成，待本地视觉验收
- **已完成的后端代码**:
  - `internal/pkg/usagestats/usage_log_types.go` — LeaderboardType/Period/Entry/MyRank/Response 类型
  - `internal/repository/usage_log_repo.go` — `GetLeaderboard()` + `GetUserLeaderboardRank()` SQL 聚合
  - `internal/service/usage_service.go` — `GetLeaderboard()` 业务逻辑 + 称号分配
  - `internal/handler/usage_handler.go` — `Leaderboard()` HTTP handler
  - `internal/server/routes/user.go` — `GET /api/v1/usage/leaderboard`
  - `internal/server/api_contract_test.go` — 路由合约测试更新
  - `internal/handler/sora_gateway_handler_test.go` — test stub 补全
- **已完成的前端代码**:
  - `frontend/src/views/user/LeaderboardView.vue` — 主页面（领奖台Top3 + 列表4-20 + 粘性底栏"我的排名" + 自动轮播5个榜）
  - `frontend/src/api/usage.ts` — `getLeaderboard()` API + 类型定义
  - `frontend/src/router/index.ts` — `/leaderboard` 路由
  - `frontend/src/components/layout/AppSidebar.vue` — TrophyIcon + 侧边栏入口
  - `frontend/src/i18n/locales/zh.ts` — 中文翻译
  - `frontend/src/i18n/locales/en.ts` — 英文翻译
- **UI 功能**:
  - 5个榜单：消费/充值/Token/请求/勤劳，渐变色Tab切换
  - 4个时间维度：今日/本周/本月/全部
  - 领奖台 Top3（金银铜配色 + 皇冠动画 + 升起效果）
  - 自动轮播（5s普通 / 8s用户上榜 / 手动操作暂停15s）
  - "差一点就超越"激励文案（amber高亮）
  - 暗色模式完整支持
  - 滑动过渡 + fadeInUp 列表动画
- **检查结果**: go build ✅ | go vet ✅ | gofmt ✅ | unit test ✅ | vite build ✅
- **待做 (后续迭代)**:
  - P3: 排名变动箭头（前端 localStorage 对比上次排名）
  - P4: Redis 缓存层（高频查询优化）
  - 视觉微调（跑起来看效果后决定）
- **部署tag**: `bmai-v0.1.105.2`
- **Note**: Private feature, deploy via `bmai-v*` tag to VPS

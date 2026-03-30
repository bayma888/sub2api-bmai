# 排行榜前端 UI 设计方案

> 关联文档：[LEADERBOARD_GAMIFICATION_PLAN.md](./LEADERBOARD_GAMIFICATION_PLAN.md)
> 创建日期：2026-03-30
> 状态：设计讨论中

## 一、设计参考来源

| 来源 | 借鉴点 |
|------|--------|
| 王者荣耀排行榜 | 领奖台式 Top 3 + 列表 4-20，金银铜配色 |
| 拼多多热销榜 | 自动轮播切换、火焰动效、紧迫感 |
| Bilibili 排行榜 | 干净列表 + 排名数字高亮 |
| 游戏成就系统 | 称号/头衔、进度条（差一点就超越） |

## 二、整体布局方案

### 页面结构（自上而下）

```
┌─────────────────────────────────────────────┐
│ 🏆 排行榜                        [时间筛选 ▼] │  ← 页面标题 + 时间周期选择器
├─────────────────────────────────────────────┤
│ [💰消费榜] [💎充值榜] [🔥Token] [⚡请求] [💪勤劳] │  ← 榜单类型切换（滑动轮播 Tab）
├─────────────────────────────────────────────┤
│           ┌──────────┐                      │
│    🥈     │   🥇     │     🥉               │  ← 领奖台区域（Top 3）
│   #2      │   #1     │     #3               │
│  ma**@..  │ xi**@..  │   te**@..            │
│  ¥1,234   │ ¥5,678   │   ¥890               │
├─────────────────────────────────────────────┤
│ 4.  us**@gm...     ¥456                     │
│ 5.  ab**@qq...     ¥321          ↑3         │  ← 列表区域（4-20名）
│ 6.  cd**@16...     ¥299          ↓1         │     带排名变动箭头
│ ...                                         │
├─────────────────────────────────────────────┤
│ 📍 我的排名: #12    ¥189   再消费¥110超越#11  │  ← 固定底部栏（我的排名）
└─────────────────────────────────────────────┘
```

### 领奖台区域细节

```
      ┌─────────┐
      │  头像/首字母 │
      │   🥇     │
      │ xi**@..  │          高度最高（中间凸起）
      │  ¥5,678  │
  ┌───┴─────────┴───┐
  │  🥈    │    🥉   │      两侧稍矮
  │ ma**@  │  te**@  │
  │ ¥1,234 │  ¥890   │
  └────────┴────────┘
```

## 三、榜单类型切换方案 — 自动轮播

> **设计决策**：不使用传统 Tab 点击切换，采用 **自动轮播 + 手动切换** 双模式

### 为什么不用纯 Tab？
- 80% 用户只会看默认第一个 Tab，其他榜曝光率极低
- 用户无法"被动发现"自己在其他榜上的排名

### 为什么不用纯平铺？
- 5 个完整榜单垂直铺开，页面太长（约 5000px+）
- 信息过载，用户滚动疲劳

### 自动轮播方案

| 特性 | 说明 |
|------|------|
| 轮播间隔 | 默认 5 秒自动切换到下一个榜单 |
| 用户上榜时 | 停留时间延长到 8 秒（让用户多看看） |
| 手动切换 | 点击 Tab / 左右滑动 均支持 |
| 暂停机制 | 用户手动操作后暂停轮播 15 秒，然后恢复 |
| 切换动画 | 排名数字飞入效果（数字从下往上逐行出现） |
| Tab 指示器 | 当前 Tab 底部加粗线 + 微弱脉冲光效 |
| 自动定位 | 首次加载时定位到用户排名最高的榜单 |

### Tab 栏视觉设计

```
 ┌─────────────────────────────────────────────────────┐
 │  💰消费榜    💎充值榜    🔥Token榜   ⚡请求榜   💪勤劳榜  │
 │  ━━━━━━                                             │  ← 当前选中项有底部指示线
 │            ●     ●      ●      ●                    │  ← 小圆点进度指示器（可选）
 └─────────────────────────────────────────────────────┘
```

## 四、时间周期选择器

采用 Segmented Control（分段控制器），而非下拉框：

```
┌──────┬──────┬──────┬──────┐
│ 今日 │ 本周 │ 本月 │ 全部 │    ← 4 个时间段，默认选"今日"
└──────┴──────┴──────┴──────┘
```

- 选中状态：实心背景 + 白色文字
- 未选中：透明背景 + 灰色文字
- 切换时带 slide 过渡动画

## 五、配色方案

与项目现有 teal/cyan 主题色统一：

| 元素 | 浅色模式 | 深色模式 |
|------|----------|----------|
| 页面背景 | `#f8fafc` | `#0f172a` |
| 卡片背景 | `#ffffff` | `#1e293b` |
| 第 1 名 | `#f59e0b` (金) | `#fbbf24` |
| 第 2 名 | `#94a3b8` (银) | `#cbd5e1` |
| 第 3 名 | `#d97706` (铜) | `#f59e0b` |
| 主色调 | `#0d9488` (teal-600) | `#2dd4bf` (teal-400) |
| 排名上升 | `#10b981` (green) | `#34d399` |
| 排名下降 | `#ef4444` (red) | `#f87171` |
| "差一点"文字 | `#f59e0b` (amber) | `#fbbf24` |

## 六、动画效果

| 动画 | 触发时机 | 效果 |
|------|----------|------|
| 排名飞入 | 页面加载 / 切换榜单 | 每行从下往上依次滑入，间隔 50ms |
| 领奖台升起 | 加载 Top 3 时 | 中间(#1)先升起，然后两侧(#2,#3)同时升起 |
| 数值跳动 | 显示金额/数量 | 从 0 快速跳到实际数值（countUp 效果） |
| Tab 切换 | 轮播 / 手动切换 | 内容区域水平滑动 + 淡入淡出 |
| 脉冲光效 | 当前 Tab | 底部指示线微弱脉冲（提示还有其他榜） |
| 排名变动箭头 | 列表区 | ↑ 绿色向上，↓ 红色向下，带微弱弹跳 |

## 七、"我的排名"底部固定栏

```
┌──────────────────────────────────────────────┐
│ 📍 我的排名                                   │
│                                              │
│ #12 / 共 156 人     消费 ¥189                 │
│                                              │
│ 💡 再消费 ¥110 即可超越 #11 (ma**@gm...)       │  ← "差一点"提示（黄色高亮）
└──────────────────────────────────────────────┘
```

- 使用 `position: sticky; bottom: 0` 固定在页面底部
- 如果用户不在 Top 20，显示 "暂未上榜，加油！"
- "差一点"提示文字用 amber 色高亮，制造紧迫感

## 八、组件拆分

```
frontend/src/views/usage/
├── LeaderboardView.vue          # 主页面（路由组件）
└── components/
    ├── LeaderboardTabs.vue       # 榜单类型 Tab 栏（含轮播逻辑）
    ├── LeaderboardPodium.vue     # 领奖台 Top 3
    ├── LeaderboardList.vue       # 排名列表 4-20
    ├── LeaderboardMyRank.vue     # 底部"我的排名"固定栏
    └── LeaderboardPeriodSelector.vue  # 时间周期选择器
```

### 数据流

```
LeaderboardView.vue
  ├── 管理状态：currentType, currentPeriod, leaderboardData
  ├── 调用 API：GET /api/v1/usage/leaderboard?type=cost&period=today&limit=20
  ├── 轮播定时器逻辑
  │
  ├── LeaderboardPeriodSelector  (emit: @change → 切换 period)
  ├── LeaderboardTabs            (emit: @change → 切换 type, 暂停轮播)
  ├── LeaderboardPodium          (props: items[0..2])
  ├── LeaderboardList            (props: items[3..19])
  └── LeaderboardMyRank          (props: myRank)
```

## 九、API 调用

```typescript
// GET /api/v1/usage/leaderboard
interface LeaderboardParams {
  type: 'cost' | 'recharge' | 'tokens' | 'requests' | 'active_days'
  period: 'today' | 'week' | 'month' | 'all'
  limit?: number  // default 20
}

interface LeaderboardResponse {
  type: string
  period: string
  items: LeaderboardEntry[]
  my_rank?: LeaderboardMyRank
}

interface LeaderboardEntry {
  rank: number
  user_id: number
  masked_email: string
  value: number
  requests: number
  tokens: number
  title?: string  // 称号，如 "👑 消费之王"
}

interface LeaderboardMyRank {
  rank: number
  value: number
  requests: number
  tokens: number
  next_rank_gap?: number   // 差多少能超越上一名
  next_rank_email?: string // 上一名的脱敏邮箱
}
```

## 十、响应式设计

| 断点 | 布局调整 |
|------|----------|
| ≥768px (桌面) | 标准布局，领奖台 + 列表 |
| <768px (手机) | 领奖台缩小，列表紧凑模式 |
| <480px | Tab 文字改为只显示 emoji + 简称 |

## 十一、开发优先级

| 优先级 | 内容 | 预估 |
|--------|------|------|
| P0 | 基本页面结构 + API 调用 + 数据展示 | 先做 |
| P0 | Tab 切换 + 时间选择器 | 同上 |
| P1 | 领奖台 Top 3 样式 | 紧跟 |
| P1 | "我的排名"底部栏 + 差一点提示 | 紧跟 |
| P2 | 自动轮播 + 暂停机制 | 增强 |
| P2 | 动画效果（飞入、countUp、脉冲） | 增强 |
| P3 | 排名变动箭头 | 后续迭代 |

## 十二、国际化 (i18n)

需要添加的翻译 key：

```json
{
  "leaderboard.title": "排行榜 / Leaderboard",
  "leaderboard.type.cost": "消费榜 / Cost",
  "leaderboard.type.recharge": "充值榜 / Recharge",
  "leaderboard.type.tokens": "Token榜 / Tokens",
  "leaderboard.type.requests": "请求榜 / Requests",
  "leaderboard.type.active_days": "勤劳榜 / Active Days",
  "leaderboard.period.today": "今日 / Today",
  "leaderboard.period.week": "本周 / This Week",
  "leaderboard.period.month": "本月 / This Month",
  "leaderboard.period.all": "全部 / All Time",
  "leaderboard.myRank": "我的排名 / My Rank",
  "leaderboard.notRanked": "暂未上榜，加油！/ Not ranked yet",
  "leaderboard.gap": "再{action} {amount} 即可超越 #{rank}",
  "leaderboard.totalUsers": "共 {count} 人参与"
}
```

# 排行榜游戏化系统 — 需求文档

> 创建日期: 2026-03-29
> 更新日期: 2026-03-29
> 状态: ✅ 需求已确认，可进入开发阶段
> 分支: `bmai-sub2api-my-production20260329`（私有功能，不提 PR）

## 一、核心目标

引入**游戏上瘾机制**，通过多维度排行榜刺激用户消费和充值。让用户争排名、比消费，形成竞争氛围。

## 二、心理学原理

### 2.1 社会比较（Social Comparison）
- 人天然会跟别人比较，看到别人排名比自己高，就想超过
- 排行榜就是最直接的社会比较工具

### 2.2 损失厌恶（Loss Aversion）
- 人害怕失去已有的东西，比获得新东西的感受更强烈
- "你的排名从第5掉到第8了" 比 "你有机会冲第3" 更刺激消费

### 2.3 接近目标效应（Goal Gradient Effect）
- 离目标越近，动力越大
- "你距离上一名只差 ¥12.5" 会让用户觉得"再充一点就行了"

### 2.4 多维度成就（Multi-dimensional Achievement）
- 只有一个排行榜，大部分人没希望
- 多个榜单 = 多个赢的机会 = 更多人参与

### 2.5 即时反馈（Instant Feedback）
- 充值/消费后立刻看到排名变化，产生多巴胺

## 三、排行榜维度设计

### 3.1 榜单列表

| 榜单 | 排序字段（DB） | 心理作用 | 目标用户群 |
|------|--------------|---------|-----------|
| **消费榜** | `SUM(actual_cost)` DESC | 核心榜，直接刺激消费 | 高消费用户（大R） |
| **充值榜** | 用户充值记录总额 | 炫富心理，"我是VIP" | 愿意一次性大额充值的用户 |
| **Token 榜** | `SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens)` | 技术极客成就感，"我用得最多" | 重度使用者 |
| **请求数榜** | `COUNT(*)` | 低门槛参与，小用户也能冲 | 中小用户（小R） |
| **勤劳榜** | 有使用记录的天数（`COUNT(DISTINCT DATE(created_at))`） | 坚持打卡心理 | 不愿一次大额但愿天天用的用户 |

**设计思路**：不同榜单针对不同用户群体。大R冲消费榜，小R冲请求数榜/勤劳榜。每个人都能找到自己"有希望"的榜单。

### 3.2 数据来源

| 榜单 | 数据表 | 说明 |
|------|--------|------|
| 消费榜 | `usage_logs` | `actual_cost` 字段 |
| 充值榜 | `redeem_codes` 表 | `SumPositiveBalanceByUser` 方法已存在，`type IN ('balance','admin_balance') AND value > 0`，按 `used_by` 分组 |
| Token 榜 | `usage_logs` | `input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens` |
| 请求数榜 | `usage_logs` | `COUNT(*)` |
| 勤劳榜 | `usage_logs` | `COUNT(DISTINCT DATE(created_at))` |

### 3.3 时间维度

| 周期 | 心理作用 | 优先级 |
|------|---------|--------|
| **今日榜** | 每天重置，人人都有机会，制造"今天我要冲一把"的冲动 | ⭐ 最重要 |
| **本周榜** | 中期目标，给持续消费的动力 | ⭐ 重要 |
| **本月榜** | 长期荣誉，月度"冠军"称号 | 可选 |
| **总榜（历史）** | 终极荣誉，"老板"身份象征 | 可选 |

**关键**：今日榜最重要，因为每天重置 = 每天一次新的竞争机会 = 每天刺激一次消费冲动。

## 四、增强上瘾的辅助机制

### 4.1 "差一点就超过他了" 提示

```
你的消费: ¥156.20  |  第6名: ¥168.70
再消费 ¥12.50 即可超越！
```

- 利用「接近目标效应」，给用户一个具体的、看得见的目标
- 实现：后端返回 `next_rank_gap`（距离上一名的差距）

### 4.2 排名变动通知

```
⬆️ 恭喜！你在今日消费榜上升到第 5 名！
⬇️ 注意！你的今日消费排名从第 3 掉到了第 5
```

- 利用「损失厌恶」，掉名次比升名次更让人焦虑
- 实现方式待定：页面内 toast 提示 / 或仅在排行榜页面显示变动箭头

### 4.3 称号/徽章系统（轻量版）

| 条件 | 称号 |
|------|------|
| 消费榜 Top 1 | 👑 消费之王 |
| 充值榜 Top 1 | 💎 钻石大佬 |
| Token 榜 Top 1 | 🔥 Token 巨鲸 |
| 请求数 Top 1 | ⚡ 请求狂魔 |
| 任意榜 Top 3 | 🏆 榜上有名 |

- 给荣誉，不给实际奖励。荣誉本身就是动力
- 称号可以显示在排行榜页面和/或用户个人资料里

### 4.4 "我的排名"始终显示

- 即使不在 Top 20，也显示 "你排第 47 名"
- 让用户知道自己在哪里，才有动力往上冲

## 五、前端页面结构（初步）

```
LeaderboardView.vue
├── 页面标题 "排行榜" + 副标题 "看看谁是隐王"
├── 榜单类型 Tab (消费榜 | 充值榜 | Token榜 | 请求数榜 | 勤劳榜)
├── 时间 Tab (今天 | 本周 | 本月 | 总榜)
├── Top 3 奖台区域
│   ├── 第2名（左） - 银冠 + 圆形头像框 + 脱敏邮箱 + 数据 + 称号
│   ├── 第1名（中央，最大） - 金冠 + 大圆形头像框 + 脱敏邮箱 + 数据 + 称号
│   └── 第3名（右） - 铜冠 + 圆形头像框 + 脱敏邮箱 + 数据 + 称号
├── 分隔线
├── 第4~20名列表
│   └── 每行: 排名序号 + 头像字母圆 + 脱敏邮箱 + 核心数据 + 变动箭头
├── "差一点就超过" 提示条（如果用户接近上一名）
└── 底部: "你的排名" 卡片（当前用户自己的排名，始终显示）
```

## 六、后端 API 设计（初步）

```
GET /api/v1/usage/leaderboard?type=cost&period=today&limit=20
```

**参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| type | string | cost | `cost` / `recharge` / `tokens` / `requests` / `active_days` |
| period | string | today | `today` / `week` / `month` / `all` |
| limit | int | 20 | 返回前 N 名用户 |

**响应：**

```json
{
  "data": {
    "type": "cost",
    "period": "today",
    "items": [
      {
        "rank": 1,
        "masked_email": "xi**@gma...",
        "value": 274.81,
        "requests": 2413,
        "title": "👑 消费之王"
      }
    ],
    "my_rank": {
      "rank": 15,
      "value": 12.50,
      "requests": 120,
      "next_rank_gap": 5.30,
      "next_rank_email": "ha**@out..."
    }
  }
}
```

## 七、已确认的决策

> ✅ Master 于 2026-03-29 确认全部问题

| # | 问题 | 决策 |
|---|------|------|
| 1 | 榜单范围 | **全做**（消费、充值、Token、请求数、勤劳，5个榜全上） |
| 2 | 充值榜数据来源 | `redeem_codes` 表，`type IN ('balance','admin_balance') AND value > 0`，已有 `SumPositiveBalanceByUser` 方法可复用，排行榜需新增按用户分组的 Top N 查询 |
| 3 | "差一点就超过" 提示 | **做**（后端返回 `next_rank_gap` + `next_rank_email` 字段） |
| 4 | 排名变动通知 | **做，页面内上下箭头**（前端本地缓存上次排名，对比显示 ⬆️/⬇️ 箭头即可） |
| 5 | 称号/徽章 | **做，先显示在排行榜页面**（后续可扩展到用户资料） |
| 6 | 隐身选项 | **不加**（所有用户都上榜，邮箱已脱敏足够保护隐私） |
| 7 | 管理员排除 | **不排除**（管理员也参与排名） |

## 八、技术实现参考

### 8.1 已有可复用基础设施

| 现有能力 | 位置 | 说明 |
|---------|------|------|
| Top N 用户聚合查询 | `usage_log_repo.go` `GetUserSpendingRanking()` | 按 actual_cost/requests/tokens 聚合 |
| 充值总额查询 | `redeem_code_repo.go` `SumPositiveBalanceByUser()` | 单用户充值总额，排行榜需扩展为 Top N |
| 用户使用趋势类型 | `usagestats/usage_log_types.go` | `UserSpendingRankingItem` 等现有类型 |
| 时间范围工具 | `timezone.Today()` 等 | 已有日/周/月时间计算 |
| 用户端路由 | `routes/user.go` | `authenticated.Group("/usage")` |
| 前端侧边栏 | `AppSidebar.vue` | `userNavItems` computed |

### 8.2 数据库字段（usage_logs 表）

- `actual_cost`: 实际扣费金额（消费榜排序字段）
- `total_cost`: 标准价格（参考）
- `input_tokens`, `output_tokens`, `cache_creation_tokens`, `cache_read_tokens`: Token 相关
- `created_at`: 时间戳（时间维度过滤 + 勤劳榜计算）
- `user_id`: 用户关联

### 8.3 邮箱脱敏

```go
// "xiaobei@gmail.com" → "xi**@gma..."
func maskEmail(email string) string {
    parts := strings.SplitN(email, "@", 2)
    if len(parts) != 2 { return "***" }
    name, domain := parts[0], parts[1]
    if len(name) > 2 { name = name[:2] + "**" }
    if len(domain) > 3 { domain = domain[:3] + "..." }
    return name + "@" + domain
}
```

## 九、开发计划（需求已确认，按顺序执行）

> 全部功能一次性开发，不分批。按依赖顺序排列。

### 后端开发
- [ ] 1. 新增 `LeaderboardEntry` + `LeaderboardResponse` 类型（`usagestats/usage_log_types.go`）
- [ ] 2. 新增 `GetLeaderboard()` repo 方法 — 支持5种 type（cost/tokens/requests/active_days），基于 `usage_logs`（`usage_log_repo.go`）
- [ ] 3. 新增 `GetRechargeLeaderboard()` repo 方法 — 基于 `redeem_codes` 表（`redeem_code_repo.go`）
- [ ] 4. 新增 `GetUserRank()` repo 方法 — 获取当前用户排名 + `next_rank_gap`
- [ ] 5. 新增 `GetLeaderboard()` service 方法 — 含邮箱脱敏 + 称号计算（`account_usage_service.go`）
- [ ] 6. 新增 `Leaderboard()` handler（`usage_handler.go`）
- [ ] 7. 注册路由 `usage.GET("/leaderboard", ...)`（`routes/user.go`）
- [ ] 8. 补全所有 test stub 的新 interface 方法（⚠️ 重要，搜索所有 Stub/Mock）

### 前端开发
- [ ] 9. 新增 API 调用函数 `getLeaderboard(type, period, limit)`
- [ ] 10. 新建 `LeaderboardView.vue` 页面（奖台 + 列表 + 称号 + 箭头 + "差一点就超过" + "我的排名"）
- [ ] 11. 新增路由 `/leaderboard`
- [ ] 12. 侧边栏加入口（`userNavItems`）
- [ ] 13. 新增 i18n 翻译（zh + en）
- [ ] 14. 排名变动箭头（前端 localStorage 缓存上次排名，对比显示 ⬆️⬇️）

### 验证
- [ ] 15. 后端单元测试通过 `go test -tags=unit ./...`
- [ ] 16. Lint 检查通过 `golangci-lint run ./...`
- [ ] 17. 前端页面视觉验证（5个榜切换、4个时间段、称号显示、箭头、"差一点就超过"）
- [ ] 18. 邮箱脱敏效果确认

### 部署
- [ ] 19. 合并到 `bmai-sub2api-my-production20260329`
- [ ] 20. 打 tag `bmai-v0.1.105.2`，推送触发 CI
- [ ] 21. VPS 一键安装验证

---

> 📌 此文档可带到其他机器继续开发。需求已全部确认，直接开干。

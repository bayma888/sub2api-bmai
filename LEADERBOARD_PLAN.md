# 消耗排行榜功能实现计划

> 创建日期: 2026-03-29
> 状态: 待审批

## 一、功能概述

在用户端新增**消耗排行榜**页面（`/leaderboard`），让所有登录用户都能看到消费排名，刺激用户消费。

### 设计参考（Master 提供的截图）
- 标题："消耗排行榜" 副标题："看看谁是隐王"
- 时间切换 Tab：今天 | 昨天 | 最近3天 | 最近7天
- Top 3 奖台展示：金银铜冠 + 头像圆圈 + 脱敏邮箱 + 消费金额 + 请求次数
- 第4名起列表展示：排名 + 头像 + 脱敏邮箱 + 请求次数 + 消费金额 + Token 数

## 二、技术方案

### 现有基础设施（可复用）
| 现有能力 | 位置 | 说明 |
|---------|------|------|
| Top N 用户 SQL 查询 | `usage_log_repo.go:1098` | `GetUserUsageTrend()` 的 `WITH top_users` CTE |
| 用户使用趋势类型 | `usagestats/usage_log_types.go:95` | `UserUsageTrendPoint` (UserID, Email, Requests, Tokens, Cost, ActualCost) |
| 用户端路由注册 | `routes/user.go` | `authenticated.Group("/usage")` 下已有 dashboard 系列 API |
| 前端侧边栏 | `AppSidebar.vue:516` | `userNavItems` computed，可直接加新条目 |

### 新增内容一览

#### 后端（4个文件改动，1个新增类型）

| # | 文件 | 改动 | 说明 |
|---|------|------|------|
| 1 | `internal/pkg/usagestats/usage_log_types.go` | **新增** `LeaderboardEntry` 类型 | `Rank, UserID, Email, MaskedEmail, Requests, Tokens, ActualCost` |
| 2 | `internal/repository/usage_log_repo.go` | **新增** `GetLeaderboard()` 方法 | 基于现有 `top_users` CTE 改造，新增邮箱脱敏、排名序号 |
| 3 | `internal/service/account_usage_service.go` | **新增** interface 方法 + service 实现 | `GetLeaderboard(ctx, period, limit)` |
| 4 | `internal/handler/usage_handler.go` | **新增** `Leaderboard()` handler | `GET /api/v1/usage/leaderboard?period=today&limit=20` |
| 5 | `internal/server/routes/user.go` | **新增** 一行路由注册 | `usage.GET("/leaderboard", h.Usage.Leaderboard)` |

#### 前端（4个文件改动，1个新增页面）

| # | 文件 | 改动 | 说明 |
|---|------|------|------|
| 1 | `src/views/user/LeaderboardView.vue` | **新建** | 排行榜主页面（奖台 + 列表） |
| 2 | `src/api/usage.ts` (或同级) | **新增** | `getLeaderboard(period, limit)` API 调用 |
| 3 | `src/router/index.ts` | **新增** | `/leaderboard` 路由 |
| 4 | `src/components/layout/AppSidebar.vue` | **修改** | `userNavItems` 加排行榜入口 |
| 5 | `src/i18n/locales/zh.ts` + `en.ts` | **修改** | 新增排行榜相关翻译 |

## 三、详细设计

### 3.1 后端 API

```
GET /api/v1/usage/leaderboard?period=today&limit=20
```

**参数：**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| period | string | today | `today` / `yesterday` / `3days` / `7days` |
| limit | int | 20 | 返回前 N 名用户 |

**响应：**
```json
{
  "data": {
    "period": "today",
    "items": [
      {
        "rank": 1,
        "user_id": 42,
        "masked_email": "大**",
        "requests": 2413,
        "tokens": 523000000,
        "actual_cost": 274.81
      }
    ],
    "my_rank": {
      "rank": 15,
      "requests": 120,
      "tokens": 15000000,
      "actual_cost": 12.50
    }
  }
}
```

**关键点：**
- `my_rank`: 当前用户自己的排名（即使不在 Top N 中也返回）
- `masked_email`: 后端做邮箱脱敏（`x**@gmail....`），前端不处理明文

### 3.2 SQL 查询（基于现有 `top_users` CTE）

```sql
WITH ranked_users AS (
    SELECT
        user_id,
        COUNT(*) as requests,
        COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) as tokens,
        COALESCE(SUM(actual_cost), 0) as actual_cost
    FROM usage_logs
    WHERE created_at >= $1 AND created_at < $2
    GROUP BY user_id
    ORDER BY actual_cost DESC
    LIMIT $3
)
SELECT
    ROW_NUMBER() OVER (ORDER BY r.actual_cost DESC) as rank,
    r.user_id,
    COALESCE(u.email, '') as email,
    r.requests,
    r.tokens,
    r.actual_cost
FROM ranked_users r
LEFT JOIN users u ON r.user_id = u.id
ORDER BY rank ASC
```

### 3.3 邮箱脱敏逻辑（后端 Go）

```go
// maskEmail 脱敏邮箱: "xiaobei@gmail.com" → "xi***@gma..."
func maskEmail(email string) string {
    parts := strings.SplitN(email, "@", 2)
    if len(parts) != 2 {
        return "***"
    }
    name := parts[0]
    domain := parts[1]

    // 用户名：保留前2字符
    if len(name) > 2 {
        name = name[:2] + "**"
    }
    // 域名：保留前3字符
    if len(domain) > 3 {
        domain = domain[:3] + "..."
    }
    return name + "@" + domain
}
```

### 3.4 时间范围计算

| period | 起始时间 | 结束时间 |
|--------|---------|---------|
| today | 今天 00:00 | now |
| yesterday | 昨天 00:00 | 今天 00:00 |
| 3days | 3天前 00:00 | now |
| 7days | 7天前 00:00 | now |

### 3.5 前端页面结构

```
LeaderboardView.vue
├── 页面标题 "消耗排行榜" + 副标题 "看看谁是隐王"
├── 时间 Tab 切换 (今天|昨天|最近3天|最近7天)
├── Top 3 奖台区域
│   ├── 第2名（左） - 银冠 + 圆形头像框 + 脱敏邮箱 + 金额 + 请求数
│   ├── 第1名（中央，最大） - 金冠 + 大圆形头像框 + 脱敏邮箱 + 金额 + 请求数
│   └── 第3名（右） - 铜冠 + 圆形头像框 + 脱敏邮箱 + 金额 + 请求数
├── 分隔线
├── 第4~20名列表
│   └── 每行: 排名序号 + 头像字母圆 + 脱敏邮箱 + 请求次数 + 金额 + Token数
└── 底部: "你的排名" 卡片（如果当前用户不在列表中）
```

**头像说明：** 用户没有真实头像，用邮箱首字母生成彩色圆圈（类似截图中的 X、H）

### 3.6 缓存策略

- 后端用 Redis 缓存，key: `leaderboard:{period}`，TTL: **5分钟**
- 排行榜不需要实时性，5分钟更新一次足够
- `my_rank` 不缓存（每个用户不同）

## 四、Todo 清单

### 后端
- [ ] 1. 新增 `LeaderboardEntry` + `LeaderboardResponse` 类型
- [ ] 2. 新增 `GetLeaderboard()` repo 方法 + `GetUserRank()` 方法
- [ ] 3. 新增 `GetLeaderboard()` service 方法（含邮箱脱敏）
- [ ] 4. 新增 `Leaderboard()` handler
- [ ] 5. 注册路由 `usage.GET("/leaderboard", ...)`
- [ ] 6. 补全所有 test stub 的新 interface 方法（⚠️ 重要）

### 前端
- [ ] 7. 新建 `LeaderboardView.vue` 页面
- [ ] 8. 新增 API 调用函数
- [ ] 9. 新增路由 `/leaderboard`
- [ ] 10. 侧边栏加入口（userNavItems + personalNavItems）
- [ ] 11. 新增 i18n 翻译（zh + en）

### 验证
- [ ] 12. 后端单元测试通过
- [ ] 13. 前端页面视觉验证
- [ ] 14. 邮箱脱敏效果确认

## 五、风险 & 注意事项

| 风险 | 应对 |
|------|------|
| interface 改动导致 test stub 编译失败 | 第6步必须搜索所有 stub/mock 补全 |
| 隐私问题：用户可能不想被看到 | 邮箱脱敏后端强制执行，前端不接触明文 |
| 大量用户时 SQL 性能 | usage_logs 表已有 `created_at` 索引，top N 查询性能 OK |
| 上游合并可能性 | 低（隐私争议），建议作为 fork 特色功能 |

## 六、不做的事情（Scope Out）

- ❌ 管理员开关（先做核心功能，开关后续加）
- ❌ 用户头像上传（用首字母生成即可）
- ❌ Token 排行和请求数排行切换（先只按消费金额排）
- ❌ Redis 缓存（先不加，数据量不大直接查 DB，后续优化）

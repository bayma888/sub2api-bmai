# Feature F2: 管理员查看用户余额/并发变动记录 - 实现总结

**完成日期**: 2026-02-03
**分支**: `feature/admin-user-balance-history`
**Commit**: `606e29d`
**Status**: ✅ Ready for PR Review

---

## 功能概述

管理员可以点击用户列表中的余额金额，打开弹框查看该用户所有的余额/并发变动历史记录，包括：
- 兑换码充值/扣除
- 管理员手动调整
- 订阅变动

弹框内可快速进行充值/退款操作，无需返回列表。

---

## 设计理念

### 1. 用户交互流程
```
用户列表 → 点击余额($xxx) → 打开变动记录弹框
                          ├── 查看历史记录（支持类型筛选、分页）
                          ├── 点击"充值"按钮 → 打开充值弹框（在上层）
                          └── 返回历史记录继续查看
```

### 2. 关键设计决策

| 决策项 | 选择 | 原因 |
|--------|------|------|
| 总充值计算 | SQL聚合查询 | 性能优于内存遍历，不受分页影响 |
| 充值/退款逻辑 | 复用现有实现 | 金钱操作须谨慎，复用确保安全 |
| 弹框层级 | history=40, action=50 | 确保充值弹框在历史上方，避免混淆 |
| Tooltip实现 | CSS即时显示 | 原生title有延迟，CSS duration-75快速响应 |
| 按钮样式 | 与菜单统一 | 视觉一致性提升用户体验 |

---

## 技术实现

### 后端架构

```
Repository Layer
  ├── ListByUserPaginated(userID, pagination, type) → []RedeemCode + PaginationResult
  └── SumPositiveBalanceByUser(userID) → float64

Service Layer
  └── GetUserBalanceHistory(userID, page, pageSize, type)
      → ([]RedeemCode, total, totalRecharged, error)

Handler Layer
  └── GET /admin/users/:id/balance-history?page=1&page_size=15&type=balance
      Response: {
        items: BalanceHistoryItem[],
        total: int,
        page: int,
        page_size: int,
        pages: int,
        total_recharged: float64
      }
```

**SQL聚合查询** (in `redeem_code_repo.go`):
```go
func (r *redeemCodeRepository) SumPositiveBalanceByUser(ctx context.Context, userID int64) (float64, error) {
    var result []struct {
        Sum float64 `json:"sum"`
    }
    err := r.client.RedeemCode.Query().
        Where(
            redeemcode.UsedByEQ(userID),
            redeemcode.ValueGT(0),
            redeemcode.TypeIn("balance", "admin_balance"),
        ).
        Aggregate(dbent.As(dbent.Sum(redeemcode.FieldValue), "sum")).
        Scan(ctx, &result)
    if err != nil {
        return 0, err
    }
    if len(result) == 0 {
        return 0, nil
    }
    return result[0].Sum, nil
}
```

### 前端架构

```
UsersView (父组件)
  ├── 管理状态: balanceHistoryUser, showBalanceHistoryModal
  ├── 事件处理:
  │   ├── handleBalanceHistory(user) → 打开弹框
  │   ├── handleDepositFromHistory() → 转发到handleDeposit
  │   └── handleWithdrawFromHistory() → 转发到handleWithdraw
  └── 渲染:
      └── 用户列表
          └── balance列 (点击打开 + 快捷充值按钮)
      └── UserBalanceHistoryModal
          └── 事件监听: @deposit, @withdraw, @close

UserBalanceHistoryModal (子组件)
  ├── Props: show, user
  ├── Emits: close, deposit, withdraw
  ├── 状态: history[], loading, typeFilter, currentPage
  ├── 渲染:
  │   ├── Header (两行布局)
  │   │   ├── Row1: avatar + email/username + created_at (left) + balance (right)
  │   │   └── Row2: notes + total_recharged
  │   ├── Filter + Action buttons
  │   │   ├── Type filter dropdown
  │   │   ├── 充值按钮 (emerald) → emit('deposit')
  │   │   └── 退款按钮 (amber) → emit('withdraw')
  │   └── History list with pagination
  └── API调用: adminAPI.users.getUserBalanceHistory()
```

**组件通信流**:
```
UserBalanceHistoryModal
  ├── emit('deposit')  ↓
UsersView
  ├── handleDepositFromHistory()
  ├── 调用 handleDeposit(balanceHistoryUser)
  ├── 打开 UserBalanceModal (z-index=50)
  └── UserBalanceModal 显示在上层 (history z-index=40)
```

### API设计

**请求**:
```
GET /admin/users/:id/balance-history?page=1&page_size=15&type=balance
```

**参数**:
- `page`: 页码（从1开始）
- `page_size`: 每页条数（默认15）
- `type`: 类型筛选（可选），支持 balance/admin_balance/concurrency/admin_concurrency/subscription

**响应**:
```json
{
  "items": [
    {
      "id": 123,
      "code": "ABC123DEF456",
      "type": "balance",
      "value": 100.50,
      "status": "used",
      "notes": "管理员备注说明原因",
      "validity_days": 30,
      "created_at": "2026-02-01T10:00:00Z",
      "used_at": "2026-02-01T10:30:00Z",
      "used_by": 456,
      "group": {
        "id": 789,
        "name": "套餐A"
      }
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 15,
  "pages": 7,
  "total_recharged": 5000.50
}
```

---

## 文件修改清单

### 后端 (5个文件)

1. **`backend/internal/service/redeem_service.go`**
   - 添加 `SumPositiveBalanceByUser` 接口方法声明

2. **`backend/internal/repository/redeem_code_repo.go`**
   - 实现 `SumPositiveBalanceByUser` 使用SQL聚合
   - 优化 `ListByUserPaginated` 支持type筛选

3. **`backend/internal/service/admin_service.go`**
   - 修改 `GetUserBalanceHistory` 返回值包含totalRecharged
   - 同时调用pagination和aggregation

4. **`backend/internal/handler/admin/user_handler.go`**
   - 修改 GetUserBalanceHistory endpoint
   - 响应JSON包含 `total_recharged` 字段

5. **`backend/internal/server/routes/admin.go`**
   - 注册新路由 `GET /admin/users/:id/balance-history`

### 前端 (7个文件)

1. **`frontend/src/components/admin/user/UserBalanceHistoryModal.vue`** [新建]
   - 完整的历史记录弹框组件
   - 两行header设计
   - 类型筛选和分页

2. **`frontend/src/views/admin/UsersView.vue`**
   - 余额列改造：添加点击打开、tooltip、快捷充值按钮
   - 菜单图标改为dollar
   - 添加事件处理器：handleDepositFromHistory/handleWithdrawFromHistory

3. **`frontend/src/components/common/BaseDialog.vue`**
   - 新增 `zIndex` prop，支持自定义弹框层级
   - 默认值50，可通过 `:z-index="40"` 自定义

4. **`frontend/src/api/admin/users.ts`**
   - 新增 `BalanceHistoryResponse` 接口
   - 新增 `getUserBalanceHistory` 函数

5. **`frontend/src/api/admin/index.ts`**
   - 导出 `BalanceHistoryResponse` 和 `getUserBalanceHistory`

6. **`frontend/src/i18n/locales/zh.ts`**
   - 添加translations:
     - `balanceHistoryTip`: "点击打开充值记录"
     - `balanceHistoryTitle`: "用户充值和并发变动记录"
     - `totalRecharged`: "总充值"
     - `createdAt`: "创建时间"

7. **`frontend/src/i18n/locales/en.ts`**
   - 对应英文translations

---

## 关键实现细节

### 1. 总充值金额计算

**为什么使用SQL聚合而不是内存遍历?**
- 性能: 数据库计算vs应用层计算（尤其当用户有成千上万条记录）
- 一致性: 不受分页限制，无论查看第几页，总充值都是完整统计

**SQL聚合逻辑**:
```go
WHERE used_by = userID
  AND value > 0
  AND type IN ('balance', 'admin_balance')
```
- `value > 0`: 只统计充值（正值），不统计扣除（负值）
- `type IN ...`: 只统计资金类型，不含订阅和并发

### 2. 弹框层级控制

**问题**: 点击历史弹框中的"充值"打开充值弹框，但充值弹框在历史弹框后面

**解决**:
- BaseDialog添加 `zIndex` prop（默认50）
- BalanceHistoryModal设置 `:z-index="40"`
- 充值弹框（BalanceModal）使用默认 z-index=50
- 结果: 充值弹框z-50 > 历史z-40，正确显示在上层

**CSS实现**:
```css
.modal-overlay {
  @apply fixed inset-0 z-50;  /* 原有代码 */
}
/* BaseDialog使用inline style覆盖 */
:style="{ zIndex: 40 }"
```

### 3. Tooltip即时显示

**问题**: 原生 `title` 属性有500ms+延迟，用户体验不好

**解决**: CSS group-hover + opacity transition
```vue
<div class="group relative">
  <button>$100.00</button>
  <div class="opacity-0 group-hover:opacity-100 duration-75 transition-opacity">
    点击打开充值记录
  </div>
</div>
```
- `duration-75`: 75ms淡入，几乎是瞬间
- `pointer-events-none`: tooltip不阻挡鼠标交互

### 4. 事件复用模式

**目标**: 避免在Modal中重复实现充值/退款逻辑

**实现**:
```typescript
// Modal中
const emit = defineEmits(['close', 'deposit', 'withdraw'])

// 用户点击
@click="emit('deposit')" // 只发送事件

// 父组件中
<UserBalanceHistoryModal @deposit="handleDepositFromHistory" />

const handleDepositFromHistory = () => {
  if (balanceHistoryUser.value) {
    handleDeposit(balanceHistoryUser.value)  // 复用现有逻辑
  }
}
```

**优势**:
- 安全: 金钱操作逻辑集中，不易出错
- 易维护: 充值逻辑改动只需改一处
- 单一职责: Modal负责展示，UsersView负责业务逻辑

---

## 测试检查清单

- [ ] 打开变动记录弹框
  - [ ] 点击余额数字打开
  - [ ] 显示用户信息（email、username、created_at、余额、备注）
  - [ ] 显示总充值金额

- [ ] 类型筛选
  - [ ] "全部类型" 显示所有记录
  - [ ] 各种类型单独筛选正确
  - [ ] 切换类型自动重新加载第1页

- [ ] 分页
  - [ ] 显示正确的页码和总页数
  - [ ] 上一页/下一页按钮状态正确
  - [ ] 翻页加载数据正确

- [ ] 总充值计算
  - [ ] 与实际充值金额相符
  - [ ] 不受类型筛选影响（始终显示全量统计）
  - [ ] 不受分页影响

- [ ] 快捷充值/退款
  - [ ] 点击弹框内"充值"按钮打开充值弹框
  - [ ] 充值弹框显示在历史弹框上方
  - [ ] 充值完成后可返回历史弹框继续查看
  - [ ] 刷新后金额正确更新

- [ ] 用户体验
  - [ ] Tooltip即时显示（无明显延迟）
  - [ ] 点击弹框外自动关闭
  - [ ] 按钮样式与菜单保持一致
  - [ ] 暗色模式正常显示

- [ ] 国际化
  - [ ] 中文显示正确
  - [ ] 英文翻译完整

---

## Git提交信息

```
feat(admin): add user balance/concurrency history modal

- Add new API endpoint GET /admin/users/:id/balance-history with pagination and type filter
- Add SumPositiveBalanceByUser for calculating total recharged amount
- Create UserBalanceHistoryModal component with:
  - User info header (email, username, created_at, current balance, notes, total recharged)
  - Type filter dropdown (all/balance/admin_balance/concurrency/admin_concurrency/subscription)
  - Quick deposit/withdraw buttons
  - Paginated history list with icons and colored values
- Add instant tooltip on balance column for better UX
- Add z-index prop to BaseDialog for modal stacking control
- Update i18n translations (zh/en)
```

---

## 后续改进建议

1. **性能优化**: 考虑Redis缓存总充值金额（如果用户数据量巨大）
2. **导出功能**: 支持导出用户的变动记录为CSV
3. **高级筛选**: 按日期范围、金额范围筛选
4. **操作日志**: 记录admin的充值/退款操作时间和原因
5. **批量操作**: 支持批量查看多个用户的余额

---

**文档完成时间**: 2026-02-03 15:30 UTC+8

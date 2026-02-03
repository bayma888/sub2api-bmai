# Sub2API 功能增强计划 - 严谨版

> **重要提示**: 此文档涉及运营资金相关功能，所有改动必须经过严格审查后才能提交PR

---

## 🎯 进度总结 (2026-02-02)

| 功能 | 状态 | PR | 完成时间 |
|------|------|-----|---------|
| **F3** | ✅ 已完成 | [#449](https://github.com/Wei-Shaw/sub2api/pull/449) | 2026-02-02 |
| **F4** | ✅ 已完成 | [#450](https://github.com/Wei-Shaw/sub2api/pull/450) | 2026-02-02 |
| **F2** | 🔄 开发中 | - | - |
| **F1** | ⏳ 计划中 | - | - |

---

## 一、需求清单

| 序号 | 功能 | 优先级 | 复杂度 | 状态 |
|------|------|--------|--------|------|
| **F1** | API Key独立配额 | 高 | 中 | ⏳ 待开发 |
| **F2** | 管理员查看用户充值记录 | 高 | 低 | 🔄 开发中 |
| **F3** | 用户搜索支持备注模糊查询 | 中 | 低 | ✅ 已完成 |
| **F4** | 用户端充值记录显示备注 | 中 | 低 | ✅ 已完成 |

---

## 二、现有代码分析

### 2.1 关键发现

#### 余额变动记录机制
**现状**: Sub2API **没有**独立的`balance_log`表，而是**复用`redeem_codes`表**作为余额变动记录：

```go
// backend/internal/service/admin_service.go:479-500
// 管理员给用户充值时，会创建一条type=admin_balance的redeem_code记录
adjustmentRecord := &RedeemCode{
    Code:   code,                         // 自动生成的唯一码
    Type:   AdjustmentTypeAdminBalance,   // "admin_balance"
    Value:  balanceDiff,                  // 变动金额（正/负）
    Status: StatusUsed,                   // 直接标记为已使用
    UsedBy: &user.ID,                     // 用户ID
    Notes:  notes,                        // 管理员备注 ✅ 已支持
}
```

#### 用户搜索机制
**现状**: 仅支持email和username搜索，**不支持notes**：

```go
// backend/internal/repository/user_repo.go:188-194
if filters.Search != "" {
    q = q.Where(
        dbuser.Or(
            dbuser.EmailContainsFold(filters.Search),
            dbuser.UsernameContainsFold(filters.Search),
            // ❌ 缺少: dbuser.NotesContainsFold(filters.Search)
        ),
    )
}
```

#### 用户端充值记录
**现状**: `RedeemCode` DTO **没有暴露notes字段**给普通用户：

```go
// backend/internal/handler/dto/types.go:188-203
type RedeemCode struct {
    // ... 其他字段
    // ❌ 没有 Notes 字段（只有 AdminRedeemCode 有）
}

type AdminRedeemCode struct {
    RedeemCode
    Notes string `json:"notes"` // ✅ 管理员DTO有
}
```

#### API Key结构
**现状**: API Key **没有配额相关字段**：

```go
// backend/ent/schema/api_key.go:33-55
func (APIKey) Fields() []ent.Field {
    return []ent.Field{
        field.Int64("user_id"),
        field.String("key"),
        field.String("name"),
        field.Int64("group_id").Optional(),
        field.String("status"),
        field.JSON("ip_whitelist", []string{}),
        field.JSON("ip_blacklist", []string{}),
        // ❌ 没有 quota, quota_used, quota_type 字段
    }
}
```

---

## 三、详细实现计划

---

### F1: API Key 独立配额功能

#### 3.1.1 需求说明
- 用户创建API Key时可设置配额（tokens或USD）
- 请求前检查配额是否用完
- 用完后立即拒绝该密钥的所有请求
- 支持管理员和用户查看/修改配额

#### 3.1.2 数据库改动

**文件**: `backend/ent/schema/api_key.go`

```go
func (APIKey) Fields() []ent.Field {
    return []ent.Field{
        // ... 现有字段保持不变 ...

        // ========== 新增配额字段 ==========

        // 配额限制值（0表示无限制）
        field.Float("quota").
            SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
            Default(0).
            Comment("Quota limit for this API key (0 = unlimited)"),

        // 已使用配额
        field.Float("quota_used").
            SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
            Default(0).
            Comment("Used quota amount"),

        // 配额类型：tokens 或 usd
        field.String("quota_type").
            MaxLen(20).
            Default("usd").
            Comment("Quota type: tokens or usd"),
    }
}

func (APIKey) Indexes() []ent.Index {
    return []ent.Index{
        // ... 现有索引 ...

        // 新增：用于快速查询配额即将用完的Key
        index.Fields("quota", "quota_used"),
    }
}
```

**迁移命令**:
```bash
cd backend
go generate ./ent
```

#### 3.1.3 Service层改动

**文件**: `backend/internal/service/api_key.go`

新增Model字段：
```go
type APIKey struct {
    ID          int64
    UserID      int64
    Key         string
    Name        string
    GroupID     *int64
    Status      string
    IPWhitelist []string
    IPBlacklist []string
    CreatedAt   time.Time
    UpdatedAt   time.Time

    // ========== 新增 ==========
    Quota      float64 // 配额限制
    QuotaUsed  float64 // 已使用配额
    QuotaType  string  // "tokens" 或 "usd"
}
```

**文件**: `backend/internal/service/api_key_service.go`

修改Request结构：
```go
type CreateAPIKeyRequest struct {
    Name        string
    GroupID     *int64
    CustomKey   *string
    IPWhitelist []string
    IPBlacklist []string

    // ========== 新增 ==========
    Quota     float64 `json:"quota"`      // 配额值（0=无限制）
    QuotaType string  `json:"quota_type"` // "tokens" 或 "usd"，默认"usd"
}

type UpdateAPIKeyRequest struct {
    Name        *string
    GroupID     *int64
    Status      *string
    IPWhitelist []string
    IPBlacklist []string

    // ========== 新增 ==========
    Quota     *float64 `json:"quota"`      // 可选更新配额
    QuotaType *string  `json:"quota_type"` // 可选更新类型
}
```

新增方法：
```go
var (
    ErrAPIKeyQuotaExhausted = infraerrors.PaymentRequired(
        "API_KEY_QUOTA_EXHAUSTED",
        "API key quota exhausted",
    )
)

// CheckQuota 检查API Key配额是否足够
// 返回nil表示可以继续，返回error表示配额不足
func (s *APIKeyService) CheckQuota(apiKey *APIKey) error {
    // 配额为0表示无限制
    if apiKey.Quota <= 0 {
        return nil
    }

    // 检查是否已用完
    if apiKey.QuotaUsed >= apiKey.Quota {
        return ErrAPIKeyQuotaExhausted
    }

    return nil
}

// UpdateQuotaUsage 更新API Key已使用的配额
// cost: 本次消耗的配额（根据quota_type，可能是tokens或usd）
func (s *APIKeyService) UpdateQuotaUsage(ctx context.Context, apiKeyID int64, cost float64) error {
    return s.apiKeyRepo.IncrementQuotaUsed(ctx, apiKeyID, cost)
}

// GetQuotaRemaining 获取剩余配额
func (s *APIKeyService) GetQuotaRemaining(apiKey *APIKey) float64 {
    if apiKey.Quota <= 0 {
        return -1 // -1表示无限制
    }
    remaining := apiKey.Quota - apiKey.QuotaUsed
    if remaining < 0 {
        return 0
    }
    return remaining
}
```

修改Create方法：
```go
func (s *APIKeyService) Create(ctx context.Context, userID int64, req CreateAPIKeyRequest) (*APIKey, error) {
    // ... 现有验证逻辑 ...

    // 验证配额类型
    quotaType := req.QuotaType
    if quotaType == "" {
        quotaType = "usd" // 默认使用USD
    }
    if quotaType != "tokens" && quotaType != "usd" {
        return nil, infraerrors.BadRequest("INVALID_QUOTA_TYPE", "quota_type must be 'tokens' or 'usd'")
    }

    // 验证配额值
    if req.Quota < 0 {
        return nil, infraerrors.BadRequest("INVALID_QUOTA", "quota must be >= 0")
    }

    apiKey := &APIKey{
        UserID:      userID,
        Key:         key,
        Name:        req.Name,
        GroupID:     req.GroupID,
        Status:      StatusActive,
        IPWhitelist: req.IPWhitelist,
        IPBlacklist: req.IPBlacklist,

        // ========== 新增 ==========
        Quota:      req.Quota,
        QuotaUsed:  0, // 新建时为0
        QuotaType:  quotaType,
    }

    // ... 后续逻辑 ...
}
```

#### 3.1.4 Repository层改动

**文件**: `backend/internal/service/api_key.go` (接口定义)

```go
type APIKeyRepository interface {
    // ... 现有方法 ...

    // ========== 新增 ==========
    // IncrementQuotaUsed 原子增加已使用配额
    IncrementQuotaUsed(ctx context.Context, id int64, amount float64) error

    // ResetQuotaUsed 重置已使用配额（管理员操作）
    ResetQuotaUsed(ctx context.Context, id int64) error
}
```

**文件**: `backend/internal/repository/api_key_repo.go`

```go
func (r *apiKeyRepository) IncrementQuotaUsed(ctx context.Context, id int64, amount float64) error {
    _, err := r.client.APIKey.UpdateOneID(id).
        AddQuotaUsed(amount).
        Save(ctx)
    return err
}

func (r *apiKeyRepository) ResetQuotaUsed(ctx context.Context, id int64) error {
    _, err := r.client.APIKey.UpdateOneID(id).
        SetQuotaUsed(0).
        Save(ctx)
    return err
}
```

#### 3.1.5 Gateway层改动

**文件**: `backend/internal/handler/gateway_handler.go`

在Messages方法开头添加配额检查：
```go
func (h *GatewayHandler) Messages(c *gin.Context) {
    // 获取认证信息（已有逻辑）
    apiKey := getAPIKeyFromContext(c)

    // ========== 新增：配额检查 ==========
    if err := h.apiKeyService.CheckQuota(apiKey); err != nil {
        response.ErrorFrom(c, err)
        return
    }

    // ... 后续处理逻辑 ...
}
```

#### 3.1.6 计费层改动

**文件**: `backend/internal/service/gateway_service.go` 或相关计费处理位置

在计费成功后更新配额：
```go
// 计费成功后
actualCost := costBreakdown.ActualCost

// ========== 新增：更新API Key配额 ==========
if apiKey.Quota > 0 {
    var costToDeduct float64
    if apiKey.QuotaType == "tokens" {
        // 如果配额类型是tokens，使用总token数
        costToDeduct = float64(usage.InputTokens + usage.OutputTokens)
    } else {
        // 默认使用USD成本
        costToDeduct = actualCost
    }

    if err := s.apiKeyService.UpdateQuotaUsage(ctx, apiKey.ID, costToDeduct); err != nil {
        // 记录错误但不中断请求（配额更新失败不应影响用户请求）
        log.Printf("[WARN] Failed to update API key quota: %v", err)
    }
}
```

#### 3.1.7 Handler层改动

**文件**: `backend/internal/handler/api_key_handler.go`

```go
type CreateAPIKeyRequest struct {
    Name        string   `json:"name" binding:"required"`
    GroupID     *int64   `json:"group_id"`
    CustomKey   *string  `json:"custom_key"`
    IPWhitelist []string `json:"ip_whitelist"`
    IPBlacklist []string `json:"ip_blacklist"`

    // ========== 新增 ==========
    Quota     float64 `json:"quota"`      // 配额值（0=无限制）
    QuotaType string  `json:"quota_type"` // "tokens" 或 "usd"
}

type UpdateAPIKeyRequest struct {
    Name        string   `json:"name"`
    GroupID     *int64   `json:"group_id"`
    Status      string   `json:"status" binding:"omitempty,oneof=active inactive"`
    IPWhitelist []string `json:"ip_whitelist"`
    IPBlacklist []string `json:"ip_blacklist"`

    // ========== 新增 ==========
    Quota     *float64 `json:"quota"`      // 可选更新配额
    QuotaType *string  `json:"quota_type"` // 可选更新类型
}
```

#### 3.1.8 DTO层改动

**文件**: `backend/internal/handler/dto/types.go`

```go
type APIKey struct {
    ID          int64     `json:"id"`
    UserID      int64     `json:"user_id"`
    Key         string    `json:"key"`
    Name        string    `json:"name"`
    GroupID     *int64    `json:"group_id"`
    Status      string    `json:"status"`
    IPWhitelist []string  `json:"ip_whitelist"`
    IPBlacklist []string  `json:"ip_blacklist"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // ========== 新增 ==========
    Quota     float64 `json:"quota"`      // 配额限制（0=无限制）
    QuotaUsed float64 `json:"quota_used"` // 已使用配额
    QuotaType string  `json:"quota_type"` // "tokens" 或 "usd"

    User  *User  `json:"user,omitempty"`
    Group *Group `json:"group,omitempty"`
}
```

#### 3.1.9 前端改动

**文件**: `frontend/src/views/user/APIKeyForm.vue` (或类似文件)

创建/编辑表单新增：
```vue
<template>
  <!-- 配额设置区域 -->
  <div class="form-group">
    <label class="input-label">{{ t('apiKey.quotaLimit') }}</label>
    <div class="flex gap-4">
      <div class="flex-1">
        <input
          v-model.number="form.quota"
          type="number"
          min="0"
          step="0.01"
          :placeholder="t('apiKey.quotaUnlimited')"
          class="input"
        />
        <p class="input-hint">{{ t('apiKey.quotaHint') }}</p>
      </div>
      <div class="w-32">
        <select v-model="form.quota_type" class="input">
          <option value="usd">USD</option>
          <option value="tokens">Tokens</option>
        </select>
      </div>
    </div>
  </div>
</template>

<script setup>
const form = reactive({
  name: '',
  group_id: null,
  quota: 0,      // 0表示无限制
  quota_type: 'usd',
  // ...
})
</script>
```

**文件**: `frontend/src/views/user/APIKeyList.vue` (或类似文件)

列表显示配额进度：
```vue
<template>
  <td>
    <div v-if="key.quota > 0" class="quota-progress">
      <div class="flex items-center gap-2">
        <div class="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
          <div
            class="h-full bg-primary-500 transition-all"
            :class="{ 'bg-red-500': quotaPercent(key) >= 90 }"
            :style="{ width: `${quotaPercent(key)}%` }"
          ></div>
        </div>
        <span class="text-xs text-gray-500">
          {{ formatQuota(key.quota_used, key.quota_type) }} /
          {{ formatQuota(key.quota, key.quota_type) }}
        </span>
      </div>
    </div>
    <span v-else class="text-gray-400">{{ t('apiKey.unlimited') }}</span>
  </td>
</template>

<script setup>
const quotaPercent = (key) => {
  if (key.quota <= 0) return 0
  return Math.min(100, (key.quota_used / key.quota) * 100)
}

const formatQuota = (value, type) => {
  if (type === 'usd') {
    return `$${value.toFixed(2)}`
  }
  return `${Math.round(value).toLocaleString()} tokens`
}
</script>
```

---

### F2: 管理员查看用户充值记录

#### 3.2.1 需求说明
- 在"用户管理"下的"充值"弹窗中，新增"充值记录"选项卡
- 显示该用户的所有金额变动记录（包括兑换码兑换、管理员充值/扣款）
- 显示时间、类型、金额、备注

#### 3.2.2 现有机制

**好消息**: 充值记录**已经存在**！存储在`redeem_codes`表中：
- `type = 'balance'`: 用户兑换余额码
- `type = 'admin_balance'`: 管理员调整余额
- `type = 'concurrency'`: 用户兑换并发码
- `type = 'admin_concurrency'`: 管理员调整并发

只需要添加**管理员查询接口**即可。

#### 3.2.3 后端改动

**文件**: `backend/internal/service/admin_service.go`

新增方法：
```go
// GetUserBalanceHistory 获取用户的余额变动历史
func (s *adminService) GetUserBalanceHistory(ctx context.Context, userID int64, page, pageSize int) ([]RedeemCode, int64, error) {
    // 查询该用户的所有余额相关记录（包括兑换和管理员调整）
    types := []string{
        RedeemTypeBalance,           // "balance"
        AdjustmentTypeAdminBalance,  // "admin_balance"
        RedeemTypeConcurrency,       // "concurrency" (可选，看需求)
        AdjustmentTypeAdminConcurrency, // "admin_concurrency" (可选)
    }

    codes, total, err := s.redeemCodeRepo.ListByUserAndTypes(ctx, userID, types, page, pageSize)
    if err != nil {
        return nil, 0, fmt.Errorf("list balance history: %w", err)
    }

    return codes, total, nil
}
```

**文件**: `backend/internal/repository/redeem_code_repo.go`

新增Repository方法：
```go
// ListByUserAndTypes 按用户ID和类型列表查询，支持分页
func (r *redeemCodeRepository) ListByUserAndTypes(
    ctx context.Context,
    userID int64,
    types []string,
    page, pageSize int,
) ([]service.RedeemCode, int64, error) {
    query := r.client.RedeemCode.Query().
        Where(
            dbredeemcode.UsedByEQ(userID),
            dbredeemcode.TypeIn(types...),
        ).
        Order(dbent.Desc(dbredeemcode.FieldUsedAt))

    total, err := query.Clone().Count(ctx)
    if err != nil {
        return nil, 0, err
    }

    offset := (page - 1) * pageSize
    codes, err := query.Offset(offset).Limit(pageSize).All(ctx)
    if err != nil {
        return nil, 0, err
    }

    result := make([]service.RedeemCode, len(codes))
    for i, c := range codes {
        result[i] = redeemCodeEntityToService(c)
    }

    return result, int64(total), nil
}
```

**文件**: `backend/internal/handler/admin/user_handler.go`

新增Handler：
```go
// GetUserBalanceHistory 获取用户余额变动历史
// GET /api/v1/admin/users/:id/balance-history
func (h *UserHandler) GetUserBalanceHistory(c *gin.Context) {
    userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        response.BadRequest(c, "Invalid user ID")
        return
    }

    page, pageSize := response.ParsePagination(c)

    history, total, err := h.adminService.GetUserBalanceHistory(c.Request.Context(), userID, page, pageSize)
    if err != nil {
        response.ErrorFrom(c, err)
        return
    }

    // 使用 AdminRedeemCode DTO（包含notes字段）
    out := make([]dto.AdminRedeemCode, len(history))
    for i := range history {
        out[i] = *dto.AdminRedeemCodeFromService(&history[i])
    }

    response.Paginated(c, out, total, page, pageSize)
}
```

**文件**: `backend/internal/server/routes/admin.go`

注册新路由：
```go
// 在 users 路由组中添加
users.GET("/:id/balance-history", userHandler.GetUserBalanceHistory)
```

#### 3.2.4 前端改动

**文件**: `frontend/src/components/admin/user/UserBalanceModal.vue`

在现有充值弹窗中新增Tab：
```vue
<template>
  <Modal :show="show" @close="close" size="lg">
    <div class="p-6">
      <!-- Tab切换 -->
      <div class="flex border-b border-gray-200 mb-4">
        <button
          @click="activeTab = 'recharge'"
          :class="['tab-btn', activeTab === 'recharge' && 'tab-btn-active']"
        >
          {{ t('admin.users.recharge') }}
        </button>
        <button
          @click="activeTab = 'history'"
          :class="['tab-btn', activeTab === 'history' && 'tab-btn-active']"
        >
          {{ t('admin.users.balanceHistory') }}
        </button>
      </div>

      <!-- 充值表单（现有） -->
      <div v-show="activeTab === 'recharge'">
        <!-- 保持原有的充值表单 -->
      </div>

      <!-- 充值记录（新增） -->
      <div v-show="activeTab === 'history'">
        <BalanceHistoryTable :user-id="userId" />
      </div>
    </div>
  </Modal>
</template>

<script setup>
const activeTab = ref('recharge')
</script>
```

**新文件**: `frontend/src/components/admin/user/BalanceHistoryTable.vue`

```vue
<template>
  <div>
    <div v-if="loading" class="flex justify-center py-8">
      <LoadingSpinner />
    </div>

    <table v-else-if="history.length > 0" class="table w-full">
      <thead>
        <tr>
          <th>{{ t('common.time') }}</th>
          <th>{{ t('common.type') }}</th>
          <th>{{ t('common.amount') }}</th>
          <th>{{ t('common.notes') }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in history" :key="item.id">
          <td>{{ formatDateTime(item.used_at) }}</td>
          <td>
            <span :class="getTypeBadgeClass(item.type)">
              {{ getTypeLabel(item.type) }}
            </span>
          </td>
          <td :class="item.value >= 0 ? 'text-green-600' : 'text-red-600'">
            {{ item.value >= 0 ? '+' : '' }}${{ item.value.toFixed(2) }}
          </td>
          <td class="text-gray-500 max-w-xs truncate" :title="item.notes">
            {{ item.notes || '-' }}
          </td>
        </tr>
      </tbody>
    </table>

    <div v-else class="text-center py-8 text-gray-500">
      {{ t('common.noData') }}
    </div>

    <!-- 分页 -->
    <Pagination
      v-if="total > pageSize"
      :current="page"
      :total="total"
      :page-size="pageSize"
      @change="loadHistory"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminAPI } from '@/api'

const props = defineProps<{ userId: number }>()

const history = ref([])
const loading = ref(true)
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const loadHistory = async (newPage = 1) => {
  loading.value = true
  page.value = newPage
  try {
    const res = await adminAPI.getUserBalanceHistory(props.userId, page.value, pageSize.value)
    history.value = res.data
    total.value = res.total
  } finally {
    loading.value = false
  }
}

const getTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    'balance': '兑换码充值',
    'admin_balance': '管理员调整',
    'concurrency': '兑换码并发',
    'admin_concurrency': '管理员调整并发',
  }
  return labels[type] || type
}

const getTypeBadgeClass = (type: string) => {
  if (type.startsWith('admin_')) return 'badge badge-warning'
  return 'badge badge-info'
}

onMounted(() => loadHistory())
</script>
```

**文件**: `frontend/src/api/admin.ts`

新增API方法：
```typescript
// 获取用户余额变动历史
async getUserBalanceHistory(userId: number, page: number, pageSize: number) {
  return request.get(`/api/v1/admin/users/${userId}/balance-history`, {
    params: { page, page_size: pageSize }
  })
}
```

---

### F3: 用户搜索支持备注模糊查询

#### 3.3.1 需求说明
- 管理员搜索用户时，除了email和username，还应该支持notes字段

#### 3.3.2 改动范围极小

**文件**: `backend/internal/repository/user_repo.go`

修改第188-194行：
```go
if filters.Search != "" {
    q = q.Where(
        dbuser.Or(
            dbuser.EmailContainsFold(filters.Search),
            dbuser.UsernameContainsFold(filters.Search),
            dbuser.NotesContainsFold(filters.Search), // ✅ 新增
        ),
    )
}
```

**就这一行改动！**

#### 3.3.3 前端改动

**可选**: 修改搜索框placeholder提示，告知用户可以搜索备注：

**文件**: `frontend/src/views/admin/UsersView.vue`

```vue
<input
  v-model="searchQuery"
  type="text"
  :placeholder="t('admin.users.searchUsersWithNotes')" <!-- 更新key -->
  class="input"
  @input="handleSearch"
/>
```

**文件**: `frontend/src/locales/zh-CN.json` (或对应语言文件)

```json
{
  "admin": {
    "users": {
      "searchUsersWithNotes": "搜索邮箱、用户名或备注..."
    }
  }
}
```

---

### F4: 用户端充值记录显示备注

#### 3.4.1 需求说明
- 用户在"兑换码"页面查看充值历史时，能看到管理员填写的备注
- 这样用户知道为什么被充值或扣款

#### 3.4.2 现状分析

目前`RedeemCode` DTO（普通用户用）**没有notes字段**，只有`AdminRedeemCode`有：

```go
type RedeemCode struct {
    // ... 没有 Notes
}

type AdminRedeemCode struct {
    RedeemCode
    Notes string `json:"notes"` // 只有管理员DTO有
}
```

这是**有意设计**，因为notes可能包含内部信息。但对于`admin_balance`类型的备注，应该让用户看到。

#### 3.4.3 安全考虑

**方案A**: 直接在`RedeemCode` DTO添加notes字段（简单但可能泄露兑换码的内部备注）

**方案B**: 只对`admin_balance`和`admin_concurrency`类型返回notes（更安全）

**推荐方案B**，实现如下：

#### 3.4.4 后端改动

**文件**: `backend/internal/handler/dto/types.go`

```go
type RedeemCode struct {
    ID        int64      `json:"id"`
    Code      string     `json:"code"`
    Type      string     `json:"type"`
    Value     float64    `json:"value"`
    Status    string     `json:"status"`
    UsedBy    *int64     `json:"used_by"`
    UsedAt    *time.Time `json:"used_at"`
    CreatedAt time.Time  `json:"created_at"`

    GroupID      *int64 `json:"group_id"`
    ValidityDays int    `json:"validity_days"`

    User  *User  `json:"user,omitempty"`
    Group *Group `json:"group,omitempty"`

    // ========== 新增 ==========
    // Notes 只有管理员调整类型才返回，普通兑换码不返回
    Notes *string `json:"notes,omitempty"`
}
```

**文件**: `backend/internal/handler/dto/mappers.go`

修改映射函数，只对admin类型返回notes：
```go
func RedeemCodeFromService(s *service.RedeemCode) *RedeemCode {
    if s == nil {
        return nil
    }

    out := &RedeemCode{
        ID:           s.ID,
        Code:         s.Code,
        Type:         s.Type,
        Value:        s.Value,
        Status:       s.Status,
        UsedBy:       s.UsedBy,
        UsedAt:       s.UsedAt,
        CreatedAt:    s.CreatedAt,
        GroupID:      s.GroupID,
        ValidityDays: s.ValidityDays,
    }

    // ========== 新增：只对管理员调整类型返回notes ==========
    if s.Type == service.AdjustmentTypeAdminBalance ||
       s.Type == service.AdjustmentTypeAdminConcurrency {
        if s.Notes != nil && *s.Notes != "" {
            out.Notes = s.Notes
        }
    }

    return out
}
```

#### 3.4.5 前端改动

**文件**: `frontend/src/views/user/RedeemView.vue`

在历史记录显示中添加备注：

```vue
<div
  v-for="item in history"
  :key="item.id"
  class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
>
  <div class="flex items-center gap-4">
    <!-- 图标（保持原有） -->
    <div :class="[...]">
      <Icon ... />
    </div>
    <div>
      <p class="text-sm font-medium text-gray-900 dark:text-white">
        {{ getHistoryItemTitle(item) }}
      </p>
      <p class="text-xs text-gray-500 dark:text-dark-400">
        {{ formatDateTime(item.used_at) }}
      </p>
      <!-- ========== 新增：显示备注 ========== -->
      <p
        v-if="item.notes"
        class="text-xs text-gray-400 dark:text-dark-500 mt-1"
      >
        {{ t('redeem.note') }}: {{ item.notes }}
      </p>
    </div>
  </div>
  <!-- 右侧金额（保持原有） -->
</div>
```

**文件**: `frontend/src/api/types.ts`

更新类型定义：
```typescript
export interface RedeemHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_at: string
  group?: { id: number; name: string }
  validity_days?: number
  notes?: string  // ✅ 新增
}
```

**文件**: `frontend/src/locales/zh-CN.json`

```json
{
  "redeem": {
    "note": "备注"
  }
}
```

---

## 四、实施顺序建议

按照**风险由低到高、依赖关系**排序：

### Phase 1: 低风险改动（可快速合并）
1. **F3 - 用户搜索支持备注** - 改动仅1行代码
2. **F4 - 用户端显示充值备注** - DTO和前端小改动

### Phase 2: 中等复杂度
3. **F2 - 管理员查看充值记录** - 新增API和前端组件

### Phase 3: 核心功能
4. **F1 - API Key配额** - 涉及数据库Schema、计费逻辑等核心改动

---

## 五、测试检查清单

### F1 - API Key配额测试
- [ ] 创建API Key时设置配额（tokens类型）
- [ ] 创建API Key时设置配额（usd类型）
- [ ] 创建API Key时配额为0（无限制）
- [ ] 发起请求，配额被正确扣除
- [ ] 配额用完后请求被正确拒绝（返回402 Payment Required）
- [ ] 更新配额后可继续使用
- [ ] 管理员可查看用户API Key的配额使用情况
- [ ] 前端正确显示配额进度条

### F2 - 充值记录测试
- [ ] 管理员可查看用户的余额变动历史
- [ ] 历史记录包含兑换码充值和管理员调整
- [ ] 分页功能正常
- [ ] 备注正确显示

### F3 - 搜索备注测试
- [ ] 按邮箱搜索正常
- [ ] 按用户名搜索正常
- [ ] 按备注搜索正常
- [ ] 混合关键词搜索正常

### F4 - 用户备注显示测试
- [ ] 用户兑换码充值不显示备注（可能没有）
- [ ] 管理员调整余额的记录显示备注
- [ ] 备注过长时正确截断/显示

---

## 六、PR提交建议

### PR拆分策略

建议拆分为**4个独立PR**，按照Phase顺序提交：

1. **PR #1: feat: support notes search in user list**
   - 改动文件：1个
   - 审查难度：低

2. **PR #2: feat: show notes in user redeem history**
   - 改动文件：3-4个
   - 审查难度：低

3. **PR #3: feat: add balance history view for admin**
   - 改动文件：5-6个
   - 审查难度：中

4. **PR #4: feat: add quota support for API keys**
   - 改动文件：10+个
   - 审查难度：高（建议仔细Review）

### Commit Message规范

```
feat(user): add notes search in user list filter

- Add notes field to search OR condition in ListWithFilters
- Users can now be found by searching their admin notes

Closes #xxx
```

```
feat(redeem): expose notes to user for admin adjustments

- Add notes field to RedeemCode DTO
- Only return notes for admin_balance and admin_concurrency types
- Update frontend to display notes in redeem history

Closes #xxx
```

---

## 六、F2详细实现方案（开发中）

### 6.1 功能需求

**目标**: 管理员可以在用户详情页查看某个用户的所有余额变动记录

**关键特性**:
- 分页显示余额变动记录
- 显示充值金额、类型、时间、备注
- 统计总充值、总消费、当前余额
- 区分兑换码充值和管理员手动调整
- 时间倒序排列

### 6.2 数据设计

**数据来源**: `redeem_codes` 表

```sql
-- 查询条件
WHERE used_by = ?
  AND type IN ('balance', 'admin_balance')  -- 只查余额相关
  AND status = 'used'
ORDER BY used_at DESC
```

### 6.3 后端改动

**文件**: `backend/internal/handler/admin/user_handler.go`

```go
// GetBalanceHistory 获取用户余额变动历史
// GET /api/v1/admin/users/:id/balance-history?page=1&page_size=20
func (h *UserHandler) GetBalanceHistory(c *gin.Context) {
    userID := c.Param("id")
    page := c.DefaultQuery("page", "1")
    pageSize := c.DefaultQuery("page_size", "20")

    // 调用Service获取记录
    history, total, err := h.redeemService.GetUserBalanceHistory(userID, page, pageSize)
    if err != nil {
        response.Error(c, err)
        return
    }

    // 获取用户当前余额
    user, _ := h.userService.GetByID(userID)

    response.Success(c, gin.H{
        "records": history,
        "total": total,
        "current_balance": user.Balance,
    })
}
```

**文件**: `backend/internal/service/redeem_service.go`

```go
// GetUserBalanceHistory 获取用户余额变动历史
func (s *RedeemService) GetUserBalanceHistory(userID int64, page, pageSize int) ([]RedeemCode, int64, error) {
    offset := (page - 1) * pageSize
    return s.redeemRepo.ListByUser(userID, []string{"balance", "admin_balance"}, offset, pageSize)
}
```

**文件**: `backend/internal/repository/redeem_repo.go`

```go
// ListByUser 查询用户的余额变动记录
func (r *redeemRepository) ListByUser(userID int64, types []string, offset, limit int) ([]RedeemCode, int64, error) {
    q := r.client.RedeemCode.Query().
        Where(
            dbredeemcode.UsedByEQ(userID),
            dbredeemcode.TypeIn(types...),
            dbredeemcode.StatusEQ("used"),
        ).
        Order(ent.Desc(dbredeemcode.FieldUsedAt))

    total := q.CountX(ctx)
    records := q.Offset(offset).Limit(limit).AllX(ctx)

    return records, total, nil
}
```

### 6.4 前端改动

**新建文件**: `frontend/src/components/admin/UserDetailModal.vue`

```vue
<template>
  <Modal v-model:open="visible" title="用户详情" width="900px">
    <Tabs v-model:active="activeTab">
      <!-- 基本信息Tab -->
      <Tab name="info" title="基本信息">...</Tab>

      <!-- API Keys Tab -->
      <Tab name="keys" title="API Keys">...</Tab>

      <!-- 💰 余额记录Tab（新增） -->
      <Tab name="balance" title="💰 余额记录">
        <!-- 统计卡片 -->
        <div class="grid grid-cols-3 gap-4 mb-6">
          <StatCard title="当前余额" :value="user.balance" type="balance" />
          <StatCard title="总充值" :value="totalRecharge" type="income" />
          <StatCard title="总消费" :value="totalDeduct" type="expense" />
        </div>

        <!-- 记录列表 -->
        <DataTable
          :columns="balanceColumns"
          :data="balanceHistory"
          :loading="balanceLoading"
          :pagination="pagination"
          @page-change="onPageChange"
        />
      </Tab>

      <!-- 使用记录Tab -->
      <Tab name="usage" title="使用记录">...</Tab>
    </Tabs>
  </Modal>
</template>

<script setup lang="ts">
const balanceColumns = [
  { key: 'used_at', label: '时间', formatter: formatDateTime },
  { key: 'type', label: '类型', formatter: formatType },
  { key: 'value', label: '金额', formatter: formatAmount },
  { key: 'notes', label: '备注', width: '200px' }
]

const formatType = (type) => {
  return type === 'admin_balance' ? '管理员调整' : '兑换码'
}

const formatAmount = (value) => {
  return value >= 0 ? `+$${value.toFixed(2)}` : `-$${Math.abs(value).toFixed(2)}`
}
</script>
```

**更新文件**: `frontend/src/api/admin/user.ts`

```typescript
// 获取用户余额记录
export async function getUserBalanceHistory(
  userId: number,
  page: number = 1,
  pageSize: number = 20
) {
  const { data } = await apiClient.get(
    `/admin/users/${userId}/balance-history`,
    { params: { page, page_size: pageSize } }
  )
  return data
}
```

### 6.5 路由配置

**文件**: `backend/internal/server/routes/admin.go`

```go
// 在admin路由中添加
apiV1.GET("/admin/users/:id/balance-history", h.userHandler.GetBalanceHistory)
```

### 6.6 改动总结

| 类型 | 文件 | 改动 |
|------|------|------|
| 后端Handler | `backend/internal/handler/admin/user_handler.go` | 新增GetBalanceHistory方法 |
| 后端Service | `backend/internal/service/redeem_service.go` | 新增GetUserBalanceHistory方法 |
| 后端Repository | `backend/internal/repository/redeem_repo.go` | 新增ListByUser方法 |
| 后端路由 | `backend/internal/server/routes/admin.go` | 新增路由配置 |
| 前端组件 | `frontend/src/components/admin/UserDetailModal.vue` | 新建/更新用户详情Modal |
| 前端API | `frontend/src/api/admin/user.ts` | 新增getUserBalanceHistory函数 |

---

## 七、风险提示

### F1 配额功能风险点

1. **并发竞争**: 多个请求同时消耗配额时，可能出现超额
   - 缓解方案：使用数据库原子操作`AddQuotaUsed()`

2. **计费失败**: 如果UpdateQuotaUsage失败，配额不会更新
   - 缓解方案：记录日志，后续可手动修复

3. **配额类型切换**: 用户将tokens改为usd时，quota_used如何处理？
   - 建议：切换类型时重置quota_used为0

### F2/F4 数据安全风险点

1. **Notes泄露**: 确保只对admin类型返回notes，普通兑换码的notes可能包含敏感信息
   - 已在方案中处理

---

**Master，以上是完整的实现计划。我可以按照这个顺序开始改代码吗？建议从F3（搜索备注）开始，因为它只需要改一行代码，风险最低，可以先验证开发流程。**

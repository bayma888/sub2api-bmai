# API Key配额功能实现计划

## 📋 需求总结
用户创建API Key时，可以为该密钥设置**单独的配额额度**。当该密钥的配额用完后，立即停止该密钥的所有请求。

## 🎯 改动范围

### **1. 数据库Schema层**
**文件**: `backend/ent/schema/api_key.go`

需要添加字段：
```go
// 配额相关字段
field.Float64("quota").
    Default(0).
    Comment("该API Key的配额限制（tokens或USD），0表示无限制"),
field.Float64("quota_used").
    Default(0).
    Comment("已使用的配额"),
field.String("quota_type").
    Default("tokens").
    Comment("配额类型：tokens 或 usd"),
```

### **2. 业务逻辑层 - APIKeyService**
**文件**: `backend/internal/service/api_key_service.go`

#### 2.1 创建Request DTO
修改 `CreateAPIKeyRequest` 结构体：
```go
type CreateAPIKeyRequest struct {
	Name        string
	GroupID     *int64
	CustomKey   *string
	IPWhitelist []string
	IPBlacklist []string

	// 新增配额字段
	Quota       float64 `json:"quota"`        // 配额值（默认0表示无限制）
	QuotaType   string  `json:"quota_type"`  // tokens 或 usd
}
```

#### 2.2 修改 Create 方法
在创建API Key时设置配额：
```go
apiKey := &APIKey{
	UserID:      userID,
	Key:         key,
	Name:        req.Name,
	GroupID:     req.GroupID,
	Status:      StatusActive,
	IPWhitelist: req.IPWhitelist,
	IPBlacklist: req.IPBlacklist,

	// 新增
	Quota:       req.Quota,
	QuotaUsed:   0,
	QuotaType:   req.QuotaType, // 默认 "tokens"
}
```

#### 2.3 修改 Update 方法
允许更新配额：
```go
type UpdateAPIKeyRequest struct {
	Name        *string
	GroupID     *int64
	Status      *string
	IPWhitelist []string
	IPBlacklist []string

	// 新增
	Quota     *float64 `json:"quota"`       // 可选更新配额
	QuotaType *string  `json:"quota_type"` // 可选更新配额类型
}

// 在Update方法中添加
if req.Quota != nil {
	apiKey.Quota = *req.Quota
}
if req.QuotaType != nil {
	apiKey.QuotaType = *req.QuotaType
}
```

#### 2.4 新增配额检查方法
```go
// CheckQuota 检查API Key的配额是否已用完
func (s *APIKeyService) CheckQuota(ctx context.Context, apiKey *APIKey) error {
	// 无配额限制时允许
	if apiKey.Quota <= 0 {
		return nil
	}

	// 配额已用完
	if apiKey.QuotaUsed >= apiKey.Quota {
		return infraerrors.BadRequest(
			"API_KEY_QUOTA_EXHAUSTED",
			fmt.Sprintf("API key quota exhausted. Used: %.2f, Limit: %.2f",
				apiKey.QuotaUsed, apiKey.Quota),
		)
	}

	return nil
}

// UpdateQuotaUsage 更新API Key已使用的配额
func (s *APIKeyService) UpdateQuotaUsage(ctx context.Context, apiKeyID int64, cost float64) error {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, apiKeyID)
	if err != nil {
		return err
	}

	apiKey.QuotaUsed += cost
	return s.apiKeyRepo.Update(ctx, apiKey)
}
```

### **3. 网关请求前置检查**
**文件**: `backend/internal/handler/gateway_handler.go` 或相关middleware

在处理请求之前检查配额：
```go
// 在Messages或CountTokens方法的开始处添加
apiKey, err := h.apiKeyService.GetByID(ctx, apiKeyID)
if err != nil {
	return err
}

// 检查配额
if err := h.apiKeyService.CheckQuota(ctx, apiKey); err != nil {
	response.BadRequest(c, err.Error())
	return
}
```

### **4. 计费逻辑层**
**文件**: `backend/internal/service/billing_service.go` 或 gateway_service.go

计算费用后立即更新配额：
```go
// 在记录使用日志后添加
cost := costBreakdown.ActualCost // 或根据quota_type计算cost

// 如果API Key有配额限制，更新已使用配额
if apiKey.Quota > 0 {
	if err := s.apiKeyService.UpdateQuotaUsage(ctx, apiKey.ID, cost); err != nil {
		log.Printf("Failed to update quota usage: %v", err)
		// 不中断请求，但记录错误
	}
}
```

### **5. 前端改动**

#### 5.1 API Key创建表单
**文件**: `frontend/src/views/user/APIKeyForm.vue` 或类似

```vue
<template>
  <div class="form-group">
    <label>配额限制</label>
    <div class="quota-container">
      <input
        v-model.number="form.quota"
        type="number"
        placeholder="0表示无限制"
        min="0"
        step="0.01"
      />
      <select v-model="form.quota_type">
        <option value="tokens">Tokens</option>
        <option value="usd">USD</option>
      </select>
    </div>
    <small>配额用完后该密钥将被自动停用</small>
  </div>
</template>

<script setup>
const form = reactive({
  name: '',
  quota: 0,
  quota_type: 'tokens',
  // ... 其他字段
})
</script>
```

#### 5.2 API Key列表展示
**文件**: `frontend/src/views/user/APIKeyList.vue` 或类似

添加配额进度列：
```vue
<table>
  <tr v-for="key in apiKeys" :key="key.id">
    <!-- 现有列 -->
    <td>
      <div class="quota-progress">
        <progress
          :value="key.quota_used"
          :max="key.quota"
          v-if="key.quota > 0"
        ></progress>
        <span v-if="key.quota > 0">
          {{ key.quota_used.toFixed(2) }} / {{ key.quota.toFixed(2) }}
          {{ key.quota_type === 'tokens' ? 'tokens' : 'USD' }}
        </span>
        <span v-else>无限制</span>
      </div>
    </td>
  </tr>
</table>
```

#### 5.3 API调用DTO
**文件**: `frontend/src/handler/dto/api_key.ts` 或类似

```typescript
export interface APIKey {
  id: number
  key: string
  name: string
  group_id?: number
  status: string
  quota: number          // 新增
  quota_used: number     // 新增
  quota_type: string     // 新增: 'tokens' | 'usd'
  ip_whitelist?: string[]
  ip_blacklist?: string[]
  created_at: string
  updated_at: string
}
```

---

## 🔄 完整流程演示

### 场景：用户创建API Key并设置500 tokens配额

**1. 前端请求**
```json
POST /api/v1/api-keys
{
  "name": "My API Key",
  "quota": 500,
  "quota_type": "tokens",
  "group_id": null
}
```

**2. 后端处理**
- Handler接收请求 → 调用APIKeyService.Create()
- Service验证并创建API Key记录 → 存入DB（quota=500, quota_used=0）
- 返回创建成功

**3. 用户发起请求**
```bash
curl -H "Authorization: Bearer sk-xxx" https://api.sub2api.com/v1/messages
```

**4. 网关处理请求**
- Middleware验证API Key
- Handler调用APIKeyService.CheckQuota()
- 检查：quota_used(100) >= quota(500)？ → 否，继续
- 调用上游API
- 返回响应给用户

**5. 计费流程**
- 解析响应，计算成本（如200 tokens消耗）
- 记录UsageLog
- 调用APIKeyService.UpdateQuotaUsage(id, 200)
- 更新DB：api_keys.quota_used = 100 + 200 = 300

**6. 配额耗尽**
- 当quota_used = 500时
- 下一个请求来到：CheckQuota() → 返回错误
- 用户收到：`API_KEY_QUOTA_EXHAUSTED` 错误
- 密钥自动无法使用，直到用户增加配额

---

## 📊 数据库迁移

运行以下命令生成Ent迁移：
```bash
cd backend
go generate ./ent
```

这会自动在 `ent/migrate` 目录生成迁移文件。

---

## ✅ 实现检查清单

- [ ] 更新 `api_key.go` Schema（添加quota字段）
- [ ] 更新 `CreateAPIKeyRequest` 结构体
- [ ] 更新 `UpdateAPIKeyRequest` 结构体
- [ ] 在APIKeyService.Create() 中处理配额
- [ ] 在APIKeyService.Update() 中处理配额更新
- [ ] 添加 CheckQuota() 方法
- [ ] 添加 UpdateQuotaUsage() 方法
- [ ] 在gateway_handler中添加CheckQuota()调用
- [ ] 在计费逻辑中调用UpdateQuotaUsage()
- [ ] 更新前端Create表单
- [ ] 更新前端List表格显示
- [ ] 更新前端DTO类型
- [ ] 生成Ent迁移: `go generate ./ent`
- [ ] 测试创建带配额的API Key
- [ ] 测试配额检查逻辑
- [ ] 测试配额更新逻辑

---

## 🚀 建议实现顺序

1. **先做数据库** - 更新Schema，生成迁移
2. **再做后端业务逻辑** - APIKeyService的方法
3. **再做网关检查** - 在请求处理中添加检查
4. **最后做前端** - UI表单和展示

这样可以边做边测试，充分验证逻辑。

---

## 💡 额外考虑

### 边界情况处理
- ✅ 配额为0时 → 无限制（已在CheckQuota中处理）
- ✅ 配额为负数时 → 视为无效，应该拒绝
- ✅ 更新配额时 → 可以增加也可以减少
- ✅ quota_used > quota的情况 → 已有的请求继续，新请求拒绝

### 性能优化
- 可以在APIKeyAuthCacheEntry中缓存配额信息
- 避免每次请求都查数据库
- 定期同步更新回DB（如每100个请求或每分钟）

### 监控和告警
- 配额即将用完时 → 给用户提示
- 配额已用完 → 在列表中高亮显示
- API返回 `X-Quota-Used` 和 `X-Quota-Limit` headers

Master，这就是完整的实现计划！你想从哪个部分开始？我可以逐个帮你改代码。

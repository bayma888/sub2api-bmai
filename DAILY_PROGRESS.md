# Sub2API 开发进度日志

## 📅 2026-02-04 工作总结

### ✅ 已完成

#### F1: API Key 独立配额和过期时间
- **状态**: ✅ 已完成并合并到上游
- **PR**: [#471](https://github.com/Wei-Shaw/sub2api/pull/471) - Merged
- **Branch**: `feature/api-key-quota-expiration`

**功能说明**:
- `quota`: 配额限制（USD），0 表示无限制
- `quota_used`: 已使用额度（USD）
- `expires_at`: 过期时间，null 表示永不过期

**后端改动**:
- 数据库迁移: `045_add_api_key_quota.sql`
- Ent Schema: `api_key.go` 添加 quota/quota_used/expires_at 字段
- 认证缓存: `APIKeyAuthSnapshot` 添加 Quota/QuotaUsed 字段
- 扣费逻辑: `gateway_service.go` 中 RecordUsage 支持 API Key 配额扣费
- Repository: `IncrementQuotaUsed` 原子更新配额

**前端改动**:
- `KeysView.vue`: 创建/编辑 API Key 时支持配额和过期时间设置
- DTO 更新: 添加 quota/quota_used/expires_at 字段

**CI 修复**:
- 修复测试 stub 缺少 `IncrementQuotaUsed` 方法
- 修复 gofmt 格式化问题
- 更新 API contract 测试期望值

---

## 📅 2026-02-03 工作总结

### ✅ 已完成

#### F2: 管理员查看用户余额/并发变动记录
- **状态**: ✅ 已完成并合并
- **Branch**: `feature/admin-user-balance-history`
- **改动文件**: 12个，588行新增

**API设计**:
```
GET /admin/users/:id/balance-history?page=1&page_size=15&type=balance
Response:
{
  "items": [...],
  "total": 100,
  "page": 1,
  "page_size": 15,
  "pages": 7,
  "total_recharged": 5000.50
}
```

---

## 📅 2026-02-02 工作总结

### ✅ 已完成

#### F3: 用户搜索支持备注字段
- **状态**: ✅ 已完成并合并
- **PR**: [#449](https://github.com/Wei-Shaw/sub2api/pull/449) - Merged

#### F4: 用户端显示管理员调整备注
- **状态**: ✅ 已完成并合并
- **PR**: [#450](https://github.com/Wei-Shaw/sub2api/pull/450) - Merged

---

## 🎯 功能完成状态

| 功能 | 状态 | PR |
|------|------|-----|
| F1 - API Key 独立配额 | ✅ 已合并 | #471 |
| F2 - 用户余额记录 | ✅ 已合并 | - |
| F3 - 搜索备注支持 | ✅ 已合并 | #449 |
| F4 - 用户端显示备注 | ✅ 已合并 | #450 |

---

## 🔧 项目信息

### 仓库结构
- **上游仓库**: `Wei-Shaw/sub2api`
- **Fork 仓库**: `bayma888/sub2api-bmai`
- **本地文档分支**: `local/dev-docs` (不提交到上游)

### 技术栈
- **后端**: Go 1.25.6 + Ent ORM + PostgreSQL + Redis
- **前端**: Vue 3 + TypeScript + TailwindCSS + Vite
- **端口**: 前端 3000, 后端 8080

### Git 配置
- **用户名**: bayma888
- **邮箱**: kubai666@126.com

### 开发工作流
```bash
# 热更新开发
cd frontend && pnpm install && pnpm dev

# 同步上游代码
git checkout main
git pull upstream main
git push origin main

# 恢复本地文档
git checkout local/dev-docs
```

---

## 📝 Claude Code 偏好设置

- **语言**: 回复使用中文
- **Git 提交**: 使用中文
- **称呼**: Master

---

**最后更新**: 2026-02-04

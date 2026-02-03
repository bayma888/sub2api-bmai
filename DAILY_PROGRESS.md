# Sub2API 开发进度日志

## 📅 2026-02-03 工作总结

### ✅ 已完成

#### F2: 管理员查看用户余额/并发变动记录
- **状态**: ✅ 已完成，待审核合并
- **Branch**: `feature/admin-user-balance-history`
- **Commit**: `606e29d`
- **改动文件**: 12个，588行新增

**后端改动**:
- `backend/internal/service/redeem_service.go`: 新增 `SumPositiveBalanceByUser` 接口
- `backend/internal/repository/redeem_code_repo.go`: 实现SQL聚合查询用户累计充值金额
- `backend/internal/service/admin_service.go`: GetUserBalanceHistory返回totalRecharged
- `backend/internal/handler/admin/user_handler.go`: 新API响应字段total_recharged
- `backend/internal/server/routes/admin.go`: 注册GET /admin/users/:id/balance-history路由

**前端改动**:
- `frontend/src/components/admin/user/UserBalanceHistoryModal.vue`: 新建变动记录弹框组件
  - 两行header设计（用户信息、余额、创建时间、备注、总充值）
  - 类型筛选dropdown（全部/余额/并发/订阅）
  - 快捷充值/退款按钮（样式与菜单保持一致）
  - 分页历史列表（带图标和彩色显示）
- `frontend/src/views/admin/UsersView.vue`:
  - 余额列添加点击打开记录功能
  - 添加即时tooltip提示"点击打开充值记录"
  - 余额列添加快捷充值按钮
  - 菜单"充值记录"图标改为dollar（金钱）
- `frontend/src/components/common/BaseDialog.vue`: 新增zIndex属性支持弹框层级控制
- `frontend/src/api/admin/users.ts`: 新增BalanceHistoryResponse接口，包含total_recharged
- i18n更新: 中英文添加balanceHistoryTip翻译

**UI优化**:
- 弹框宽度设为"wide"，显示更舒适
- 点击弹框外自动关闭（closeOnClickOutside=true）
- 余额tooltip采用CSS即时显示（duration-75）
- 充值/退款弹框z-index=50，变动记录弹框z-index=40，确保充值弹框在上面

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
- **改动**:
  - 文件: `backend/internal/repository/user_repo.go`
  - 添加: `dbuser.NotesContainsFold(filters.Search)` 到搜索条件
  - 效果: 管理员可以通过用户备注搜索用户

#### F4: 用户端显示管理员调整备注
- **状态**: ✅ 已完成并合并
- **PR**: [#450](https://github.com/Wei-Shaw/sub2api/pull/450) - Merged
- **改动**:
  - `backend/internal/handler/dto/types.go`: 为RedeemCode添加notes字段
  - `backend/internal/handler/dto/mappers.go`: 条件性填充notes（仅admin_balance/admin_concurrency）
  - `frontend/src/api/redeem.ts`: 添加notes到TypeScript接口
  - `frontend/src/views/user/RedeemView.vue`: 在历史记录中显示备注
  - 效果: 用户可以看到管理员调整余额的原因

### 🔄 进行中

#### F2: 管理员查看用户余额记录
- **状态**: ✅ 已完成开发，等待PR审核

### ⏳ 计划中

#### F1: API Key独立配额
- **状态**: ⏳ 待开发
- **复杂度**: 中等
- **影响**: 数据库Schema、Service层、Handler层、前端表单

---

## 🔧 技术调整

### 热更新开发环境
- **问题**: 每次前端改动都需要Docker重新构建，效率低下
- **解决**:
  - 设置Vite dev server at localhost:3000
  - 配置API代理指向Docker backend (localhost:8080)
  - 前端改动自动HMR (Hot Module Replacement)
  - `pnpm install && pnpm dev` 启动开发服务器
  - 显著提升开发效率：改动后<1秒自动更新

### Git配置修复
- **问题**: 之前的PR使用`xiaobei@example.com`邮箱，无法关联GitHub账号头像
- **解决**:
  - 更新git全局邮箱: `kubai666@126.com`
  - 更新git全局用户名: `bayma888`
  - 已应用于新分支，以后的PR会正确显示头像

---

## 📚 文档更新

- ✅ `DAILY_PROGRESS.md`: 添加2026-02-03工作总结
- ✅ `LOCAL_DEV_GUIDE.md`: Vite热更新开发指南

---

## 🎯 下一步计划

1. **F2 PR审核** (优先级: 高)
   - [ ] 创建PR并等待审核
   - [ ] 根据feedback进行调整
   - [ ] 合并到main分支

2. **F1规划** (优先级: 中)
   - [ ] 数据库Schema设计
   - [ ] 配额检查逻辑
   - [ ] 前端配额管理界面

---

## 📝 技术要点

### F2功能架构
- **分层设计**: Repository → Service → Handler → API
- **数据一致性**: 使用Ent ORM确保数据完整性
- **性能优化**: 使用SQL聚合而不是内存遍历计算总充值
- **UI/UX**: 即时tooltip + 快捷操作 + 合理弹框层级

### 前端最佳实践
- **组件复用**: 充值/退款逻辑由UsersView管理，BalanceHistoryModal仅发送事件
- **样式一致性**: 按钮样式与菜单保持统一
- **可访问性**: 使用BaseDialog的焦点管理和Esc关闭
- **i18n**: 中英文支持完整

### 开发工作流
- 使用Vite HMR进行快速迭代
- 关键改动后使用Docker镜像验证完整功能
- 分支命名规范: feature/xxxxx
- Commit信息详细，便于追踪

---

**最后更新**: 2026-02-03 15:00 UTC+8

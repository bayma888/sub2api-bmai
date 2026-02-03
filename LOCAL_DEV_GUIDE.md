# Sub2API 本地开发指南

## 📌 快速开始

### 当前状态
✅ **Docker容器已启动** - PostgreSQL + Redis + Sub2API完全运行
- PostgreSQL 18: `sub2api-postgres`
- Redis 8: `sub2api-redis`
- Sub2API: `sub2api` (http://localhost:8080)

### 访问Web界面
```
URL: http://localhost:8080
邮箱: admin@sub2api.local
密码: admin123456
```

---

## 🛠️ 本地开发工作流

### 场景1：修改后端代码（Go）

**当前状态：** Sub2API在Docker容器中运行

**如果要本地修改代码调试：**
1. 需要安装Go 1.25.6（当前未安装）
2. 本地运行 `go run ./cmd/server`
3. 连接到Docker中的PostgreSQL + Redis

**建议：** 暂时使用Docker版本（快速迭代），如果需要深度调试再安装Go

### 场景2：修改前端代码（Vue）

**当前状态：** 前端已内置在Docker镜像中

**如果要实时修改前端：**
```bash
cd frontend
pnpm install
pnpm run dev
# 会在 http://localhost:5173 启动开发服务器（热重载）
```

### 场景3：修改数据库Schema

1. 编辑：`backend/ent/schema/*.go`
2. 重生成：`cd backend && go generate ./ent`
3. 修改Docker-Compose配置后重启容器

---

## 📝 常用Docker命令

### 查看所有容器
```bash
docker-compose -f deploy/docker-compose.yml ps
```

### 查看Sub2API日志
```bash
docker-compose -f deploy/docker-compose.yml logs -f sub2api
```

### 查看PostgreSQL日志
```bash
docker-compose -f deploy/docker-compose.yml logs -f postgres
```

### 查看Redis日志
```bash
docker-compose -f deploy/docker-compose.yml logs -f redis
```

### 进入Sub2API容器
```bash
docker exec -it sub2api sh
```

### 进入PostgreSQL容器
```bash
docker exec -it sub2api-postgres psql -U sub2api -d sub2api
```

### 重启所有服务
```bash
cd deploy
docker-compose restart
```

### 停止所有服务（保留数据）
```bash
cd deploy
docker-compose down
```

### 删除所有数据重新开始
```bash
cd deploy
docker-compose down -v
```

---

## 🔧 配置文件

### Docker环境变量
**文件：** `deploy/.env`

修改后需要重启容器：
```bash
cd deploy
docker-compose restart sub2api
```

关键变量：
- `SERVER_MODE` - debug / release
- `ADMIN_EMAIL` - 管理员邮箱
- `ADMIN_PASSWORD` - 管理员密码
- `DATABASE_*` - PostgreSQL连接配置
- `REDIS_*` - Redis配置

### 本地配置文件（如使用本地Go）
**文件：** `backend/config.yaml`

复制模板：
```bash
cp deploy/config.example.yaml backend/config.yaml
```

修改连接到Docker数据库：
```yaml
database:
  host: "localhost"  # 改为 docker host IP 如果远程
  port: 5432
  user: "sub2api"
  password: "dev_password_123"

redis:
  host: "localhost"  # 改为 docker host IP 如果远程
  port: 6379
```

---

## 📊 架构总结

```
┌─────────────────────────────────────────────────┐
│           用户浏览器                            │
│         http://localhost:8080                   │
└──────────────┬──────────────────────────────────┘
               │
        ┌──────▼──────────────────┐
        │  Sub2API Docker容器     │
        │  (weishaw/sub2api)      │
        │  Port: 8080             │
        └──────┬─────────┬────────┘
               │         │
      ┌────────▼┐    ┌───▼────────┐
      │ PG 18   │    │ Redis 8    │
      │  :5432  │    │  :6379     │
      └─────────┘    └────────────┘
```

---

## 🎯 下一步

### 如果要修改功能：

1. **查看代码位置**
   - 核心业务逻辑：`backend/internal/service/`
   - HTTP处理：`backend/internal/handler/`
   - 数据库：`backend/ent/schema/`
   - 前端：`frontend/src/`

2. **修改代码**
   - 对于Go代码：建议安装Go本地运行（开发更快）
   - 对于前端：用 `pnpm run dev` 热重载开发

3. **测试修改**
   - Go修改：会自动编译提示错误
   - 前端修改：热重载自动刷新浏览器

4. **提交PR**
   ```bash
   git add .
   git commit -m "feat: your feature description"
   git push origin feature/your-branch
   # 然后在GitHub创建PR到 Wei-Shaw/sub2api
   ```

---

## 🚨 常见问题

### Q: 容器重启后数据会丢失吗？
**A:** 不会。Docker Compose使用命名卷持久化数据：
- `postgres_data` - 保存数据库
- `redis_data` - 保存Redis数据
- `sub2api_data` - 保存配置和应用数据

### Q: 如何访问数据库查询？
**A:** 进入PostgreSQL容器：
```bash
docker exec -it sub2api-postgres psql -U sub2api -d sub2api
# 然后输入SQL命令
```

### Q: 如何修改管理员密码？
**A:** 通过Web界面登录后，在设置中修改密码

### Q: 容器一直在Restarting怎么办？
**A:** 检查日志：
```bash
docker-compose logs sub2api
```

查看错误信息，通常是配置问题。最常见的是TOTP_ENCRYPTION_KEY格式不对。

---

## 📚 项目结构

```
sub2api/
├── backend/
│   ├── cmd/server/          # 入口点
│   ├── internal/
│   │   ├── handler/         # HTTP处理器
│   │   ├── service/         # 业务逻辑（40+个Service）
│   │   ├── repository/      # 数据访问层
│   │   └── config/          # 配置管理
│   ├── ent/
│   │   └── schema/          # 数据库Schema定义
│   └── go.mod               # Go依赖
│
├── frontend/
│   ├── src/
│   │   ├── views/           # 页面组件
│   │   ├── components/      # 可复用组件
│   │   ├── api/             # API调用
│   │   ├── stores/          # Pinia状态管理
│   │   └── router/          # 路由配置
│   └── package.json         # Node依赖
│
└── deploy/
    ├── docker-compose.yml   # Docker配置
    ├── .env                 # 环境变量
    └── config.example.yaml  # 配置模板
```

---

**Master, 现在Docker全部配置好了！你可以开始修改功能了。需要什么帮助？**

# sub2api-bmai 本地开发环境搭建指南

> 最后更新：2026-04-02
> 适用系统：Windows 10/11
> 用途：换电脑、重装系统后，照着这个文档一步步来就能跑起来

---

## 一、前置条件（需要安装的东西）

| 软件 | 版本 | 用途 | 安装方式 |
|------|------|------|----------|
| Go | 1.25.7+ | 后端编译运行 | https://go.dev/dl/ |
| Node.js | 18+ | 前端运行 | https://nodejs.org/ |
| pnpm | 最新版 | 前端包管理（不是 npm！） | `npm install -g pnpm` |
| PostgreSQL | 16 | 数据库 | https://www.postgresql.org/download/windows/ |
| Docker Desktop | 最新版 | 运行 Redis | https://www.docker.com/products/docker-desktop/ |
| Git | 最新版 | 代码管理 | https://git-scm.com/ |

---

## 二、数据库配置

### 2.1 PostgreSQL（Windows 原生服务，不是 Docker）

安装 PostgreSQL 16 后，默认会创建 Windows 服务 `postgresql-x64-16`。

**创建项目数据库和用户：**

```bash
# 用超级用户连接（密码：postgres）
"C:\Program Files\PostgreSQL\16\bin\psql.exe" -U postgres -h 127.0.0.1

# 在 psql 里执行：
CREATE USER sub2api WITH PASSWORD 'sub2api';
CREATE DATABASE sub2api OWNER sub2api;
GRANT ALL PRIVILEGES ON DATABASE sub2api TO sub2api;
\q
```

**验证连接：**
```bash
"C:\Program Files\PostgreSQL\16\bin\psql.exe" -U sub2api -h 127.0.0.1 -d sub2api
# 输入密码：sub2api
# 能进去就说明OK，输入 \q 退出
```

### 2.2 Redis（Docker 容器）

```bash
# 首次创建 Redis 容器（只需要执行一次）
docker run -d --name redis -p 6379:6379 redis:latest

# 以后每次启动只需要：
docker start redis

# 验证 Redis 是否运行：
docker ps | grep redis
# 应该看到 0.0.0.0:6379->6379/tcp
```

---

## 三、项目配置

### 3.1 克隆代码

```bash
git clone https://github.com/bayma888/sub2api-bmai.git
cd sub2api-bmai
```

### 3.2 后端配置文件

`backend/config.yaml` 已经在仓库里了，内容如下：

```yaml
server:
    host: 0.0.0.0
    port: 8080
    mode: release
database:
    host: localhost
    port: 5432
    user: sub2api
    password: sub2api
    dbname: sub2api
    sslmode: disable
redis:
    host: localhost
    port: 6379
    password: ""
    db: 0
jwt:
    secret: 7959746ed81c8622a71f1a96f277eb79c715869e89d534c108165e41bc77fd73
    expire_hour: 24
default:
    user_concurrency: 5
    user_balance: 0
    api_key_prefix: sk-
    rate_multiplier: 1
rate_limit:
    requests_per_minute: 60
    burst_size: 10
timezone: Asia/Shanghai
```

### 3.3 创建 .installed 锁文件

后端启动时会检测这个文件，没有的话会进入 setup 向导模式：

```bash
# 在 backend 目录下创建空文件
cd backend
echo "" > .installed
```

> **注意**：`.installed` 文件在 `.gitignore` 里，每台电脑都要手动创建一次

### 3.4 前端依赖安装

```bash
cd frontend
pnpm install
```

> **注意**：必须用 pnpm，不能用 npm！否则 lock 文件不一致会导致 CI 失败

---

## 四、启动服务（每次开发前执行）

### 快速启动流程（3个终端）

**终端1 — 启动 Redis：**
```bash
docker start redis
```

**终端2 — 启动后端：**
```bash
cd backend
export DATA_DIR="$(pwd)"
go run ./cmd/server/
```

> **关键**：必须设置 `DATA_DIR` 环境变量指向 backend 目录，否则会进入 setup 向导！

**终端3 — 启动前端：**
```bash
cd frontend
pnpm dev
```

### 启动成功标志

| 服务 | 成功标志 | 地址 |
|------|----------|------|
| PostgreSQL | Windows 服务显示"正在运行" | `127.0.0.1:5432` |
| Redis | `docker ps` 显示 redis 容器 Up | `127.0.0.1:6379` |
| 后端 | 日志显示 `Server started on 0.0.0.0:8080` | `http://localhost:8080` |
| 前端 | 显示 `VITE ready` + `Local: http://localhost:3000/` | `http://localhost:3000` |

### 访问网站

浏览器打开 **http://localhost:3000**

---

## 五、测试账号

| 项目 | 值 |
|------|-----|
| 邮箱 | `admin@126.com` |
| 密码 | `admin123` |
| 角色 | admin（可访问管理后台） |
| 用户 ID | 2 |

### 如果忘记密码，重置方法：

**第1步：生成新密码的 bcrypt hash：**
```bash
cd backend
cat > /tmp/gen_hash.go << 'EOF'
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("你的新密码"), 10)
    fmt.Println(string(hash))
}
EOF
go run /tmp/gen_hash.go
```

**第2步：写入 SQL 文件执行（不能直接命令行，$ 符号会被 shell 吃掉）：**
```bash
# 把上一步生成的 hash 写入 SQL 文件
cat > /tmp/reset_pwd.sql << 'EOF'
UPDATE users SET password_hash = '这里粘贴上一步生成的hash', updated_at = NOW() WHERE email = 'admin@126.com';
EOF

# 执行
PGPASSWORD=sub2api "C:/Program Files/PostgreSQL/16/bin/psql.exe" -U sub2api -h 127.0.0.1 -d sub2api -f /tmp/reset_pwd.sql
```

> **坑：** bcrypt hash 里有 `$` 符号，PowerShell/Bash 会把它当变量解析。一定要通过 SQL 文件执行，不能直接写在命令行里！

---

## 六、常见问题排查

### Q1: 后端启动后进入了 Setup 向导页面

**原因：** `backend/.installed` 文件不存在，或 `DATA_DIR` 没设置

**解决：**
```bash
cd backend
echo "" > .installed
export DATA_DIR="$(pwd)"
go run ./cmd/server/
```

### Q2: 后端报 `bind: Only one usage of each socket address`

**原因：** 8080 端口被占用（上次没关干净）

**解决：**
```bash
# 找到占用进程并杀掉
netstat -ano | grep ":8080 " | grep LISTENING
# 看最后一列的 PID，然后：
taskkill //PID 进程号 //F
```

### Q3: 前端报 `Port 3000 is in use`

**解决：** 同上，杀掉占用 3000 端口的进程

### Q4: Redis 连接失败

**解决：**
```bash
# 检查 Docker Desktop 是否启动
# 检查 redis 容器是否运行
docker ps | grep redis

# 如果没有，启动它
docker start redis

# 如果容器不存在，重新创建
docker run -d --name redis -p 6379:6379 redis:latest
```

### Q5: PostgreSQL 连接失败

**检查 Windows 服务是否运行：**
```powershell
Get-Service postgresql-x64-16
# 如果 Stopped，启动它：
Start-Service postgresql-x64-16
```

### Q6: psql 命令找不到

**解决：** 用完整路径：
```bash
"C:\Program Files\PostgreSQL\16\bin\psql.exe" -U sub2api -h 127.0.0.1 -d sub2api
```

> **注意：** 一定要用 `127.0.0.1`，不要用 `localhost`！Windows 上 localhost 可能走 IPv6 导致连不上。

---

## 七、数据库相关操作

```bash
# 连接数据库
PGPASSWORD=sub2api "C:/Program Files/PostgreSQL/16/bin/psql.exe" -U sub2api -h 127.0.0.1 -d sub2api

# 查看所有表
\dt

# 查看用户列表
SELECT id, email, role, status, balance FROM users WHERE deleted_at IS NULL;

# 执行 SQL 文件（避免中文路径问题）
PGPASSWORD=sub2api "C:/Program Files/PostgreSQL/16/bin/psql.exe" -U sub2api -h 127.0.0.1 -d sub2api -f "C:/temp/xxx.sql"
```

> **坑：** psql 不支持中文路径！如果 SQL 文件在中文目录下，先复制到 `C:\temp\` 再执行。

---

## 八、开发常用命令

```bash
# 后端单元测试
cd backend && go test -tags=unit ./...

# 后端 lint 检查
cd backend && golangci-lint run ./...

# Ent ORM 代码生成（改了 ent/schema 后必须执行）
cd backend && go generate ./ent

# 前端构建（检查是否能正常打包）
cd frontend && pnpm build

# 前端类型检查
cd frontend && npx vue-tsc --noEmit
```

---

## 九、端口清单

| 服务 | 端口 | 说明 |
|------|------|------|
| 前端 Vite | 3000 | 开发服务器，自动代理 API 到 8080 |
| 后端 Go | 8080 | API 服务器 |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |

---

## 十、给 Claude Code 的提示

如果你是 Claude Code 在读这个文档，帮 Master 启动本地环境时按这个顺序来：

1. 检查 PostgreSQL 服务是否运行：`Get-Service postgresql-x64-16` 或 `netstat -ano | grep :5432`
2. 启动 Redis：`docker start redis`，检查 `docker ps | grep redis`
3. 检查端口 8080 和 3000 是否被占用，被占了就杀掉
4. 确认 `backend/.installed` 文件存在，不存在就创建
5. 启动后端：`cd backend && export DATA_DIR="$(pwd)" && go run ./cmd/server/`（后台运行）
6. 启动前端：`cd frontend && pnpm dev`（后台运行）
7. 告诉 Master 访问 http://localhost:3000，账号 admin@126.com / admin123

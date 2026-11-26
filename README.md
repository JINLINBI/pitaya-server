# pitaya-server

个人游戏服务器框架（[Pitaya](https://github.com/topfreegames/pitaya) + GORM + Luban 配置热更）。

提供：**账户注册/登录（JWT）**、**connector 绑会话**、大厅聊天、**邮箱示例玩法**，以及单体 / 集群（connector × N + account + room）部署。

约定见 [CLAUDE.md](./CLAUDE.md)。新增业务照抄 [docs/adding-a-service.md](./docs/adding-a-service.md)。

## 一键验收（推荐）

集群是默认验收路径：

```bash
make verify
# 等价：docker compose up -d --build → 等健康口 → scripts/cluster_smoke.sh
```

常用运维：

| 命令 | 作用 |
|------|------|
| `make up` | 构建并后台启动集群 |
| `make down` | 停止容器 |
| `make rebuild` | 强制重建并重启 |
| `make logs` / `make ps` | 看日志 / 状态 |
| `make wait-healthy` | 轮询 `18081–18084/healthz` |
| `make cluster-smoke` | 仅烟雾（假定已 up） |
| `make test` | `go test ./...` |
| `make clean` | down -v + 清理 bin |

可选：复制 `.env.example` 为 `.env`，统一改 `AUTH_JWT_SECRET`（所有服务必须一致）。

## 本地单体（开发）

```bash
export AUTH_JWT_SECRET=dev-secret
export ADMIN_TOKEN=dev-admin   # 可选：开启 POST /admin/cfg/reload

go run ./cmd/server --type monolith --port 3250

# 注册 + 登录 + 绑会话 + 进大厅
go run ./cmd/client --addr 127.0.0.1:3250 --user alice --password secret --register
```

无密码 demo 登录（仅开发，`ALLOW_DEMO_LOGIN=1`）：

```bash
ALLOW_DEMO_LOGIN=1 go run ./cmd/server --type monolith --port 3250
go run ./cmd/client --addr 127.0.0.1:3250 --user alice
```

## 配置热更（Luban JSON）

数据目录默认 `data/cfg`（`LUBAN_DATA_DIR` 可改）。修改例如 `data/cfg/tbglobalconfig.json` 后：

```bash
curl -X POST http://127.0.0.1:8008/admin/cfg/reload \
  -H "X-Admin-Token: $ADMIN_TOKEN"
```

表代码：`gen/cfg`；策划源与生成脚本见 `configs/luban`、`scripts/gen_luban.sh`（需 .NET + Luban）。

## 本地集群（手动）

```bash
make up && make wait-healthy

# 集群需 --cluster：客户端发三元组路由（account.account.* / room.room.*）
go run ./cmd/client --addr 127.0.0.1:38080 --user alice --password secret --register --cluster
go run ./cmd/client --addr 127.0.0.1:38081 --user bob --password secret --register --cluster
```

健康口：connector `18081`/`18082`，account `18083`，room `18084`。

账户与邮箱数据在 **account** 服务 SQLite 卷；connector / room 不连库。

## 启动参数

| 参数 | 说明 |
|------|------|
| `--type` | `monolith` / `connector` / `room` / **`account`** |
| `--port` | TCP 监听（仅 frontend） |
| `--node_id` | 节点编号（BTA） |
| `--health_port` | 默认 `8008`，`/healthz` |

## 环境变量

| 变量 | 说明 |
|------|------|
| `PITAYA_ETCD_ENDPOINTS` / `PITAYA_NATS_URL` | **集群必填**（etcd 亦用于共享 lobby 群组） |
| `AUTH_JWT_SECRET` | JWT 密钥（connector + account 需一致）；集群 / `PITAYA_ENV=production` 禁止默认 `dev-change-me`，且至少 16 字符 |
| `PITAYA_ENV` | `development`（默认）\| `production`；生产拒绝 `ALLOW_DEMO_LOGIN` |
| `DATABASE_DSN` / `DATABASE_DRIVER` | 默认 SQLite 文件 `pitaya.db`；生产可 postgres/mysql DSN |
| `LUBAN_DATA_DIR` | Luban 导出 JSON 目录 |
| `ADMIN_TOKEN` | 设置后才暴露 `POST /admin/cfg/reload`；生产缺省不设更安全 |
| `ALLOW_DEMO_LOGIN` | `1` 时允许无密码 `connector.login` |

启动失败时 bootstrap 会打印多行 `invalid runtime config:` 说明缺哪项。

## 客户端路由

单体（两段）：

1. `account.register` / `account.login` → JWT
2. `connector.bind` `{"token":"..."}`
3. `room.join` → `room.message` notify
4. 示例：`mailbox.send` / `mailbox.list` / `mailbox.markread`
5. Push：`onMessage` / `onNewUser` / `onMembers`

集群（CLI 加 `--cluster`，映射为三元组 `serverType.service.method`）：

- `account.account.register` / `account.account.login`
- `connector.bind`（仍在 frontend）
- `room.room.join` / `room.room.message`
- `account.mailbox.*`（mailbox 与 account 同进程）

## 仓库

```text
https://github.com/JINLINBI/pitaya-server
```

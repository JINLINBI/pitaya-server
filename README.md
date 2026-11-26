# pitaya-server

个人游戏服务器框架（[Pitaya](https://github.com/topfreegames/pitaya) + GORM + Luban 配置热更）。

提供：**账户注册/登录（JWT）**、**connector 绑会话**、大厅聊天，以及单体 / 集群（connector × N + account + room）部署。

约定见 [CLAUDE.md](./CLAUDE.md)。

## 本地单体

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

## 本地集群

```bash
docker compose up -d --build

go run ./cmd/client --addr 127.0.0.1:38080 --user alice --password secret --register
go run ./cmd/client --addr 127.0.0.1:38081 --user bob --password secret --register
```

账户数据仅在 **account** 服务容器内的 SQLite 卷；connector 不连库。

集群烟雾（compose 已启动后）：

```bash
make cluster-smoke
```

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
| `PITAYA_ETCD_ENDPOINTS` / `PITAYA_NATS_URL` | 集群（etcd 亦用于共享 **lobby** 群组） |
| `DATABASE_DSN` / `DATABASE_DRIVER` | 默认 SQLite 文件 `pitaya.db`；生产可 postgres/mysql DSN |
| `AUTH_JWT_SECRET` | JWT 密钥（connector + account 需一致）；**集群模式禁止**使用默认 `dev-change-me` |
| `LUBAN_DATA_DIR` | Luban 导出 JSON 目录 |
| `ADMIN_TOKEN` | 配置后启用 `POST /admin/cfg/reload` |
| `ALLOW_DEMO_LOGIN` | `1` 时允许无密码 `connector.login` |

## 客户端路由

1. `account.register` / `account.login` → JWT
2. `connector.bind` `{"token":"..."}`
3. `room.join` → `room.message` notify
4. Push：`onMessage` / `onNewUser` / `onMembers`

## 仓库

```text
https://github.com/JINLINBI/pitaya-server
```
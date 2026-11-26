# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

个人 **Pitaya 游戏服务器框架**起步仓库。当前阶段提供：

- GORM 账户（注册/登录 JWT）+ connector 绑会话 + 大厅聊天
- Luban 服务端配置（`gen/cfg` + `data/cfg`）与 HTTP 热重载
- TCP connector（可多实例）+ room 后端
- CLI 客户端，可指定不同 `--addr` 验证多容器 hostPort 暴露
- 与 **backend-tool-assistant（BTA）** 对接的轻量镜像 / 启动参数约定

模块路径：`github.com/JINLINBI/pitaya-server`（Go module，不是可安装库；从仓库根直接 `go run`）。

本仓库目标是演进成个人服务器框架；**先把聊天骨架与部署契约做稳**，再加玩法系统。

## 常用命令

```bash
# 默认验收（集群 up + 健康等待 + 烟雾）
make verify

# 单体（同进程 connector + account + mailbox + room）
AUTH_JWT_SECRET=dev-secret go run ./cmd/server --type monolith --port 3250
go run ./cmd/client --addr 127.0.0.1:3250 --user alice --password secret --register

# 集群运维（CLI 需 --cluster）
make up / make down / make rebuild / make logs / make cluster-smoke
go run ./cmd/client --addr 127.0.0.1:38080 --user alice --password secret --register --cluster

# 构建与单测
make build && make test
```

健康检查：单体 `GET http://127.0.0.1:8008/healthz`；compose 下 connector `18081/18082`、account `18083`、room `18084`。

新增业务：复制 [docs/adding-a-service.md](./docs/adding-a-service.md)（对照 `mailbox`）。

## 架构

```
cmd/server                    monolith | connector | account | room
internal/bootstrap            装配 DB / JWT / cfg.Manager / domain services
internal/platform             db, auth, cfg 热更, env（集群校验 JWT/etcd/nats）
internal/domain/account       账户用例（无 Pitaya 依赖）
internal/domain/mailbox       邮箱示例领域（无 Pitaya 依赖）
internal/services/account     account.register / account.login
internal/services/mailbox     mailbox.send / list / markread
internal/services/connector   connector.bind（JWT）；ALLOW_DEMO_LOGIN 时 connector.login
internal/services/room        大厅聊天
internal/health               /healthz；ADMIN_TOKEN 时 /admin/cfg/reload
gen/cfg + data/cfg            Luban 生成物与 JSON 数据
configs/pitaya.yaml           Pitaya 参考配置
docs/adding-a-service.md      新服务注册约定
```

### 运行模式

| `--type` | 说明 |
|----------|------|
| `monolith` / `standalone` | `pitaya.Standalone`；同进程注册 connector + room；无需 etcd/nats |
| `connector` | 集群 frontend；TCP acceptor；需 etcd + nats |
| `room` | 集群 backend；注册 handler + remote；需 etcd + nats |
| `account` | 集群 backend；账户与 DB；注册 handler + remote |

环境变量（集群）：

- `PITAYA_ETCD_ENDPOINTS`（逗号分隔，如 `etcd:2379`；集群下同时用于 **Groups** 与 SD）
- `PITAYA_NATS_URL`（如 `nats://nats:4222`）

CLI / 容器参数（对齐 BTA workload）：

- `--type`、`--port`（仅 frontend）、`--node_id`、`--health_port`

默认 connector 容器内端口 **3250**；compose / BTA 集群模板把两个 connector 分别映到宿主机 **38080 / 38081**。

环境变量（平台）：

- `DATABASE_DSN` / `DATABASE_DRIVER`（默认 SQLite `pitaya.db`）
- `AUTH_JWT_SECRET`（集群 / production 禁止默认 `dev-change-me`，≥16 字符）
- `PITAYA_ENV`（`development`\|`production`；生产拒绝 demo 登录）
- `LUBAN_DATA_DIR`（默认 `data/cfg`）
- `ADMIN_TOKEN`（配置后启用热更 API）
- `ALLOW_DEMO_LOGIN`（默认关）

集群启动时 `env.Validate` 会校验 etcd / nats / JWT，失败信息为多行列表。

### 客户端路由

1. `account.register` / `account.login`
2. `connector.bind` — `{"token":"..."}`
3. `room.join` → `room.message`
4. `mailbox.send` / `mailbox.list` / `mailbox.markread`（示例玩法；集群挂 account 后端）
5. Push：`onMessage` / `onNewUser` / `onMembers`

大厅 group 名：`lobby`。集群下 `GroupBroadcast` 的 frontend 类型固定为 `connector`。

## 与 BTA 的契约

backend-tool-assistant 内置模板 `pitaya-server` / `pitaya-server-cluster` 会：

- 用本仓库 Dockerfile 风格的 inline 构建（`go build ./cmd/server`）
- 集群 ephemeral：`etcd` + `nats`，注入 `PITAYA_ETCD_ENDPOINTS` / `PITAYA_NATS_URL`
- workload `container_args` 走 `--type` / `--node_id`（覆盖镜像 CMD）
- 健康口：`container_port=8008`，`health_path=/healthz`

改启动参数、环境变量名或健康路径时，必须同步考虑 BTA 模板，否则一键部署会断。

GitLab 项目占位常为 `demo/pitaya-server`；真实流水线需改成可达仓库。

## 代码约定

- 注释、日志、面向用户的 CLI 文案、提交说明用**中文**；提交用 Conventional Commits（如 `feat(room): ...`、`fix(client): ...`）
- 包分层：`cmd` 只装配；`internal/domain/*` 无 Pitaya；`internal/services/*` 薄适配（`Deps`，不直接依赖 `bootstrap`）；集群 Groups 经 `internal/serverwire` 接 etcd
- 依赖 Pitaya v3（`github.com/topfreegames/pitaya/v3`）；传输默认 **TCP**（`pitaya/pkg/client` + `TCPAcceptor`）
- 持久化默认 SQLite（`github.com/glebarez/sqlite` 纯 Go，镜像 CGO_ENABLED=0）；生产用 `DATABASE_DSN` 切 Postgres/MySQL
- 镜像：`CGO_ENABLED=0` 静态编译；`ENTRYPOINT` 为 `/app/pitaya-server`，`CMD` 可被 K8s args 覆盖

## 非目标（当前阶段）

- 完整玩法 / 数值
- WebSocket 网关或 Unity 客户端（可后续加，勿破坏现有 TCP CLI 验收）
- 在本仓库内嵌 BTA 或改 BTA 源码（BTA 在独立仓库）

## 验收清单

本地单体：单进程 login + 聊天通。

本地集群：`make verify`（推荐）；或两个 CLI 分别连 `38080` / `38081`。

BTA：集群模板部署后，同样用 CLI 验双端口。

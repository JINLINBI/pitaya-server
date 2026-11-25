# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

个人 **Pitaya 游戏服务器框架**起步仓库。当前阶段提供：

- 登录（会话 Bind）+ 大厅聊天
- TCP connector（可多实例）+ room 后端
- CLI 客户端，可指定不同 `--addr` 验证多容器 hostPort 暴露
- 与 **backend-tool-assistant（BTA）** 对接的轻量镜像 / 启动参数约定

模块路径：`github.com/JINLINBI/pitaya-server`（Go module，不是可安装库；从仓库根直接 `go run`）。

本仓库目标是演进成个人服务器框架；**先把聊天骨架与部署契约做稳**，再加玩法系统。

## 常用命令

```bash
# 单体（同进程 connector + room，Standalone）
go run ./cmd/server --type monolith --port 3250

# CLI
go run ./cmd/client --addr 127.0.0.1:3250 --user alice

# 本地集群（etcd + nats + connector×2 + room）
docker compose up -d --build
go run ./cmd/client --addr 127.0.0.1:38080 --user alice
go run ./cmd/client --addr 127.0.0.1:38081 --user bob

# 构建
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
go test ./...
```

健康检查：`GET http://127.0.0.1:8008/healthz`（`--health_port` 可改）。

## 架构

```
cmd/server          按 --type 启动 monolith | connector | room
cmd/client          TCP CLI：login → join → 交互发消息
internal/services/connector   connector.login（Bind UID）
internal/services/room        room.join / room.message；推送 onMessage/onNewUser/onMembers
internal/health               HTTP /healthz 与 /
configs/pitaya.yaml           参考配置；运行时优先环境变量
```

### 运行模式

| `--type` | 说明 |
|----------|------|
| `monolith` / `standalone` | `pitaya.Standalone`；同进程注册 connector + room；无需 etcd/nats |
| `connector` | 集群 frontend；TCP acceptor；需 etcd + nats |
| `room` | 集群 backend；注册 handler + remote；需 etcd + nats |

环境变量（集群）：

- `PITAYA_ETCD_ENDPOINTS`（逗号分隔，如 `etcd:2379`）
- `PITAYA_NATS_URL`（如 `nats://nats:4222`）

CLI / 容器参数（对齐 BTA workload）：

- `--type`、`--port`（仅 frontend）、`--node_id`、`--health_port`

默认 connector 容器内端口 **3250**；compose / BTA 集群模板把两个 connector 分别映到宿主机 **38080 / 38081**。

### 客户端路由

1. Request `connector.login` — `{"username":"..."}`（无密码，demo）
2. Request `room.join`
3. Notify `room.message` — `{"name":"...","content":"..."}`
4. Push：`onMessage` / `onNewUser` / `onMembers`

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
- 包分层：`cmd` 只装配；业务在 `internal/services/*`；不要在 `cmd` 堆 handler
- 依赖 Pitaya v3（`github.com/topfreegames/pitaya/v3`）；传输默认 **TCP**（`pitaya/pkg/client` + `TCPAcceptor`）
- 不引入 MySQL；demo 无真实账号体系
- 镜像：`CGO_ENABLED=0` 静态编译；`ENTRYPOINT` 为 `/app/pitaya-server`，`CMD` 可被 K8s args 覆盖

## 非目标（当前阶段）

- 完整玩法 / 数值 / 持久化账号
- WebSocket 网关或 Unity 客户端（可后续加，勿破坏现有 TCP CLI 验收）
- 在本仓库内嵌 BTA 或改 BTA 源码（BTA 在独立仓库）

## 验收清单

本地单体：单进程 login + 聊天通。

本地集群：两个 CLI 分别连 `38080` 与 `38081`，互发消息可见。

BTA：集群模板部署后，同样用 CLI 验双端口。

# pitaya-server

个人游戏服务器框架起步项目（Pitaya）。当前提供：登录 + 大厅聊天，以及可连不同 connector 端口的 CLI，用来验证容器 hostPort 暴露。

本目录是独立 Go module / 独立 git 仓库，目前以嵌套方式放在 `backend-tool-assistant/pitaya-server` 下开发；推送到远端后可改成正式 git submodule。

## 本地单体

```bash
go run ./cmd/server --type monolith --port 3250
# 另一终端
go run ./cmd/client --addr 127.0.0.1:3250 --user alice
```

## 本地集群（双 connector 验端口）

```bash
docker compose up -d --build
go run ./cmd/client --addr 127.0.0.1:38080 --user alice
go run ./cmd/client --addr 127.0.0.1:38081 --user bob
```

两边互发消息可见，即跨 connector RPC + 端口映射均正常。

## 启动参数

| 参数 | 说明 |
|------|------|
| `--type` | `monolith` / `connector` / `room` |
| `--port` | TCP 监听（仅 frontend） |
| `--node_id` | 节点编号（对齐 BTA workload） |
| `--health_port` | 默认 `8008`，路径 `/healthz` |

环境变量：

- `PITAYA_ETCD_ENDPOINTS`（如 `etcd:2379`）
- `PITAYA_NATS_URL`（如 `nats://nats:4222`）

## 客户端路由

1. `connector.login` `{"username":"..."}`
2. `room.join`
3. `room.message` notify `{"name":"...","content":"..."}`
4. 推送：`onMessage` / `onNewUser` / `onMembers`

# BTA 应用模板（本仓库为权威源）

本目录 JSON 是 **backend-tool-assistant（BTA）** 部署 Pitaya 的应用模板。助手侧不再内嵌 Pitaya 模板；通过 `app_preset_dir` 导入即可。

## 包含模板

| 文件 | `id` | 用途 |
|------|------|------|
| `pitaya-server.json` | `pitaya-server` | 单体（monolith），hostPort `38080→3250` |
| `pitaya-server-cluster.json` | `pitaya-server-cluster` | 集群：connector×2 + account + room；临时 etcd/nats |

健康口：`container_port=8008`，`health_path=/healthz`。集群 connector 宿主机端口与本地 compose 一致：`38080` / `38081`。

## 让助手导入

在 BTA 的 `configs/config.yaml`（或等价配置）中指向本目录的**绝对路径**（或助手机器上的 checkout 路径）：

```yaml
app_preset_dir: /path/to/pitaya-server/deploy/bta
# 默认 missing_only：已有同 id 自定义模板则跳过
# app_preset_dir_import_mode: always
# app_preset_dir_strict: false
```

也可用环境变量：

```bash
export ASSISTANT_APP_PRESET_DIR=/path/to/pitaya-server/deploy/bta
```

本地开发示例（两仓并列 checkout）：

```yaml
app_preset_dir: ../pitaya-server/deploy/bta
```

也可在 UI **应用模板 → 导入 JSON**，逐个选本目录下的文件。

## 验收

模板部署到 sandbox 后，用本仓 CLI（集群加 `--cluster`）：

```bash
go run ./cmd/client --addr <public_host>:38080 --user alice --password secret --register --cluster
go run ./cmd/client --addr <public_host>:38081 --user bob --password secret --register --cluster
```

## 注意

- `gitlab_project` 默认为占位 `demo/pitaya-server`；对接真实流水线时改成可达 GitLab 路径。
- 沙盒 JWT 字面量仅供 demo；生产请改 `AUTH_JWT_SECRET`。
- 改启动参数 / 健康路径 / 环境变量名时，请同步改本目录 JSON 与根 `Dockerfile` / `docker-compose.yml`。

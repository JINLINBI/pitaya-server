# 新增业务服务（复制此模式）

本仓库分层约定：**domain 无 Pitaya 依赖** → **services 薄适配** → **bootstrap 装配 Deps** → **cmd/server 按 `--type` 注册**。

邮箱（`mailbox`）是可运行的示例，可直接对照复制。

## 1. Domain（`internal/domain/<name>/`）

```text
model.go       # 实体 / GORM 标签
repository.go  # 持久化
service.go     # 用例；只依赖 stdlib + platform/db 相关，禁止 import pitaya
service_test.go
```

- 错误用包级 `var ErrXxx = errors.New(...)`
- 不读 `os.Getenv`；配置经构造注入（见 `account.Service` 的 `cfg func()`）

## 2. Service handler（`internal/services/<name>/`）

```text
deps.go     # type Deps struct { ... }  — 不依赖 bootstrap
<name>.go   # Pitaya component；从 session 取 UID，调用 domain
```

```go
type Deps struct {
    Mailbox *domain.Service
}

func New(app pitaya.Pitaya, d Deps) *Handler { ... }
```

## 3. Bootstrap

在 `internal/bootstrap/bootstrap.go`：

1. `Platform` 增加字段（如 `Mailbox *mailbox.Service`）
2. `NeedDB` 为真时 `NewRepository` + `NewService` 赋给 Platform
3. 若有新表：在 `internal/platform/db/db.go` 的 `AutoMigrate` 追加实体

## 4. 注册（`cmd/server/main.go`）

| 模式 | 怎么挂 |
|------|--------|
| `monolith` / `standalone` | `app.Register`（本进程前端可直达） |
| 集群 backend（需 DB，如 account） | `Register` + `RegisterRemote`，与 account 同进程或新建 `--type` |
| 仅前端（connector） | 一般不挂业务；由 backend remote 处理 |

示例（单体）：

```go
func registerMailbox(app pitaya.Pitaya, platform *bootstrap.Platform) {
    h := mailbox.New(app, mailbox.Deps{Mailbox: platform.Mailbox})
    app.Register(h, component.WithName("mailbox"), component.WithNameFunc(strings.ToLower))
}
```

路由名 = 组件名 + 方法名小写，如 `mailbox.send` / `mailbox.list` / `mailbox.markread`。

## 5. 清单

- [ ] domain 单元测试（不启 Pitaya）
- [ ] `go test ./...`
- [ ] 单体：`AUTH_JWT_SECRET=dev-secret go run ./cmd/server --type monolith --port 3250`
- [ ] 集群：业务挂在带 DB 的 backend；`make verify`

## 反模式

- 在 `cmd/server` 写业务逻辑
- domain 引用 `pitaya` / `component`
- services 直接 import `bootstrap`（用 `Deps` 注入）

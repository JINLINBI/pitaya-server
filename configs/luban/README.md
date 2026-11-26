# Luban 配置

- `Defines/`、`Datas/`：表结构与策划数据（可用 Excel/JSON）。
- 生成：仓库根执行 `scripts/gen_luban.sh`（需 .NET 8 + [Luban](https://github.com/focus-creative-games/luban)）。

当前仓库已 **检入** `gen/cfg` 与 `data/cfg` 示例，便于无 Luban 环境直接 `go build`。修改表后请重新生成并提交。
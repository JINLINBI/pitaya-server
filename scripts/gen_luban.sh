#!/usr/bin/env bash
# 使用 Luban.ClientServer 生成 Go 代码与 JSON 数据（server 组）。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LUBAN_DLL="${LUBAN_DLL:-Luban.ClientServer.dll}"

if [[ ! -f "$LUBAN_DLL" ]]; then
  echo "请设置 LUBAN_DLL 指向 Luban.ClientServer.dll，或从 Luban Release 下载后放在仓库根。" >&2
  echo "示例: LUBAN_DLL=/path/to/Luban.ClientServer.dll $0" >&2
  exit 1
fi

cd "$ROOT"
dotnet "$LUBAN_DLL" -j cfg -- \
  --conf "configs/luban/luban.conf" \
  -s server \
  --gen_types code_go_json,data_json \
  -x lubanGoModule=github.com/JINLINBI/pitaya-server/gen/cfg \
  -x go-json.outputCodeDir=gen/cfg \
  -x json.outputDataDir=data/cfg

echo "Luban 生成完成: gen/cfg data/cfg"
#!/usr/bin/env bash
# 集群烟雾验收：健康检查 + 双 connector 注册登录。
# 用法：
#   ./scripts/cluster_smoke.sh          # 假定 compose 已启动
#   CLUSTER_SMOKE_UP=1 ./scripts/cluster_smoke.sh  # 自动 up + wait
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ "${CLUSTER_SMOKE_UP:-0}" == "1" ]]; then
	echo "== docker compose up =="
	docker compose up -d --build
	chmod +x ./scripts/wait_healthy.sh
	./scripts/wait_healthy.sh
fi

health_urls=(
	"http://127.0.0.1:18081/healthz"
	"http://127.0.0.1:18082/healthz"
	"http://127.0.0.1:18083/healthz"
	"http://127.0.0.1:18084/healthz"
)
for u in "${health_urls[@]}"; do
	if ! curl -sf "$u" >/dev/null; then
		echo "health check failed: $u" >&2
		echo "hint: run 'make up' (or CLUSTER_SMOKE_UP=1 make cluster-smoke), then retry" >&2
		exit 1
	fi
done

SUFFIX="${RANDOM}"
ALICE="smoke_alice_${SUFFIX}"
BOB="smoke_bob_${SUFFIX}"
PASS="smoke-secret"

echo "== register $ALICE on :38080 =="
printf 'hi from alice\n' | go run ./cmd/client --addr 127.0.0.1:38080 --user "$ALICE" --password "$PASS" --register --cluster

echo "== register $BOB on :38081 =="
printf 'hi from bob\n' | go run ./cmd/client --addr 127.0.0.1:38081 --user "$BOB" --password "$PASS" --register --cluster

echo "cluster smoke ok"

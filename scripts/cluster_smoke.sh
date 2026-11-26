#!/usr/bin/env bash
# 集群烟雾：需先 docker compose up -d --build
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

health_urls=(
	"http://127.0.0.1:18081/healthz"
	"http://127.0.0.1:18082/healthz"
)
for u in "${health_urls[@]}"; do
	if ! curl -sf "$u" >/dev/null; then
		echo "health check failed: $u (is compose up?)"
		exit 1
	fi
done

SUFFIX="${RANDOM}"
ALICE="smoke_alice_${SUFFIX}"
BOB="smoke_bob_${SUFFIX}"
PASS="smoke-secret"

echo "== register $ALICE on :38080 =="
printf 'hi from alice\n' | go run ./cmd/client --addr 127.0.0.1:38080 --user "$ALICE" --password "$PASS" --register

echo "== register $BOB on :38081 =="
printf 'hi from bob\n' | go run ./cmd/client --addr 127.0.0.1:38081 --user "$BOB" --password "$PASS" --register

echo "cluster smoke ok"
#!/usr/bin/env bash
# 等待 compose 集群健康口就绪。
set -euo pipefail

TIMEOUT="${WAIT_HEALTHY_TIMEOUT:-120}"
INTERVAL="${WAIT_HEALTHY_INTERVAL:-2}"

urls=(
	"http://127.0.0.1:18081/healthz"
	"http://127.0.0.1:18082/healthz"
	"http://127.0.0.1:18083/healthz"
	"http://127.0.0.1:18084/healthz"
)

deadline=$((SECONDS + TIMEOUT))
echo "waiting for health (timeout=${TIMEOUT}s)..."
while (( SECONDS < deadline )); do
	ok=1
	for u in "${urls[@]}"; do
		if ! curl -sf "$u" >/dev/null 2>&1; then
			ok=0
			break
		fi
	done
	if (( ok == 1 )); then
		echo "all health checks passed"
		exit 0
	fi
	sleep "$INTERVAL"
done

echo "health wait timed out after ${TIMEOUT}s" >&2
for u in "${urls[@]}"; do
	echo "  probe $u -> $(curl -s -o /dev/null -w '%{http_code}' "$u" || echo fail)" >&2
done
exit 1

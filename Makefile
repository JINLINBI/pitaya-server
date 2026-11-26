.PHONY: build test luban up down rebuild logs ps wait-healthy cluster-smoke verify clean

build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

test:
	go test ./...

luban:
	./scripts/gen_luban.sh

# --- 集群部署 ---

up:
	docker compose up -d --build

down:
	docker compose down

rebuild:
	docker compose up -d --build --force-recreate

logs:
	docker compose logs -f --tail=200

ps:
	docker compose ps

wait-healthy:
	chmod +x ./scripts/wait_healthy.sh
	./scripts/wait_healthy.sh

cluster-smoke:
	chmod +x ./scripts/cluster_smoke.sh
	./scripts/cluster_smoke.sh

# 默认验收路径：拉起集群 → 等健康 → 烟雾
verify: up wait-healthy cluster-smoke

clean:
	docker compose down -v --remove-orphans
	rm -f bin/server bin/client

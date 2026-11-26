.PHONY: build test luban cluster-smoke

build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

test:
	go test ./...

luban:
	./scripts/gen_luban.sh

cluster-smoke:
	chmod +x ./scripts/cluster_smoke.sh
	./scripts/cluster_smoke.sh
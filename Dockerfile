# pitaya-server — 轻量静态二进制，供 BTA / 本地 compose 使用。
ARG BUILD_IMAGE=golang:1.25-alpine
ARG RUN_IMAGE=alpine:3.20

FROM ${BUILD_IMAGE} AS builder

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY} \
	CGO_ENABLED=0 \
	GOTOOLCHAIN=local

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/pitaya-server ./cmd/server \
	&& go build -trimpath -ldflags="-s -w" -o /out/pitaya-client ./cmd/client

FROM ${RUN_IMAGE}
WORKDIR /app
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /out/pitaya-server /app/pitaya-server
COPY --from=builder /out/pitaya-client /app/pitaya-client
COPY data/cfg /app/data/cfg
RUN chmod +x /app/pitaya-server /app/pitaya-client
EXPOSE 3250 8008
HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=3 \
	CMD wget -q -O /dev/null http://127.0.0.1:8008/healthz || exit 1
ENTRYPOINT ["/app/pitaya-server"]
CMD ["--type", "monolith", "--port", "3250"]

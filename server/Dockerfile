# syntax=docker/dockerfile:1

# ---------- 构建阶段 ----------
# SQLite 驱动（mattn/go-sqlite3）依赖 CGO，构建镜像需要 gcc / musl-dev
FROM golang:1.27-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /src

# 先复制依赖清单，代码变更时可复用依赖层缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/lfdd .

# ---------- 运行阶段 ----------
FROM alpine:3.22

# ca-certificates：HTTPS 出站请求；tzdata：本地时间；libgcc：CGO 运行时依赖
RUN apk add --no-cache ca-certificates tzdata libgcc \
    && mkdir -p /app/data

WORKDIR /app

COPY --from=builder /out/lfdd /app/lfdd

# 容器内默认走 HTTP，TLS 交给外层反向代理（如需容器内 HTTPS，运行时覆盖 LFDD_SERVER_HTTPS=true 并挂载证书）
# 其余配置（端口/数据库/上传目录等）默认值已指向 /app/data，可直接通过挂载的 data/.env 或 docker -e 修改

VOLUME ["/app/data"]
EXPOSE 9081

ENTRYPOINT ["/app/lfdd"]

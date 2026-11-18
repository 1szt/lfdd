# syntax=docker/dockerfile:1

# ---------- 构建阶段 ----------
FROM node:24-alpine AS builder

RUN npm install -g pnpm@12

WORKDIR /app

# 源码必须在 install 之前就位：package.json 的 postinstall 会执行 quasar prepare，读取 quasar.config.ts
COPY . .
RUN pnpm install --frozen-lockfile

# 后端地址在构建时写入产物（Vite 静态注入，运行期无法修改；留空表示与页面同源）
ARG API_BASE=""
RUN if [ -n "$API_BASE" ]; then echo "VITE_API_BASE=$API_BASE" > .env; fi \
    && pnpm quasar build -m spa

# ---------- 运行阶段 ----------
FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist/spa /usr/share/nginx/html

EXPOSE 80

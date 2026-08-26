# 运行阶段：二进制由本地编译好后放入构建上下文（仓库根目录）
# 注意：容器为 debian/linux，请放置 Linux 版二进制（GOOS=linux）
FROM node:24-bookworm-slim
WORKDIR /root
ENV TZ="Asia/Shanghai"

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ffmpeg \
        fonts-noto-* \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY arknights_bot /root/arknights_bot
COPY renderer/package.json renderer/package-lock.json /root/renderer/
WORKDIR /root/renderer
RUN npm ci --omit=dev
COPY renderer/*.mjs /root/renderer/
COPY renderer/lib /root/renderer/lib
COPY renderer/components /root/renderer/components
WORKDIR /root
CMD ["/root/arknights_bot"]

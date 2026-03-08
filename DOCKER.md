# Docker 使用指南

## 🐳 镜像信息

- **镜像名称**: `infinite-refill`
- **标签**: `latest`, `v1.0`
- **大小**: 18MB
- **架构**: linux/amd64
- **基础镜像**: alpine:latest

## 📦 镜像已构建

```bash
$ docker images | grep infinite-refill
infinite-refill   latest   ebedbb3984e9   18MB
infinite-refill   v1.0     ebedbb3984e9   18MB
```

## 🚀 快速开始

### 1. 检查版本

```bash
docker run --rm infinite-refill:latest version
```

输出：
```
refill version v1.0 (built 2026-03-08_09:08:22)
```

### 2. 检查配置

```bash
docker run --rm \
  -e SERVER_URL="https://your-server.com" \
  -e USER_KEY="your-user-key-here" \
  -e ACCOUNTS_DIR="/data/accounts" \
  infinite-refill:latest check
```

### 3. 单次续杯

```bash
docker run --rm \
  -v ./accounts:/data/accounts \
  -v ./out:/data/out \
  -e SERVER_URL="https://your-server.com" \
  -e USER_KEY="your-user-key-here" \
  -e ACCOUNTS_DIR="/data/accounts" \
  infinite-refill:latest run
```

### 4. 同步所有账号

```bash
docker run --rm \
  -v ./accounts:/data/accounts \
  -v ./out:/data/out \
  -e SERVER_URL="https://your-server.com" \
  -e USER_KEY="your-user-key-here" \
  -e ACCOUNTS_DIR="/data/accounts" \
  infinite-refill:latest sync
```

### 5. 清理失效账号

```bash
# 预览模式
docker run --rm \
  -v ./accounts:/data/accounts \
  -v ./out:/data/out \
  -e SERVER_URL="https://your-server.com" \
  -e USER_KEY="your-user-key-here" \
  -e ACCOUNTS_DIR="/data/accounts" \
  infinite-refill:latest clean

# 实际删除
docker run --rm \
  -v ./accounts:/data/accounts \
  -v ./out:/data/out \
  -e SERVER_URL="https://your-server.com" \
  -e USER_KEY="your-user-key-here" \
  -e ACCOUNTS_DIR="/data/accounts" \
  infinite-refill:latest clean --apply
```

### 6. 启动定时任务

```bash
docker run -d \
  --name refill-scheduler \
  --restart unless-stopped \
  -v ./accounts:/data/accounts \
  -v ./out:/data/out \
  -e SERVER_URL="https://your-server.com" \
  -e USER_KEY="your-user-key-here" \
  -e ACCOUNTS_DIR="/data/accounts" \
  -e SCHEDULER_INTERVAL_MINUTES=30 \
  infinite-refill:latest scheduler start
```

## 📋 Docker Compose

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  refill:
    image: infinite-refill:latest
    container_name: refill-scheduler
    restart: unless-stopped
    volumes:
      - ./accounts:/data/accounts
      - ./out:/data/out
    environment:
      - SERVER_URL=https://your-server.com
      - USER_KEY=your-user-key-here
      - ACCOUNTS_DIR=/data/accounts
      - TARGET_POOL_SIZE=10
      - TOTAL_HOLD_LIMIT=30
      - SCHEDULER_INTERVAL_MINUTES=30
      - PROBE_PARALLEL=6
      - PROXY_MODE=mixed
    command: scheduler start
```

启动：

```bash
docker-compose up -d
```

查看日志：

```bash
docker-compose logs -f
```

停止：

```bash
docker-compose down
```

## 🔧 环境变量

### 必需变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `SERVER_URL` | 服务器地址 | `https://your-server.com` |
| `USER_KEY` | 用户密钥 | `your-user-key-here` |
| `ACCOUNTS_DIR` | 账号目录 | `/data/accounts` |

### 可选变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `TARGET_POOL_SIZE` | 目标池大小 | `10` |
| `TOTAL_HOLD_LIMIT` | 总持有上限 | `50` |
| `SCHEDULER_INTERVAL_MINUTES` | 定时间隔（分钟） | `30` |
| `PROBE_PARALLEL` | 并发探测数 | `6` |
| `PROBE_WAIT_TIMEOUT` | 探测超时（秒） | `600` |
| `PROXY_MODE` | 代理模式 | `auto` |
| `PROXY_URL` | 代理地址 | `` |

## 📂 卷挂载

| 容器路径 | 说明 | 建议挂载 |
|----------|------|----------|
| `/data/accounts` | 账号文件目录 | `./accounts` |
| `/data/out` | 输出报告目录 | `./out` |

## 🎯 常用命令

### 查看运行中的容器

```bash
docker ps | grep refill
```

### 查看日志

```bash
docker logs -f refill-scheduler
```

### 停止容器

```bash
docker stop refill-scheduler
```

### 删除容器

```bash
docker rm refill-scheduler
```

### 进入容器

```bash
docker exec -it refill-scheduler sh
```

### 查看容器内文件

```bash
docker exec refill-scheduler ls -lh /data/accounts
docker exec refill-scheduler ls -lh /data/out
```

## 🔍 故障排查

### 问题 1: 权限错误

如果遇到权限问题，确保挂载的目录有正确的权限：

```bash
chmod 755 accounts
chmod 755 out
```

### 问题 2: 容器无法启动

查看日志：

```bash
docker logs refill-scheduler
```

### 问题 3: 账号目录为空

首次运行时，账号目录可能为空。运行 `sync` 或 `run` 命令获取账号：

```bash
docker run --rm \
  -v ./accounts:/data/accounts \
  -e SERVER_URL="..." \
  -e USER_KEY="..." \
  -e ACCOUNTS_DIR="/data/accounts" \
  infinite-refill:latest run
```

### 问题 4: 网络连接问题

如果需要使用代理：

```bash
docker run --rm \
  -v ./accounts:/data/accounts \
  -e SERVER_URL="..." \
  -e USER_KEY="..." \
  -e ACCOUNTS_DIR="/data/accounts" \
  -e PROXY_MODE=proxy \
  -e PROXY_URL="http://your-proxy.com:8080" \
  infinite-refill:latest run
```

## 🏗️ 构建镜像

如果需要重新构建镜像：

```bash
docker build -t infinite-refill:latest \
  --build-arg VERSION=v1.0 \
  --build-arg BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S') \
  .
```

构建多架构镜像：

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t infinite-refill:latest \
  --build-arg VERSION=v1.0 \
  --build-arg BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S') \
  .
```

## 📊 镜像特性

- ✅ **轻量级**: 仅 18MB
- ✅ **多阶段构建**: 最小化镜像大小
- ✅ **Alpine 基础**: 安全且高效
- ✅ **健康检查**: 自动监控容器状态
- ✅ **时区支持**: 包含 tzdata
- ✅ **CA 证书**: 支持 HTTPS 连接

## 🔐 安全建议

1. **不要在镜像中硬编码密钥**
   - 使用环境变量或 Docker secrets

2. **使用只读卷**（可选）
   ```bash
   -v ./accounts:/data/accounts:ro
   ```

3. **限制容器资源**
   ```bash
   docker run --memory="256m" --cpus="0.5" ...
   ```

4. **使用非 root 用户**（已在 Dockerfile 中配置）

## 📝 示例脚本

创建 `run-docker.sh`:

```bash
#!/bin/bash

# 配置
SERVER_URL="https://your-server.com"
USER_KEY="your-user-key-here"
ACCOUNTS_DIR="/data/accounts"

# 运行容器
docker run --rm \
  -v ./accounts:/data/accounts \
  -v ./out:/data/out \
  -e SERVER_URL="$SERVER_URL" \
  -e USER_KEY="$USER_KEY" \
  -e ACCOUNTS_DIR="$ACCOUNTS_DIR" \
  infinite-refill:latest "$@"
```

使用：

```bash
chmod +x run-docker.sh
./run-docker.sh check
./run-docker.sh run
./run-docker.sh sync
```

---

**镜像版本**: v1.0
**构建时间**: 2026-03-08
**状态**: ✅ 就绪
